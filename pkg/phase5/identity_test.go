package phase5

import (
	"bytes"
	"crypto/ed25519"
	"errors"
	"math"
	"runtime"
	"testing"
	"unsafe"
)

func cloneEnvelope(env StorageEnvelope) StorageEnvelope {
	return StorageEnvelope{
		Salt:       append([]byte(nil), env.Salt...),
		Nonce:      append([]byte(nil), env.Nonce...),
		Ciphertext: append([]byte(nil), env.Ciphertext...),
	}
}

func cloneWordParts(parts [][]byte) [][]byte {
	cloned := make([][]byte, len(parts))
	for i, part := range parts {
		cloned[i] = append([]byte(nil), part...)
	}
	return cloned
}

func TestSeedEncodingRoundTrip(t *testing.T) {
	seed := make([]byte, 32)
	copy(seed, []byte("0123456789abcdef0123456789abcdef"))
	encoded := EncodeSeed(seed)
	decoded, err := ParseSeed(encoded)
	if err != nil {
		t.Fatalf("expected parse to succeed, got %v", err)
	}
	if string(decoded) != string(seed) {
		t.Fatalf("round trip seed mismatch")
	}

	_, err = ParseSeed("badhex")
	if err == nil {
		t.Fatalf("expected parse failure for bad hex")
	}
}

func TestStorageEnvelopeSerialization(t *testing.T) {
	_, priv, err := DeriveKeyFromSeed(bytes.Repeat([]byte{0x01}, 32))
	if err != nil {
		t.Fatalf("derive key failed: %v", err)
	}
	env, err := SealPrivateKey(priv, "passphrase")
	if err != nil {
		t.Fatalf("seal private key failed: %v", err)
	}
	data, err := env.Marshal()
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	restored, err := UnmarshalEnvelope(data)
	if err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	opened, err := OpenPrivateKey(restored, "passphrase")
	if err != nil {
		t.Fatalf("open failed: %v", err)
	}
	if !ed25519.PrivateKey(opened).Equal(priv) {
		t.Fatalf("opened private key does not match")
	}
	if len(restored.Ciphertext) == 0 {
		t.Fatalf("ciphertext should be populated")
	}
}

func TestEnvelopeValidationErrors(t *testing.T) {
	_, err := UnmarshalEnvelope([]byte{'b', 'a', 'd'})
	if err == nil {
		t.Fatalf("expected error on malformed envelope")
	}
	env := StorageEnvelope{Salt: make([]byte, saltSize)}
	if err := env.Validate(); err == nil {
		t.Fatalf("expected validation error for incomplete envelope")
	}
}

func TestIdentityLifecycleFreshInstallFlow(t *testing.T) {
	seed := bytes.Repeat([]byte{0x11}, seedSize)
	pub, priv, err := DeriveKeyFromSeed(seed)
	if err != nil {
		t.Fatalf("derive key failed: %v", err)
	}
	passphrase := "fresh-install-passphrase"
	env, err := SealPrivateKey(priv, passphrase)
	if err != nil {
		t.Fatalf("seal private key failed: %v", err)
	}
	serialized, err := env.Marshal()
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	if len(serialized) == 0 {
		t.Fatalf("expected serialized envelope output")
	}
	restored, err := UnmarshalEnvelope(serialized)
	if err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	reopened, err := OpenPrivateKey(restored, passphrase)
	if err != nil {
		t.Fatalf("open failed: %v", err)
	}
	if !ed25519.PrivateKey(reopened).Equal(priv) {
		t.Fatalf("private key mismatch after reopen")
	}
	reopenedPub := reopened.Public().(ed25519.PublicKey)
	if !pub.Equal(reopenedPub) {
		t.Fatalf("public key mismatch after reopen")
	}
}

func TestIdentityLifecycleRestartPersistence(t *testing.T) {
	seed := bytes.Repeat([]byte{0x22}, seedSize)
	_, priv, err := DeriveKeyFromSeed(seed)
	if err != nil {
		t.Fatalf("derive key failed: %v", err)
	}
	passphrase := "restart-passphrase"
	env, err := SealPrivateKey(priv, passphrase)
	if err != nil {
		t.Fatalf("seal failed: %v", err)
	}
	serialized, err := env.Marshal()
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	restored, err := UnmarshalEnvelope(serialized)
	if err != nil {
		t.Fatalf("initial unmarshal failed: %v", err)
	}
	reencoded, err := restored.Marshal()
	if err != nil {
		t.Fatalf("re-marshal failed: %v", err)
	}
	if !bytes.Equal(serialized, reencoded) {
		t.Fatalf("expected envelope marshal to be stable across restarts")
	}
	const restartCount = 5
	for i := 0; i < restartCount; i++ {
		reloaded, err := UnmarshalEnvelope(serialized)
		if err != nil {
			t.Fatalf("restart %d unmarshal failed: %v", i, err)
		}
		reopened, err := OpenPrivateKey(reloaded, passphrase)
		if err != nil {
			t.Fatalf("restart %d open failed: %v", i, err)
		}
		if !ed25519.PrivateKey(reopened).Equal(priv) {
			t.Fatalf("restart %d private key mismatch", i)
		}
	}
}

func TestDefaultMigrationPlanValidate(t *testing.T) {
	plan := DefaultMigrationPlan()
	if err := plan.Validate(); err != nil {
		t.Fatalf("expected default plan valid, got %v", err)
	}
	if plan.IdentityVersion != identitySchemaVersion {
		t.Fatalf("identity version mismatch: got %d want %d", plan.IdentityVersion, identitySchemaVersion)
	}
	if plan.ProfileVersion != profileSchemaVersion {
		t.Fatalf("profile version mismatch: got %d want %d", plan.ProfileVersion, profileSchemaVersion)
	}
	if len(plan.IdentitySteps) != 1 || len(plan.ProfileSteps) != 1 {
		t.Fatalf("expected single bootstrap step for identity/profile")
	}
	if plan.IdentitySteps[0].FromVersion != 0 || plan.IdentitySteps[0].ToVersion != identitySchemaVersion {
		t.Fatalf("unexpected identity bootstrap step: %+v", plan.IdentitySteps[0])
	}
	if plan.ProfileSteps[0].FromVersion != 0 || plan.ProfileSteps[0].ToVersion != profileSchemaVersion {
		t.Fatalf("unexpected profile bootstrap step: %+v", plan.ProfileSteps[0])
	}
}

func TestMigrationPlanValidateRejectsInvalidShapes(t *testing.T) {
	tests := []struct {
		name string
		plan MigrationPlan
	}{
		{
			name: "identity target mismatch",
			plan: MigrationPlan{
				IdentityVersion: 2,
				ProfileVersion:  1,
				IdentitySteps: []MigrationStep{{
					FromVersion: 0,
					ToVersion:   1,
					Description: "bootstrap",
				}},
				ProfileSteps: []MigrationStep{{
					FromVersion: 0,
					ToVersion:   1,
					Description: "bootstrap",
				}},
			},
		},
		{
			name: "identity non contiguous",
			plan: MigrationPlan{
				IdentityVersion: 2,
				ProfileVersion:  1,
				IdentitySteps: []MigrationStep{
					{FromVersion: 0, ToVersion: 1, Description: "bootstrap"},
					{FromVersion: 0, ToVersion: 2, Description: "bad jump"},
				},
				ProfileSteps: []MigrationStep{{
					FromVersion: 0,
					ToVersion:   1,
					Description: "bootstrap",
				}},
			},
		},
		{
			name: "empty description",
			plan: MigrationPlan{
				IdentityVersion: 1,
				ProfileVersion:  1,
				IdentitySteps: []MigrationStep{{
					FromVersion: 0,
					ToVersion:   1,
					Description: "",
				}},
				ProfileSteps: []MigrationStep{{
					FromVersion: 0,
					ToVersion:   1,
					Description: "bootstrap",
				}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.plan.Validate(); err == nil {
				t.Fatalf("expected validate error for %s", tt.name)
			}
		})
	}
}

func TestValidateStartVersions(t *testing.T) {
	plan := DefaultMigrationPlan()
	if err := plan.ValidateStartVersions(SchemaVersionSet{Identity: 0, Profile: 0}); err != nil {
		t.Fatalf("fresh install versions should be valid: %v", err)
	}
	if err := plan.ValidateStartVersions(SchemaVersionSet{Identity: identitySchemaVersion, Profile: profileSchemaVersion}); err != nil {
		t.Fatalf("current versions should be valid: %v", err)
	}
	if err := plan.ValidateStartVersions(SchemaVersionSet{Identity: -1, Profile: 0}); err == nil {
		t.Fatalf("expected negative identity version error")
	}
	if err := plan.ValidateStartVersions(SchemaVersionSet{Identity: identitySchemaVersion + 1, Profile: profileSchemaVersion}); err == nil {
		t.Fatalf("expected unsupported newer identity version error")
	}
	if err := plan.ValidateStartVersions(SchemaVersionSet{Identity: identitySchemaVersion, Profile: profileSchemaVersion + 1}); err == nil {
		t.Fatalf("expected unsupported newer profile version error")
	}
}

func TestDefaultBackupPlan(t *testing.T) {
	backup := DefaultBackupPlan()
	if !backup.Required {
		t.Fatalf("backup should be required before migration")
	}
	if backup.BackupSuffix != ".bak" {
		t.Fatalf("unexpected backup suffix: %q", backup.BackupSuffix)
	}
	if backup.PreflightCheck == "" {
		t.Fatalf("expected non-empty preflight check")
	}
}

func TestDefaultMigrationFixtures(t *testing.T) {
	fixtures := DefaultMigrationFixtures()
	if len(fixtures) < 3 {
		t.Fatalf("expected baseline fixture set, got %d", len(fixtures))
	}

	plan := DefaultMigrationPlan()
	for _, fixture := range fixtures {
		err := plan.ValidateStartVersions(fixture.StartVersions)
		if fixture.ExpectErrorText == "" {
			if err != nil {
				t.Fatalf("fixture %q expected valid start version, got %v", fixture.Name, err)
			}
			continue
		}
		if err == nil {
			t.Fatalf("fixture %q expected error containing %q", fixture.Name, fixture.ExpectErrorText)
		}
		if !bytes.Contains([]byte(err.Error()), []byte(fixture.ExpectErrorText)) {
			t.Fatalf("fixture %q error %q missing %q", fixture.Name, err.Error(), fixture.ExpectErrorText)
		}
	}
}

func TestGenerateMnemonicRecoverSeedRoundTrip(t *testing.T) {
	seed := bytes.Repeat([]byte{0x5a}, seedSize)
	mnemonic, err := GenerateMnemonic(seed)
	if err != nil {
		t.Fatalf("generate mnemonic failed: %v", err)
	}
	if mnemonic == "" {
		t.Fatalf("expected non-empty mnemonic")
	}
	recovered, err := RecoverSeedFromMnemonic(mnemonic)
	if err != nil {
		t.Fatalf("recover seed failed: %v", err)
	}
	if !bytes.Equal(recovered, seed) {
		t.Fatalf("recovered seed mismatch")
	}

	pubExpected, privExpected, err := DeriveKeyFromSeed(seed)
	if err != nil {
		t.Fatalf("derive expected key failed: %v", err)
	}
	pubRecovered, privRecovered, err := RecoverKeyFromMnemonic(mnemonic)
	if err != nil {
		t.Fatalf("recover key from mnemonic failed: %v", err)
	}
	if !pubExpected.Equal(pubRecovered) {
		t.Fatalf("public key mismatch from recovered mnemonic")
	}
	if !privExpected.Equal(privRecovered) {
		t.Fatalf("private key mismatch from recovered mnemonic")
	}
}

func TestMnemonicInputValidationAndChecksum(t *testing.T) {
	seed := bytes.Repeat([]byte{0x21}, seedSize)
	mnemonic, err := GenerateMnemonic(seed)
	if err != nil {
		t.Fatalf("generate mnemonic failed: %v", err)
	}
	parts := bytes.Fields([]byte(mnemonic))
	if len(parts) != mnemonicWords {
		t.Fatalf("expected mnemonic words to equal %d", mnemonicWords)
	}
	joinWords := func(p [][]byte) string {
		return string(bytes.Join(p, []byte(" ")))
	}
	unknownWordParts := cloneWordParts(parts)
	unknownWordParts[5] = []byte("invalidword")
	checksumParts := cloneWordParts(parts)
	checksumParts[len(checksumParts)-1] = []byte("albar")
	tests := []struct {
		name   string
		phrase string
	}{
		{name: "wrong word count", phrase: "only two"},
		{name: "unknown word token", phrase: joinWords(unknownWordParts)},
		{name: "checksum mismatch", phrase: joinWords(checksumParts)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := RecoverSeedFromMnemonic(tt.phrase); err == nil {
				t.Fatalf("expected %s error", tt.name)
			}
		})
	}
}

func TestStorageEnvelopeRecoveryFailures(t *testing.T) {
	seed := bytes.Repeat([]byte{0x33}, seedSize)
	_, priv, err := DeriveKeyFromSeed(seed)
	if err != nil {
		t.Fatalf("derive key failed: %v", err)
	}
	passphrase := "correct-passphrase"
	env, err := SealPrivateKey(priv, passphrase)
	if err != nil {
		t.Fatalf("seal failed: %v", err)
	}
	cases := []struct {
		name       string
		mutate     func(StorageEnvelope) StorageEnvelope
		passphrase string
	}{
		{
			name: "wrong passphrase",
			mutate: func(e StorageEnvelope) StorageEnvelope {
				return cloneEnvelope(e)
			},
			passphrase: "incorrect-passphrase",
		},
		{
			name: "ciphertext corrupted",
			mutate: func(e StorageEnvelope) StorageEnvelope {
				corrupted := cloneEnvelope(e)
				corrupted.Ciphertext[0] ^= 0xFF
				return corrupted
			},
			passphrase: passphrase,
		},
		{
			name: "ciphertext truncated",
			mutate: func(e StorageEnvelope) StorageEnvelope {
				corrupted := cloneEnvelope(e)
				if len(corrupted.Ciphertext) == 0 {
					t.Fatalf("ciphertext should not be empty")
				}
				corrupted.Ciphertext = corrupted.Ciphertext[:len(corrupted.Ciphertext)-1]
				return corrupted
			},
			passphrase: passphrase,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			mutated := tt.mutate(env)
			if _, err := OpenPrivateKey(mutated, tt.passphrase); err == nil {
				t.Fatalf("expected failure for %s", tt.name)
			}
		})
	}
}

func TestStorageEnvelopeLengthWithinBounds(t *testing.T) {
	if !storageEnvelopeLengthWithinBounds(math.MaxUint32) {
		t.Fatalf("expected MaxUint32 to be within bounds")
	}
	if storageEnvelopeLengthWithinBounds(uint64(math.MaxUint32) + 1) {
		t.Fatalf("expected MaxUint32+1 to be out of bounds")
	}
}

func TestStorageEnvelopeMarshalLengthsWithinBounds(t *testing.T) {
	if !storageEnvelopeMarshalLengthsWithinBounds(16, 12, 1) {
		t.Fatalf("expected normal envelope lengths to be within bounds")
	}
	if storageEnvelopeMarshalLengthsWithinBounds(uint64(math.MaxUint32)+1, 12, 1) {
		t.Fatalf("expected salt length overflow to be out of bounds")
	}
	if storageEnvelopeMarshalLengthsWithinBounds(16, uint64(math.MaxUint32)+1, 1) {
		t.Fatalf("expected nonce length overflow to be out of bounds")
	}
	if storageEnvelopeMarshalLengthsWithinBounds(16, 12, uint64(math.MaxUint32)+1) {
		t.Fatalf("expected ciphertext length overflow to be out of bounds")
	}
}

func TestStorageEnvelopeMarshalReturnsInvalidEnvelopeForSyntheticLengthOverflow(t *testing.T) {
	overflowLen := uint64(math.MaxUint32) + 1
	if storageEnvelopeMarshalLengthsWithinBounds(uint64(saltSize), uint64(nonceSize), overflowLen) {
		t.Fatalf("expected helper to reject out-of-range ciphertext length")
	}

	syntheticCiphertext, keepAlive := syntheticByteSliceWithLength(int(overflowLen))
	env := StorageEnvelope{
		Salt:       bytes.Repeat([]byte{0x01}, saltSize),
		Nonce:      bytes.Repeat([]byte{0x02}, nonceSize),
		Ciphertext: syntheticCiphertext,
	}

	_, err := env.Marshal()
	runtime.KeepAlive(keepAlive)
	if !errors.Is(err, errInvalidEnvelope) {
		t.Fatalf("Marshal() error = %v, want errInvalidEnvelope", err)
	}
}

func syntheticByteSliceWithLength(length int) ([]byte, []byte) {
	backing := []byte{0xAA}
	return unsafe.Slice((*byte)(unsafe.Pointer(&backing[0])), length), backing
}
