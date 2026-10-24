// Package httpx contains the small HTTP helpers shared by all adapters.
package httpx

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const UserAgent = "Mozilla/5.0 (compatible; Cryptomon/1.0; +https://github.com/)"

var Client = &http.Client{
	Timeout: 15 * time.Second,
	Transport: &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           (&net.Dialer{Timeout: 7 * time.Second, KeepAlive: 30 * time.Second}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          64,
		MaxIdleConnsPerHost:   4,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   7 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
	},
}

// StatusError is returned for non-2xx responses.
type StatusError struct {
	Code int
	Body string
}

func (e *StatusError) Error() string {
	if e.Body == "" {
		return fmt.Sprintf("HTTP %d", e.Code)
	}
	return fmt.Sprintf("HTTP %d: %s", e.Code, e.Body)
}

// Get performs a GET request and returns the body of a 2xx response.
func Get(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Accept", "application/json, text/xml, */*")
	resp, err := Client.Do(req)
	if err != nil {
		return nil, shorten(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 16<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, &StatusError{Code: resp.StatusCode, Body: snippet(body)}
	}
	return body, nil
}

// GetJSON performs a GET request and decodes a JSON response into v.
func GetJSON(ctx context.Context, url string, v any) error {
	body, err := Get(ctx, url)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(body, v); err != nil {
		return fmt.Errorf("bad JSON: %v (%s)", err, snippet(body))
	}
	return nil
}

// F parses a decimal number leniently: empty strings become 0 and a comma
// decimal separator (used by the Bank of Russia) is accepted.
func F(s string) float64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	s = strings.ReplaceAll(s, ",", ".")
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return v
}

func snippet(b []byte) string {
	s := strings.TrimSpace(string(b))
	s = strings.Join(strings.Fields(s), " ")
	if len(s) > 160 {
		s = s[:160] + "…"
	}
	return s
}

// shorten strips the noisy "Get \"https://...\":" prefix from net errors.
func shorten(err error) error {
	msg := err.Error()
	if i := strings.LastIndex(msg, "\": "); i >= 0 {
		msg = msg[i+3:]
	}
	return fmt.Errorf("%s", msg)
}
