package linkpreview

import (
	"fmt"
	"net/url"
	"sort"
	"strings"
)

// Metadata holds OpenGraph and Twitter preview fields.
type Metadata struct {
	OG      map[string]string
	Twitter map[string]string
}

// NormalizeURL produces a deterministic URL for preview eligibility.
func NormalizeURL(raw string) (string, error) {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return "", fmt.Errorf("normalize url: %w", err)
	}
	parsed.Fragment = ""
	if parsed.Scheme == "" {
		parsed.Scheme = "https"
	}
	return parsed.String(), nil
}

// PreviewEligibility checks that the URL can be previewed.
func PreviewEligibility(normalized string) bool {
	return strings.HasPrefix(normalized, "https://")
}

// MetadataPrecedence applies the OG-first policy, falling back to Twitter.
func MetadataPrecedence(meta Metadata) map[string]string {
	result := make(map[string]string)
	for k, v := range meta.Twitter {
		result[k] = v
	}
	// Override Twitter with OG values when present.
	for k, v := range meta.OG {
		result[k] = v
	}
	return result
}

// RenderState captures deterministic cache metadata.
type RenderState struct {
	URL        string
	Cached     bool
	CacheState string
}

// CacheKey returns the deterministic cache key for render state.
func CacheKey(state RenderState) string {
	parts := []string{state.URL}
	if state.Cached {
		parts = append(parts, "cached")
	}
	if state.CacheState != "" {
		parts = append(parts, state.CacheState)
	}
	sort.Strings(parts)
	return strings.Join(parts, ":")
}
