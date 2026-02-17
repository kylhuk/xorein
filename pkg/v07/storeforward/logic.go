package storeforward

import (
	"crypto/sha256"
	"fmt"
	"sort"
	"strings"
	"time"
)

const ReplicationTarget = 20

// Record represents a deterministic store-and-forward record.
type Record struct {
	ID        string
	Payload   []byte
	Metadata  map[string]string
	CreatedAt time.Time
	TTL       time.Duration
}

// NewRecord builds a record with predictable ID metadata.
func NewRecord(payload []byte, ttl time.Duration, metadata map[string]string) Record {
	copied := make(map[string]string)
	for k, v := range metadata {
		copied[k] = v
	}
	timestamp := time.Now().UTC()
	id := fmt.Sprintf("sf-%x", sha256.Sum256(append(payload, []byte(timestamp.String())...)))
	return Record{
		ID:        id,
		Payload:   payload,
		Metadata:  copied,
		CreatedAt: timestamp,
		TTL:       ttl,
	}
}

func (r Record) Expiration() time.Time {
	if r.TTL <= 0 {
		return r.CreatedAt
	}
	return r.CreatedAt.Add(r.TTL)
}

func (r Record) Remaining(now time.Time) time.Duration {
	return r.Expiration().Sub(now)
}

func (r Record) IsExpired(now time.Time) bool {
	return r.Remaining(now) <= 0
}

type PurgeReason string

const (
	ReasonExpired          PurgeReason = "purge.expired"
	ReasonPolicyEnforced   PurgeReason = "purge.policy"
	ReasonStoragePressured PurgeReason = "purge.storage"
	ReasonDegradedSignal   PurgeReason = "purge.degraded"
	ReasonRetained         PurgeReason = "purge.retained"
)

type PurgeClassification struct {
	RecordID    string
	Reason      PurgeReason
	Description string
	NextCheck   time.Time
	IsDegraded  bool
}

func ClassifyPurge(record Record, now time.Time, policyDays int, storagePressure float64) PurgeClassification {
	reason := ReasonRetained
	desc := "record retained"
	nextCheck := now.Add(24 * time.Hour)
	degraded := false

	if record.IsExpired(now) {
		reason = ReasonExpired
		desc = "ttl expired"
		nextCheck = record.Expiration()
	} else if policyDays > 0 {
		policyExpiry := record.CreatedAt.Add(time.Duration(policyDays) * 24 * time.Hour)
		if !now.Before(policyExpiry) {
			reason = ReasonPolicyEnforced
			desc = "retention policy triggered"
			nextCheck = policyExpiry
		}
	}

	if reason == ReasonRetained && storagePressure >= 0.8 {
		reason = ReasonStoragePressured
		desc = "storage pressure high"
		nextCheck = minTime(nextCheck, now.Add(time.Hour))
		if storagePressure >= 0.9 {
			degraded = true
		}
	}

	if reason == ReasonRetained && decryptedDegraded(record) {
		reason = ReasonDegradedSignal
		desc = "degraded replication"
		nextCheck = now.Add(30 * time.Minute)
		degraded = true
	}

	if reason == ReasonRetained {
		recordExpiry := record.Expiration()
		if recordExpiry.After(now) {
			nextCheck = minTime(nextCheck, recordExpiry)
		}
	}

	if storagePressure >= 0.9 {
		degraded = true
	}

	return PurgeClassification{
		RecordID:    record.ID,
		Reason:      reason,
		Description: desc,
		NextCheck:   nextCheck,
		IsDegraded:  degraded,
	}
}

func decryptedDegraded(record Record) bool {
	hint := record.Metadata["replication"]
	return strings.Contains(hint, "degraded")
}

func minTime(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}

// PeerInfo represents a candidate replication target.
type PeerInfo struct {
	ID    string
	Score int
	Role  string
}

type ReplicationPlan struct {
	RecordID string
	Targets  []PeerInfo
	Expected int
	Reason   string
}

func BuildReplicationPlan(record Record, peers []PeerInfo) ReplicationPlan {
	expected := ReplicationTarget
	if expected <= 0 {
		expected = ReplicationTarget
	}

	selected := make([]PeerInfo, len(peers))
	copy(selected, peers)
	sort.Slice(selected, func(i, j int) bool {
		if selected[i].Score == selected[j].Score {
			return selected[i].ID < selected[j].ID
		}
		return selected[i].Score > selected[j].Score
	})

	if len(selected) > expected {
		selected = selected[:expected]
	}

	reason := "replication.target.met"
	if len(selected) < expected {
		reason = fmt.Sprintf("replication.degraded.target=%d", expected)
	}

	return ReplicationPlan{
		RecordID: record.ID,
		Targets:  selected,
		Expected: expected,
		Reason:   reason,
	}
}

func (p ReplicationPlan) NeedsRepair() bool {
	return len(p.Targets) < p.Expected
}

type TTLRange struct {
	MinSeconds int
	MaxSeconds int
}

func NormalizeTTL(ttl time.Duration, fallback TTLRange) time.Duration {
	if ttl <= 0 {
		return time.Duration(fallback.MaxSeconds) * time.Second
	}
	if fallback.MinSeconds > 0 && ttl < time.Duration(fallback.MinSeconds)*time.Second {
		return time.Duration(fallback.MinSeconds) * time.Second
	}
	if fallback.MaxSeconds > 0 && ttl > time.Duration(fallback.MaxSeconds)*time.Second {
		return time.Duration(fallback.MaxSeconds) * time.Second
	}
	return ttl
}
