package localapi

import "testing"

func TestValidateLocalBindRefusesNetwork(t *testing.T) {
	cfg := ListenerConfig{Network: "tcp", Address: "localhost:8000"}
	err := cfg.ValidateLocalBind()
	if err == nil {
		t.Fatalf("expected refusal for tcp")
	}

	if refusal, ok := err.(RefusalError); !ok || refusal.Reason != RefusalReasonNonLocalBind {
		t.Fatalf("expected non-local refusal, got %v", err)
	}
}

func TestValidateLocalBindAllowsUnix(t *testing.T) {
	cfg := ListenerConfig{Network: "unix", Address: "/tmp/xorein.sock"}
	if err := cfg.ValidateLocalBind(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateLocalBindRefusesOwnershipMismatch(t *testing.T) {
	expected := 1000
	actual := 1001
	cfg := ListenerConfig{
		Network:          "unix",
		Address:          "/tmp/xorein.sock",
		ExpectedOwnerUID: &expected,
		ActualOwnerUID:   &actual,
	}
	if err := cfg.ValidateLocalBind(); err == nil {
		t.Fatalf("expected ownership mismatch refusal")
	} else if refusal, ok := err.(RefusalError); !ok || refusal.Reason != RefusalReasonOwnershipMismatch {
		t.Fatalf("unexpected refusal reason: %v", err)
	}
}

func TestValidateOwnershipAllowsMatch(t *testing.T) {
	expected := 1000
	cfg := ListenerConfig{
		Network:          "unix",
		Address:          "/tmp/xorein.sock",
		ExpectedOwnerUID: &expected,
		ActualOwnerUID:   &expected,
	}
	if err := cfg.ValidateLocalBind(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
