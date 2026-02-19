package durability

import "testing"

func TestReplicaAccountingHealthy(t *testing.T) {
	ra := NewReplicaAccounting(3, 2, 2)
	ra.ApplyEvent(ReplicaEvent{NodeID: "node-a", Online: true})
	ra.ApplyEvent(ReplicaEvent{NodeID: "node-b", Online: true})
	ra.ApplyEvent(ReplicaEvent{NodeID: "node-c", Online: true})

	status := ra.Snapshot()
	if status.State != DurabilityStateHealthy {
		t.Fatalf("expected healthy, got %s", status.State)
	}
	if status.Reason != ReasonTargetMet {
		t.Fatalf("expected reason %s, got %s", ReasonTargetMet, status.Reason)
	}
}

func TestReplicaAccountingDegradedTargetUnmet(t *testing.T) {
	ra := NewReplicaAccounting(3, 2, 3)
	ra.ApplyEvent(ReplicaEvent{NodeID: "node-a", Online: true})
	ra.ApplyEvent(ReplicaEvent{NodeID: "node-b", Online: true})

	status := ra.Snapshot()
	if status.State != DurabilityStateDegraded {
		t.Fatalf("expected degraded, got %s", status.State)
	}
	if status.Reason != ReasonTargetUnmet {
		t.Fatalf("expected reason %s, got %s", ReasonTargetUnmet, status.Reason)
	}
}

func TestReplicaAccountingPartialFailure(t *testing.T) {
	ra := NewReplicaAccounting(3, 2, 3)
	ra.ApplyEvent(ReplicaEvent{NodeID: "node-a", Online: true})

	status := ra.Snapshot()
	if status.State != DurabilityStateUnavailable {
		t.Fatalf("expected unavailable, got %s", status.State)
	}
	if status.Reason != ReasonPartialFailure {
		t.Fatalf("expected reason %s, got %s", ReasonPartialFailure, status.Reason)
	}
}

func TestReplicaAccountingChurnDetected(t *testing.T) {
	ra := NewReplicaAccounting(2, 1, 3)
	ra.ApplyEvent(ReplicaEvent{NodeID: "node-a", Online: true})
	ra.ApplyEvent(ReplicaEvent{NodeID: "node-b", Online: true})
	// cause churn by toggling twice per node
	ra.ApplyEvent(ReplicaEvent{NodeID: "node-a", Online: false})
	ra.ApplyEvent(ReplicaEvent{NodeID: "node-a", Online: true})
	ra.ApplyEvent(ReplicaEvent{NodeID: "node-b", Online: false})
	ra.ApplyEvent(ReplicaEvent{NodeID: "node-b", Online: true})
	status := ra.Snapshot()
	if status.State != DurabilityStateDegraded {
		t.Fatalf("expected degraded on churn, got %s", status.State)
	}
	if status.Reason != ReasonChurnDetected {
		t.Fatalf("expected churn reason, got %s", status.Reason)
	}
}
