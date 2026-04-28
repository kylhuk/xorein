// Package chat implements the /aether/chat/0.1.0 protocol family.
// Source: docs/spec/v0.1/41-family-chat.md
package chat

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	libp2pprotocol "github.com/libp2p/go-libp2p/core/protocol"

	proto "github.com/aether/code_aether/pkg/v0_1/protocol"
	"github.com/aether/code_aether/pkg/v0_1/transport"
)

const protocolID = "/aether/chat/0.1.0"

// MessageRecord holds a received chat message.
type MessageRecord struct {
	ID         string
	ScopeID    string
	SenderID   string
	ScopeType  string // "channel" or "dm"
	Body       []byte
	ReceivedAt time.Time
}

// MuteChecker is satisfied by the moderation.Handler.
type MuteChecker interface {
	IsMuted(serverID, peerID string) bool
}

// Handler implements transport.FamilyHandler for the Chat family.
type Handler struct {
	mu          sync.RWMutex
	LocalPeerID string
	messages    []MessageRecord
	seenMsgIDs  map[string]struct{} // for DUPLICATE_DELIVERY detection
	channels    map[string]struct{} // known channel IDs
	moderation  MuteChecker         // optional; nil disables mute enforcement
}

// WithModeration wires in a moderation handler for mute enforcement.
func (h *Handler) WithModeration(m MuteChecker) { h.moderation = m }

// RegisterChannel registers a channel ID so chat.join can verify membership.
func (h *Handler) RegisterChannel(channelID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.channels == nil {
		h.channels = make(map[string]struct{})
	}
	h.channels[channelID] = struct{}{}
}

var _ transport.FamilyHandler = (*Handler)(nil)

func (h *Handler) ProtocolID() libp2pprotocol.ID { return libp2pprotocol.ID(protocolID) }

func (h *Handler) HandleStream(ctx context.Context, req *proto.PeerStreamRequest) *proto.PeerStreamResponse {
	required, known := proto.RequiredCapsFor(req.Operation)
	if !known {
		return errorResp(req.RequestID, proto.CodeUnsupportedOperation, "unknown: "+req.Operation, nil)
	}
	remote := toFeatureFlags(req.AdvertisedCaps)
	result := proto.NegotiateCapabilities(h.localCaps(), remote, required)
	if len(result.MissingRequired) > 0 {
		return errorResp(req.RequestID, proto.CodeMissingRequiredCapability, "missing caps", result.MissingRequired)
	}

	switch req.Operation {
	case "chat.send":
		return h.handleSend(req)
	case "chat.join":
		return h.handleJoin(req)
	case "chat.history":
		return h.handleHistory(req)
	default:
		return errorResp(req.RequestID, proto.CodeUnsupportedOperation, "not implemented: "+req.Operation, nil)
	}
}

func (h *Handler) localCaps() []proto.FeatureFlag {
	return []proto.FeatureFlag{"cap.chat"}
}

type sendPayload struct {
	MessageID string `json:"message_id,omitempty"`
	ScopeID   string `json:"scope_id"`
	ScopeType string `json:"scope_type"`
	SenderID  string `json:"sender_id"`
	Body      []byte `json:"body"`
	Mode      string `json:"mode,omitempty"` // expected security mode for this scope
}

func (h *Handler) handleSend(req *proto.PeerStreamRequest) *proto.PeerStreamResponse {
	var p sendPayload
	if err := json.Unmarshal(req.Payload, &p); err != nil {
		return errorResp(req.RequestID, proto.CodeOperationFailed, "invalid payload", nil)
	}
	// Mute enforcement (spec 50 §3.5): reject if sender is currently muted.
	if h.moderation != nil && p.SenderID != "" && h.moderation.IsMuted(p.ScopeID, p.SenderID) {
		return errorResp(req.RequestID, proto.CodeOperationFailed, "muted", nil)
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	// Duplicate delivery check: reject if message_id was already seen.
	if p.MessageID != "" {
		if h.seenMsgIDs == nil {
			h.seenMsgIDs = make(map[string]struct{})
		}
		if _, dup := h.seenMsgIDs[p.MessageID]; dup {
			return errorResp(req.RequestID, proto.CodeDuplicateDelivery, "duplicate message_id", nil)
		}
		h.seenMsgIDs[p.MessageID] = struct{}{}
	}
	h.messages = append(h.messages, MessageRecord{
		ID:         p.MessageID,
		ScopeID:    p.ScopeID,
		SenderID:   p.SenderID,
		ScopeType:  p.ScopeType,
		Body:       p.Body,
		ReceivedAt: time.Now(),
	})
	resp, _ := json.Marshal(map[string]any{"ok": true, "scope_id": p.ScopeID})
	return &proto.PeerStreamResponse{NegotiatedProtocol: protocolID, Payload: resp}
}

type joinPayload struct {
	ScopeID   string `json:"scope_id"`
	ScopeType string `json:"scope_type"`
	PeerID    string `json:"peer_id"`
}

func (h *Handler) handleJoin(req *proto.PeerStreamRequest) *proto.PeerStreamResponse {
	var p joinPayload
	if err := json.Unmarshal(req.Payload, &p); err != nil {
		return errorResp(req.RequestID, proto.CodeOperationFailed, "invalid payload", nil)
	}
	// Channel existence check: if channels are registered and scope_type is "channel",
	// reject with CHANNEL_NOT_FOUND when the scope is unknown.
	if p.ScopeType == "channel" {
		h.mu.RLock()
		_, known := h.channels[p.ScopeID]
		hasChannels := len(h.channels) > 0
		h.mu.RUnlock()
		if hasChannels && !known {
			return errorResp(req.RequestID, proto.CodeChannelNotFound, "channel not found: "+p.ScopeID, nil)
		}
	}
	resp, _ := json.Marshal(map[string]any{"ok": true, "scope_id": p.ScopeID, "peer_id": p.PeerID})
	return &proto.PeerStreamResponse{NegotiatedProtocol: protocolID, Payload: resp}
}

type historyPayload struct {
	ScopeID string `json:"scope_id"`
	Limit   int    `json:"limit"`
}

func (h *Handler) handleHistory(req *proto.PeerStreamRequest) *proto.PeerStreamResponse {
	var p historyPayload
	if err := json.Unmarshal(req.Payload, &p); err != nil {
		return errorResp(req.RequestID, proto.CodeOperationFailed, "invalid payload", nil)
	}
	h.mu.RLock()
	var msgs []MessageRecord
	for _, m := range h.messages {
		if m.ScopeID == p.ScopeID {
			msgs = append(msgs, m)
		}
	}
	h.mu.RUnlock()
	if p.Limit > 0 && len(msgs) > p.Limit {
		msgs = msgs[len(msgs)-p.Limit:]
	}
	resp, _ := json.Marshal(map[string]any{"scope_id": p.ScopeID, "count": len(msgs), "messages": msgs})
	return &proto.PeerStreamResponse{NegotiatedProtocol: protocolID, Payload: resp}
}

func (h *Handler) Messages(scopeID string) []MessageRecord {
	h.mu.RLock()
	defer h.mu.RUnlock()
	var out []MessageRecord
	for _, m := range h.messages {
		if m.ScopeID == scopeID {
			out = append(out, m)
		}
	}
	return out
}

func errorResp(requestID, code, msg string, missing []string) *proto.PeerStreamResponse {
	return &proto.PeerStreamResponse{RequestID: requestID, Error: proto.NewPeerStreamError(code, msg, missing)}
}

func toFeatureFlags(ss []string) []proto.FeatureFlag {
	out := make([]proto.FeatureFlag, len(ss))
	for i, s := range ss {
		out[i] = proto.FeatureFlag(s)
	}
	return out
}
