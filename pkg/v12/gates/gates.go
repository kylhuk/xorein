package gates

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	DefaultStatusDir = "artifacts/v12/gates"
	staleThreshold   = 24 * time.Hour
)

// GateID identifies v12 promotion gates.
type GateID string

const (
	GateG0 GateID = "G0"
	GateG1 GateID = "G1"
	GateG2 GateID = "G2"
	GateG3 GateID = "G3"
	GateG4 GateID = "G4"
	GateG5 GateID = "G5"
	GateG6 GateID = "G6"
	GateG7 GateID = "G7"
	GateG8 GateID = "G8"
	GateG9 GateID = "G9"
)

// AllGateIDs defines deterministic gate ordering for output and checks.
var AllGateIDs = []GateID{
	GateG0,
	GateG1,
	GateG2,
	GateG3,
	GateG4,
	GateG5,
	GateG6,
	GateG7,
	GateG8,
	GateG9,
}

// GateState captures machine-readable lifecycle state.
type GateState string

const (
	StateOpen     GateState = "open"
	StateBlocked  GateState = "blocked"
	StatePromoted GateState = "promoted"
)

// GateSummary captures status of one gate artifact.
type GateSummary struct {
	ID        GateID
	State     GateState
	UpdatedAt time.Time
	Missing   bool
	Stale     bool
	Path      string
}

// GateRunResult captures aggregate execution details.
type GateRunResult struct {
	Summaries   []GateSummary
	EvaluatedAt time.Time
	Passed      bool
	Missing     []GateID
	Stale       []GateID
}

type gateStatusFile struct {
	GateID    string `json:"gateId"`
	State     string `json:"state"`
	UpdatedAt string `json:"updatedAt"`
}

// FreshnessThreshold exposes stale threshold for command output.
func FreshnessThreshold() time.Duration {
	return staleThreshold
}

// RunGateChecks evaluates all v12 gate status artifacts.
func RunGateChecks(statusDir string, now time.Time) (GateRunResult, error) {
	statusDir = strings.TrimSpace(statusDir)
	if statusDir == "" {
		return GateRunResult{}, errors.New("status directory is required")
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}

	result := GateRunResult{
		Summaries:   make([]GateSummary, 0, len(AllGateIDs)),
		EvaluatedAt: now,
		Passed:      true,
		Missing:     []GateID{},
		Stale:       []GateID{},
	}

	for _, gateID := range AllGateIDs {
		statusPath := StatusFilePath(statusDir, gateID)
		summary := GateSummary{ID: gateID, Path: statusPath}

		status, err := ParseGateStatusFile(statusPath, gateID)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				summary.Missing = true
				summary.State = StateOpen
				result.Missing = append(result.Missing, gateID)
				result.Passed = false
				result.Summaries = append(result.Summaries, summary)
				continue
			}
			return GateRunResult{}, fmt.Errorf("parse %s: %w", statusPath, err)
		}

		summary.State = status.State
		summary.UpdatedAt = status.UpdatedAt

		if now.Sub(status.UpdatedAt) > staleThreshold {
			summary.Stale = true
			result.Stale = append(result.Stale, gateID)
			result.Passed = false
		}
		if status.State != StatePromoted {
			result.Passed = false
		}

		result.Summaries = append(result.Summaries, summary)
	}

	return result, nil
}

// StatusFilePath returns canonical status path for one gate.
func StatusFilePath(statusDir string, gateID GateID) string {
	return filepath.Join(statusDir, fmt.Sprintf("%s.status.json", gateID))
}

// ParseGateStatusFile parses one gate status file and validates schema.
func ParseGateStatusFile(path string, expected GateID) (GateSummary, error) {
	base := filepath.Base(path)
	expectedBase := fmt.Sprintf("%s.status.json", expected)
	if base != expectedBase {
		return GateSummary{}, fmt.Errorf("unexpected status filename %q (expected %q)", base, expectedBase)
	}

	// #nosec G304 -- path is constrained to expected gate status filename.
	data, err := os.ReadFile(path)
	if err != nil {
		return GateSummary{}, err
	}

	var raw gateStatusFile
	if err := json.Unmarshal(data, &raw); err != nil {
		return GateSummary{}, fmt.Errorf("decode json: %w", err)
	}

	updatedAt := strings.TrimSpace(raw.UpdatedAt)
	if updatedAt == "" {
		return GateSummary{}, errors.New("updatedAt is required")
	}
	parsedUpdatedAt, err := time.Parse(time.RFC3339, updatedAt)
	if err != nil {
		return GateSummary{}, fmt.Errorf("parse updatedAt: %w", err)
	}

	gateID := strings.TrimSpace(raw.GateID)
	if gateID == "" {
		gateID = string(expected)
	}
	if GateID(gateID) != expected {
		return GateSummary{}, fmt.Errorf("gateId %q does not match expected %q", gateID, expected)
	}

	return GateSummary{
		ID:        expected,
		State:     parseGateState(raw.State),
		UpdatedAt: parsedUpdatedAt,
	}, nil
}

func parseGateState(raw string) GateState {
	switch GateState(strings.TrimSpace(raw)) {
	case StateOpen:
		return StateOpen
	case StateBlocked:
		return StateBlocked
	case StatePromoted:
		return StatePromoted
	default:
		return StateOpen
	}
}

// ValidGateIDs returns a stable sorted list for diagnostics.
func ValidGateIDs() []string {
	ids := make([]string, 0, len(AllGateIDs))
	for _, gateID := range AllGateIDs {
		ids = append(ids, string(gateID))
	}
	sort.Strings(ids)
	return ids
}
