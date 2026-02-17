package archivist

import (
	"fmt"
	"sort"
	"time"
)

type CapabilityAnnouncement struct {
	PeerID  string
	Score   int
	Epoch   int
	State   EnrollmentState
	Expires time.Time
}

type SelectionResult struct {
	PeerID string
	Score  int
	Rank   int
}

const withdrawalGrace = 48 * time.Hour

func Advertise(peer string, score int) CapabilityAnnouncement {
	return CapabilityAnnouncement{
		PeerID:  peer,
		Score:   score,
		Epoch:   time.Now().UTC().Day(),
		State:   StateActive,
		Expires: time.Now().UTC().Add(withdrawalGrace),
	}
}

func SelectArchivists(candidates []CapabilityAnnouncement, limit int) []SelectionResult {
	if limit <= 0 {
		limit = 3
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].Score == candidates[j].Score {
			return candidates[i].PeerID < candidates[j].PeerID
		}
		return candidates[i].Score > candidates[j].Score
	})

	result := make([]SelectionResult, 0, limit)
	for idx, candidate := range candidates {
		if idx >= limit {
			break
		}
		result = append(result, SelectionResult{PeerID: candidate.PeerID, Score: candidate.Score, Rank: idx + 1})
	}
	return result
}

type CoverageSignal struct {
	PeerID   string
	Coverage float64
	Message  string
}

func CoverageDrop(peerID string, coverage float64) CoverageSignal {
	msg := fmt.Sprintf("archivist.coverage=%.2f", coverage)
	if coverage < 0.5 {
		msg = fmt.Sprintf("archivist.coverage.low=%.2f", coverage)
	}
	return CoverageSignal{PeerID: peerID, Coverage: coverage, Message: msg}
}

func GracefulWithdrawal(peer string) CapabilityAnnouncement {
	ann := Advertise(peer, 0)
	ann.State = StateWithdrawn
	ann.Score = 0
	ann.Expires = time.Now().UTC().Add(withdrawalGrace)
	return ann
}
