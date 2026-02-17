package emoji

import (
	"strings"
	"testing"
)

func TestCanAddReactionQuota(t *testing.T) {
	if !CanAddReaction(MaxEmojiQuota - 1) {
		t.Fatalf("should allow reaction when under quota: %d", MaxEmojiQuota-1)
	}
	if CanAddReaction(MaxEmojiQuota) {
		t.Fatalf("should reject reaction at quota boundary: %d", MaxEmojiQuota)
	}
}

func TestShortcodeNormalizeAndValidate(t *testing.T) {
	tests := []struct {
		name       string
		value      string
		normalized string
		valid      bool
	}{
		{name: "trim colons", value: ":ThumbsUp:", normalized: "thumbsup", valid: true},
		{name: "space padded", value: "  :wave:  ", normalized: "wave", valid: true},
		{name: "empty", value: "::", normalized: "", valid: false},
		{name: "too long", value: strings.Repeat("a", 33), normalized: strings.Repeat("a", 33), valid: false},
		{name: "bad char", value: ":bad!char:", normalized: "bad!char", valid: false},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := NormalizeShortcode(tc.value)
			if tc.normalized != "" && got != tc.normalized {
				t.Fatalf("normalized => %s", got)
			}

			if valid := ValidShortcode(tc.value); valid != tc.valid {
				t.Fatalf("ValidShortcode(%s) => %v", tc.value, valid)
			}
		})
	}
}

func TestTransitionStateConvergence(t *testing.T) {
	cases := []struct {
		name    string
		current ReactionState
		action  ReactionAction
		want    ReactionState
	}{
		{name: "add from removed", current: ReactionRemoved, action: ReactionActionAdd, want: ReactionPending},
		{name: "add from added", current: ReactionAdded, action: ReactionActionAdd, want: ReactionAdded},
		{name: "remove", current: ReactionAdded, action: ReactionActionRemove, want: ReactionRemoved},
		{name: "unknown action", current: ReactionPending, action: ReactionAction("noop"), want: ReactionPending},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if got := TransitionState(tc.current, tc.action); got != tc.want {
				t.Fatalf("TransitionState => %s, want %s", got, tc.want)
			}
		})
	}
}
