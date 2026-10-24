package gateway

import (
	"cryptomon/internal/catalog"
	"cryptomon/internal/market"
	"cryptomon/internal/model"
)

// SourceView is one row of the exchanges table, already converted to the
// display currency and with the user's fee settings applied.
type SourceView struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Type      string   `json:"type"`
	URL       string   `json:"url"`
	Pair      string   `json:"pair,omitempty"`
	Native    string   `json:"native,omitempty"`
	Bid       float64  `json:"bid"`
	Ask       float64  `json:"ask"`
	Rate      float64  `json:"rate"`
	Volume    float64  `json:"volume,omitempty"`
	Change24h *float64 `json:"change24h,omitempty"`
	Fee       float64  `json:"fee"`
	FeeCustom bool     `json:"feeCustom,omitempty"`
	EffBuy    float64  `json:"effBuy"`
	EffSell   float64  `json:"effSell"`
	SpreadPct float64  `json:"spreadPct"`
	Time      int64    `json:"time"`
	Fetched   int64    `json:"fetched"`
	Latency   int64    `json:"latency"`
	State     string   `json:"state"` // ok, error, waiting, excluded, na (pair not offered)
	Error     string   `json:"error,omitempty"`
	BestBuy   bool     `json:"bestBuy,omitempty"`
	BestSell  bool     `json:"bestSell,omitempty"`
}

// AssetView is everything the UI shows for one currency.
type AssetView struct {
	Code      string       `json:"code"`
	Name      string       `json:"name"`
	Kind      model.Kind   `json:"kind"`
	Symbol    string       `json:"symbol"`
	Color     string       `json:"color"`
	Vs        string       `json:"vs"`
	Time      int64        `json:"time"`
	Price     float64      `json:"price"`
	Change24h *float64     `json:"change24h,omitempty"`
	Best      model.Best   `json:"best"`
	Sources   []SourceView `json:"sources"`
	Online    int          `json:"online"`
	Total     int          `json:"total"`
	Status    string       `json:"status"` // up, down, waiting
	Error     string       `json:"error,omitempty"`
	USDTRate  float64      `json:"usdtRate,omitempty"`
}

// effectiveVs avoids pricing a currency in itself (USD in USD shows EUR).
func effectiveVs(code, vs string) string {
	if !catalog.IsFiat(vs) {
		vs = "USD"
	}
	if code == vs {
		if code == "USD" {
			return "EUR"
		}
		return "USD"
	}
	return vs
}

func combine(a, b *float64) *float64 {
	if a == nil {
		return nil
	}
	if b == nil {
		return a
	}
	v := ((1+*a/100)*(1+*b/100) - 1) * 100
	return &v
}

// quotesIn returns the asset's quotes expressed in vs, plus the factor and
// the FX change used for crypto conversion. Caller holds g.mu (read).
func (g *Gateway) quotesIn(st *svcState, vs string) (quotes []model.Quote, m *model.Market, effVs string, factor float64, fxChange *float64) {
	snap := st.snap
	factor = 1
	effVs = effectiveVs(st.ref.Code, vs)
	if snap.Kind == model.Fiat {
		m = snap.Markets[effVs]
		for _, q := range snap.Quotes {
			if q.Vs == effVs || q.Vs == "*" {
				quotes = append(quotes, q)
			}
		}
		return
	}
	m = snap.Markets["USD"]
	if effVs != "USD" {
		usd := g.svcs["USD"]
		var fm *model.Market
		if usd != nil && usd.snap != nil {
			fm = usd.snap.Markets[effVs]
		}
		if fm == nil || fm.Price <= 0 {
			effVs = "USD" // no FX rate yet: fall back to dollars
		} else {
			factor, fxChange = fm.Price, fm.Change24h
		}
	}
	for _, q := range snap.Quotes {
		if q.Vs != "USD" {
			continue
		}
		q.Vs = effVs
		q.Bid, q.Ask, q.Last, q.Rate = q.Bid*factor, q.Ask*factor, q.Last*factor, q.Rate*factor
		q.Change24h = combine(q.Change24h, fxChange)
		quotes = append(quotes, q)
	}
	return
}

// view builds the AssetView of code in vs. Caller must not hold g.mu.
func (g *Gateway) view(code, vs string) AssetView {
	g.mu.RLock()
	defer g.mu.RUnlock()
	settings := g.settings.Get()
	a, _ := catalog.Find(code)
	v := AssetView{Code: a.Code, Name: a.Name, Kind: a.Kind, Symbol: a.Symbol, Color: a.Color,
		Vs: effectiveVs(code, vs), Status: "waiting", Sources: []SourceView{}}
	st := g.svcs[code]
	if st == nil {
		v.Status, v.Error = "down", "service not configured"
		return v
	}
	if !st.connected && st.err != "" {
		v.Status, v.Error = "down", st.err
	}
	if st.info != nil {
		v.Total = len(st.info.Sources)
	}
	if st.snap == nil || st.info == nil {
		return v
	}
	if st.connected {
		v.Status = "up"
	}
	quotes, m, effVs, factor, fxChange := g.quotesIn(st, vs)
	v.Vs, v.Time, v.USDTRate = effVs, st.snap.Time, st.snap.USDTRate
	if m != nil {
		v.Price = m.Price * factor
		v.Change24h = combine(m.Change24h, fxChange)
	}

	feeOf := func(q model.Quote) float64 {
		if f, ok := settings.Fees[q.Source]; ok {
			return f
		}
		return q.Fee
	}
	include := func(id string) bool { return !settings.disabled(id) }
	v.Best = market.ComputeBest(quotes, effVs, feeOf, include)

	byID := map[string]model.Quote{}
	for _, q := range quotes {
		byID[q.Source] = q
	}
	// Sources that answer but do not price this pair (e.g. the ECB has no RUB).
	alive := map[string]bool{}
	for _, q := range st.snap.Quotes {
		alive[q.Source] = true
	}
	for _, info := range st.info.Sources {
		sv := SourceView{ID: info.ID, Name: info.Name, Type: info.Type, URL: info.URL, Pair: info.Pair, Fee: info.Fee, State: "waiting"}
		if f, ok := settings.Fees[info.ID]; ok {
			sv.Fee, sv.FeeCustom = f, true
		}
		q, ok := byID[info.ID]
		if !ok && alive[info.ID] {
			sv.State = "na"
		}
		if ok {
			sv.Native, sv.Bid, sv.Ask, sv.Rate, sv.Volume = q.Native, q.Bid, q.Ask, q.Rate, q.Volume
			sv.Change24h, sv.Time, sv.Fetched, sv.Latency = q.Change24h, q.Time, q.Fetched, q.Latency
			if q.Pair != "" {
				sv.Pair = q.Pair
			}
			if q.OK && q.Rate > 0 {
				sv.State = "ok"
				v.Online++
				sv.EffBuy = market.BuyPrice(q) * (1 + sv.Fee)
				sv.EffSell = market.SellPrice(q) * (1 - sv.Fee)
				if q.Bid > 0 && q.Ask > 0 {
					sv.SpreadPct = (q.Ask - q.Bid) / q.Rate * 100
				}
				if !include(info.ID) {
					sv.State = "excluded"
				}
			} else {
				sv.State, sv.Error = "error", q.Error
			}
		}
		if v.Best.Buy != nil && v.Best.Buy.Source == info.ID {
			sv.BestBuy = true
		}
		if v.Best.Sell != nil && v.Best.Sell.Source == info.ID {
			sv.BestSell = true
		}
		v.Sources = append(v.Sources, sv)
	}
	return v
}

// views returns every asset in catalog order.
func (g *Gateway) views(vs string) []AssetView {
	out := make([]AssetView, 0, len(g.order))
	for _, code := range g.order {
		out = append(out, g.view(code, vs))
	}
	return out
}

// currentPrice is used by the alert engine: the aggregated price, or one
// source's mid price, in vs.
func (g *Gateway) currentPrice(asset, vs, source string) (float64, bool) {
	v := g.view(asset, vs)
	if v.Vs != effectiveVs(asset, vs) || v.Vs != vs {
		return 0, false
	}
	if source == "" || source == "agg" {
		return v.Price, v.Price > 0 && v.Status != "waiting"
	}
	for _, s := range v.Sources {
		if s.ID == source && (s.State == "ok" || s.State == "excluded") {
			return s.Rate, s.Rate > 0
		}
	}
	return 0, false
}
