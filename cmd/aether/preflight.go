package main

import (
	"fmt"
	"io"
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
