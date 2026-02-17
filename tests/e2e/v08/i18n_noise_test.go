package v08e2e

import (
	"fmt"
	"strings"
	"testing"

	"github.com/aether/code_aether/pkg/v08/i18n"
	"github.com/aether/code_aether/pkg/v08/voice"
)

func TestI18nNoiseFlow(t *testing.T) {
	if got := i18n.LocaleFallback("es-MX"); got != "es-ES" {
		t.Fatalf("expected es-ES fallback, got %s", got)
	}
	msg := i18n.MissingKeyMessage("de-DE", "welcome")
	if !strings.Contains(msg, "[en-US]") {
		t.Fatalf("missing fallback locale in message: %s", msg)
	}

	if got := i18n.FormatLocalizedNumber("fr-CA", 10); got != "fr-FR:10" {
		t.Fatalf("unexpected formatted number %s", got)
	}

	decision := voice.SelectNoiseReducer(false, fmt.Errorf("signal drop"))
	if decision.Selected != voice.DTLN {
		t.Fatalf("expected dtln fallback, got %s", decision.Selected)
	}
	if !strings.Contains(decision.Reason, "signal drop") {
		t.Fatalf("expected fallback reason to contain error, got %s", decision.Reason)
	}
}
