package gateway

import (
	"math"
	"testing"

	"cryptomon/internal/model"
)

func TestConvertUsesHistoricalRates(t *testing.T) {
	prices := []model.Point{{T: 10, V: 2}, {T: 20, V: 3}, {T: 30, V: 4}}
	rates := []model.Point{{T: 5, V: 100}, {T: 25, V: 90}}
	got := convert(prices, rates)
	want := []float64{200, 300, 360}
	for i, p := range got {
		if p.V != want[i] || p.T != prices[i].T {
			t.Fatalf("point %d = %v, want %v", i, p, want[i])
		}
	}
	// Points before the first rate use the earliest rate.
	if got := convert([]model.Point{{T: 1, V: 1}}, rates); got[0].V != 100 {
		t.Fatalf("early point = %v", got[0])
	}
}

func TestEffectiveVs(t *testing.T) {
	cases := map[[2]string]string{
		{"BTC", "RUB"}: "RUB",
		{"USD", "USD"}: "EUR",
		{"EUR", "EUR"}: "USD",
		{"ETH", "XXX"}: "USD",
	}
	for in, want := range cases {
		if got := effectiveVs(in[0], in[1]); got != want {
			t.Fatalf("effectiveVs(%v) = %s, want %s", in, got, want)
		}
	}
}

func TestCombineChanges(t *testing.T) {
	a, b := 10.0, -10.0
	got := combine(&a, &b)
	if math.Abs(*got-(-1)) > 1e-9 {
		t.Fatalf("+10%% then -10%% = %v, want -1%%", *got)
	}
	if combine(nil, &b) != nil {
		t.Fatal("unknown base change must stay unknown")
	}
	if got := combine(&a, nil); *got != 10 {
		t.Fatal("unknown FX change falls back to the base change")
	}
}

func TestSanitizeSettings(t *testing.T) {
	s := sanitize(Settings{Theme: "neon", Vs: "rub", Scale: 9, Range: "2w", Fees: map[string]float64{"x": 5, "y": 0.001}})
	if s.Theme != "system" || s.Vs != "RUB" || s.Scale != 0 || s.Range != "24h" {
		t.Fatalf("sanitize = %+v", s)
	}
	if _, ok := s.Fees["x"]; ok || s.Fees["y"] != 0.001 {
		t.Fatalf("fees = %v", s.Fees)
	}
}
