package voice

import "sort"

// ElectCoordinator returns the lex-smallest peer_id among participants,
// which becomes the SFU coordinator per spec 52 §3.12.
func ElectCoordinator(participants []string) string {
	if len(participants) == 0 {
		return ""
	}
	sorted := make([]string, len(participants))
	copy(sorted, participants)
	sort.Strings(sorted)
	return sorted[0]
}

// IsCoordinator checks whether peerID is the current coordinator for sessionID.
func (h *Handler) IsCoordinator(sessionID, peerID string) bool {
	h.mu.RLock()
	s := h.sessions[sessionID]
	h.mu.RUnlock()
	if s == nil {
		return false
	}
	return ElectCoordinator(s.Participants) == peerID
}
