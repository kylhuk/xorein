package prekey

import "testing"

func TestBundleRecordValidate(t *testing.T) {
	ok := BundleRecord{IdentityID: "alice", SignedPrekeyID: "spk-1", OneTimePrekeyCount: 20, PublishedAtUnix: 1, ExpiresAtUnix: 2}
	if err := ok.Validate(); err != nil {
		t.Fatalf("validate ok: %v", err)
	}
	bad := BundleRecord{IdentityID: "alice", SignedPrekeyID: "spk-1", OneTimePrekeyCount: -1}
	if err := bad.Validate(); err == nil {
		t.Fatal("expected validation error")
	}
}

func TestDecideRotation(t *testing.T) {
	if got := DecideRotation(1, 3, 10, 20); got != RotationDecisionRepublish {
		t.Fatalf("got %q want %q", got, RotationDecisionRepublish)
	}
	if got := DecideRotation(10, 3, 20, 20); got != RotationDecisionRepublish {
		t.Fatalf("got %q want %q", got, RotationDecisionRepublish)
	}
	if got := DecideRotation(10, 3, 10, 20); got != RotationDecisionKeep {
		t.Fatalf("got %q want %q", got, RotationDecisionKeep)
	}
}
