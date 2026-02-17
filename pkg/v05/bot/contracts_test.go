package bot

import "testing"

func TestDefaultSecurityGating(t *testing.T) {
	if DefaultSecurityMode != SecurityModeClear {
		t.Fatalf("expected default security mode clear, got %s", DefaultSecurityMode)
	}

	if !AllowsMode(DefaultSecurityMode) {
		t.Fatal("default security mode should be allowed")
	}

	if AllowsMode(SecurityModeEncrypted) {
		t.Fatal("encrypted mode is not in the default whitelist")
	}
}

func TestAllowsModeVariants(t *testing.T) {
	cases := []struct {
		name     string
		mode     SecurityMode
		expected bool
	}{
		{name: "clear", mode: SecurityModeClear, expected: true},
		{name: "encrypted", mode: SecurityModeEncrypted, expected: false},
		{name: "deferred", mode: SecurityModeDeferred, expected: false},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if got := AllowsMode(tc.mode); got != tc.expected {
				t.Fatalf("AllowsMode(%s) => %v, want %v", tc.mode, got, tc.expected)
			}
		})
	}
}

func TestCommandLifecycleTrust(t *testing.T) {
	cases := []struct {
		name     string
		auth     AuthOutcome
		expected TrustClass
	}{
		{name: "success", auth: AuthSuccess, expected: TrustClassHigh},
		{name: "challenge", auth: AuthChallenge, expected: TrustClassMedium},
		{name: "failure", auth: AuthFailure, expected: TrustClassLow},
		{name: "not attempted", auth: AuthNotAttempted, expected: TrustClassUnknown},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			lifecycle := CommandLifecycle{Stage: StageExecuting, Reason: ReasonCommandInvoked, Auth: tc.auth}
			if got := lifecycle.Trust(); got != tc.expected {
				t.Fatalf("Trust() => %s, want %s", got, tc.expected)
			}
		})
	}
}

func TestCommandLifecycleStages(t *testing.T) {
	lifecycle := CommandLifecycle{Stage: StageCompleted, Auth: AuthSuccess}
	if !lifecycle.IsTerminal() {
		t.Fatal("completed stage should be terminal")
	}

	lifecycle.Stage = StageExecuting
	if lifecycle.IsTerminal() {
		t.Fatal("non-completed stage should not be terminal")
	}

	if !lifecycle.IsPlaintextOnly() {
		t.Fatal("IsPlaintextOnly should follow default gating")
	}
}

func TestNormalizeReason(t *testing.T) {
	if got := NormalizeReason(ReasonAuthMismatch); got != "reason.auth.mismatch" {
		t.Fatalf("NormalizeReason => %s", got)
	}
}
