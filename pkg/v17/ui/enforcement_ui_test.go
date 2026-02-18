package ui_test

import (
	"testing"

	"github.com/aether/code_aether/pkg/v17/ui"
)

func TestStatusSignalSummary(t *testing.T) {
	signal := ui.NewStatusSignal(ui.EnforcementModeStrict, 5, "lockdown active", false)
	summary := signal.Summary()
	expected := "[strict/5] lockdown active (unverified enforcement)"
	if summary != expected {
		t.Fatalf("unexpected summary %q", summary)
	}
}

func TestStatusSignalIsStrict(t *testing.T) {
	signal := ui.NewStatusSignal(ui.EnforcementModeRelaxed, 3, "slow mode", true)
	if signal.IsStrict() {
		t.Fatalf("relaxed should not be strict")
	}
	strict := ui.NewStatusSignal(ui.EnforcementModeStrict, 4, "ban", true)
	if !strict.IsStrict() {
		t.Fatalf("strict signal must report strict")
	}
}
