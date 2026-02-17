package ipfs

import "testing"

func TestContentAddressDependsOnMetadata(t *testing.T) {
	t.Parallel()

	base := ContentMeta{Name: "item", SizeBytes: 12, Owner: "owner", CreatedAt: 100}
	first := ContentAddress(base)
	if first == "" {
		t.Fatalf("expected non-empty address")
	}
	base.Owner = "owner-2"
	second := ContentAddress(base)
	if second == first {
		t.Fatalf("expected address to change when metadata changes")
	}
}

func TestClassifyRetention(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		size   int64
		pinned bool
		want   RetentionState
	}{
		{name: "pinned large", size: 1_500_000_000, pinned: true, want: RetentionDurable},
		{name: "pinned small", size: 10_000, pinned: true, want: RetentionStaged},
		{name: "unpinned", size: 500, pinned: false, want: RetentionEphemeral},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if got := ClassifyRetention(c.size, c.pinned); got != c.want {
				t.Fatalf("Retention = %s, want %s", got, c.want)
			}
		})
	}
}

func TestDegradedBehavior(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		pinned    bool
		reachable bool
		wantLevel string
	}{
		{name: "reachable even pinned", pinned: true, reachable: true, wantLevel: "nominal"},
		{name: "pinned unreachable", pinned: true, reachable: false, wantLevel: "degraded"},
		{name: "unpinned unreachable", pinned: false, reachable: false, wantLevel: "critical"},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if got := DegradedBehavior(c.pinned, c.reachable); got.Level != c.wantLevel {
				t.Fatalf("level = %s, want %s", got.Level, c.wantLevel)
			}
		})
	}
}

func TestNextPinStageSequence(t *testing.T) {
	t.Parallel()

	if stage := NextPinStage(PinStageQueued); stage != PinStageActive {
		t.Fatalf("expected queued -> active, got %s", stage)
	}
	if stage := NextPinStage(PinStageActive); stage != PinStageExpired {
		t.Fatalf("expected active -> expired, got %s", stage)
	}
	if stage := NextPinStage("unknown"); stage != PinStageExpired {
		t.Fatalf("expected default unknown -> expired, got %s", stage)
	}
}
