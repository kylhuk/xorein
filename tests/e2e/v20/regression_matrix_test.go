package v20

import (
	"testing"

	"github.com/aether/code_aether/pkg/v11/relaypolicy"
	"github.com/aether/code_aether/pkg/v20/security"
)

func TestRegressionMatrixSecurity(t *testing.T) {
	profile := security.HardeningProfile{
		Name: "v20-alpha",
		Controls: []security.HardeningCheck{
			{Name: "crypto_integrity", Passed: true, Severity: security.SeverityHigh},
			{Name: "relay_policy", Passed: true, Severity: security.SeverityCritical},
		},
	}

	result := security.Evaluate(profile)
	if !result.Pass {
		t.Fatalf("expected hardening gate to pass, got %v", result.Message)
	}
}

func TestRelayNoDataRegression(t *testing.T) {
	if err := relaypolicy.ValidateMode(relaypolicy.PersistenceModeDurableMessageBody); err == nil {
		t.Fatalf("expected relay mode durable-message-body to be rejected")
	}

	if err := relaypolicy.ValidateMode(relaypolicy.PersistenceModeSessionMetadata); err != nil {
		t.Fatalf("expected session-metadata to be allowed: %v", err)
	}

	forbidden := relaypolicy.ForbiddenClasses()
	var found bool
	for _, cls := range forbidden {
		if cls == relaypolicy.StorageClassDurableMessageBody {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected durable message bodies listed as forbidden")
	}
}
