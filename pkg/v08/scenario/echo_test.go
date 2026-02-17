package scenario

import "testing"

func TestRunEchoContracts(t *testing.T) {
	t.Parallel()
	if err := RunEchoContracts(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}
