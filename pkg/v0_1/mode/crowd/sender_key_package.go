// sender_key_package.go implements Crowd mode key distribution per spec 13 §3.3.
// Source: docs/spec/v0.1/13-mode-crowd.md §3.3
package crowd

import (
	"encoding/json"
	"fmt"
	"time"
)

// RotationTrigger identifies the reason for an epoch rotation.
type RotationTrigger string

const (
	RotationMembershipChange RotationTrigger = "membership_change"
	RotationMessageLimit     RotationTrigger = "message_limit"
	RotationTimeLimit        RotationTrigger = "time_limit"
)

// SenderKeyPackage is the cleartext payload that is Seal-encrypted and sent
// to each member via a crowd_key_distribution DM (spec 13 §3.3).
//
// Before encryption:
//
//	{
//	  "epoch_id": "<base64url 16B>",
//	  "epoch_root_secret": "<base64url 32B>",
//	  "rotation_trigger": "<membership_change|message_limit|time_limit>",
//	  "issued_at": 1234567890000,
//	  "expires_at": 1234567890000,
//	  "member_peer_ids": ["peer_id_1", ...]
//	}
type SenderKeyPackage struct {
	EpochID          []byte          `json:"epoch_id"`           // 16 bytes random
	EpochRootSecret  []byte          `json:"epoch_root_secret"`  // 32 bytes
	RotationTrigger  RotationTrigger `json:"rotation_trigger"`
	IssuedAt         int64           `json:"issued_at"`          // Unix ms
	ExpiresAt        int64           `json:"expires_at"`         // Unix ms (IssuedAt + EpochTTL)
	MemberPeerIDs    []string        `json:"member_peer_ids"`
}

// BuildSenderKeyPackage JSON-encodes the sender key package for a given epoch.
// The returned bytes are ready to be Seal-encrypted and sent as crowd_key_distribution.
//
// epochIDBytes must be 16 bytes of random epoch ID.
// epochRootSecret is the 32-byte epoch root secret for this epoch.
// trigger identifies why the rotation happened.
// memberPeerIDs is the list of all current member peer IDs.
//
// Note: encryption is left to the caller using their Seal session
// (pkg/v0_1/mode/seal) to maintain separation of concerns.
func BuildSenderKeyPackage(
	epochIDBytes []byte,
	epochRootSecret []byte,
	trigger RotationTrigger,
	memberPeerIDs []string,
) ([]byte, error) {
	if len(epochIDBytes) != 16 {
		return nil, fmt.Errorf("crowd: sender key package: epoch_id must be 16 bytes, got %d", len(epochIDBytes))
	}
	if len(epochRootSecret) != 32 {
		return nil, fmt.Errorf("crowd: sender key package: epoch_root_secret must be 32 bytes, got %d", len(epochRootSecret))
	}
	now := time.Now().UnixMilli()
	skp := &SenderKeyPackage{
		EpochID:         append([]byte(nil), epochIDBytes...),
		EpochRootSecret: append([]byte(nil), epochRootSecret...),
		RotationTrigger: trigger,
		IssuedAt:        now,
		ExpiresAt:       now + EpochTTL.Milliseconds(),
		MemberPeerIDs:   memberPeerIDs,
	}
	return json.Marshal(skp)
}

// DistributeSenderKeyPackage is a stub that returns the JSON-encoded SenderKeyPackage bytes.
// In production, the caller would iterate over members and use their Seal session to
// encrypt and send these bytes as a crowd_key_distribution delivery.
func DistributeSenderKeyPackage(skpJSON []byte) ([]byte, error) {
	return skpJSON, nil
}

// UnmarshalSenderKeyPackage deserializes a SenderKeyPackage from JSON.
func UnmarshalSenderKeyPackage(data []byte) (*SenderKeyPackage, error) {
	var skp SenderKeyPackage
	if err := json.Unmarshal(data, &skp); err != nil {
		return nil, fmt.Errorf("crowd: unmarshal sender key package: %w", err)
	}
	return &skp, nil
}
