package control

import (
	"net/http"
	"time"
	"unicode/utf8"
)

func (s *Server) handleListDMs(w http.ResponseWriter, r *http.Request) {
	s.st.mu.RLock()
	out := append([]*DMRecord(nil), s.st.dms...)
	s.st.mu.RUnlock()
	writeJSON(w, http.StatusOK, out)
}

type createDMRequest struct {
	PeerID string `json:"peer_id"`
}

func (s *Server) handleCreateDM(w http.ResponseWriter, r *http.Request) {
	var req createDMRequest
	if !decodeBody(w, r, &req) {
		return
	}
	if req.PeerID == "" {
		writeError(w, http.StatusBadRequest, CodeInvalidRequest, "peer_id is required")
		return
	}

	// Return existing DM record if one exists for this peer (200, not 201).
	s.st.mu.Lock()
	for _, dm := range s.st.dms {
		if dm.PeerID == req.PeerID {
			s.st.mu.Unlock()
			writeJSON(w, http.StatusOK, dm)
			return
		}
	}
	dm := &DMRecord{
		ID:        newID("dm"),
		PeerID:    req.PeerID,
		CreatedAt: time.Now().UTC(),
	}
	s.st.dms = append(s.st.dms, dm)
	s.st.mu.Unlock()
	s.sse.Publish(Event{Type: "dm", Data: dm})
	writeJSON(w, http.StatusCreated, dm)
}

func (s *Server) handleSendDMMessage(w http.ResponseWriter, r *http.Request) {
	dmID := r.PathValue("dmID")
	var req sendMessageRequest
	if !decodeBody(w, r, &req) {
		return
	}
	if utf8.RuneCountInString(req.Body) == 0 {
		writeError(w, http.StatusBadRequest, CodeInvalidRequest, "body must not be empty")
		return
	}
	if len(req.Body) > maxMsgBodyBytes {
		writeError(w, http.StatusRequestEntityTooLarge, CodeInvalidRequest, "message body exceeds 64 KiB limit")
		return
	}

	// Verify DM exists.
	s.st.mu.RLock()
	var found bool
	for _, dm := range s.st.dms {
		if dm.ID == dmID {
			found = true
			break
		}
	}
	s.st.mu.RUnlock()
	if !found {
		writeError(w, http.StatusNotFound, CodeNotFound, "DM not found")
		return
	}

	msg := &MessageRecord{
		ID:           newID("dmsg"),
		ScopeID:      dmID,
		ScopeType:    "dm",
		SenderPeerID: s.hs.PeerID,
		Body:         req.Body,
		CreatedAt:    time.Now().UTC(),
	}
	s.st.mu.Lock()
	s.st.messages = append(s.st.messages, msg)
	s.st.mu.Unlock()
	s.sse.Publish(Event{Type: "message", Data: msg})
	writeJSON(w, http.StatusCreated, msg)
}
