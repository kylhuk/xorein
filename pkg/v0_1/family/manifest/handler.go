// Package manifest implements the /aether/manifest/0.1.0 protocol family.
// Source: docs/spec/v0.1/42-family-manifest.md
package manifest

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	libp2pprotocol "github.com/libp2p/go-libp2p/core/protocol"

	"github.com/aether/code_aether/pkg/v0_1/envelope"
	proto "github.com/aether/code_aether/pkg/v0_1/protocol"
	"github.com/aether/code_aether/pkg/v0_1/transport"
)

const protocolID = "/aether/manifest/0.1.0"

// ManifestRecord is a published manifest entry.
type ManifestRecord struct {
	ID          string          `json:"id"`
	ScopeID     string          `json:"scope_id"`
	PublisherID string          `json:"publisher_id"`
	Hash        string          `json:"hash"` // SHA-256 of Content
	Content     json.RawMessage `json:"content"`
	PublishedAt time.Time       `json:"published_at"`
}

// Handler implements transport.FamilyHandler for the Manifest family.
type Handler struct {
	mu        sync.RWMutex
	manifests map[string]*ManifestRecord // hash → manifest
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
	case "manifest.publish":
		return h.handlePublish(req)
	case "manifest.fetch":
		return h.handleFetch(req)
	default:
		return errorResp(req.RequestID, proto.CodeUnsupportedOperation, "not implemented: "+req.Operation, nil)
	}
}

func (h *Handler) localCaps() []proto.FeatureFlag {
	return []proto.FeatureFlag{"cap.manifest"}
}

type publishPayload struct {
	ScopeID     string          `json:"scope_id"`
	PublisherID string          `json:"publisher_id"`
	Content     json.RawMessage `json:"content"`
}

// defaultManifestSecurityModes are the defaults per spec 04 §4 and spec 42 §2.
var defaultOfferedSecurityModes = []string{"tree", "crowd", "channel", "seal", "clear"}

func (h *Handler) handlePublish(req *proto.PeerStreamRequest) *proto.PeerStreamResponse {
	var p publishPayload
	if err := json.Unmarshal(req.Payload, &p); err != nil {
		return errorResp(req.RequestID, proto.CodeOperationFailed, "invalid payload", nil)
	}

	// Inject spec defaults for security_mode and offered_security_modes if absent (spec 04 §3.2).
	var contentMap map[string]any
	if err := json.Unmarshal(p.Content, &contentMap); err == nil {
		changed := false
		if _, ok := contentMap["security_mode"]; !ok {
			contentMap["security_mode"] = "tree"
			changed = true
		}
		if _, ok := contentMap["offered_security_modes"]; !ok {
			contentMap["offered_security_modes"] = defaultOfferedSecurityModes
			changed = true
		}
		if changed {
			if b, err := json.Marshal(contentMap); err == nil {
				p.Content = json.RawMessage(b)
			}
		}
	}

	// Compute hash using spec 02 §4: base64url-no-pad 32-char prefix of SHA-256.
	hash := envelope.ManifestHash(p.Content)

	h.mu.Lock()
	defer h.mu.Unlock()
	if h.manifests == nil {
		h.manifests = make(map[string]*ManifestRecord)
	}

	// Version check: if an existing manifest exists for this scope and publisher,
	// reject if the new content's updated_at is not strictly newer.
	if existing := h.manifestForScope(p.ScopeID); existing != nil {
		var newContent, existContent map[string]any
		if json.Unmarshal(p.Content, &newContent) == nil &&
			json.Unmarshal(existing.Content, &existContent) == nil {
			newUA, hasNew := newContent["updated_at"]
			existUA, hasExist := existContent["updated_at"]
			if hasNew && hasExist {
				newUAStr, _ := newUA.(string)
				existUAStr, _ := existUA.(string)
				if newUAStr != "" && existUAStr != "" && newUAStr <= existUAStr {
					return errorResp(req.RequestID, proto.CodeManifestNotNewer, "manifest is not newer than stored version", nil)
				}
			}
		}
	}

	record := &ManifestRecord{
		ID:          fmt.Sprintf("%s/%s", p.ScopeID, hash[:8]),
		ScopeID:     p.ScopeID,
		PublisherID: p.PublisherID,
		Hash:        hash,
		Content:     p.Content,
		PublishedAt: time.Now(),
	}
	h.manifests[hash] = record
	resp, _ := json.Marshal(map[string]any{"id": record.ID, "hash": hash})
	return &proto.PeerStreamResponse{NegotiatedProtocol: protocolID, Payload: resp}
}

// manifestForScope returns the most recently published manifest for a scope, or nil.
// Caller must hold h.mu (read or write).
func (h *Handler) manifestForScope(scopeID string) *ManifestRecord {
	var latest *ManifestRecord
	for _, m := range h.manifests {
		if m.ScopeID == scopeID {
			if latest == nil || m.PublishedAt.After(latest.PublishedAt) {
				latest = m
			}
		}
	}
	return latest
}

type fetchPayload struct {
	Hash    string `json:"hash,omitempty"`
	ScopeID string `json:"scope_id,omitempty"`
}

func (h *Handler) handleFetch(req *proto.PeerStreamRequest) *proto.PeerStreamResponse {
	var p fetchPayload
	if err := json.Unmarshal(req.Payload, &p); err != nil {
		return errorResp(req.RequestID, proto.CodeOperationFailed, "invalid payload", nil)
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	if p.Hash != "" {
		m := h.manifests[p.Hash]
		if m == nil {
			return errorResp(req.RequestID, proto.CodeManifestNotFound, "manifest not found: "+p.Hash, nil)
		}
		resp, _ := json.Marshal(m)
		return &proto.PeerStreamResponse{NegotiatedProtocol: protocolID, Payload: resp}
	}
	// Return all manifests for scope.
	var results []*ManifestRecord
	for _, m := range h.manifests {
		if p.ScopeID == "" || m.ScopeID == p.ScopeID {
			results = append(results, m)
		}
	}
	resp, _ := json.Marshal(map[string]any{"manifests": results})
	return &proto.PeerStreamResponse{NegotiatedProtocol: protocolID, Payload: resp}
}

// Publish stores a manifest locally (for direct access without wire protocol).
func (h *Handler) Publish(scopeID, publisherID string, content json.RawMessage) (*ManifestRecord, error) {
	req := &proto.PeerStreamRequest{
		Operation: "manifest.publish",
	}
	req.Payload, _ = json.Marshal(publishPayload{ScopeID: scopeID, PublisherID: publisherID, Content: content})
	resp := h.HandleStream(context.Background(), req)
	if resp.Error != nil {
		return nil, fmt.Errorf("manifest publish: %s", resp.Error.Message)
	}
	var result struct{ Hash string `json:"hash"` }
	json.Unmarshal(resp.Payload, &result)
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.manifests[result.Hash], nil
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
