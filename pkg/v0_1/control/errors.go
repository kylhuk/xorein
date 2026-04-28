// Package control implements the local control API (spec 60).
package control

import (
	"encoding/json"
	"net/http"
)

const (
	CodeUnauthorized     = "unauthorized"
	CodeForbidden        = "forbidden"
	CodeMethodNotAllowed = "method_not_allowed"
	CodeInvalidRequest   = "invalid_request"
	CodeNotFound         = "not_found"
	CodeInvalidSignature = "invalid_signature"
	CodeExpiredInvite    = "expired_invite"
	CodeJoinFailed       = "join_failed"
	CodePreviewFailed    = "preview_failed"
	CodeBackupFailed     = "backup_failed"
	CodeUnsupported      = "unsupported"
)

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeError(w http.ResponseWriter, status int, code, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(apiError{Code: code, Message: msg})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
