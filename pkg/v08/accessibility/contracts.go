package accessibility

import (
	"sync"
	"time"
)

// SupportedRoles enumerates deterministic role labels.
var SupportedRoles = map[string]struct{}{
	"button":     {},
	"link":       {},
	"dialog":     {},
	"main":       {},
	"navigation": {},
}

// ValidLabel returns true when the label is recognized by the contract.
func ValidLabel(label string) bool {
	_, ok := SupportedRoles[label]
	return ok
}

// Announcer tracks deterministic announcements to throttle duplicates.
type Announcer struct {
	mu    sync.Mutex
	seen  map[string]time.Time
	limit time.Duration
}

// NewAnnouncer creates an announcer with a throttle window.
func NewAnnouncer(limit time.Duration) *Announcer {
	return &Announcer{seen: make(map[string]time.Time), limit: limit}
}

// ShouldAnnounce returns false when the announcement was emitted too recently.
func (a *Announcer) ShouldAnnounce(key string, now time.Time) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	last, seen := a.seen[key]
	if seen && now.Sub(last) < a.limit {
		return false
	}
	a.seen[key] = now
	return true
}

// HighContrastToken returns a override token when enabled.
func HighContrastToken(enabled bool, original string) string {
	if enabled {
		return "#000000"
	}
	return original
}

// FocusGraph captures deterministic adjacency for keyboard navigation.
type FocusGraph map[string][]string

// AddEdge registers a deterministic focus path.
// It is safe to call on the zero value because the map is lazily initialized.
func (g *FocusGraph) AddEdge(from, to string) {
	if *g == nil {
		*g = make(map[string][]string)
	}
	(*g)[from] = append((*g)[from], to)
}

// Neighbors returns the deterministic focus targets for a node.
func (g FocusGraph) Neighbors(node string) []string {
	return g[node]
}
