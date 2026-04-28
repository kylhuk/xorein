package crypto

// HKDF label constants for each protocol context.
// Every Expand / DeriveKey call in this codebase MUST use one of these constants —
// never a raw string — so that label collisions across modes are structurally impossible.
//
// Naming convention: "xorein/<mode>/<version>/<purpose>".
// Spec reference column points at docs/spec/v0.1/ sections (filled as waves land).

const (
	// Seal mode (DM, X3DH + Double Ratchet) — spec: 10-mode-seal.md
	LabelSealX3DHMasterSecret = "xorein/seal/v1/x3dh-master-secret"
	LabelSealRootKey          = "xorein/seal/v1/root-key"
	LabelSealChainKeySend     = "xorein/seal/v1/chain-key-send"
	LabelSealChainKeyRecv     = "xorein/seal/v1/chain-key-recv"
	LabelSealMessageKey       = "xorein/seal/v1/message-key"
	LabelSealRatchetStep      = "xorein/seal/v1/ratchet-step"

	// Tree mode (small groups, MLS) — spec: 11-mode-tree.md
	LabelTreeMLSExporterBase    = "xorein/tree/v1/mls-exporter-base"
	LabelTreeMediaShieldExport  = "xorein/mediashield/v1"

	// Crowd mode (large groups, sender keys) — spec: 12-mode-crowd.md
	LabelCrowdSenderKeyRoot    = "xorein/crowd/v1/sender-key-root"
	LabelCrowdSenderKeyChain   = "xorein/crowd/v1/sender-key-chain"
	LabelCrowdEpochKey         = "xorein/crowd/v1/epoch-key"
	LabelCrowdMediaShield      = "xorein/crowd/v1/mediashield"

	// Channel mode (server-wide) — spec: 13-mode-channel.md
	LabelChannelSenderKeyRoot  = "xorein/channel/v1/sender-key-root"
	LabelChannelEpochKey       = "xorein/channel/v1/epoch-key"

	// MediaShield (SFrame voice) — spec: 14-mode-mediashield.md
	// Per-participant key derived via the parent mode's labeled exporter.
	LabelMediaShieldParticipant = "xorein/mediashield/v1/participant"

	// StreamShield (screen-share) — spec: 15-mode-streamshield.md
	LabelStreamShieldParticipant = "xorein/streamshield/v1/participant"

	// Identity / verification — spec: 21-verification.md
	LabelVerificationSafetyNumber = "xorein/verification/v1/safety-number"
)
