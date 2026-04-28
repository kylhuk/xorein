package presence

// All returns a snapshot of all known presence records.
func (h *Handler) All() map[string]*PresenceRecord {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if h.presence == nil {
		return map[string]*PresenceRecord{}
	}
	out := make(map[string]*PresenceRecord, len(h.presence))
	for k, v := range h.presence {
		cp := *v
		out[k] = &cp
	}
	return out
}
