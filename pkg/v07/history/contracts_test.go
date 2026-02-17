package history

import "testing"

func TestCanonicalRoot(t *testing.T) {
	root := CanonicalRoot("base", 5, ModeEpochHistory)
	if root != "root:base:epoch:5" {
		t.Fatalf("unexpected canonical root %s", root)
	}
}

func TestClassifyProof(t *testing.T) {
	match := ClassifyProof("a", "a")
	if !match.Matched {
		t.Fatalf("expected match true")
	}
	mismatch := ClassifyProof("a", "b")
	if mismatch.Matched {
		t.Fatalf("expected mismatch false")
	}
}

func TestResumeSync(t *testing.T) {
	state := SyncState{Epoch: 3}
	next := ResumeSync(state)
	if !next.Ready {
		t.Fatalf("expected resume to ta ready")
	}
}

func TestNewCapsuleMetadata(t *testing.T) {
	meta := NewCapsuleMetadata("base", 1, ModeRolling)
	if meta.Root != CanonicalRoot("base", 1, ModeRolling) {
		t.Fatalf("expected root to match")
	}
}
