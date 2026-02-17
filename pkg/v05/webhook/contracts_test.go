package webhook

import "testing"

func TestEndpointPolicyModeGate(t *testing.T) {
	tests := []struct {
		name            string
		auth            AuthScheme
		wantModes       []SecurityMode
		wantPlaintextOK bool
	}{
		{
			name:            "none allows plaintext",
			auth:            AuthSchemeNone,
			wantModes:       []SecurityMode{SecurityModePlain},
			wantPlaintextOK: true,
		},
		{
			name:            "hmac disallows plaintext",
			auth:            AuthSchemeHMAC,
			wantModes:       []SecurityMode{SecurityModeSigned, SecurityModeEncrypted},
			wantPlaintextOK: false,
		},
		{
			name:            "bearer disallows plaintext",
			auth:            AuthSchemeBearer,
			wantModes:       []SecurityMode{SecurityModeSigned, SecurityModeEncrypted},
			wantPlaintextOK: false,
		},
		{
			name:            "mutual denies plaintext",
			auth:            AuthSchemeMutual,
			wantModes:       []SecurityMode{SecurityModeEncrypted},
			wantPlaintextOK: false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			policy := NewEndpointPolicy(tc.auth, IdempotencyOptional, DefaultSecurityMode)
			if got := policy.AllowedSecurityModes(); !equalModes(got, tc.wantModes) {
				t.Fatalf("AllowedSecurityModes => %v, want %v", got, tc.wantModes)
			}
			if got := policy.IsPlaintextAllowed(); got != tc.wantPlaintextOK {
				t.Fatalf("IsPlaintextAllowed => %v", got)
			}
		})
	}
}

func equalModes(a, b []SecurityMode) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestRetryClassMapping(t *testing.T) {
	tests := []struct {
		name   string
		status int
		want   RetryClass
	}{
		{name: "known deferred", status: 503, want: RetryClassDeferred},
		{name: "server error", status: 550, want: RetryClassDeferred},
		{name: "client abort", status: 429, want: RetryClassDeferred},
		{name: "bad request", status: 400, want: RetryClassAbort},
		{name: "success", status: 200, want: RetryClassImmediate},
		{name: "redirect", status: 302, want: RetryClassImmediate},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if got := RetryClassForStatus(tc.status); got != tc.want {
				t.Fatalf("RetryClassForStatus(%d) => %s", tc.status, got)
			}
		})
	}
}
