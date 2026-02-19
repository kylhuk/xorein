package security

import "fmt"

// RefusalReason enumerates deterministic reasons for refusing abuse requests.
type RefusalReason string

const (
	// RefusalReasonQuotaEntriesExceeded indicates the per-Archivist entry quota was exceeded.
	RefusalReasonQuotaEntriesExceeded RefusalReason = "quota_entries_exceeded"
	// RefusalReasonQuotaRetentionExceeded indicates retention policy would be violated.
	RefusalReasonQuotaRetentionExceeded RefusalReason = "quota_retention_exceeded"
)

// RefusalError surfaces deterministic refusal reasons with optional context.
type RefusalError struct {
	Reason  RefusalReason
	Details string
}

// Error implements the error interface.
func (e *RefusalError) Error() string {
	if e == nil {
		return ""
	}
	if e.Details == "" {
		return fmt.Sprintf("%s", e.Reason)
	}
	return fmt.Sprintf("%s: %s", e.Reason, e.Details)
}

// QuotaEnforcer enforces conservative quotas and retention defaults.
type QuotaEnforcer struct {
	MaxEntries       int
	MaxRetentionDays int
}

// Enforce ensures the provided usage stays within configured limits.
func (q *QuotaEnforcer) Enforce(entries int, retentionDays int) error {
	if q == nil {
		return nil
	}
	if q.MaxEntries > 0 && entries > q.MaxEntries {
		return &RefusalError{
			Reason:  RefusalReasonQuotaEntriesExceeded,
			Details: fmt.Sprintf("requesting %d/%d entries", entries, q.MaxEntries),
		}
	}
	if q.MaxRetentionDays > 0 && retentionDays > q.MaxRetentionDays {
		return &RefusalError{
			Reason:  RefusalReasonQuotaRetentionExceeded,
			Details: fmt.Sprintf("retention %d days exceeds %d-day default", retentionDays, q.MaxRetentionDays),
		}
	}
	return nil
}
