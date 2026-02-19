package v23

import (
	"testing"

	"github.com/aether/code_aether/pkg/v23/security"
)

func TestAssistedSearchOffByDefault(t *testing.T) {
	gate := security.AssistedSearchGate{}
	if gate.Allows("consent") {
		t.Fatalf("assisted mode must be off without explicit opt-in")
	}
}

func TestAssistedSearchRequiresConsentToken(t *testing.T) {
	gate := security.AssistedSearchGate{}.Enable("token-123")
	if !gate.Allows("token-123") {
		t.Fatalf("matching consent token should enable assisted mode")
	}
	if gate.Allows("wrong") {
		t.Fatalf("non-matching token must remain blocked")
	}
}
