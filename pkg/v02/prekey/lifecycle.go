package prekey

import "fmt"

// BundleRecord is the additive prekey publication contract.
type BundleRecord struct {
	IdentityID         string
	SignedPrekeyID     string
	OneTimePrekeyCount int
	PublishedAtUnix    uint64
	ExpiresAtUnix      uint64
}

func (r BundleRecord) Validate() error {
	if r.IdentityID == "" {
		return fmt.Errorf("identity id required")
	}
	if r.SignedPrekeyID == "" {
		return fmt.Errorf("signed prekey id required")
	}
	if r.OneTimePrekeyCount < 0 {
		return fmt.Errorf("one-time prekey count cannot be negative")
	}
	if r.ExpiresAtUnix != 0 && r.ExpiresAtUnix <= r.PublishedAtUnix {
		return fmt.Errorf("expiry must be after published timestamp")
	}
	return nil
}

type RotationDecision string

const (
	RotationDecisionKeep      RotationDecision = "keep"
	RotationDecisionRepublish RotationDecision = "republish"
)

func DecideRotation(remainingOneTime, minThreshold int, nowUnix, expiresAtUnix uint64) RotationDecision {
	if remainingOneTime <= minThreshold {
		return RotationDecisionRepublish
	}
	if expiresAtUnix != 0 && nowUnix >= expiresAtUnix {
		return RotationDecisionRepublish
	}
	return RotationDecisionKeep
}
