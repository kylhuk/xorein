package discovery

import (
	"fmt"
	"sort"
	"time"

	"github.com/aether/code_aether/pkg/v06/conformance"
)

type DiscoveryReasonClass string

const (
	DiscoveryReasonFreshness   DiscoveryReasonClass = "VA-D1:discovery.freshness"
	DiscoveryReasonPoisoning   DiscoveryReasonClass = "VA-D2:discovery.poison"
	DiscoveryReasonConsistency DiscoveryReasonClass = "VA-D3:discovery.consistency"
)

type DiscoveryContract struct {
	ScopeID  string
	TaskID   string
	Artifact string
	Gate     conformance.GateID
	Reason   DiscoveryReasonClass
}

func NewDiscoveryContract(scope, task, artifact string, gate conformance.GateID, reason DiscoveryReasonClass) DiscoveryContract {
	return DiscoveryContract{
		ScopeID:  scope,
		TaskID:   task,
		Artifact: artifact,
		Gate:     gate,
		Reason:   reason,
	}
}

func (c DiscoveryContract) EvidenceAnchor() string {
	return fmt.Sprintf("%s#%s", c.Artifact, c.TaskID)
}

func (c DiscoveryContract) ReasonLabel() string {
	return string(c.Reason)
}

func DiscoveryReasonClasses() []DiscoveryReasonClass {
	return []DiscoveryReasonClass{
		DiscoveryReasonFreshness,
		DiscoveryReasonPoisoning,
		DiscoveryReasonConsistency,
	}
}

type FreshnessStatus string

const (
	FreshnessStatusValid FreshnessStatus = "valid"
	FreshnessStatusStale FreshnessStatus = "stale"
	FreshnessStatusRetry FreshnessStatus = "retry"
)

type FreshnessState struct {
	Status      FreshnessStatus
	EntryAge    time.Duration
	TTL         time.Duration
	Attempts    int
	ReasonLabel string
}

func AssessFreshness(entryAge, ttl time.Duration, attempts, maxAttempts int) FreshnessState {
	state := FreshnessState{
		EntryAge: entryAge,
		TTL:      ttl,
		Attempts: attempts,
	}
	switch {
	case entryAge <= ttl:
		state.Status = FreshnessStatusValid
		state.ReasonLabel = "discovery.freshness.success"
	case attempts < maxAttempts:
		state.Status = FreshnessStatusRetry
		state.ReasonLabel = "discovery.freshness.retry"
	default:
		state.Status = FreshnessStatusStale
		state.ReasonLabel = "discovery.freshness.stale"
	}
	return state
}

func RetryBackoff(attempt int, base time.Duration) time.Duration {
	if attempt <= 0 {
		return base
	}
	maxBackoff := 30 * time.Second
	backoff := time.Duration(1<<uint(attempt)) * base
	if backoff > maxBackoff {
		backoff = maxBackoff
	}
	return backoff
}

type PoisonClassification struct {
	Classification string
	ReasonLabel    string
}

func ClassifyPoisonAttempt(validSignature bool, conflictingPeers int) PoisonClassification {
	switch {
	case validSignature:
		return PoisonClassification{Classification: "clean", ReasonLabel: "discovery.poison.clean"}
	case conflictingPeers == 0:
		return PoisonClassification{Classification: "detected", ReasonLabel: "discovery.poison.detected"}
	default:
		return PoisonClassification{Classification: "invalid", ReasonLabel: "discovery.poison.invalid"}
	}
}

type ConflictResolution struct {
	CanonicalSource string
	ConflictCount   int
	ReasonLabel     string
}

func ResolveMultiSourceConflict(sources map[string]int64) ConflictResolution {
	if len(sources) == 0 {
		return ConflictResolution{ReasonLabel: "discovery.consistency.conflict"}
	}
	keys := make([]string, 0, len(sources))
	for k := range sources {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	canonical := keys[0]
	highest := sources[canonical]
	conflicts := 0
	for _, key := range keys[1:] {
		value := sources[key]
		if value > highest || (value == highest && key < canonical) {
			canonical = key
			highest = value
		}
		conflicts++
	}
	reason := "discovery.consistency.success"
	if conflicts > 0 {
		reason = "discovery.consistency.conflict"
	}
	return ConflictResolution{
		CanonicalSource: canonical,
		ConflictCount:   conflicts,
		ReasonLabel:     reason,
	}
}
