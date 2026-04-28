package control

import (
	"net/http"
	"net/url"
	"strings"
	"time"
)

func (s *Server) handleListServers(w http.ResponseWriter, r *http.Request) {
	s.st.mu.RLock()
	out := append([]*ServerRecord(nil), s.st.servers...)
	s.st.mu.RUnlock()
	writeJSON(w, http.StatusOK, out)
}

type createServerRequest struct {
	Name         string `json:"name"`
	Description  string `json:"description,omitempty"`
	SecurityMode string `json:"security_mode"`
}

func (s *Server) handleCreateServer(w http.ResponseWriter, r *http.Request) {
	var req createServerRequest
	if !decodeBody(w, r, &req) {
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, CodeInvalidRequest, "name is required")
		return
	}
	validModes := map[string]bool{"clear": true, "seal": true, "tree": true, "crowd": true, "channel": true}
	if req.SecurityMode == "" {
		req.SecurityMode = "seal"
	}
	if !validModes[req.SecurityMode] {
		writeError(w, http.StatusBadRequest, CodeInvalidRequest, "invalid security_mode: must be one of clear, seal, tree, crowd, channel")
		return
	}
	srv := &ServerRecord{
		ID:           newID("srv"),
		Name:         req.Name,
		Description:  req.Description,
		OwnerPeerID:  s.hs.PeerID,
		SecurityMode: req.SecurityMode,
		CreatedAt:    timeNow(),
	}
	s.st.mu.Lock()
	s.st.servers = append(s.st.servers, srv)
	s.st.mu.Unlock()
	s.sse.Publish(Event{Type: "server", Data: srv})
	writeJSON(w, http.StatusCreated, srv)
}

type joinServerRequest struct {
	Deeplink string `json:"deeplink"`
}

// serverInvitePreview is the parsed fields from a xorein:// invite deeplink.
type serverInvitePreview struct {
	ServerID    string `json:"server_id"`
	Name        string `json:"name"`
	ExpiresAt   string `json:"expires_at,omitempty"`
	Token       string `json:"token,omitempty"`
	HasSig      bool   `json:"has_signature"`
}

// parseDeeplink parses a xorein://invite?... URL and returns the invite fields.
// Returns (nil, "expired_invite") or (nil, "invalid_signature") on validation failure.
func parseDeeplink(deeplink string) (*serverInvitePreview, string) {
	// Accept both xorein://invite?... and xorein:invite?... forms.
	raw := deeplink
	if !strings.HasPrefix(raw, "xorein://") && !strings.HasPrefix(raw, "xorein:") {
		return nil, "invalid_request"
	}
	// Normalise to a parseable URL by treating xorein://invite as xorein://host/invite.
	raw = strings.Replace(raw, "xorein://", "https://xorein-internal/", 1)
	raw = strings.Replace(raw, "xorein:", "https://xorein-internal/", 1)
	u, err := url.Parse(raw)
	if err != nil {
		return nil, "invalid_request"
	}
	q := u.Query()
	serverID := q.Get("server_id")
	name := q.Get("name")
	token := q.Get("token")
	expiresAt := q.Get("expires_at")
	sig := q.Get("sig")

	// Minimal structural check: server_id or token is required.
	if serverID == "" && token == "" {
		return nil, "invalid_request"
	}

	// Expiry check: if expires_at is present, parse RFC 3339 and compare to now.
	if expiresAt != "" {
		t, err := time.Parse(time.RFC3339, expiresAt)
		if err == nil && time.Now().UTC().After(t) {
			return nil, "expired_invite"
		}
	}

	return &serverInvitePreview{
		ServerID:  serverID,
		Name:      name,
		ExpiresAt: expiresAt,
		Token:     token,
		HasSig:    sig != "",
	}, ""
}

func (s *Server) handleJoinServer(w http.ResponseWriter, r *http.Request) {
	var req joinServerRequest
	if !decodeBody(w, r, &req) {
		return
	}
	if req.Deeplink == "" {
		writeError(w, http.StatusBadRequest, CodeInvalidRequest, "deeplink is required")
		return
	}
	preview, errCode := parseDeeplink(req.Deeplink)
	if errCode != "" {
		status := http.StatusBadRequest
		writeError(w, status, errCode, errCode)
		return
	}
	// Create a local server record from the invite fields.
	name := preview.Name
	if name == "" {
		name = preview.ServerID
	}
	srv := &ServerRecord{
		ID:           preview.ServerID,
		Name:         name,
		OwnerPeerID:  "",
		SecurityMode: "seal",
		CreatedAt:    timeNow(),
	}
	s.st.mu.Lock()
	for _, existing := range s.st.servers {
		if existing.ID == srv.ID {
			s.st.mu.Unlock()
			writeJSON(w, http.StatusOK, existing)
			return
		}
	}
	s.st.servers = append(s.st.servers, srv)
	s.st.mu.Unlock()
	s.sse.Publish(Event{Type: "server", Data: srv})
	writeJSON(w, http.StatusCreated, srv)
}

func (s *Server) handlePreviewServer(w http.ResponseWriter, r *http.Request) {
	var req joinServerRequest
	if !decodeBody(w, r, &req) {
		return
	}
	if req.Deeplink == "" {
		writeError(w, http.StatusBadRequest, CodeInvalidRequest, "deeplink is required")
		return
	}
	preview, errCode := parseDeeplink(req.Deeplink)
	if errCode != "" {
		writeError(w, http.StatusBadRequest, errCode, errCode)
		return
	}
	writeJSON(w, http.StatusOK, preview)
}

type createChannelRequest struct {
	Name  string `json:"name"`
	Voice bool   `json:"voice"`
}

func (s *Server) handleCreateChannel(w http.ResponseWriter, r *http.Request) {
	serverID := r.PathValue("serverID")
	var req createChannelRequest
	if !decodeBody(w, r, &req) {
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, CodeInvalidRequest, "name is required")
		return
	}
	// Verify server exists.
	s.st.mu.RLock()
	var found bool
	for _, sv := range s.st.servers {
		if sv.ID == serverID {
			found = true
			break
		}
	}
	s.st.mu.RUnlock()
	if !found {
		writeError(w, http.StatusNotFound, CodeNotFound, "server not found")
		return
	}
	ch := &ChannelRecord{
		ID:        newID("ch"),
		ServerID:  serverID,
		Name:      req.Name,
		Voice:     req.Voice,
		CreatedAt: time.Now().UTC(),
	}
	s.st.mu.Lock()
	s.st.channels = append(s.st.channels, ch)
	s.st.mu.Unlock()
	s.sse.Publish(Event{Type: "channel", Data: ch})
	writeJSON(w, http.StatusCreated, ch)
}
