package identity

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	SeedSize               = ed25519.SeedSize
	CurrentMetadataVersion = 1
)

const defaultKeyReference = "keyring://local/identity"

var (
	ErrStorePathRequired             = errors.New("identity store path is required")
	ErrSeedInvalid                   = errors.New("identity seed must be 32 bytes")
	ErrIdentityNotFound              = errors.New("identity not found")
	ErrIdentityCorrupt               = errors.New("identity state is corrupt")
	ErrIdentityDuplicate             = errors.New("identity duplicate detected")
	ErrIdentityImmutable             = errors.New("identity is immutable")
	ErrIdentityIDRequired            = errors.New("identity id is required")
	ErrPublicKeyFingerprintRequired  = errors.New("public key fingerprint is required")
	ErrKeyReferenceRequired          = errors.New("key reference is required")
	ErrCreatedTimestampRequired      = errors.New("created timestamp is required")
	ErrMetadataVersionInvalid        = errors.New("metadata version invalid")
	ErrPublicKeyFingerprintMalformed = errors.New("public key fingerprint malformed")
)

// BackupStatus tracks whether a local backup has been exported.
type BackupStatus string

const (
	BackupStatusUnknown BackupStatus = "unknown"
	BackupStatusMissing BackupStatus = "missing"
	BackupStatusPresent BackupStatus = "present"
)

// Record stores immutable identity metadata and key references.
type Record struct {
	IdentityID           string       `json:"identity_id"`
	PublicKeyFingerprint string       `json:"public_key_fingerprint"`
	CreatedAt            time.Time    `json:"created_at"`
	MetadataVersion      int          `json:"metadata_version"`
	BackupStatus         BackupStatus `json:"backup_status"`
	KeyReference         string       `json:"key_reference"`
}

type stateFile struct {
	SchemaVersion int    `json:"schema_version"`
	Record        Record `json:"record"`
}

// CreateFromSeed constructs a deterministic immutable identity from a seed.
func CreateFromSeed(seed []byte, now time.Time, keyReference string) (Record, ed25519.PrivateKey, error) {
	if len(seed) != SeedSize {
		return Record{}, nil, ErrSeedInvalid
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	privateKey := ed25519.NewKeyFromSeed(seed)
	publicKey := privateKey.Public().(ed25519.PublicKey)
	record := Record{
		IdentityID:           IdentityIDFromPublicKey(publicKey),
		PublicKeyFingerprint: FingerprintPublicKey(publicKey),
		CreatedAt:            now.UTC(),
		MetadataVersion:      CurrentMetadataVersion,
		BackupStatus:         BackupStatusMissing,
		KeyReference:         strings.TrimSpace(keyReference),
	}
	if record.KeyReference == "" {
		record.KeyReference = defaultKeyReference
	}
	if err := record.Validate(); err != nil {
		return Record{}, nil, err
	}
	return record, privateKey, nil
}

// IdentityIDFromPublicKey computes an immutable globally unique identifier.
func IdentityIDFromPublicKey(publicKey ed25519.PublicKey) string {
	digest := sha256.Sum256(publicKey)
	return fmt.Sprintf("xid-%s", hex.EncodeToString(digest[:16]))
}

// FingerprintPublicKey computes the full SHA-256 fingerprint for key references.
func FingerprintPublicKey(publicKey ed25519.PublicKey) string {
	digest := sha256.Sum256(publicKey)
	return hex.EncodeToString(digest[:])
}

// EnsureImmutable creates and persists identity on first run; later runs must match.
func EnsureImmutable(path string, seed []byte, now time.Time, keyReference string) (Record, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return Record{}, ErrStorePathRequired
	}
	candidate, _, err := CreateFromSeed(seed, now, keyReference)
	if err != nil {
		return Record{}, err
	}

	existing, err := Load(path)
	if err != nil {
		if errors.Is(err, ErrIdentityNotFound) {
			if persistErr := Persist(path, candidate); persistErr != nil {
				return Record{}, persistErr
			}
			return candidate, nil
		}
		return Record{}, err
	}

	if existing.IdentityID != candidate.IdentityID || existing.PublicKeyFingerprint != candidate.PublicKeyFingerprint {
		return Record{}, ErrIdentityDuplicate
	}
	if existing.MetadataVersion != candidate.MetadataVersion {
		return Record{}, ErrIdentityImmutable
	}
	return existing, nil
}

// Persist writes record metadata to a local JSON state file.
func Persist(path string, record Record) error {
	path = strings.TrimSpace(path)
	if path == "" {
		return ErrStorePathRequired
	}
	if err := record.Validate(); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create identity state dir: %w", err)
	}
	data, err := json.MarshalIndent(stateFile{SchemaVersion: CurrentMetadataVersion, Record: record}, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal identity state: %w", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write identity state: %w", err)
	}
	return nil
}

// Load reads and validates an identity record from local JSON state.
func Load(path string) (Record, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return Record{}, ErrStorePathRequired
	}
	// #nosec G304 -- path is explicit local identity state path selected by runtime config.
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Record{}, ErrIdentityNotFound
		}
		return Record{}, fmt.Errorf("read identity state: %w", err)
	}

	var state stateFile
	if err := json.Unmarshal(data, &state); err != nil {
		return Record{}, fmt.Errorf("%w: decode json: %v", ErrIdentityCorrupt, err)
	}
	if state.SchemaVersion > CurrentMetadataVersion {
		return Record{}, fmt.Errorf("%w: schema version %d unsupported", ErrIdentityCorrupt, state.SchemaVersion)
	}
	if err := state.Record.Validate(); err != nil {
		return Record{}, fmt.Errorf("%w: %v", ErrIdentityCorrupt, err)
	}
	return state.Record, nil
}

// Validate ensures the identity record is complete and deterministic.
func (r Record) Validate() error {
	if strings.TrimSpace(r.IdentityID) == "" {
		return ErrIdentityIDRequired
	}
	fingerprint := strings.TrimSpace(r.PublicKeyFingerprint)
	if fingerprint == "" {
		return ErrPublicKeyFingerprintRequired
	}
	if _, err := hex.DecodeString(fingerprint); err != nil {
		return fmt.Errorf("%w: %v", ErrPublicKeyFingerprintMalformed, err)
	}
	if len(fingerprint) != 64 {
		return fmt.Errorf("%w: expected 64 hex chars", ErrPublicKeyFingerprintMalformed)
	}
	if r.CreatedAt.IsZero() {
		return ErrCreatedTimestampRequired
	}
	if r.MetadataVersion < 1 {
		return ErrMetadataVersionInvalid
	}
	if strings.TrimSpace(r.KeyReference) == "" {
		return ErrKeyReferenceRequired
	}
	return nil
}
