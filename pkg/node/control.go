package node

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

func (s *Service) controlMux() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/state", s.handleState)
	mux.HandleFunc("/v1/events", s.handleEvents)
	mux.HandleFunc("/v1/identities", s.handleIdentities)
	mux.HandleFunc("/v1/identities/backup", s.handleIdentityBackup)
	mux.HandleFunc("/v1/identities/restore", s.handleIdentityRestore)
	mux.HandleFunc("/v1/servers", s.handleServers)
	mux.HandleFunc("/v1/servers/join", s.handleServersJoin)
	mux.HandleFunc("/v1/servers/preview", s.handleServersPreview)
	mux.HandleFunc("/v1/servers/", s.handleServerSubresources)
	mux.HandleFunc("/v1/presence", s.handlePresence)
	mux.HandleFunc("/v1/notifications/search", s.handleNotificationsSearch)
	mux.HandleFunc("/v1/notifications/summary", s.handleNotificationsSummary)
	mux.HandleFunc("/v1/notifications/read", s.handleNotificationsRead)
	mux.HandleFunc("/v1/mentions/search", s.handleMentionsSearch)
	mux.HandleFunc("/v1/peers/manual", s.handleManualPeers)
	mux.HandleFunc("/v1/dms", s.handleDMs)
	mux.HandleFunc("/v1/dms/", s.handleDMSubresources)
	mux.HandleFunc("/v1/channels/", s.handleChannelSubresources)
	mux.HandleFunc("/v1/messages/search", s.handleMessagesSearch)
	mux.HandleFunc("/v1/messages/", s.handleMessageSubresources)
	mux.HandleFunc("/v1/voice/", s.handleVoiceSubresources)
	return withLocalOnly(s.requireAuth(mux))
}

func withLocalOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if addr, ok := r.Context().Value(http.LocalAddrContextKey).(net.Addr); ok {
			if network := strings.TrimSpace(addr.Network()); network == "unix" || strings.HasPrefix(network, "unix") {
				next.ServeHTTP(w, r)
				return
			}
		}
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			host = strings.TrimSpace(r.RemoteAddr)
		}
		ip := net.ParseIP(host)
		if ip == nil || !ip.IsLoopback() {
			writeError(w, http.StatusForbidden, APIError{Code: "forbidden", Message: "control API is local-only"})
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Service) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := strings.TrimSpace(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
		if token == "" || token != s.ControlToken() {
			writeError(w, http.StatusUnauthorized, APIError{Code: "unauthorized", Message: "missing or invalid control token"})
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Service) handleState(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, APIError{Code: "method_not_allowed", Message: "method not allowed"})
		return
	}
	writeJSON(w, http.StatusOK, s.Snapshot())
}

func (s *Service) handleEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, APIError{Code: "method_not_allowed", Message: "method not allowed"})
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, APIError{Code: "unsupported", Message: "streaming unsupported"})
		return
	}
	events, cancel := s.Subscribe()
	defer cancel()
	writer := bufio.NewWriter(w)
	defer writer.Flush()
	fmt.Fprintf(writer, "event: ready\ndata: {\"version\":\"%s\"}\n\n", ControlAPIVersion)
	writer.Flush()
	flusher.Flush()
	for {
		select {
		case <-r.Context().Done():
			return
		case event, ok := <-events:
			if !ok {
				return
			}
			raw, _ := json.Marshal(event)
			fmt.Fprintf(writer, "event: %s\ndata: %s\n\n", event.Type, raw)
			writer.Flush()
			flusher.Flush()
		}
	}
}

func (s *Service) handleIdentities(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, s.Snapshot().Identity)
	case http.MethodPost:
		var req CreateIdentityRequest
		if err := decodeJSON(r.Body, &req); err != nil {
			writeError(w, http.StatusBadRequest, APIError{Code: "invalid_request", Message: err.Error()})
			return
		}
		identity, err := s.CreateIdentity(req.DisplayName, req.Bio)
		if err != nil {
			writeError(w, http.StatusBadRequest, APIError{Code: "invalid_request", Message: err.Error()})
			return
		}
		writeJSON(w, http.StatusCreated, identity)
	default:
		writeError(w, http.StatusMethodNotAllowed, APIError{Code: "method_not_allowed", Message: "method not allowed"})
	}
}

func (s *Service) handleIdentityBackup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, APIError{Code: "method_not_allowed", Message: "method not allowed"})
		return
	}
	raw, err := s.BackupIdentity()
	if err != nil {
		writeError(w, http.StatusInternalServerError, APIError{Code: "backup_failed", Message: err.Error()})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(raw)
}

func (s *Service) handleIdentityRestore(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, APIError{Code: "method_not_allowed", Message: "method not allowed"})
		return
	}
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, APIError{Code: "invalid_request", Message: err.Error()})
		return
	}
	identity, err := s.RestoreIdentity(raw)
	if err != nil {
		writeError(w, http.StatusBadRequest, APIError{Code: "invalid_request", Message: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, identity)
}

func (s *Service) handleServers(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, s.Snapshot().Servers)
	case http.MethodPost:
		var req CreateServerRequest
		if err := decodeJSON(r.Body, &req); err != nil {
			writeError(w, http.StatusBadRequest, APIError{Code: "invalid_request", Message: err.Error()})
			return
		}
		server, err := s.CreateServer(req.Name, req.Description)
		if err != nil {
			writeError(w, http.StatusBadRequest, APIError{Code: "invalid_request", Message: err.Error()})
			return
		}
		writeJSON(w, http.StatusCreated, server)
	default:
		writeError(w, http.StatusMethodNotAllowed, APIError{Code: "method_not_allowed", Message: "method not allowed"})
	}
}

func (s *Service) handleServersJoin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, APIError{Code: "method_not_allowed", Message: "method not allowed"})
		return
	}
	var req JoinServerRequest
	if err := decodeJSON(r.Body, &req); err != nil {
		writeError(w, http.StatusBadRequest, APIError{Code: "invalid_request", Message: err.Error()})
		return
	}
	server, err := s.JoinByDeeplink(req.Deeplink)
	if err != nil {
		code := "join_failed"
		if strings.Contains(err.Error(), "expired") {
			code = "expired_invite"
		}
		writeError(w, http.StatusBadRequest, APIError{Code: code, Message: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, server)
}

func (s *Service) handleServersPreview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, APIError{Code: "method_not_allowed", Message: "method not allowed"})
		return
	}
	var req PreviewServerRequest
	if err := decodeJSON(r.Body, &req); err != nil {
		writeError(w, http.StatusBadRequest, APIError{Code: "invalid_request", Message: err.Error()})
		return
	}
	preview, err := s.PreviewDeeplink(req.Deeplink)
	if err != nil {
		code := "preview_failed"
		if strings.Contains(err.Error(), "expired") {
			code = "expired_invite"
		}
		writeError(w, http.StatusBadRequest, APIError{Code: code, Message: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, preview)
}

func (s *Service) handleServerSubresources(w http.ResponseWriter, r *http.Request) {
	trimmed := strings.TrimPrefix(r.URL.Path, "/v1/servers/")
	parts := strings.Split(trimmed, "/")
	if len(parts) == 2 && parts[1] == "channels" && r.Method == http.MethodPost {
		var req CreateChannelRequest
		if err := decodeJSON(r.Body, &req); err != nil {
			writeError(w, http.StatusBadRequest, APIError{Code: "invalid_request", Message: err.Error()})
			return
		}
		channel, err := s.CreateChannel(parts[0], req.Name, req.Voice)
		if err != nil {
			writeError(w, http.StatusBadRequest, APIError{Code: "invalid_request", Message: err.Error()})
			return
		}
		writeJSON(w, http.StatusCreated, channel)
		return
	}
	writeError(w, http.StatusNotFound, APIError{Code: "not_found", Message: "endpoint not found"})
}

func (s *Service) handlePresence(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, APIError{Code: "method_not_allowed", Message: "method not allowed"})
		return
	}
	writeJSON(w, http.StatusOK, PresenceResponse{Peers: s.Presence()})
}

func (s *Service) handleNotificationsSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, APIError{Code: "method_not_allowed", Message: "method not allowed"})
		return
	}
	var req SearchNotificationsRequest
	if err := decodeJSON(r.Body, &req); err != nil {
		writeError(w, http.StatusBadRequest, APIError{Code: "invalid_request", Message: err.Error()})
		return
	}
	resp, err := s.SearchNotifications(req)
	if err != nil {
		writeError(w, http.StatusBadRequest, APIError{Code: "invalid_request", Message: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Service) handleNotificationsRead(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, APIError{Code: "method_not_allowed", Message: "method not allowed"})
		return
	}
	var req MarkNotificationsReadRequest
	if err := decodeJSON(r.Body, &req); err != nil {
		writeError(w, http.StatusBadRequest, APIError{Code: "invalid_request", Message: err.Error()})
		return
	}
	readThrough, err := s.MarkNotificationsReadScoped(req)
	if err != nil {
		writeError(w, http.StatusBadRequest, APIError{Code: "invalid_request", Message: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, s.notificationsReadResponse(strings.TrimSpace(req.ServerID), strings.TrimSpace(req.ScopeType), strings.TrimSpace(req.ScopeID), readThrough))
}

func (s *Service) handleNotificationsSummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, APIError{Code: "method_not_allowed", Message: "method not allowed"})
		return
	}
	writeJSON(w, http.StatusOK, s.NotificationSummary())
}

func (s *Service) handleMessagesSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, APIError{Code: "method_not_allowed", Message: "method not allowed"})
		return
	}
	var req SearchMessagesRequest
	if err := decodeJSON(r.Body, &req); err != nil {
		writeError(w, http.StatusBadRequest, APIError{Code: "invalid_request", Message: err.Error()})
		return
	}
	messages, err := s.SearchMessages(req)
	if err != nil {
		writeError(w, http.StatusBadRequest, APIError{Code: "invalid_request", Message: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, SearchMessagesResponse{Messages: messages, Results: s.messageSearchRecords(messages)})
}

func (s *Service) handleMentionsSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, APIError{Code: "method_not_allowed", Message: "method not allowed"})
		return
	}
	var req SearchMentionsRequest
	if err := decodeJSON(r.Body, &req); err != nil {
		writeError(w, http.StatusBadRequest, APIError{Code: "invalid_request", Message: err.Error()})
		return
	}
	mentions, err := s.SearchMentions(req)
	if err != nil {
		writeError(w, http.StatusBadRequest, APIError{Code: "invalid_request", Message: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, SearchMentionsResponse{Mentions: mentions})
}

func (s *Service) handleManualPeers(w http.ResponseWriter, r *http.Request) {
	var req ManualPeerRequest
	switch r.Method {
	case http.MethodPost:
		if err := decodeJSON(r.Body, &req); err != nil {
			writeError(w, http.StatusBadRequest, APIError{Code: "invalid_request", Message: err.Error()})
			return
		}
		if err := s.AddManualPeer(req.Address); err != nil {
			writeError(w, http.StatusBadRequest, APIError{Code: "invalid_request", Message: err.Error()})
			return
		}
		writeJSON(w, http.StatusNoContent, nil)
	case http.MethodDelete:
		if err := decodeJSON(r.Body, &req); err != nil {
			writeError(w, http.StatusBadRequest, APIError{Code: "invalid_request", Message: err.Error()})
			return
		}
		if err := s.RemoveManualPeer(req.Address); err != nil {
			writeError(w, http.StatusBadRequest, APIError{Code: "invalid_request", Message: err.Error()})
			return
		}
		writeJSON(w, http.StatusNoContent, nil)
	default:
		writeError(w, http.StatusMethodNotAllowed, APIError{Code: "method_not_allowed", Message: "method not allowed"})
	}
}

func (s *Service) handleDMs(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, s.Snapshot().DMs)
	case http.MethodPost:
		var req CreateDMRequest
		if err := decodeJSON(r.Body, &req); err != nil {
			writeError(w, http.StatusBadRequest, APIError{Code: "invalid_request", Message: err.Error()})
			return
		}
		dm, err := s.CreateDM(req.PeerID)
		if err != nil {
			writeError(w, http.StatusBadRequest, APIError{Code: "invalid_request", Message: err.Error()})
			return
		}
		writeJSON(w, http.StatusCreated, dm)
	default:
		writeError(w, http.StatusMethodNotAllowed, APIError{Code: "method_not_allowed", Message: "method not allowed"})
	}
}

func (s *Service) handleDMSubresources(w http.ResponseWriter, r *http.Request) {
	trimmed := strings.TrimPrefix(r.URL.Path, "/v1/dms/")
	parts := strings.Split(trimmed, "/")
	if len(parts) == 2 && parts[1] == "messages" && r.Method == http.MethodPost {
		var req SendMessageRequest
		if err := decodeJSON(r.Body, &req); err != nil {
			writeError(w, http.StatusBadRequest, APIError{Code: "invalid_request", Message: err.Error()})
			return
		}
		msg, err := s.SendDMMessage(parts[0], req.Body)
		if err != nil {
			writeError(w, http.StatusBadRequest, APIError{Code: "invalid_request", Message: err.Error()})
			return
		}
		writeJSON(w, http.StatusCreated, msg)
		return
	}
	writeError(w, http.StatusNotFound, APIError{Code: "not_found", Message: "endpoint not found"})
}

func (s *Service) handleChannelSubresources(w http.ResponseWriter, r *http.Request) {
	trimmed := strings.TrimPrefix(r.URL.Path, "/v1/channels/")
	parts := strings.Split(trimmed, "/")
	if len(parts) == 2 && parts[1] == "messages" && r.Method == http.MethodPost {
		var req SendMessageRequest
		if err := decodeJSON(r.Body, &req); err != nil {
			writeError(w, http.StatusBadRequest, APIError{Code: "invalid_request", Message: err.Error()})
			return
		}
		msg, err := s.SendChannelMessage(parts[0], req.Body)
		if err != nil {
			writeError(w, http.StatusBadRequest, APIError{Code: "invalid_request", Message: err.Error()})
			return
		}
		writeJSON(w, http.StatusCreated, msg)
		return
	}
	writeError(w, http.StatusNotFound, APIError{Code: "not_found", Message: "endpoint not found"})
}

func (s *Service) handleMessageSubresources(w http.ResponseWriter, r *http.Request) {
	messageID := strings.TrimPrefix(r.URL.Path, "/v1/messages/")
	switch r.Method {
	case http.MethodPatch:
		var req EditMessageRequest
		if err := decodeJSON(r.Body, &req); err != nil {
			writeError(w, http.StatusBadRequest, APIError{Code: "invalid_request", Message: err.Error()})
			return
		}
		msg, err := s.EditMessage(messageID, req.Body)
		if err != nil {
			writeError(w, http.StatusBadRequest, APIError{Code: "invalid_request", Message: err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, msg)
	case http.MethodDelete:
		if err := s.DeleteMessage(messageID); err != nil {
			writeError(w, http.StatusBadRequest, APIError{Code: "invalid_request", Message: err.Error()})
			return
		}
		writeJSON(w, http.StatusNoContent, nil)
	default:
		writeError(w, http.StatusMethodNotAllowed, APIError{Code: "method_not_allowed", Message: "method not allowed"})
	}
}

func (s *Service) handleVoiceSubresources(w http.ResponseWriter, r *http.Request) {
	trimmed := strings.TrimPrefix(r.URL.Path, "/v1/voice/")
	parts := strings.Split(trimmed, "/")
	if len(parts) != 2 {
		writeError(w, http.StatusNotFound, APIError{Code: "not_found", Message: "endpoint not found"})
		return
	}
	channelID, action := parts[0], parts[1]
	switch action {
	case "join":
		var req VoiceJoinRequest
		if err := decodeJSON(r.Body, &req); err != nil {
			writeError(w, http.StatusBadRequest, APIError{Code: "invalid_request", Message: err.Error()})
			return
		}
		if err := s.JoinVoice(channelID, req.Muted); err != nil {
			writeError(w, http.StatusBadRequest, APIError{Code: "invalid_request", Message: err.Error()})
			return
		}
		writeJSON(w, http.StatusNoContent, nil)
	case "leave":
		if err := s.LeaveVoice(channelID); err != nil {
			writeError(w, http.StatusBadRequest, APIError{Code: "invalid_request", Message: err.Error()})
			return
		}
		writeJSON(w, http.StatusNoContent, nil)
	case "mute":
		var req VoiceJoinRequest
		if err := decodeJSON(r.Body, &req); err != nil {
			writeError(w, http.StatusBadRequest, APIError{Code: "invalid_request", Message: err.Error()})
			return
		}
		if err := s.SetVoiceMuted(channelID, req.Muted); err != nil {
			writeError(w, http.StatusBadRequest, APIError{Code: "invalid_request", Message: err.Error()})
			return
		}
		writeJSON(w, http.StatusNoContent, nil)
	case "frames":
		var req VoiceFrameRequest
		if err := decodeJSON(r.Body, &req); err != nil {
			writeError(w, http.StatusBadRequest, APIError{Code: "invalid_request", Message: err.Error()})
			return
		}
		if err := s.SendVoiceFrame(channelID, req.Data); err != nil {
			writeError(w, http.StatusBadRequest, APIError{Code: "invalid_request", Message: err.Error()})
			return
		}
		writeJSON(w, http.StatusNoContent, nil)
	default:
		writeError(w, http.StatusNotFound, APIError{Code: "not_found", Message: "endpoint not found"})
	}
}

func NewControlClient(endpoint, token string) (*http.Client, string) {
	endpoint = strings.TrimSpace(endpoint)
	if runtime.GOOS == "windows" {
		return &http.Client{Timeout: time.Second}, "http://" + strings.TrimRight(endpoint, "/")
	}
	transport := &http.Transport{}
	transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		var d net.Dialer
		return d.DialContext(ctx, "unix", endpoint)
	}
	client := &http.Client{Transport: transport, Timeout: 2 * time.Second}
	return client, "http://unix"
}

func CallControlJSON(endpoint, token, method, path string, body any, out any) error {
	client, base := NewControlClient(endpoint, token)
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(raw)
	}
	req, err := http.NewRequest(method, base+path, reader)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return parseAPIError(resp)
	}
	if out == nil || resp.StatusCode == http.StatusNoContent {
		return nil
	}
	return decodeJSON(resp.Body, out)
}

func ControlTokenFromDataDir(dataDir string) (string, error) {
	raw, err := os.ReadFile(filepath.Join(dataDir, "control.token"))
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(raw)), nil
}
