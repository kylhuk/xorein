// Package seal implements Seal mode: hybrid X3DH + Double Ratchet.
// Source: docs/spec/v0.1/11-mode-seal.md
package seal

import (
	"bytes"
	"crypto/ed25519"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	v0crypto "github.com/aether/code_aether/pkg/v0_1/crypto"
)

// ErrBundleExpired is returned when a prekey bundle's expires_at has passed.
var ErrBundleExpired = errors.New("seal: prekey bundle expired")

// ErrBundleSignatureInvalid is returned when a bundle's signature verification fails.
var ErrBundleSignatureInvalid = errors.New("seal: bundle signature invalid")

// ErrNoOPKAvailable is returned when a bundle has no one-time prekeys left.
var ErrNoOPKAvailable = errors.New("seal: no one-time prekey available")

// PrekeyBundle is the prekey bundle published by each node per spec 11 §1.1.
// All byte fields are raw bytes (not base64).
type PrekeyBundle struct {
	PeerID               string      `json:"peer_id"`
	IdentityKeyEd25519   []byte      `json:"identity_key_ed25519"`   // 32 B
	IdentityKeyMLDSA65   []byte      `json:"identity_key_ml_dsa_65"` // 1952 B
	SignedPrekeyX25519   []byte      `json:"signed_prekey_x25519"`   // 32 B
	SignedPrekeySignature []byte     `json:"signed_prekey_signature"` // 3373 B hybrid
	OneTimePrekeys       [][]byte    `json:"one_time_prekeys_x25519"` // each 32 B
	MLKEM768PK           []byte      `json:"ml_kem_768_pk"`           // 1184 B
	MLKEM768PKSignature  []byte      `json:"ml_kem_768_pk_signature"` // 3373 B hybrid
	PublishedAt          int64       `json:"published_at"`
	ExpiresAt            int64       `json:"expires_at"`
	BundleSignature      []byte      `json:"bundle_signature"` // 3373 B hybrid
}

// PrekeyPrivate holds the secret key material corresponding to a PrekeyBundle.
type PrekeyPrivate struct {
	SPKPriv     [32]byte   // signed prekey X25519 private key
	OPKPrivs    [][32]byte // one-time prekey X25519 private keys (same order as bundle)
	MLKEM768SK  []byte     // ML-KEM-768 decapsulation key
}

// InitialMessage is the data the initiator sends with the first message.
type InitialMessage struct {
	EKPub    [32]byte // initiator's ephemeral X25519 public key
	CTMLKEM  []byte   // ML-KEM-768 ciphertext (1088 B)
	OPKIndex int      // index of OPK used; -1 if no OPK
}

// BuildBundle generates a fresh prekey bundle and its private key material.
// opkCount is the number of one-time prekeys to generate (1–100).
func BuildBundle(
	peerID string,
	idEdPriv ed25519.PrivateKey,
	idMLDSAPriv []byte,
	opkCount int,
) (*PrekeyBundle, *PrekeyPrivate, error) {
	if opkCount < 1 {
		opkCount = 1
	}
	if opkCount > 100 {
		opkCount = 100
	}

	// Generate signed prekey (SPK).
	spkPriv, spkPub, err := v0crypto.GenerateX25519Keypair()
	if err != nil {
		return nil, nil, fmt.Errorf("bundle: generate SPK: %w", err)
	}

	// Sign SPK: spec 11 §1.2: sign(label || spk_pub).
	spkCanonical := append([]byte(v0crypto.LabelSealSPKSign), spkPub[:]...)
	spkSig, err := v0crypto.HybridSign(idEdPriv, idMLDSAPriv, spkCanonical)
	if err != nil {
		return nil, nil, fmt.Errorf("bundle: sign SPK: %w", err)
	}

	// Generate one-time prekeys.
	opkPubs := make([][]byte, opkCount)
	opkPrivs := make([][32]byte, opkCount)
	for i := range opkCount {
		priv, pub, err := v0crypto.GenerateX25519Keypair()
		if err != nil {
			return nil, nil, fmt.Errorf("bundle: generate OPK[%d]: %w", i, err)
		}
		opkPrivs[i] = priv
		opkPubs[i] = pub[:]
	}

	// Generate ML-KEM-768 keypair.
	mlkemPK, mlkemSK, err := v0crypto.GenerateMLKEM768Keypair()
	if err != nil {
		return nil, nil, fmt.Errorf("bundle: generate ML-KEM: %w", err)
	}

	// Sign ML-KEM public key.
	mlkemLabel := []byte("xorein/seal/v1/mlkem-pk-sign")
	mlkemCanonical := append(mlkemLabel, mlkemPK...)
	mlkemSig, err := v0crypto.HybridSign(idEdPriv, idMLDSAPriv, mlkemCanonical)
	if err != nil {
		return nil, nil, fmt.Errorf("bundle: sign ML-KEM PK: %w", err)
	}

	// Derive ML-DSA public key from private key.
	mldsaPub, err := v0crypto.MLDSA65PublicKeyFromPrivate(idMLDSAPriv)
	if err != nil {
		return nil, nil, fmt.Errorf("bundle: derive ML-DSA pub: %w", err)
	}

	now := time.Now()
	bundle := &PrekeyBundle{
		PeerID:               peerID,
		IdentityKeyEd25519:   []byte(idEdPriv.Public().(ed25519.PublicKey)),
		IdentityKeyMLDSA65:   mldsaPub,
		SignedPrekeyX25519:   spkPub[:],
		SignedPrekeySignature: spkSig,
		OneTimePrekeys:        opkPubs,
		MLKEM768PK:            mlkemPK,
		MLKEM768PKSignature:   mlkemSig,
		PublishedAt:           now.UnixMilli(),
		ExpiresAt:             now.Add(7 * 24 * time.Hour).UnixMilli(),
	}

	// Sign entire bundle (without BundleSignature field set).
	bundleCanon, err := canonicalBundleBytes(bundle)
	if err != nil {
		return nil, nil, fmt.Errorf("bundle: canonical: %w", err)
	}
	bundleSig, err := v0crypto.HybridSign(idEdPriv, idMLDSAPriv, bundleCanon)
	if err != nil {
		return nil, nil, fmt.Errorf("bundle: sign bundle: %w", err)
	}
	bundle.BundleSignature = bundleSig

	priv := &PrekeyPrivate{
		SPKPriv:    spkPriv,
		OPKPrivs:   opkPrivs,
		MLKEM768SK: mlkemSK,
	}
	return bundle, priv, nil
}

// VerifyBundle verifies all signatures in a prekey bundle.
func VerifyBundle(b *PrekeyBundle, now time.Time) error {
	if now.UnixMilli() > b.ExpiresAt {
		return ErrBundleExpired
	}
	if b.ExpiresAt-b.PublishedAt > int64(7*24*time.Hour/time.Millisecond) {
		return ErrBundleExpired
	}

	edPub := ed25519.PublicKey(b.IdentityKeyEd25519)
	mldsaPub := b.IdentityKeyMLDSA65

	// Verify SPK signature.
	spkCanonical := append([]byte(v0crypto.LabelSealSPKSign), b.SignedPrekeyX25519...)
	if err := v0crypto.HybridVerify(edPub, mldsaPub, spkCanonical, b.SignedPrekeySignature); err != nil {
		return fmt.Errorf("%w: SPK signature", ErrBundleSignatureInvalid)
	}

	// Verify ML-KEM PK signature.
	mlkemLabel := []byte("xorein/seal/v1/mlkem-pk-sign")
	mlkemCanonical := append(mlkemLabel, b.MLKEM768PK...)
	if err := v0crypto.HybridVerify(edPub, mldsaPub, mlkemCanonical, b.MLKEM768PKSignature); err != nil {
		return fmt.Errorf("%w: ML-KEM PK signature", ErrBundleSignatureInvalid)
	}

	// Verify bundle signature (over bundle without BundleSignature field).
	saved := b.BundleSignature
	b.BundleSignature = nil
	bundleCanon, err := canonicalBundleBytes(b)
	b.BundleSignature = saved
	if err != nil {
		return fmt.Errorf("%w: canonical bytes: %v", ErrBundleSignatureInvalid, err)
	}
	if err := v0crypto.HybridVerify(edPub, mldsaPub, bundleCanon, saved); err != nil {
		return fmt.Errorf("%w: bundle signature", ErrBundleSignatureInvalid)
	}

	return nil
}

// Initiate runs the initiator side of hybrid X3DH.
// Returns the initial message data (to send with first message) and a RatchetState.
func Initiate(
	bundle *PrekeyBundle,
	ourEdPriv ed25519.PrivateKey,
) (*InitialMessage, *RatchetState, error) {
	// Convert initiator identity key to X25519.
	ikPriv, err := v0crypto.Ed25519PrivateToX25519(ourEdPriv)
	if err != nil {
		return nil, nil, fmt.Errorf("x3dh init: convert IK: %w", err)
	}

	// Convert responder identity key to X25519.
	ikrPub, err := v0crypto.Ed25519PublicToX25519(ed25519.PublicKey(bundle.IdentityKeyEd25519))
	if err != nil {
		return nil, nil, fmt.Errorf("x3dh init: convert IKr: %w", err)
	}

	// Parse responder SPK.
	var spkPub [32]byte
	copy(spkPub[:], bundle.SignedPrekeyX25519)

	// Generate ephemeral key.
	ekPriv, ekPub, err := v0crypto.GenerateX25519Keypair()
	if err != nil {
		return nil, nil, fmt.Errorf("x3dh init: generate EK: %w", err)
	}

	// DH1: initiator IK × responder SPK.
	dh1, err := v0crypto.X25519DH(ikPriv, spkPub)
	if err != nil {
		return nil, nil, fmt.Errorf("x3dh init: DH1: %w", err)
	}
	// DH2: initiator EK × responder IK.
	dh2, err := v0crypto.X25519DH(ekPriv, ikrPub)
	if err != nil {
		return nil, nil, fmt.Errorf("x3dh init: DH2: %w", err)
	}
	// DH3: initiator EK × responder SPK.
	dh3, err := v0crypto.X25519DH(ekPriv, spkPub)
	if err != nil {
		return nil, nil, fmt.Errorf("x3dh init: DH3: %w", err)
	}

	x3dhSecret := append(dh1[:], append(dh2[:], dh3[:]...)...)
	opkIndex := -1

	// DH4: initiator EK × responder OPK (if available).
	if len(bundle.OneTimePrekeys) > 0 {
		var opkPub [32]byte
		copy(opkPub[:], bundle.OneTimePrekeys[0])
		dh4, err := v0crypto.X25519DH(ekPriv, opkPub)
		if err != nil {
			return nil, nil, fmt.Errorf("x3dh init: DH4: %w", err)
		}
		x3dhSecret = append(x3dhSecret, dh4[:]...)
		opkIndex = 0
	}

	// ML-KEM-768 encapsulation.
	ctMLKEM, ssMLKEM, err := v0crypto.MLKEM768Encapsulate(bundle.MLKEM768PK)
	if err != nil {
		return nil, nil, fmt.Errorf("x3dh init: ML-KEM encap: %w", err)
	}

	// Hybrid combine.
	hybridMaster, err := v0crypto.CombineKEM(x3dhSecret, ssMLKEM)
	if err != nil {
		return nil, nil, fmt.Errorf("x3dh init: combine KEM: %w", err)
	}

	// Root + chain key derivation.
	rs, err := initRatchetFromMaster(hybridMaster, ekPriv, ekPub, spkPub, true)
	if err != nil {
		return nil, nil, fmt.Errorf("x3dh init: ratchet init: %w", err)
	}

	im := &InitialMessage{
		EKPub:    ekPub,
		CTMLKEM:  ctMLKEM,
		OPKIndex: opkIndex,
	}
	return im, rs, nil
}

// Respond runs the responder side of hybrid X3DH.
// theirEdPub is the initiator's Ed25519 identity public key.
func Respond(
	im *InitialMessage,
	priv *PrekeyPrivate,
	bundle *PrekeyBundle,
	theirEdPub ed25519.PublicKey,
) (*RatchetState, error) {
	// Convert initiator identity key to X25519.
	ikPub, err := v0crypto.Ed25519PublicToX25519(theirEdPub)
	if err != nil {
		return nil, fmt.Errorf("x3dh respond: convert IK: %w", err)
	}

	// Parse SPK.
	var spkPub [32]byte
	copy(spkPub[:], bundle.SignedPrekeyX25519)

	// Convert responder identity key to X25519 (for DH1 and DH2).
	// DH1: initiator IK × responder SPK → X25519(SPK_priv, IK_pub).
	dh1, err := v0crypto.X25519DH(priv.SPKPriv, ikPub)
	if err != nil {
		return nil, fmt.Errorf("x3dh respond: DH1: %w", err)
	}
	// DH2: initiator EK × responder IK → X25519(IK_priv, EK_pub).
	// We need our own IK X25519 private key.
	// Actually we can compute X25519(IK_priv, EK_pub) = X25519(EK_priv, IK_pub) for DH2.
	// But we only have SPK priv here, not IK priv. We need the responder's IK X25519 priv.
	// The responder holds their Ed25519 private key — it's not in PrekeyPrivate.
	// This means Respond() must take the responder's idEdPriv as well.
	_ = dh1

	return nil, fmt.Errorf("x3dh respond: not implemented (needs idEdPriv)")
}

// RespondFull runs the responder side of hybrid X3DH with access to identity keys.
func RespondFull(
	im *InitialMessage,
	priv *PrekeyPrivate,
	bundle *PrekeyBundle,
	ourEdPriv ed25519.PrivateKey,
	theirEdPub ed25519.PublicKey,
) (*RatchetState, error) {
	// Convert our identity key to X25519.
	ourIKPriv, err := v0crypto.Ed25519PrivateToX25519(ourEdPriv)
	if err != nil {
		return nil, fmt.Errorf("x3dh respond: convert our IK: %w", err)
	}

	// Convert initiator identity key to X25519.
	theirIKPub, err := v0crypto.Ed25519PublicToX25519(theirEdPub)
	if err != nil {
		return nil, fmt.Errorf("x3dh respond: convert their IK: %w", err)
	}

	// Parse SPK.
	var spkPub [32]byte
	copy(spkPub[:], bundle.SignedPrekeyX25519)

	// DH1: X25519(SPK_priv, their_IK_pub).
	dh1, err := v0crypto.X25519DH(priv.SPKPriv, theirIKPub)
	if err != nil {
		return nil, fmt.Errorf("x3dh respond: DH1: %w", err)
	}
	// DH2: X25519(our_IK_priv, their_EK_pub).
	dh2, err := v0crypto.X25519DH(ourIKPriv, im.EKPub)
	if err != nil {
		return nil, fmt.Errorf("x3dh respond: DH2: %w", err)
	}
	// DH3: X25519(SPK_priv, their_EK_pub).
	dh3, err := v0crypto.X25519DH(priv.SPKPriv, im.EKPub)
	if err != nil {
		return nil, fmt.Errorf("x3dh respond: DH3: %w", err)
	}

	x3dhSecret := append(dh1[:], append(dh2[:], dh3[:]...)...)

	// DH4: X25519(OPK_priv, their_EK_pub).
	if im.OPKIndex >= 0 && im.OPKIndex < len(priv.OPKPrivs) {
		dh4, err := v0crypto.X25519DH(priv.OPKPrivs[im.OPKIndex], im.EKPub)
		if err != nil {
			return nil, fmt.Errorf("x3dh respond: DH4: %w", err)
		}
		x3dhSecret = append(x3dhSecret, dh4[:]...)
	}

	// ML-KEM-768 decapsulation.
	ssMLKEM, err := v0crypto.MLKEM768Decapsulate(priv.MLKEM768SK, im.CTMLKEM)
	if err != nil {
		return nil, fmt.Errorf("x3dh respond: ML-KEM decap: %w", err)
	}

	// Hybrid combine.
	hybridMaster, err := v0crypto.CombineKEM(x3dhSecret, ssMLKEM)
	if err != nil {
		return nil, fmt.Errorf("x3dh respond: combine KEM: %w", err)
	}

	// Responder uses SPK as initial ratchet key; remote is EK pub.
	rs, err := initRatchetFromMaster(hybridMaster, priv.SPKPriv, spkPub, im.EKPub, false)
	if err != nil {
		return nil, fmt.Errorf("x3dh respond: ratchet init: %w", err)
	}

	return rs, nil
}

// initRatchetFromMaster derives root+chain keys from hybridMaster and builds initial RatchetState.
// isInitiator=true → SendChainKey=chain_key; isInitiator=false → RecvChainKey=chain_key.
func initRatchetFromMaster(hybridMaster [32]byte, localRatchetPriv, localRatchetPub, remoteRatchetPub [32]byte, isInitiator bool) (*RatchetState, error) {
	okm, err := v0crypto.DeriveKey(hybridMaster[:], nil, v0crypto.LabelSealRootKey, 64)
	if err != nil {
		return nil, fmt.Errorf("derive root key: %w", err)
	}
	var rootKey, chainKey [32]byte
	copy(rootKey[:], okm[:32])
	copy(chainKey[:], okm[32:])

	rs := &RatchetState{
		RootKey:          rootKey,
		SendRatchetPriv:  localRatchetPriv,
		SendRatchetPub:   localRatchetPub,
		RemoteRatchetPub: remoteRatchetPub,
		SkipList:         make(map[SkipKey][32]byte),
	}
	if isInitiator {
		rs.SendChainKey = chainKey
	} else {
		rs.RecvChainKey = chainKey
	}
	return rs, nil
}

// canonicalBundleBytes returns a deterministic JSON encoding of the bundle
// (without BundleSignature) for signing.
func canonicalBundleBytes(b *PrekeyBundle) ([]byte, error) {
	type bundleForSig struct {
		PeerID               string   `json:"peer_id"`
		IdentityKeyEd25519   []byte   `json:"identity_key_ed25519"`
		IdentityKeyMLDSA65   []byte   `json:"identity_key_ml_dsa_65"`
		SignedPrekeyX25519   []byte   `json:"signed_prekey_x25519"`
		SignedPrekeySignature []byte  `json:"signed_prekey_signature"`
		OneTimePrekeys       [][]byte `json:"one_time_prekeys_x25519"`
		MLKEM768PK           []byte   `json:"ml_kem_768_pk"`
		MLKEM768PKSignature  []byte   `json:"ml_kem_768_pk_signature"`
		PublishedAt          int64    `json:"published_at"`
		ExpiresAt            int64    `json:"expires_at"`
	}
	s := bundleForSig{
		PeerID:               b.PeerID,
		IdentityKeyEd25519:   b.IdentityKeyEd25519,
		IdentityKeyMLDSA65:   b.IdentityKeyMLDSA65,
		SignedPrekeyX25519:   b.SignedPrekeyX25519,
		SignedPrekeySignature: b.SignedPrekeySignature,
		OneTimePrekeys:        b.OneTimePrekeys,
		MLKEM768PK:            b.MLKEM768PK,
		MLKEM768PKSignature:   b.MLKEM768PKSignature,
		PublishedAt:           b.PublishedAt,
		ExpiresAt:             b.ExpiresAt,
	}
	data, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	// Prefix with a domain label.
	prefix := []byte("xorein/seal/v1/bundle-sign\x00")
	return append(prefix, data...), nil
}

// BundlesEqual compares two bundles by their wire fields (not private material).
func BundlesEqual(a, b *PrekeyBundle) bool {
	if a.PeerID != b.PeerID {
		return false
	}
	return bytes.Equal(a.BundleSignature, b.BundleSignature)
}
