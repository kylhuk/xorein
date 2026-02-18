package backup

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aether/code_aether/pkg/v12/identity"
	"golang.org/x/crypto/argon2"
)

const (
	EnvelopeVersion = 1

	saltSize  = 16
	nonceSize = 12

	kdfIterations uint32 = 3
	kdfMemory     uint32 = 64 * 1024
	kdfThreads    uint8  = 1
	kdfKeyLength  uint32 = 32
)

// Reason captures deterministic backup and identity restore reasons.
type Reason string

const (
	ReasonUnspecified       Reason = "unspecified"
	ReasonBackupPassword    Reason = "backup-password"
	ReasonBackupCorrupt     Reason = "backup-corrupt"
	ReasonIdentityMismatch  Reason = "identity-mismatch"
	ReasonIdentityDuplicate Reason = "identity-duplicate"
	ReasonBackupOutdated    Reason = "backup-outdated"
)

var (
	ErrBackupPasswordRequired = errors.New("backup password is required")
	ErrBackupCorrupt          = errors.New(string(ReasonBackupCorrupt))
	ErrBackupPassword         = errors.New(string(ReasonBackupPassword))
	ErrIdentityMismatch       = errors.New(string(ReasonIdentityMismatch))
	ErrIdentityDuplicate      = errors.New(string(ReasonIdentityDuplicate))
	ErrBackupOutdated         = errors.New(string(ReasonBackupOutdated))
)

// Payload is the local encrypted backup content.
type Payload struct {
	Identity identity.Record   `json:"identity"`
	Config   map[string]string `json:"config,omitempty"`
}

// Envelope is the versioned backup transport envelope.
type Envelope struct {
	Version          int       `json:"version"`
	BackupID         string    `json:"backup_id"`
	CreatedAt        time.Time `json:"created_at"`
	Salt             string    `json:"salt"`
	Nonce            string    `json:"nonce"`
	Ciphertext       string    `json:"ciphertext"`
	CiphertextSHA256 string    `json:"ciphertext_sha256"`
}

// Export seals a payload using BackupPassword with Argon2id+AEAD.
func Export(payload Payload, backupPassword string, now time.Time) (Envelope, error) {
	if strings.TrimSpace(backupPassword) == "" {
		return Envelope{}, ErrBackupPasswordRequired
	}
	if err := payload.Identity.Validate(); err != nil {
		return Envelope{}, fmt.Errorf("validate identity payload: %w", err)
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}

	plaintext, err := json.Marshal(payload)
	if err != nil {
		return Envelope{}, fmt.Errorf("marshal backup payload: %w", err)
	}

	salt := make([]byte, saltSize)
	if _, err := rand.Read(salt); err != nil {
		return Envelope{}, fmt.Errorf("generate backup salt: %w", err)
	}
	nonce := make([]byte, nonceSize)
	if _, err := rand.Read(nonce); err != nil {
		return Envelope{}, fmt.Errorf("generate backup nonce: %w", err)
	}

	key := deriveKey(backupPassword, salt)
	block, err := aes.NewCipher(key)
	if err != nil {
		return Envelope{}, fmt.Errorf("create backup cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return Envelope{}, fmt.Errorf("create backup gcm: %w", err)
	}
	// #nosec G407 -- nonce is generated per-export via crypto/rand.
	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)
	cipherDigest := sha256.Sum256(ciphertext)

	envelope := Envelope{
		Version:          EnvelopeVersion,
		BackupID:         buildBackupID(payload.Identity.IdentityID, cipherDigest),
		CreatedAt:        now.UTC(),
		Salt:             base64.StdEncoding.EncodeToString(salt),
		Nonce:            base64.StdEncoding.EncodeToString(nonce),
		Ciphertext:       base64.StdEncoding.EncodeToString(ciphertext),
		CiphertextSHA256: hex.EncodeToString(cipherDigest[:]),
	}
	return envelope, nil
}

// MarshalEnvelope serializes a backup envelope.
func MarshalEnvelope(envelope Envelope) ([]byte, error) {
	data, err := json.MarshalIndent(envelope, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal backup envelope: %w", err)
	}
	return data, nil
}

// UnmarshalEnvelope parses a backup envelope from JSON.
func UnmarshalEnvelope(data []byte) (Envelope, error) {
	var envelope Envelope
	if err := json.Unmarshal(data, &envelope); err != nil {
		return Envelope{}, fmt.Errorf("decode backup envelope: %w", err)
	}
	if strings.TrimSpace(envelope.BackupID) == "" {
		return Envelope{}, fmt.Errorf("%w: missing backup_id", ErrBackupCorrupt)
	}
	if envelope.CreatedAt.IsZero() {
		return Envelope{}, fmt.Errorf("%w: missing created_at", ErrBackupCorrupt)
	}
	return envelope, nil
}

// Restore opens and validates a backup envelope.
func Restore(data []byte, backupPassword string, existing *identity.Record) (Payload, error) {
	if strings.TrimSpace(backupPassword) == "" {
		return Payload{}, ErrBackupPasswordRequired
	}
	envelope, err := UnmarshalEnvelope(data)
	if err != nil {
		return Payload{}, fmt.Errorf("%w: %v", ErrBackupCorrupt, err)
	}
	if envelope.Version != EnvelopeVersion {
		return Payload{}, fmt.Errorf("%w: version %d", ErrBackupOutdated, envelope.Version)
	}

	salt, err := base64.StdEncoding.DecodeString(envelope.Salt)
	if err != nil || len(salt) != saltSize {
		return Payload{}, fmt.Errorf("%w: invalid salt", ErrBackupCorrupt)
	}
	nonce, err := base64.StdEncoding.DecodeString(envelope.Nonce)
	if err != nil || len(nonce) != nonceSize {
		return Payload{}, fmt.Errorf("%w: invalid nonce", ErrBackupCorrupt)
	}
	ciphertext, err := base64.StdEncoding.DecodeString(envelope.Ciphertext)
	if err != nil || len(ciphertext) == 0 {
		return Payload{}, fmt.Errorf("%w: invalid ciphertext", ErrBackupCorrupt)
	}
	if strings.TrimSpace(envelope.CiphertextSHA256) == "" {
		return Payload{}, fmt.Errorf("%w: missing ciphertext digest", ErrBackupCorrupt)
	}

	digest := sha256.Sum256(ciphertext)
	if hex.EncodeToString(digest[:]) != strings.ToLower(strings.TrimSpace(envelope.CiphertextSHA256)) {
		return Payload{}, fmt.Errorf("%w: ciphertext digest mismatch", ErrBackupCorrupt)
	}

	key := deriveKey(backupPassword, salt)
	block, err := aes.NewCipher(key)
	if err != nil {
		return Payload{}, fmt.Errorf("create backup cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return Payload{}, fmt.Errorf("create backup gcm: %w", err)
	}
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return Payload{}, fmt.Errorf("%w: authentication failed", ErrBackupPassword)
	}

	var payload Payload
	if err := json.Unmarshal(plaintext, &payload); err != nil {
		return Payload{}, fmt.Errorf("%w: decode payload: %v", ErrBackupCorrupt, err)
	}
	if err := payload.Identity.Validate(); err != nil {
		return Payload{}, fmt.Errorf("%w: invalid identity payload: %v", ErrBackupCorrupt, err)
	}

	if existing != nil {
		if payload.Identity.IdentityID != existing.IdentityID {
			return Payload{}, fmt.Errorf("%w: existing=%s incoming=%s", ErrIdentityMismatch, existing.IdentityID, payload.Identity.IdentityID)
		}
		if payload.Identity.PublicKeyFingerprint != existing.PublicKeyFingerprint {
			return Payload{}, fmt.Errorf("%w: identity fingerprint conflict", ErrIdentityDuplicate)
		}
	}

	return payload, nil
}

// ReasonFromError maps restore errors to deterministic reason taxonomy.
func ReasonFromError(err error) Reason {
	switch {
	case err == nil:
		return ReasonUnspecified
	case errors.Is(err, ErrBackupPassword):
		return ReasonBackupPassword
	case errors.Is(err, ErrBackupCorrupt):
		return ReasonBackupCorrupt
	case errors.Is(err, ErrIdentityMismatch):
		return ReasonIdentityMismatch
	case errors.Is(err, ErrIdentityDuplicate):
		return ReasonIdentityDuplicate
	case errors.Is(err, ErrBackupOutdated):
		return ReasonBackupOutdated
	default:
		return ReasonUnspecified
	}
}

func buildBackupID(identityID string, digest [32]byte) string {
	seed := append([]byte(strings.TrimSpace(identityID)), digest[:]...)
	backupDigest := sha256.Sum256(seed)
	return fmt.Sprintf("BKP-%s", strings.ToUpper(hex.EncodeToString(backupDigest[:8])))
}

func deriveKey(backupPassword string, salt []byte) []byte {
	return argon2.IDKey([]byte(backupPassword), salt, kdfIterations, kdfMemory, kdfThreads, kdfKeyLength)
}
