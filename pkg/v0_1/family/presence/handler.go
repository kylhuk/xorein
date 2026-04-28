// Package presence implements the /aether/presence/0.2.0 protocol family.
// Source: docs/spec/v0.1/48-family-presence.md
package presence

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	libp2pprotocol "github.com/libp2p/go-libp2p/core/protocol"

	proto "github.com/aether/code_aether/pkg/v0_1/protocol"
	"github.com/aether/code_aether/pkg/v0_1/transport"
)

const protocolID = "/aether/presence/0.2.0"

// Status represents a peer's online status.
type Status string

const (
	StatusOnline  Status = "online"
	StatusAway    Status = "away"
	StatusOffline Status = "offline"
	StatusIdle    Status = "idle"
	StatusDND     Status = "dnd"       // Do Not Disturb
	StatusInvisible Status = "invisible"
)

// PresenceRecord is a peer's current presence state.
type PresenceRecord struct {
	PeerID          string    `json:"peer_id"`
	Status          Status    `json:"status"`
	StatusText      string    `json:"status_text,omitempty"`
	UpdatedAt       time.Time `json:"updated_at"`
	StatusVersion   uint64    `json:"status_version,omitempty"` // monotonic counter
	IsTyping        bool      `json:"is_typing,omitempty"`
	TypingInScope   string    `json:"typing_in_scope,omitempty"`
}

// Handler implements transport.FamilyHandler for the Presence family.
type Handler struct {
	mu       sync.RWMutex
	presence map[string]*PresenceRecord
}

var _ transport.FamilyHandler = (*Handler)(nil)

func (h *Handler) ProtocolID() libp2pprotocol.ID { return libp2pprotocol.ID(protocolID) }

func (h *Handler) HandleStream(ctx context.Context, req *proto.PeerStreamRequest) *proto.PeerStreamResponse {
	switch req.Operation {
	case "presence.update":
		return h.handleUpdate(req)
	case "presence.query":
		return h.handleQuery(req)
	default:
		return errorResp(req.RequestID, proto.CodeUnsupportedOperation, "unknown: "+req.Operation, nil)
	}
}

type updatePayload struct {
	PeerID        string `json:"peer_id"`
	Status        Status `json:"status"`
	StatusText    string `json:"status_text,omitempty"`
	StatusVersion uint64 `json:"status_version,omitempty"`
	IsTyping      bool   `json:"is_typing,omitempty"`
	TypingInScope string `json:"typing_in_scope,omitempty"`
}

func (h *Handler) handleUpdate(req *proto.PeerStreamRequest) *proto.PeerStreamResponse {
	var p updatePayload
	if err := json.Unmarshal(req.Payload, &p); err != nil {
		return errorResp(req.RequestID, proto.CodeOperationFailed, "invalid payload", nil)
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.presence == nil {
		h.presence = make(map[string]*PresenceRecord)
	}
	// Monotonic version check: reject if new status_version <= existing status_version.
	if existing := h.presence[p.PeerID]; existing != nil && p.StatusVersion > 0 {
		if p.StatusVersion <= existing.StatusVersion {
			resp, _ := json.Marshal(map[string]any{"ok": false, "reason": "stale_status_version"})
			return &proto.PeerStreamResponse{NegotiatedProtocol: protocolID, Payload: resp}
		}
	}
	h.presence[p.PeerID] = &PresenceRecord{
		PeerID:        p.PeerID,
		Status:        p.Status,
		StatusText:    p.StatusText,
		UpdatedAt:     time.Now(),
		StatusVersion: p.StatusVersion,
		IsTyping:      p.IsTyping,
		TypingInScope: p.TypingInScope,
	}
	resp, _ := json.Marshal(map[string]any{"ok": true})
	return &proto.PeerStreamResponse{NegotiatedProtocol: protocolID, Payload: resp}
}

type queryPayload struct {
	PeerIDs []string `json:"peer_ids"`
}

func (h *Handler) handleQuery(req *proto.PeerStreamRequest) *proto.PeerStreamResponse {
	var p queryPayload
	if err := json.Unmarshal(req.Payload, &p); err != nil {
		return errorResp(req.RequestID, proto.CodeOperationFailed, "invalid payload", nil)
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	var results []*PresenceRecord
	for _, pid := range p.PeerIDs {
		if r := h.presence[pid]; r != nil {
			results = append(results, r)
		}
	}
	resp, _ := json.Marshal(map[string]any{"presence": results})
	return &proto.PeerStreamResponse{NegotiatedProtocol: protocolID, Payload: resp}
}

func errorResp(requestID, code, msg string, missing []string) *proto.PeerStreamResponse {
	return &proto.PeerStreamResponse{RequestID: requestID, Error: proto.NewPeerStreamError(code, msg, missing)}
}
