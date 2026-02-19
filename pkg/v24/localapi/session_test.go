package localapi

import "testing"

func TestHandshakeVersionMismatchRefusal(t *testing.T) {
	middleware := NewHandshakeMiddleware("v1")
	if _, err := middleware.EstablishSession("v2", "nonce-a"); err == nil {
		t.Fatalf("expected version mismatch refusal")
	} else if refusal, ok := err.(RefusalError); !ok || refusal.Reason != RefusalReasonUnauthorizedCapability {
		t.Fatalf("unexpected refusal reason: %v", err)
	}
}

func TestHandshakeVersionDowngradeRefusal(t *testing.T) {
	middleware := NewHandshakeMiddlewareWithMinVersion("v2", "v2")
	if _, err := middleware.EstablishSession("v1", "nonce-down"); err == nil {
		t.Fatalf("expected version downgrade refusal")
	} else if refusal, ok := err.(RefusalError); !ok || refusal.Reason != RefusalReasonVersionDowngrade {
		t.Fatalf("unexpected refusal reason: %v", err)
	}
}

func TestHandshakeNonceReplayRefusal(t *testing.T) {
	middleware := NewHandshakeMiddleware("v1")
	if _, err := middleware.EstablishSession("v1", "nonce-replay"); err != nil {
		t.Fatalf("unexpected error for first nonce: %v", err)
	}
	if _, err := middleware.EstablishSession("v1", "nonce-replay"); err == nil {
		t.Fatalf("expected nonce replay refusal")
	} else if refusal, ok := err.(RefusalError); !ok || refusal.Reason != RefusalReasonNonceReplay {
		t.Fatalf("unexpected refusal reason: %v", err)
	}
}

func TestAuthenticateMissingToken(t *testing.T) {
	middleware := NewHandshakeMiddleware("v1")
	if _, err := middleware.AuthenticateSession("missing"); err == nil {
		t.Fatalf("expected invalid token refusal")
	}
}
