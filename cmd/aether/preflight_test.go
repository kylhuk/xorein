package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
)

func TestRunPreflightRunsSnapshotThenBaseline(t *testing.T) {
	resetPreflightHooks(t)
	calls := make([]string, 0, 2)
	runRepoSnapshotFn = func(w io.Writer, root string) error {
		calls = append(calls, "snapshot:"+root)
		_, _ = fmt.Fprintln(w, "snapshot ok")
		return nil
	}
	runBaselineHealthFn = func(w io.Writer, root string) error {
		calls = append(calls, "baseline:"+root)
		_, _ = fmt.Fprintln(w, "baseline ok")
		return nil
	}

	var buf bytes.Buffer
	if err := runPreflight(&buf, "/tmp/repo"); err != nil {
		t.Fatalf("runPreflight() error = %v", err)
	}
	if got, want := calls, []string{"snapshot:/tmp/repo", "baseline:/tmp/repo"}; len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("calls = %#v, want %#v", got, want)
	}
	output := buf.String()
	for _, needle := range []string{"Preflight report", "snapshot ok", "baseline ok"} {
		if !strings.Contains(output, needle) {
			t.Fatalf("expected %q in output %q", needle, output)
		}
	}
	if !strings.Contains(output, "snapshot ok\n\nbaseline ok") {
		t.Fatalf("expected blank line between sections in %q", output)
	}
}

func TestRunPreflightStopsAfterSnapshotError(t *testing.T) {
	resetPreflightHooks(t)
	baselineCalled := false
	runRepoSnapshotFn = func(w io.Writer, root string) error {
		return errors.New("snapshot failed")
	}
	runBaselineHealthFn = func(w io.Writer, root string) error {
		baselineCalled = true
		return nil
	}

	var buf bytes.Buffer
	err := runPreflight(&buf, "/tmp/repo")
	if err == nil || err.Error() != "snapshot failed" {
		t.Fatalf("runPreflight() error = %v", err)
	}
	if baselineCalled {
		t.Fatal("baseline health should not run after snapshot failure")
	}
}

func resetPreflightHooks(t *testing.T) {
	t.Helper()
	runRepoSnapshotFn = runRepoSnapshot
	runBaselineHealthFn = runBaselineHealth
	t.Cleanup(func() {
		runRepoSnapshotFn = runRepoSnapshot
		runBaselineHealthFn = runBaselineHealth
	})
}
