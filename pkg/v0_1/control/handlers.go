package control

import (
	chatpkg "github.com/aether/code_aether/pkg/v0_1/family/chat"
	dmpkg "github.com/aether/code_aether/pkg/v0_1/family/dm"
	friendspkg "github.com/aether/code_aether/pkg/v0_1/family/friends"
	groupdmpkg "github.com/aether/code_aether/pkg/v0_1/family/groupdm"
	governancepkg "github.com/aether/code_aether/pkg/v0_1/family/governance"
	identitypkg "github.com/aether/code_aether/pkg/v0_1/family/identity"
	manifestpkg "github.com/aether/code_aether/pkg/v0_1/family/manifest"
	moderationpkg "github.com/aether/code_aether/pkg/v0_1/family/moderation"
	notifypkg "github.com/aether/code_aether/pkg/v0_1/family/notify"
	peerfamily "github.com/aether/code_aether/pkg/v0_1/family/peer"
	presencepkg "github.com/aether/code_aether/pkg/v0_1/family/presence"
	syncpkg "github.com/aether/code_aether/pkg/v0_1/family/sync"
	voicepkg "github.com/aether/code_aether/pkg/v0_1/family/voice"
)

// Handlers bundles all family handler references needed by the control API.
type Handlers struct {
	// PeerID is the local libp2p peer ID (string form).
	PeerID string
	// DisplayName is the local identity display name, if set.
	DisplayName string

	// ConnectionTypeFn returns the connection type string for a given peer ID.
	// Injected by the runtime to avoid an import cycle.
	ConnectionTypeFn func(peerID string) string

	// LatencyFn returns the EWMA latency in milliseconds for a given peer ID,
	// or -1 if unavailable (spec 32 §5 — latency indicator on connection endpoint).
	// Injected by the runtime to avoid an import cycle.
	LatencyFn func(peerID string) int64

	// AddManualPeerFn propagates a newly-added manual peer address to the discovery
	// layer at runtime. Optional; if nil, only the control-layer state is updated.
	AddManualPeerFn func(addr string)

	// RemoveManualPeerFn removes a manual peer address from the discovery layer.
	// Optional; if nil, only the control-layer state is updated.
	RemoveManualPeerFn func(addr string)

	// AddRelayFn propagates a newly-added relay multiaddr to the relay-client config.
	// Optional; if nil, only the control-layer state is updated.
	AddRelayFn func(multiaddr string)

	// BackupKeyFn serialises the local Ed25519 private key for backup.
	// Returns raw 64-byte Ed25519 private key bytes. Optional; if nil, backup returns an error.
	BackupKeyFn func() ([]byte, error)

	// RestoreKeyFn reinstalls a restored Ed25519 private key.
	// Optional; if nil, restore returns an error.
	RestoreKeyFn func(privKey []byte) error

	Chat       *chatpkg.Handler
	DM         *dmpkg.Handler
	GroupDM    *groupdmpkg.Handler
	Voice      *voicepkg.Handler
	Identity   *identitypkg.Handler
	Manifest   *manifestpkg.Handler
	Presence   *presencepkg.Handler
	Notify     *notifypkg.Handler
	Friends    *friendspkg.Handler
	Peer       *peerfamily.Handler
	Sync       *syncpkg.Handler
	Governance *governancepkg.Handler
	Moderation *moderationpkg.Handler
}
