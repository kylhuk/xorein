package storage

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	sqlcipher "github.com/mutecomm/go-sqlcipher/v4"
	"golang.org/x/crypto/argon2"
)

const (
	FormatVersion      = 3
	StoreFileName      = "state.db"
	StoreDirName       = "state.store"
	StoreMetaFileName  = "state.db.meta.json"
	storeKeyName       = "state.key"
	secretEnvName      = "XOREIN_STATE_KEY"
	metadataTableName  = "store_metadata"
	bucketTableName    = "store_buckets"
	metadataSingleton  = 1
	migrationBackupTag = ".migrated-"
	cipherPageSize     = 4096
)

var (
	ErrStoreNotFound = errors.New("storage: store not found")
	ErrWrongKey      = errors.New("storage: wrong key")
	ErrCorrupt       = errors.New("storage: corrupt store")
)

type Snapshot struct {
	SchemaVersion int
	Buckets       map[string][]byte
}

type metadata struct {
	FormatVersion int
	SchemaVersion int
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type storeKeyMetadata struct {
	FormatVersion    int       `json:"format_version"`
	Salt             string    `json:"salt"`
	KeyCheck         string    `json:"key_check"`
	KDF              string    `json:"kdf,omitempty"`
	Argon2idMemory   uint32    `json:"argon2id_memory,omitempty"`
	Argon2idTime     uint32    `json:"argon2id_time,omitempty"`
	Argon2idThreads  uint8     `json:"argon2id_threads,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

func Load(dataDir string) (Snapshot, error) {
	db, meta, err := openStore(dataDir, false)
	if err != nil {
		if errors.Is(err, ErrStoreNotFound) {
			snapshot, migrated, migrateErr := migrateLegacyStore(dataDir)
			if migrated || migrateErr == nil {
				return snapshot, migrateErr
			}
			if !errors.Is(migrateErr, ErrStoreNotFound) {
				return Snapshot{}, migrateErr
			}
		}
		return Snapshot{}, err
	}
	defer db.Close()

	rows, err := db.Query(`SELECT name, payload FROM ` + bucketTableName + ` ORDER BY name`)
	if err != nil {
		return Snapshot{}, mapSQLError(err)
	}
	defer rows.Close()
	buckets := map[string][]byte{}
	for rows.Next() {
		var (
			name    string
			payload []byte
		)
		if err := rows.Scan(&name, &payload); err != nil {
			return Snapshot{}, mapSQLError(err)
		}
		buckets[name] = append([]byte(nil), payload...)
	}
	if err := rows.Err(); err != nil {
		return Snapshot{}, mapSQLError(err)
	}
	return Snapshot{SchemaVersion: meta.SchemaVersion, Buckets: buckets}, nil
}

func Save(dataDir string, snapshot Snapshot) error {
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return fmt.Errorf("create data dir: %w", err)
	}
	db, _, err := openStore(dataDir, true)
	if err != nil {
		return err
	}
	defer db.Close()
	tx, err := db.Begin()
	if err != nil {
		return mapSQLError(err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()
	now := time.Now().UTC()
	if _, err := tx.Exec(`UPDATE `+metadataTableName+` SET format_version = ?, schema_version = ?, updated_at = ? WHERE id = ?`, FormatVersion, snapshot.SchemaVersion, now.Format(time.RFC3339Nano), metadataSingleton); err != nil {
		return mapSQLError(err)
	}
	if _, err := tx.Exec(`DELETE FROM ` + bucketTableName); err != nil {
		return mapSQLError(err)
	}
	bucketNames := make([]string, 0, len(snapshot.Buckets))
	for name := range snapshot.Buckets {
		bucketNames = append(bucketNames, name)
	}
	sort.Strings(bucketNames)
	for _, name := range bucketNames {
		if _, err := tx.Exec(`INSERT INTO `+bucketTableName+` (name, payload) VALUES (?, ?)`, name, snapshot.Buckets[name]); err != nil {
			return mapSQLError(err)
		}
	}
	if err := tx.Commit(); err != nil {
		return mapSQLError(err)
	}
	committed = true
	return nil
}

func openStore(dataDir string, create bool) (*sql.DB, metadata, error) {
	path := storePath(dataDir)
	hasDB, err := pathExists(path)
	if err != nil {
		return nil, metadata{}, fmt.Errorf("stat store: %w", err)
	}
	if !hasDB && !create {
		return nil, metadata{}, ErrStoreNotFound
	}
	keyMeta, key, err := resolveStoreKeyMetadata(dataDir, hasDB, create)
	if err != nil {
		return nil, metadata{}, err
	}
	db, err := sql.Open("sqlite3", sqlcipherDSN(path, key))
	if err != nil {
		return nil, metadata{}, err
	}
	// Spec 70 §1.1: all files MUST be 0600. SQLite creates the DB before we can chmod;
	// do so immediately after opening (before configureDB writes WAL files).
	if chErr := os.Chmod(path, 0o600); chErr != nil && !os.IsNotExist(chErr) {
		_ = db.Close()
		return nil, metadata{}, fmt.Errorf("chmod state.db: %w", chErr)
	}
	if err := configureDB(db); err != nil {
		_ = db.Close()
		return nil, metadata{}, err
	}
	// WAL companion files may have been created by configureDB.
	for _, companion := range []string{path + "-wal", path + "-shm"} {
		if chErr := os.Chmod(companion, 0o600); chErr != nil && !os.IsNotExist(chErr) {
			_ = db.Close()
			return nil, metadata{}, fmt.Errorf("chmod %s: %w", companion, chErr)
		}
	}
	meta, ok, err := readMetadata(db)
	if err != nil {
		_ = db.Close()
		return nil, metadata{}, mapSQLError(err)
	}
	if !ok {
		if !create {
			_ = db.Close()
			return nil, metadata{}, ErrCorrupt
		}
		meta, err = initializeMetadata(db, keyMeta.CreatedAt)
		if err != nil {
			_ = db.Close()
			return nil, metadata{}, err
		}
	}
	return db, meta, nil
}

func configureDB(db *sql.DB) error {
	if err := db.Ping(); err != nil {
		return mapSQLError(err)
	}
	for _, stmt := range []string{
		`PRAGMA journal_mode = WAL`,
		`PRAGMA foreign_keys = ON`,
		`PRAGMA busy_timeout = 5000`,
		`CREATE TABLE IF NOT EXISTS ` + metadataTableName + ` (
			id INTEGER PRIMARY KEY CHECK (id = 1),
			format_version INTEGER NOT NULL,
			schema_version INTEGER NOT NULL,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS ` + bucketTableName + ` (
			name TEXT PRIMARY KEY,
			payload BLOB NOT NULL
		)`,
	} {
		if _, err := db.Exec(stmt); err != nil {
			return mapSQLError(err)
		}
	}
	return nil
}

func initializeMetadata(db *sql.DB, createdAt time.Time) (metadata, error) {
	now := time.Now().UTC()
	if createdAt.IsZero() {
		createdAt = now
	}
	meta := metadata{
		FormatVersion: FormatVersion,
		SchemaVersion: 0,
		CreatedAt:     createdAt,
		UpdatedAt:     now,
	}
	if _, err := db.Exec(`INSERT INTO `+metadataTableName+` (id, format_version, schema_version, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`, metadataSingleton, meta.FormatVersion, meta.SchemaVersion, meta.CreatedAt.Format(time.RFC3339Nano), meta.UpdatedAt.Format(time.RFC3339Nano)); err != nil {
		return metadata{}, mapSQLError(err)
	}
	return meta, nil
}

func readMetadata(db *sql.DB) (metadata, bool, error) {
	var (
		meta      metadata
		createdAt string
		updatedAt string
	)
	err := db.QueryRow(`SELECT format_version, schema_version, created_at, updated_at FROM `+metadataTableName+` WHERE id = ?`, metadataSingleton).Scan(&meta.FormatVersion, &meta.SchemaVersion, &createdAt, &updatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return metadata{}, false, nil
	}
	if err != nil {
		return metadata{}, false, err
	}
	if meta.FormatVersion == 0 {
		return metadata{}, false, ErrCorrupt
	}
	var parseErr error
	meta.CreatedAt, parseErr = time.Parse(time.RFC3339Nano, createdAt)
	if parseErr != nil {
		return metadata{}, false, ErrCorrupt
	}
	meta.UpdatedAt, parseErr = time.Parse(time.RFC3339Nano, updatedAt)
	if parseErr != nil {
		return metadata{}, false, ErrCorrupt
	}
	return meta, true, nil
}

func resolveStoreKeyMetadata(dataDir string, hasDB, create bool) (storeKeyMetadata, []byte, error) {
	keyMeta, err := readStoreKeyMetadata(dataDir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			if hasDB {
				return storeKeyMetadata{}, nil, ErrCorrupt
			}
			if !create {
				return storeKeyMetadata{}, nil, ErrStoreNotFound
			}
			return createStoreKeyMetadata(dataDir)
		}
		if errors.Is(err, ErrCorrupt) {
			return storeKeyMetadata{}, nil, ErrCorrupt
		}
		return storeKeyMetadata{}, nil, err
	}
	salt := decodeBase64(keyMeta.Salt)
	var key []byte
	if keyMeta.FormatVersion >= 3 {
		// §4: Argon2id KDF for format_version 3+.
		// Spec §11: derive key → verify key_check → THEN sql.Open (enforced by caller).
		secret, sErr := resolveSecret(dataDir, create)
		if sErr != nil {
			if errors.Is(sErr, fs.ErrNotExist) {
				return storeKeyMetadata{}, nil, ErrWrongKey
			}
			return storeKeyMetadata{}, nil, sErr
		}
		key = deriveStoreKeyArgon2id(salt, secret)
	} else {
		key, err = deriveStoreKey(dataDir, salt, create)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				return storeKeyMetadata{}, nil, ErrWrongKey
			}
			return storeKeyMetadata{}, nil, err
		}
	}
	if !matchesKeyCheck(key, keyMeta.KeyCheck) {
		return storeKeyMetadata{}, nil, ErrWrongKey
	}
	return keyMeta, key, nil
}

func readStoreKeyMetadata(dataDir string) (storeKeyMetadata, error) {
	raw, err := os.ReadFile(storeMetadataPath(dataDir))
	if err != nil {
		return storeKeyMetadata{}, err
	}
	var keyMeta storeKeyMetadata
	if err := json.Unmarshal(raw, &keyMeta); err != nil {
		return storeKeyMetadata{}, ErrCorrupt
	}
	if keyMeta.FormatVersion == 0 || strings.TrimSpace(keyMeta.Salt) == "" || strings.TrimSpace(keyMeta.KeyCheck) == "" || keyMeta.CreatedAt.IsZero() {
		return storeKeyMetadata{}, ErrCorrupt
	}
	if keyMeta.UpdatedAt.IsZero() {
		keyMeta.UpdatedAt = keyMeta.CreatedAt
	}
	return keyMeta, nil
}

func createStoreKeyMetadata(dataDir string) (storeKeyMetadata, []byte, error) {
	salt := make([]byte, 32)
	if _, err := rand.Read(salt); err != nil {
		return storeKeyMetadata{}, nil, fmt.Errorf("generate store salt: %w", err)
	}
	secret, err := resolveSecret(dataDir, true)
	if err != nil {
		return storeKeyMetadata{}, nil, err
	}
	// §4: new stores use Argon2id (format_version 3).
	key := deriveStoreKeyArgon2id(salt, secret)
	now := time.Now().UTC()
	keyMeta := storeKeyMetadata{
		FormatVersion:   FormatVersion,
		Salt:            base64.RawURLEncoding.EncodeToString(salt),
		KeyCheck:        keyCheck(key),
		KDF:             "argon2id",
		Argon2idMemory:  64 * 1024,
		Argon2idTime:    3,
		Argon2idThreads: 2,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := writeStoreKeyMetadata(dataDir, keyMeta); err != nil {
		return storeKeyMetadata{}, nil, err
	}
	return keyMeta, key, nil
}

func writeStoreKeyMetadata(dataDir string, keyMeta storeKeyMetadata) error {
	raw, err := json.MarshalIndent(keyMeta, "", "  ")
	if err != nil {
		return fmt.Errorf("encode store metadata: %w", err)
	}
	raw = append(raw, '\n')
	path := storeMetadataPath(dataDir)
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, raw, 0o600); err != nil {
		return fmt.Errorf("write store metadata: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("replace store metadata: %w", err)
	}
	return nil
}

func migrateLegacyStore(dataDir string) (Snapshot, bool, error) {
	if !legacyStoreExists(dataDir) {
		return Snapshot{}, false, ErrStoreNotFound
	}
	snapshot, err := loadLegacyStore(dataDir)
	if err != nil {
		return Snapshot{}, false, err
	}
	if err := Save(dataDir, snapshot); err != nil {
		return Snapshot{}, false, err
	}
	if err := archiveLegacyStore(dataDir); err != nil {
		return Snapshot{}, false, err
	}
	return snapshot, true, nil
}

func archiveLegacyStore(dataDir string) error {
	path := legacyStorePath(dataDir)
	if _, err := os.Stat(path); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("stat legacy store: %w", err)
	}
	target := filepath.Join(dataDir, StoreDirName+migrationBackupTag+time.Now().UTC().Format("20060102T150405.000000000"))
	if err := os.Rename(path, target); err != nil {
		return fmt.Errorf("archive legacy store: %w", err)
	}
	return nil
}

func sqlcipherDSN(path string, key []byte) string {
	query := url.Values{}
	query.Set("_pragma_key", fmt.Sprintf("x'%s'", hex.EncodeToString(key)))
	query.Set("_pragma_cipher_page_size", fmt.Sprintf("%d", cipherPageSize))
	query.Set("_busy_timeout", "5000")
	query.Set("_journal_mode", "WAL")
	query.Set("_foreign_keys", "on")
	return (&url.URL{Scheme: "file", Path: path, RawQuery: query.Encode()}).String()
}

func storePath(dataDir string) string {
	return filepath.Join(dataDir, StoreFileName)
}

func storeMetadataPath(dataDir string) string {
	return filepath.Join(dataDir, StoreMetaFileName)
}

func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	return false, err
}

func isEncryptedDatabase(path string) bool {
	encrypted, err := sqlcipher.IsEncrypted(path)
	return err == nil && encrypted
}

func mapSQLError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, ErrWrongKey) || errors.Is(err, ErrCorrupt) {
		return err
	}
	message := strings.ToLower(err.Error())
	for _, marker := range []string{
		"not a database",
		"malformed",
		"database disk image is malformed",
		"file is not a database",
		"file is encrypted or is not a database",
	} {
		if strings.Contains(message, marker) {
			return ErrCorrupt
		}
	}
	return err
}

// deriveStoreKeyArgon2id derives a 32-byte store key using Argon2id per spec §4.
// Parameters: m=65536 KiB, t=3, p=2.
func deriveStoreKeyArgon2id(salt, secret []byte) []byte {
	return argon2.IDKey(secret, salt, 3, 64*1024, 2, 32)
}

func deriveStoreKey(dataDir string, salt []byte, create bool) ([]byte, error) {
	secret, err := resolveSecret(dataDir, create)
	if err != nil {
		return nil, err
	}
	sum := sha256.Sum256(append(append([]byte(nil), salt...), secret...))
	key := make([]byte, len(sum))
	copy(key, sum[:])
	return key, nil
}

func resolveSecret(dataDir string, create bool) ([]byte, error) {
	if secret := strings.TrimSpace(os.Getenv(secretEnvName)); secret != "" {
		return []byte(secret), nil
	}
	path := filepath.Join(dataDir, storeKeyName)
	raw, err := os.ReadFile(path)
	if err == nil {
		decoded, err := base64.RawURLEncoding.DecodeString(strings.TrimSpace(string(raw)))
		if err != nil {
			return nil, ErrWrongKey
		}
		return decoded, nil
	}
	if !errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("read store key: %w", err)
	}
	if !create {
		return nil, err
	}
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		return nil, fmt.Errorf("generate store key: %w", err)
	}
	encoded := base64.RawURLEncoding.EncodeToString(secret)
	if err := os.WriteFile(path, []byte(encoded+"\n"), 0o600); err != nil {
		return nil, fmt.Errorf("write store key: %w", err)
	}
	return secret, nil
}

func keyCheck(key []byte) string {
	sum := sha256.Sum256(append(append([]byte(nil), key...), []byte("xorein-state-store-key-check")...))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func matchesKeyCheck(key []byte, want string) bool {
	return keyCheck(key) == want
}

func decodeBase64(raw string) []byte {
	decoded, err := base64.RawURLEncoding.DecodeString(strings.TrimSpace(raw))
	if err != nil {
		return nil
	}
	return decoded
}

// resolveStoreKeyFromMeta derives the store key from the meta file, dispatching
// to Argon2id for format_version 3+ and SHA-256 for older formats.
func resolveStoreKeyFromMeta(dataDir string, keyMeta storeKeyMetadata) ([]byte, error) {
	salt := decodeBase64(keyMeta.Salt)
	if keyMeta.FormatVersion >= 3 {
		secret, err := resolveSecret(dataDir, false)
		if err != nil {
			return nil, err
		}
		return deriveStoreKeyArgon2id(salt, secret), nil
	}
	return deriveStoreKey(dataDir, salt, false)
}

// SealPayload encrypts an opaque payload using the store secret material.
func SealPayload(dataDir string, plaintext []byte) ([]byte, error) {
	keyMeta, err := readStoreKeyMetadata(dataDir)
	if err != nil {
		return nil, err
	}
	key, err := resolveStoreKeyFromMeta(dataDir, keyMeta)
	if err != nil {
		return nil, err
	}
	return encryptLegacyBucket(key, plaintext)
}

// OpenPayload decrypts a payload previously sealed with SealPayload.
func OpenPayload(dataDir string, ciphertext []byte) ([]byte, error) {
	keyMeta, err := readStoreKeyMetadata(dataDir)
	if err != nil {
		return nil, err
	}
	key, err := resolveStoreKeyFromMeta(dataDir, keyMeta)
	if err != nil {
		return nil, err
	}
	return decryptLegacyBucket(key, ciphertext)
}
