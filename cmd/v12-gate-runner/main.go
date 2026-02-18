package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/aether/code_aether/pkg/v12/gates"
)

func main() {
	statusDir := flag.String("status-dir", gates.DefaultStatusDir, "directory containing v12 gate status artifacts")
	flag.Parse()

	now := time.Now().UTC()
	result, err := gates.RunGateChecks(*statusDir, now)
	if err != nil {
		fmt.Fprintf(os.Stderr, "v12 gate runner failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("v12 gate runner summary")
	fmt.Printf("status directory: %s\n", *statusDir)
	fmt.Printf("evaluated at: %s\n", result.EvaluatedAt.Format(time.RFC3339))
	fmt.Printf("freshness threshold: %s\n", gates.FreshnessThreshold())

	for _, summary := range result.Summaries {
		switch {
		case summary.Missing:
			fmt.Printf("%s blocked (missing %s)\n", summary.ID, summary.Path)
		case summary.Stale:
			fmt.Printf("%s %s (stale, updated %s)\n", summary.ID, summary.State, summary.UpdatedAt.Format(time.RFC3339))
		default:
			fmt.Printf("%s %s (updated %s)\n", summary.ID, summary.State, summary.UpdatedAt.Format(time.RFC3339))
		}
	}

	if result.Passed {
		fmt.Println("PASS: All gates promoted and artifact is fresh.")
		return
	}

	fmt.Println("FAIL: gate artifact indicates at least one failed condition.")
	if len(result.Stale) > 0 {
		fmt.Printf("stale gates: %v\n", result.Stale)
	}
	if len(result.Missing) > 0 {
		fmt.Printf("missing gates: %v\n", result.Missing)
	}
	os.Exit(1)
}
