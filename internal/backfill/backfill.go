// Package backfill loads historical prices on service start-up so that
// history charts are useful immediately, not only after days of running.
package backfill

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"

	"cryptomon/internal/catalog"
	"cryptomon/internal/httpx"
	"cryptomon/internal/model"
	"cryptomon/internal/sources"
	"cryptomon/internal/storage"
)

// AggSeries is the storage series name of an aggregated price.
func AggSeries(vs string) string { return "agg:" + vs }

// SourceSeries is the storage series name of one source's price.
func SourceSeries(source, vs string) string { return "src:" + source + ":" + vs }

// Crypto loads minute (12h), hourly (30d) and daily (2y) candles in USD
// from Kraken, falling back to Binance klines.
func Crypto(ctx context.Context, a catalog.Asset, st *storage.Store) error {
	series := AggSeries("USD")
	plan := []struct {
		kraken  int
		binance string
		tier    storage.Tier
	}{
		{1440, "1d", storage.Hour},
		{60, "1h", storage.Hour},
		{1, "1m", storage.Minute},
	}
	var errs []string
	for _, p := range plan {
		pts, err := krakenOHLC(ctx, a.KrakenOHLC, p.kraken)
		if (err != nil || len(pts) == 0) && a.BinanceKlines != "" {
			pts, err = binanceKlines(ctx, a.BinanceKlines, p.binance)
		}
		if err != nil {
			errs = append(errs, err.Error())
			continue
		}
		if err := st.Put(p.tier, series, pts); err != nil {
			return err
		}
	}
	if len(errs) == len(plan) {
		return errors.New(strings.Join(errs, "; "))
	}
	return nil
}

func krakenOHLC(ctx context.Context, pair string, interval int) ([]model.Point, error) {
	if pair == "" {
		return nil, errors.New("no kraken pair")
	}
	var r struct {
		Error  []string                   `json:"error"`
		Result map[string]json.RawMessage `json:"result"`
	}
	u := fmt.Sprintf("https://api.kraken.com/0/public/OHLC?pair=%s&interval=%d", url.QueryEscape(pair), interval)
	if err := httpx.GetJSON(ctx, u, &r); err != nil {
		return nil, err
	}
	if len(r.Error) > 0 {
		return nil, fmt.Errorf("kraken: %s", r.Error[0])
	}
	for k, raw := range r.Result {
		if k == "last" {
			continue
		}
		var rows [][]any
		if err := json.Unmarshal(raw, &rows); err != nil {
			return nil, err
		}
		pts := make([]model.Point, 0, len(rows))
		for _, row := range rows {
			if len(row) < 5 {
				continue
			}
			t, _ := row[0].(float64)
			c, _ := row[4].(string)
			pts = append(pts, model.Point{T: int64(t) * 1000, V: httpx.F(c)})
		}
		return pts, nil
	}
	return nil, errors.New("kraken: empty OHLC")
}

func binanceKlines(ctx context.Context, symbol, interval string) ([]model.Point, error) {
	var rows [][]any
	u := fmt.Sprintf("https://api.binance.com/api/v3/klines?symbol=%s&interval=%s&limit=1000", url.QueryEscape(symbol), interval)
	if err := httpx.GetJSON(ctx, u, &rows); err != nil {
		return nil, err
	}
	pts := make([]model.Point, 0, len(rows))
	for _, row := range rows {
		if len(row) < 5 {
			continue
		}
		t, _ := row[0].(float64)
		c, _ := row[4].(string)
		pts = append(pts, model.Point{T: int64(t), V: httpx.F(c)})
	}
	return pts, nil
}

// Fiat loads daily official rates for the last five years (ECB through
// Frankfurter plus the Bank of Russia for RUB) and hourly/15-minute rates
// for recent weeks from Yahoo Finance, for every quote currency.
func Fiat(ctx context.Context, code string, st *storage.Store) error {
	var errs []string
	if err := fiatDaily(ctx, code, st); err != nil {
		errs = append(errs, "daily: "+err.Error())
	}
	for _, p := range []struct {
		rng, interval string
		tier          storage.Tier
	}{{"1mo", "1h", storage.Hour}, {"5d", "15m", storage.Minute}} {
		if err := yahooSpark(ctx, code, p.rng, p.interval, p.tier, st); err != nil {
			errs = append(errs, "yahoo "+p.interval+": "+err.Error())
		}
	}
	if len(errs) == 3 {
		return errors.New(strings.Join(errs, "; "))
	}
	return nil
}

func fiatDaily(ctx context.Context, code string, st *storage.Store) error {
	start := time.Now().AddDate(-5, 0, 0)
	others := []string{}
	for _, c := range catalog.FiatCodes {
		if c != "USD" && c != "RUB" {
			others = append(others, c)
		}
	}
	var fr struct {
		Rates map[string]map[string]float64 `json:"rates"`
	}
	u := fmt.Sprintf("https://api.frankfurter.dev/v1/%s..?base=USD&symbols=%s", start.Format("2006-01-02"), strings.Join(others, ","))
	if err := httpx.GetJSON(ctx, u, &fr); err != nil {
		return err
	}
	rub, rubErr := cbrDynamic(ctx, "R01235", start, time.Now()) // USD/RUB
	dates := make([]string, 0, len(fr.Rates))
	for d := range fr.Rates {
		dates = append(dates, d)
	}
	sort.Strings(dates)

	series := map[string][]model.Point{}
	ri := 0
	var lastRub float64
	for _, d := range dates {
		day, err := time.Parse("2006-01-02", d)
		if err != nil {
			continue
		}
		usdTo := map[string]float64{"USD": 1}
		for c, v := range fr.Rates[d] {
			usdTo[c] = v
		}
		for ri < len(rub) && rub[ri].T <= day.UnixMilli() {
			lastRub = rub[ri].V
			ri++
		}
		if lastRub > 0 {
			usdTo["RUB"] = lastRub
		}
		x, ok := usdTo[code]
		if !ok {
			continue
		}
		ts := day.Add(15 * time.Hour).UnixMilli()
		for q, v := range usdTo {
			if q != code {
				series[q] = append(series[q], model.Point{T: ts, V: v / x})
			}
		}
	}
	for q, pts := range series {
		if err := st.Put(storage.Hour, AggSeries(q), pts); err != nil {
			return err
		}
	}
	if rubErr != nil {
		return fmt.Errorf("RUB history: %w", rubErr)
	}
	return nil
}

// cbrDynamic loads the Bank of Russia's daily rate for one currency id.
func cbrDynamic(ctx context.Context, id string, from, to time.Time) ([]model.Point, error) {
	u := fmt.Sprintf("https://www.cbr.ru/scripts/XML_dynamic.asp?date_req1=%s&date_req2=%s&VAL_NM_RQ=%s",
		from.Format("02/01/2006"), to.Format("02/01/2006"), id)
	body, err := httpx.Get(ctx, u)
	if err != nil {
		return nil, err
	}
	var r struct {
		Records []struct {
			Date    string `xml:"Date,attr"`
			Nominal string `xml:"Nominal"`
			Value   string `xml:"Value"`
		} `xml:"Record"`
	}
	if err := sources.DecodeLegacyXML(body, &r); err != nil {
		return nil, err
	}
	pts := make([]model.Point, 0, len(r.Records))
	for _, rec := range r.Records {
		d, err := time.Parse("02.01.2006", rec.Date)
		if err != nil {
			continue
		}
		n, v := httpx.F(rec.Nominal), httpx.F(rec.Value)
		if n > 0 && v > 0 {
			pts = append(pts, model.Point{T: d.UnixMilli(), V: v / n})
		}
	}
	return pts, nil
}

func yahooSpark(ctx context.Context, code, rng, interval string, tier storage.Tier, st *storage.Store) error {
	series, err := sources.YahooSpark(ctx, code, catalog.FiatCodes, rng, interval)
	if err != nil {
		return err
	}
	for q, pts := range series {
		if err := st.Put(tier, AggSeries(q), pts); err != nil {
			return err
		}
	}
	return nil
}
