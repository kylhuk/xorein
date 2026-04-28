package discovery

import (
	"crypto/ed25519"
	"fmt"
	"time"

	"github.com/aether/code_aether/pkg/v0_1/crypto"
	"github.com/aether/code_aether/pkg/v0_1/envelope"
)

// canonicalPeerRecord is the subset of PeerRecord that is signed (spec 31 §7.1).
// It omits Source, LastSeen, ExpiresAt, and Signature to produce a stable canonical form.
type canonicalPeerRecord struct {
	PeerID           string   `json:"peer_id"`
	Addresses        []string `json:"addresses"`
	Role             string   `json:"role"`
	Capabilities     []string `json:"capabilities"`
	SigningPublicKey  []byte   `json:"signing_public_key,omitempty"`
	MLDSA65PublicKey []byte   `json:"mldsa65_public_key,omitempty"`
	SignedAt         int64    `json:"signed_at"`
}

// PeerRecordCanonicalBytes returns the deterministic bytes to sign for a peer record
// per spec 31 §7.1: canonical JSON of {peer_id, addresses, role, capabilities,
// signing_public_key, mldsa65_public_key, signed_at} with sorted keys, no signature field.
func PeerRecordCanonicalBytes(r *PeerRecord) ([]byte, error) {
	c := canonicalPeerRecord{
		PeerID:           r.PeerID,
		Addresses:        r.Addresses,
		Role:             r.Role,
		Capabilities:     r.Caps,
		SigningPublicKey:  r.SigningPublicKey,
		MLDSA65PublicKey: r.MLDSA65PublicKey,
		SignedAt:         r.SignedAt,
	}
	b, err := envelope.CanonicalJSON(c)
	if err != nil {
		return nil, fmt.Errorf("peer record canonical bytes: %w", err)
	}
	return b, nil
}

// SignPeerRecord signs r with the given Ed25519 and ML-DSA-65 private keys.
// It sets r.SignedAt (unix seconds), r.SigningPublicKey, r.MLDSA65PublicKey, and r.Signature.
func SignPeerRecord(r *PeerRecord, edPriv ed25519.PrivateKey, mldsaPriv []byte) error {
	r.SignedAt = time.Now().Unix()
	r.SigningPublicKey = []byte(edPriv.Public().(ed25519.PublicKey))

	mldsaPub, err := crypto.MLDSA65PublicKeyFromPrivate(mldsaPriv)
	if err != nil {
		return fmt.Errorf("sign peer record: derive mldsa65 public key: %w", err)
	}
	r.MLDSA65PublicKey = mldsaPub

	msg, err := PeerRecordCanonicalBytes(r)
	if err != nil {
		return fmt.Errorf("sign peer record: canonical bytes: %w", err)
	}

	sig, err := crypto.HybridSign(edPriv, mldsaPriv, msg)
	if err != nil {
		return fmt.Errorf("sign peer record: hybrid sign: %w", err)
	}

	// sig is Ed25519_sig (64 B) || ML-DSA-65_sig (3309 B).
	edSig := sig[:crypto.Ed25519SignatureSize]
	mldsaSig := sig[crypto.Ed25519SignatureSize:]
	r.Signature = envelope.EncodeHybridSig(edSig, mldsaSig)
	return nil
}

// VerifyPeerRecord verifies r.Signature against r's canonical bytes using
// r.SigningPublicKey (Ed25519) and r.MLDSA65PublicKey. Returns nil on success.
func VerifyPeerRecord(r *PeerRecord) error {
	if r.Signature == "" {
		return fmt.Errorf("verify peer record: no signature")
	}
	if len(r.SigningPublicKey) == 0 {
		return fmt.Errorf("verify peer record: missing signing public key")
	}
	if len(r.MLDSA65PublicKey) == 0 {
		return fmt.Errorf("verify peer record: missing mldsa65 public key")
	}

	edSig, mldsaSig, err := envelope.DecodeHybridSig(r.Signature)
	if err != nil {
		return fmt.Errorf("verify peer record: decode signature: %w", err)
	}

	msg, err := PeerRecordCanonicalBytes(r)
	if err != nil {
		return fmt.Errorf("verify peer record: canonical bytes: %w", err)
	}

	sig := make([]byte, len(edSig)+len(mldsaSig))
	copy(sig, edSig)
	copy(sig[len(edSig):], mldsaSig)

	if err := crypto.HybridVerify(ed25519.PublicKey(r.SigningPublicKey), r.MLDSA65PublicKey, msg, sig); err != nil {
		return fmt.Errorf("verify peer record: %w", err)
	}
	return nil
}
