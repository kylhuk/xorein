package storage

import (
	"bytes"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/mutecomm/go-sqlcipher/v4"
)

func TestStoreSaveLoadWrongKeyAndEncryption(t *testing.T) {
	dataDir := t.TempDir()
	snapshot := Snapshot{SchemaVersion: 2, Buckets: map[string][]byte{"identity": []byte(`{"peer_id":"peer-1"}`), "settings": []byte(`{"mode":"client"}`)}}
	if err := Save(dataDir, snapshot); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	if !isEncryptedDatabase(storePath(dataDir)) {
		t.Fatal("expected SQLCipher database to be encrypted on disk")
	}
	loaded, err := Load(dataDir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if loaded.SchemaVersion != snapshot.SchemaVersion {
		t.Fatalf("loaded schema version = %d want %d", loaded.SchemaVersion, snapshot.SchemaVersion)
	}
	if string(loaded.Buckets["settings"]) != string(snapshot.Buckets["settings"]) {
		t.Fatalf("loaded settings = %s want %s", loaded.Buckets["settings"], snapshot.Buckets["settings"])
	}
	wrongSecret := make([]byte, 32)
	copy(wrongSecret, []byte("different-sql-store-secret-value!"))
	if err := os.WriteFile(filepath.Join(dataDir, storeKeyName), []byte(base64.RawURLEncoding.EncodeToString(wrongSecret)+"\n"), 0o600); err != nil {
		t.Fatalf("WriteFile(state.key) error = %v", err)
	}
	if _, err := Load(dataDir); err == nil || err != ErrWrongKey {
		t.Fatalf("Load() wrong key error = %v want %v", err, ErrWrongKey)
	}
}

func TestStoreDetectsCorruptAndMissingMetadata(t *testing.T) {
	dataDir := t.TempDir()
	if err := Save(dataDir, Snapshot{SchemaVersion: 2, Buckets: map[string][]byte{"identity": []byte(`{"peer_id":"peer-1"}`)}}); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	if err := os.WriteFile(storePath(dataDir), []byte("not-a-sqlite-db"), 0o600); err != nil {
		t.Fatalf("WriteFile(state.db) error = %v", err)
	}
	if _, err := Load(dataDir); err == nil || err != ErrCorrupt {
		t.Fatalf("Load() corrupt db error = %v want %v", err, ErrCorrupt)
	}

	missingMetaDir := t.TempDir()
	if err := Save(missingMetaDir, Snapshot{SchemaVersion: 2, Buckets: map[string][]byte{"identity": []byte(`{"peer_id":"peer-2"}`)}}); err != nil {
		t.Fatalf("Save(missingMeta) error = %v", err)
	}
	if err := os.Remove(storeMetadataPath(missingMetaDir)); err != nil {
		t.Fatalf("Remove(state.db.meta.json) error = %v", err)
	}
	if _, err := Load(missingMetaDir); err == nil || err != ErrCorrupt {
		t.Fatalf("Load() missing metadata error = %v want %v", err, ErrCorrupt)
	}

	missingDBMetaDir := t.TempDir()
	_, key, err := resolveStoreKeyMetadata(missingDBMetaDir, false, true)
	if err != nil {
		t.Fatalf("resolveStoreKeyMetadata() error = %v", err)
	}
	db, err := sql.Open("sqlite3", sqlcipherDSN(storePath(missingDBMetaDir), key))
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	if err := configureDB(db); err != nil {
		_ = db.Close()
		t.Fatalf("configureDB() error = %v", err)
	}
	if _, err := db.Exec(`DROP TABLE ` + metadataTableName); err != nil {
		_ = db.Close()
		t.Fatalf("DROP TABLE store_metadata error = %v", err)
	}
	_ = db.Close()
	if _, err := Load(missingDBMetaDir); err == nil || err != ErrCorrupt {
		t.Fatalf("Load() missing DB metadata error = %v want %v", err, ErrCorrupt)
	}
}

func TestStoreMigratesLegacyDirectory(t *testing.T) {
	dataDir := t.TempDir()
	legacySnapshot := Snapshot{SchemaVersion: 2, Buckets: map[string][]byte{"identity": []byte(`{"peer_id":"peer-1"}`), "settings": []byte(`{"mode":"relay"}`)}}
	if err := writeLegacyStoreFixture(dataDir, legacySnapshot); err != nil {
		t.Fatalf("writeLegacyStoreFixture() error = %v", err)
	}

	loaded, err := Load(dataDir)
	if err != nil {
		t.Fatalf("Load() migration error = %v", err)
	}
	if string(loaded.Buckets["settings"]) != string(legacySnapshot.Buckets["settings"]) {
		t.Fatalf("loaded settings = %s want %s", loaded.Buckets["settings"], legacySnapshot.Buckets["settings"])
	}
	if _, err := os.Stat(storePath(dataDir)); err != nil {
		t.Fatalf("Stat(state.db) error = %v", err)
	}
	if !isEncryptedDatabase(storePath(dataDir)) {
		t.Fatal("expected migrated SQLCipher database to be encrypted on disk")
	}
	matches, err := filepath.Glob(filepath.Join(dataDir, StoreDirName+migrationBackupTag+"*"))
	if err != nil {
		t.Fatalf("Glob(legacy backup) error = %v", err)
	}
	if len(matches) == 0 {
		t.Fatal("expected legacy store archive after migration")
	}
	if _, err := os.Stat(legacyStorePath(dataDir)); !os.IsNotExist(err) {
		t.Fatalf("legacy store dir still present: %v", err)
	}
}

func TestSealAndOpenPayloadRoundTrip(t *testing.T) {
	dataDir := t.TempDir()
	if err := Save(dataDir, Snapshot{SchemaVersion: 2, Buckets: map[string][]byte{}}); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	plaintext := []byte(`{"kind":"channel_message","body":"relay secret"}`)
	sealed, err := SealPayload(dataDir, plaintext)
	if err != nil {
		t.Fatalf("SealPayload() error = %v", err)
	}
	if bytes.Contains(sealed, plaintext) {
		t.Fatalf("sealed payload leaked plaintext: %q", string(sealed))
	}
	opened, err := OpenPayload(dataDir, sealed)
	if err != nil {
		t.Fatalf("OpenPayload() error = %v", err)
	}
	if string(opened) != string(plaintext) {
		t.Fatalf("opened payload = %s want %s", opened, plaintext)
	}
}

func writeLegacyStoreFixture(dataDir string, snapshot Snapshot) error {
	secret := make([]byte, 32)
	for i := range secret {
		secret[i] = byte(i + 1)
	}
	if err := os.WriteFile(filepath.Join(dataDir, storeKeyName), []byte(base64.RawURLEncoding.EncodeToString(secret)+"\n"), 0o600); err != nil {
		return err
	}
	salt := make([]byte, 32)
	for i := range salt {
		salt[i] = byte(255 - i)
	}
	key, err := deriveStoreKey(dataDir, salt, false)
	if err != nil {
		return err
	}
	storeDir := legacyStorePath(dataDir)
	genDir := filepath.Join(storeDir, "gen-000001")
	if err := os.MkdirAll(genDir, 0o700); err != nil {
		return err
	}
	meta := legacyMetadata{
		FormatVersion:     1,
		SchemaVersion:     snapshot.SchemaVersion,
		Salt:              base64.RawURLEncoding.EncodeToString(salt),
		KeyCheck:          keyCheck(key),
		CurrentGeneration: "gen-000001",
	}
	rawMeta, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(storeDir, legacyStoreMetaName), rawMeta, 0o600); err != nil {
		return err
	}
	for name, payload := range snapshot.Buckets {
		ciphertext, err := encryptLegacyBucket(key, payload)
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(genDir, name+legacyBucketExtension), ciphertext, 0o600); err != nil {
			return err
		}
	}
	return nil
}
