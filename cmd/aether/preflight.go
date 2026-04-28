package main

import (
	"fmt"
	"io"
)

var (
	runRepoSnapshotFn   = runRepoSnapshot
	runBaselineHealthFn = runBaselineHealth
)

func runPreflight(w io.Writer, root string) error {
	_, _ = fmt.Fprintln(w, "Preflight report")
	_, _ = fmt.Fprintln(w)
	if err := runRepoSnapshotFn(w, root); err != nil {
		return err
	}
	_, _ = fmt.Fprintln(w)
	return runBaselineHealthFn(w, root)
}
