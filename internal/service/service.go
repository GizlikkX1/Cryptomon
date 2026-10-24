// Package service is the per-currency microservice. One process tracks one
// asset: it polls every source, aggregates prices, finds the best rates,
// stores history in its own database and streams snapshots over SSE.
package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"math/rand/v2"
	"net"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"cryptomon/internal/backfill"
	"cryptomon/internal/catalog"
	"cryptomon/internal/httpx"
	"cryptomon/internal/market"
	"cryptomon/internal/model"
	"cryptomon/internal/sources"
	"cryptomon/internal/sse"
	"cryptomon/internal/storage"
)

type Config struct {
	Asset      catalog.Asset
	Addr       string        // listen address, e.g. 127.0.0.1:9101
	DataDir    string        // directory for the service database
	USDTURL    string        // base URL of the USDT service, for USDT→USD normalisation
	Interval   time.Duration // crypto exchange polling interval
	NoBackfill bool
	Version    string
}

type source struct {
	info     model.SourceInfo
	interval time.Duration
	fetch    func(ctx context.Context) ([]model.Quote, error)
}

type sourceState struct {
	quotes  []model.Quote
	lastOK  int64
	err     string
	fails   int
	latency int64
}

type refEntry struct {
	checked time.Time
	value   float64
}

type Service struct {
	cfg     Config
	log     *log.Logger
	info    model.AssetInfo
	store   *storage.Store
	broker  *sse.Broker
	sources []*source
	started time.Time

	mu    sync.RWMutex
	state map[string]*sourceState
	snap  *model.Snapshot
	seq   uint64

	usdt      atomic.Uint64 // float64 bits of the USDT/USD rate
	resetRefs atomic.Bool   // set after a backfill: 24h references must be re-read

	stored map[string]model.Point // last stored sample per series (snapshot goroutine only)
	refs   map[string]refEntry    // 24h reference per vs (snapshot goroutine only)
}

func New(cfg Config, logger *log.Logger) (*Service, error) {
	if cfg.Interval <= 0 {
		cfg.Interval = 3 * time.Second
	}
	a := cfg.Asset
	s := &Service{
		cfg:     cfg,
		log:     logger,
		broker:  sse.NewBroker(),
		state:   map[string]*sourceState{},
		stored:  map[string]model.Point{},
		refs:    map[string]refEntry{},
		started: time.Now(),
	}
	s.usdt.Store(math.Float64bits(1))
	s.info = model.AssetInfo{Code: a.Code, Name: a.Name, Kind: a.Kind, Symbol: a.Symbol, Color: a.Color, Version: cfg.Version}
	if a.Kind == model.Crypto {
		s.info.Quotes = []string{"USD"}
		for _, m := range a.Markets {
			if err := s.addExchange(m); err != nil {
				return nil, err
			}
		}
	} else {
		for _, c := range catalog.FiatCodes {
			if c != a.Code {
				s.info.Quotes = append(s.info.Quotes, c)
			}
		}
		for _, fx := range sources.FXSources {
			if fx.Supports(a.Code) {
				s.addFX(fx)
			}
		}
	}
	for _, src := range s.sources {
		s.info.Sources = append(s.info.Sources, src.info)
		s.state[src.info.ID] = &sourceState{}
	}
	return s, nil
}

func (s *Service) addExchange(m catalog.Market) error {
	ex, ok := catalog.Exchanges[m.Exchange]
	fn, ok2 := sources.Exchanges[m.Exchange]
	if !ok || !ok2 {
		return fmt.Errorf("unknown exchange %q", m.Exchange)
	}
	pair := m.Symbol
	if m.Inverse {
		pair += " (1/x)"
	}
	info := model.SourceInfo{ID: ex.ID, Name: ex.Name, Type: "exchange", Fee: ex.Fee, URL: ex.URL, Pair: pair}
	s.sources = append(s.sources, &source{
		info:     info,
		interval: s.cfg.Interval,
		fetch: func(ctx context.Context) ([]model.Quote, error) {
			t, err := fn(ctx, m.Symbol)
			if err != nil {
				return nil, err
			}
			if m.Inverse {
				t = t.Invert()
			}
			return []model.Quote{{
				Source: ex.ID, Vs: "USD", Pair: pair, Native: m.Native,
				Bid: t.Bid, Ask: t.Ask, Last: t.Last, Volume: t.Volume, Change24h: t.Change24h,
				Fee: ex.Fee, Time: t.Time,
			}}, nil
		},
	})
	return nil
}

func (s *Service) addFX(fx sources.FXSource) {
	code := s.cfg.Asset.Code
	vs := s.info.Quotes
	s.sources = append(s.sources, &source{
		info:     model.SourceInfo{ID: fx.ID, Name: fx.Name, Type: fx.Type, URL: fx.URL},
		interval: fx.Interval,
		fetch: func(ctx context.Context) ([]model.Quote, error) {
			t, err := fx.Fetch(ctx, code, catalog.FiatCodes)
			if err != nil {
				return nil, err
			}
			ts := t.Time
			if ts == 0 {
				ts = time.Now().UnixMilli()
			}
			var out []model.Quote
			for _, q := range vs {
				if r, ok := t.Cross(code, q); ok {
					out = append(out, model.Quote{Source: fx.ID, Vs: q, Pair: code + "/" + q, Last: r, Rate: r, Time: ts})
				}
			}
			if len(out) == 0 {
				return nil, fmt.Errorf("no %s rates in the response", code)
			}
			return out, nil
		},
	})
}

// Run starts polling, backfill and the HTTP API; it blocks until ctx ends.
func (s *Service) Run(ctx context.Context) error {
	ln, err := net.Listen("tcp", s.cfg.Addr)
	if err != nil {
		return err
	}
	st, err := storage.Open(filepath.Join(s.cfg.DataDir, strings.ToLower(s.cfg.Asset.Code)+".db"))
	if err != nil {
		ln.Close()
		return err
	}
	s.store = st
	defer st.Close()

	var wg sync.WaitGroup
	run := func(f func()) {
		wg.Add(1)
		go func() { defer wg.Done(); f() }()
	}
	for _, src := range s.sources {
		src := src
		run(func() { s.pollSource(ctx, src) })
	}
	if s.cfg.Asset.Kind == model.Crypto && s.cfg.Asset.Code != "USDT" && s.cfg.USDTURL != "" {
		run(func() { s.followUSDT(ctx) })
	}
	run(func() { s.snapshotLoop(ctx) })
	run(func() { s.maintenance(ctx) })

	srv := &http.Server{Handler: s.routes(), ReadHeaderTimeout: 10 * time.Second}
	go func() {
		<-ctx.Done()
		sctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = srv.Shutdown(sctx)
	}()
	s.log.Printf("listening on http://%s with %d sources", ln.Addr(), len(s.sources))
	err = srv.Serve(ln)
	wg.Wait()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func sleepCtx(ctx context.Context, d time.Duration) bool {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-t.C:
		return true
	}
}

func (s *Service) pollSource(ctx context.Context, src *source) {
	// Spread the first requests so all services don't hit the network at once.
	if !sleepCtx(ctx, time.Duration(rand.IntN(1500))*time.Millisecond) {
		return
	}
	for {
		start := time.Now()
		cctx, cancel := context.WithTimeout(ctx, 12*time.Second)
		quotes, err := src.fetch(cctx)
		cancel()
		if ctx.Err() != nil {
			return
		}
		latency := time.Since(start).Milliseconds()

		s.mu.Lock()
		st := s.state[src.info.ID]
		st.latency = latency
		if err == nil {
			if st.err != "" {
				s.log.Printf("%s recovered", src.info.Name)
			}
			st.quotes, st.err, st.fails = quotes, "", 0
			st.lastOK = time.Now().UnixMilli()
		} else {
			st.fails++
			if st.err == "" { // log transitions only, not every retry
				s.log.Printf("%s: %v", src.info.Name, err)
			}
			st.err = err.Error()
		}
		fails := st.fails
		s.mu.Unlock()

		wait := src.interval
		if fails > 0 {
			backoff := 5 * time.Second << min(fails-1, 6)
			if backoff > 5*time.Minute {
				backoff = 5 * time.Minute
			}
			if backoff < wait || wait < 5*time.Second {
				wait = backoff
			}
		}
		// Jitter avoids synchronised bursts across services.
		wait += time.Duration(rand.IntN(400)) * time.Millisecond
		if !sleepCtx(ctx, wait) {
			return
		}
	}
}

// followUSDT keeps the USDT/USD rate used to normalise USDT-quoted pairs,
// asking the USDT microservice.
func (s *Service) followUSDT(ctx context.Context) {
	for {
		var snap model.Snapshot
		cctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		err := httpx.GetJSON(cctx, strings.TrimRight(s.cfg.USDTURL, "/")+"/api/v1/snapshot", &snap)
		cancel()
		if err == nil {
			if m := snap.Markets["USD"]; m != nil && m.Price > 0.9 && m.Price < 1.1 {
				s.usdt.Store(math.Float64bits(m.Price))
			}
		}
		if !sleepCtx(ctx, 15*time.Second) {
			return
		}
	}
}

func (s *Service) usdtRate() float64 { return math.Float64frombits(s.usdt.Load()) }

func (s *Service) factor(native string) float64 {
	if native == "USDT" {
		return s.usdtRate()
	}
	return 1 // USD, and USDC treated as 1 USD
}

func staleAfter(src *source) time.Duration {
	d := 3 * src.interval
	if d < 30*time.Second {
		d = 30 * time.Second
	}
	return d
}

func (s *Service) snapshotLoop(ctx context.Context) {
	every := s.cfg.Interval
	if s.cfg.Asset.Kind == model.Fiat {
		every = 5 * time.Second
	}
	if !sleepCtx(ctx, 2*time.Second) {
		return
	}
	t := time.NewTicker(every)
	defer t.Stop()
	for {
		s.buildSnapshot()
		select {
		case <-ctx.Done():
			return
		case <-t.C:
		}
	}
}

func (s *Service) buildSnapshot() {
	now := time.Now()
	nowMs := now.UnixMilli()
	isCrypto := s.cfg.Asset.Kind == model.Crypto

	s.mu.RLock()
	var quotes []model.Quote
	for _, src := range s.sources {
		st := s.state[src.info.ID]
		if st.quotes == nil {
			if st.err != "" {
				vs := "*"
				if isCrypto {
					vs = "USD"
				}
				quotes = append(quotes, model.Quote{Source: src.info.ID, Vs: vs, Pair: src.info.Pair,
					Fee: src.info.Fee, OK: false, Error: st.err, Latency: st.latency})
			}
			continue
		}
		stale := nowMs-st.lastOK > staleAfter(src).Milliseconds()
		for _, q := range st.quotes {
			q.Fetched, q.Latency, q.OK = st.lastOK, st.latency, true
			if stale {
				q.OK = false
				q.Error = "stale data"
				if st.err != "" {
					q.Error = st.err
				}
			}
			if isCrypto {
				f := s.factor(q.Native)
				q.Bid, q.Ask, q.Last = q.Bid*f, q.Ask*f, q.Last*f
			}
			if q.Bid > 0 && q.Ask > 0 {
				q.Rate = (q.Bid + q.Ask) / 2
			} else if q.Rate == 0 {
				q.Rate = q.Last
			}
			quotes = append(quotes, q)
		}
	}
	s.mu.RUnlock()

	snap := &model.Snapshot{Code: s.cfg.Asset.Code, Kind: s.cfg.Asset.Kind, Time: nowMs, Markets: map[string]*model.Market{}}
	if isCrypto {
		snap.USDTRate = s.usdtRate()
	}
	var samples []storage.Sample
	for _, vs := range s.info.Quotes {
		price, online := market.Aggregate(quotes, vs)
		if price <= 0 {
			continue
		}
		m := &model.Market{Vs: vs, Price: price, Online: online}
		m.Best = market.ComputeBest(quotes, vs, nil, nil)
		if ref := s.reference(vs, now); ref > 0 {
			m.Change24h = market.Change(price, ref)
		}
		snap.Markets[vs] = m
		samples = s.sample(samples, backfill.AggSeries(vs), nowMs, price)
	}
	for _, q := range quotes {
		if q.OK && q.Rate > 0 {
			samples = s.sample(samples, backfill.SourceSeries(q.Source, q.Vs), nowMs, q.Rate)
		}
	}
	snap.Quotes = quotes
	if err := s.store.Append(samples); err != nil {
		s.log.Printf("storage: %v", err)
	}

	s.mu.Lock()
	s.seq++
	snap.Seq = s.seq
	s.snap = snap
	s.mu.Unlock()

	if data, err := json.Marshal(snap); err == nil {
		s.broker.Publish("snapshot", data)
	}
}

// sample stores at most one point per 10s per series, and skips unchanged
// values for up to 5 minutes (fiat feeds change rarely).
func (s *Service) sample(out []storage.Sample, series string, t int64, v float64) []storage.Sample {
	last, ok := s.stored[series]
	if ok {
		dt := t - last.T
		if dt < 10_000 || (v == last.V && dt < 300_000) {
			return out
		}
	}
	s.stored[series] = model.Point{T: t, V: v}
	return append(out, storage.Sample{Series: series, T: t, V: v})
}

// reference returns the aggregated price ~24h ago, cached for a minute.
func (s *Service) reference(vs string, now time.Time) float64 {
	if s.resetRefs.Swap(false) {
		s.refs = map[string]refEntry{}
	}
	if r, ok := s.refs[vs]; ok && now.Sub(r.checked) < time.Minute {
		return r.value
	}
	target := now.Add(-24 * time.Hour)
	tolerance := 3 * time.Hour
	if s.cfg.Asset.Kind == model.Fiat {
		tolerance = 72 * time.Hour // daily reference data, weekends
	}
	var v float64
	if p, ok := s.store.ValueAt(backfill.AggSeries(vs), target.UnixMilli()); ok && target.UnixMilli()-p.T <= tolerance.Milliseconds() {
		v = p.V
	}
	s.refs[vs] = refEntry{checked: now, value: v}
	return v
}

// maintenance backfills history on start (unless it was done within the
// last hour) and then every 6 hours, and prunes expired samples hourly.
func (s *Service) maintenance(ctx context.Context) {
	refresh := time.Hour
	nextPrune := time.Now().Add(time.Minute)
	for {
		if !s.cfg.NoBackfill {
			last := time.UnixMilli(s.store.GetMeta("backfill"))
			if time.Since(last) > refresh {
				s.runBackfill(ctx)
			}
			refresh = 6 * time.Hour
		}
		if time.Now().After(nextPrune) {
			if err := s.store.Prune(time.Now()); err != nil {
				s.log.Printf("prune: %v", err)
			}
			nextPrune = time.Now().Add(time.Hour)
		}
		if !sleepCtx(ctx, 10*time.Minute) {
			return
		}
	}
}

func (s *Service) runBackfill(ctx context.Context) {
	for attempt := 1; attempt <= 3; attempt++ {
		cctx, cancel := context.WithTimeout(ctx, 90*time.Second)
		var err error
		if s.cfg.Asset.Kind == model.Crypto {
			err = backfill.Crypto(cctx, s.cfg.Asset, s.store)
		} else {
			err = backfill.Fiat(cctx, s.cfg.Asset.Code, s.store)
		}
		cancel()
		if ctx.Err() != nil {
			return
		}
		if err == nil {
			s.log.Printf("history backfill done")
			_ = s.store.SetMeta("backfill", time.Now().UnixMilli())
			s.resetRefs.Store(true)
			return
		}
		s.log.Printf("history backfill (attempt %d): %v", attempt, err)
		if !sleepCtx(ctx, time.Duration(attempt)*20*time.Second) {
			return
		}
	}
}

// ---- HTTP API ------------------------------------------------------------------

func (s *Service) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]any{"status": "ok", "code": s.cfg.Asset.Code, "uptime": int(time.Since(s.started).Seconds()), "version": s.cfg.Version})
	})
	mux.HandleFunc("GET /api/v1/info", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, s.info)
	})
	mux.HandleFunc("GET /api/v1/snapshot", func(w http.ResponseWriter, r *http.Request) {
		s.mu.RLock()
		snap := s.snap
		s.mu.RUnlock()
		if snap == nil {
			http.Error(w, "warming up", http.StatusServiceUnavailable)
			return
		}
		writeJSON(w, snap)
	})
	mux.HandleFunc("GET /api/v1/history", s.handleHistory)
	mux.HandleFunc("GET /api/v1/value-at", s.handleValueAt)
	mux.Handle("GET /api/v1/stream", s.broker)
	return mux
}

func (s *Service) vsParam(r *http.Request) (string, error) {
	vs := strings.ToUpper(r.URL.Query().Get("vs"))
	if vs == "" {
		return s.info.Quotes[0], nil
	}
	for _, q := range s.info.Quotes {
		if q == vs {
			return vs, nil
		}
	}
	return "", fmt.Errorf("%s is not quoted in %s (supported: %s)", s.cfg.Asset.Code, vs, strings.Join(s.info.Quotes, ","))
}

func (s *Service) handleHistory(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	vs, err := s.vsParam(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	now := time.Now().UnixMilli()
	to := parseInt(q.Get("to"), now)
	from := parseInt(q.Get("from"), to-24*3600*1000)
	points := int(parseInt(q.Get("points"), 500))
	points = max(10, min(points, 5000))
	ids := strings.Split(q.Get("sources"), ",")
	if q.Get("sources") == "" {
		ids = []string{"agg"}
	}
	h := model.History{Code: s.cfg.Asset.Code, Vs: vs, From: from, To: to, Series: map[string][]model.Point{}}
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		series := backfill.AggSeries(vs)
		if id != "agg" {
			series = backfill.SourceSeries(id, vs)
		}
		pts, err := s.store.Query(series, from, to, points)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if pts == nil {
			pts = []model.Point{}
		}
		h.Series[id] = pts
	}
	writeJSON(w, h)
}

func (s *Service) handleValueAt(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	vs, err := s.vsParam(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	series := backfill.AggSeries(vs)
	if src := q.Get("source"); src != "" && src != "agg" {
		series = backfill.SourceSeries(src, vs)
	}
	p, ok := s.store.ValueAt(series, parseInt(q.Get("t"), time.Now().UnixMilli()))
	writeJSON(w, map[string]any{"ok": ok, "t": p.T, "v": p.V})
}

func parseInt(s string, def int64) int64 {
	if v, err := strconv.ParseInt(s, 10, 64); err == nil {
		return v
	}
	return def
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	_ = json.NewEncoder(w).Encode(v)
}
