// Package catalog is the static description of every tracked currency:
// which exchanges list it, under which ticker, and which port its
// microservice listens on by default.
package catalog

import (
	"fmt"
	"sort"
	"strings"

	"cryptomon/internal/model"
)

// Exchange is a crypto exchange with its default spot taker fee.
// Fees are approximate base-tier values and can be overridden in the UI.
type Exchange struct {
	ID   string
	Name string
	URL  string
	Fee  float64
}

var Exchanges = map[string]Exchange{
	"binance":  {"binance", "Binance", "https://www.binance.com", 0.001},
	"bybit":    {"bybit", "Bybit", "https://www.bybit.com", 0.001},
	"okx":      {"okx", "OKX", "https://www.okx.com", 0.001},
	"kraken":   {"kraken", "Kraken", "https://www.kraken.com", 0.004},
	"coinbase": {"coinbase", "Coinbase", "https://www.coinbase.com", 0.006},
	"kucoin":   {"kucoin", "KuCoin", "https://www.kucoin.com", 0.001},
	"gate":     {"gate", "Gate", "https://www.gate.com", 0.002},
	"htx":      {"htx", "HTX", "https://www.htx.com", 0.002},
	"bitfinex": {"bitfinex", "Bitfinex", "https://www.bitfinex.com", 0.002},
	"mexc":     {"mexc", "MEXC", "https://www.mexc.com", 0.0005},
	"bitget":   {"bitget", "Bitget", "https://www.bitget.com", 0.001},
	"bitstamp": {"bitstamp", "Bitstamp", "https://www.bitstamp.net", 0.004},
}

// Market is one tradable pair of a crypto asset on an exchange.
type Market struct {
	Exchange string
	Symbol   string // ticker exactly as the exchange API expects it
	Native   string // quote currency of the pair: USD, USDT or USDC
	// Inverse means the pair is quoted the other way round
	// (e.g. USDC/USDT used to price USDT) and must be inverted.
	Inverse bool
}

type Asset struct {
	Code   string
	Name   string
	Symbol string
	Color  string
	Kind   model.Kind
	Port   int

	Markets []Market // crypto only

	// History backfill sources (crypto only).
	KrakenOHLC    string
	BinanceKlines string
}

// FiatCodes are the fiat currencies that get their own service and can be
// used as the display currency.
var FiatCodes = []string{"USD", "EUR", "RUB", "CNY", "JPY", "INR", "CHF", "GBP"}

var Assets = []Asset{
	{
		Code: "BTC", Name: "Bitcoin", Symbol: "₿", Color: "#f7931a", Kind: model.Crypto, Port: 9101,
		KrakenOHLC: "XBTUSD", BinanceKlines: "BTCUSDT",
		Markets: []Market{
			{"binance", "BTCUSDT", "USDT", false},
			{"bybit", "BTCUSDT", "USDT", false},
			{"okx", "BTC-USDT", "USDT", false},
			{"kraken", "XBTUSD", "USD", false},
			{"coinbase", "BTC-USD", "USD", false},
			{"kucoin", "BTC-USDT", "USDT", false},
			{"gate", "BTC_USDT", "USDT", false},
			{"htx", "btcusdt", "USDT", false},
			{"bitfinex", "tBTCUSD", "USD", false},
			{"mexc", "BTCUSDT", "USDT", false},
			{"bitget", "BTCUSDT", "USDT", false},
			{"bitstamp", "btcusd", "USD", false},
		},
	},
	{
		Code: "ETH", Name: "Ethereum", Symbol: "Ξ", Color: "#627eea", Kind: model.Crypto, Port: 9102,
		KrakenOHLC: "ETHUSD", BinanceKlines: "ETHUSDT",
		Markets: []Market{
			{"binance", "ETHUSDT", "USDT", false},
			{"bybit", "ETHUSDT", "USDT", false},
			{"okx", "ETH-USDT", "USDT", false},
			{"kraken", "ETHUSD", "USD", false},
			{"coinbase", "ETH-USD", "USD", false},
			{"kucoin", "ETH-USDT", "USDT", false},
			{"gate", "ETH_USDT", "USDT", false},
			{"htx", "ethusdt", "USDT", false},
			{"bitfinex", "tETHUSD", "USD", false},
			{"mexc", "ETHUSDT", "USDT", false},
			{"bitget", "ETHUSDT", "USDT", false},
			{"bitstamp", "ethusd", "USD", false},
		},
	},
	{
		// USDT is priced in USD directly where a USD pair exists, otherwise
		// through the USDC/USDT book (USDC is treated as 1 USD).
		Code: "USDT", Name: "Tether USDt", Symbol: "₮", Color: "#26a17b", Kind: model.Crypto, Port: 9103,
		KrakenOHLC: "USDTUSD", BinanceKlines: "",
		Markets: []Market{
			{"kraken", "USDTUSD", "USD", false},
			{"coinbase", "USDT-USD", "USD", false},
			{"bitstamp", "usdtusd", "USD", false},
			{"bitfinex", "tUSTUSD", "USD", false},
			{"binance", "USDCUSDT", "USDC", true},
			{"bybit", "USDCUSDT", "USDC", true},
			{"okx", "USDC-USDT", "USDC", true},
			{"kucoin", "USDC-USDT", "USDC", true},
			{"gate", "USDC_USDT", "USDC", true},
			{"htx", "usdcusdt", "USDC", true},
			{"mexc", "USDCUSDT", "USDC", true},
			{"bitget", "USDCUSDT", "USDC", true},
		},
	},
	{
		// Toncoin trades as GRAM on most exchanges after the rebrand;
		// Kraken, Coinbase, Bitfinex and Bitstamp still use TON.
		Code: "TON", Name: "Toncoin (Gram)", Symbol: "◆", Color: "#0098ea", Kind: model.Crypto, Port: 9104,
		KrakenOHLC: "TONUSD", BinanceKlines: "GRAMUSDT",
		Markets: []Market{
			{"binance", "GRAMUSDT", "USDT", false},
			{"bybit", "GRAMUSDT", "USDT", false},
			{"okx", "GRAM-USDT", "USDT", false},
			{"kraken", "TONUSD", "USD", false},
			{"coinbase", "TON-USD", "USD", false},
			{"kucoin", "GRAM-USDT", "USDT", false},
			{"gate", "GRAM_USDT", "USDT", false},
			{"htx", "gramusdt", "USDT", false},
			{"bitfinex", "tTONUSD", "USD", false},
			{"mexc", "GRAMUSDT", "USDT", false},
			{"bitget", "GRAMUSDT", "USDT", false},
			{"bitstamp", "tonusd", "USD", false},
		},
	},
	{Code: "USD", Name: "US Dollar", Symbol: "$", Color: "#2e7d32", Kind: model.Fiat, Port: 9111},
	{Code: "EUR", Name: "Euro", Symbol: "€", Color: "#1e4fd8", Kind: model.Fiat, Port: 9112},
	{Code: "RUB", Name: "Russian Ruble", Symbol: "₽", Color: "#c62828", Kind: model.Fiat, Port: 9113},
	{Code: "CNY", Name: "Chinese Yuan", Symbol: "元", Color: "#de2910", Kind: model.Fiat, Port: 9114},
	{Code: "JPY", Name: "Japanese Yen", Symbol: "円", Color: "#bc002d", Kind: model.Fiat, Port: 9115},
	{Code: "INR", Name: "Indian Rupee", Symbol: "₹", Color: "#ff9933", Kind: model.Fiat, Port: 9116},
	{Code: "CHF", Name: "Swiss Franc", Symbol: "₣", Color: "#d52b1e", Kind: model.Fiat, Port: 9117},
	{Code: "GBP", Name: "British Pound", Symbol: "£", Color: "#5b2c83", Kind: model.Fiat, Port: 9118},
}

// Find returns the asset with the given code (case-insensitive).
func Find(code string) (Asset, error) {
	code = strings.ToUpper(strings.TrimSpace(code))
	for _, a := range Assets {
		if a.Code == code {
			return a, nil
		}
	}
	codes := make([]string, 0, len(Assets))
	for _, a := range Assets {
		codes = append(codes, a.Code)
	}
	sort.Strings(codes)
	return Asset{}, fmt.Errorf("unknown asset %q (known: %s)", code, strings.Join(codes, ", "))
}

func IsFiat(code string) bool {
	for _, c := range FiatCodes {
		if c == code {
			return true
		}
	}
	return false
}
