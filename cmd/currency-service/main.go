// Command currency-service runs one currency microservice, e.g.
// `currency-service --asset BTC --addr 0.0.0.0:9101`. Used for container
// or multi-host deployments; the cryptomon launcher embeds the same code.
package main

import (
	"fmt"
	"os"

	"cryptomon/internal/app"
)

func main() {
	if err := app.RunService(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
