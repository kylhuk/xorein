// Package groupdm implements the /aether/groupdm/0.2.0 protocol family (Tree-mode group DMs).
// Source: docs/spec/v0.1/46-family-groupdm.md
package groupdm

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	libp2pprotocol "github.com/libp2p/go-libp2p/core/protocol"

	tree "github.com/aether/code_aether/pkg/v0_1/mode/tree"
	proto "github.com/aether/code_aether/pkg/v0_1/protocol"
	"github.com/aether/code_aether/pkg/v0_1/transport"
)

const protocolID = "/aether/groupdm/0.2.0"

// Handler implements transport.FamilyHandler for the GroupDM family.
type Handler struct {
	mu       sync.RWMutex
	LocalPeerID string
	groups   map[string]*tree.GroupState // groupID → state
	messages map[string][]MessageRecord  // groupID → messages
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
	case "groupdm.create":
		return h.handleCreate(req)
	case "groupdm.send":
		return h.handleSend(req)
	case "groupdm.add":
		return h.handleAdd(req)
	case "groupdm.remove":
		return h.handleRemove(req)
	default:
		return errorResp(req.RequestID, proto.CodeUnsupportedOperation, "not implemented: "+req.Operation, nil)
	}
}

func (h *Handler) localCaps() []proto.FeatureFlag {
	return []proto.FeatureFlag{"cap.group-dm", "mode.tree"}
}

// MessageRecord holds a decrypted GroupDM message.
type MessageRecord struct {
	ID        string
	GroupID   string
	SenderID  string
	Plaintext []byte
	EpochID   uint64
	ReceivedAt time.Time
}

// CreateGroup creates a new group and returns its ID.
func (h *Handler) CreateGroup(groupID string, creator tree.Member) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.groups == nil {
		h.groups = make(map[string]*tree.GroupState)
	}
	if _, exists := h.groups[groupID]; exists {
		return fmt.Errorf("groupdm: group %q already exists", groupID)
	}
	g, err := tree.NewGroup(groupID, creator)
	if err != nil {
		return fmt.Errorf("groupdm: create: %w", err)
	}
	h.groups[groupID] = g
	return nil
}

// AddMember adds a peer to an existing group.
func (h *Handler) AddMember(groupID string, member tree.Member) (*tree.Commit, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	g, err := h.getGroup(groupID)
	if err != nil {
		return nil, err
	}
	return tree.AddMember(g, member)
}

// RemoveMember removes a peer from a group.
func (h *Handler) RemoveMember(groupID, peerID string) (*tree.Commit, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	g, err := h.getGroup(groupID)
	if err != nil {
		return nil, err
	}
	return tree.RemoveMember(g, peerID)
}

// SendMessage encrypts and stores a message in the group.
func (h *Handler) SendMessage(groupID, senderID string, plaintext []byte) (*tree.Ciphertext, *tree.Commit, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	g, err := h.getGroup(groupID)
	if err != nil {
		return nil, nil, err
	}
	ct, commit, err := tree.Encrypt(g, senderID, plaintext)
	if err != nil {
		return nil, nil, fmt.Errorf("groupdm: encrypt: %w", err)
	}
	return ct, commit, nil
}

// ReceiveMessage decrypts a received group ciphertext and stores it.
func (h *Handler) ReceiveMessage(groupID string, ct *tree.Ciphertext) ([]byte, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	g, err := h.getGroup(groupID)
	if err != nil {
		return nil, err
	}
	pt, err := tree.Decrypt(g, ct)
	if err != nil {
		return nil, fmt.Errorf("groupdm: decrypt: %w", err)
	}
	if h.messages == nil {
		h.messages = make(map[string][]MessageRecord)
	}
	h.messages[groupID] = append(h.messages[groupID], MessageRecord{
		GroupID:    groupID,
		SenderID:   ct.SenderID,
		Plaintext:  pt,
		EpochID:    ct.EpochID,
		ReceivedAt: time.Now(),
	})
	return pt, nil
}

// Messages returns decrypted messages for a group.
func (h *Handler) Messages(groupID string) []MessageRecord {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.messages[groupID]
}

// GroupStateForTest returns the raw GroupState for test introspection. Not for production use.
func (h *Handler) GroupStateForTest(groupID string) *tree.GroupState {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.groups[groupID]
}

// IsMember checks if a peer is in the group.
func (h *Handler) IsMember(groupID, peerID string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	g := h.groups[groupID]
	if g == nil {
		return false
	}
	return tree.IsMember(g, peerID)
}

func (h *Handler) getGroup(groupID string) (*tree.GroupState, error) {
	g := h.groups[groupID]
	if g == nil {
		return nil, fmt.Errorf("groupdm: group %q not found", groupID)
	}
	return g, nil
}

// --- wire operations ---

type createPayload struct {
	GroupID  string      `json:"group_id"`
	Creator  tree.Member `json:"creator"`
}

type sendPayload struct {
	GroupID    string `json:"group_id"`
	SenderID   string `json:"sender_id"`
	Ciphertext string `json:"ciphertext"` // base64url(JSON(tree.Ciphertext))
}

func (h *Handler) handleCreate(req *proto.PeerStreamRequest) *proto.PeerStreamResponse {
	var p createPayload
	if err := json.Unmarshal(req.Payload, &p); err != nil {
		return errorResp(req.RequestID, proto.CodeOperationFailed, "invalid payload", nil)
	}
	if err := h.CreateGroup(p.GroupID, p.Creator); err != nil {
		return errorResp(req.RequestID, proto.CodeOperationFailed, err.Error(), nil)
	}
	resp, _ := json.Marshal(map[string]any{"group_id": p.GroupID, "ok": true})
	return &proto.PeerStreamResponse{NegotiatedProtocol: protocolID, Payload: resp}
}

func (h *Handler) handleSend(req *proto.PeerStreamRequest) *proto.PeerStreamResponse {
	var p sendPayload
	if err := json.Unmarshal(req.Payload, &p); err != nil {
		return errorResp(req.RequestID, proto.CodeOperationFailed, "invalid payload", nil)
	}
	ctJSON, err := base64.RawURLEncoding.DecodeString(p.Ciphertext)
	if err != nil {
		return errorResp(req.RequestID, proto.CodeOperationFailed, "invalid ciphertext encoding", nil)
	}
	var ct tree.Ciphertext
	if err := json.Unmarshal(ctJSON, &ct); err != nil {
		return errorResp(req.RequestID, proto.CodeOperationFailed, "invalid ciphertext", nil)
	}
	pt, err := h.ReceiveMessage(p.GroupID, &ct)
	if err != nil {
		return errorResp(req.RequestID, proto.CodeOperationFailed, "decrypt: "+err.Error(), nil)
	}
	resp, _ := json.Marshal(map[string]any{"group_id": p.GroupID, "plaintext_len": len(pt)})
	return &proto.PeerStreamResponse{NegotiatedProtocol: protocolID, Payload: resp}
}

type addPayload struct {
	GroupID string      `json:"group_id"`
	Member  tree.Member `json:"member"`
}

func (h *Handler) handleAdd(req *proto.PeerStreamRequest) *proto.PeerStreamResponse {
	var p addPayload
	if err := json.Unmarshal(req.Payload, &p); err != nil {
		return errorResp(req.RequestID, proto.CodeOperationFailed, "invalid payload", nil)
	}
	commit, err := h.AddMember(p.GroupID, p.Member)
	if err != nil {
		return errorResp(req.RequestID, proto.CodeOperationFailed, err.Error(), nil)
	}
	resp, _ := json.Marshal(commit)
	return &proto.PeerStreamResponse{NegotiatedProtocol: protocolID, Payload: resp}
}

type removePayload struct {
	GroupID string `json:"group_id"`
	PeerID  string `json:"peer_id"`
}

func (h *Handler) handleRemove(req *proto.PeerStreamRequest) *proto.PeerStreamResponse {
	var p removePayload
	if err := json.Unmarshal(req.Payload, &p); err != nil {
		return errorResp(req.RequestID, proto.CodeOperationFailed, "invalid payload", nil)
	}
	commit, err := h.RemoveMember(p.GroupID, p.PeerID)
	if err != nil {
		return errorResp(req.RequestID, proto.CodeOperationFailed, err.Error(), nil)
	}
	resp, _ := json.Marshal(commit)
	return &proto.PeerStreamResponse{NegotiatedProtocol: protocolID, Payload: resp}
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
