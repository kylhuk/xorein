package gates

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
)

func TestRunGateChecksMissingGateArtifact(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	now := time.Unix(1_730_002_000, 0).UTC()
	for _, gateID := range AllGateIDs {
		if gateID == GateG4 {
			continue
		}
		writeGateStatus(t, dir, gateID, StatePromoted, now)
	}

	result, err := RunGateChecks(dir, now)
	if err != nil {
		t.Fatalf("RunGateChecks returned error: %v", err)
	}
	if result.Passed {
		t.Fatal("expected run to fail when one gate file is missing")
	}
	if len(result.Missing) != 1 || result.Missing[0] != GateG4 {
		t.Fatalf("missing gates mismatch: %+v", result.Missing)
	}

	var missingSummary GateSummary
	for _, summary := range result.Summaries {
		if summary.ID == GateG4 {
			missingSummary = summary
			break
		}
	}
	if !missingSummary.Missing {
		t.Fatal("expected missing summary flag for G4")
	}
}

func TestRunGateChecksStaleStatus(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	now := time.Unix(1_730_002_100, 0).UTC()
	staleAt := now.Add(-48 * time.Hour)
	for _, gateID := range AllGateIDs {
		writeGateStatus(t, dir, gateID, StatePromoted, staleAt)
	}

	result, err := RunGateChecks(dir, now)
	if err != nil {
		t.Fatalf("RunGateChecks returned error: %v", err)
	}
	if result.Passed {
		t.Fatal("expected stale artifacts to fail run")
	}
	if len(result.Stale) != len(AllGateIDs) {
		t.Fatalf("expected all gates stale, got %d", len(result.Stale))
	}
}

func TestRunGateChecksAllPromotedFresh(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	now := time.Unix(1_730_002_200, 0).UTC()
	freshAt := now.Add(-1 * time.Hour)
	for _, gateID := range AllGateIDs {
		writeGateStatus(t, dir, gateID, StatePromoted, freshAt)
	}

	result, err := RunGateChecks(dir, now)
	if err != nil {
		t.Fatalf("RunGateChecks returned error: %v", err)
	}
	if !result.Passed {
		t.Fatal("expected all promoted and fresh gates to pass")
	}
	if len(result.Missing) != 0 || len(result.Stale) != 0 {
		t.Fatalf("expected no missing/stale gates, got missing=%v stale=%v", result.Missing, result.Stale)
	}
	if len(result.Summaries) != len(AllGateIDs) {
		t.Fatalf("summary length mismatch: got %d want %d", len(result.Summaries), len(AllGateIDs))
	}

	ids := make([]GateID, 0, len(result.Summaries))
	for _, summary := range result.Summaries {
		ids = append(ids, summary.ID)
	}
	if !slices.Equal(ids, AllGateIDs) {
		t.Fatalf("summary ordering mismatch: got %v want %v", ids, AllGateIDs)
	}
}

func TestRunGateChecksMismatchedGateIDFails(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	now := time.Unix(1_730_002_300, 0).UTC()
	for _, gateID := range AllGateIDs {
		if gateID == GateG0 {
			writeGateStatusPayload(t, dir, gateID, gateStatusFile{GateID: "G9", State: string(StatePromoted), UpdatedAt: now.Format(time.RFC3339)})
			continue
		}
		writeGateStatus(t, dir, gateID, StatePromoted, now)
	}

	_, err := RunGateChecks(dir, now)
	if err == nil {
		t.Fatal("expected mismatched gate id to fail")
	}
	if !strings.Contains(err.Error(), "does not match expected") {
		t.Fatalf("expected mismatch error, got %v", err)
	}
}

func TestParseGateStatusFileRejectsUnexpectedFilename(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "unexpected.status.json")
	if err := os.WriteFile(path, []byte(`{"gateId":"G0","state":"promoted","updatedAt":"2024-01-01T00:00:00Z"}`), 0o600); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	_, err := ParseGateStatusFile(path, GateG0)
	if err == nil {
		t.Fatal("expected unexpected filename failure")
	}
}

func TestRunGateChecksRequiresStatusDir(t *testing.T) {
	t.Parallel()

	_, err := RunGateChecks("", time.Now().UTC())
	if err == nil {
		t.Fatal("expected status directory validation error")
	}
}

func TestParseGateStatusFileMissingFile(t *testing.T) {
	t.Parallel()

	_, err := ParseGateStatusFile(filepath.Join(t.TempDir(), "G0.status.json"), GateG0)
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected os.ErrNotExist, got %v", err)
	}
}

func writeGateStatus(t *testing.T, dir string, gateID GateID, state GateState, updatedAt time.Time) {
	t.Helper()
	writeGateStatusPayload(t, dir, gateID, gateStatusFile{
		GateID:    string(gateID),
		State:     string(state),
		UpdatedAt: updatedAt.Format(time.RFC3339),
	})
}

func writeGateStatusPayload(t *testing.T, dir string, gateID GateID, payload gateStatusFile) {
	t.Helper()
	path := StatusFilePath(dir, gateID)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Marshal returned error: %v", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}
}
