package v24

import (
	"testing"

	"github.com/aether/code_aether/pkg/v24/localapi"
)

func TestLocalAPIHandshakeNonceReplay(t *testing.T) {
	middleware := localapi.NewHandshakeMiddleware("v24")
	if _, err := middleware.EstablishSession("v24", "nonce-e2e"); err != nil {
		t.Fatalf("unexpected error for first nonce: %v", err)
	}
	if _, err := middleware.EstablishSession("v24", "nonce-e2e"); err == nil {
		t.Fatalf("expected nonce replay refusal")
	} else if refusal, ok := err.(localapi.RefusalError); !ok || refusal.Reason != localapi.RefusalReasonNonceReplay {
		t.Fatalf("unexpected refusal reason: %v", err)
	}
}

func TestLocalAPIHandshakeDowngrade(t *testing.T) {
	middleware := localapi.NewHandshakeMiddlewareWithMinVersion("v24", "v24")
	if _, err := middleware.EstablishSession("v23", "nonce-downgrade"); err == nil {
		t.Fatalf("expected version downgrade refusal")
	} else if refusal, ok := err.(localapi.RefusalError); !ok || refusal.Reason != localapi.RefusalReasonVersionDowngrade {
		t.Fatalf("unexpected refusal reason: %v", err)
	}
}
