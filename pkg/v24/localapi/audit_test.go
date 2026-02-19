package localapi

import (
	"strings"
	"testing"
	"time"
)

func TestAuditSummaryAvoidsPayload(t *testing.T) {
	record := AuditRecord{
		Timestamp: time.Now(),
		RPC:       "Attach",
		Reason:    RefusalReasonUnauthorizedCapability,
		Outcome:   "denied",
	}

	summary := record.Summary()
	if containsPayload(summary) {
		t.Fatalf("audit summary must not expose payload: %s", summary)
	}
}

func containsPayload(text string) bool {
	return strings.Contains(text, "payload")
}
