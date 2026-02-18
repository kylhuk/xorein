package v11e2e

import (
	"errors"
	"testing"

	relaypolicy "github.com/aether/code_aether/pkg/v11/relaypolicy"
)

func TestRelayPolicyBoundaries(t *testing.T) {
	t.Run("allowed-session-metadata", func(t *testing.T) {
		mode, err := relaypolicy.ParsePersistenceMode("session-metadata")
		if err != nil {
			t.Fatalf("ParsePersistenceMode() error = %v", err)
		}
		if err := relaypolicy.ValidateMode(mode); err != nil {
			t.Fatalf("ValidateMode() = %v, want nil", err)
		}
	})

	t.Run("forbidden-durable-message", func(t *testing.T) {
		mode := relaypolicy.PersistenceModeDurableMessageBody
		err := relaypolicy.ValidateMode(mode)
		if err == nil {
			t.Fatalf("ValidateMode() = nil, want ValidationError")
		}
		var policyErr *relaypolicy.ValidationError
		if !errors.As(err, &policyErr) {
			t.Fatalf("expected ValidationError, got %T", err)
		}
		if policyErr.Mode != mode {
			t.Fatalf("expected mode %q, got %q", mode, policyErr.Mode)
		}
		if len(policyErr.ForbiddenClasses) != 1 || policyErr.ForbiddenClasses[0] != relaypolicy.StorageClassDurableMessageBody {
			t.Fatalf("unexpected forbidden classes: %v", policyErr.ForbiddenClasses)
		}
	})
}
