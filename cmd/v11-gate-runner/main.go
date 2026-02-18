package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/aether/code_aether/pkg/v11/gates"
)

func main() {
	statusDir := flag.String("status-dir", gates.DefaultStatusDir, "directory containing v11 gate status artifacts")
	flag.Parse()

	result, err := gates.RunGateChecks(*statusDir, time.Now().UTC())
	if err != nil {
		fmt.Fprintf(os.Stderr, "v11 gate runner: failed to evaluate status artifacts in %q: %v\n", *statusDir, err)
		os.Exit(1)
	}

	fmt.Println("v11 gate runner summary")
	fmt.Printf(" status directory: %s\n", *statusDir)
	fmt.Printf(" evaluated at: %s\n", result.EvaluatedAt.Format(time.RFC3339))
	fmt.Printf(" freshness threshold: %s\n", gates.FreshnessThreshold())
	for _, summary := range result.Summaries {
		updated := "n/a"
		if !summary.UpdatedAt.IsZero() {
			updated = summary.UpdatedAt.Format(time.RFC3339)
		}
		if summary.Missing {
			fmt.Printf(" %s blocked (missing %s)\n", summary.ID, summary.Path)
			continue
		}
		if summary.Stale {
			fmt.Printf(" %s %s (stale, updated %s)\n", summary.ID, summary.State, updated)
			continue
		}
		fmt.Printf(" %s %s (updated %s)\n", summary.ID, summary.State, updated)
	}

	if result.Passed {
		fmt.Println("PASS: All gates promoted and artifact is fresh.")
		os.Exit(0)
	}

	fmt.Println("FAIL: gate artifact indicates at least one failed condition.")
	if len(result.Stale) > 0 {
		fmt.Printf(" stale gates: %v\n", result.Stale)
	}
	if len(result.Missing) > 0 {
		fmt.Printf(" missing gate artifacts: %v\n", result.Missing)
	}
	os.Exit(1)
}
