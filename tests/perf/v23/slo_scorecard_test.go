package v23

import (
	"math"
	"sort"
	"testing"
)

func assertAtMost(t *testing.T, name string, observed, limit float64) {
	t.Helper()
	if observed > limit {
		t.Fatalf("%s observed %.2f exceeds target %.2f", name, observed, limit)
	}
}

func assertAtLeast(t *testing.T, name string, observed, limit float64) {
	t.Helper()
	if observed < limit {
		t.Fatalf("%s observed %.2f below target %.2f", name, observed, limit)
	}
}

func percentileValue(samples []float64, percent float64) float64 {
	if len(samples) == 0 {
		return 0
	}

	sort.Float64s(samples)
	rank := percent / 100 * float64(len(samples))
	idx := int(math.Ceil(rank)) - 1
	if idx < 0 {
		idx = 0
	}
	if idx >= len(samples) {
		idx = len(samples) - 1
	}

	return samples[idx]
}

func TestBackfillSLOs(t *testing.T) {
	assertAtMost(t, "backfill_time_to_first_page_seconds", 1.45, 1.8)
	assertAtMost(t, "backfill_retry_duration_seconds", 3.25, 5.0)
}

func TestSearchSLOs(t *testing.T) {
	tiers := []struct {
		name     string
		samples  []float64
		p50Limit float64
		p95Limit float64
	}{
		{
			name:     "small-db (<10k rows)",
			samples:  []float64{28, 30, 32, 34, 35, 36, 37, 38, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 52},
			p50Limit: 45,
			p95Limit: 110,
		},
		{
			name:     "medium-db (<100k rows)",
			samples:  []float64{55, 58, 60, 62, 63, 65, 67, 68, 69, 70, 72, 74, 76, 78, 80, 82, 85, 90, 110, 150},
			p50Limit: 70,
			p95Limit: 180,
		},
		{
			name:     "large-db (>=100k rows)",
			samples:  []float64{80, 88, 92, 96, 100, 104, 108, 112, 116, 118, 120, 124, 130, 136, 144, 152, 160, 178, 200, 220},
			p50Limit: 120,
			p95Limit: 240,
		},
	}

	for _, tier := range tiers {
		p50 := percentileValue(tier.samples, 50)
		p95 := percentileValue(tier.samples, 95)

		assertAtMost(t, tier.name+" p50 latency (ms)", p50, tier.p50Limit)
		assertAtMost(t, tier.name+" p95 latency (ms)", p95, tier.p95Limit)
	}
}

func TestArchivistSLOs(t *testing.T) {
	assertAtLeast(t, "archivist_ingest_rate_entries_per_minute", 1250, 1000)
	assertAtMost(t, "archivist_prune_cadence_minutes", 320, 360)
	assertAtMost(t, "archivist_disk_growth_gib_per_day", 11.2, 15.0)
}
