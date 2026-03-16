package node

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

func GenerateIdentity(displayName string) (Identity, error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return Identity{}, fmt.Errorf("generate identity: %w", err)
	}
	createdAt := time.Now().UTC()
	id := Identity{
		ID:         randomID("identity"),
		PeerID:     derivePeerID(pub),
		PublicKey:  base64.RawURLEncoding.EncodeToString(pub),
		PrivateKey: base64.RawURLEncoding.EncodeToString(priv),
		CreatedAt:  createdAt,
		Profile:    Profile{DisplayName: strings.TrimSpace(displayName)},
	}
	if id.Profile.DisplayName == "" {
		id.Profile.DisplayName = id.PeerID[:12]
	}
	return id, nil
}

func RestoreIdentity(raw []byte) (Identity, error) {
	var identity Identity
	if err := json.Unmarshal(raw, &identity); err != nil {
		return Identity{}, fmt.Errorf("decode identity: %w", err)
	}
	if err := identity.Validate(); err != nil {
		return Identity{}, err
	}
	return identity, nil
}

func (i Identity) Validate() error {
	if i.PeerID == "" || i.PublicKey == "" || i.PrivateKey == "" {
		return errors.New("identity is incomplete")
	}
	pub, err := i.publicKeyBytes()
	if err != nil {
		return fmt.Errorf("identity public key: %w", err)
	}
	if got := derivePeerID(pub); got != i.PeerID {
		return fmt.Errorf("identity peer id mismatch: got %s want %s", got, i.PeerID)
	}
	priv, err := i.privateKeyBytes()
	if err != nil {
		return fmt.Errorf("identity private key: %w", err)
	}
	if len(priv) != ed25519.PrivateKeySize {
		return errors.New("identity private key size invalid")
	}
	return nil
}

func (i Identity) Backup() ([]byte, error) {
	if err := i.Validate(); err != nil {
		return nil, err
	}
	return json.MarshalIndent(i, "", "  ")
}

func (i Identity) PublicPeer() PeerRecord {
	return PeerRecord{
		PeerID:    i.PeerID,
		PublicKey: i.PublicKey,
	}
}

func (i Identity) publicKeyBytes() (ed25519.PublicKey, error) {
	pub, err := base64.RawURLEncoding.DecodeString(i.PublicKey)
	if err != nil {
		return nil, err
	}
	if len(pub) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("invalid public key size %d", len(pub))
	}
	return ed25519.PublicKey(pub), nil
}

func (i Identity) privateKeyBytes() (ed25519.PrivateKey, error) {
	priv, err := base64.RawURLEncoding.DecodeString(i.PrivateKey)
	if err != nil {
		return nil, err
	}
	if len(priv) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("invalid private key size %d", len(priv))
	}
	return ed25519.PrivateKey(priv), nil
}

func (i Identity) SignCanonical(payload []byte) (string, error) {
	priv, err := i.privateKeyBytes()
	if err != nil {
		return "", err
	}
	sig := ed25519.Sign(priv, payload)
	return base64.RawURLEncoding.EncodeToString(sig), nil
}

func VerifyCanonical(publicKey string, payload []byte, signature string) error {
	pubRaw, err := base64.RawURLEncoding.DecodeString(publicKey)
	if err != nil {
		return fmt.Errorf("decode public key: %w", err)
	}
	sigRaw, err := base64.RawURLEncoding.DecodeString(signature)
	if err != nil {
		return fmt.Errorf("decode signature: %w", err)
	}
	if !ed25519.Verify(ed25519.PublicKey(pubRaw), payload, sigRaw) {
		return errors.New("signature verification failed")
	}
	return nil
}

func derivePeerID(pub ed25519.PublicKey) string {
	sum := sha256.Sum256(pub)
	return hex.EncodeToString(sum[:16])
}

func randomID(prefix string) string {
	var buf [8]byte
	if _, err := rand.Read(buf[:]); err != nil {
		sum := sha256.Sum256([]byte(time.Now().UTC().Format(time.RFC3339Nano)))
		return fmt.Sprintf("%s-%s", prefix, hex.EncodeToString(sum[:6]))
	}
	return fmt.Sprintf("%s-%s", prefix, hex.EncodeToString(buf[:]))
}
