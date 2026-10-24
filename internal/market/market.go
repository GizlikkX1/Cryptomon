// Package market holds the pricing maths shared by the services and the
// gateway: robust aggregation and the best-rate search.
package market

import (
	"math"
	"sort"

	"cryptomon/internal/model"
)

// OutlierTolerance is the max relative deviation from the median before a
// source is excluded as broken or stale.
const OutlierTolerance = 0.08

// Median returns the median of vals (0 for an empty slice).
func Median(vals []float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	s := append([]float64(nil), vals...)
	sort.Float64s(s)
	n := len(s)
	if n%2 == 1 {
		return s[n/2]
	}
	return (s[n/2-1] + s[n/2]) / 2
}

// Aggregate returns the median rate of healthy quotes for vs and marks
// outliers as not OK (in place).
func Aggregate(quotes []model.Quote, vs string) (price float64, online int) {
	var rates []float64
	for _, q := range quotes {
		if q.OK && q.Vs == vs && q.Rate > 0 {
			rates = append(rates, q.Rate)
		}
	}
	med := Median(rates)
	if med == 0 {
		return 0, 0
	}
	var clean []float64
	for i := range quotes {
		q := &quotes[i]
		if !q.OK || q.Vs != vs || q.Rate <= 0 {
			continue
		}
		if len(rates) >= 3 && math.Abs(q.Rate/med-1) > OutlierTolerance {
			q.OK = false
			q.Error = "outlier: deviates from the median by more than 8%"
			continue
		}
		clean = append(clean, q.Rate)
	}
	return Median(clean), len(clean)
}

// BuyPrice is what one unit costs on this source before fees.
func BuyPrice(q model.Quote) float64 {
	if q.Ask > 0 {
		return q.Ask
	}
	return q.Rate
}

// SellPrice is what one unit fetches on this source before fees.
func SellPrice(q model.Quote) float64 {
	if q.Bid > 0 {
		return q.Bid
	}
	return q.Rate
}

// ComputeBest finds the cheapest buy and the most profitable sell among
// healthy quotes for vs. fee overrides the per-quote fee when non-nil;
// include filters sources (nil means all).
func ComputeBest(quotes []model.Quote, vs string, fee func(model.Quote) float64, include func(string) bool) model.Best {
	if fee == nil {
		fee = func(q model.Quote) float64 { return q.Fee }
	}
	var b model.Best
	for _, q := range quotes {
		if !q.OK || q.Vs != vs || q.Rate <= 0 {
			continue
		}
		if include != nil && !include(q.Source) {
			continue
		}
		f := fee(q)
		buy := &model.BestPrice{Source: q.Source, Price: BuyPrice(q), Effective: BuyPrice(q) * (1 + f)}
		sell := &model.BestPrice{Source: q.Source, Price: SellPrice(q), Effective: SellPrice(q) * (1 - f)}
		if b.Buy == nil || buy.Effective < b.Buy.Effective {
			b.Buy = buy
		}
		if b.WorstBuy == nil || buy.Effective > b.WorstBuy.Effective {
			b.WorstBuy = buy
		}
		if b.Sell == nil || sell.Effective > b.Sell.Effective {
			b.Sell = sell
		}
		if b.WorstSell == nil || sell.Effective < b.WorstSell.Effective {
			b.WorstSell = sell
		}
	}
	if b.Buy != nil && b.Sell != nil && b.Buy.Effective > 0 {
		b.SpreadPct = (b.Sell.Effective - b.Buy.Effective) / b.Buy.Effective * 100
	}
	if b.Buy != nil && b.WorstBuy != nil && b.WorstBuy.Effective > 0 {
		b.SavingPct = (b.WorstBuy.Effective - b.Buy.Effective) / b.WorstBuy.Effective * 100
	}
	return b
}

// Change returns the percentage change from ref to cur.
func Change(cur, ref float64) *float64 {
	if ref <= 0 || cur <= 0 {
		return nil
	}
	v := (cur - ref) / ref * 100
	return &v
}
