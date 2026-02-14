package main

import (
	"flag"
	"fmt"
	"os"
)

var (
	runMode = flag.String("mode", "client", "runtime mode: client|relay|bootstrap")
)

func main() {
	flag.Parse()
	switch *runMode {
	case "client", "relay", "bootstrap":
		// Valid mode for single-binary scaffolding.
	default:
		fmt.Fprintf(os.Stderr, "invalid --mode %q; expected client|relay|bootstrap\n", *runMode)
		os.Exit(2)
	}

	fmt.Printf("Phase 2 foundation stub: cmd/aether mode=%s\n", *runMode)
	// TODO: wire the protocol/network/crypto/storage/ui seams from pkg/app once interfaces stabilize.
	// TODO: replace fmt.Printf with structured logging after logging strategy is approved.
}
