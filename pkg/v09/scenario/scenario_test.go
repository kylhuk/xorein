package scenario

import "testing"

func TestRunForgeScenario(t *testing.T) {
	t.Parallel()
	if err := RunForgeScenario(); err != nil {
		t.Fatalf("expected forge scenario to pass, got %v", err)
	}
}
