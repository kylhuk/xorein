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

// Rotate generates a new salt and secret, re-encrypts the SQLCipher database
// with PRAGMA rekey, and atomically rewrites state.db.meta.json.
//
// The 7-step procedure follows docs/spec/v0.1/70-storage-and-key-derivation.md §7:
//  1. Open DB with current key to verify it still works.
//  2. Generate 32-byte new salt.
//  3. Generate 32-byte new secret; write to state.key at 0600 (overwriting old).
//  4. Derive new key: SHA-256(newSalt || newSecret).
//  5. Execute PRAGMA rekey = x'<newKey>'.
//  6. Atomically rewrite state.db.meta.json with new salt + key_check.
//  7. On any failure in steps 5–6, attempt to rekey back to old key and return error.
func Rotate(dataDir string) error {
	// Step 1: verify current key opens the DB.
	keyMeta, oldKey, err := resolveStoreKeyMetadata(dataDir, true, false)
	if err != nil {
		return fmt.Errorf("rotate: open current store: %w", err)
	}
	_ = keyMeta

	// Step 2: generate new salt.
	newSalt := make([]byte, 32)
	if _, err := rand.Read(newSalt); err != nil {
		return fmt.Errorf("rotate: generate new salt: %w", err)
	}

	// Step 3: generate new secret and write to state.key at 0600.
	newSecret := make([]byte, 32)
	if _, err := rand.Read(newSecret); err != nil {
		return fmt.Errorf("rotate: generate new secret: %w", err)
	}
	keyFilePath := storeKeyFilePath(dataDir)
	encoded := base64.RawURLEncoding.EncodeToString(newSecret)
	if err := os.WriteFile(keyFilePath, []byte(encoded+"\n"), 0o600); err != nil {
		return fmt.Errorf("rotate: write new state.key: %w", err)
	}

	// Step 4: derive new key.
	h := sha256.Sum256(append(append([]byte(nil), newSalt...), newSecret...))
	newKey := h[:]

	// Step 5: open DB with old key and execute PRAGMA rekey.
	dbPath := storePath(dataDir)
	db, err := sql.Open("sqlite3", sqlcipherDSN(dbPath, oldKey))
	if err != nil {
		return fmt.Errorf("rotate: open db for rekey: %w", err)
	}
	defer db.Close()

	newKeyHex := hex.EncodeToString(newKey)
	if _, err := db.Exec(fmt.Sprintf("PRAGMA rekey = x'%s'", newKeyHex)); err != nil {
		// Attempt rollback: restore old secret.
		oldEncoded := base64.RawURLEncoding.EncodeToString(oldKey)
		_ = os.WriteFile(keyFilePath, []byte(oldEncoded+"\n"), 0o600)
		return fmt.Errorf("rotate: PRAGMA rekey failed: %w", err)
	}

	// Step 6: atomically rewrite state.db.meta.json.
	newMeta := storeKeyMetadata{
		FormatVersion: FormatVersion,
		Salt:          base64.RawURLEncoding.EncodeToString(newSalt),
		KeyCheck:      keyCheck(newKey),
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}
	if err := writeStoreKeyMetadata(dataDir, newMeta); err != nil {
		// Step 7: rollback — rekey back to old key.
		if _, rekeyErr := db.Exec(fmt.Sprintf("PRAGMA rekey = x'%s'", hex.EncodeToString(oldKey))); rekeyErr != nil {
			return fmt.Errorf("rotate: meta write failed AND rollback failed: meta=%w; rekey=%v", err, rekeyErr)
		}
		// Restore old secret.
		oldEncoded := base64.RawURLEncoding.EncodeToString(oldKey)
		_ = os.WriteFile(keyFilePath, []byte(oldEncoded+"\n"), 0o600)
		return fmt.Errorf("rotate: write new meta (rolled back): %w", err)
	}

	// Fsync the DB to flush WAL.
	if _, err := db.Exec("PRAGMA wal_checkpoint(FULL)"); err != nil {
		// Non-fatal: checkpoint failure doesn't invalidate the rotation.
		_ = err
	}

	return nil
}

func storeKeyFilePath(dataDir string) string {
	return filepath.Join(dataDir, storeKeyName)
}
