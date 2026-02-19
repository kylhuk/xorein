package security

import (
	"sort"
	"strings"
	"sync"
)

// TelemetryAggregator keeps privacy-preserving counters keyed by aggregated labels.
type TelemetryAggregator struct {
	mu     sync.Mutex
	counts map[string]int
}

// NewTelemetryAggregator returns a ready-to-use aggregator.
func NewTelemetryAggregator() *TelemetryAggregator {
	return &TelemetryAggregator{counts: make(map[string]int)}
}

// Record increments the counter for the provided aggregated labels.
func (t *TelemetryAggregator) Record(labels map[string]string) {
	if t == nil {
		return
	}
	key := telemetryKey(labels)
	t.mu.Lock()
	defer t.mu.Unlock()
	t.counts[key]++
}

// Snapshot returns a copy of the current counters.
func (t *TelemetryAggregator) Snapshot() map[string]int {
	if t == nil {
		return nil
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	copy := make(map[string]int, len(t.counts))
	for k, v := range t.counts {
		copy[k] = v
	}
	return copy
}

func telemetryKey(labels map[string]string) string {
	if len(labels) == 0 {
		return ""
	}
	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var builder strings.Builder
	for _, k := range keys {
		builder.WriteString(k)
		builder.WriteString("=")
		builder.WriteString(labels[k])
		builder.WriteString(";")
	}
	return builder.String()
}
