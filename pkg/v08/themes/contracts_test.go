package themes

import (
	"encoding/json"
	"testing"
)

func TestValidateToken(t *testing.T) {
	cases := []struct {
		token string
		want  bool
	}{
		{"#fff", true},
		{"#1234", true},
		{"fff", false},
		{"#1", false},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.token, func(t *testing.T) {
			t.Parallel()
			if got := ValidateToken(tc.token); got != tc.want {
				t.Fatalf("token %q expected %t", tc.token, tc.want)
			}
		})
	}
}

func TestValidateCustomTheme(t *testing.T) {
	valid := []byte(`{"name":"echo","tokens":{"background":"#112233","text":"#ffffff"}}`)
	def, err := ValidateCustomTheme(valid, "night")
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	if def.Name != "echo" {
		t.Fatalf("expected echo, got %s", def.Name)
	}

	invalidJSON := []byte(`#invalid`)
	fallback1, err := ValidateCustomTheme(invalidJSON, "night")
	if err == nil || fallback1.Name != "night" {
		t.Fatalf("expected fallback night, got %s err %v", fallback1.Name, err)
	}

	invalidToken := []byte(`{"name":"echo","tokens":{"background":"rgb(0,0,0)"}}`)
	fallback2, err := ValidateCustomTheme(invalidToken, "night")
	if err == nil || fallback2.Name != "night" {
		t.Fatalf("expected invalid token fallback, got %s err %v", fallback2.Name, err)
	}

	emptyTokens := ThemeDefinition{Name: "empty", Tokens: map[string]string{}}
	emptyData, _ := json.Marshal(emptyTokens)
	fallback3, err := ValidateCustomTheme(emptyData, "night")
	if err != nil {
		t.Fatalf("expected tokens to fallback, got %v", err)
	}
	if fallback3.Name != "empty" {
		t.Fatalf("expected empty name preserved, got %s", fallback3.Name)
	}
	if len(fallback3.Tokens) == 0 {
		t.Fatalf("fallback tokens should be populated")
	}

	unknownFallback, err := ValidateCustomTheme(invalidJSON, "unknown")
	if err == nil || unknownFallback.Name != "dawn" {
		t.Fatalf("expected dawn fallback, got %s err %v", unknownFallback.Name, err)
	}
}
