package gateway

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
	"sync"

	"cryptomon/internal/alerts"
	"cryptomon/internal/catalog"
)

// Settings are the user's preferences. They live on the server so that
// they survive switching between the app window and a normal browser.
type Settings struct {
	Lang          string              `json:"lang"`  // "" = auto-detect in the UI
	Theme         string              `json:"theme"` // system, light, dark, midnight, ocean, sepia
	Scale         float64             `json:"scale"` // 0 = automatic from screen size
	Vs            string              `json:"vs"`    // display currency
	ChartStyle    string              `json:"chartStyle"`
	Range         string              `json:"range"` // default history range
	OSNotify      bool                `json:"osNotify"`
	BrowserNotify bool                `json:"browserNotify"`
	Sound         bool                `json:"sound"`
	Disabled      []string            `json:"disabled"` // sources excluded from best-rate search
	Fees          map[string]float64  `json:"fees"`     // taker fee overrides, fraction
	Watch         map[string][]string `json:"watch"`    // per-asset sources shown on the chart
}

func defaultSettings() Settings {
	return Settings{Theme: "system", Vs: "USD", ChartStyle: "area", Range: "24h",
		OSNotify: true, BrowserNotify: false, Sound: true,
		Disabled: []string{}, Fees: map[string]float64{}, Watch: map[string][]string{}}
}

type settingsStore struct {
	mu   sync.RWMutex
	path string
	s    Settings
}

func loadSettings(path string) (*settingsStore, error) {
	st := &settingsStore{path: path, s: defaultSettings()}
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return st, nil
	}
	if err != nil {
		return nil, err
	}
	s := defaultSettings()
	if err := json.Unmarshal(b, &s); err != nil {
		return st, nil // corrupt file: start with defaults
	}
	st.s = sanitize(s)
	return st, nil
}

func sanitize(s Settings) Settings {
	d := defaultSettings()
	switch s.Theme {
	case "system", "light", "dark", "midnight", "ocean", "sepia":
	default:
		s.Theme = d.Theme
	}
	if !catalog.IsFiat(strings.ToUpper(s.Vs)) {
		s.Vs = d.Vs
	}
	s.Vs = strings.ToUpper(s.Vs)
	if s.Scale != 0 && (s.Scale < 0.5 || s.Scale > 3) {
		s.Scale = 0
	}
	if s.ChartStyle != "line" {
		s.ChartStyle = "area"
	}
	if _, ok := ranges[s.Range]; !ok {
		s.Range = d.Range
	}
	if s.Disabled == nil {
		s.Disabled = []string{}
	}
	if s.Fees == nil {
		s.Fees = map[string]float64{}
	}
	for k, v := range s.Fees {
		if v < 0 || v > 0.1 {
			delete(s.Fees, k)
		}
	}
	if s.Watch == nil {
		s.Watch = map[string][]string{}
	}
	if len(s.Lang) > 8 {
		s.Lang = ""
	}
	return s
}

func (st *settingsStore) Get() Settings {
	st.mu.RLock()
	defer st.mu.RUnlock()
	return st.s
}

func (st *settingsStore) Set(s Settings) (Settings, error) {
	s = sanitize(s)
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return s, err
	}
	st.mu.Lock()
	defer st.mu.Unlock()
	st.s = s
	return s, alerts.WriteFileAtomic(st.path, b)
}

func (s Settings) disabled(id string) bool {
	for _, d := range s.Disabled {
		if d == id {
			return true
		}
	}
	return false
}
