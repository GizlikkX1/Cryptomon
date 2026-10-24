package alerts

import (
	"path/filepath"
	"testing"
	"time"
)

type fakeMarket struct {
	price map[string]float64
	ref   float64
}

func newEngine(t *testing.T, m *fakeMarket) (*Engine, *[]Event) {
	t.Helper()
	e, err := Load(filepath.Join(t.TempDir(), "alerts.json"))
	if err != nil {
		t.Fatal(err)
	}
	var fired []Event
	e.Price = func(asset, vs, source string) (float64, bool) {
		p, ok := m.price[asset+vs]
		return p, ok
	}
	e.RefAt = func(asset, vs, source string, at time.Time) (float64, bool) { return m.ref, m.ref > 0 }
	e.OnEvent = func(ev Event) { fired = append(fired, ev) }
	return e, &fired
}

func TestRisePercentSinceRepeats(t *testing.T) {
	m := &fakeMarket{price: map[string]float64{"BTCUSD": 100}}
	e, fired := newEngine(t, m)
	r, err := e.Add(Rule{Asset: "btc", Vs: "usd", Condition: RisePct, Value: 5, Repeat: true})
	if err != nil {
		t.Fatal(err)
	}
	if r.Baseline != 100 || r.Asset != "BTC" || r.Window != "since" || r.Source != "agg" {
		t.Fatalf("defaults not applied: %+v", r)
	}
	now := time.Now()
	m.price["BTCUSD"] = 104
	e.Evaluate("BTC", now)
	if len(*fired) != 0 {
		t.Fatal("+4% must not fire a 5% rule")
	}
	m.price["BTCUSD"] = 105.5
	e.Evaluate("BTC", now)
	if len(*fired) != 1 || (*fired)[0].ChangePct < 5 {
		t.Fatalf("expected one event, got %+v", *fired)
	}
	// Re-armed from the new baseline: the next 5% is measured from 105.5.
	m.price["BTCUSD"] = 108
	e.Evaluate("BTC", now)
	if len(*fired) != 1 {
		t.Fatal("must wait for another 5% after re-arming")
	}
	m.price["BTCUSD"] = 111
	e.Evaluate("BTC", now)
	if len(*fired) != 2 {
		t.Fatalf("second notification expected, got %d", len(*fired))
	}
	if got := e.Rules()[0]; !got.Enabled || got.Triggers != 2 {
		t.Fatalf("rule state %+v", got)
	}
}

func TestOneShotDisables(t *testing.T) {
	m := &fakeMarket{price: map[string]float64{"ETHEUR": 2000}}
	e, fired := newEngine(t, m)
	_, _ = e.Add(Rule{Asset: "ETH", Vs: "EUR", Condition: FallAbs, Value: 100})
	m.price["ETHEUR"] = 1890
	e.Evaluate("ETH", time.Now())
	e.Evaluate("ETH", time.Now())
	if len(*fired) != 1 || e.Rules()[0].Enabled {
		t.Fatalf("one-shot rule must fire once and disable itself (%d events)", len(*fired))
	}
}

func TestAboveIsEdgeTriggered(t *testing.T) {
	m := &fakeMarket{price: map[string]float64{"USDRUB": 90}}
	e, fired := newEngine(t, m)
	_, _ = e.Add(Rule{Asset: "USD", Vs: "RUB", Condition: Above, Value: 95, Repeat: true})
	now := time.Now()
	for _, p := range []float64{94, 96, 97, 96} {
		m.price["USDRUB"] = p
		e.Evaluate("USD", now)
	}
	if len(*fired) != 1 {
		t.Fatalf("crossing once must notify once, got %d", len(*fired))
	}
	// Back below (beyond hysteresis) and up again: fires again.
	for _, p := range []float64{94.5, 95.2} {
		m.price["USDRUB"] = p
		e.Evaluate("USD", now)
	}
	if len(*fired) != 2 {
		t.Fatalf("re-crossing must notify again, got %d", len(*fired))
	}
}

func TestRollingWindowCooldown(t *testing.T) {
	m := &fakeMarket{price: map[string]float64{"TONUSD": 1.10}, ref: 1.0}
	e, fired := newEngine(t, m)
	_, _ = e.Add(Rule{Asset: "TON", Vs: "USD", Condition: MovePct, Value: 5, Window: "1h", Repeat: true})
	now := time.Now()
	e.Evaluate("TON", now)
	e.Evaluate("TON", now.Add(time.Minute))
	if len(*fired) != 1 {
		t.Fatalf("cooldown must suppress repeats within the window, got %d", len(*fired))
	}
	e.refs = map[string]cachedRef{} // drop the 30s reference cache
	e.Evaluate("TON", now.Add(61*time.Minute))
	if len(*fired) != 2 {
		t.Fatalf("after the window it may fire again, got %d", len(*fired))
	}
}

func TestValidationAndPersistence(t *testing.T) {
	m := &fakeMarket{price: map[string]float64{"BTCUSD": 100}}
	e, _ := newEngine(t, m)
	bad := []Rule{
		{Asset: "BTC", Vs: "USD", Condition: "nope", Value: 1},
		{Asset: "BTC", Vs: "USD", Condition: RisePct, Value: 0},
		{Asset: "BTC", Vs: "USD", Condition: RisePct, Value: 1, Window: "3y"},
		{Vs: "USD", Condition: RisePct, Value: 1},
	}
	for _, r := range bad {
		if _, err := e.Add(r); err == nil {
			t.Fatalf("rule %+v must be rejected", r)
		}
	}
	r, _ := e.Add(Rule{Asset: "BTC", Vs: "USD", Condition: Below, Value: 50, Note: "dip"})
	r.Value = 60
	if _, err := e.Update(r.ID, r); err != nil {
		t.Fatal(err)
	}
	again, err := Load(e.path)
	if err != nil {
		t.Fatal(err)
	}
	if rs := again.Rules(); len(rs) != 1 || rs[0].Value != 60 || rs[0].Note != "dip" {
		t.Fatalf("persisted rules = %+v", rs)
	}
	if err := e.Delete(r.ID); err != nil || len(e.Rules()) != 0 {
		t.Fatal("delete failed")
	}
}
