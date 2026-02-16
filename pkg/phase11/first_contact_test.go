package phase11

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNormalizeOptionsDefaults(t *testing.T) {
	t.Parallel()

	got := normalizeOptions(Options{})

	if got.Runs != 3 {
		t.Fatalf("expected default runs=3, got %d", got.Runs)
	}
	if got.OutputDir != "artifacts/generated/first-contact" {
		t.Fatalf("expected default output dir, got %q", got.OutputDir)
	}
	if got.ServerIDPrefix != "first-contact-server" {
		t.Fatalf("expected default server prefix, got %q", got.ServerIDPrefix)
	}
	if got.IdentityPrefix != "first-contact-identity" {
		t.Fatalf("expected default identity prefix, got %q", got.IdentityPrefix)
	}
	if got.TargetDuration != 5*time.Minute {
		t.Fatalf("expected default target duration=5m, got %s", got.TargetDuration)
	}
	if got.RegressionOut != "artifacts/generated/regression" {
		t.Fatalf("expected default regression output dir, got %q", got.RegressionOut)
	}
}

func TestSummarizeAggregatesCountsAndDurations(t *testing.T) {
	t.Parallel()

	runs := []RunResult{
		{
			RunID:          1,
			Success:        true,
			TargetMet:      true,
			DurationMillis: 100,
			EventCounts:    map[string]int{"chat_roundtrip": 1},
			ReasonCodeCounts: map[string]int{
				"chat_ok": 1,
			},
			FallbackCounts: map[string]int{"voice_capacity_fallback": 0},
		},
		{
			RunID:          2,
			Success:        false,
			TargetMet:      false,
			DurationMillis: 300,
			FailureReason:  "voice_connect: failed",
			FailureOwner:   "Realtime Engineer",
			EventCounts:    map[string]int{"failure": 1},
			ReasonCodeCounts: map[string]int{
				"voice_failed": 1,
			},
			FallbackCounts: map[string]int{"voice_capacity_fallback": 2},
		},
	}

	summary := summarize(Options{Runs: 2, TargetDuration: 5 * time.Minute}, runs)

	if summary.RunsRequested != 2 || summary.RunsCompleted != 2 {
		t.Fatalf("unexpected run counts: requested=%d completed=%d", summary.RunsRequested, summary.RunsCompleted)
	}
	if summary.PassedRuns != 1 || summary.FailedRuns != 1 {
		t.Fatalf("unexpected pass/fail counts: passed=%d failed=%d", summary.PassedRuns, summary.FailedRuns)
	}
	if summary.MeanDurationMS != 200 {
		t.Fatalf("expected mean duration 200ms, got %d", summary.MeanDurationMS)
	}
	if summary.MedianDurationMS != 200 {
		t.Fatalf("expected median duration 200ms, got %d", summary.MedianDurationMS)
	}
	if summary.EventCounts["chat_roundtrip"] != 1 || summary.EventCounts["failure"] != 1 {
		t.Fatalf("unexpected event counts: %#v", summary.EventCounts)
	}
	if summary.ReasonCodeCounts["chat_ok"] != 1 || summary.ReasonCodeCounts["voice_failed"] != 1 {
		t.Fatalf("unexpected reason-code counts: %#v", summary.ReasonCodeCounts)
	}
	if summary.FallbackCounts["voice_capacity_fallback"] != 2 {
		t.Fatalf("expected fallback count=2, got %d", summary.FallbackCounts["voice_capacity_fallback"])
	}
	if len(summary.Failures) != 1 {
		t.Fatalf("expected exactly one failure trace, got %d", len(summary.Failures))
	}
}

func TestRunFirstContactWritesArtifacts(t *testing.T) {
	t.Parallel()

	outDir := t.TempDir()
	regressionOut := filepath.Join(outDir, "regression")
	gotSummary, gotRuns, err := RunFirstContact(context.Background(), Options{
		Runs:           2,
		OutputDir:      outDir,
		ServerIDPrefix: "p11-server",
		IdentityPrefix: "p11-identity",
		TargetDuration: 5 * time.Minute,
		RegressionOut:  regressionOut,
	})
	if err != nil {
		t.Fatalf("RunFirstContact returned error: %v", err)
	}
	if gotSummary == nil {
		t.Fatal("expected non-nil summary")
	}
	if len(gotRuns) != 2 {
		t.Fatalf("expected 2 runs, got %d", len(gotRuns))
	}

	wantFiles := []string{
		"run-01.json",
		"run-02.json",
		"summary.json",
		"summary.md",
	}
	for _, name := range wantFiles {
		path := filepath.Join(outDir, name)
		info, statErr := os.Stat(path)
		if statErr != nil {
			t.Fatalf("expected artifact %s: %v", name, statErr)
		}
		if info.Size() == 0 {
			t.Fatalf("artifact %s is empty", name)
		}
	}

	regressionFiles := []string{
		"report.txt",
		"defects.json",
	}
	for _, name := range regressionFiles {
		path := filepath.Join(regressionOut, name)
		info, statErr := os.Stat(path)
		if statErr != nil {
			t.Fatalf("expected regression artifact %s: %v", name, statErr)
		}
		if info.Size() == 0 {
			t.Fatalf("regression artifact %s is empty", name)
		}
	}
}

func TestWriteDefectStatusCreatesCleanSentinelWhenNoFailures(t *testing.T) {
	t.Parallel()

	out := filepath.Join(t.TempDir(), "defects.json")
	err := writeDefectStatus(out, []RunResult{{RunID: 1, Success: true}})
	if err != nil {
		t.Fatalf("writeDefectStatus returned error: %v", err)
	}

	content, readErr := os.ReadFile(out)
	if readErr != nil {
		t.Fatalf("read defects artifact: %v", readErr)
	}
	text := string(content)
	if !containsAll(text, []string{"P11-T2-REGRESSION-CLEAN", "\"status\": \"closed\"", "\"reason_code\": \"none\""}) {
		t.Fatalf("clean defects artifact missing sentinel fields: %s", text)
	}
}

func TestWriteDefectStatusCreatesFailureEntries(t *testing.T) {
	t.Parallel()

	out := filepath.Join(t.TempDir(), "defects.json")
	runs := []RunResult{
		{
			RunID:         2,
			Success:       false,
			FailureReason: "voice_connect: quota exceeded",
			FailureOwner:  "Realtime Engineer",
			ReasonCodeCounts: map[string]int{
				"voice_failed":      2,
				"relay_unavailable": 1,
			},
		},
	}
	err := writeDefectStatus(out, runs)
	if err != nil {
		t.Fatalf("writeDefectStatus returned error: %v", err)
	}

	content, readErr := os.ReadFile(out)
	if readErr != nil {
		t.Fatalf("read defects artifact: %v", readErr)
	}
	text := string(content)
	if !containsAll(text, []string{"P11-T2-RUN-02", "\"reason_code\": \"voice_failed\"", "\"owner\": \"Realtime Engineer\"", "\"status\": \"open\""}) {
		t.Fatalf("failure defects artifact missing expected fields: %s", text)
	}
}

func containsAll(s string, needles []string) bool {
	for _, needle := range needles {
		if !contains(s, needle) {
			return false
		}
	}
	return true
}

func contains(s, needle string) bool {
	return len(needle) == 0 || (len(s) >= len(needle) && indexOf(s, needle) >= 0)
}

func indexOf(s, needle string) int {
	for i := 0; i+len(needle) <= len(s); i++ {
		if s[i:i+len(needle)] == needle {
			return i
		}
	}
	return -1
}
