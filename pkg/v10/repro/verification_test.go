package repro

import "testing"

func TestDeterministicBuildCommandReferencesVerifier(t *testing.T) {
	t.Parallel()

	cmd := DeterministicBuildCommand()
	if cmd == "" {
		t.Fatal("expected deterministic build command to be non-empty")
	}
	if cmd != "go build ./... && go test ./... && pkg/v10/repro verify" {
		t.Fatalf("unexpected deterministic command: %q", cmd)
	}

	if len(BuildPins()) == 0 {
		t.Fatal("expected build pins to be seeded")
	}
	if len(VerificationSteps()) == 0 {
		t.Fatal("expected verification steps to be defined")
	}
}
