//go:build !(js && wasm)

// Command recuerdo-web is Recuerdo's web version; it is built for the
// browser: GOOS=js GOARCH=wasm (scripts/build-web.sh).
package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Fprintln(os.Stderr, "recuerdo-web runs in a browser: build it with scripts/build-web.sh (GOOS=js GOARCH=wasm)")
	os.Exit(2)
}
