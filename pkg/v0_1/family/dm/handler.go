// Package dm implements the /aether/dm/0.2.0 protocol family (Seal-mode direct messages).
// Source: docs/spec/v0.1/45-family-dm.md
package dm

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	libp2pprotocol "github.com/libp2p/go-libp2p/core/protocol"

	v0crypto "github.com/aether/code_aether/pkg/v0_1/crypto"
	idstore "github.com/aether/code_aether/pkg/v0_1/family/identity/store"
	clearmode "github.com/aether/code_aether/pkg/v0_1/mode/clear"
	seal "github.com/aether/code_aether/pkg/v0_1/mode/seal"
	proto "github.com/aether/code_aether/pkg/v0_1/protocol"
	"github.com/aether/code_aether/pkg/v0_1/transport"
)

const protocolID = "/aether/dm/0.2.0"

// MessageRecord holds a decrypted DM message.
type MessageRecord struct {
	ID           string
	ScopeID      string
	SenderPeerID string
	Plaintext    []byte
	ReceivedAt   time.Time
}

// dmRateWindow is the sliding window duration for DM rate limiting.
const dmRateWindow = time.Minute

// dmRateLimit is the maximum number of DM messages per sender per minute.
const dmRateLimit = 60

// Handler implements transport.FamilyHandler for the DM family.
// OwnPrivate must be set for decryption (responder path).
type Handler struct {
	mu           sync.RWMutex
	LocalPeerID  string
	LocalEdPriv  ed25519.PrivateKey
	OwnBundle    *seal.PrekeyBundle  // our published bundle
	OwnPrivate   *seal.PrekeyPrivate // our private key material
	BundleStore  idstore.BundleStore // peer bundles for lookup
	sessions     map[string]*seal.RatchetState
	scopeModes   map[string]clearmode.SecurityMode
	messages     []MessageRecord
	rateTrack    map[string][]time.Time // senderPeerID → recent send times
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
	case "dm.send":
		return h.handleSend(req)
	default:
		return errorResp(req.RequestID, proto.CodeUnsupportedOperation, "not implemented: "+req.Operation, nil)
	}
}

func (h *Handler) localCaps() []proto.FeatureFlag {
	return []proto.FeatureFlag{"cap.dm", "mode.seal"}
}

// Delivery is the wire-level DM delivery per spec 45 §4.2.
type Delivery struct {
	ID               string          `json:"id"`
	Kind             string          `json:"kind"`
	ScopeID          string          `json:"scope_id"`
	ScopeType        string          `json:"scope_type"`
	SenderPeerID     string          `json:"sender_peer_id"`
	RecipientPeerIDs []string        `json:"recipient_peer_ids"`
	Body             string          `json:"body"`
	Data             json.RawMessage `json:"data,omitempty"`
	CreatedAt        string          `json:"created_at"`
	Signature        string          `json:"signature,omitempty"`
}

// DeliveryData carries per-message header + optional X3DH fields.
type DeliveryData struct {
	Header           string `json:"data"`
	CiphertextFormat string `json:"ciphertext_format"`
	EKPub            string `json:"ek_pub,omitempty"`
	CTMLKEM          string `json:"ct_mlkem,omitempty"`
	OPKIndex         int    `json:"opk_index,omitempty"`
}

func (h *Handler) handleSend(req *proto.PeerStreamRequest) *proto.PeerStreamResponse {
	var payload struct {
		Delivery json.RawMessage `json:"delivery"`
	}
	if err := json.Unmarshal(req.Payload, &payload); err != nil {
		return errorResp(req.RequestID, proto.CodeOperationFailed, "invalid payload", nil)
	}
	var d Delivery
	if err := json.Unmarshal(payload.Delivery, &d); err != nil {
		return errorResp(req.RequestID, proto.CodeOperationFailed, "invalid delivery", nil)
	}

	// Timestamp validity check (spec 45): reject deliveries older than 24h or more
	// than 5 min in the future to prevent replay and clock-skew attacks.
	if d.CreatedAt != "" {
		if createdAt, err := time.Parse(time.RFC3339Nano, d.CreatedAt); err == nil {
			now := time.Now()
			if createdAt.After(now.Add(5*time.Minute)) || createdAt.Before(now.Add(-24*time.Hour)) {
				return errorResp(req.RequestID, proto.CodeExpiredSignature, "delivery timestamp out of acceptable range", nil)
			}
		}
	}

	// Rate limiting: 60 messages per minute per sender.
	if d.SenderPeerID != "" {
		h.mu.Lock()
		if h.rateTrack == nil {
			h.rateTrack = make(map[string][]time.Time)
		}
		now := time.Now()
		cutoff := now.Add(-dmRateWindow)
		var recent []time.Time
		for _, t := range h.rateTrack[d.SenderPeerID] {
			if t.After(cutoff) {
				recent = append(recent, t)
			}
		}
		if len(recent) >= dmRateLimit {
			h.mu.Unlock()
			return errorResp(req.RequestID, proto.CodeDMRateLimited, "DM rate limit exceeded (60/min)", nil)
		}
		h.rateTrack[d.SenderPeerID] = append(recent, now)
		h.mu.Unlock()
	}

	// Mode continuity.
	h.mu.RLock()
	existingMode := h.scopeModes[d.ScopeID]
	h.mu.RUnlock()
	if err := clearmode.EnforceModeContinuity(existingMode, clearmode.ModeSeal); err != nil {
		return errorResp(req.RequestID, proto.CodeModeIncompatible, err.Error(), nil)
	}

	ct, err := base64.RawURLEncoding.DecodeString(d.Body)
	if err != nil {
		return errorResp(req.RequestID, proto.CodeOperationFailed, "invalid body", nil)
	}

	var dd DeliveryData
	if len(d.Data) > 0 {
		if err := json.Unmarshal(d.Data, &dd); err != nil {
			return errorResp(req.RequestID, proto.CodeOperationFailed, "invalid data field", nil)
		}
	}

	hdrBytes, err := base64.RawURLEncoding.DecodeString(dd.Header)
	if err != nil || len(hdrBytes) != seal.HeaderSize {
		return errorResp(req.RequestID, proto.CodeOperationFailed, fmt.Sprintf("bad header len %d", len(hdrBytes)), nil)
	}
	var hdr [seal.HeaderSize]byte
	copy(hdr[:], hdrBytes)

	plaintext, rs, err := h.decrypt(d.ScopeID, d.SenderPeerID, hdr, ct, &dd)
	if err != nil {
		return errorResp(req.RequestID, proto.CodeOperationFailed, "decrypt: "+err.Error(), nil)
	}

	h.mu.Lock()
	if h.sessions == nil {
		h.sessions = make(map[string]*seal.RatchetState)
	}
	h.sessions[d.ScopeID] = rs
	if h.scopeModes == nil {
		h.scopeModes = make(map[string]clearmode.SecurityMode)
	}
	h.scopeModes[d.ScopeID] = clearmode.ModeSeal
	h.messages = append(h.messages, MessageRecord{
		ID: d.ID, ScopeID: d.ScopeID,
		SenderPeerID: d.SenderPeerID, Plaintext: plaintext,
		ReceivedAt: time.Now(),
	})
	h.mu.Unlock()

	resp, _ := json.Marshal(map[string]any{
		"delivery_id":         d.ID,
		"acknowledged_at_ms":  time.Now().UnixMilli(),
	})
	return &proto.PeerStreamResponse{NegotiatedProtocol: protocolID, Payload: resp}
}

func (h *Handler) decrypt(
	scopeID, senderPeerID string,
	hdr [seal.HeaderSize]byte,
	ct []byte,
	dd *DeliveryData,
) ([]byte, *seal.RatchetState, error) {
	h.mu.RLock()
	rs := h.sessions[scopeID]
	h.mu.RUnlock()

	if rs == nil {
		// First message — X3DH respond.
		if dd.EKPub == "" || dd.CTMLKEM == "" {
			return nil, nil, fmt.Errorf("first message missing X3DH fields")
		}
		ekPubB, err := base64.RawURLEncoding.DecodeString(dd.EKPub)
		if err != nil || len(ekPubB) != 32 {
			return nil, nil, fmt.Errorf("invalid ek_pub")
		}
		ctMLKEM, err := base64.RawURLEncoding.DecodeString(dd.CTMLKEM)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid ct_mlkem")
		}
		if h.OwnPrivate == nil || h.OwnBundle == nil {
			return nil, nil, fmt.Errorf("own bundle/private not configured")
		}
		senderBundle, err := h.BundleStore.Get(senderPeerID)
		if err != nil || senderBundle == nil {
			return nil, nil, fmt.Errorf("sender bundle not found: %v", err)
		}
		var ekPub [32]byte
		copy(ekPub[:], ekPubB)
		im := &seal.InitialMessage{EKPub: ekPub, CTMLKEM: ctMLKEM, OPKIndex: dd.OPKIndex}
		rs, err = seal.RespondFull(im, h.OwnPrivate, h.OwnBundle, h.LocalEdPriv, ed25519.PublicKey(senderBundle.IdentityKeyEd25519))
		if err != nil {
			return nil, nil, fmt.Errorf("X3DH respond: %w", err)
		}
	}

	pt, err := seal.Decrypt(rs, hdr, ct)
	if err != nil {
		return nil, nil, err
	}
	return pt, rs, nil
}

// SendMessage encrypts and builds a Delivery for the given peer.
func (h *Handler) SendMessage(scopeID string, peerBundle *seal.PrekeyBundle, plaintext []byte) (*Delivery, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.sessions == nil {
		h.sessions = make(map[string]*seal.RatchetState)
	}
	if h.scopeModes == nil {
		h.scopeModes = make(map[string]clearmode.SecurityMode)
	}

	if err := clearmode.EnforceModeContinuity(h.scopeModes[scopeID], clearmode.ModeSeal); err != nil {
		return nil, err
	}

	rs := h.sessions[scopeID]
	var dd DeliveryData
	var im *seal.InitialMessage

	if rs == nil {
		var err error
		im, rs, err = seal.Initiate(peerBundle, h.LocalEdPriv)
		if err != nil {
			return nil, fmt.Errorf("dm send: initiate: %w", err)
		}
		dd.EKPub = base64.RawURLEncoding.EncodeToString(im.EKPub[:])
		dd.CTMLKEM = base64.RawURLEncoding.EncodeToString(im.CTMLKEM)
		dd.OPKIndex = im.OPKIndex
	}

	hdr, ciphertext, err := seal.Encrypt(rs, plaintext)
	if err != nil {
		return nil, fmt.Errorf("dm send: encrypt: %w", err)
	}
	h.sessions[scopeID] = rs
	h.scopeModes[scopeID] = clearmode.ModeSeal

	dd.Header = base64.RawURLEncoding.EncodeToString(hdr[:])
	dd.CiphertextFormat = "seal/v1"
	dataJSON, _ := json.Marshal(dd)

	kind := "dm_message"
	if im != nil {
		kind = "dm_create"
	}
	return &Delivery{
		ID:               genID(),
		Kind:             kind,
		ScopeID:          scopeID,
		ScopeType:        "dm",
		SenderPeerID:     h.LocalPeerID,
		RecipientPeerIDs: []string{peerBundle.PeerID},
		Body:             base64.RawURLEncoding.EncodeToString(ciphertext),
		Data:             json.RawMessage(dataJSON),
		CreatedAt:        time.Now().UTC().Format(time.RFC3339Nano),
	}, nil
}

// Messages returns received messages for a scope.
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

func genID() string {
	b, _ := v0crypto.RandomBytes(16)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
