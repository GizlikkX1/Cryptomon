// Package sources contains adapters for public, key-less market data APIs:
// crypto exchanges (this file) and fiat reference/FX feeds (fx.go).
package sources

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"cryptomon/internal/httpx"
)

// Ticker is the best bid/ask snapshot of one pair, in the pair's own quote
// currency.
type Ticker struct {
	Bid, Ask, Last float64
	Volume         float64 // 24h volume in base units
	Change24h      *float64
	Time           int64 // unix ms
}

// TickerFunc fetches a ticker for an exchange-specific symbol.
type TickerFunc func(ctx context.Context, symbol string) (Ticker, error)

// Exchanges maps exchange ids (see catalog.Exchanges) to their adapter.
var Exchanges = map[string]TickerFunc{
	"binance":  binance,
	"bybit":    bybit,
	"okx":      okx,
	"kraken":   kraken,
	"coinbase": coinbase,
	"kucoin":   kucoin,
	"gate":     gate,
	"htx":      htx,
	"bitfinex": bitfinex,
	"mexc":     mexc,
	"bitget":   bitget,
	"bitstamp": bitstamp,
}

var errNoBook = errors.New("no order book (market halted or delisted)")

func now() int64 { return time.Now().UnixMilli() }

// pct returns the percentage change from open to last, or nil.
func pct(last, open float64) *float64 {
	if open <= 0 || last <= 0 {
		return nil
	}
	v := (last - open) / open * 100
	return &v
}

func fracPct(frac float64, ok bool) *float64 {
	if !ok {
		return nil
	}
	v := frac * 100
	return &v
}

func validate(t Ticker) (Ticker, error) {
	if t.Bid <= 0 || t.Ask <= 0 {
		return t, errNoBook
	}
	if t.Last <= 0 {
		t.Last = (t.Bid + t.Ask) / 2
	}
	if t.Time == 0 {
		t.Time = now()
	}
	return t, nil
}

// Invert turns a ticker of X/Y into Y/X.
func (t Ticker) Invert() Ticker {
	inv := Ticker{Time: t.Time}
	if t.Ask > 0 {
		inv.Bid = 1 / t.Ask
	}
	if t.Bid > 0 {
		inv.Ask = 1 / t.Bid
	}
	if t.Last > 0 {
		inv.Last = 1 / t.Last
	}
	if t.Change24h != nil {
		v := (1/(1+*t.Change24h/100) - 1) * 100
		inv.Change24h = &v
	}
	return inv
}

func binance(ctx context.Context, sym string) (Ticker, error) {
	var r struct {
		Bid   string `json:"bidPrice"`
		Ask   string `json:"askPrice"`
		Last  string `json:"lastPrice"`
		Vol   string `json:"volume"`
		Open  string `json:"openPrice"`
		Close int64  `json:"closeTime"`
	}
	if err := httpx.GetJSON(ctx, "https://api.binance.com/api/v3/ticker/24hr?symbol="+url.QueryEscape(sym), &r); err != nil {
		return Ticker{}, err
	}
	last := httpx.F(r.Last)
	return validate(Ticker{Bid: httpx.F(r.Bid), Ask: httpx.F(r.Ask), Last: last, Volume: httpx.F(r.Vol),
		Change24h: pct(last, httpx.F(r.Open)), Time: r.Close})
}

func bybit(ctx context.Context, sym string) (Ticker, error) {
	var r struct {
		RetCode int    `json:"retCode"`
		RetMsg  string `json:"retMsg"`
		Time    int64  `json:"time"`
		Result  struct {
			List []struct {
				Bid  string `json:"bid1Price"`
				Ask  string `json:"ask1Price"`
				Last string `json:"lastPrice"`
				Prev string `json:"prevPrice24h"`
				Vol  string `json:"volume24h"`
			} `json:"list"`
		} `json:"result"`
	}
	if err := httpx.GetJSON(ctx, "https://api.bybit.com/v5/market/tickers?category=spot&symbol="+url.QueryEscape(sym), &r); err != nil {
		return Ticker{}, err
	}
	if r.RetCode != 0 || len(r.Result.List) == 0 {
		return Ticker{}, fmt.Errorf("bybit: %s", r.RetMsg)
	}
	t := r.Result.List[0]
	last := httpx.F(t.Last)
	return validate(Ticker{Bid: httpx.F(t.Bid), Ask: httpx.F(t.Ask), Last: last, Volume: httpx.F(t.Vol),
		Change24h: pct(last, httpx.F(t.Prev)), Time: r.Time})
}

func okx(ctx context.Context, sym string) (Ticker, error) {
	var r struct {
		Code string `json:"code"`
		Msg  string `json:"msg"`
		Data []struct {
			Bid  string `json:"bidPx"`
			Ask  string `json:"askPx"`
			Last string `json:"last"`
			Open string `json:"open24h"`
			Vol  string `json:"vol24h"`
			Ts   string `json:"ts"`
		} `json:"data"`
	}
	if err := httpx.GetJSON(ctx, "https://www.okx.com/api/v5/market/ticker?instId="+url.QueryEscape(sym), &r); err != nil {
		return Ticker{}, err
	}
	if r.Code != "0" || len(r.Data) == 0 {
		return Ticker{}, fmt.Errorf("okx: %s", r.Msg)
	}
	d := r.Data[0]
	last := httpx.F(d.Last)
	ts, _ := strconv.ParseInt(d.Ts, 10, 64)
	return validate(Ticker{Bid: httpx.F(d.Bid), Ask: httpx.F(d.Ask), Last: last, Volume: httpx.F(d.Vol),
		Change24h: pct(last, httpx.F(d.Open)), Time: ts})
}

func kraken(ctx context.Context, sym string) (Ticker, error) {
	var r struct {
		Error  []string `json:"error"`
		Result map[string]struct {
			A []string `json:"a"`
			B []string `json:"b"`
			C []string `json:"c"`
			V []string `json:"v"`
		} `json:"result"`
	}
	if err := httpx.GetJSON(ctx, "https://api.kraken.com/0/public/Ticker?pair="+url.QueryEscape(sym), &r); err != nil {
		return Ticker{}, err
	}
	if len(r.Error) > 0 {
		return Ticker{}, fmt.Errorf("kraken: %s", r.Error[0])
	}
	for _, t := range r.Result {
		if len(t.A) == 0 || len(t.B) == 0 || len(t.C) == 0 {
			break
		}
		var vol float64
		if len(t.V) > 1 {
			vol = httpx.F(t.V[1])
		}
		// Kraken's "o" is today's UTC open, not a rolling 24h value, so the
		// 24h change is left to our own history.
		return validate(Ticker{Bid: httpx.F(t.B[0]), Ask: httpx.F(t.A[0]), Last: httpx.F(t.C[0]), Volume: vol})
	}
	return Ticker{}, errors.New("kraken: empty result")
}

func coinbase(ctx context.Context, sym string) (Ticker, error) {
	var r struct {
		Bid     string `json:"bid"`
		Ask     string `json:"ask"`
		Price   string `json:"price"`
		Vol     string `json:"volume"`
		Time    string `json:"time"`
		Message string `json:"message"`
	}
	if err := httpx.GetJSON(ctx, "https://api.exchange.coinbase.com/products/"+url.PathEscape(sym)+"/ticker", &r); err != nil {
		return Ticker{}, err
	}
	if r.Message != "" {
		return Ticker{}, fmt.Errorf("coinbase: %s", r.Message)
	}
	var ts int64
	if t, err := time.Parse(time.RFC3339Nano, r.Time); err == nil {
		ts = t.UnixMilli()
	}
	return validate(Ticker{Bid: httpx.F(r.Bid), Ask: httpx.F(r.Ask), Last: httpx.F(r.Price), Volume: httpx.F(r.Vol), Time: ts})
}

func kucoin(ctx context.Context, sym string) (Ticker, error) {
	var r struct {
		Code string `json:"code"`
		Msg  string `json:"msg"`
		Data *struct {
			Time   int64  `json:"time"`
			Buy    string `json:"buy"`
			Sell   string `json:"sell"`
			Last   string `json:"last"`
			Vol    string `json:"vol"`
			Change string `json:"changeRate"`
		} `json:"data"`
	}
	if err := httpx.GetJSON(ctx, "https://api.kucoin.com/api/v1/market/stats?symbol="+url.QueryEscape(sym), &r); err != nil {
		return Ticker{}, err
	}
	if r.Code != "200000" || r.Data == nil || r.Data.Buy == "" {
		return Ticker{}, errNoBook
	}
	d := r.Data
	return validate(Ticker{Bid: httpx.F(d.Buy), Ask: httpx.F(d.Sell), Last: httpx.F(d.Last), Volume: httpx.F(d.Vol),
		Change24h: fracPct(httpx.F(d.Change), d.Change != ""), Time: d.Time})
}

func gate(ctx context.Context, sym string) (Ticker, error) {
	body, err := httpx.Get(ctx, "https://api.gateio.ws/api/v4/spot/tickers?currency_pair="+url.QueryEscape(sym))
	if err != nil {
		var se *httpx.StatusError
		if errors.As(err, &se) {
			var e struct {
				Message string `json:"message"`
			}
			if json.Unmarshal([]byte(se.Body), &e) == nil && e.Message != "" {
				return Ticker{}, fmt.Errorf("gate: %s", e.Message)
			}
		}
		return Ticker{}, err
	}
	var r []struct {
		Bid    string `json:"highest_bid"`
		Ask    string `json:"lowest_ask"`
		Last   string `json:"last"`
		Vol    string `json:"base_volume"`
		Change string `json:"change_percentage"`
	}
	if err := json.Unmarshal(body, &r); err != nil || len(r) == 0 {
		return Ticker{}, errors.New("gate: unexpected response")
	}
	t := r[0]
	var chg *float64
	if t.Change != "" {
		chg = new(float64)
		*chg = httpx.F(t.Change)
	}
	return validate(Ticker{Bid: httpx.F(t.Bid), Ask: httpx.F(t.Ask), Last: httpx.F(t.Last), Volume: httpx.F(t.Vol), Change24h: chg})
}

func htx(ctx context.Context, sym string) (Ticker, error) {
	var r struct {
		Status string `json:"status"`
		ErrMsg string `json:"err-msg"`
		Ts     int64  `json:"ts"`
		Tick   struct {
			Open   float64   `json:"open"`
			Close  float64   `json:"close"`
			Amount float64   `json:"amount"`
			Bid    []float64 `json:"bid"`
			Ask    []float64 `json:"ask"`
		} `json:"tick"`
	}
	if err := httpx.GetJSON(ctx, "https://api.huobi.pro/market/detail/merged?symbol="+url.QueryEscape(sym), &r); err != nil {
		return Ticker{}, err
	}
	if r.Status != "ok" {
		return Ticker{}, fmt.Errorf("htx: %s", r.ErrMsg)
	}
	t := Ticker{Last: r.Tick.Close, Volume: r.Tick.Amount, Change24h: pct(r.Tick.Close, r.Tick.Open), Time: r.Ts}
	if len(r.Tick.Bid) > 0 {
		t.Bid = r.Tick.Bid[0]
	}
	if len(r.Tick.Ask) > 0 {
		t.Ask = r.Tick.Ask[0]
	}
	return validate(t)
}

func bitfinex(ctx context.Context, sym string) (Ticker, error) {
	body, err := httpx.Get(ctx, "https://api-pub.bitfinex.com/v2/ticker/"+url.PathEscape(sym))
	if err != nil {
		return Ticker{}, err
	}
	var raw []json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		return Ticker{}, errors.New("bitfinex: unexpected response")
	}
	if len(raw) < 8 {
		if len(raw) > 0 && string(raw[0]) == `"error"` {
			return Ticker{}, fmt.Errorf("bitfinex: %s", string(body))
		}
		return Ticker{}, errors.New("bitfinex: unknown symbol")
	}
	num := func(i int) float64 {
		var v float64
		_ = json.Unmarshal(raw[i], &v)
		return v
	}
	// [BID, BID_SIZE, ASK, ASK_SIZE, DAILY_CHANGE, DAILY_CHANGE_RELATIVE, LAST_PRICE, VOLUME, HIGH, LOW]
	return validate(Ticker{Bid: num(0), Ask: num(2), Last: num(6), Volume: num(7), Change24h: fracPct(num(5), true)})
}

func mexc(ctx context.Context, sym string) (Ticker, error) {
	var r struct {
		Bid   string `json:"bidPrice"`
		Ask   string `json:"askPrice"`
		Last  string `json:"lastPrice"`
		Open  string `json:"openPrice"`
		Vol   string `json:"volume"`
		Close int64  `json:"closeTime"`
		Msg   string `json:"msg"`
	}
	if err := httpx.GetJSON(ctx, "https://api.mexc.com/api/v3/ticker/24hr?symbol="+url.QueryEscape(sym), &r); err != nil {
		return Ticker{}, err
	}
	last := httpx.F(r.Last)
	return validate(Ticker{Bid: httpx.F(r.Bid), Ask: httpx.F(r.Ask), Last: last, Volume: httpx.F(r.Vol),
		Change24h: pct(last, httpx.F(r.Open)), Time: r.Close})
}

func bitget(ctx context.Context, sym string) (Ticker, error) {
	var r struct {
		Code string `json:"code"`
		Msg  string `json:"msg"`
		Data []struct {
			Bid  string `json:"bidPr"`
			Ask  string `json:"askPr"`
			Last string `json:"lastPr"`
			Open string `json:"open"`
			Vol  string `json:"baseVolume"`
			Ts   string `json:"ts"`
		} `json:"data"`
	}
	if err := httpx.GetJSON(ctx, "https://api.bitget.com/api/v2/spot/market/tickers?symbol="+url.QueryEscape(sym), &r); err != nil {
		return Ticker{}, err
	}
	if r.Code != "00000" || len(r.Data) == 0 {
		return Ticker{}, fmt.Errorf("bitget: %s", r.Msg)
	}
	d := r.Data[0]
	last := httpx.F(d.Last)
	ts, _ := strconv.ParseInt(d.Ts, 10, 64)
	return validate(Ticker{Bid: httpx.F(d.Bid), Ask: httpx.F(d.Ask), Last: last, Volume: httpx.F(d.Vol),
		Change24h: pct(last, httpx.F(d.Open)), Time: ts})
}

func bitstamp(ctx context.Context, sym string) (Ticker, error) {
	var r struct {
		Bid    string `json:"bid"`
		Ask    string `json:"ask"`
		Last   string `json:"last"`
		Vol    string `json:"volume"`
		Open24 string `json:"open_24"`
		Ts     string `json:"timestamp"`
	}
	if err := httpx.GetJSON(ctx, "https://www.bitstamp.net/api/v2/ticker/"+url.PathEscape(sym)+"/", &r); err != nil {
		return Ticker{}, err
	}
	last := httpx.F(r.Last)
	ts, _ := strconv.ParseInt(r.Ts, 10, 64)
	return validate(Ticker{Bid: httpx.F(r.Bid), Ask: httpx.F(r.Ask), Last: last, Volume: httpx.F(r.Vol),
		Change24h: pct(last, httpx.F(r.Open24)), Time: ts * 1000})
}
