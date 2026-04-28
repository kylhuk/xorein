package control

import "net/http"

type presenceEntry struct {
	Status        string `json:"status"`
	StatusText    string `json:"status_text,omitempty"`
	TypingInScope string `json:"typing_in_scope,omitempty"`
	UpdatedAt     string `json:"updated_at"`
}

func (s *Server) handleGetPresence(w http.ResponseWriter, r *http.Request) {
	peers := make(map[string]*presenceEntry)
	if s.hs.Presence != nil {
		for id, rec := range s.hs.Presence.All() {
			peers[id] = &presenceEntry{
				Status:        string(rec.Status),
				StatusText:    rec.StatusText,
				TypingInScope: rec.TypingInScope,
				UpdatedAt:     rec.UpdatedAt.UTC().Format("2006-01-02T15:04:05.999999999Z"),
			}
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"peers": peers})
}
