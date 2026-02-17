package archivist

type EnrollmentState string

const (
	StateUnmapped  EnrollmentState = "unmapped"
	StateEnrolling EnrollmentState = "enrolling"
	StateActive    EnrollmentState = "active"
	StateWithdrawn EnrollmentState = "withdrawn"
	StateSuspended EnrollmentState = "suspended"
)

func CanTransition(from, to EnrollmentState) bool {
	if from == to {
		return true
	}
	switch from {
	case StateUnmapped:
		return to == StateEnrolling
	case StateEnrolling:
		return to == StateActive || to == StateWithdrawn
	case StateActive:
		return to == StateWithdrawn || to == StateSuspended
	case StateSuspended:
		return to == StateActive || to == StateWithdrawn
	default:
		return false
	}
}

type ArchivistCapability string

const (
	CapabilityVolunteer ArchivistCapability = "archivist.volunteer"
)

func IsVolunteerCapability(cap ArchivistCapability) bool {
	return cap == CapabilityVolunteer
}

type TransitionRecord struct {
	From   EnrollmentState
	To     EnrollmentState
	Reason string
}

func RecordTransition(from, to EnrollmentState, reason string) TransitionRecord {
	if !CanTransition(from, to) {
		reason = "invalid transition"
	}
	return TransitionRecord{From: from, To: to, Reason: reason}
}
