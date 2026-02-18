package ui

// CallState models the UI state for a voice session.
type CallState struct {
	Joined         bool
	Muted          bool
	Deafened       bool
	SelectedDevice string
	QualityScore   int
	RecoveryHint   bool
}

// ActionLabel returns the deterministic join/leave label.
func (s CallState) ActionLabel() string {
	if s.Joined {
		return "Leave"
	}
	return "Join"
}

// MuteLabel returns the label for mute/deafen controls.
func (s CallState) MuteLabel() string {
	if s.Muted {
		return "Unmute"
	}
	return "Mute"
}

// DeafLabel returns label for deafening.
func (s CallState) DeafLabel() string {
	if s.Deafened {
		return "Undeafen"
	}
	return "Deafen"
}

// DeviceLabel returns a deterministic device selection hint.
func (s CallState) DeviceLabel() string {
	if s.SelectedDevice == "" {
		return "Select Device"
	}
	return s.SelectedDevice
}

// QualityBadge derives a badge from the quality score.
func QualityBadge(score int) string {
	switch {
	case score >= 85:
		return "HD"
	case score >= 60:
		return "SD"
	default:
		return "Degraded"
	}
}

// NoLimboMessage surfaces deterministic guidance when call is unstable.
func NoLimboMessage(score int, recovery bool) string {
	if recovery {
		return "Reconnecting with recovery-first flow..."
	}
	if score < 50 {
		return "Network degraded; limited audio until recovery completes."
	}
	return "Audio stable"
}
