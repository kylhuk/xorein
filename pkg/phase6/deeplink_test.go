package phase6

import "testing"

func TestParseJoinDeepLink(t *testing.T) {
	tests := []struct {
		name      string
		raw       string
		wantID    string
		wantError string
	}{
		{name: "valid link", raw: "aether://join/server123", wantID: "server123"},
		{name: "case-insensitive scheme and host", raw: "AETHER://JOIN/server123", wantID: "server123"},
		{name: "empty input", raw: "", wantError: "deeplink validation: empty deeplink"},
		{name: "wrong scheme", raw: "http://join/server", wantError: "deeplink validation: invalid scheme, expected aether"},
		{name: "wrong host", raw: "aether://connect/server", wantError: "deeplink validation: deeplink host must be join"},
		{name: "missing id", raw: "aether://join/", wantError: "deeplink validation: missing server identifier"},
		{name: "invalid id", raw: "aether://join/!!@@", wantError: "deeplink validation: server identifier invalid (alphanumeric/_/- only, 3-64 chars)"},
		{name: "query not allowed", raw: "aether://join/server123?invite=1", wantError: "deeplink validation: query parameters and fragments are not allowed"},
		{name: "fragment not allowed", raw: "aether://join/server123#frag", wantError: "deeplink validation: query parameters and fragments are not allowed"},
		{name: "userinfo not allowed", raw: "aether://user@join/server123", wantError: "deeplink validation: userinfo is not allowed"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			link, err := ParseJoinDeepLink(tc.raw)
			if tc.wantError != "" {
				if err == nil {
					t.Fatalf("expected error for %s", tc.name)
				}
				if err.Error() != tc.wantError {
					t.Fatalf("unexpected error, want %q got %q", tc.wantError, err.Error())
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if link.ServerID != tc.wantID {
				t.Fatalf("server id mismatch, want %q got %q", tc.wantID, link.ServerID)
			}
		})
	}
}
