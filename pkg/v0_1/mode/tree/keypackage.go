// KeyPackage implements MLS KeyPackage with hybrid credential per spec 12 §2–3.
// Source: docs/spec/v0.1/12-mode-tree.md §2.1
package tree

import (
	"crypto/ed25519"
	"encoding/binary"
	"errors"
	"fmt"

	v0crypto "github.com/aether/code_aether/pkg/v0_1/crypto"
)

// ErrInvalidKeyPackage is returned when a KeyPackage signature fails verification.
var ErrInvalidKeyPackage = errors.New("tree/mls: key package signature verification failed")

// HybridCredential carries the dual identity keys per spec 12 §3.2.
// leaf_node.credential carries both the Ed25519 and ML-DSA-65 public keys.
type HybridCredential struct {
	PeerID     string `json:"peer_id"`
	Ed25519Pub []byte `json:"ed25519_pub"`  // 32 bytes
	MLDSA65Pub []byte `json:"mldsa65_pub"`  // 1952 bytes
}

// KeyPackage is the MLS key package with hybrid credential per spec 12 §3.
//
// Fields beyond the standard MLS spec:
//   - Credential.MLDSA65Pub — ML-DSA-65 verification key (1952 bytes)
//   - MLKEMPub — ML-KEM-768 encapsulation public key (1184 bytes)
//   - InitKey carries the X25519 ephemeral init key (32 bytes)
//
// The combined hybrid KEM init key is InitKey (X25519) || MLKEMPub (ML-KEM-768).
// Source: spec 12 §2.1, §1.2.
type KeyPackage struct {
	// MLS version; always MLSVersion (1).
	Version uint8 `json:"version"`
	// CiphersuiteID is always CiphersuiteID (0xFF01).
	CiphersuiteID uint16 `json:"ciphersuite_id"`
	// InitKey is the X25519 ephemeral init public key (32 bytes).
	// For the hybrid KEM, the full init key is InitKey || MLKEMPub.
	InitKey []byte `json:"init_key"`
	// MLKEMPub is the ML-KEM-768 encapsulation public key (1184 bytes).
	MLKEMPub []byte `json:"mlkem_pub"`
	// Credential carries the member's dual identity keys.
	Credential HybridCredential `json:"credential"`
	// Signature is the hybrid signature (3373 bytes: Ed25519 || ML-DSA-65)
	// over the canonical serialisation of this KeyPackage (all fields except Signature).
	Signature []byte `json:"signature"`
}

// GenerateKeyPackage generates a new KeyPackage for the given identity.
//
// Parameters:
//   - peerID: the member's peer ID string
//   - edPriv: Ed25519 private key (64 bytes)
//   - mldsaPriv: ML-DSA-65 private key (4032 bytes)
//   - edPub: Ed25519 public key (32 bytes)
//   - mldsaPub: ML-DSA-65 public key (1952 bytes)
//
// Generates a fresh X25519 + ML-KEM-768 ephemeral init keypair.
// Returns the KeyPackage (with signature) and the combined hybrid init private key
// (HybridPrivateKeySize bytes: x25519_priv || mlkem_priv) for use in TreeKEM /
// Welcome decryption operations.
func GenerateKeyPackage(
	peerID string,
	edPriv ed25519.PrivateKey,
	mldsaPriv []byte,
	edPub, mldsaPub []byte,
) (*KeyPackage, []byte, error) {
	// Generate ephemeral X25519 init keypair.
	x25519Priv, x25519Pub, err := v0crypto.GenerateX25519Keypair()
	if err != nil {
		return nil, nil, fmt.Errorf("tree/mls: generate x25519 init key: %w", err)
	}

	// Generate ephemeral ML-KEM-768 init keypair.
	mlkemPub, mlkemPriv, err := v0crypto.GenerateMLKEM768Keypair()
	if err != nil {
		return nil, nil, fmt.Errorf("tree/mls: generate mlkem init key: %w", err)
	}

	kp := &KeyPackage{
		Version:       MLSVersion,
		CiphersuiteID: CiphersuiteID,
		InitKey:       x25519Pub[:],
		MLKEMPub:      mlkemPub,
		Credential: HybridCredential{
			PeerID:     peerID,
			Ed25519Pub: edPub,
			MLDSA65Pub: mldsaPub,
		},
	}

	// Sign over the canonical form.
	canonical := keyPackageSigningBytes(kp)
	sig, err := v0crypto.HybridSign(edPriv, mldsaPriv, canonical)
	if err != nil {
		return nil, nil, fmt.Errorf("tree/mls: sign key package: %w", err)
	}
	kp.Signature = sig

	// Return the combined hybrid private key: x25519_priv(32) || mlkem_priv(2400).
	hybridPriv := make([]byte, v0crypto.X25519KeySize+len(mlkemPriv))
	copy(hybridPriv[:v0crypto.X25519KeySize], x25519Priv[:])
	copy(hybridPriv[v0crypto.X25519KeySize:], mlkemPriv)

	return kp, hybridPriv, nil
}

// VerifyKeyPackage verifies the hybrid signature on kp.
// Returns ErrInvalidKeyPackage on failure.
func VerifyKeyPackage(kp *KeyPackage) error {
	if kp.CiphersuiteID != CiphersuiteID {
		return fmt.Errorf("tree/mls: unsupported ciphersuite 0x%04X", kp.CiphersuiteID)
	}
	canonical := keyPackageSigningBytes(kp)
	if err := v0crypto.HybridVerify(
		ed25519.PublicKey(kp.Credential.Ed25519Pub),
		kp.Credential.MLDSA65Pub,
		canonical,
		kp.Signature,
	); err != nil {
		return ErrInvalidKeyPackage
	}
	return nil
}

// keyPackageSigningBytes returns the canonical byte form of a KeyPackage for signing.
// Encodes: version(1) || ciphersuite_id(2 BE) || len(init_key)(2 BE) || init_key
//          || len(mlkem_pub)(2 BE) || mlkem_pub
//          || len(peer_id)(2 BE) || peer_id
//          || len(ed25519_pub)(2 BE) || ed25519_pub
//          || len(mldsa65_pub)(2 BE) || mldsa65_pub
// The Signature field is NOT included.
func keyPackageSigningBytes(kp *KeyPackage) []byte {
	var buf []byte

	// version
	buf = append(buf, kp.Version)
	// ciphersuite_id
	var cs [2]byte
	binary.BigEndian.PutUint16(cs[:], kp.CiphersuiteID)
	buf = append(buf, cs[:]...)
	// init_key
	buf = appendLenPrefixed(buf, kp.InitKey)
	// mlkem_pub
	buf = appendLenPrefixed(buf, kp.MLKEMPub)
	// peer_id
	buf = appendLenPrefixed(buf, []byte(kp.Credential.PeerID))
	// ed25519_pub
	buf = appendLenPrefixed(buf, kp.Credential.Ed25519Pub)
	// mldsa65_pub
	buf = appendLenPrefixed(buf, kp.Credential.MLDSA65Pub)

	return buf
}

// appendLenPrefixed appends a 2-byte big-endian length then data to buf.
func appendLenPrefixed(buf, data []byte) []byte {
	var lenBuf [2]byte
	binary.BigEndian.PutUint16(lenBuf[:], uint16(len(data)))
	buf = append(buf, lenBuf[:]...)
	return append(buf, data...)
}
