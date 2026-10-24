package gateway

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"cryptomon/internal/httpx"
	"cryptomon/internal/model"
)

var ranges = map[string]time.Duration{
	"1h":  time.Hour,
	"6h":  6 * time.Hour,
	"24h": 24 * time.Hour,
	"7d":  7 * 24 * time.Hour,
	"30d": 30 * 24 * time.Hour,
	"90d": 90 * 24 * time.Hour,
	"1y":  365 * 24 * time.Hour,
	"5y":  5 * 365 * 24 * time.Hour,
}

func (g *Gateway) serviceURL(code string) (string, model.Kind, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	st := g.svcs[code]
	if st == nil {
		return "", "", fmt.Errorf("unknown asset %s", code)
	}
	kind := model.Fiat
	if st.info != nil {
		kind = st.info.Kind
	} else if code == "BTC" || code == "ETH" || code == "USDT" || code == "TON" {
		kind = model.Crypto
	}
	return st.ref.URL, kind, nil
}

func fetchHistory(ctx context.Context, base, vs string, sources []string, from, to int64, points int) (*model.History, error) {
	u := fmt.Sprintf("%s/api/v1/history?vs=%s&sources=%s&from=%d&to=%d&points=%d",
		base, url.QueryEscape(vs), url.QueryEscape(strings.Join(sources, ",")), from, to, points)
	var h model.History
	if err := httpx.GetJSON(ctx, u, &h); err != nil {
		return nil, err
	}
	return &h, nil
}

// history returns price series of code in vs. Crypto services store USD
// only; other currencies are converted point by point with the USD
// service's own history, so past values use past exchange rates.
func (g *Gateway) history(ctx context.Context, code, vs string, from, to int64, sources []string, points int) (*model.History, error) {
	base, kind, err := g.serviceURL(code)
	if err != nil {
		return nil, err
	}
	vs = effectiveVs(code, vs)
	if kind == model.Fiat {
		return fetchHistory(ctx, base, vs, sources, from, to, points)
	}
	h, err := fetchHistory(ctx, base, "USD", sources, from, to, points)
	if err != nil || vs == "USD" {
		return h, err
	}
	usdURL, _, err := g.serviceURL("USD")
	if err != nil {
		return h, nil
	}
	// Look back further so the first points have an FX rate (weekends).
	fx, err := fetchHistory(ctx, usdURL, vs, []string{"agg"}, from-7*24*3600*1000, to, 3000)
	if err != nil || len(fx.Series["agg"]) == 0 {
		return nil, errors.New("no " + vs + " exchange-rate history yet")
	}
	rates := fx.Series["agg"]
	for id, pts := range h.Series {
		h.Series[id] = convert(pts, rates)
	}
	h.Vs = vs
	return h, nil
}

// convert multiplies each point by the latest rate at or before it.
func convert(pts, rates []model.Point) []model.Point {
	out := make([]model.Point, 0, len(pts))
	j := 0
	for _, p := range pts {
		for j+1 < len(rates) && rates[j+1].T <= p.T {
			j++
		}
		out = append(out, model.Point{T: p.T, V: p.V * rates[j].V})
	}
	return out
}

func valueAt(ctx context.Context, base, vs, source string, t int64) (float64, bool) {
	var r struct {
		OK bool    `json:"ok"`
		T  int64   `json:"t"`
		V  float64 `json:"v"`
	}
	u := fmt.Sprintf("%s/api/v1/value-at?vs=%s&source=%s&t=%d", base, url.QueryEscape(vs), url.QueryEscape(source), t)
	if err := httpx.GetJSON(ctx, u, &r); err != nil || !r.OK {
		return 0, false
	}
	// A reference much older than requested means a gap in history.
	if t-r.T > 3*24*3600*1000 {
		return 0, false
	}
	return r.V, r.V > 0
}

// referencePrice is the alert engine's look-back: price of asset in vs at t.
func (g *Gateway) referencePrice(asset, vs, source string, t time.Time) (float64, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	base, kind, err := g.serviceURL(asset)
	if err != nil {
		return 0, false
	}
	if source == "" {
		source = "agg"
	}
	ms := t.UnixMilli()
	if kind == model.Fiat {
		return valueAt(ctx, base, vs, source, ms)
	}
	v, ok := valueAt(ctx, base, "USD", source, ms)
	if !ok || vs == "USD" {
		return v, ok
	}
	usdURL, _, err := g.serviceURL("USD")
	if err != nil {
		return 0, false
	}
	fx, ok := valueAt(ctx, usdURL, vs, "agg", ms)
	return v * fx, ok
}
