// Package model holds the data types shared by the currency microservices
// and the gateway. Everything here is serialised as JSON over HTTP/SSE.
package model

import (
	"encoding/json"
	"fmt"
	"math"
)

type Kind string

const (
	Crypto Kind = "crypto"
	Fiat   Kind = "fiat"
)

// SourceInfo describes one market data source: an exchange or a reference
// rate provider (central bank, FX feed).
type SourceInfo struct {
	ID   string  `json:"id"`
	Name string  `json:"name"`
	Type string  `json:"type"` // "exchange" | "reference"
	Fee  float64 `json:"fee"`  // default taker fee as a fraction (0.001 = 0.1%)
	URL  string  `json:"url"`
	Pair string  `json:"pair,omitempty"`
}

// AssetInfo is what a currency service reports about itself.
type AssetInfo struct {
	Code    string       `json:"code"`
	Name    string       `json:"name"`
	Kind    Kind         `json:"kind"`
	Symbol  string       `json:"symbol"`
	Color   string       `json:"color"`
	Quotes  []string     `json:"quotes"` // currencies this service prices its asset in
	Sources []SourceInfo `json:"sources"`
	Version string       `json:"version"`
}

// Quote is the price of an asset on one source against the Vs currency.
// For crypto Vs is always USD (USDT pairs are normalised with the live
// USDT/USD rate); fiat services quote against every other fiat currency.
type Quote struct {
	Source    string   `json:"source"`
	Vs        string   `json:"vs"`
	Pair      string   `json:"pair,omitempty"`
	Native    string   `json:"native,omitempty"` // quote currency of the original pair
	Bid       float64  `json:"bid"`
	Ask       float64  `json:"ask"`
	Last      float64  `json:"last"`
	Rate      float64  `json:"rate"` // representative price: mid of bid/ask, else last
	Volume    float64  `json:"volume,omitempty"`
	Change24h *float64 `json:"change24h,omitempty"`
	Fee       float64  `json:"fee"`
	Time      int64    `json:"time"`    // data timestamp, unix ms
	Fetched   int64    `json:"fetched"` // when we last got it, unix ms
	Latency   int64    `json:"latency"` // ms
	OK        bool     `json:"ok"`
	Error     string   `json:"error,omitempty"`
}

// BestPrice is the best place to buy or sell.
type BestPrice struct {
	Source    string  `json:"source"`
	Price     float64 `json:"price"`     // raw ask (buy) or bid (sell)
	Effective float64 `json:"effective"` // price after the taker fee
}

// Best summarises where it is cheapest to buy and most profitable to sell.
type Best struct {
	Buy       *BestPrice `json:"buy,omitempty"`
	Sell      *BestPrice `json:"sell,omitempty"`
	WorstBuy  *BestPrice `json:"worstBuy,omitempty"`
	WorstSell *BestPrice `json:"worstSell,omitempty"`
	// SpreadPct is the arbitrage margin after fees: buy on Buy.Source,
	// sell on Sell.Source. Positive means a theoretical profit.
	SpreadPct float64 `json:"spreadPct"`
	// SavingPct is how much cheaper the best buy is than the worst one.
	SavingPct float64 `json:"savingPct"`
}

// Market is the aggregated state of an asset against one currency.
type Market struct {
	Vs        string   `json:"vs"`
	Price     float64  `json:"price"` // median of healthy sources
	Change24h *float64 `json:"change24h,omitempty"`
	Best      Best     `json:"best"`
	Online    int      `json:"online"`
}

// Snapshot is the full live state of one currency service.
type Snapshot struct {
	Code     string             `json:"code"`
	Kind     Kind               `json:"kind"`
	Time     int64              `json:"time"`
	Seq      uint64             `json:"seq"`
	Markets  map[string]*Market `json:"markets"`
	Quotes   []Quote            `json:"quotes"`
	USDTRate float64            `json:"usdtRate,omitempty"`
}

// Point is a single time-series sample, encoded as [t, v] to keep history
// payloads small.
type Point struct {
	T int64
	V float64
}

func (p Point) MarshalJSON() ([]byte, error) {
	if math.IsNaN(p.V) || math.IsInf(p.V, 0) {
		return []byte(fmt.Sprintf("[%d,null]", p.T)), nil
	}
	return json.Marshal([2]any{p.T, p.V})
}

func (p *Point) UnmarshalJSON(b []byte) error {
	var raw [2]*float64
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	if raw[0] != nil {
		p.T = int64(*raw[0])
	}
	p.V = math.NaN()
	if raw[1] != nil {
		p.V = *raw[1]
	}
	return nil
}

// History is a set of series for one asset, keyed by source id ("agg" is
// the aggregated median).
type History struct {
	Code   string             `json:"code"`
	Vs     string             `json:"vs"`
	From   int64              `json:"from"`
	To     int64              `json:"to"`
	Series map[string][]Point `json:"series"`
}

// Ptr is a small helper for optional numbers.
func Ptr(v float64) *float64 { return &v }
