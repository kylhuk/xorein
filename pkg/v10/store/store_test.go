package store

import "testing"

func TestBootstrapNodesAndLabels(t *testing.T) {
	t.Parallel()

	nodes := BootstrapNodes(-1)
	if len(nodes) != 10 {
		t.Fatalf("BootstrapNodes fallback len = %d, want 10", len(nodes))
	}

	labels := NodeLabels()
	if len(labels) != len(nodes) {
		t.Fatalf("node label count = %d, want %d", len(labels), len(nodes))
	}
	if labels[nodes[0]] != "VA-N1" {
		t.Fatalf("expected label VA-N1 for %s, got %q", nodes[0], labels[nodes[0]])
	}
}
