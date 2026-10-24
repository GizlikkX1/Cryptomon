package sources

import (
	"math"
	"testing"
)

func TestTableCross(t *testing.T) {
	// ECB style: EUR base.
	tb := Table{Base: "EUR", Rates: map[string]float64{"USD": 1.12, "JPY": 175.5}}
	if v, ok := tb.Cross("USD", "JPY"); !ok || math.Abs(v-175.5/1.12) > 1e-9 {
		t.Fatalf("USD→JPY = %v", v)
	}
	if v, ok := tb.Cross("EUR", "USD"); !ok || v != 1.12 {
		t.Fatalf("EUR→USD = %v", v)
	}
	if _, ok := tb.Cross("USD", "RUB"); ok {
		t.Fatal("missing currency must not cross")
	}
}

func TestTickerInvert(t *testing.T) {
	chg := 1.0
	inv := Ticker{Bid: 1.0010, Ask: 1.0012, Last: 1.0011, Change24h: &chg}.Invert()
	if math.Abs(inv.Bid-1/1.0012) > 1e-12 || math.Abs(inv.Ask-1/1.0010) > 1e-12 {
		t.Fatalf("inverted book %+v", inv)
	}
	if inv.Bid >= inv.Ask {
		t.Fatal("inverted bid must stay below ask")
	}
	if *inv.Change24h >= 0 {
		t.Fatal("a rising pair falls when inverted")
	}
}

func TestDecodeLegacyXML(t *testing.T) {
	// windows-1251 bytes in a field we do not need must not break parsing.
	body := []byte("<?xml version=\"1.0\" encoding=\"windows-1251\"?><ValCurs Date=\"24.10.2026\">" +
		"<Valute><CharCode>USD</CharCode><Nominal>1</Nominal><Name>\xc4\xee\xeb\xeb\xe0\xf0</Name><Value>80,5381</Value></Valute></ValCurs>")
	var r struct {
		Date   string `xml:"Date,attr"`
		Valute []struct {
			Code  string `xml:"CharCode"`
			Value string `xml:"Value"`
		} `xml:"Valute"`
	}
	if err := DecodeLegacyXML(body, &r); err != nil {
		t.Fatal(err)
	}
	if r.Date != "24.10.2026" || len(r.Valute) != 1 || r.Valute[0].Code != "USD" || r.Valute[0].Value != "80,5381" {
		t.Fatalf("decoded %+v", r)
	}
}

func TestValidate(t *testing.T) {
	if _, err := validate(Ticker{Bid: 0, Ask: 0, Last: 5}); err == nil {
		t.Fatal("empty book must be rejected (halted market)")
	}
	tk, err := validate(Ticker{Bid: 1, Ask: 3})
	if err != nil || tk.Last != 2 || tk.Time == 0 {
		t.Fatalf("validate filled %+v %v", tk, err)
	}
}

func TestEverySupportedFiatHasSources(t *testing.T) {
	for _, code := range []string{"USD", "EUR", "RUB", "CNY", "JPY", "INR", "CHF", "GBP"} {
		n := 0
		for _, s := range FXSources {
			if s.Supports(code) {
				n++
			}
		}
		if n < 4 {
			t.Fatalf("%s has only %d sources", code, n)
		}
	}
}
