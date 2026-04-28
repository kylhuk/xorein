// Package notify implements the /aether/notify/0.2.0 protocol family.
// Source: docs/spec/v0.1/49-family-notify.md
package notify

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	libp2pprotocol "github.com/libp2p/go-libp2p/core/protocol"

	proto "github.com/aether/code_aether/pkg/v0_1/protocol"
	"github.com/aether/code_aether/pkg/v0_1/transport"
)

const protocolID = "/aether/notify/0.2.0"

// notifyRateWindow is the sliding window duration for notify rate limiting.
const notifyRateWindow = time.Minute

// notifyRateLimit is the maximum notifications pushed per sender per minute.
const notifyRateLimit = 60

// Notification is a push notification record.
type Notification struct {
	ID          string          `json:"id"`
	RecipientID string          `json:"recipient_id"`
	SenderID    string          `json:"sender_id,omitempty"`
	Type        string          `json:"type"`
	Payload     json.RawMessage `json:"payload"`
	BodyPreview string          `json:"body_preview,omitempty"` // spec 49: plaintext excerpt
	CreatedAt   time.Time       `json:"created_at"`
}

// Handler implements transport.FamilyHandler for the Notify family.
type Handler struct {
	mu            sync.RWMutex
	notifications map[string][]*Notification // recipientID → queue
	rateTrack     map[string][]time.Time      // senderID → recent push times
}

var _ transport.FamilyHandler = (*Handler)(nil)

func (h *Handler) ProtocolID() libp2pprotocol.ID { return libp2pprotocol.ID(protocolID) }

func (h *Handler) HandleStream(ctx context.Context, req *proto.PeerStreamRequest) *proto.PeerStreamResponse {
	switch req.Operation {
	case "notify.push":
		return h.handlePush(req)
	case "notify.drain":
		return h.handleDrain(req)
	default:
		return errorResp(req.RequestID, proto.CodeUnsupportedOperation, "unknown: "+req.Operation, nil)
	}
}

type pushPayload struct {
	RecipientID string          `json:"recipient_id"`
	SenderID    string          `json:"sender_id,omitempty"`
	Type        string          `json:"type"`
	Payload     json.RawMessage `json:"payload"`
	BodyPreview string          `json:"body_preview,omitempty"`
}

func (h *Handler) handlePush(req *proto.PeerStreamRequest) *proto.PeerStreamResponse {
	var p pushPayload
	if err := json.Unmarshal(req.Payload, &p); err != nil {
		return errorResp(req.RequestID, proto.CodeOperationFailed, "invalid payload", nil)
	}
	// Rate limiting: 60 pushes per sender per minute.
	if p.SenderID != "" {
		h.mu.Lock()
		if h.rateTrack == nil {
			h.rateTrack = make(map[string][]time.Time)
		}
		now := time.Now()
		cutoff := now.Add(-notifyRateWindow)
		var recent []time.Time
		for _, t := range h.rateTrack[p.SenderID] {
			if t.After(cutoff) {
				recent = append(recent, t)
			}
		}
		if len(recent) >= notifyRateLimit {
			h.mu.Unlock()
			return errorResp(req.RequestID, proto.CodeRateLimited, "notify rate limit exceeded (60/min)", nil)
		}
		h.rateTrack[p.SenderID] = append(recent, now)
		h.mu.Unlock()
	}
	n := &Notification{
		RecipientID: p.RecipientID,
		SenderID:    p.SenderID,
		Type:        p.Type,
		Payload:     p.Payload,
		BodyPreview: p.BodyPreview,
		CreatedAt:   time.Now(),
	}
	h.mu.Lock()
	if h.notifications == nil {
		h.notifications = make(map[string][]*Notification)
	}
	h.notifications[p.RecipientID] = append(h.notifications[p.RecipientID], n)
	h.mu.Unlock()
	resp, _ := json.Marshal(map[string]any{"ok": true})
	return &proto.PeerStreamResponse{NegotiatedProtocol: protocolID, Payload: resp}
}

type drainPayload struct {
	RecipientID string `json:"recipient_id"`
}

func (h *Handler) handleDrain(req *proto.PeerStreamRequest) *proto.PeerStreamResponse {
	var p drainPayload
	if err := json.Unmarshal(req.Payload, &p); err != nil {
		return errorResp(req.RequestID, proto.CodeOperationFailed, "invalid payload", nil)
	}
	h.mu.Lock()
	ns := h.notifications[p.RecipientID]
	delete(h.notifications, p.RecipientID)
	h.mu.Unlock()
	resp, _ := json.Marshal(map[string]any{"notifications": ns, "count": len(ns)})
	return &proto.PeerStreamResponse{NegotiatedProtocol: protocolID, Payload: resp}
}

// Push queues a notification (direct access).
func (h *Handler) Push(recipientID, notifType string, payload json.RawMessage) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.notifications == nil {
		h.notifications = make(map[string][]*Notification)
	}
	h.notifications[recipientID] = append(h.notifications[recipientID], &Notification{
		RecipientID: recipientID,
		Type:        notifType,
		Payload:     payload,
		CreatedAt:   time.Now(),
	})
}

// Drain returns and clears queued notifications for a recipient (direct access).
func (h *Handler) Drain(recipientID string) []*Notification {
	h.mu.Lock()
	defer h.mu.Unlock()
	ns := h.notifications[recipientID]
	delete(h.notifications, recipientID)
	return ns
}

func errorResp(requestID, code, msg string, missing []string) *proto.PeerStreamResponse {
	return &proto.PeerStreamResponse{RequestID: requestID, Error: proto.NewPeerStreamError(code, msg, missing)}
}
