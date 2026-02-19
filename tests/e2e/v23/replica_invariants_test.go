package v23

import (
	"testing"

	"github.com/aether/code_aether/pkg/v23/durability"
)

func TestReplicaChurnAndPartialFailure(t *testing.T) {
	ra := durability.NewReplicaAccounting(3, 2, 2)
	// bring up three replicas so target is satisfied
	ra.ApplyEvent(durability.ReplicaEvent{NodeID: "A", Online: true})
	ra.ApplyEvent(durability.ReplicaEvent{NodeID: "B", Online: true})
	ra.ApplyEvent(durability.ReplicaEvent{NodeID: "C", Online: true})

	// simulate churn by toggling the first replica twice
	ra.ApplyEvent(durability.ReplicaEvent{NodeID: "A", Online: false})
	ra.ApplyEvent(durability.ReplicaEvent{NodeID: "A", Online: true})
	status := ra.Snapshot()
	if status.State != durability.DurabilityStateDegraded {
		t.Fatalf("expected degraded while churn occurs, got %s", status.State)
	}
	if status.Reason != durability.ReasonChurnDetected {
		t.Fatalf("expected churn reason, got %s", status.Reason)
	}

	// simulate partial failure by dropping below target but staying at minimum acceptable replicas
	ra.ApplyEvent(durability.ReplicaEvent{NodeID: "B", Online: false})
	status = ra.Snapshot()
	if status.State != durability.DurabilityStateDegraded {
		t.Fatalf("expected degraded when still at minimum replicas, got %s", status.State)
	}
	if status.Reason != durability.ReasonTargetUnmet {
		t.Fatalf("expected target unmet reason when at min replicas, got %s", status.Reason)
	}
}

func TestReplicaTargetUnmetButServiceFunctional(t *testing.T) {
	ra := durability.NewReplicaAccounting(4, 2, 5)
	ra.ApplyEvent(durability.ReplicaEvent{NodeID: "A", Online: true})
	ra.ApplyEvent(durability.ReplicaEvent{NodeID: "B", Online: true})
	status := ra.Snapshot()
	if status.State != durability.DurabilityStateDegraded {
		t.Fatalf("expected degraded state, got %s", status.State)
	}
	if status.Reason != durability.ReasonTargetUnmet {
		t.Fatalf("expected target unmet reason, got %s", status.Reason)
	}
	if status.ReadyReplicas < status.MinReplicas {
		t.Fatalf("ready replicas %d are below minimum %d", status.ReadyReplicas, status.MinReplicas)
	}
}
