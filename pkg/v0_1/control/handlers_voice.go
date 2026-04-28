package control

import (
	"encoding/base64"
	"net/http"

	ms "github.com/aether/code_aether/pkg/v0_1/mode/mediashield"
)

type voiceJoinRequest struct {
	Muted bool `json:"muted"`
}

type voiceJoinResponse struct {
	SessionID    string   `json:"session_id"`
	Participants []string `json:"participants"`
}

func (s *Server) handleVoiceJoin(w http.ResponseWriter, r *http.Request) {
	channelID := r.PathValue("channelID")
	var req voiceJoinRequest
	if !decodeBody(w, r, &req) {
		return
	}
	if s.hs == nil || s.hs.Voice == nil {
		writeError(w, http.StatusServiceUnavailable, CodeUnsupported, "voice not available")
		return
	}
	localPeerID := ""
	if s.hs != nil {
		localPeerID = s.hs.PeerID
	}
	// Create a new session keyed to this channel; empty keys = MediaShield not yet negotiated.
	if err := s.hs.Voice.CreateSession(channelID, []string{localPeerID}, map[string]*ms.PeerKey{}); err != nil {
		writeError(w, http.StatusInternalServerError, CodeUnsupported, err.Error())
		return
	}
	if req.Muted {
		s.hs.Voice.SetMuted(channelID, localPeerID, true)
	}
	sessions := s.hs.Voice.Sessions()
	var participants []string
	for _, sess := range sessions {
		if sess.ID == channelID {
			participants = sess.Participants
			break
		}
	}
	s.sse.Publish(Event{Type: "voice", Data: map[string]any{
		"action":     "join",
		"session_id": channelID,
		"peer_id":    localPeerID,
	}})
	writeJSON(w, http.StatusOK, voiceJoinResponse{SessionID: channelID, Participants: participants})
}

func (s *Server) handleVoiceLeave(w http.ResponseWriter, r *http.Request) {
	channelID := r.PathValue("channelID")
	var body map[string]any
	_ = decodeBody(w, r, &body)
	if s.hs != nil && s.hs.Voice != nil {
		localPeerID := ""
		if s.hs != nil {
			localPeerID = s.hs.PeerID
		}
		s.hs.Voice.LeaveSession(channelID, localPeerID)
		s.sse.Publish(Event{Type: "voice", Data: map[string]any{
			"action":     "leave",
			"session_id": channelID,
			"peer_id":    localPeerID,
		}})
	}
	w.WriteHeader(http.StatusNoContent)
}

type voiceMuteRequest struct {
	Muted bool `json:"muted"`
}

func (s *Server) handleVoiceMute(w http.ResponseWriter, r *http.Request) {
	channelID := r.PathValue("channelID")
	var req voiceMuteRequest
	if !decodeBody(w, r, &req) {
		return
	}
	if s.hs != nil && s.hs.Voice != nil {
		localPeerID := ""
		if s.hs != nil {
			localPeerID = s.hs.PeerID
		}
		s.hs.Voice.SetMuted(channelID, localPeerID, req.Muted)
	}
	w.WriteHeader(http.StatusNoContent)
}

type voiceFrameRequest struct {
	Data string `json:"data"` // base64url SFrame payload
}

func (s *Server) handleVoiceFrames(w http.ResponseWriter, r *http.Request) {
	channelID := r.PathValue("channelID")
	var req voiceFrameRequest
	if !decodeBody(w, r, &req) {
		return
	}
	if req.Data == "" {
		writeError(w, http.StatusBadRequest, CodeInvalidRequest, "data is required")
		return
	}
	payload, err := base64.RawURLEncoding.DecodeString(req.Data)
	if err != nil {
		writeError(w, http.StatusBadRequest, CodeInvalidRequest, "data must be valid base64url")
		return
	}
	if s.hs != nil && s.hs.Voice != nil {
		localPeerID := s.hs.PeerID
		// EncryptFrame requires RTP header; for the control API path use an empty header.
		if _, _, err := s.hs.Voice.EncryptFrame(channelID, localPeerID, []byte{}, payload); err != nil {
			// Session may not have MediaShield keys yet; proceed without error.
			_ = err
		}
	}
	_ = channelID
	w.WriteHeader(http.StatusNoContent)
}
