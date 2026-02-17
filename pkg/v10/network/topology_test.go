package network

import "testing"

func TestTopologyMaintainsRegionMapping(t *testing.T) {
	t.Parallel()

	nodes := []string{"bootstrap-node-a", "bootstrap-node-b", "bootstrap-node-c"}
	topo := Topology(nodes)
	if len(topo) != len(nodes) {
		t.Fatalf("expected topology entries for each node, got %d", len(topo))
	}

	continuity := ContinuityPlan()
	if got, want := len(continuity), 4; got != want {
		t.Fatalf("continuity plan length = %d, want %d", got, want)
	}
}
