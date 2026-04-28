package control

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"
	"net/http"
	"time"

	"golang.org/x/crypto/argon2"
)

type identityResponse struct {
	PeerID      string    `json:"peer_id"`
	DisplayName string    `json:"display_name,omitempty"`
	Bio         string    `json:"bio,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

func (s *Server) handleGetIdentities(w http.ResponseWriter, r *http.Request) {
	resp := identityResponse{
		PeerID:      s.hs.PeerID,
		DisplayName: s.hs.DisplayName,
		CreatedAt:   time.Time{},
	}
	writeJSON(w, http.StatusOK, resp)
}

type createIdentityRequest struct {
	DisplayName string `json:"display_name"`
	Bio         string `json:"bio,omitempty"`
}

func (s *Server) handleCreateIdentity(w http.ResponseWriter, r *http.Request) {
	var req createIdentityRequest
	if !decodeBody(w, r, &req) {
		return
	}
	if req.DisplayName == "" {
		writeError(w, http.StatusBadRequest, CodeInvalidRequest, "display_name is required")
		return
	}
	s.hs.DisplayName = req.DisplayName
	resp := identityResponse{
		PeerID:      s.hs.PeerID,
		DisplayName: req.DisplayName,
		Bio:         req.Bio,
		CreatedAt:   timeNow(),
	}
	writeJSON(w, http.StatusCreated, resp)
}

// backupDocument is the on-disk format for an identity backup.
type backupDocument struct {
	Version    int    `json:"version"`
	Alg        string `json:"alg"`
	PeerID     string `json:"peer_id"`
	Salt       string `json:"salt"`       // base64std
	Nonce      string `json:"nonce"`      // base64std
	Ciphertext string `json:"ciphertext"` // base64std
}

type backupRequest struct {
	Passphrase string `json:"passphrase"`
}

type restoreRequest struct {
	Passphrase string         `json:"passphrase"`
	Backup     backupDocument `json:"backup"`
}

// deriveBackupKey derives a 32-byte AES key from passphrase + salt using Argon2id.
func deriveBackupKey(passphrase string, salt []byte) []byte {
	return argon2.IDKey([]byte(passphrase), salt, 3, 64*1024, 2, 32)
}

func (s *Server) handleBackupIdentity(w http.ResponseWriter, r *http.Request) {
	var req backupRequest
	if !decodeBody(w, r, &req) {
		return
	}
	if req.Passphrase == "" {
		writeError(w, http.StatusBadRequest, CodeInvalidRequest, "passphrase is required")
		return
	}
	if s.hs == nil || s.hs.BackupKeyFn == nil {
		writeError(w, http.StatusInternalServerError, CodeBackupFailed, "identity backup not wired in this runtime")
		return
	}
	privKey, err := s.hs.BackupKeyFn()
	if err != nil {
		writeError(w, http.StatusInternalServerError, CodeBackupFailed, err.Error())
		return
	}

	salt := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		writeError(w, http.StatusInternalServerError, CodeBackupFailed, "rng failure")
		return
	}
	key := deriveBackupKey(req.Passphrase, salt)
	block, err := aes.NewCipher(key)
	if err != nil {
		writeError(w, http.StatusInternalServerError, CodeBackupFailed, "cipher init failed")
		return
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		writeError(w, http.StatusInternalServerError, CodeBackupFailed, "gcm init failed")
		return
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		writeError(w, http.StatusInternalServerError, CodeBackupFailed, "rng failure")
		return
	}
	ct := gcm.Seal(nil, nonce, privKey, nil)
	doc := backupDocument{
		Version:    1,
		Alg:        "argon2id-aes256gcm",
		PeerID:     s.hs.PeerID,
		Salt:       base64.StdEncoding.EncodeToString(salt),
		Nonce:      base64.StdEncoding.EncodeToString(nonce),
		Ciphertext: base64.StdEncoding.EncodeToString(ct),
	}
	writeJSON(w, http.StatusOK, doc)
}

func (s *Server) handleRestoreIdentity(w http.ResponseWriter, r *http.Request) {
	var req restoreRequest
	if !decodeBody(w, r, &req) {
		return
	}
	if req.Passphrase == "" {
		writeError(w, http.StatusBadRequest, CodeInvalidRequest, "passphrase is required")
		return
	}
	if req.Backup.Version == 0 {
		writeError(w, http.StatusBadRequest, CodeInvalidRequest, "backup document required (must include backup.version)")
		return
	}
	if req.Backup.Alg != "argon2id-aes256gcm" {
		writeError(w, http.StatusBadRequest, CodeInvalidRequest, "unsupported backup algorithm: "+req.Backup.Alg)
		return
	}
	salt, err := base64.StdEncoding.DecodeString(req.Backup.Salt)
	if err != nil {
		writeError(w, http.StatusBadRequest, CodeInvalidRequest, "invalid salt encoding")
		return
	}
	nonce, err := base64.StdEncoding.DecodeString(req.Backup.Nonce)
	if err != nil {
		writeError(w, http.StatusBadRequest, CodeInvalidRequest, "invalid nonce encoding")
		return
	}
	ct, err := base64.StdEncoding.DecodeString(req.Backup.Ciphertext)
	if err != nil {
		writeError(w, http.StatusBadRequest, CodeInvalidRequest, "invalid ciphertext encoding")
		return
	}
	key := deriveBackupKey(req.Passphrase, salt)
	block, err := aes.NewCipher(key)
	if err != nil {
		writeError(w, http.StatusInternalServerError, CodeBackupFailed, "cipher init failed")
		return
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		writeError(w, http.StatusInternalServerError, CodeBackupFailed, "gcm init failed")
		return
	}
	privKey, err := gcm.Open(nil, nonce, ct, nil)
	if err != nil {
		writeError(w, http.StatusBadRequest, CodeInvalidSignature, "wrong passphrase or corrupt backup")
		return
	}
	if s.hs == nil || s.hs.RestoreKeyFn == nil {
		writeError(w, http.StatusInternalServerError, CodeBackupFailed, "identity restore not wired in this runtime")
		return
	}
	if err := s.hs.RestoreKeyFn(privKey); err != nil {
		writeError(w, http.StatusInternalServerError, CodeBackupFailed, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "peer_id": req.Backup.PeerID})
}
