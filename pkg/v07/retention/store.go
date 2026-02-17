package retention

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"time"
)

// PolicyRecord tracks audited retention policy changes.
type PolicyRecord struct {
	Policy    RetentionPolicy
	SignedBy  string
	Timestamp time.Time
	Signature string
	ChangeID  string
}

// PolicyStore keeps the last policy and history.
type PolicyStore struct {
	current PolicyRecord
	history []PolicyRecord
}

// NewPolicyStore bootstrap with an initial policy and signer.
func NewPolicyStore(initial RetentionPolicy, signer string) *PolicyStore {
	rec := PolicyRecord{
		Policy:    initial,
		SignedBy:  signer,
		Timestamp: time.Now().UTC(),
	}
	rec.Signature = computeSignature(rec)
	rec.ChangeID = fmt.Sprintf("policy-%s", rec.Signature[:8])
	return &PolicyStore{current: rec}
}

func computeSignature(rec PolicyRecord) string {
	data := fmt.Sprintf("%s:%s:%d:%d", rec.SignedBy, rec.Policy.Tier, rec.Policy.Days, rec.Timestamp.Unix())
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:8])
}

func (s *PolicyStore) CurrentPolicy() RetentionPolicy {
	return s.current.Policy
}

// ApplyChange enforces signature order deterministically.
func (s *PolicyStore) ApplyChange(policy RetentionPolicy, actor string) PolicyRecord {
	rec := PolicyRecord{
		Policy:    policy,
		SignedBy:  actor,
		Timestamp: time.Now().UTC(),
	}
	rec.Signature = computeSignature(rec)
	rec.ChangeID = fmt.Sprintf("policy-%s", rec.Signature[:8])
	s.history = append(s.history, s.current)
	s.current = rec
	return rec
}

// HistoryBoundary returns a cutoff timestamp for allowed history serving.
func (s *PolicyStore) HistoryBoundary(now time.Time) time.Time {
	days := s.current.Policy.Days
	if days <= 0 {
		days = 30
	}
	return now.Add(-time.Duration(days) * 24 * time.Hour)
}

// AcceptRecord decides if a record may be stored.
func (s *PolicyStore) AcceptRecord(created time.Time) bool {
	cutoff := s.HistoryBoundary(time.Now().UTC())
	return !created.Before(cutoff)
}

// BuildAuditTrail returns signed change history.
func (s *PolicyStore) BuildAuditTrail() []PolicyRecord {
	records := append([]PolicyRecord{}, s.history...)
	sort.Slice(records, func(i, j int) bool {
		return records[i].Timestamp.Before(records[j].Timestamp)
	})
	records = append(records, s.current)
	return records
}

// PolicyChangeEvent describes what changed.
type PolicyChangeEvent struct {
	Previous RetentionPolicy
	Current  RetentionPolicy
	Actor    string
	Record   PolicyRecord
}

// DescribeChange builds event data for audit events.
func (s *PolicyStore) DescribeChange(actor string) PolicyChangeEvent {
	prev := RetentionPolicy{Tier: TierEdge, Days: 14}
	if len(s.history) > 0 {
		prev = s.history[len(s.history)-1].Policy
	}
	return PolicyChangeEvent{
		Previous: prev,
		Current:  s.current.Policy,
		Actor:    actor,
		Record:   s.current,
	}
}

// EnforcePurge returns when next purge window occurs deterministically.
func (s *PolicyStore) EnforcePurge() (time.Duration, string) {
	days := s.current.Policy.Days
	if days <= 0 {
		days = 30
	}
	return time.Duration(days) * 24 * time.Hour, fmt.Sprintf("retention.%s", s.current.Policy.Tier)
}
