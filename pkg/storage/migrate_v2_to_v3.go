package storage

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// MigrateV2ToV3 migrates a format-version-2 store (SHA-256 KDF) to format-version-3 (Argon2id).
// It reads the existing DB using the v2 key, derives the v3 key using Argon2id with a new random
// salt, re-keys the SQLCipher database, rewrites the meta file with v3 parameters, and archives
// the old meta as state.store.migrated-<unix_timestamp>.json.
//
// dataDir is the directory containing state.db and state.db.meta.json.
// secret is the raw 32-byte secret (from env var or state.key file).
func MigrateV2ToV3(dataDir string, secret []byte) error {
	// Step a: read meta and verify format_version == 2.
	keyMeta, err := readStoreKeyMetadata(dataDir)
	if err != nil {
		return fmt.Errorf("migrate v2→v3: read meta: %w", err)
	}
	if keyMeta.FormatVersion != 2 {
		return fmt.Errorf("migrate v2→v3: expected format_version=2, got %d", keyMeta.FormatVersion)
	}

	// Step b: derive v2 key (SHA-256) and verify key_check.
	salt := decodeBase64(keyMeta.Salt)
	if len(salt) == 0 {
		return fmt.Errorf("migrate v2→v3: invalid salt in meta")
	}
	v2Key := deriveKeyLegacySHA256(salt, secret)
	if !matchesKeyCheck(v2Key, keyMeta.KeyCheck) {
		return ErrWrongKey
	}

	// Step c: generate new 32-byte random salt.
	newSalt := make([]byte, 32)
	if _, err := rand.Read(newSalt); err != nil {
		return fmt.Errorf("migrate v2→v3: generate new salt: %w", err)
	}

	// Step d: derive v3 key (Argon2id).
	v3Key := deriveStoreKeyArgon2id(newSalt, secret)

	// Step e: open DB with v2 key and execute PRAGMA rekey to v3 key.
	dbPath := storePath(dataDir)
	db, err := sql.Open("sqlite3", sqlcipherDSN(dbPath, v2Key))
	if err != nil {
		return fmt.Errorf("migrate v2→v3: open db: %w", err)
	}
	defer db.Close()

	newKeyHex := hex.EncodeToString(v3Key)
	if _, err := db.Exec(fmt.Sprintf("PRAGMA rekey = x'%s'", newKeyHex)); err != nil {
		return fmt.Errorf("migrate v2→v3: PRAGMA rekey failed: %w", err)
	}

	// Flush WAL before writing meta (non-fatal on failure).
	if _, cpErr := db.Exec("PRAGMA wal_checkpoint(FULL)"); cpErr != nil {
		_ = cpErr
	}

	// Step g: archive old meta as state.store.migrated-<unix_timestamp>.json.
	oldMetaPath := storeMetadataPath(dataDir)
	archivePath := filepath.Join(dataDir, "state.store.migrated-"+fmt.Sprintf("%d", time.Now().UTC().Unix())+".json")
	if oldMetaBytes, readErr := os.ReadFile(oldMetaPath); readErr == nil {
		_ = os.WriteFile(archivePath, oldMetaBytes, 0o600)
	}

	// Step f: rewrite meta file atomically with format_version=3, new salt, Argon2id params.
	now := time.Now().UTC()
	newMeta := storeKeyMetadata{
		FormatVersion:   3,
		Salt:            base64.RawURLEncoding.EncodeToString(newSalt),
		KeyCheck:        keyCheck(v3Key),
		KDF:             "argon2id",
		Argon2idMemory:  64 * 1024,
		Argon2idTime:    3,
		Argon2idThreads: 2,
		CreatedAt:       keyMeta.CreatedAt,
		UpdatedAt:       now,
	}
	if err := writeStoreKeyMetadata(dataDir, newMeta); err != nil {
		// Attempt rollback: rekey back to v2 key.
		oldKeyHex := hex.EncodeToString(v2Key)
		if _, rekeyErr := db.Exec(fmt.Sprintf("PRAGMA rekey = x'%s'", oldKeyHex)); rekeyErr != nil {
			return fmt.Errorf("migrate v2→v3: meta write failed AND rollback failed: meta=%w; rekey=%v", err, rekeyErr)
		}
		return fmt.Errorf("migrate v2→v3: write new meta (rolled back): %w", err)
	}

	return nil
}

// deriveKeyLegacySHA256 derives a store key using the legacy v2 SHA-256 method.
// Used only for v2→v3 migration reads; not for new stores.
func deriveKeyLegacySHA256(salt, secret []byte) []byte {
	combined := make([]byte, 0, len(salt)+len(secret))
	combined = append(combined, salt...)
	combined = append(combined, secret...)
	sum := sha256.Sum256(combined)
	out := make([]byte, 32)
	copy(out, sum[:])
	return out
}
