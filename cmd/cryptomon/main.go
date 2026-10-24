// Command cryptomon is the all-in-one launcher. Without arguments it
// starts every currency microservice and the gateway as separate
// processes and opens the UI. The same binary also runs a single
// component: `cryptomon service --asset BTC` or `cryptomon gateway`.
package main

import (
	"errors"
	"fmt"
	"os"

	"cryptomon/internal/app"
	"cryptomon/internal/gateway"
	"cryptomon/internal/launcher"
)

func main() {
	args := os.Args[1:]
	cmd := "run"
	if len(args) > 0 && len(args[0]) > 0 && args[0][0] != '-' {
		cmd, args = args[0], args[1:]
	}
	var err error
	switch cmd {
	case "run":
		err = app.RunAll(args)
	case "service":
		err = app.RunService(args)
	case "gateway":
		err = app.RunGateway(args)
	case "version":
		fmt.Println("cryptomon", app.Version)
	case "help", "-h", "--help":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n", cmd)
		usage()
		os.Exit(2)
	}
	if errors.Is(err, gateway.ErrShutdown) {
		os.Exit(launcher.ExitShutdown)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprint(os.Stderr, `Cryptomon — real-time crypto and fiat exchange-rate monitor

Usage:
  cryptomon [run] [flags]          start all microservices + gateway and open the UI
  cryptomon service --asset CODE   run one currency microservice
  cryptomon gateway [flags]        run the API gateway / web UI only
  cryptomon version

Run "cryptomon <command> -h" for the flags of a command.
`)
}
