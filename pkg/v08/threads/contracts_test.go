package threads

import "testing"

func TestValidateReplyLineage(t *testing.T) {
	cases := []struct {
		name  string
		trace ThreadTrace
		want  bool
	}{
		{"valid progression", ThreadTrace{ID: "trace", CreatedDepth: 2, ReplyDepth: 3}, false},
		{"missing id", ThreadTrace{CreatedDepth: 1, ReplyDepth: 1}, true},
		{"negative reply", ThreadTrace{ID: "trace", CreatedDepth: 0, ReplyDepth: -1}, true},
		{"reply too deep", ThreadTrace{ID: "trace", CreatedDepth: 1, ReplyDepth: 3}, true},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateReplyLineage(tc.trace)
			if (err != nil) != tc.want {
				t.Fatalf("expected error=%t, got %v", tc.want, err)
			}
		})
	}
}

func TestClassifyLifecycle(t *testing.T) {
	cases := []struct {
		depth int
		want  Lifecycle
		desc  string
	}{
		{0, LifecycleDormant, "zero depth"},
		{1, LifecycleActive, "shallow"},
		{3, LifecycleActive, "mid depth"},
		{4, LifecycleArchived, "archived floor"},
		{10, LifecycleArchived, "archived high"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()
			if got := ClassifyLifecycle(tc.depth); got != tc.want {
				t.Fatalf("depth %d expected %s, got %s", tc.depth, tc.want, got)
			}
		})
	}
}

func TestDescribeLifecycle(t *testing.T) {
	if got := DescribeLifecycle(LifecycleActive); got != "thread-lifecycle:active" {
		t.Fatalf("unexpected description %s", got)
	}
}
