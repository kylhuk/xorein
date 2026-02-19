package limits

import (
	"testing"
	"time"
)

func TestRequestBudgetNormalUsage(t *testing.T) {
	budget := NewRequestBudget(ScopeBackfillVerification, BackfillVerificationBudget)
	if err := budget.ConsumeCPU(100 * time.Millisecond); err != nil {
		t.Fatalf("unexpected CPU refusal: %v", err)
	}
	if err := budget.ConsumeCPU(120 * time.Millisecond); err != nil {
		t.Fatalf("unexpected CPU refusal: %v", err)
	}
	if err := budget.ConsumeIO(8 * 1024 * 1024); err != nil {
		t.Fatalf("unexpected IO refusal: %v", err)
	}
	if err := budget.ConsumeIO(7 * 1024 * 1024); err != nil {
		t.Fatalf("unexpected IO refusal: %v", err)
	}
}

func TestRequestBudgetCPURefusal(t *testing.T) {
	budget := NewRequestBudget(ScopeIndexing, ScopeLimits{CPULimit: 30 * time.Millisecond, IOLimitBytes: 1})
	if err := budget.ConsumeCPU(20 * time.Millisecond); err != nil {
		t.Fatalf("setup consumption failed: %v", err)
	}
	err := budget.ConsumeCPU(20 * time.Millisecond)
	refusal := expectRefusal(t, err, CodeCPUExceeded, ScopeIndexing, remediationCPU)
	if refusal.Message == "" {
		t.Fatalf("expected detail message on refusal")
	}
}

func TestRequestBudgetIORefusal(t *testing.T) {
	budget := NewRequestBudget(ScopeBackfillVerification, ScopeLimits{CPULimit: time.Second, IOLimitBytes: 100})
	err := budget.ConsumeIO(120)
	expectRefusal(t, err, CodeIOExceeded, ScopeBackfillVerification, remediationIO)
}

func TestDiskGuardAlarmAndRefusal(t *testing.T) {
	guard, err := NewDiskGuard(100, 200, 0)
	if err != nil {
		t.Fatalf("failed to create guard: %v", err)
	}

	state, err := guard.AddUsage(120)
	if err != nil {
		t.Fatalf("expected no refusal before hard limit: %v", err)
	}
	if !state.Alarmed {
		t.Fatalf("expected alarm after crossing threshold")
	}
	if state.UsageBytes != 120 {
		t.Fatalf("unexpected usage: got %d", state.UsageBytes)
	}

	state, err = guard.AddUsage(50)
	if err != nil {
		t.Fatalf("unexpected refusal while below hard limit: %v", err)
	}
	if state.UsageBytes != 170 {
		t.Fatalf("bad tracked usage: %d", state.UsageBytes)
	}

	state, err = guard.AddUsage(40)
	refusal := expectRefusal(t, err, CodeDiskHardLimit, ScopeDiskGrowth, remediationDisk)
	if refusal.Message == "" {
		t.Fatalf("expected detail message on disk refusal")
	}
	if state.UsageBytes != 200 {
		t.Fatalf("usage should be clamped to hard limit; got %d", state.UsageBytes)
	}

	if _, err := guard.AddUsage(10); err == nil {
		t.Fatalf("expected failure after hard limit was hit")
	} else {
		expectRefusal(t, err, CodeDiskHardLimit, ScopeDiskGrowth, remediationDisk)
	}
}

func expectRefusal(t *testing.T, err error, code ReasonCode, scope Scope, remediation string) *Refusal {
	t.Helper()
	if err == nil {
		t.Fatalf("expected refusal %s/%s", code, scope)
	}
	refusal, ok := err.(*Refusal)
	if !ok {
		t.Fatalf("expected Refusal but got %T", err)
	}
	if refusal.Code != code {
		t.Fatalf("code mismatch: want %s got %s", code, refusal.Code)
	}
	if refusal.Scope != scope {
		t.Fatalf("scope mismatch: want %s got %s", scope, refusal.Scope)
	}
	if refusal.Remediation != remediation {
		t.Fatalf("remediation mismatch: got %s", refusal.Remediation)
	}
	return refusal
}
