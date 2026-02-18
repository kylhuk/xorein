package gates

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// We intentionally keep G0..G7 in a fixed slice to enforce deterministic output order.
const DefaultStatusDir = "artifacts/v11/gates"

const staleThreshold = 24 * time.Hour

var AllGateIDs = []GateID{
	"G0", "G1", "G2", "G3", "G4", "G5", "G6", "G7",
}

type GateID string

type GateState string

const (
	StateOpen     GateState = "open"
	StateBlocked  GateState = "blocked"
	StatePromoted GateState = "promoted"
)

// GateSummary reflects the runtime state emitted by the status artifact.
type GateSummary struct {
	ID        GateID
	State     GateState
	UpdatedAt time.Time
	Missing   bool
	Stale     bool
	Path      string
}

// GateRunResult summarizes the gate artifact evaluation.
type GateRunResult struct {
	Summaries   []GateSummary
	EvaluatedAt time.Time
	Passed      bool
	Missing     []GateID
	Stale       []GateID
}

func FreshnessThreshold() time.Duration {
	return staleThreshold
}

func RunGateChecks(statusDir string, now time.Time) (GateRunResult, error) {
	if strings.TrimSpace(statusDir) == "" {
		return GateRunResult{}, errors.New("status directory is required")
	}

	summaries := make([]GateSummary, 0, len(AllGateIDs))
	missing := make([]GateID, 0)
	stale := make([]GateID, 0)
	allPromoted := true

	for _, id := range AllGateIDs {
		path := StatusFilePath(statusDir, id)
		status, err := ParseGateStatusFile(path, id)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				missing = append(missing, id)
				summaries = append(summaries, GateSummary{ID: id, State: StateOpen, Missing: true, Path: path})
				allPromoted = false
				continue
			}
			return GateRunResult{}, err
		}

		isStale := now.Sub(status.UpdatedAt) > staleThreshold
		if isStale {
			stale = append(stale, id)
			allPromoted = false
		}
		if status.State != StatePromoted {
			allPromoted = false
		}

		summaries = append(summaries, GateSummary{
			ID:        id,
			State:     status.State,
			UpdatedAt: status.UpdatedAt,
			Stale:     isStale,
			Path:      path,
		})
	}

	return GateRunResult{
		Summaries:   summaries,
		EvaluatedAt: now.UTC(),
		Passed:      allPromoted && len(missing) == 0 && len(stale) == 0,
		Missing:     missing,
		Stale:       stale,
	}, nil
}

func StatusFilePath(statusDir string, id GateID) string {
	return filepath.Join(statusDir, fmt.Sprintf("%s.status.json", id))
}

func ParseGateStatusFile(path string, expected GateID) (gateStatus, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return gateStatus{}, fmt.Errorf("read %s: %w", path, err)
	}

	var raw gateStatusFile
	if err := json.Unmarshal(data, &raw); err != nil {
		return gateStatus{}, fmt.Errorf("parse %s: %w", path, err)
	}

	if strings.TrimSpace(raw.UpdatedAt) == "" {
		return gateStatus{}, fmt.Errorf("parse %s: updatedAt is required", path)
	}
	updatedAt, err := time.Parse(time.RFC3339, raw.UpdatedAt)
	if err != nil {
		return gateStatus{}, fmt.Errorf("parse %s: invalid updatedAt: %w", path, err)
	}

	gateID := GateID(strings.ToUpper(strings.TrimSpace(raw.GateID)))
	if gateID == "" {
		gateID = expected
	}
	if gateID != expected {
		return gateStatus{}, fmt.Errorf("parse %s: gateId %q does not match expected %q", path, gateID, expected)
	}

	return gateStatus{
		State:     parseGateState(raw.State),
		UpdatedAt: updatedAt,
	}, nil
}

type gateStatusFile struct {
	GateID     string `json:"gateId"`
	State      string `json:"state"`
	UpdatedAt  string `json:"updatedAt"`
	EvidenceID string `json:"evidenceId,omitempty"`
	Owner      string `json:"owner,omitempty"`
	Approver   string `json:"approver,omitempty"`
	Notes      string `json:"notes,omitempty"`
}

type gateStatus struct {
	State     GateState
	UpdatedAt time.Time
}

func parseGateState(raw string) GateState {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case string(StatePromoted):
		return StatePromoted
	case string(StateBlocked):
		return StateBlocked
	case string(StateOpen):
		fallthrough
	default:
		return StateOpen
	}
}
