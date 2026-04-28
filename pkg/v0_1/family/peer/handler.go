// Package peer implements the /aether/peer/0.1.0 protocol family.
// Source: docs/spec/v0.1/40-family-peer.md
package peer

import (
	"context"
	"encoding/json"
	"time"

	libp2pprotocol "github.com/libp2p/go-libp2p/core/protocol"

	apb "github.com/aether/code_aether/gen/go/proto"
	proto "github.com/aether/code_aether/pkg/v0_1/protocol"
	"github.com/aether/code_aether/pkg/v0_1/transport"
	"github.com/aether/code_aether/pkg/v0_1/nat"
)

const protocolID = "/aether/peer/0.1.0"

// Handler implements transport.FamilyHandler for the peer family.
type Handler struct {
	// LocalIdentity is the node's own IdentityProfile (for peer.info response).
	LocalIdentity *apb.IdentityProfile
	// LocalCaps is the list of capability flags this node advertises.
	LocalCaps []proto.FeatureFlag
	// RelayQueue is the store-and-forward queue used by peer.relay.store and
	// peer.relay.drain (spec 32 §4, charter §3.4). If nil, those operations
	// return CodeUnsupportedOperation.
	RelayQueue *nat.RelayQueue
}

var _ transport.FamilyHandler = (*Handler)(nil)

func (h *Handler) ProtocolID() libp2pprotocol.ID {
	return libp2pprotocol.ID(protocolID)
}

func (h *Handler) HandleStream(ctx context.Context, req *proto.PeerStreamRequest) *proto.PeerStreamResponse {
	// Capability negotiation.
	required, known := proto.RequiredCapsFor(req.Operation)
	if !known {
		return errorResp(req.RequestID, proto.CodeUnsupportedOperation, "unknown operation: "+req.Operation, nil)
	}
	remote := toFeatureFlags(req.AdvertisedCaps)
	result := proto.NegotiateCapabilities(h.LocalCaps, remote, required)
	if len(result.MissingRequired) > 0 {
		return errorResp(req.RequestID, proto.CodeMissingRequiredCapability, "missing capabilities", result.MissingRequired)
	}

	switch req.Operation {
	case "peer.info":
		return h.handlePeerInfo(req, result)
	case "peer.exchange":
		return h.handlePeerExchange(req, result)
	case "peer.relay.store":
		return h.handleRelayStore(req, result)
	case "peer.relay.drain":
		return h.handleRelayDrain(req, result)
	default:
		return errorResp(req.RequestID, proto.CodeUnsupportedOperation, "operation not implemented in this wave: "+req.Operation, nil)
	}
}

// PeerInfoResponse is the JSON payload for peer.info.
type PeerInfoResponse struct {
	IdentityID   string   `json:"identity_id"`
	DisplayName  string   `json:"display_name,omitempty"`
	Capabilities []string `json:"capabilities"`
	ProtocolID   string   `json:"protocol_id"`
	ServerTime   int64    `json:"server_time_ms"`
}

func (h *Handler) handlePeerInfo(req *proto.PeerStreamRequest, result proto.CapabilityNegResult) *proto.PeerStreamResponse {
	info := PeerInfoResponse{
		ProtocolID:   protocolID,
		ServerTime:   time.Now().UnixMilli(),
		Capabilities: toStringSlice(h.LocalCaps),
	}
	if h.LocalIdentity != nil {
		info.IdentityID = h.LocalIdentity.IdentityId
		info.DisplayName = h.LocalIdentity.DisplayName
	}
	payload, _ := json.Marshal(info)
	return &proto.PeerStreamResponse{
		NegotiatedProtocol: protocolID,
		AcceptedCaps:       toStringSlice(result.Accepted),
		IgnoredCaps:        result.IgnoredRemote,
		Payload:            payload,
	}
}

// PeerExchangeRequest is the JSON payload for peer.exchange.
type PeerExchangeRequest struct {
	KnownPeerIDs []string `json:"known_peer_ids,omitempty"`
}

// PeerRecord is a single peer advertisement in a peer.exchange response.
type PeerRecord struct {
	PeerID   string   `json:"peer_id"`
	Addrs    []string `json:"addrs,omitempty"`
	Role     string   `json:"role,omitempty"`
	LastSeen int64    `json:"last_seen_ms,omitempty"`
}

func (h *Handler) handlePeerExchange(_ *proto.PeerStreamRequest, result proto.CapabilityNegResult) *proto.PeerStreamResponse {
	// TODO(W5): return real DHT/mDNS-discovered peers.
	payload, _ := json.Marshal([]PeerRecord{})
	return &proto.PeerStreamResponse{
		NegotiatedProtocol: protocolID,
		AcceptedCaps:       toStringSlice(result.Accepted),
		IgnoredCaps:        result.IgnoredRemote,
		Payload:            payload,
	}
}

// relayStoreRequest is the JSON payload for peer.relay.store.
type relayStoreRequest struct {
	RecipientID string `json:"recipient_id"`
	ID          string `json:"id"`
	Body        string `json:"body"` // base64url-encoded ciphertext
}

// relayStoreResponse is the JSON payload returned on success.
type relayStoreResponse struct {
	Queued bool `json:"queued"`
}

func (h *Handler) handleRelayStore(req *proto.PeerStreamRequest, result proto.CapabilityNegResult) *proto.PeerStreamResponse {
	if h.RelayQueue == nil {
		return errorResp(req.RequestID, proto.CodeUnsupportedOperation, "relay queue not available on this node", nil)
	}

	var r relayStoreRequest
	if err := json.Unmarshal(req.Payload, &r); err != nil {
		return errorResp(req.RequestID, proto.CodeOperationFailed, "invalid relay.store payload: "+err.Error(), nil)
	}
	if r.RecipientID == "" || r.ID == "" {
		return errorResp(req.RequestID, proto.CodeOperationFailed, "recipient_id and id are required", nil)
	}

	if err := h.RelayQueue.Store(r.RecipientID, r.ID, []byte(r.Body)); err != nil {
		code := proto.CodeOperationFailed
		if err == nat.ErrRelayOpacityViolation {
			code = proto.CodeRelayOpacityViolation
		}
		return errorResp(req.RequestID, code, err.Error(), nil)
	}

	payload, _ := json.Marshal(relayStoreResponse{Queued: true})
	return &proto.PeerStreamResponse{
		NegotiatedProtocol: protocolID,
		AcceptedCaps:       toStringSlice(result.Accepted),
		IgnoredCaps:        result.IgnoredRemote,
		Payload:            payload,
	}
}

// relayDrainRequest is the JSON payload for peer.relay.drain.
type relayDrainRequest struct {
	RecipientID string `json:"recipient_id"`
}

// relayDrainEntry is a single entry in the drain response.
type relayDrainEntry struct {
	ID        string `json:"id"`
	Body      string `json:"body"`
	StoredAt  int64  `json:"stored_at_ms"`
	ExpiresAt int64  `json:"expires_at_ms"`
}

// relayDrainResponse is the JSON payload returned by peer.relay.drain.
type relayDrainResponse struct {
	Entries []relayDrainEntry `json:"entries"`
}

func (h *Handler) handleRelayDrain(req *proto.PeerStreamRequest, result proto.CapabilityNegResult) *proto.PeerStreamResponse {
	if h.RelayQueue == nil {
		return errorResp(req.RequestID, proto.CodeUnsupportedOperation, "relay queue not available on this node", nil)
	}

	var r relayDrainRequest
	if err := json.Unmarshal(req.Payload, &r); err != nil {
		return errorResp(req.RequestID, proto.CodeOperationFailed, "invalid relay.drain payload: "+err.Error(), nil)
	}
	if r.RecipientID == "" {
		return errorResp(req.RequestID, proto.CodeOperationFailed, "recipient_id is required", nil)
	}

	drained := h.RelayQueue.Drain(r.RecipientID)
	entries := make([]relayDrainEntry, 0, len(drained))
	for _, e := range drained {
		entries = append(entries, relayDrainEntry{
			ID:        e.ID,
			Body:      string(e.Body),
			StoredAt:  e.StoredAt.UnixMilli(),
			ExpiresAt: e.ExpiresAt.UnixMilli(),
		})
	}

	payload, _ := json.Marshal(relayDrainResponse{Entries: entries})
	return &proto.PeerStreamResponse{
		NegotiatedProtocol: protocolID,
		AcceptedCaps:       toStringSlice(result.Accepted),
		IgnoredCaps:        result.IgnoredRemote,
		Payload:            payload,
	}
}

func errorResp(requestID, code, msg string, missing []string) *proto.PeerStreamResponse {
	return &proto.PeerStreamResponse{
		RequestID: requestID,
		Error:     proto.NewPeerStreamError(code, msg, missing),
	}
}

func toFeatureFlags(ss []string) []proto.FeatureFlag {
	out := make([]proto.FeatureFlag, len(ss))
	for i, s := range ss {
		out[i] = proto.FeatureFlag(s)
	}
	return out
}

func toStringSlice(ff []proto.FeatureFlag) []string {
	out := make([]string, len(ff))
	for i, f := range ff {
		out[i] = string(f)
	}
	return out
}
