// Package gateway is the backend-for-frontend: it follows every currency
// microservice over SSE, converts prices into the user's display currency,
// evaluates alerts, sends desktop notifications and serves the web UI.
package gateway

import (
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"cryptomon/internal/alerts"
	"cryptomon/internal/catalog"
	"cryptomon/internal/httpx"
	"cryptomon/internal/model"
	"cryptomon/internal/notify"
	"cryptomon/internal/sse"
	"cryptomon/web"
)

// ErrShutdown is returned by Run when the user pressed "Quit" in the UI.
var ErrShutdown = errors.New("shutdown requested from the UI")

type ServiceRef struct {
	Code string `json:"code"`
	URL  string `json:"url"`
}

type Config struct {
	Addr     string
	DataDir  string
	Services []ServiceRef
	Version  string
}

type svcState struct {
	ref       ServiceRef
	info      *model.AssetInfo
	snap      *model.Snapshot
	connected bool
	lastSeen  time.Time
	err       string
}

// ServiceStatus is reported on the settings page.
type ServiceStatus struct {
	Code     string `json:"code"`
	URL      string `json:"url"`
	Status   string `json:"status"`
	LastSeen int64  `json:"lastSeen"`
	Error    string `json:"error,omitempty"`
	Sources  int    `json:"sources"`
	Online   int    `json:"online"`
	Version  string `json:"version,omitempty"`
}

type clientMsg struct {
	event string
	code  string // for "view": asset to render in the client's currency
	data  []byte
}

type client struct {
	vs string
	ch chan clientMsg
}

type Gateway struct {
	cfg      Config
	log      *log.Logger
	settings *settingsStore
	alerts   *alerts.Engine
	notifier *notify.Notifier

	mu      sync.RWMutex
	svcs    map[string]*svcState
	order   []string
	clients map[*client]struct{}

	evalCh   chan string
	shutdown chan struct{}
	once     sync.Once
}

func New(cfg Config, logger *log.Logger) (*Gateway, error) {
	if err := os.MkdirAll(cfg.DataDir, 0o755); err != nil {
		return nil, err
	}
	st, err := loadSettings(filepath.Join(cfg.DataDir, "settings.json"))
	if err != nil {
		return nil, err
	}
	eng, err := alerts.Load(filepath.Join(cfg.DataDir, "alerts.json"))
	if err != nil {
		return nil, err
	}
	g := &Gateway{
		cfg: cfg, log: logger, settings: st, alerts: eng,
		notifier: notify.New(cfg.DataDir),
		svcs:     map[string]*svcState{},
		clients:  map[*client]struct{}{},
		evalCh:   make(chan string, 64),
		shutdown: make(chan struct{}),
	}
	for _, ref := range cfg.Services {
		g.svcs[ref.Code] = &svcState{ref: ref}
	}
	// Keep catalog order for the UI regardless of flag order.
	for _, a := range catalog.Assets {
		if _, ok := g.svcs[a.Code]; ok {
			g.order = append(g.order, a.Code)
		}
	}
	eng.Price = g.currentPrice
	eng.RefAt = g.referencePrice
	eng.OnEvent = g.onAlert
	return g, nil
}

func (g *Gateway) Run(ctx context.Context) error {
	ln, err := net.Listen("tcp", g.cfg.Addr)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	for _, code := range g.order {
		go g.follow(ctx, g.svcs[code])
	}
	go g.alertLoop(ctx)

	srv := &http.Server{Handler: g.guard(g.routes()), ReadHeaderTimeout: 10 * time.Second}
	var requested atomic.Bool
	go func() {
		select {
		case <-ctx.Done():
		case <-g.shutdown:
			requested.Store(true)
		}
		sctx, c := context.WithTimeout(context.Background(), 3*time.Second)
		defer c()
		_ = srv.Shutdown(sctx)
	}()
	g.log.Printf("UI and API on http://%s", ln.Addr())
	err = srv.Serve(ln)
	if requested.Load() {
		return ErrShutdown
	}
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

// follow keeps one service's info and live snapshots up to date.
func (g *Gateway) follow(ctx context.Context, st *svcState) {
	code := st.ref.Code
	for ctx.Err() == nil {
		var info model.AssetInfo
		cctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		err := httpx.GetJSON(cctx, st.ref.URL+"/api/v1/info", &info)
		cancel()
		if err == nil {
			g.mu.Lock()
			st.info = &info
			g.mu.Unlock()
			break
		}
		g.setState(st, false, err)
		select {
		case <-ctx.Done():
			return
		case <-time.After(2 * time.Second):
		}
	}
	sse.Subscribe(ctx, st.ref.URL+"/api/v1/stream",
		func(connected bool, err error) { g.setState(st, connected, err) },
		func(event string, data []byte) {
			if event != "snapshot" {
				return
			}
			var snap model.Snapshot
			if json.Unmarshal(data, &snap) != nil {
				return
			}
			g.mu.Lock()
			st.snap, st.lastSeen, st.connected, st.err = &snap, time.Now(), true, ""
			g.mu.Unlock()
			g.broadcast(clientMsg{event: "view", code: code})
			select {
			case g.evalCh <- code:
			default:
			}
		})
}

func (g *Gateway) setState(st *svcState, connected bool, err error) {
	g.mu.Lock()
	wasSeen := !st.lastSeen.IsZero()
	changed := st.connected != connected
	st.connected = connected
	if err != nil {
		changed = changed || st.err == ""
		st.err = err.Error()
	} else {
		st.err = ""
	}
	g.mu.Unlock()
	if changed {
		if !connected && wasSeen {
			g.log.Printf("%s service unavailable: %v", st.ref.Code, err)
		}
		g.broadcastStatus()
		g.broadcast(clientMsg{event: "view", code: st.ref.Code})
	}
}

func (g *Gateway) services() []ServiceStatus {
	g.mu.RLock()
	defer g.mu.RUnlock()
	out := make([]ServiceStatus, 0, len(g.order))
	for _, code := range g.order {
		st := g.svcs[code]
		s := ServiceStatus{Code: code, URL: st.ref.URL, Status: "waiting", Error: st.err}
		if st.connected {
			s.Status = "up"
		} else if st.err != "" {
			s.Status = "down"
		}
		if !st.lastSeen.IsZero() {
			s.LastSeen = st.lastSeen.UnixMilli()
		}
		if st.info != nil {
			s.Sources, s.Version = len(st.info.Sources), st.info.Version
		}
		if st.snap != nil {
			seen := map[string]bool{}
			for _, q := range st.snap.Quotes {
				if q.OK && !seen[q.Source] {
					seen[q.Source] = true
					s.Online++
				}
			}
		}
		out = append(out, s)
	}
	return out
}

func (g *Gateway) broadcast(m clientMsg) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	for c := range g.clients {
		select {
		case c.ch <- m:
		default: // slow client: it will get the next update
		}
	}
}

func (g *Gateway) broadcastStatus() {
	if b, err := json.Marshal(g.services()); err == nil {
		g.broadcast(clientMsg{event: "status", data: b})
	}
}

func (g *Gateway) alertLoop(ctx context.Context) {
	tick := time.NewTicker(30 * time.Second)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case code := <-g.evalCh:
			g.alerts.Evaluate(code, time.Now())
		case <-tick.C:
			if b, err := json.Marshal(g.alerts.Rules()); err == nil {
				g.broadcast(clientMsg{event: "rules", data: b})
			}
		}
	}
}

func (g *Gateway) sourceName(asset, id string) string {
	if id == "" || id == "agg" {
		return ""
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	if st := g.svcs[asset]; st != nil && st.info != nil {
		for _, s := range st.info.Sources {
			if s.ID == id {
				return s.Name
			}
		}
	}
	return id
}

func (g *Gateway) onAlert(ev alerts.Event) {
	s := g.settings.Get()
	g.log.Printf("alert: %s/%s %s %.4g (price %.6g)", ev.Asset, ev.Vs, ev.Condition, ev.Value, ev.Price)
	if s.OSNotify {
		title, body := notify.AlertText(s.Lang, ev, g.sourceName(ev.Asset, ev.Source))
		go func() {
			if err := g.notifier.Send(title, body); err != nil {
				g.log.Printf("desktop notification failed: %v", err)
			}
		}()
	}
	if b, err := json.Marshal(ev); err == nil {
		g.broadcast(clientMsg{event: "alert", data: b})
	}
	if b, err := json.Marshal(g.alerts.Rules()); err == nil {
		g.broadcast(clientMsg{event: "rules", data: b})
	}
}

// ---- HTTP ----------------------------------------------------------------------

// guard protects the local API from other web pages: mutating requests
// need a custom header (which forces a CORS preflight we never allow), and
// on loopback the Host header must be local (DNS rebinding).
func (g *Gateway) guard(next http.Handler) http.Handler {
	loopback := strings.HasPrefix(g.cfg.Addr, "127.") || strings.HasPrefix(g.cfg.Addr, "localhost") || strings.HasPrefix(g.cfg.Addr, "[::1]")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if loopback {
			host := r.Host
			if h, _, err := net.SplitHostPort(host); err == nil {
				host = h
			}
			if host != "127.0.0.1" && host != "localhost" && host != "::1" && host != "[::1]" {
				http.Error(w, "forbidden host", http.StatusForbidden)
				return
			}
		}
		if r.Method != http.MethodGet && r.Method != http.MethodHead && r.Header.Get("X-Cryptomon") != "1" {
			http.Error(w, "missing X-Cryptomon header", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (g *Gateway) routes() http.Handler {
	mux := http.NewServeMux()
	static, _ := fs.Sub(web.Static, "static")
	if dir := os.Getenv("CRYPTOMON_WEB_DIR"); dir != "" { // UI development: serve files from disk
		static = os.DirFS(dir)
	}
	files := http.FileServer(http.FS(static))
	mux.Handle("GET /", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache")
		files.ServeHTTP(w, r)
	}))

	mux.HandleFunc("GET /api/meta", g.handleMeta)
	mux.HandleFunc("GET /api/views", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, g.views(vsOf(r)))
	})
	mux.HandleFunc("GET /api/views/{code}", func(w http.ResponseWriter, r *http.Request) {
		code := strings.ToUpper(r.PathValue("code"))
		if _, ok := g.svcs[code]; !ok {
			http.NotFound(w, r)
			return
		}
		writeJSON(w, g.view(code, vsOf(r)))
	})
	mux.HandleFunc("GET /api/history/{code}", g.handleHistory)
	mux.HandleFunc("GET /api/stream", g.handleStream)
	mux.HandleFunc("GET /api/services", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, g.services())
	})

	mux.HandleFunc("GET /api/settings", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, g.settings.Get())
	})
	mux.HandleFunc("PUT /api/settings", func(w http.ResponseWriter, r *http.Request) {
		var s Settings
		if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
			httpError(w, err, http.StatusBadRequest)
			return
		}
		saved, err := g.settings.Set(s)
		if err != nil {
			httpError(w, err, http.StatusInternalServerError)
			return
		}
		writeJSON(w, saved)
		// Fees and excluded sources change the best-rate view.
		for _, code := range g.order {
			g.broadcast(clientMsg{event: "view", code: code})
		}
	})

	mux.HandleFunc("GET /api/alerts", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, g.alerts.Rules())
	})
	mux.HandleFunc("POST /api/alerts", func(w http.ResponseWriter, r *http.Request) {
		var rule alerts.Rule
		if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
			httpError(w, err, http.StatusBadRequest)
			return
		}
		saved, err := g.alerts.Add(rule)
		if err != nil {
			httpError(w, err, http.StatusBadRequest)
			return
		}
		writeJSON(w, saved)
		g.queueEval(saved.Asset)
	})
	mux.HandleFunc("PUT /api/alerts/{id}", func(w http.ResponseWriter, r *http.Request) {
		var rule alerts.Rule
		if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
			httpError(w, err, http.StatusBadRequest)
			return
		}
		saved, err := g.alerts.Update(r.PathValue("id"), rule)
		if errors.Is(err, os.ErrNotExist) {
			http.NotFound(w, r)
			return
		}
		if err != nil {
			httpError(w, err, http.StatusBadRequest)
			return
		}
		writeJSON(w, saved)
		g.queueEval(saved.Asset)
	})
	mux.HandleFunc("DELETE /api/alerts/{id}", func(w http.ResponseWriter, r *http.Request) {
		if err := g.alerts.Delete(r.PathValue("id")); err != nil {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("GET /api/alerts/events", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, g.alerts.Events())
	})
	mux.HandleFunc("DELETE /api/alerts/events", func(w http.ResponseWriter, r *http.Request) {
		if err := g.alerts.ClearEvents(); err != nil {
			httpError(w, err, http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("POST /api/notify/test", func(w http.ResponseWriter, r *http.Request) {
		s := g.settings.Get()
		if err := g.notifier.Send("Cryptomon", notify.TestText(s.Lang)); err != nil {
			httpError(w, err, http.StatusInternalServerError)
			return
		}
		writeJSON(w, map[string]bool{"ok": true})
	})
	mux.HandleFunc("POST /api/shutdown", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]bool{"ok": true})
		g.once.Do(func() {
			go func() {
				time.Sleep(300 * time.Millisecond)
				close(g.shutdown)
			}()
		})
	})
	return mux
}

func (g *Gateway) queueEval(code string) {
	select {
	case g.evalCh <- code:
	default:
	}
}

func vsOf(r *http.Request) string {
	vs := strings.ToUpper(r.URL.Query().Get("vs"))
	if !catalog.IsFiat(vs) {
		return "USD"
	}
	return vs
}

type metaAsset struct {
	model.AssetInfo
	Status string `json:"status"`
}

func (g *Gateway) handleMeta(w http.ResponseWriter, r *http.Request) {
	g.mu.RLock()
	assets := make([]metaAsset, 0, len(g.order))
	for _, code := range g.order {
		st := g.svcs[code]
		a, _ := catalog.Find(code)
		info := model.AssetInfo{Code: a.Code, Name: a.Name, Kind: a.Kind, Symbol: a.Symbol, Color: a.Color, Sources: []model.SourceInfo{}}
		if st.info != nil {
			info = *st.info
		}
		status := "waiting"
		if st.connected {
			status = "up"
		} else if st.err != "" {
			status = "down"
		}
		assets = append(assets, metaAsset{AssetInfo: info, Status: status})
	}
	g.mu.RUnlock()
	ranges := []string{"1h", "6h", "24h", "7d", "30d", "90d", "1y", "5y"}
	writeJSON(w, map[string]any{
		"version": g.cfg.Version, "fiat": catalog.FiatCodes, "assets": assets, "ranges": ranges,
	})
}

func (g *Gateway) handleHistory(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	code := strings.ToUpper(r.PathValue("code"))
	now := time.Now().UnixMilli()
	span, ok := ranges[q.Get("range")]
	if !ok {
		span = 24 * time.Hour
	}
	to := now
	from := now - span.Milliseconds()
	if v, err := strconv.ParseInt(q.Get("from"), 10, 64); err == nil {
		from = v
	}
	points, _ := strconv.Atoi(q.Get("points"))
	if points <= 0 {
		points = 600
	}
	sources := []string{"agg"}
	if s := q.Get("sources"); s != "" {
		sources = strings.Split(s, ",")
	}
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()
	h, err := g.history(ctx, code, vsOf(r), from, to, sources, points)
	if err != nil {
		httpError(w, err, http.StatusBadGateway)
		return
	}
	writeJSON(w, h)
}

func (g *Gateway) handleStream(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}
	sse.PrepareHeaders(w)
	c := &client{vs: vsOf(r), ch: make(chan clientMsg, 256)}
	g.mu.Lock()
	g.clients[c] = struct{}{}
	g.mu.Unlock()
	defer func() {
		g.mu.Lock()
		delete(g.clients, c)
		g.mu.Unlock()
	}()

	send := func(event string, v any) bool {
		b, err := json.Marshal(v)
		if err != nil {
			return true
		}
		if _, err := w.Write(sse.Format(event, b)); err != nil {
			return false
		}
		return true
	}
	send("views", g.views(c.vs))
	send("status", g.services())
	send("rules", g.alerts.Rules())
	flusher.Flush()

	heartbeat := time.NewTicker(15 * time.Second)
	defer heartbeat.Stop()
	for {
		select {
		case <-r.Context().Done():
			return
		case <-g.shutdown:
			send("bye", map[string]bool{"ok": true})
			flusher.Flush()
			return
		case m := <-c.ch:
			ok := true
			if m.event == "view" {
				ok = send("view", g.view(m.code, c.vs))
			} else if _, err := w.Write(sse.Format(m.event, m.data)); err != nil {
				ok = false
			}
			if !ok {
				return
			}
			flusher.Flush()
		case <-heartbeat.C:
			if _, err := w.Write([]byte(": ping\n\n")); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	_ = json.NewEncoder(w).Encode(v)
}

func httpError(w http.ResponseWriter, err error, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
}
