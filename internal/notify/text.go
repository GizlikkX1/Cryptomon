package notify

import (
	"fmt"
	"math"
	"strings"

	"cryptomon/internal/alerts"
)

// Texts for native notifications. The UI has its own, larger dictionary;
// only what the OS notification needs lives here.
var texts = map[string]map[string]string{
	"en": {"up": "{asset} rose by {chg} {window}", "down": "{asset} fell by {chg} {window}", "move": "{asset} changed by {chg} {window}",
		"above": "{asset} rose above {level}", "below": "{asset} fell below {level}", "since": "since the alert was set", "in": "in {w}",
		"price": "Price", "min": "min", "h": "h", "d": "d", "test": "Notifications are working"},
	"ru": {"up": "{asset} вырос на {chg} {window}", "down": "{asset} упал на {chg} {window}", "move": "{asset} изменился на {chg} {window}",
		"above": "{asset} поднялся выше {level}", "below": "{asset} опустился ниже {level}", "since": "с момента создания уведомления", "in": "за {w}",
		"price": "Цена", "min": "мин", "h": "ч", "d": "дн", "test": "Уведомления работают"},
	"zh": {"up": "{asset} {window}上涨了 {chg}", "down": "{asset} {window}下跌了 {chg}", "move": "{asset} {window}变动了 {chg}",
		"above": "{asset} 涨破 {level}", "below": "{asset} 跌破 {level}", "since": "自设置提醒以来", "in": "在{w}内",
		"price": "价格", "min": "分钟", "h": "小时", "d": "天", "test": "通知功能正常"},
	"hi": {"up": "{asset} {window} {chg} बढ़ा", "down": "{asset} {window} {chg} गिरा", "move": "{asset} {window} {chg} बदला",
		"above": "{asset} {level} से ऊपर गया", "below": "{asset} {level} से नीचे गया", "since": "अलर्ट सेट करने के बाद से", "in": "{w} में",
		"price": "मूल्य", "min": "मिनट", "h": "घंटे", "d": "दिन", "test": "सूचनाएँ काम कर रही हैं"},
	"es": {"up": "{asset} subió {chg} {window}", "down": "{asset} bajó {chg} {window}", "move": "{asset} cambió {chg} {window}",
		"above": "{asset} superó {level}", "below": "{asset} cayó por debajo de {level}", "since": "desde que se creó la alerta", "in": "en {w}",
		"price": "Precio", "min": "min", "h": "h", "d": "d", "test": "Las notificaciones funcionan"},
	"fr": {"up": "{asset} a augmenté de {chg} {window}", "down": "{asset} a baissé de {chg} {window}", "move": "{asset} a varié de {chg} {window}",
		"above": "{asset} est passé au-dessus de {level}", "below": "{asset} est passé sous {level}", "since": "depuis la création de l'alerte", "in": "en {w}",
		"price": "Prix", "min": "min", "h": "h", "d": "j", "test": "Les notifications fonctionnent"},
	"ar": {"up": "ارتفع {asset} بمقدار {chg} {window}", "down": "انخفض {asset} بمقدار {chg} {window}", "move": "تغيّر {asset} بمقدار {chg} {window}",
		"above": "ارتفع {asset} فوق {level}", "below": "انخفض {asset} دون {level}", "since": "منذ إنشاء التنبيه", "in": "خلال {w}",
		"price": "السعر", "min": "دقيقة", "h": "ساعة", "d": "يوم", "test": "الإشعارات تعمل"},
	"bn": {"up": "{asset} {window} {chg} বেড়েছে", "down": "{asset} {window} {chg} কমেছে", "move": "{asset} {window} {chg} পরিবর্তিত হয়েছে",
		"above": "{asset} {level}-এর উপরে উঠেছে", "below": "{asset} {level}-এর নিচে নেমেছে", "since": "অ্যালার্ট সেট করার পর থেকে", "in": "{w}-এ",
		"price": "মূল্য", "min": "মিনিট", "h": "ঘণ্টা", "d": "দিন", "test": "বিজ্ঞপ্তি কাজ করছে"},
	"pt": {"up": "{asset} subiu {chg} {window}", "down": "{asset} caiu {chg} {window}", "move": "{asset} variou {chg} {window}",
		"above": "{asset} subiu acima de {level}", "below": "{asset} caiu abaixo de {level}", "since": "desde a criação do alerta", "in": "em {w}",
		"price": "Preço", "min": "min", "h": "h", "d": "d", "test": "As notificações estão funcionando"},
	"ur": {"up": "{asset} میں {window} {chg} اضافہ ہوا", "down": "{asset} میں {window} {chg} کمی ہوئی", "move": "{asset} میں {window} {chg} تبدیلی ہوئی",
		"above": "{asset} {level} سے اوپر چلا گیا", "below": "{asset} {level} سے نیچے آ گیا", "since": "الرٹ سیٹ کرنے کے بعد سے", "in": "{w} میں",
		"price": "قیمت", "min": "منٹ", "h": "گھنٹے", "d": "دن", "test": "اطلاعات کام کر رہی ہیں"},
}

func dict(lang string) map[string]string {
	if d, ok := texts[lang]; ok {
		return d
	}
	return texts["en"]
}

func fill(s string, kv ...string) string {
	return strings.NewReplacer(kv...).Replace(s)
}

// TestText is the body of the "send test notification" button.
func TestText(lang string) string { return dict(lang)["test"] }

// AlertText renders an alert event as a notification title and body.
// sourceName is the human-readable source ("" for the aggregate price).
func AlertText(lang string, ev alerts.Event, sourceName string) (string, string) {
	d := dict(lang)
	asset := ev.Asset
	title := "Cryptomon · " + ev.Asset + "/" + ev.Vs
	if sourceName != "" {
		title += " · " + sourceName
	}
	window := d["since"]
	if ev.Window != "since" {
		window = fill(d["in"], "{w}", windowText(d, ev.Window))
	}
	var body string
	switch ev.Condition {
	case alerts.Above, alerts.Below:
		key := "above"
		if ev.Condition == alerts.Below {
			key = "below"
		}
		body = fill(d[key], "{asset}", asset, "{level}", Number(ev.Value)+" "+ev.Vs)
	default:
		// The verb ("rose"/"fell") carries the direction: unsigned numbers.
		pct := math.Abs(ev.ChangePct)
		chg := fmt.Sprintf("%.2f%%", pct)
		if pct < 0.1 {
			chg = fmt.Sprintf("%.4f%%", pct)
		}
		if ev.Condition == alerts.RiseAbs || ev.Condition == alerts.FallAbs {
			chg = Number(math.Abs(ev.ChangeAbs)) + " " + ev.Vs
		}
		key := "move"
		switch {
		case ev.Condition == alerts.RisePct || ev.Condition == alerts.RiseAbs || (ev.Condition == alerts.MovePct && ev.ChangePct > 0):
			key = "up"
		case ev.Condition == alerts.FallPct || ev.Condition == alerts.FallAbs || (ev.Condition == alerts.MovePct && ev.ChangePct < 0):
			key = "down"
		}
		body = fill(d[key], "{asset}", asset, "{chg}", chg, "{window}", window)
	}
	body = strings.Join(strings.Fields(body), " ") + "\n" + d["price"] + ": " + Number(ev.Price) + " " + ev.Vs
	if ev.Note != "" {
		body += "\n" + ev.Note
	}
	return title, body
}

func windowText(d map[string]string, w string) string {
	if len(w) < 2 {
		return w
	}
	num, unit := w[:len(w)-1], w[len(w)-1:]
	switch unit {
	case "m":
		return num + " " + d["min"]
	case "h":
		return num + " " + d["h"]
	case "d":
		return num + " " + d["d"]
	}
	return w
}

// Number formats a price with grouping and a precision suited to its size.
func Number(v float64) string {
	a := math.Abs(v)
	dec := 2
	switch {
	case a == 0:
	case a < 0.01:
		dec = 6
	case a < 1:
		dec = 4
	case a >= 100000:
		dec = 0
	}
	s := fmt.Sprintf("%.*f", dec, a)
	intPart, frac, _ := strings.Cut(s, ".")
	var b strings.Builder
	for i, c := range intPart {
		if i > 0 && (len(intPart)-i)%3 == 0 {
			b.WriteRune(' ')
		}
		b.WriteRune(c)
	}
	out := b.String()
	if frac != "" {
		out += "." + frac
	}
	if v < 0 {
		out = "-" + out
	}
	return out
}
