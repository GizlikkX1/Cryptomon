// Package alerts evaluates user-defined price notifications such as
// "BTC rose by 5% since I set this" or "USD/RUB above 95".
package alerts

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Conditions.
const (
	RisePct = "rise_pct"
	FallPct = "fall_pct"
	MovePct = "move_pct"
	RiseAbs = "rise_abs"
	FallAbs = "fall_abs"
	Above   = "above"
	Below   = "below"
)

// Windows: "since" compares against the price when the rule was created
// (or last triggered, for repeating rules); the others are rolling.
var Windows = map[string]time.Duration{
	"since": 0,
	"5m":    5 * time.Minute,
	"15m":   15 * time.Minute,
	"1h":    time.Hour,
	"4h":    4 * time.Hour,
	"24h":   24 * time.Hour,
	"7d":    7 * 24 * time.Hour,
}

type Rule struct {
	ID        string  `json:"id"`
	Asset     string  `json:"asset"`
	Vs        string  `json:"vs"`
	Source    string  `json:"source"` // "agg" (aggregated price) or a source id
	Condition string  `json:"condition"`
	Value     float64 `json:"value"`
	Window    string  `json:"window"`
	Repeat    bool    `json:"repeat"`
	Enabled   bool    `json:"enabled"`
	Note      string  `json:"note,omitempty"`

	Baseline      float64  `json:"baseline"`
	BaselineTime  int64    `json:"baselineTime"`
	Armed         bool     `json:"armed"`
	CooldownUntil int64    `json:"cooldownUntil"`
	CreatedAt     int64    `json:"createdAt"`
	LastTriggered int64    `json:"lastTriggered"`
	Triggers      int      `json:"triggers"`
	LastPrice     float64  `json:"lastPrice"`
	LastChange    *float64 `json:"lastChange,omitempty"`

	rev int
}

type Event struct {
	ID        string  `json:"id"`
	RuleID    string  `json:"ruleId"`
	Asset     string  `json:"asset"`
	Vs        string  `json:"vs"`
	Source    string  `json:"source"`
	Condition string  `json:"condition"`
	Window    string  `json:"window"`
	Value     float64 `json:"value"`
	Price     float64 `json:"price"`
	Ref       float64 `json:"ref"`
	ChangePct float64 `json:"changePct"`
	ChangeAbs float64 `json:"changeAbs"`
	Note      string  `json:"note,omitempty"`
	Time      int64   `json:"time"`
}

// PriceFunc returns the current price of asset in vs on source.
type PriceFunc func(asset, vs, source string) (float64, bool)

// RefFunc returns the price at time t.
type RefFunc func(asset, vs, source string, t time.Time) (float64, bool)

type Engine struct {
	mu     sync.Mutex
	path   string
	rules  []*Rule
	events []Event

	Price   PriceFunc
	RefAt   RefFunc
	OnEvent func(Event)

	refs map[string]cachedRef
}

type cachedRef struct {
	at    time.Time
	value float64
	ok    bool
}

const maxEvents = 300

type fileFormat struct {
	Rules  []*Rule `json:"rules"`
	Events []Event `json:"events"`
}

// Load reads rules from path; a missing file is not an error.
func Load(path string) (*Engine, error) {
	e := &Engine{path: path, refs: map[string]cachedRef{}}
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return e, nil
	}
	if err != nil {
		return nil, err
	}
	var f fileFormat
	if err := json.Unmarshal(b, &f); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	e.rules, e.events = f.Rules, f.Events
	return e, nil
}

func (e *Engine) saveLocked() error {
	f := fileFormat{Rules: e.rules, Events: e.events}
	if f.Rules == nil {
		f.Rules = []*Rule{}
	}
	if f.Events == nil {
		f.Events = []Event{}
	}
	b, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return err
	}
	return WriteFileAtomic(e.path, b)
}

// WriteFileAtomic writes via a temporary file and rename.
func WriteFileAtomic(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func newID() string {
	var b [6]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}

func (e *Engine) Rules() []Rule {
	e.mu.Lock()
	defer e.mu.Unlock()
	out := make([]Rule, len(e.rules))
	for i, r := range e.rules {
		out[i] = *r
	}
	return out
}

func (e *Engine) Events() []Event {
	e.mu.Lock()
	defer e.mu.Unlock()
	out := make([]Event, len(e.events))
	copy(out, e.events)
	return out
}

func (e *Engine) ClearEvents() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.events = nil
	return e.saveLocked()
}

func validate(r *Rule) error {
	r.Asset = strings.ToUpper(strings.TrimSpace(r.Asset))
	r.Vs = strings.ToUpper(strings.TrimSpace(r.Vs))
	if r.Source == "" {
		r.Source = "agg"
	}
	if r.Window == "" {
		r.Window = "since"
	}
	if r.Asset == "" || r.Vs == "" {
		return errors.New("asset and vs are required")
	}
	switch r.Condition {
	case RisePct, FallPct, MovePct, RiseAbs, FallAbs, Above, Below:
	default:
		return fmt.Errorf("unknown condition %q", r.Condition)
	}
	if _, ok := Windows[r.Window]; !ok {
		return fmt.Errorf("unknown window %q", r.Window)
	}
	if !(r.Value > 0) || math.IsInf(r.Value, 0) {
		return errors.New("value must be a positive number")
	}
	if len(r.Note) > 200 {
		r.Note = r.Note[:200]
	}
	return nil
}

// Add creates a rule; percentage/amount rules with the "since" window take
// the current price as their baseline.
func (e *Engine) Add(r Rule) (Rule, error) {
	if err := validate(&r); err != nil {
		return Rule{}, err
	}
	now := time.Now().UnixMilli()
	r.ID, r.CreatedAt, r.Enabled, r.Armed = newID(), now, true, true
	r.Triggers, r.LastTriggered, r.CooldownUntil, r.LastChange = 0, 0, 0, nil
	r.Baseline, r.BaselineTime = 0, 0
	if p, ok := e.price(r); ok {
		r.Baseline, r.BaselineTime, r.LastPrice = p, now, p
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	e.rules = append(e.rules, &r)
	return r, e.saveLocked()
}

// Update replaces the editable fields of a rule. Changing what is
// measured, or re-enabling it, resets its baseline.
func (e *Engine) Update(id string, in Rule) (Rule, error) {
	if err := validate(&in); err != nil {
		return Rule{}, err
	}
	p, havePrice := e.price(in)
	e.mu.Lock()
	defer e.mu.Unlock()
	for _, r := range e.rules {
		if r.ID != id {
			continue
		}
		reset := r.Asset != in.Asset || r.Vs != in.Vs || r.Source != in.Source ||
			r.Condition != in.Condition || r.Window != in.Window || r.Value != in.Value ||
			(!r.Enabled && in.Enabled)
		r.Asset, r.Vs, r.Source, r.Condition = in.Asset, in.Vs, in.Source, in.Condition
		r.Value, r.Window, r.Repeat, r.Enabled, r.Note = in.Value, in.Window, in.Repeat, in.Enabled, in.Note
		if reset {
			r.Armed, r.CooldownUntil, r.LastChange = true, 0, nil
			r.Baseline, r.BaselineTime = 0, 0
			if havePrice {
				r.Baseline, r.BaselineTime = p, time.Now().UnixMilli()
			}
		}
		r.rev++
		return *r, e.saveLocked()
	}
	return Rule{}, os.ErrNotExist
}

func (e *Engine) Delete(id string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	for i, r := range e.rules {
		if r.ID == id {
			e.rules = append(e.rules[:i], e.rules[i+1:]...)
			return e.saveLocked()
		}
	}
	return os.ErrNotExist
}

func (e *Engine) price(r Rule) (float64, bool) {
	if e.Price == nil {
		return 0, false
	}
	return e.Price(r.Asset, r.Vs, r.Source)
}

func (e *Engine) refAt(r Rule, window time.Duration, now time.Time) (float64, bool) {
	k := r.Asset + "|" + r.Vs + "|" + r.Source + "|" + r.Window
	if c, ok := e.refs[k]; ok && now.Sub(c.at) < 30*time.Second {
		return c.value, c.ok
	}
	var v float64
	var ok bool
	if e.RefAt != nil {
		v, ok = e.RefAt(r.Asset, r.Vs, r.Source, now.Add(-window))
	}
	e.refs[k] = cachedRef{at: now, value: v, ok: ok}
	return v, ok
}

// Evaluate checks every enabled rule for asset ("" = all). It must be
// called from a single goroutine; price lookups happen outside the lock.
func (e *Engine) Evaluate(asset string, now time.Time) {
	e.mu.Lock()
	var todo []Rule
	for _, r := range e.rules {
		if r.Enabled && (asset == "" || r.Asset == asset) {
			todo = append(todo, *r)
		}
	}
	e.mu.Unlock()

	var fired []Event
	changed := false
	for _, snapshot := range todo {
		cur, ok := e.price(snapshot)
		if !ok || cur <= 0 {
			continue
		}
		var ref float64
		window := Windows[snapshot.Window]
		isLevel := snapshot.Condition == Above || snapshot.Condition == Below
		if !isLevel && window > 0 {
			if now.UnixMilli() < snapshot.CooldownUntil {
				ref = -1 // still show progress below
			} else if ref, ok = e.refAt(snapshot, window, now); !ok {
				continue
			}
		}

		e.mu.Lock()
		r := e.find(snapshot.ID)
		if r == nil || r.rev != snapshot.rev || !r.Enabled {
			e.mu.Unlock()
			continue
		}
		ev, hit := step(r, cur, ref, now)
		if hit {
			fired = append(fired, ev)
			e.events = append([]Event{ev}, e.events...)
			if len(e.events) > maxEvents {
				e.events = e.events[:maxEvents]
			}
			changed = true
		}
		e.mu.Unlock()
	}
	if changed {
		e.mu.Lock()
		_ = e.saveLocked()
		e.mu.Unlock()
	}
	if e.OnEvent != nil {
		for _, ev := range fired {
			e.OnEvent(ev)
		}
	}
}

func (e *Engine) find(id string) *Rule {
	for _, r := range e.rules {
		if r.ID == id {
			return r
		}
	}
	return nil
}

// step advances one rule with the current price; ref is the rolling-window
// reference (0 for "since" rules, -1 during cooldown). Caller holds the lock.
func step(r *Rule, cur, ref float64, now time.Time) (Event, bool) {
	nowMs := now.UnixMilli()
	r.LastPrice = cur
	ev := Event{ID: newID(), RuleID: r.ID, Asset: r.Asset, Vs: r.Vs, Source: r.Source, Condition: r.Condition,
		Window: r.Window, Value: r.Value, Price: cur, Note: r.Note, Time: nowMs}

	trigger := func() (Event, bool) {
		r.LastTriggered = nowMs
		r.Triggers++
		if !r.Repeat {
			r.Enabled = false
		}
		r.rev++
		return ev, true
	}

	switch r.Condition {
	case Above, Below:
		hit := (r.Condition == Above && cur >= r.Value) || (r.Condition == Below && cur <= r.Value)
		ev.Ref = r.Value
		ev.ChangeAbs = cur - r.Value
		ev.ChangePct = (cur - r.Value) / r.Value * 100
		chg := ev.ChangePct
		r.LastChange = &chg
		if hit && r.Armed {
			r.Armed = false
			return trigger()
		}
		// Re-arm with a small hysteresis once the price is back across.
		if !hit && !r.Armed &&
			((r.Condition == Above && cur < r.Value*0.999) || (r.Condition == Below && cur > r.Value*1.001)) {
			r.Armed = true
		}
		return ev, false
	}

	if r.Window == "since" {
		if r.Baseline <= 0 {
			r.Baseline, r.BaselineTime = cur, nowMs
			return ev, false
		}
		ref = r.Baseline
	}
	if ref <= 0 {
		return ev, false
	}
	chg := (cur - ref) / ref * 100
	abs := cur - ref
	r.LastChange = &chg
	ev.Ref, ev.ChangePct, ev.ChangeAbs = ref, chg, abs

	var hit bool
	switch r.Condition {
	case RisePct:
		hit = chg >= r.Value
	case FallPct:
		hit = chg <= -r.Value
	case MovePct:
		hit = math.Abs(chg) >= r.Value
	case RiseAbs:
		hit = abs >= r.Value
	case FallAbs:
		hit = -abs >= r.Value
	}
	if !hit {
		return ev, false
	}
	if r.Window == "since" {
		// Repeating rules notify again after each further move of Value.
		r.Baseline, r.BaselineTime = cur, nowMs
	} else {
		r.CooldownUntil = nowMs + Windows[r.Window].Milliseconds()
	}
	return trigger()
}
