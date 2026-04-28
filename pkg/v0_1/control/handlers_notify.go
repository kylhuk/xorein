package control

import (
	"net/http"
	"time"
)

type searchNotificationsRequest struct {
	ServerID  string `json:"server_id,omitempty"`
	ScopeType string `json:"scope_type,omitempty"`
	ScopeID   string `json:"scope_id,omitempty"`
	UnreadOnly bool   `json:"unread_only,omitempty"`
	Limit      int    `json:"limit,omitempty"`
}

func (s *Server) handleSearchNotifications(w http.ResponseWriter, r *http.Request) {
	var req searchNotificationsRequest
	if !decodeBody(w, r, &req) {
		return
	}
	if req.Limit <= 0 || req.Limit > 200 {
		req.Limit = 50
	}

	s.st.mu.RLock()
	all := s.st.notifications
	s.st.mu.RUnlock()

	var results []*NotificationRecord
	for _, n := range all {
		if req.UnreadOnly && n.Read {
			continue
		}
		if req.ServerID != "" && n.ServerID != req.ServerID {
			continue
		}
		if req.ScopeType != "" && n.ScopeType != req.ScopeType {
			continue
		}
		if req.ScopeID != "" && n.ScopeID != req.ScopeID {
			continue
		}
		results = append(results, n)
		if len(results) >= req.Limit {
			break
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"notifications": results})
}

func (s *Server) handleNotificationSummary(w http.ResponseWriter, r *http.Request) {
	s.st.mu.RLock()
	all := s.st.notifications
	s.st.mu.RUnlock()

	var total, dmsUnread int
	byServer := make(map[string]map[string]int)
	for _, n := range all {
		if n.Read {
			continue
		}
		total++
		if n.ScopeType == "dm" {
			dmsUnread++
			continue
		}
		if n.ServerID == "" {
			continue
		}
		if byServer[n.ServerID] == nil {
			byServer[n.ServerID] = map[string]int{"unread": 0, "mentions": 0}
		}
		byServer[n.ServerID]["unread"]++
		if n.Type == "mention" {
			byServer[n.ServerID]["mentions"]++
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"total_unread": total,
		"by_server":    byServer,
		"dms_unread":   dmsUnread,
	})
}

type markReadRequest struct {
	ServerID              string `json:"server_id,omitempty"`
	ScopeType             string `json:"scope_type,omitempty"`
	ScopeID               string `json:"scope_id,omitempty"`
	ReadThroughMessageID  string `json:"read_through_message_id"`
}

func (s *Server) handleMarkRead(w http.ResponseWriter, r *http.Request) {
	var req markReadRequest
	if !decodeBody(w, r, &req) {
		return
	}

	s.st.mu.Lock()
	key := req.ScopeID
	if key == "" {
		key = req.ServerID
	}
	rt := &ReadThrough{
		ScopeID:          req.ScopeID,
		ScopeType:        req.ScopeType,
		ReadThroughMsgID: req.ReadThroughMessageID,
		UpdatedAt:        time.Now().UTC(),
	}
	s.st.readThrough[key] = rt
	// Mark matching notifications as read, respecting read_through_message_id cutoff.
	// IDs are "<prefix>-<nanosecond_timestamp>" so lexicographic comparison is time-ordered.
	for _, n := range s.st.notifications {
		if req.ScopeID != "" && n.ScopeID != req.ScopeID {
			continue
		}
		if req.ScopeType != "" && n.ScopeType != req.ScopeType {
			continue
		}
		// If a cutoff message ID is given, only mark notifications whose MessageID
		// is ≤ the cutoff (notifications with no MessageID are always marked read).
		if req.ReadThroughMessageID != "" && n.MessageID != "" && n.MessageID > req.ReadThroughMessageID {
			continue
		}
		n.Read = true
	}
	s.st.mu.Unlock()
	writeJSON(w, http.StatusOK, rt)
}

type searchMentionsRequest struct {
	ServerID string `json:"server_id,omitempty"`
	ScopeID  string `json:"scope_id,omitempty"`
	Limit    int    `json:"limit,omitempty"`
}

func (s *Server) handleSearchMentions(w http.ResponseWriter, r *http.Request) {
	var req searchMentionsRequest
	if !decodeBody(w, r, &req) {
		return
	}
	if req.Limit <= 0 || req.Limit > 200 {
		req.Limit = 50
	}

	s.st.mu.RLock()
	all := s.st.notifications
	s.st.mu.RUnlock()

	var results []*NotificationRecord
	for _, n := range all {
		if n.Type != "mention" {
			continue
		}
		if req.ServerID != "" && n.ServerID != req.ServerID {
			continue
		}
		if req.ScopeID != "" && n.ScopeID != req.ScopeID {
			continue
		}
		results = append(results, n)
		if len(results) >= req.Limit {
			break
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"mentions": results})
}
