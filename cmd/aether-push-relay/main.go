package main

import (
	"flag"
	"fmt"
	"os"
)

var (
	modeFlag          = flag.String("mode", "relay", "runtime mode: relay|probe")
	pushRelayDispatch = dispatchPushRelay
)

type relayHandlers struct {
	runRelay func()
	runProbe func()
}

func main() {
	flag.Parse()
	exitCode := pushRelayDispatch(*modeFlag, relayHandlers{runRelay: runRelayMode, runProbe: runProbeMode})
	if exitCode != 0 {
		os.Exit(exitCode)
	}
}

func dispatchPushRelay(mode string, handlers relayHandlers) int {
	switch mode {
	case "relay":
		handlers.runRelay()
		return 0
	case "probe":
		handlers.runProbe()
		return 0
	default:
		fmt.Fprintf(os.Stderr, "invalid mode %q; expected relay|probe\n", mode)
		return 2
	}
}

func runRelayMode() {
	fmt.Println("push relay runtime ready")
}

func runProbeMode() {
	fmt.Println("push relay probe ok")
}
