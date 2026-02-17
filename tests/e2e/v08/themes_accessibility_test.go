package v08e2e

import (
	"testing"

	"github.com/aether/code_aether/pkg/v08/accessibility"
	"github.com/aether/code_aether/pkg/v08/themes"
)

func TestThemesAccessibilityFlow(t *testing.T) {
	def, err := themes.ValidateCustomTheme([]byte(`{"name":"v08","tokens":{"background":"#101010","text":"#f0f0f0"}}`), "night")
	if err != nil {
		t.Fatalf("unexpected theme error %v", err)
	}
	if def.Name != "v08" {
		t.Fatalf("unexpected theme name %s", def.Name)
	}

	if got := accessibility.HighContrastToken(true, def.Tokens["text"]); got != "#000000" {
		t.Fatalf("high contrast override missing")
	}

	graph := accessibility.FocusGraph{}
	graph.AddEdge("primary", "secondary")
	graph.AddEdge("primary", "tertiary")
	if len(graph.Neighbors("primary")) != 2 {
		t.Fatalf("focus graph missing edges")
	}

	// Invalid theme triggers fallback
	fallback, err := themes.ValidateCustomTheme([]byte(`{"name":"none","tokens":{"text":"bad"}}`), "dawn")
	if err == nil {
		t.Fatalf("expected error for invalid token")
	}
	if fallback.Name != "dawn" {
		t.Fatalf("expected dawn fallback, got %s", fallback.Name)
	}
}
