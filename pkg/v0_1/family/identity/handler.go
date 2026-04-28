// Package identity implements the /aether/identity/0.1.0 protocol family.
// Source: docs/spec/v0.1/43-family-identity.md
package identity

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	libp2pprotocol "github.com/libp2p/go-libp2p/core/protocol"

	"github.com/aether/code_aether/pkg/v0_1/family/identity/store"
	seal "github.com/aether/code_aether/pkg/v0_1/mode/seal"
	proto "github.com/aether/code_aether/pkg/v0_1/protocol"
	"github.com/aether/code_aether/pkg/v0_1/transport"
)

const protocolID = "/aether/identity/0.1.0"

// Handler implements transport.FamilyHandler for the identity family.
type Handler struct {
	mu       sync.RWMutex
	bundles  map[string]*seal.PrekeyBundle // peer_id → bundle
	BundleStore store.BundleStore
}

var _ transport.FamilyHandler = (*Handler)(nil)

func (h *Handler) ProtocolID() libp2pprotocol.ID {
	return libp2pprotocol.ID(protocolID)
}

func (h *Handler) HandleStream(ctx context.Context, req *proto.PeerStreamRequest) *proto.PeerStreamResponse {
	required, known := proto.RequiredCapsFor(req.Operation)
	if !known {
		return errorResp(req.RequestID, proto.CodeUnsupportedOperation, "unknown operation: "+req.Operation, nil)
	}
	remote := toFeatureFlags(req.AdvertisedCaps)
	result := proto.NegotiateCapabilities(h.localCaps(), remote, required)
	if len(result.MissingRequired) > 0 {
		return errorResp(req.RequestID, proto.CodeMissingRequiredCapability, "missing capabilities", result.MissingRequired)
	}

	switch req.Operation {
	case "identity.publish":
		return h.handlePublish(req)
	case "identity.fetch":
		return h.handleFetch(req)
	default:
		return errorResp(req.RequestID, proto.CodeUnsupportedOperation, "operation not implemented: "+req.Operation, nil)
	}
}

func (h *Handler) localCaps() []proto.FeatureFlag {
	return []proto.FeatureFlag{"cap.identity"}
}

// publishRequest is the wire format for identity.publish.
type publishRequest struct {
	Bundle json.RawMessage `json:"bundle"`
}

// fetchRequest is the wire format for identity.fetch.
type fetchRequest struct {
	PeerID      string `json:"peer_id"`
	OPKConsumed bool   `json:"opk_consumed,omitempty"`
}

// fetchResponse is the wire format for identity.fetch response.
type fetchResponse struct {
	Bundle        *seal.PrekeyBundle `json:"bundle"`
	RemainingOPKs int                `json:"remaining_opks"`
}

func (h *Handler) handlePublish(req *proto.PeerStreamRequest) *proto.PeerStreamResponse {
	var pr publishRequest
	if err := json.Unmarshal(req.Payload, &pr); err != nil {
		return errorResp(req.RequestID, proto.CodeOperationFailed, "invalid publish payload: "+err.Error(), nil)
	}
	var bundle seal.PrekeyBundle
	if err := json.Unmarshal(pr.Bundle, &bundle); err != nil {
		return errorResp(req.RequestID, proto.CodeOperationFailed, "invalid bundle JSON: "+err.Error(), nil)
	}
	if err := seal.VerifyBundle(&bundle, time.Now()); err != nil {
		return errorResp(req.RequestID, proto.CodeSignatureMismatch, "bundle verification failed: "+err.Error(), nil)
	}

	h.storeBundleLocal(bundle.PeerID, &bundle)
	if h.BundleStore != nil {
		_ = h.BundleStore.Put(bundle.PeerID, &bundle)
	}

	payload, _ := json.Marshal(map[string]string{"status": "ok"})
	return &proto.PeerStreamResponse{
		NegotiatedProtocol: protocolID,
		Payload:            payload,
	}
}

func (h *Handler) handleFetch(req *proto.PeerStreamRequest) *proto.PeerStreamResponse {
	var fr fetchRequest
	if err := json.Unmarshal(req.Payload, &fr); err != nil {
		return errorResp(req.RequestID, proto.CodeOperationFailed, "invalid fetch payload: "+err.Error(), nil)
	}

	bundle := h.getBundleLocal(fr.PeerID)
	if bundle == nil && h.BundleStore != nil {
		bundle, _ = h.BundleStore.Get(fr.PeerID)
	}
	if bundle == nil {
		return errorResp(req.RequestID, proto.CodeOperationFailed, "BUNDLE_NOT_FOUND", nil)
	}
	if time.Now().UnixMilli() > bundle.ExpiresAt {
		return errorResp(req.RequestID, proto.CodeOperationFailed, "BUNDLE_EXPIRED", nil)
	}

	remaining := len(bundle.OneTimePrekeys)
	if fr.OPKConsumed {
		if remaining == 0 {
			// All one-time prekeys exhausted; spec 43 requires NO_OPK_AVAILABLE.
			return errorResp(req.RequestID, proto.CodeNoOPKAvailable, "no one-time prekey available", nil)
		}
		// Pop the first OPK (consumed by the caller who just fetched it).
		h.consumeOPK(fr.PeerID, bundle)
		remaining--
	}
	// Sanity cap: reject bundles carrying more than 100 OPKs to limit payload size.
	if remaining > 100 {
		return errorResp(req.RequestID, proto.CodeOperationFailed, "too many one-time prekeys (max 100)", nil)
	}

	resp := fetchResponse{Bundle: bundle, RemainingOPKs: remaining}
	payload, _ := json.Marshal(resp)
	return &proto.PeerStreamResponse{
		NegotiatedProtocol: protocolID,
		Payload:            payload,
	}
}

func (h *Handler) storeBundleLocal(peerID string, b *seal.PrekeyBundle) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.bundles == nil {
		h.bundles = make(map[string]*seal.PrekeyBundle)
	}
	h.bundles[peerID] = b
}

func (h *Handler) getBundleLocal(peerID string) *seal.PrekeyBundle {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.bundles[peerID]
}

func (h *Handler) consumeOPK(peerID string, bundle *seal.PrekeyBundle) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if len(bundle.OneTimePrekeys) > 0 {
		bundle.OneTimePrekeys = bundle.OneTimePrekeys[1:]
	}
	if h.bundles == nil {
		h.bundles = make(map[string]*seal.PrekeyBundle)
	}
	h.bundles[peerID] = bundle
}

// PublishLocal stores a bundle locally without network exchange (for self-publication).
func (h *Handler) PublishLocal(b *seal.PrekeyBundle) {
	h.storeBundleLocal(b.PeerID, b)
	if h.BundleStore != nil {
		_ = h.BundleStore.Put(b.PeerID, b)
	}
}

// FetchLocal returns the bundle for peerID from the local cache.
func (h *Handler) FetchLocal(peerID string) *seal.PrekeyBundle {
	b := h.getBundleLocal(peerID)
	if b == nil && h.BundleStore != nil {
		b, _ = h.BundleStore.Get(peerID)
	}
	return b
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
