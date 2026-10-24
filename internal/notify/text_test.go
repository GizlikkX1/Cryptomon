package notify

import (
	"strings"
	"testing"

	"cryptomon/internal/alerts"
)

func TestNumber(t *testing.T) {
	cases := map[float64]string{
		80458.67:  "80 458.67",
		6481194.4: "6 481 194",
		1.25305:   "1.25",
		0.0124165: "0.0124",
		0.0063767: "0.006377",
		-1234.5:   "-1 234.50",
	}
	for in, want := range cases {
		if got := Number(in); got != want {
			t.Errorf("Number(%v) = %q, want %q", in, got, want)
		}
	}
}

func TestAlertText(t *testing.T) {
	ev := alerts.Event{Asset: "BTC", Vs: "RUB", Condition: alerts.RisePct, Window: "since", Value: 5, Price: 6800000, ChangePct: 5.2}
	title, body := AlertText("ru", ev, "")
	if title != "Cryptomon · BTC/RUB" {
		t.Fatalf("title %q", title)
	}
	if !strings.Contains(body, "BTC вырос на 5.20% с момента создания уведомления") || !strings.Contains(body, "Цена: 6 800 000 RUB") {
		t.Fatalf("body %q", body)
	}

	ev = alerts.Event{Asset: "ETH", Vs: "USD", Condition: alerts.FallPct, Window: "1h", Value: 3, Price: 2200, ChangePct: -3.4}
	_, body = AlertText("en", ev, "Binance")
	if !strings.HasPrefix(body, "ETH fell by 3.40% in 1 h") {
		t.Fatalf("body %q", body)
	}

	ev = alerts.Event{Asset: "USD", Vs: "RUB", Condition: alerts.Above, Value: 95, Price: 95.3}
	_, body = AlertText("zz", ev, "") // unknown language falls back to English
	if !strings.HasPrefix(body, "USD rose above 95.00 RUB") {
		t.Fatalf("body %q", body)
	}
}

func TestAllLanguagesComplete(t *testing.T) {
	for lang, d := range texts {
		for k := range texts["en"] {
			if d[k] == "" {
				t.Errorf("%s: missing %q", lang, k)
			}
		}
	}
	if len(texts) != 10 {
		t.Fatalf("expected 10 languages, have %d", len(texts))
	}
}
