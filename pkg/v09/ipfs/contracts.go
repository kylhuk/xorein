package ipfs

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// ContentMeta captures the deterministic metadata used for content addressing.
type ContentMeta struct {
	Name      string
	SizeBytes int64
	Owner     string
	CreatedAt int64
}

// ContentAddress returns a deterministic identifier for content metadata.
func ContentAddress(meta ContentMeta) string {
	hasher := sha256.New()
	hasher.Write([]byte(meta.Name))
	hasher.Write([]byte(meta.Owner))
	hasher.Write([]byte(fmt.Sprintf("%d", meta.SizeBytes)))
	hasher.Write([]byte(fmt.Sprintf("%d", meta.CreatedAt)))
	return hex.EncodeToString(hasher.Sum(nil))
}

// PinRole denotes who is responsible for a pin.
type PinRole string

const (
	PinRoleServerOwner PinRole = "server-owner"
	PinRoleRelay       PinRole = "relay"
)

// PinLifecycleStage tracks the lifecycle state of a pin.
type PinLifecycleStage string

const (
	PinStageQueued  PinLifecycleStage = "queued"
	PinStageActive  PinLifecycleStage = "active"
	PinStageExpired PinLifecycleStage = "expired"
)

// NextPinStage returns the next deterministic pin stage.
func NextPinStage(stage PinLifecycleStage) PinLifecycleStage {
	switch stage {
	case PinStageQueued:
		return PinStageActive
	case PinStageActive:
		return PinStageExpired
	default:
		return PinStageExpired
	}
}

// RetentionState describes how long content must remain available.
type RetentionState string

const (
	RetentionEphemeral RetentionState = "ephemeral"
	RetentionStaged    RetentionState = "staged"
	RetentionDurable   RetentionState = "durable"
)

// ClassifyRetention deterministically categorizes retention expectations.
func ClassifyRetention(sizeBytes int64, pinned bool) RetentionState {
	if pinned && sizeBytes >= 1_000_000_000 {
		return RetentionDurable
	}
	if pinned {
		return RetentionStaged
	}
	return RetentionEphemeral
}

// DegradedOutcome describes recovery guidance when content is unreachable.
type DegradedOutcome struct {
	Level    string
	Action   string
	Operator string
}

// DegradedBehavior classifies degraded-mode responses.
func DegradedBehavior(pinned bool, reachable bool) DegradedOutcome {
	if reachable {
		return DegradedOutcome{Level: "nominal", Action: "log-metrics", Operator: "n/a"}
	}
	if pinned {
		return DegradedOutcome{Level: "degraded", Action: "retry replicas", Operator: "server-owner"}
	}
	return DegradedOutcome{Level: "critical", Action: "fallback to recovery storage", Operator: "relay"}
}
