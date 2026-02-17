package linkpreview

import (
	"reflect"
	"testing"
)

func TestNormalizeURL(t *testing.T) {
	cases := []struct {
		name    string
		raw     string
		want    string
		wantErr bool
	}{
		{"already https", "https://example.com", "https://example.com", false},
		{"missing scheme", "example.com/path", "https://example.com/path", false},
		{"with fragment", "https://example.com/page#section", "https://example.com/page", false},
		{"invalid", "://bad://", "", true},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := NormalizeURL(tc.raw)
			if (err != nil) != tc.wantErr {
				t.Fatalf("expected error=%t, got %v", tc.wantErr, err)
			}
			if got != tc.want {
				t.Fatalf("expected %q, got %q", tc.want, got)
			}
		})
	}
}

func TestPreviewEligibility(t *testing.T) {
	if !PreviewEligibility("https://foo") {
		t.Fatalf("expected https to be eligible")
	}
	if PreviewEligibility("http://foo") {
		t.Fatalf("http should be ineligible")
	}
}

func TestMetadataPrecedence(t *testing.T) {
	meta := Metadata{
		OG:      map[string]string{"title": "og"},
		Twitter: map[string]string{"title": "tw", "description": "desc"},
	}
	got := MetadataPrecedence(meta)
	want := map[string]string{"title": "og", "description": "desc"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected precedence map: %v", got)
	}
}

func TestCacheKey(t *testing.T) {
	state := RenderState{URL: "https://example", Cached: true, CacheState: "warm"}
	if got := CacheKey(state); got != "cached:https://example:warm" {
		t.Fatalf("unexpected key %s", got)
	}
	if CacheKey(RenderState{URL: "https://example"}) != "https://example" {
		t.Fatalf("unexpected cache key without extras")
	}
}
