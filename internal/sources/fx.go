package sources

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"math"
	"net/url"
	"strings"
	"time"

	"cryptomon/internal/httpx"
	"cryptomon/internal/model"
)

// Table is a rate table: Rates[c] is how many units of c one unit of Base
// buys. Base itself is implicitly 1.
type Table struct {
	Base  string
	Rates map[string]float64
	Time  int64 // unix ms of the data (publication time), 0 if unknown
}

func (t Table) rate(c string) (float64, bool) {
	if c == t.Base {
		return 1, true
	}
	v, ok := t.Rates[c]
	return v, ok && v > 0 && !math.IsInf(v, 0) && !math.IsNaN(v)
}

// Cross returns how many units of q one unit of x buys.
func (t Table) Cross(x, q string) (float64, bool) {
	rx, ok1 := t.rate(x)
	rq, ok2 := t.rate(q)
	if !ok1 || !ok2 {
		return 0, false
	}
	return rq / rx, true
}

// FXSource is a fiat rate provider.
type FXSource struct {
	ID       string
	Name     string
	URL      string
	Type     string // "reference" (central bank / aggregator) or "exchange"
	Interval time.Duration
	// Covers lists the currencies the source can price; nil means all.
	Covers []string
	Fetch  func(ctx context.Context, asset string, vs []string) (Table, error)
}

func (s FXSource) Supports(code string) bool {
	if s.Covers == nil {
		return true
	}
	for _, c := range s.Covers {
		if c == code {
			return true
		}
	}
	return false
}

// FXSources is the list of fiat providers, in display order.
var FXSources = []FXSource{
	{ID: "cbr", Name: "Bank of Russia", URL: "https://www.cbr.ru", Type: "reference", Interval: 30 * time.Minute, Fetch: fetchCBR},
	{ID: "ecb", Name: "European Central Bank", URL: "https://www.ecb.europa.eu", Type: "reference", Interval: 30 * time.Minute,
		Covers: []string{"USD", "EUR", "CNY", "JPY", "INR", "CHF", "GBP"}, Fetch: fetchECB},
	{ID: "yahoo", Name: "Yahoo Finance", URL: "https://finance.yahoo.com", Type: "reference", Interval: time.Minute, Fetch: fetchYahoo},
	{ID: "cbrates", Name: "Coinbase Rates", URL: "https://www.coinbase.com", Type: "reference", Interval: time.Minute, Fetch: fetchCoinbaseRates},
	{ID: "erapi", Name: "ExchangeRate-API", URL: "https://www.exchangerate-api.com", Type: "reference", Interval: time.Hour, Fetch: fetchERAPI},
	{ID: "currencyapi", Name: "Currency-API", URL: "https://github.com/fawazahmed0/exchange-api", Type: "reference", Interval: time.Hour, Fetch: fetchCurrencyAPI},
	{ID: "krakenfx", Name: "Kraken FX", URL: "https://www.kraken.com", Type: "exchange", Interval: 15 * time.Second,
		Covers: []string{"USD", "EUR", "GBP", "CHF"}, Fetch: fetchKrakenFX},
	{ID: "moex", Name: "Moscow Exchange", URL: "https://www.moex.com", Type: "exchange", Interval: 30 * time.Second,
		Covers: []string{"CNY", "RUB"}, Fetch: fetchMOEX},
}

// ---- Bank of Russia ---------------------------------------------------------

func fetchCBR(ctx context.Context, _ string, _ []string) (Table, error) {
	t, err := fetchCBRJSON(ctx)
	if err == nil {
		return t, nil
	}
	if t2, err2 := fetchCBRXML(ctx); err2 == nil {
		return t2, nil
	}
	return Table{}, err
}

// The cbr-xml-daily.ru mirror republishes the official XML as JSON.
func fetchCBRJSON(ctx context.Context) (Table, error) {
	var r struct {
		Date   string `json:"Date"`
		Valute map[string]struct {
			Nominal float64 `json:"Nominal"`
			Value   float64 `json:"Value"`
		} `json:"Valute"`
	}
	if err := httpx.GetJSON(ctx, "https://www.cbr-xml-daily.ru/daily_json.js", &r); err != nil {
		return Table{}, err
	}
	t := Table{Base: "RUB", Rates: map[string]float64{}}
	for code, v := range r.Valute {
		if v.Value > 0 && v.Nominal > 0 {
			t.Rates[code] = v.Nominal / v.Value
		}
	}
	if ts, err := time.Parse(time.RFC3339, r.Date); err == nil {
		t.Time = ts.UnixMilli()
	}
	if len(t.Rates) == 0 {
		return Table{}, errors.New("cbr: empty table")
	}
	return t, nil
}

func fetchCBRXML(ctx context.Context) (Table, error) {
	body, err := httpx.Get(ctx, "https://www.cbr.ru/scripts/XML_daily.asp")
	if err != nil {
		return Table{}, err
	}
	var r struct {
		Date   string `xml:"Date,attr"`
		Valute []struct {
			Code    string `xml:"CharCode"`
			Nominal string `xml:"Nominal"`
			Value   string `xml:"Value"`
		} `xml:"Valute"`
	}
	if err := DecodeLegacyXML(body, &r); err != nil {
		return Table{}, err
	}
	t := Table{Base: "RUB", Rates: map[string]float64{}}
	for _, v := range r.Valute {
		n, val := httpx.F(v.Nominal), httpx.F(v.Value)
		if n > 0 && val > 0 {
			t.Rates[v.Code] = n / val
		}
	}
	if ts, err := time.ParseInLocation("02.01.2006", r.Date, moscow()); err == nil {
		t.Time = ts.UnixMilli()
	}
	return t, nil
}

// DecodeLegacyXML decodes XML in a single-byte legacy encoding
// (windows-1251). Only ASCII fields are needed, so non-ASCII bytes are
// replaced instead of pulling in a full charset library.
func DecodeLegacyXML(body []byte, v any) error {
	d := xml.NewDecoder(bytes.NewReader(body))
	d.CharsetReader = func(_ string, in io.Reader) (io.Reader, error) {
		b, err := io.ReadAll(in)
		if err != nil {
			return nil, err
		}
		for i, c := range b {
			if c >= 0x80 {
				b[i] = '?'
			}
		}
		return bytes.NewReader(b), nil
	}
	return d.Decode(v)
}

func moscow() *time.Location {
	if loc, err := time.LoadLocation("Europe/Moscow"); err == nil {
		return loc
	}
	return time.FixedZone("MSK", 3*3600)
}

// ---- European Central Bank -------------------------------------------------

func fetchECB(ctx context.Context, _ string, _ []string) (Table, error) {
	body, err := httpx.Get(ctx, "https://www.ecb.europa.eu/stats/eurofxref/eurofxref-daily.xml")
	if err != nil {
		return Table{}, err
	}
	var r struct {
		Cube struct {
			Cube struct {
				Time  string `xml:"time,attr"`
				Rates []struct {
					Currency string `xml:"currency,attr"`
					Rate     string `xml:"rate,attr"`
				} `xml:"Cube"`
			} `xml:"Cube"`
		} `xml:"Cube"`
	}
	if err := xml.Unmarshal(body, &r); err != nil {
		return Table{}, err
	}
	t := Table{Base: "EUR", Rates: map[string]float64{}}
	for _, c := range r.Cube.Cube.Rates {
		t.Rates[c.Currency] = httpx.F(c.Rate)
	}
	if ts, err := time.Parse("2006-01-02", r.Cube.Cube.Time); err == nil {
		t.Time = ts.Add(15 * time.Hour).UnixMilli() // published ~16:00 CET
	}
	if len(t.Rates) == 0 {
		return Table{}, errors.New("ecb: empty table")
	}
	return t, nil
}

// ---- Yahoo Finance (near real-time FX) --------------------------------------

// YahooSpark loads price series of asset against every q in vs. Yahoo
// rounds quotes to ~4 decimals, so small rates (JPY→USD = 0.0064) are
// taken from the inverse pair (USDJPY = 156.8) and inverted.
func YahooSpark(ctx context.Context, asset string, vs []string, rng, interval string) (map[string][]model.Point, error) {
	var syms []string
	for _, q := range vs {
		if q != asset {
			syms = append(syms, asset+q+"=X", q+asset+"=X")
		}
	}
	u := fmt.Sprintf("https://query1.finance.yahoo.com/v8/finance/spark?range=%s&interval=%s&symbols=%s",
		rng, interval, url.QueryEscape(strings.Join(syms, ",")))
	var r map[string]struct {
		Timestamp []int64    `json:"timestamp"`
		Close     []*float64 `json:"close"`
	}
	if err := httpx.GetJSON(ctx, u, &r); err != nil {
		return nil, err
	}
	series := func(sym string, invert bool) []model.Point {
		s, ok := r[sym]
		if !ok {
			return nil
		}
		pts := make([]model.Point, 0, len(s.Timestamp))
		for i, t := range s.Timestamp {
			if i < len(s.Close) && s.Close[i] != nil && *s.Close[i] > 0 {
				v := *s.Close[i]
				if invert {
					v = 1 / v
				}
				pts = append(pts, model.Point{T: t * 1000, V: v})
			}
		}
		return pts
	}
	out := map[string][]model.Point{}
	for _, q := range vs {
		if q == asset {
			continue
		}
		direct := series(asset+q+"=X", false)
		if len(direct) > 0 && direct[len(direct)-1].V >= 1 {
			out[q] = direct
		} else if inv := series(q+asset+"=X", true); len(inv) > 0 {
			out[q] = inv
		} else if len(direct) > 0 {
			out[q] = direct
		}
	}
	if len(out) == 0 {
		return nil, errors.New("yahoo: no data")
	}
	return out, nil
}

func fetchYahoo(ctx context.Context, asset string, vs []string) (Table, error) {
	series, err := YahooSpark(ctx, asset, vs, "1d", "5m")
	if err != nil {
		return Table{}, err
	}
	t := Table{Base: asset, Rates: map[string]float64{}}
	for q, pts := range series {
		last := pts[len(pts)-1]
		t.Rates[q] = last.V
		t.Time = max(t.Time, last.T)
	}
	return t, nil
}

// ---- Coinbase exchange-rates ------------------------------------------------

func fetchCoinbaseRates(ctx context.Context, _ string, _ []string) (Table, error) {
	var r struct {
		Data struct {
			Currency string            `json:"currency"`
			Rates    map[string]string `json:"rates"`
		} `json:"data"`
	}
	if err := httpx.GetJSON(ctx, "https://api.coinbase.com/v2/exchange-rates?currency=USD", &r); err != nil {
		return Table{}, err
	}
	t := Table{Base: "USD", Rates: map[string]float64{}, Time: time.Now().UnixMilli()}
	for c, v := range r.Data.Rates {
		t.Rates[c] = httpx.F(v)
	}
	if len(t.Rates) == 0 {
		return Table{}, errors.New("coinbase: empty table")
	}
	return t, nil
}

// ---- open.er-api.com -------------------------------------------------------

func fetchERAPI(ctx context.Context, _ string, _ []string) (Table, error) {
	var r struct {
		Result string             `json:"result"`
		Update int64              `json:"time_last_update_unix"`
		Rates  map[string]float64 `json:"rates"`
	}
	if err := httpx.GetJSON(ctx, "https://open.er-api.com/v6/latest/USD", &r); err != nil {
		return Table{}, err
	}
	if r.Result != "success" {
		return Table{}, fmt.Errorf("erapi: %s", r.Result)
	}
	return Table{Base: "USD", Rates: r.Rates, Time: r.Update * 1000}, nil
}

// ---- fawazahmed0/exchange-api ----------------------------------------------

func fetchCurrencyAPI(ctx context.Context, _ string, _ []string) (Table, error) {
	urls := []string{
		"https://cdn.jsdelivr.net/npm/@fawazahmed0/currency-api@latest/v1/currencies/usd.json",
		"https://latest.currency-api.pages.dev/v1/currencies/usd.json",
	}
	var lastErr error
	for _, u := range urls {
		var r struct {
			Date string             `json:"date"`
			USD  map[string]float64 `json:"usd"`
		}
		if err := httpx.GetJSON(ctx, u, &r); err != nil {
			lastErr = err
			continue
		}
		t := Table{Base: "USD", Rates: map[string]float64{}}
		for c, v := range r.USD {
			t.Rates[strings.ToUpper(c)] = v
		}
		if ts, err := time.Parse("2006-01-02", r.Date); err == nil {
			t.Time = ts.UnixMilli()
		}
		return t, nil
	}
	return Table{}, lastErr
}

// ---- Kraken spot FX ---------------------------------------------------------

func fetchKrakenFX(ctx context.Context, _ string, _ []string) (Table, error) {
	var r struct {
		Error  []string `json:"error"`
		Result map[string]struct {
			A []string `json:"a"`
			B []string `json:"b"`
			C []string `json:"c"`
		} `json:"result"`
	}
	if err := httpx.GetJSON(ctx, "https://api.kraken.com/0/public/Ticker?pair=EURUSD,GBPUSD,USDCHF", &r); err != nil {
		return Table{}, err
	}
	if len(r.Error) > 0 {
		return Table{}, fmt.Errorf("kraken: %s", r.Error[0])
	}
	t := Table{Base: "USD", Rates: map[string]float64{}, Time: time.Now().UnixMilli()}
	for key, v := range r.Result {
		if len(v.A) == 0 || len(v.B) == 0 {
			continue
		}
		ask, bid := httpx.F(v.A[0]), httpx.F(v.B[0])
		mid := (ask + bid) / 2
		if mid <= 0 || (ask-bid)/mid > 0.005 { // thin book, skip
			continue
		}
		switch key {
		case "ZEURZUSD", "EURUSD":
			t.Rates["EUR"] = 1 / mid
		case "ZGBPZUSD", "GBPUSD":
			t.Rates["GBP"] = 1 / mid
		case "USDCHF", "ZUSDZCHF":
			t.Rates["CHF"] = mid
		}
	}
	if len(t.Rates) == 0 {
		return Table{}, errors.New("kraken: no liquid FX pairs")
	}
	return t, nil
}

// ---- Moscow Exchange (CNY/RUB) -------------------------------------------------

func fetchMOEX(ctx context.Context, _ string, _ []string) (Table, error) {
	var r struct {
		Marketdata struct {
			Data [][]json.RawMessage `json:"data"`
		} `json:"marketdata"`
	}
	u := "https://iss.moex.com/iss/engines/currency/markets/selt/boards/CETS/securities/CNYRUB_TOM.json" +
		"?iss.meta=off&iss.only=marketdata&marketdata.columns=LAST,BID,OFFER,SYSTIME"
	if err := httpx.GetJSON(ctx, u, &r); err != nil {
		return Table{}, err
	}
	if len(r.Marketdata.Data) == 0 || len(r.Marketdata.Data[0]) < 4 {
		return Table{}, errors.New("moex: no data")
	}
	row := r.Marketdata.Data[0]
	var last, bid, offer float64
	var sys string
	_ = json.Unmarshal(row[0], &last)
	_ = json.Unmarshal(row[1], &bid)
	_ = json.Unmarshal(row[2], &offer)
	_ = json.Unmarshal(row[3], &sys)
	price := last
	if bid > 0 && offer > 0 {
		price = (bid + offer) / 2
	}
	if price <= 0 {
		return Table{}, errors.New("moex: no price")
	}
	t := Table{Base: "RUB", Rates: map[string]float64{"CNY": 1 / price}}
	if ts, err := time.ParseInLocation("2006-01-02 15:04:05", sys, moscow()); err == nil {
		t.Time = ts.UnixMilli()
	}
	return t, nil
}
