package crypto

// KDF label constants for the v0.1 locked cryptographic profile.
// All labels follow the pattern xorein/<mode>/<version>/<purpose>.
// Source: docs/spec/v0.1/01-cryptographic-primitives.md §3.1
const (
	LabelSealX3DHMaster  = "xorein/seal/v1/x3dh-master-secret"
	LabelSealHybridMaster = "xorein/seal/v1/hybrid-master"
	LabelSealRootKey     = "xorein/seal/v1/root-key"
	LabelSealChainKey    = "xorein/seal/v1/chain-key"
	LabelSealMessageKey  = "xorein/seal/v1/message-key"
	LabelSealRatchetStep = "xorein/seal/v1/ratchet-step"
	LabelSealSPKSign     = "xorein/seal/v1/spk-sign"

	LabelTreeExporter  = "xorein/tree/v1/exporter"
	LabelTreeEpochRoot = "xorein/tree/v1/epoch-root"

	LabelCrowdSenderKey = "xorein/crowd/v1/sender-key"
	LabelCrowdEpochRoot = "xorein/crowd/v1/epoch-root"

	LabelChannelSenderKey = "xorein/channel/v1/sender-key"

	LabelMediaShield     = "xorein/mediashield/v1"
	LabelMediaShieldPeer = "xorein/mediashield/v1/peer/" // append peer_id at use
)
