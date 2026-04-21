package storage

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	legacyStoreMetaName   = "meta.json"
	legacyBucketExtension = ".bin"
)

type legacyMetadata struct {
	FormatVersion      int    `json:"format_version"`
	SchemaVersion      int    `json:"schema_version"`
	Salt               string `json:"salt"`
	KeyCheck           string `json:"key_check"`
	CurrentGeneration  string `json:"current_generation"`
	PreviousGeneration string `json:"previous_generation,omitempty"`
}

type legacyBucketEnvelope struct {
	Nonce      string `json:"nonce"`
	Ciphertext string `json:"ciphertext"`
}

func legacyStoreExists(dataDir string) bool {
	path := legacyStorePath(dataDir)
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func loadLegacyStore(dataDir string) (Snapshot, error) {
	meta, key, err := openLegacyMetadata(dataDir)
	if err != nil {
		return Snapshot{}, err
	}
	storeDir := legacyStorePath(dataDir)
	generation, ok, err := resolveLegacyGeneration(storeDir, meta.CurrentGeneration)
	if err != nil {
		return Snapshot{}, err
	}
	if !ok {
		return Snapshot{SchemaVersion: meta.SchemaVersion, Buckets: map[string][]byte{}}, nil
	}
	currentDir := filepath.Join(storeDir, generation)
	entries, err := os.ReadDir(currentDir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return Snapshot{SchemaVersion: meta.SchemaVersion, Buckets: map[string][]byte{}}, nil
		}
		return Snapshot{}, fmt.Errorf("read current generation: %w", err)
	}
	buckets := make(map[string][]byte, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), legacyBucketExtension) {
			continue
		}
		bucketName := strings.TrimSuffix(entry.Name(), legacyBucketExtension)
		raw, err := os.ReadFile(filepath.Join(currentDir, entry.Name()))
		if err != nil {
			return Snapshot{}, fmt.Errorf("read bucket %s: %w", bucketName, err)
		}
		plaintext, err := decryptLegacyBucket(key, raw)
		if err != nil {
			if errors.Is(err, ErrWrongKey) {
				return Snapshot{}, err
			}
			return Snapshot{}, ErrCorrupt
		}
		buckets[bucketName] = plaintext
	}
	return Snapshot{SchemaVersion: meta.SchemaVersion, Buckets: buckets}, nil
}

func openLegacyMetadata(dataDir string) (legacyMetadata, []byte, error) {
	storeDir := legacyStorePath(dataDir)
	raw, err := os.ReadFile(filepath.Join(storeDir, legacyStoreMetaName))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			if hasLegacyPendingGenerations(storeDir) {
				return legacyMetadata{}, nil, ErrCorrupt
			}
			return legacyMetadata{}, nil, ErrStoreNotFound
		}
		return legacyMetadata{}, nil, fmt.Errorf("read store metadata: %w", err)
	}
	var meta legacyMetadata
	if err := json.Unmarshal(raw, &meta); err != nil {
		return legacyMetadata{}, nil, ErrCorrupt
	}
	if meta.FormatVersion == 0 || meta.Salt == "" || meta.KeyCheck == "" {
		return legacyMetadata{}, nil, ErrCorrupt
	}
	key, err := deriveStoreKey(dataDir, decodeBase64(meta.Salt), false)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return legacyMetadata{}, nil, ErrWrongKey
		}
		return legacyMetadata{}, nil, err
	}
	if !matchesKeyCheck(key, meta.KeyCheck) {
		return legacyMetadata{}, nil, ErrWrongKey
	}
	return meta, key, nil
}

func resolveLegacyGeneration(storeDir, current string) (string, bool, error) {
	generations, err := legacyGenerationDirectories(storeDir)
	if err != nil {
		return "", false, err
	}
	if len(generations) > 0 {
		return generations[len(generations)-1], true, nil
	}
	current = strings.TrimSpace(current)
	if current != "" {
		return current, true, nil
	}
	return "", false, nil
}

func legacyGenerationDirectories(storeDir string) ([]string, error) {
	entries, err := os.ReadDir(storeDir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("read generation directories: %w", err)
	}
	generations := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasPrefix(name, ".") || !strings.HasPrefix(name, "gen-") {
			continue
		}
		generations = append(generations, name)
	}
	sort.Strings(generations)
	return generations, nil
}

func hasLegacyPendingGenerations(storeDir string) bool {
	entries, err := os.ReadDir(storeDir)
	if err != nil {
		return false
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".gen-") && strings.HasSuffix(entry.Name(), ".pending") {
			return true
		}
	}
	return false
}

func decryptLegacyBucket(key []byte, raw []byte) ([]byte, error) {
	var envelope legacyBucketEnvelope
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return nil, ErrCorrupt
	}
	nonce := decodeBase64(envelope.Nonce)
	ciphertext := decodeBase64(envelope.Ciphertext)
	if len(nonce) == 0 || len(ciphertext) == 0 {
		return nil, ErrCorrupt
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	plaintext, err := aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, ErrCorrupt
	}
	return plaintext, nil
}

func encryptLegacyBucket(key []byte, plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}
	envelope := legacyBucketEnvelope{
		Nonce:      base64.RawURLEncoding.EncodeToString(nonce),
		Ciphertext: base64.RawURLEncoding.EncodeToString(aead.Seal(nil, nonce, plaintext, nil)),
	}
	return json.Marshal(envelope)
}

func legacyStorePath(dataDir string) string {
	return filepath.Join(dataDir, StoreDirName)
}
