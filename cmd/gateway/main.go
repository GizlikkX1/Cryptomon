// Command gateway runs the API gateway and web UI on its own, e.g.
// `gateway --addr 0.0.0.0:8777 --services BTC=http://btc:9101,...`.
package main

import (
	"errors"
	"fmt"
	"os"

	"cryptomon/internal/app"
	"cryptomon/internal/gateway"
)

func main() {
	err := app.RunGateway(os.Args[1:])
	if errors.Is(err, gateway.ErrShutdown) {
		return
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
