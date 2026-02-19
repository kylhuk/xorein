package v24

import (
	"strings"
	"testing"
	"time"

	"github.com/aether/code_aether/pkg/v24/localapi"
)

func TestNonLocalBindRefusal(t *testing.T) {
	cfg := localapi.ListenerConfig{Network: "tcp", Address: "127.0.0.1:9999"}
	if err := cfg.ValidateLocalBind(); err == nil {
		t.Fatalf("expected non-local refusal")
	} else if refusal, ok := err.(localapi.RefusalError); !ok || refusal.Reason != localapi.RefusalReasonNonLocalBind {
		t.Fatalf("unexpected refusal reason: %v", err)
	}
}

func TestTokenCapabilityGating(t *testing.T) {
	middleware := localapi.NewHandshakeMiddleware("v1")
	if _, err := middleware.EstablishSession("v2", "nonce-auth"); err == nil {
		t.Fatalf("expected unauthorized capability refusal")
	}
	if _, err := middleware.AuthenticateSession("missing"); err == nil {
		t.Fatalf("expected invalid token refusal")
	}
}

func TestAuditLogSanitized(t *testing.T) {
	record := localapi.AuditRecord{
		Timestamp: time.Now(),
		RPC:       "Stream",
		Reason:    localapi.RefusalReasonUnauthorizedCapability,
		Outcome:   "denied",
	}
	summary := record.Summary()
	if strings.Contains(summary, "payload") {
		t.Fatalf("audit summary contains payload hint: %s", summary)
	}
}
