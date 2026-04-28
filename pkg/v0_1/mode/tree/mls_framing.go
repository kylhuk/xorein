// mls_framing.go implements MLS message framing per RFC 9420 §6 for the
// Xorein hybrid ciphersuite (spec 12 §4).
// Source: docs/spec/v0.1/12-mode-tree.md §4
package tree

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"

	v0crypto "github.com/aether/code_aether/pkg/v0_1/crypto"
)

// ContentType constants for MLSMessage.
const (
	ContentTypeWelcome     = "welcome"
	ContentTypeCommit      = "commit"
	ContentTypeProposal    = "proposal"
	ContentTypeApplication = "application"
)

// ProposalType constants for Proposal.Type.
const (
	ProposalTypeAdd    = "add"
	ProposalTypeRemove = "remove"
	ProposalTypeUpdate = "update"
)

var (
	// ErrMLSMessageVerify is returned when an MLSMessage signature fails.
	ErrMLSMessageVerify = errors.New("tree/mls: message signature verification failed")
	// ErrWelcomeNoMatch is returned when ApplyWelcome cannot find a matching encrypted group secret.
	ErrWelcomeNoMatch = errors.New("tree/mls: welcome contains no encrypted group secret for this member")
)

// MLSMessage wraps any MLS content type per RFC 9420 §6.
type MLSMessage struct {
	Version     uint8  `json:"version"`
	ContentType string `json:"content_type"` // "welcome" | "commit" | "proposal" | "application"
	GroupID     string `json:"group_id"`
	EpochID     uint64 `json:"epoch_id"`
	Payload     []byte `json:"payload"`
	// Signature is the hybrid signature (Ed25519 || ML-DSA-65) over the
	// canonical form of all fields except Signature itself.
	Signature []byte `json:"signature"`
}

// Welcome is sent to new members to give them the current group state.
// It contains per-member encrypted copies of the epoch secrets.
type Welcome struct {
	GroupID               string                 `json:"group_id"`
	EpochID               uint64                 `json:"epoch_id"`
	TreeHash              []byte                 `json:"tree_hash"` // SHA-256 of member list
	EncryptedGroupSecrets []EncryptedGroupSecret `json:"encrypted_group_secrets"`
}

// EncryptedGroupSecret holds the epoch secrets encrypted to one new member's init key.
type EncryptedGroupSecret struct {
	// RecipientInitKeyPub is the recipient's hybrid init public key (used to identify recipient).
	RecipientInitKeyPub []byte `json:"recipient_init_key_pub"`
	// EncryptedSecret is the hybrid KEM-encapsulated EpochSecrets JSON.
	EncryptedSecret []byte `json:"encrypted_secret"`
}

// MLSCommit advances the epoch: carries the new path update and any proposals.
type MLSCommit struct {
	ProposerLeaf uint32       `json:"proposer_leaf"`
	// UpdatePath holds the per-co-path encrypted path secrets from RatchetTree.UpdatePath.
	UpdatePath   [][]byte     `json:"update_path"`
	Proposals    []MLSProposal `json:"proposals"`
}

// MLSProposal is an Add, Remove, or Update operation.
type MLSProposal struct {
	Type    string          `json:"type"`    // "add" | "remove" | "update"
	Payload json.RawMessage `json:"payload"` // type-specific JSON
}

// epochSecretsEnvelope is the JSON envelope used inside EncryptedGroupSecret.
type epochSecretsEnvelope struct {
	EpochID        uint64 `json:"epoch_id"`
	HandshakeKey   []byte `json:"handshake_key"`
	ApplicationKey []byte `json:"application_key"`
	ExporterSecret []byte `json:"exporter_secret"`
}

// NewWelcome builds a Welcome message for a new member using their KeyPackage.
// The epoch secrets are encrypted to the member's hybrid init key.
func NewWelcome(g *GroupState, es *EpochSecrets, newMemberKP *KeyPackage) (*Welcome, error) {
	// Serialize epoch secrets into envelope.
	env := epochSecretsEnvelope{
		EpochID:        es.EpochID,
		HandshakeKey:   es.HandshakeKey[:],
		ApplicationKey: es.ApplicationKey[:],
		ExporterSecret: es.ExporterSecret[:],
	}
	envJSON, err := json.Marshal(env)
	if err != nil {
		return nil, fmt.Errorf("tree/mls: marshal epoch secrets: %w", err)
	}

	// Build the recipient's hybrid init public key from their KeyPackage.
	recipientInitPub := make([]byte, HybridPublicKeySize)
	copy(recipientInitPub[:v0crypto.X25519KeySize], newMemberKP.InitKey)
	copy(recipientInitPub[v0crypto.X25519KeySize:], newMemberKP.MLKEMPub)

	encSecret, err := hybridKEMEncapsulate(recipientInitPub, envJSON)
	if err != nil {
		return nil, fmt.Errorf("tree/mls: encrypt group secret: %w", err)
	}

	welcome := &Welcome{
		GroupID:  g.GroupID,
		EpochID:  es.EpochID,
		TreeHash: groupTreeHash(g),
		EncryptedGroupSecrets: []EncryptedGroupSecret{
			{
				RecipientInitKeyPub: recipientInitPub,
				EncryptedSecret:     encSecret,
			},
		},
	}
	return welcome, nil
}

// ApplyWelcome processes a Welcome and initializes the receiving node's GroupState.
// myKP is this member's KeyPackage; myHybridPriv is the corresponding init private key
// (HybridPrivateKeySize bytes: x25519_priv || mlkem_priv).
//
// Returns a new GroupState with the epoch secrets loaded from the Welcome.
// The caller should populate GroupState.Members after receiving the member list.
func ApplyWelcome(w *Welcome, myKP *KeyPackage, myHybridPriv []byte) (*GroupState, error) {
	// Build the recipient's init public key to find our entry.
	myInitPub := make([]byte, HybridPublicKeySize)
	copy(myInitPub[:v0crypto.X25519KeySize], myKP.InitKey)
	copy(myInitPub[v0crypto.X25519KeySize:], myKP.MLKEMPub)

	var encSecret []byte
	for _, egs := range w.EncryptedGroupSecrets {
		if bytesEqual(egs.RecipientInitKeyPub, myInitPub) {
			encSecret = egs.EncryptedSecret
			break
		}
	}
	if encSecret == nil {
		return nil, ErrWelcomeNoMatch
	}

	// Decrypt the epoch secrets envelope.
	envJSON, err := hybridKEMDecapsulate(myHybridPriv, encSecret)
	if err != nil {
		return nil, fmt.Errorf("tree/mls: decrypt group secret: %w", err)
	}

	var env epochSecretsEnvelope
	if err := json.Unmarshal(envJSON, &env); err != nil {
		return nil, fmt.Errorf("tree/mls: unmarshal epoch secrets: %w", err)
	}

	es := &EpochSecrets{
		EpochID: env.EpochID,
		GroupID: w.GroupID,
	}
	copy(es.HandshakeKey[:], env.HandshakeKey)
	copy(es.ApplicationKey[:], env.ApplicationKey)
	copy(es.ExporterSecret[:], env.ExporterSecret)

	// Build the GroupState from the welcome data.
	// The epoch key for the existing GroupState.CurrentEpoch is derived from
	// EpochSecrets.ApplicationKey (AES-128-GCM uses first 16 bytes).
	g := &GroupState{
		GroupID: w.GroupID,
		CurrentEpoch: &EpochState{
			EpochID:  env.EpochID,
			EpochKey: append([]byte(nil), es.ApplicationKey[:]...),
		},
		RootKey:  append([]byte(nil), es.ExporterSecret[:]...),
		MLSEpoch: es,
	}
	return g, nil
}

// SignMLSMessage signs an MLSMessage with the given hybrid identity keys.
// edPriv must be an ed25519.PrivateKey (64 bytes); mldsaPriv is the ML-DSA-65 private key.
func SignMLSMessage(msg *MLSMessage, edPriv ed25519.PrivateKey, mldsaPriv []byte) error {
	canonical := mlsMessageSigningBytes(msg)
	sig, err := v0crypto.HybridSign(edPriv, mldsaPriv, canonical)
	if err != nil {
		return fmt.Errorf("tree/mls: sign message: %w", err)
	}
	msg.Signature = sig
	return nil
}

// VerifyMLSMessage verifies the hybrid signature on an MLSMessage.
func VerifyMLSMessage(msg *MLSMessage, edPub ed25519.PublicKey, mldsaPub []byte) error {
	canonical := mlsMessageSigningBytes(msg)
	if err := v0crypto.HybridVerify(edPub, mldsaPub, canonical, msg.Signature); err != nil {
		return ErrMLSMessageVerify
	}
	return nil
}

// mlsMessageSigningBytes returns the canonical byte form of an MLSMessage for signing.
// Encodes: version(1) || content_type(len-prefixed) || group_id(len-prefixed)
//          || epoch_id(8 BE) || payload(4-byte-len-prefixed)
func mlsMessageSigningBytes(msg *MLSMessage) []byte {
	var buf []byte
	buf = append(buf, msg.Version)
	buf = appendLenPrefixed(buf, []byte(msg.ContentType))
	buf = appendLenPrefixed(buf, []byte(msg.GroupID))
	var epochBuf [8]byte
	binary.BigEndian.PutUint64(epochBuf[:], msg.EpochID)
	buf = append(buf, epochBuf[:]...)
	// 4-byte payload length prefix
	var plenBuf [4]byte
	binary.BigEndian.PutUint32(plenBuf[:], uint32(len(msg.Payload)))
	buf = append(buf, plenBuf[:]...)
	buf = append(buf, msg.Payload...)
	return buf
}

// groupTreeHash returns a deterministic hash of the group's member list.
func groupTreeHash(g *GroupState) []byte {
	h := sha256.New()
	for _, m := range g.Members {
		h.Write([]byte(m.PeerID))
		h.Write([]byte{0})
	}
	return h.Sum(nil)
}

func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
