package gates

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRunGateChecks_MissingGateArtifact(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	now := time.Date(2026, 2, 17, 12, 0, 0, 0, time.UTC)
	missingGate := AllGateIDs[len(AllGateIDs)-1]

	for _, id := range AllGateIDs {
		if id == missingGate {
			continue
		}
		writeGateStatus(t, dir, id, StatePromoted, now)
	}

	result, err := RunGateChecks(dir, now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Passed {
		t.Fatal("expected failed run when a gate artifact is missing")
	}
	if len(result.Missing) != 1 || result.Missing[0] != missingGate {
		t.Fatalf("expected missing gate %s, got %v", missingGate, result.Missing)
	}
	if !result.Summaries[len(result.Summaries)-1].Missing {
		t.Fatal("expected final summary entry to report missing artifact")
	}
}

func TestRunGateChecks_StaleStatus(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	now := time.Date(2026, 2, 17, 12, 0, 0, 0, time.UTC)
	updatedAt := now.Add(-48 * time.Hour)

	for _, id := range AllGateIDs {
		writeGateStatus(t, dir, id, StatePromoted, updatedAt)
	}

	result, err := RunGateChecks(dir, now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Stale) != len(AllGateIDs) {
		t.Fatalf("expected all gates to be stale, got %v", result.Stale)
	}
	if result.Passed {
		t.Fatalf("expected failure because status file was stale")
	}
}

func TestRunGateChecks_AllPromotedFresh(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	now := time.Date(2026, 2, 17, 12, 0, 0, 0, time.UTC)
	updatedAt := now.Add(-1 * time.Hour)

	for _, id := range AllGateIDs {
		writeGateStatus(t, dir, id, StatePromoted, updatedAt)
	}

	result, err := RunGateChecks(dir, now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Passed {
		t.Fatalf("expected pass when all gates promoted and fresh")
	}
	if len(result.Stale) != 0 {
		t.Fatalf("expected no stale gates, got %v", result.Stale)
	}
	if len(result.Missing) != 0 {
		t.Fatalf("expected no missing gates, got %v", result.Missing)
	}
	for i, id := range AllGateIDs {
		if result.Summaries[i].ID != id {
			t.Fatalf("expected summary order %s, got %s", id, result.Summaries[i].ID)
		}
	}
}

func TestRunGateChecks_MismatchedGateIDFails(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	now := time.Date(2026, 2, 17, 12, 0, 0, 0, time.UTC)
	updatedAt := now.Add(-1 * time.Hour)

	for _, id := range AllGateIDs {
		writeGateStatus(t, dir, id, StatePromoted, updatedAt)
	}

	path := StatusFilePath(dir, "G3")
	writeGateStatusPayload(t, path, map[string]string{
		"gateId":    "G7",
		"state":     string(StatePromoted),
		"updatedAt": updatedAt.Format(time.RFC3339),
	})

	_, err := RunGateChecks(dir, now)
	if err == nil {
		t.Fatal("expected mismatch error")
	}
	if !strings.Contains(err.Error(), "does not match expected") {
		t.Fatalf("expected mismatch error, got %v", err)
	}
}

func TestParseGateStatusFileRejectsUnexpectedFilename(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	updatedAt := time.Date(2026, 2, 18, 8, 0, 0, 0, time.UTC)
	path := filepath.Join(dir, "custom.status.json")
	writeGateStatusPayload(t, path, map[string]string{
		"gateId":    "G3",
		"state":     string(StatePromoted),
		"updatedAt": updatedAt.Format(time.RFC3339),
	})

	_, err := ParseGateStatusFile(path, "G3")
	if err == nil {
		t.Fatal("expected unexpected filename error")
	}
	if !strings.Contains(err.Error(), "unexpected status file name") {
		t.Fatalf("expected unexpected filename error, got %v", err)
	}
}

func writeGateStatus(t *testing.T, dir string, id GateID, state GateState, updatedAt time.Time) {
	t.Helper()
	writeGateStatusPayload(t, StatusFilePath(dir, id), map[string]string{
		"gateId":    string(id),
		"state":     string(state),
		"updatedAt": updatedAt.Format(time.RFC3339),
	})
}

func writeGateStatusPayload(t *testing.T, path string, payload map[string]string) {
	t.Helper()
	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal fixture: %v", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("failed to write status file: %v", err)
	}
}
