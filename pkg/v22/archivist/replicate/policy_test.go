package replicate

import "testing"

func TestReplicateSuccess(t *testing.T) {
	policy := Policy{R: 2, RMin: 2}
	result := Replicate(policy, []EndpointID{"a", "b"}, func(EndpointID) error { return nil })
	if !result.TargetMet || result.SuccessCount != 2 || result.Health != HealthDurable {
		t.Fatalf("expected durable target met, got %+v", result)
	}
}

func TestReplicatePartial(t *testing.T) {
	policy := Policy{R: 3, RMin: 2}
	attempts := 0
	result := Replicate(policy, []EndpointID{"a", "b", "c"}, func(id EndpointID) error {
		attempts++
		if id == "c" {
			return errDummy
		}
		return nil
	})

	if result.SuccessCount != 2 || result.Reason != ResultReplicaWritePartial || result.Health != HealthDegraded {
		t.Fatalf("unexpected partial result %+v", result)
	}
	if attempts != 3 {
		t.Fatalf("expected all endpoints attempted")
	}
}

func TestReplicateTargetUnmet(t *testing.T) {
	policy := Policy{R: 3, RMin: 2}
	result := Replicate(policy, []EndpointID{"x"}, func(_ EndpointID) error { return errDummy })
	if result.Reason != ResultReplicaTargetUnmet || result.TargetMet {
		t.Fatalf("unexpected result %+v", result)
	}
}

func TestHealRecoversMissing(t *testing.T) {
	policy := Policy{R: 3, RMin: 2}
	known := []EndpointID{"a"}
	candidates := []EndpointID{"b", "c"}
	successes := 0
	result := Heal(policy, known, candidates, func(id EndpointID) error {
		if id == "b" {
			successes++
			return nil
		}
		return errDummy
	})

	if len(result.HealedTokens) != 1 || result.HealedTokens[0] != "b" {
		t.Fatalf("expected healing for b, got %+v", result)
	}
	if result.Reason != ResultReplicaHealingInProgress {
		t.Fatalf("unexpected reason %s", result.Reason)
	}
	if result.SuccessTotal != 2 {
		t.Fatalf("expected success total 2, got %d", result.SuccessTotal)
	}
	if successes != 1 {
		t.Fatalf("writer should have run once")
	}
}

func TestHealTargetUnmet(t *testing.T) {
	policy := Policy{R: 3, RMin: 3}
	known := []EndpointID{}
	candidates := []EndpointID{"x"}
	result := Heal(policy, known, candidates, func(EndpointID) error { return errDummy })
	if result.Reason != ResultReplicaTargetUnmet {
		t.Fatalf("expected target unmet reason, got %s", result.Reason)
	}
}

func TestHealNoWritesTargetUnmet(t *testing.T) {
	policy := Policy{R: 3, RMin: 2}
	known := []EndpointID{"a"}
	candidates := []EndpointID{"b", "c"}
	attempts := 0
	result := Heal(policy, known, candidates, func(EndpointID) error {
		attempts++
		return errDummy
	})
	if len(result.HealedTokens) != 0 {
		t.Fatalf("expected no healed tokens, got %+v", result.HealedTokens)
	}
	if result.Reason != ResultReplicaTargetUnmet {
		t.Fatalf("expected target unmet reason, got %s", result.Reason)
	}
	if result.SuccessTotal != 1 {
		t.Fatalf("expected success total 1, got %d", result.SuccessTotal)
	}
	if attempts != len(candidates) {
		t.Fatalf("expected attempts on all candidates, got %d", attempts)
	}
}

var errDummy = errorString("dummy")

type errorString string

func (e errorString) Error() string { return string(e) }
