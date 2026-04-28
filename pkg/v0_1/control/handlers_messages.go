package control

import (
	"net/http"
	"strings"
	"time"
	"unicode/utf8"
)

const maxMsgBodyBytes = 64 << 10 // 64 KiB per spec 60 §6

type sendMessageRequest struct {
	Body string `json:"body"`
}

func (s *Server) handleSendChannelMessage(w http.ResponseWriter, r *http.Request) {
	channelID := r.PathValue("channelID")
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
	msg := &MessageRecord{
		ID:           newID("msg"),
		ScopeID:      channelID,
		ScopeType:    "channel",
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

type searchMessagesRequest struct {
	Query        string    `json:"query,omitempty"`
	ScopeType    string    `json:"scope_type,omitempty"`
	ScopeID      string    `json:"scope_id,omitempty"`
	ServerID     string    `json:"server_id,omitempty"`
	SenderPeerID string    `json:"sender_peer_id,omitempty"`
	Before       time.Time `json:"before,omitempty"`
	After        time.Time `json:"after,omitempty"`
	Limit        int       `json:"limit,omitempty"`
}

func (s *Server) handleSearchMessages(w http.ResponseWriter, r *http.Request) {
	var req searchMessagesRequest
	if !decodeBody(w, r, &req) {
		return
	}
	if req.Limit <= 0 || req.Limit > 200 {
		req.Limit = 50
	}

	s.st.mu.RLock()
	all := s.st.messages
	s.st.mu.RUnlock()

	var results []*MessageRecord
	q := strings.ToLower(req.Query)
	for _, m := range all {
		if m.DeletedAt != nil {
			continue
		}
		if req.ScopeType != "" && m.ScopeType != req.ScopeType {
			continue
		}
		if req.ScopeID != "" && m.ScopeID != req.ScopeID {
			continue
		}
		if req.SenderPeerID != "" && m.SenderPeerID != req.SenderPeerID {
			continue
		}
		if q != "" && !strings.Contains(strings.ToLower(m.Body), q) {
			continue
		}
		if !req.Before.IsZero() && !m.CreatedAt.Before(req.Before) {
			continue
		}
		if !req.After.IsZero() && !m.CreatedAt.After(req.After) {
			continue
		}
		results = append(results, m)
		if len(results) >= req.Limit {
			break
		}
	}

	ids := make([]string, len(results))
	for i, m := range results {
		ids[i] = m.ID
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"messages": ids,
		"results":  results,
	})
}

type editMessageRequest struct {
	Body string `json:"body"`
}

func (s *Server) handleEditMessage(w http.ResponseWriter, r *http.Request) {
	messageID := r.PathValue("messageID")
	var req editMessageRequest
	if !decodeBody(w, r, &req) {
		return
	}
	if len(req.Body) > maxMsgBodyBytes {
		writeError(w, http.StatusRequestEntityTooLarge, CodeInvalidRequest, "message body exceeds 64 KiB limit")
		return
	}

	s.st.mu.Lock()
	var found *MessageRecord
	for _, m := range s.st.messages {
		if m.ID == messageID {
			found = m
			break
		}
	}
	if found != nil && found.SenderPeerID == s.hs.PeerID {
		now := time.Now().UTC()
		found.Body = req.Body
		found.EditedAt = &now
	}
	s.st.mu.Unlock()

	if found == nil {
		writeError(w, http.StatusNotFound, CodeNotFound, "message not found")
		return
	}
	if found.SenderPeerID != s.hs.PeerID {
		writeError(w, http.StatusForbidden, CodeForbidden, "cannot edit another peer's message")
		return
	}
	writeJSON(w, http.StatusOK, found)
}

func (s *Server) handleDeleteMessage(w http.ResponseWriter, r *http.Request) {
	messageID := r.PathValue("messageID")

	s.st.mu.Lock()
	var found *MessageRecord
	for _, m := range s.st.messages {
		if m.ID == messageID {
			found = m
			break
		}
	}
	if found != nil && found.SenderPeerID == s.hs.PeerID {
		now := time.Now().UTC()
		found.DeletedAt = &now
	}
	s.st.mu.Unlock()

	if found == nil {
		writeError(w, http.StatusNotFound, CodeNotFound, "message not found")
		return
	}
	if found.SenderPeerID != s.hs.PeerID {
		writeError(w, http.StatusForbidden, CodeForbidden, "cannot delete another peer's message")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
