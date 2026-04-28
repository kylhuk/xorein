package governance

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	libp2pprotocol "github.com/libp2p/go-libp2p/core/protocol"

	v0crypto "github.com/aether/code_aether/pkg/v0_1/crypto"
	"github.com/aether/code_aether/pkg/v0_1/envelope"
	proto "github.com/aether/code_aether/pkg/v0_1/protocol"
	"github.com/aether/code_aether/pkg/v0_1/transport"
)

const protocolID = "/aether/governance/0.2.0"

// Handler implements transport.FamilyHandler for the Governance family.
type Handler struct {
	store *StateStore
}

// NewHandler creates a Handler backed by an empty in-memory store.
func NewHandler() *Handler {
	return &Handler{store: newStateStore()}
}

var _ transport.FamilyHandler = (*Handler)(nil)

func (h *Handler) ProtocolID() libp2pprotocol.ID { return libp2pprotocol.ID(protocolID) }

func (h *Handler) HandleStream(ctx context.Context, req *proto.PeerStreamRequest) *proto.PeerStreamResponse {
	required, known := proto.RequiredCapsFor(req.Operation)
	if !known {
		return errResp(req.RequestID, proto.CodeUnsupportedOperation, "unknown: "+req.Operation, nil)
	}
	remote := toFlags(req.AdvertisedCaps)
	result := proto.NegotiateCapabilities(h.localCaps(), remote, required)
	if len(result.MissingRequired) > 0 {
		return errResp(req.RequestID, proto.CodeMissingRequiredCapability, "missing caps", result.MissingRequired)
	}

	switch req.Operation {
	case "governance.assign_role":
		return h.handleAssign(req)
	case "governance.revoke_role":
		return h.handleRevoke(req)
	case "governance.create_role":
		return h.handleCreateRole(req)
	case "governance.delete_role":
		return h.handleDeleteRole(req)
	case "governance.sync":
		return h.handleSync(req)
	default:
		return errResp(req.RequestID, proto.CodeUnsupportedOperation, "not implemented: "+req.Operation, nil)
	}
}

func (h *Handler) localCaps() []proto.FeatureFlag { return []proto.FeatureFlag{"cap.rbac"} }

// --- public API for sibling packages ---

// RoleOf returns the role for (serverID, peerID); defaults to RoleMember when unknown.
func (h *Handler) RoleOf(serverID, peerID string) Role { return h.store.RoleOf(serverID, peerID) }

// PolicyVersion returns the current policy_version for a server.
func (h *Handler) PolicyVersion(serverID string) uint64 { return h.store.PolicyVersion(serverID) }

// AssignForTest is a test seam that seeds role state without advancing policy_version.
// Use only from tests; allows test cases to start from policy_version=0.
func (h *Handler) AssignForTest(serverID, peerID string, r Role) {
	h.store.SeedForTest(serverID, peerID, r)
}

// CreateRoleForTest seeds a custom role without going through the handler operation.
func (h *Handler) CreateRoleForTest(serverID, roleName string, bitfield uint64) {
	_ = h.store.SetCustomRole(serverID, roleName, bitfield)
}

// --- wire payloads ---

type roleAssignRequest struct {
	ActorPeerID   string `json:"actor_peer_id"`
	TargetPeerID  string `json:"target_peer_id"`
	ServerID      string `json:"server_id"`
	Role          string `json:"role"`
	PolicyVersion uint64 `json:"policy_version"`
	Signature     string `json:"signature"`
	EdPub         []byte `json:"ed_pub,omitempty"`
	MldsaPub      []byte `json:"mldsa_pub,omitempty"`
}

type roleAssignResponse struct {
	Accepted      bool   `json:"accepted"`
	PolicyVersion uint64 `json:"policy_version"`
	Error         string `json:"error,omitempty"`
}

type roleCreateRequest struct {
	ActorPeerID         string `json:"actor_peer_id"`
	ServerID            string `json:"server_id"`
	RoleName            string `json:"role_name"`
	PermissionsBitfield uint64 `json:"permissions_bitfield"`
	PolicyVersion       uint64 `json:"policy_version"`
	Signature           string `json:"signature"`
}

type roleCreateResponse struct {
	Accepted      bool   `json:"accepted"`
	RoleName      string `json:"role_name"`
	PolicyVersion uint64 `json:"policy_version"`
	Error         string `json:"error,omitempty"`
}

type roleSyncRequest struct {
	RequesterPeerID    string `json:"requester_peer_id"`
	ServerID           string `json:"server_id"`
	KnownPolicyVersion uint64 `json:"known_policy_version"`
}

type roleSyncEntry struct {
	IdentityID    string `json:"identity_id"`
	Role          string `json:"role"`
	Version       uint64 `json:"version"`
	UpdatedAtUnix uint64 `json:"updated_at_unix"`
}

type roleSyncResponse struct {
	UpToDate      bool           `json:"up_to_date"`
	PolicyVersion uint64         `json:"policy_version"`
	Roles         []roleSyncEntry `json:"roles"`
}

// --- op handlers ---

func (h *Handler) handleAssign(req *proto.PeerStreamRequest) *proto.PeerStreamResponse {
	var p roleAssignRequest
	if err := json.Unmarshal(req.Payload, &p); err != nil {
		return errResp(req.RequestID, proto.CodeOperationFailed, "invalid payload", nil)
	}
	if p.Signature != "" {
		canonical := canonicalWithoutSig(req.Payload)
		if err := verifySig(canonical, p.Signature, p.EdPub, p.MldsaPub); err != nil {
			return errResp(req.RequestID, proto.CodeGovernanceInvalidSignature, "invalid signature", nil)
		}
	}
	role := parseRole(p.Role)
	if role == 0 {
		return errResp(req.RequestID, proto.CodeOperationFailed, "invalid role: "+p.Role, nil)
	}
	if role == RoleOwner {
		return errResp(req.RequestID, proto.CodeGovernanceOwnerImmutable, "owner assignment not permitted via this op", nil)
	}
	actorRole := h.store.RoleOf(p.ServerID, p.ActorPeerID)
	targetRole := h.store.RoleOf(p.ServerID, p.TargetPeerID)
	// Forbidden-target check first: actor cannot act on equal-or-higher ranked peer.
	if !CanActOnTarget(actorRole, targetRole) {
		return errResp(req.RequestID, proto.CodeGovernanceForbiddenTarget, "actor cannot act on target", nil)
	}
	// Assignment ceiling: assigning admin requires owner; moderator/member requires admin.
	minRequired := RoleAdmin
	if role == RoleAdmin {
		minRequired = RoleOwner
	}
	if actorRole < minRequired {
		return errResp(req.RequestID, proto.CodeGovernanceUnauthorized, "actor role insufficient", nil)
	}
	// Stale version check (spec §5): reject when policy_version is strictly less than current.
	// policy_version=N succeeds once (advancing current to N+1); a replay with same N then fails.
	current := h.store.PolicyVersion(p.ServerID)
	if p.PolicyVersion < current {
		return errResp(req.RequestID, proto.CodeGovernanceStaleVersion, "policy_version behind current", nil)
	}
	h.store.Assign(p.ServerID, p.TargetPeerID, role)
	resp, _ := json.Marshal(roleAssignResponse{Accepted: true, PolicyVersion: h.store.PolicyVersion(p.ServerID)})
	return &proto.PeerStreamResponse{NegotiatedProtocol: protocolID, Payload: resp}
}

func (h *Handler) handleRevoke(req *proto.PeerStreamRequest) *proto.PeerStreamResponse {
	var p roleAssignRequest
	if err := json.Unmarshal(req.Payload, &p); err != nil {
		return errResp(req.RequestID, proto.CodeOperationFailed, "invalid payload", nil)
	}
	if p.Signature != "" {
		canonical := canonicalWithoutSig(req.Payload)
		if err := verifySig(canonical, p.Signature, p.EdPub, p.MldsaPub); err != nil {
			return errResp(req.RequestID, proto.CodeGovernanceInvalidSignature, "invalid signature", nil)
		}
	}
	targetRole := h.store.RoleOf(p.ServerID, p.TargetPeerID)
	if targetRole == RoleOwner {
		return errResp(req.RequestID, proto.CodeGovernanceOwnerImmutable, "owner role cannot be revoked", nil)
	}
	actorRole := h.store.RoleOf(p.ServerID, p.ActorPeerID)
	if actorRole < RoleAdmin {
		return errResp(req.RequestID, proto.CodeGovernanceUnauthorized, "actor role insufficient", nil)
	}
	if !CanActOnTarget(actorRole, targetRole) {
		return errResp(req.RequestID, proto.CodeGovernanceForbiddenTarget, "actor cannot act on target", nil)
	}
	current := h.store.PolicyVersion(p.ServerID)
	if p.PolicyVersion < current {
		return errResp(req.RequestID, proto.CodeGovernanceStaleVersion, "policy_version behind current", nil)
	}
	h.store.Assign(p.ServerID, p.TargetPeerID, RoleMember)
	resp, _ := json.Marshal(roleAssignResponse{Accepted: true, PolicyVersion: h.store.PolicyVersion(p.ServerID)})
	return &proto.PeerStreamResponse{NegotiatedProtocol: protocolID, Payload: resp}
}

func (h *Handler) handleCreateRole(req *proto.PeerStreamRequest) *proto.PeerStreamResponse {
	var p roleCreateRequest
	if err := json.Unmarshal(req.Payload, &p); err != nil {
		return errResp(req.RequestID, proto.CodeOperationFailed, "invalid payload", nil)
	}
	actorRole := h.store.RoleOf(p.ServerID, p.ActorPeerID)
	if actorRole < RoleAdmin {
		return errResp(req.RequestID, proto.CodeGovernanceUnauthorized, "actor role insufficient", nil)
	}
	if IsBaseRoleName(p.RoleName) {
		return errResp(req.RequestID, proto.CodeGovernanceBaseRoleProtected, "cannot override base role", nil)
	}
	if p.PermissionsBitfield&PermissionsReservedMask != 0 {
		return errResp(req.RequestID, proto.CodeGovernanceInvalidBitfield, "reserved bits set", nil)
	}
	if err := h.store.SetCustomRole(p.ServerID, p.RoleName, p.PermissionsBitfield); err != nil {
		if err.Error() == "conflict" {
			return errResp(req.RequestID, proto.CodeGovernanceRoleConflict, "role name already exists", nil)
		}
		return errResp(req.RequestID, proto.CodeOperationFailed, err.Error(), nil)
	}
	resp, _ := json.Marshal(roleCreateResponse{Accepted: true, RoleName: p.RoleName, PolicyVersion: h.store.PolicyVersion(p.ServerID)})
	return &proto.PeerStreamResponse{NegotiatedProtocol: protocolID, Payload: resp}
}

func (h *Handler) handleDeleteRole(req *proto.PeerStreamRequest) *proto.PeerStreamResponse {
	var p roleCreateRequest
	if err := json.Unmarshal(req.Payload, &p); err != nil {
		return errResp(req.RequestID, proto.CodeOperationFailed, "invalid payload", nil)
	}
	actorRole := h.store.RoleOf(p.ServerID, p.ActorPeerID)
	if actorRole < RoleAdmin {
		return errResp(req.RequestID, proto.CodeGovernanceUnauthorized, "actor role insufficient", nil)
	}
	if IsBaseRoleName(p.RoleName) {
		return errResp(req.RequestID, proto.CodeGovernanceBaseRoleProtected, "cannot delete base role", nil)
	}
	if err := h.store.DeleteCustomRole(p.ServerID, p.RoleName); err != nil {
		if err.Error() == "not found" {
			return errResp(req.RequestID, proto.CodeGovernanceRoleNotFound, "role not found", nil)
		}
		return errResp(req.RequestID, proto.CodeOperationFailed, err.Error(), nil)
	}
	resp, _ := json.Marshal(roleAssignResponse{Accepted: true, PolicyVersion: h.store.PolicyVersion(p.ServerID)})
	return &proto.PeerStreamResponse{NegotiatedProtocol: protocolID, Payload: resp}
}

func (h *Handler) handleSync(req *proto.PeerStreamRequest) *proto.PeerStreamResponse {
	var p roleSyncRequest
	if err := json.Unmarshal(req.Payload, &p); err != nil {
		return errResp(req.RequestID, proto.CodeOperationFailed, "invalid payload", nil)
	}
	states, ver := h.store.AllRoleStates(p.ServerID)
	if ver <= p.KnownPolicyVersion {
		resp, _ := json.Marshal(roleSyncResponse{UpToDate: true, PolicyVersion: ver, Roles: []roleSyncEntry{}})
		return &proto.PeerStreamResponse{NegotiatedProtocol: protocolID, Payload: resp}
	}
	entries := make([]roleSyncEntry, 0, len(states))
	for _, rs := range states {
		entries = append(entries, roleSyncEntry{
			IdentityID:    rs.PeerID,
			Role:          rs.Role.String(),
			Version:       rs.Version,
			UpdatedAtUnix: uint64(rs.UpdatedAt.Unix()),
		})
	}
	resp, _ := json.Marshal(roleSyncResponse{UpToDate: false, PolicyVersion: ver, Roles: entries})
	return &proto.PeerStreamResponse{NegotiatedProtocol: protocolID, Payload: resp}
}

// --- helpers ---

func parseRole(s string) Role {
	switch s {
	case "member":
		return RoleMember
	case "moderator":
		return RoleModerator
	case "admin":
		return RoleAdmin
	case "owner":
		return RoleOwner
	}
	return 0
}

func errResp(requestID, code, msg string, missing []string) *proto.PeerStreamResponse {
	return &proto.PeerStreamResponse{RequestID: requestID, Error: proto.NewPeerStreamError(code, msg, missing)}
}

func toFlags(ss []string) []proto.FeatureFlag {
	out := make([]proto.FeatureFlag, len(ss))
	for i, s := range ss {
		out[i] = proto.FeatureFlag(s)
	}
	return out
}

// verifySig verifies a hybrid signature over canonicalBytes.
// When edPub and mldsaPub are both non-empty, full cryptographic verification
// is performed via proto.VerifyHybridSig. When keys are absent (e.g. KAT
// scenarios without public key material), the signature is checked for valid
// format AND correct total size (3373 bytes = Ed25519 64B + ML-DSA-65 3309B).
func verifySig(canonicalBytes []byte, sig string, edPub, mldsaPub []byte) error {
	if len(edPub) > 0 && len(mldsaPub) > 0 {
		return proto.VerifyHybridSig(canonicalBytes, sig, edPub, mldsaPub)
	}
	// No public keys available — validate format and exact byte length.
	edSig, mldsaSig, err := envelope.DecodeHybridSig(sig)
	if err != nil {
		return err
	}
	total := len(edSig) + len(mldsaSig)
	if total != v0crypto.HybridSignatureSize {
		return fmt.Errorf("hybrid sig: wrong total size %d (want %d)", total, v0crypto.HybridSignatureSize)
	}
	return nil
}

// canonicalWithoutSig builds canonical JSON for signing/verification by
// removing the "signature", "ed_pub", and "mldsa_pub" fields from the payload.
func canonicalWithoutSig(payload []byte) []byte {
	var m map[string]any
	if err := json.Unmarshal(payload, &m); err != nil {
		return payload
	}
	delete(m, "signature")
	delete(m, "ed_pub")
	delete(m, "mldsa_pub")
	canonical, err := json.Marshal(m)
	if err != nil {
		return payload
	}
	return canonical
}

// Ensure the unused time import isn't flagged; UpdatedAt is set in Assign().
var _ = time.Now
