package app

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"cryptomon/internal/catalog"
	"cryptomon/internal/launcher"
)

// RunAll is the default mode: one microservice process per currency plus
// the gateway, supervised, with the UI opened in an app window.
func RunAll(args []string) error {
	fs := flag.NewFlagSet("run", flag.ExitOnError)
	data := fs.String("data", DefaultDataDir(), "data directory")
	port := fs.Int("port", 8777, "gateway (UI) port")
	ui := fs.String("ui", "app", "how to show the UI: app (separate window), browser, none")
	interval := fs.Duration("interval", 3*time.Second, "exchange polling interval for crypto assets")
	noBackfill := fs.Bool("no-backfill", false, "do not download historical prices")
	only := fs.String("assets", "", "comma-separated subset of assets to run (default: all)")
	_ = fs.Parse(args)

	logger := log.New(os.Stderr, "[main] ", log.LstdFlags)
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(*data, 0o755); err != nil {
		return err
	}

	gwBase := fmt.Sprintf("http://127.0.0.1:%d", *port)
	if launcher.IsCryptomon(gwBase) {
		logger.Printf("Cryptomon is already running at %s, opening it", gwBase)
		return launcher.OpenUI(*ui, gwBase, filepath.Join(*data, "ui-profile"))
	}
	gwPort, err := launcher.PickPort(*port)
	if err != nil {
		return err
	}
	gwBase = fmt.Sprintf("http://127.0.0.1:%d", gwPort)

	assets := catalog.Assets
	if *only != "" {
		assets = nil
		for _, code := range strings.Split(*only, ",") {
			a, err := catalog.Find(code)
			if err != nil {
				return err
			}
			assets = append(assets, a)
		}
	}

	var children []launcher.Child
	var services []string
	urls := map[string]string{}
	for _, a := range assets {
		p, err := launcher.PickPort(a.Port)
		if err != nil {
			return err
		}
		urls[a.Code] = fmt.Sprintf("http://127.0.0.1:%d", p)
	}
	usdtURL := urls["USDT"]
	for _, a := range assets {
		args := []string{"service", "--asset", a.Code, "--addr", strings.TrimPrefix(urls[a.Code], "http://"),
			"--data", *data, "--interval", interval.String(), "--watch-stdin"}
		if usdtURL != "" {
			args = append(args, "--usdt-url", usdtURL)
		}
		if *noBackfill {
			args = append(args, "--no-backfill")
		}
		children = append(children, launcher.Child{Name: a.Code + " service", Args: args})
		services = append(services, a.Code+"="+urls[a.Code])
	}
	children = append(children, launcher.Child{Name: "gateway", Args: []string{
		"gateway", "--addr", fmt.Sprintf("127.0.0.1:%d", gwPort), "--data", *data,
		"--services", strings.Join(services, ","), "--watch-stdin",
	}})

	ctx, cancel := signalContext(false)
	defer cancel()
	sup := launcher.NewSupervisor(exe, logger, children)
	go func() {
		if err := launcher.WaitHTTP(ctx, gwBase+"/api/meta", 30*time.Second); err != nil {
			logger.Printf("gateway is not responding: %v", err)
			return
		}
		logger.Printf("Cryptomon %s is running: %s  (Ctrl+C to stop)", Version, gwBase)
		if err := launcher.OpenUI(*ui, gwBase, filepath.Join(*data, "ui-profile")); err != nil {
			logger.Printf("could not open the UI (%v); open %s in a browser", err, gwBase)
		}
	}()
	logger.Printf("starting %d microservices and the gateway, data in %s", len(assets), *data)
	sup.Run(ctx)
	logger.Printf("stopped")
	return nil
}
