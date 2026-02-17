package archivist

import "testing"

func TestCanTransition(t *testing.T) {
	if !CanTransition(StateUnmapped, StateEnrolling) {
		t.Fatalf("expected unmapped -> enrolling to be allowed")
	}
	if CanTransition(StateActive, StateEnrolling) {
		t.Fatalf("expected active -> enrolling to be disallowed")
	}
}

func TestIsVolunteerCapability(t *testing.T) {
	if !IsVolunteerCapability(CapabilityVolunteer) {
		t.Fatalf("expected volunteer capability to be recognized")
	}
	if IsVolunteerCapability(ArchivistCapability("archivist.unknown")) {
		t.Fatalf("expected unknown capability to be rejected")
	}
}

func TestRecordTransitionNotesInvalid(t *testing.T) {
	rec := RecordTransition(StateActive, StateEnrolling, "test")
	if rec.Reason != "invalid transition" {
		t.Fatalf("expected invalid transition reason, got %s", rec.Reason)
	}
}
