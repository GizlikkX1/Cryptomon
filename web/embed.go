// Package web embeds the browser UI into the gateway binary.
package web

import "embed"

//go:embed static
var Static embed.FS
