// Package app wires command-line flags to the service, gateway and
// launcher, so that several binaries (cryptomon, currency-service,
// gateway) can share the same entry points.
package app

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"cryptomon/internal/catalog"
	"cryptomon/internal/gateway"
	"cryptomon/internal/service"
)

var Version = "1.0.0"

// DefaultDataDir is the per-user data directory on every OS
// (~/.config/cryptomon, ~/Library/Application Support/cryptomon,
// %AppData%\cryptomon).
func DefaultDataDir() string {
	if d := os.Getenv("CRYPTOMON_DATA"); d != "" {
		return d
	}
	if d, err := os.UserConfigDir(); err == nil {
		return filepath.Join(d, "cryptomon")
	}
	return "cryptomon-data"
}

// signalContext is cancelled on Ctrl+C / SIGTERM and, when watchStdin is
// set, when stdin closes — the launcher holds the write end of that pipe,
// so children exit even if the launcher is killed (works on Windows too).
func signalContext(watchStdin bool) (context.Context, context.CancelFunc) {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	if watchStdin {
		go func() {
			_, _ = io.Copy(io.Discard, os.Stdin)
			cancel()
		}()
	}
	return ctx, cancel
}

// RunService runs one currency microservice.
func RunService(args []string) error {
	fs := flag.NewFlagSet("service", flag.ExitOnError)
	asset := fs.String("asset", "", "asset code: "+codes())
	addr := fs.String("addr", "", "listen address (default 127.0.0.1:<asset port>)")
	data := fs.String("data", DefaultDataDir(), "data directory")
	usdt := fs.String("usdt-url", envOr("CRYPTOMON_USDT_URL", "http://127.0.0.1:9103"), "USDT service URL used to convert USDT prices to USD")
	interval := fs.Duration("interval", 3*time.Second, "exchange polling interval for crypto assets")
	noBackfill := fs.Bool("no-backfill", false, "do not download historical prices")
	watch := fs.Bool("watch-stdin", false, "exit when stdin closes (used by the launcher)")
	_ = fs.Parse(args)

	if *asset == "" {
		*asset = os.Getenv("CRYPTOMON_ASSET")
	}
	a, err := catalog.Find(*asset)
	if err != nil {
		return err
	}
	if *addr == "" {
		*addr = envOr("CRYPTOMON_ADDR", fmt.Sprintf("127.0.0.1:%d", a.Port))
	}
	logger := log.New(os.Stderr, fmt.Sprintf("[%-4s] ", a.Code), log.LstdFlags)
	svc, err := service.New(service.Config{
		Asset: a, Addr: *addr, DataDir: *data, USDTURL: *usdt,
		Interval: *interval, NoBackfill: *noBackfill, Version: Version,
	}, logger)
	if err != nil {
		return err
	}
	ctx, cancel := signalContext(*watch)
	defer cancel()
	return svc.Run(ctx)
}

// RunGateway runs the API gateway that serves the UI.
func RunGateway(args []string) error {
	fs := flag.NewFlagSet("gateway", flag.ExitOnError)
	addr := fs.String("addr", envOr("CRYPTOMON_ADDR", "127.0.0.1:8777"), "listen address")
	data := fs.String("data", DefaultDataDir(), "data directory")
	services := fs.String("services", os.Getenv("CRYPTOMON_SERVICES"), "comma-separated CODE=URL list (default: every asset on 127.0.0.1:<port>)")
	watch := fs.Bool("watch-stdin", false, "exit when stdin closes (used by the launcher)")
	_ = fs.Parse(args)

	svcs, err := ParseServices(*services)
	if err != nil {
		return err
	}
	logger := log.New(os.Stderr, "[gate] ", log.LstdFlags)
	gw, err := gateway.New(gateway.Config{Addr: *addr, DataDir: *data, Services: svcs, Version: Version}, logger)
	if err != nil {
		return err
	}
	ctx, cancel := signalContext(*watch)
	defer cancel()
	return gw.Run(ctx)
}

// ParseServices parses "BTC=http://host:9101,ETH=..." into an ordered list;
// an empty string means every catalog asset on localhost.
func ParseServices(s string) ([]gateway.ServiceRef, error) {
	var out []gateway.ServiceRef
	if strings.TrimSpace(s) == "" {
		for _, a := range catalog.Assets {
			out = append(out, gateway.ServiceRef{Code: a.Code, URL: fmt.Sprintf("http://127.0.0.1:%d", a.Port)})
		}
		return out, nil
	}
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		code, url, ok := strings.Cut(part, "=")
		if !ok {
			return nil, fmt.Errorf("bad service %q, want CODE=URL", part)
		}
		a, err := catalog.Find(code)
		if err != nil {
			return nil, err
		}
		out = append(out, gateway.ServiceRef{Code: a.Code, URL: strings.TrimRight(url, "/")})
	}
	return out, nil
}

func codes() string {
	var c []string
	for _, a := range catalog.Assets {
		c = append(c, a.Code)
	}
	return strings.Join(c, ", ")
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
