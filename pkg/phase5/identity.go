package phase5

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"strings"
)

const (
	seedSize       = ed25519.SeedSize
	privateKeySize = ed25519.PrivateKeySize
	saltSize       = 16
	nonceSize      = 12
	mnemonicWords  = seedSize + 1
)

var (
	errSeedLength      = fmt.Errorf("phase5: seed must be %d bytes", seedSize)
	errInvalidEnvelope = errors.New("phase5: invalid storage envelope")
	errMnemonicLength  = fmt.Errorf("phase5: mnemonic must contain %d words", mnemonicWords)
	errMnemonicCRC     = errors.New("phase5: mnemonic checksum mismatch")
)

var (
	mnemonicPrefix = [16]string{"al", "be", "co", "di", "el", "fi", "go", "ha", "io", "ju", "ka", "lo", "mi", "nu", "or", "pa"}
	mnemonicSuffix = [16]string{"bar", "cis", "den", "far", "gan", "hex", "jin", "kel", "lun", "mon", "nar", "pos", "ron", "sun", "tor", "vek"}
	mnemonicLookup = buildMnemonicLookup()
)

// DeriveKeyFromSeed converts a deterministic seed into the Ed25519 key pair that
// forms the canonical identity. Seed validation ensures deterministic format
// (32 bytes only) so repeated imports always produce the same key material.
func DeriveKeyFromSeed(seed []byte) (ed25519.PublicKey, ed25519.PrivateKey, error) {
	if len(seed) != seedSize {
		return nil, nil, errSeedLength
	}
	private := ed25519.NewKeyFromSeed(seed)
	return private.Public().(ed25519.PublicKey), private, nil
}

// EncodeSeed writes the canonical hex representation for storage or CLI flags.
func EncodeSeed(seed []byte) string {
	return hex.EncodeToString(seed)
}

// ParseSeed decodes and validates a deterministic seed hex string.
func ParseSeed(value string) ([]byte, error) {
	data, err := hex.DecodeString(value)
	if err != nil {
		return nil, fmt.Errorf("phase5: invalid seed format: %w", err)
	}
	if len(data) != seedSize {
		return nil, errSeedLength
	}
	return data, nil
}

// GenerateMnemonic converts a seed to the deterministic recovery phrase format.
// The phrase contains 33 words: 32 data words + 1 checksum word.
func GenerateMnemonic(seed []byte) (string, error) {
	if len(seed) != seedSize {
		return "", errSeedLength
	}
	words := make([]string, 0, mnemonicWords)
	for _, b := range seed {
		words = append(words, mnemonicTokenForByte(b))
	}
	words = append(words, mnemonicTokenForByte(mnemonicChecksum(seed)))
	return strings.Join(words, " "), nil
}

// RecoverSeedFromMnemonic validates and decodes the mnemonic phrase back to the
// canonical 32-byte seed. Errors are deterministic and do not include sensitive
// seed or phrase material.
func RecoverSeedFromMnemonic(mnemonic string) ([]byte, error) {
	parts := strings.Fields(strings.TrimSpace(mnemonic))
	if len(parts) != mnemonicWords {
		return nil, errMnemonicLength
	}

	seed := make([]byte, seedSize)
	for i := 0; i < seedSize; i++ {
		word := strings.ToLower(parts[i])
		value, ok := mnemonicLookup[word]
		if !ok {
			return nil, fmt.Errorf("phase5: mnemonic contains unknown word at position %d", i+1)
		}
		seed[i] = value
	}

	providedChecksum := strings.ToLower(parts[seedSize])
	expectedChecksum := mnemonicTokenForByte(mnemonicChecksum(seed))
	if providedChecksum != expectedChecksum {
		return nil, errMnemonicCRC
	}

	return seed, nil
}

// RecoverKeyFromMnemonic restores the identity keypair from the mnemonic
// recovery phrase.
func RecoverKeyFromMnemonic(mnemonic string) (ed25519.PublicKey, ed25519.PrivateKey, error) {
	seed, err := RecoverSeedFromMnemonic(mnemonic)
	if err != nil {
		return nil, nil, err
	}
	return DeriveKeyFromSeed(seed)
}

// StorageEnvelope protects the raw private key using AES-GCM. The envelope
// structure ensures replay-safe salts/nonces and explicit corruption recovery
// paths for AC compliance (deterministic init + secure storage). The private
// key itself is never revealed in logs or error text.
type StorageEnvelope struct {
	Salt       []byte
	Nonce      []byte
	Ciphertext []byte
}

// SealPrivateKey encrypts the provided private key with a passphrase, returning
// the secure envelope that can be persisted. Random salt/nonce ensure each
// envelope is unique even for deterministic identities.
func SealPrivateKey(private ed25519.PrivateKey, passphrase string) (StorageEnvelope, error) {
	if len(private) != privateKeySize {
		return StorageEnvelope{}, fmt.Errorf("phase5: expected %d byte private key", privateKeySize)
	}
	salt := make([]byte, saltSize)
	if _, err := rand.Read(salt); err != nil {
		return StorageEnvelope{}, fmt.Errorf("phase5: salt generation failed: %w", err)
	}
	key := deriveAESKey(passphrase, salt)

	block, err := aes.NewCipher(key)
	if err != nil {
		return StorageEnvelope{}, fmt.Errorf("phase5: encryption setup failed: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return StorageEnvelope{}, fmt.Errorf("phase5: encryption setup failed: %w", err)
	}

	nonce := make([]byte, nonceSize)
	if _, err := rand.Read(nonce); err != nil {
		return StorageEnvelope{}, fmt.Errorf("phase5: nonce generation failed: %w", err)
	}

	ciphertext := gcm.Seal(nil, nonce, private, nil)
	return StorageEnvelope{Salt: salt, Nonce: nonce, Ciphertext: ciphertext}, nil
}

// OpenPrivateKey decrypts the envelope using the passphrase, enforcing corruption
// detection with explicit errors without leaking private key bytes.
func OpenPrivateKey(env StorageEnvelope, passphrase string) (ed25519.PrivateKey, error) {
	if err := env.Validate(); err != nil {
		return nil, err
	}
	key := deriveAESKey(passphrase, env.Salt)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("phase5: decryption setup failed: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("phase5: decryption setup failed: %w", err)
	}
	privBytes, err := gcm.Open(nil, env.Nonce, env.Ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("phase5: envelope corrupted or passphrase invalid: %w", err)
	}
	if len(privBytes) != privateKeySize {
		return nil, errInvalidEnvelope
	}
	private := make(ed25519.PrivateKey, privateKeySize)
	copy(private, privBytes)
	return private, nil
}

// Validate ensures the envelope contains the expected byte lengths.
func (env StorageEnvelope) Validate() error {
	if len(env.Salt) != saltSize || len(env.Nonce) != nonceSize || len(env.Ciphertext) == 0 {
		return errInvalidEnvelope
	}
	return nil
}

// Marshal serializes the envelope to bytes for deterministic storage.
// Format: [saltLen][salt][nonceLen][nonce][cipherLen][ciphertext] (each len uint32 BE).
func (env StorageEnvelope) Marshal() ([]byte, error) {
	if err := env.Validate(); err != nil {
		return nil, err
	}
	if !storageEnvelopeMarshalLengthsWithinBounds(uint64(len(env.Salt)), uint64(len(env.Nonce)), uint64(len(env.Ciphertext))) {
		return nil, errInvalidEnvelope
	}
	saltLen := uint32(len(env.Salt))         // #nosec G115 -- bounded by MaxUint32 check above
	nonceLen := uint32(len(env.Nonce))       // #nosec G115 -- bounded by MaxUint32 check above
	cipherLen := uint32(len(env.Ciphertext)) // #nosec G115 -- bounded by MaxUint32 check above
	buf := make([]byte, 4+len(env.Salt)+4+len(env.Nonce)+4+len(env.Ciphertext))
	offset := 0
	binary.BigEndian.PutUint32(buf[offset:], saltLen)
	offset += 4
	copy(buf[offset:], env.Salt)
	offset += len(env.Salt)
	binary.BigEndian.PutUint32(buf[offset:], nonceLen)
	offset += 4
	copy(buf[offset:], env.Nonce)
	offset += len(env.Nonce)
	binary.BigEndian.PutUint32(buf[offset:], cipherLen)
	offset += 4
	copy(buf[offset:], env.Ciphertext)
	return buf, nil
}

func storageEnvelopeLengthWithinBounds(length uint64) bool {
	return length <= math.MaxUint32
}

func storageEnvelopeMarshalLengthsWithinBounds(saltLength uint64, nonceLength uint64, ciphertextLength uint64) bool {
	return storageEnvelopeLengthWithinBounds(saltLength) &&
		storageEnvelopeLengthWithinBounds(nonceLength) &&
		storageEnvelopeLengthWithinBounds(ciphertextLength)
}

// Unmarshal loads an envelope from bytes while validating corruption.
func UnmarshalEnvelope(data []byte) (StorageEnvelope, error) {
	env := StorageEnvelope{}
	offset := 0
	if len(data) < 12 {
		return env, errInvalidEnvelope
	}
	if offset+4 > len(data) {
		return env, errInvalidEnvelope
	}
	saltLen := int(binary.BigEndian.Uint32(data[offset:]))
	offset += 4
	if offset+saltLen > len(data) {
		return env, errInvalidEnvelope
	}
	env.Salt = append([]byte{}, data[offset:offset+saltLen]...)
	offset += saltLen
	if offset+4 > len(data) {
		return env, errInvalidEnvelope
	}
	nonceLen := int(binary.BigEndian.Uint32(data[offset:]))
	offset += 4
	if offset+nonceLen > len(data) {
		return env, errInvalidEnvelope
	}
	env.Nonce = append([]byte{}, data[offset:offset+nonceLen]...)
	offset += nonceLen
	if offset+4 > len(data) {
		return env, errInvalidEnvelope
	}
	cipherLen := int(binary.BigEndian.Uint32(data[offset:]))
	offset += 4
	if offset+cipherLen > len(data) {
		return env, errInvalidEnvelope
	}
	env.Ciphertext = append([]byte{}, data[offset:offset+cipherLen]...)
	if err := env.Validate(); err != nil {
		return StorageEnvelope{}, err
	}
	return env, nil
}

func deriveAESKey(passphrase string, salt []byte) []byte {
	hasher := sha256.New()
	hasher.Write(salt)
	hasher.Write([]byte(passphrase))
	hasher.Write([]byte{0x01})
	return hasher.Sum(nil)
}

func buildMnemonicLookup() map[string]byte {
	lookup := make(map[string]byte, 256)
	for i := 0; i <= 0xFF; i++ {
		value := byte(i)
		lookup[mnemonicTokenForByte(value)] = value
	}
	return lookup
}

func mnemonicTokenForByte(value byte) string {
	prefix := mnemonicPrefix[value>>4]
	suffix := mnemonicSuffix[value&0x0F]
	return prefix + suffix
}

func mnemonicChecksum(seed []byte) byte {
	digest := sha256.Sum256(seed)
	return digest[0]
}

const (
	identitySchemaVersion = 1
	profileSchemaVersion  = 1
)

// SchemaVersionSet captures local DB schema versions for identity/profile data.
// Version values are forward-only and monotonic by policy.
type SchemaVersionSet struct {
	Identity int
	Profile  int
}

// MigrationStep represents one forward-only migration transition.
type MigrationStep struct {
	FromVersion int
	ToVersion   int
	Description string
}

// MigrationPlan defines migration policy for local identity/profile storage.
type MigrationPlan struct {
	IdentityVersion int
	ProfileVersion  int
	IdentitySteps   []MigrationStep
	ProfileSteps    []MigrationStep
}

// DefaultMigrationPlan returns the v0.1 baseline migration policy.
func DefaultMigrationPlan() MigrationPlan {
	identitySteps := []MigrationStep{{
		FromVersion: 0,
		ToVersion:   identitySchemaVersion,
		Description: "initialize identity schema v1",
	}}
	profileSteps := []MigrationStep{{
		FromVersion: 0,
		ToVersion:   profileSchemaVersion,
		Description: "initialize profile schema v1",
	}}
	return MigrationPlan{
		IdentityVersion: identitySchemaVersion,
		ProfileVersion:  profileSchemaVersion,
		IdentitySteps:   append([]MigrationStep(nil), identitySteps...),
		ProfileSteps:    append([]MigrationStep(nil), profileSteps...),
	}
}

// Validate enforces migration policy invariants used for rollback-safe planning.
func (p MigrationPlan) Validate() error {
	if p.IdentityVersion < 1 || p.ProfileVersion < 1 {
		return errors.New("phase5: schema versions must be >= 1")
	}
	if err := validateSteps(p.IdentitySteps, p.IdentityVersion, "identity"); err != nil {
		return err
	}
	if err := validateSteps(p.ProfileSteps, p.ProfileVersion, "profile"); err != nil {
		return err
	}
	return nil
}

// BackupPlan captures backup-before-migration behavior requirements.
type BackupPlan struct {
	Required       bool
	BackupSuffix   string
	PreflightCheck string
}

// DefaultBackupPlan returns the baseline backup policy for migrations.
func DefaultBackupPlan() BackupPlan {
	return BackupPlan{
		Required:       true,
		BackupSuffix:   ".bak",
		PreflightCheck: "ensure source DB exists and backup path is writable",
	}
}

// MigrationFixture defines positive and negative migration test expectations.
type MigrationFixture struct {
	Name            string
	StartVersions   SchemaVersionSet
	ExpectErrorText string
}

// DefaultMigrationFixtures returns baseline fixtures used by migration tests.
func DefaultMigrationFixtures() []MigrationFixture {
	return []MigrationFixture{
		{
			Name:          "fresh install initializes v1",
			StartVersions: SchemaVersionSet{Identity: 0, Profile: 0},
		},
		{
			Name:            "unknown identity version rejected",
			StartVersions:   SchemaVersionSet{Identity: identitySchemaVersion + 1, Profile: profileSchemaVersion},
			ExpectErrorText: "identity schema version",
		},
		{
			Name:            "unknown profile version rejected",
			StartVersions:   SchemaVersionSet{Identity: identitySchemaVersion, Profile: profileSchemaVersion + 1},
			ExpectErrorText: "profile schema version",
		},
	}
}

// ValidateStartVersions validates whether a start version set is migratable.
func (p MigrationPlan) ValidateStartVersions(start SchemaVersionSet) error {
	if start.Identity < 0 {
		return errors.New("phase5: identity schema version must be >= 0")
	}
	if start.Profile < 0 {
		return errors.New("phase5: profile schema version must be >= 0")
	}
	if start.Identity > p.IdentityVersion {
		return fmt.Errorf("phase5: identity schema version %d is newer than supported %d", start.Identity, p.IdentityVersion)
	}
	if start.Profile > p.ProfileVersion {
		return fmt.Errorf("phase5: profile schema version %d is newer than supported %d", start.Profile, p.ProfileVersion)
	}
	return nil
}

func validateSteps(steps []MigrationStep, target int, domain string) error {
	if len(steps) == 0 {
		return fmt.Errorf("phase5: %s migration steps required", domain)
	}
	prevTo := -1
	for i, step := range steps {
		if step.FromVersion < 0 {
			return fmt.Errorf("phase5: %s migration step %d has negative from-version", domain, i)
		}
		if step.ToVersion <= step.FromVersion {
			return fmt.Errorf("phase5: %s migration step %d must increase version", domain, i)
		}
		if step.Description == "" {
			return fmt.Errorf("phase5: %s migration step %d description required", domain, i)
		}
		if i == 0 {
			if step.FromVersion != 0 {
				return fmt.Errorf("phase5: %s migration must begin at version 0", domain)
			}
		} else if step.FromVersion != prevTo {
			return fmt.Errorf("phase5: %s migration step %d is not contiguous", domain, i)
		}
		prevTo = step.ToVersion
	}
	if prevTo != target {
		return fmt.Errorf("phase5: %s migration target mismatch: got %d want %d", domain, prevTo, target)
	}
	return nil
}
