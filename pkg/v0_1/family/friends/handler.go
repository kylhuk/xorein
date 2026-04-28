// Package friends implements the /aether/friends/0.2.0 protocol family.
// Source: docs/spec/v0.1/47-family-friends.md
package friends

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	libp2pprotocol "github.com/libp2p/go-libp2p/core/protocol"

	proto "github.com/aether/code_aether/pkg/v0_1/protocol"
	"github.com/aether/code_aether/pkg/v0_1/transport"
)

const protocolID = "/aether/friends/0.2.0"

const (
	maxRequestRatePerHour = 10
	requestExpiryDays     = 7
)

var ErrRateLimited = errors.New("friends: request rate limit exceeded (10/hour)")

// FriendRequest represents a pending or accepted friend request.
type FriendRequest struct {
	FromPeerID string    `json:"from_peer_id"`
	ToPeerID   string    `json:"to_peer_id"`
	Status     string    `json:"status"` // "pending", "accepted", "rejected"
	CreatedAt  time.Time `json:"created_at"`
	ExpiresAt  time.Time `json:"expires_at"`
}

// Handler implements transport.FamilyHandler for the Friends family.
type Handler struct {
	mu          sync.RWMutex
	LocalPeerID string
	requests    map[string]*FriendRequest // fromID+toID → request
	friends     map[string][]string       // peerID → list of friend peer IDs
	rateTrack   map[string][]time.Time    // peerID → recent request times
	blocked     map[string]struct{}       // "fromID:toID" → blocked
}

var _ transport.FamilyHandler = (*Handler)(nil)

func (h *Handler) ProtocolID() libp2pprotocol.ID { return libp2pprotocol.ID(protocolID) }

func (h *Handler) HandleStream(ctx context.Context, req *proto.PeerStreamRequest) *proto.PeerStreamResponse {
	switch req.Operation {
	case "friends.request":
		return h.handleRequest(req)
	case "friends.accept":
		return h.handleAccept(req)
	case "friends.remove":
		return h.handleRemove(req)
	case "friends.block":
		return h.handleBlock(req)
	case "friends.list":
		return h.handleList(req)
	default:
		return errorResp(req.RequestID, proto.CodeUnsupportedOperation, "unknown: "+req.Operation, nil)
	}
}

type requestPayload struct {
	FromPeerID string `json:"from_peer_id"`
	ToPeerID   string `json:"to_peer_id"`
}

func (h *Handler) handleRequest(req *proto.PeerStreamRequest) *proto.PeerStreamResponse {
	var p requestPayload
	if err := json.Unmarshal(req.Payload, &p); err != nil {
		return errorResp(req.RequestID, proto.CodeOperationFailed, "invalid payload", nil)
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.rateTrack == nil {
		h.rateTrack = make(map[string][]time.Time)
	}
	// Rate limit: 10 requests per hour.
	now := time.Now()
	cutoff := now.Add(-time.Hour)
	var recent []time.Time
	for _, t := range h.rateTrack[p.FromPeerID] {
		if t.After(cutoff) {
			recent = append(recent, t)
		}
	}
	if len(recent) >= maxRequestRatePerHour {
		return errorResp(req.RequestID, proto.CodeRateLimited, ErrRateLimited.Error(), nil)
	}
	h.rateTrack[p.FromPeerID] = append(recent, now)

	key := p.FromPeerID + ":" + p.ToPeerID
	if h.requests == nil {
		h.requests = make(map[string]*FriendRequest)
	}
	// Duplicate request detection: if a pending request already exists from this peer, return it.
	if existing := h.requests[key]; existing != nil && existing.Status == "pending" && now.Before(existing.ExpiresAt) {
		resp, _ := json.Marshal(map[string]any{"ok": true, "status": "pending", "duplicate": true})
		return &proto.PeerStreamResponse{NegotiatedProtocol: protocolID, Payload: resp}
	}
	h.requests[key] = &FriendRequest{
		FromPeerID: p.FromPeerID,
		ToPeerID:   p.ToPeerID,
		Status:     "pending",
		CreatedAt:  now,
		ExpiresAt:  now.Add(requestExpiryDays * 24 * time.Hour),
	}
	resp, _ := json.Marshal(map[string]any{"ok": true, "status": "pending"})
	return &proto.PeerStreamResponse{NegotiatedProtocol: protocolID, Payload: resp}
}

type acceptPayload struct {
	FromPeerID string `json:"from_peer_id"`
	ToPeerID   string `json:"to_peer_id"`
}

func (h *Handler) handleAccept(req *proto.PeerStreamRequest) *proto.PeerStreamResponse {
	var p acceptPayload
	if err := json.Unmarshal(req.Payload, &p); err != nil {
		return errorResp(req.RequestID, proto.CodeOperationFailed, "invalid payload", nil)
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	key := p.FromPeerID + ":" + p.ToPeerID
	r := h.requests[key]
	if r == nil || r.Status != "pending" {
		return errorResp(req.RequestID, proto.CodeOperationFailed, "no pending request", nil)
	}
	if time.Now().After(r.ExpiresAt) {
		return errorResp(req.RequestID, proto.CodeOperationFailed, "request expired", nil)
	}
	r.Status = "accepted"
	if h.friends == nil {
		h.friends = make(map[string][]string)
	}
	h.friends[p.FromPeerID] = append(h.friends[p.FromPeerID], p.ToPeerID)
	h.friends[p.ToPeerID] = append(h.friends[p.ToPeerID], p.FromPeerID)
	resp, _ := json.Marshal(map[string]any{"ok": true, "status": "accepted"})
	return &proto.PeerStreamResponse{NegotiatedProtocol: protocolID, Payload: resp}
}

type listPayload struct {
	PeerID string `json:"peer_id"`
}

func (h *Handler) handleList(req *proto.PeerStreamRequest) *proto.PeerStreamResponse {
	var p listPayload
	if err := json.Unmarshal(req.Payload, &p); err != nil {
		return errorResp(req.RequestID, proto.CodeOperationFailed, "invalid payload", nil)
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	friends := h.friends[p.PeerID]
	resp, _ := json.Marshal(map[string]any{"peer_id": p.PeerID, "friends": friends})
	return &proto.PeerStreamResponse{NegotiatedProtocol: protocolID, Payload: resp}
}

type removePayload struct {
	PeerID       string `json:"peer_id"`
	FriendPeerID string `json:"friend_peer_id"`
}

func (h *Handler) handleRemove(req *proto.PeerStreamRequest) *proto.PeerStreamResponse {
	var p removePayload
	if err := json.Unmarshal(req.Payload, &p); err != nil {
		return errorResp(req.RequestID, proto.CodeOperationFailed, "invalid payload", nil)
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.friends == nil {
		h.friends = make(map[string][]string)
	}
	// Remove from both sides.
	h.friends[p.PeerID] = removePeer(h.friends[p.PeerID], p.FriendPeerID)
	h.friends[p.FriendPeerID] = removePeer(h.friends[p.FriendPeerID], p.PeerID)
	resp, _ := json.Marshal(map[string]any{"ok": true})
	return &proto.PeerStreamResponse{NegotiatedProtocol: protocolID, Payload: resp}
}

type blockPayload struct {
	ActorPeerID  string `json:"actor_peer_id"`
	TargetPeerID string `json:"target_peer_id"`
}

func (h *Handler) handleBlock(req *proto.PeerStreamRequest) *proto.PeerStreamResponse {
	var p blockPayload
	if err := json.Unmarshal(req.Payload, &p); err != nil {
		return errorResp(req.RequestID, proto.CodeOperationFailed, "invalid payload", nil)
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.blocked == nil {
		h.blocked = make(map[string]struct{})
	}
	key := p.ActorPeerID + ":" + p.TargetPeerID
	h.blocked[key] = struct{}{}
	// Also remove from friends.
	if h.friends != nil {
		h.friends[p.ActorPeerID] = removePeer(h.friends[p.ActorPeerID], p.TargetPeerID)
		h.friends[p.TargetPeerID] = removePeer(h.friends[p.TargetPeerID], p.ActorPeerID)
	}
	resp, _ := json.Marshal(map[string]any{"ok": true})
	return &proto.PeerStreamResponse{NegotiatedProtocol: protocolID, Payload: resp}
}

// removePeer removes targetID from a peer list, returning the updated slice.
func removePeer(list []string, targetID string) []string {
	out := list[:0]
	for _, id := range list {
		if id != targetID {
			out = append(out, id)
		}
	}
	return out
}

// ExpireRequestForTest backdates a pending request's ExpiresAt for test purposes.
func (h *Handler) ExpireRequestForTest(fromID, toID string, expiredAt time.Time) {
	h.mu.Lock()
	defer h.mu.Unlock()
	key := fromID + ":" + toID
	if r := h.requests[key]; r != nil {
		r.ExpiresAt = expiredAt
	}
}

func errorResp(requestID, code, msg string, missing []string) *proto.PeerStreamResponse {
	return &proto.PeerStreamResponse{RequestID: requestID, Error: proto.NewPeerStreamError(code, msg, missing)}
}
