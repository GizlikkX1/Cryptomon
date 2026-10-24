package market

import (
	"math"
	"testing"

	"cryptomon/internal/model"
)

func near(a, b float64) bool { return math.Abs(a-b) < 1e-9 }

func TestMedian(t *testing.T) {
	if Median(nil) != 0 {
		t.Fatal("empty median must be 0")
	}
	if got := Median([]float64{3, 1, 2}); got != 2 {
		t.Fatalf("odd median = %v", got)
	}
	if got := Median([]float64{4, 1, 2, 3}); got != 2.5 {
		t.Fatalf("even median = %v", got)
	}
}

func TestAggregateDropsOutliers(t *testing.T) {
	q := []model.Quote{
		{Source: "a", Vs: "USD", Rate: 100, OK: true},
		{Source: "b", Vs: "USD", Rate: 101, OK: true},
		{Source: "c", Vs: "USD", Rate: 99, OK: true},
		{Source: "d", Vs: "USD", Rate: 150, OK: true}, // broken feed
		{Source: "e", Vs: "EUR", Rate: 90, OK: true},  // other currency
		{Source: "f", Vs: "USD", Rate: 100, OK: false},
	}
	price, online := Aggregate(q, "USD")
	if price != 100 || online != 3 {
		t.Fatalf("price=%v online=%v, want 100/3", price, online)
	}
	if q[3].OK || q[3].Error == "" {
		t.Fatal("outlier must be marked as not OK")
	}
}

func TestComputeBestUsesFees(t *testing.T) {
	q := []model.Quote{
		// Cheapest raw ask but a high fee.
		{Source: "cheap", Vs: "USD", Bid: 99.9, Ask: 100.0, Rate: 99.95, Fee: 0.01, OK: true},
		{Source: "lowfee", Vs: "USD", Bid: 100.1, Ask: 100.2, Rate: 100.15, Fee: 0.001, OK: true},
		{Source: "down", Vs: "USD", Bid: 1, Ask: 1, Rate: 1, OK: false},
	}
	b := ComputeBest(q, "USD", nil, nil)
	if b.Buy.Source != "lowfee" || !near(b.Buy.Effective, 100.2*1.001) {
		t.Fatalf("best buy = %+v", b.Buy)
	}
	if b.Sell.Source != "lowfee" || !near(b.Sell.Effective, 100.1*0.999) {
		t.Fatalf("best sell = %+v", b.Sell)
	}
	if b.WorstBuy.Source != "cheap" || b.SavingPct <= 0 {
		t.Fatalf("worst buy = %+v saving=%v", b.WorstBuy, b.SavingPct)
	}
	if b.SpreadPct >= 0 {
		t.Fatalf("no arbitrage expected, spread=%v", b.SpreadPct)
	}

	// Excluding a source and overriding fees changes the answer.
	b = ComputeBest(q, "USD", func(model.Quote) float64 { return 0 }, func(id string) bool { return id != "lowfee" })
	if b.Buy.Source != "cheap" || b.Buy.Effective != 100 {
		t.Fatalf("filtered best buy = %+v", b.Buy)
	}
}

func TestComputeBestArbitrage(t *testing.T) {
	q := []model.Quote{
		{Source: "a", Vs: "USD", Bid: 99, Ask: 100, Rate: 99.5, OK: true},
		{Source: "b", Vs: "USD", Bid: 102, Ask: 103, Rate: 102.5, OK: true},
	}
	b := ComputeBest(q, "USD", nil, nil)
	if b.Buy.Source != "a" || b.Sell.Source != "b" || !near(b.SpreadPct, 2) {
		t.Fatalf("arbitrage = %+v", b)
	}
}

func TestReferenceRatesWithoutBook(t *testing.T) {
	q := []model.Quote{
		{Source: "cbr", Vs: "RUB", Rate: 80.5, OK: true},
		{Source: "ecb", Vs: "RUB", Rate: 81.0, OK: true},
	}
	b := ComputeBest(q, "RUB", nil, nil)
	if b.Buy.Source != "cbr" || b.Sell.Source != "ecb" {
		t.Fatalf("reference best = %+v", b)
	}
}

func TestChange(t *testing.T) {
	if c := Change(110, 100); c == nil || !near(*c, 10) {
		t.Fatalf("change = %v", c)
	}
	if Change(1, 0) != nil {
		t.Fatal("zero reference must give nil")
	}
}
