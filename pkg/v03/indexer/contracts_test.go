package indexer

import "testing"

func TestVerificationDecisionVariants(t *testing.T) {
	tests := []struct {
		state    SignatureVerificationState
		expected string
	}{
		{SignatureValid, "Accept payload; signature verified"},
		{SignatureMissing, "Treat payload as advisory; fall back to authoritative directory"},
		{SignatureStale, "Signal stale data; request fresh proof"},
		{SignatureInvalid, "Reject payload; signature verification failed"},
	}

	for _, tt := range tests {
		if got := VerificationDecision(tt.state); got != tt.expected {
			t.Fatalf("VerificationDecision(%q) = %q", tt.state, got)
		}
	}
}

func TestMultiIndexerMergeHintAuthoritative(t *testing.T) {
	if got := MultiIndexerMergeHint(true); got != "Honor authoritative source (if present) before community-run indexes" {
		t.Fatalf("unexpected authoritative hint: %q", got)
	}
	if got := MultiIndexerMergeHint(false); got != "Merge sorted by deterministic fingerprint and drop duplicates" {
		t.Fatalf("unexpected non-authoritative hint: %q", got)
	}
}
