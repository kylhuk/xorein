package v24

import (
	"testing"

	"github.com/aether/code_aether/pkg/v24/localapi"
)

func TestListenerOwnershipMismatchRefusal(t *testing.T) {
	expected := 1000
	actual := 1001
	cfg := localapi.ListenerConfig{
		Network:          "unix",
		Address:          "/tmp/xorein.sock",
		ExpectedOwnerUID: &expected,
		ActualOwnerUID:   &actual,
	}
	if err := cfg.ValidateLocalBind(); err == nil {
		t.Fatalf("expected ownership mismatch refusal")
	} else if refusal, ok := err.(localapi.RefusalError); !ok || refusal.Reason != localapi.RefusalReasonOwnershipMismatch {
		t.Fatalf("unexpected refusal reason: %v", err)
	}
}
