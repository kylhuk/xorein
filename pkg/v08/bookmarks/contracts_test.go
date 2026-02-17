package bookmarks

import (
	"testing"
	"time"
)

func TestValidatePrivacy(t *testing.T) {
	cases := []struct {
		name string
		b    Bookmark
		want bool
	}{
		{"valid personal", Bookmark{ID: "id", Privacy: PrivacyPersonal}, true},
		{"missing id", Bookmark{Privacy: PrivacyShared}, false},
		{"invalid privacy", Bookmark{ID: "id", Privacy: "public"}, false},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := ValidatePrivacy(tc.b)
			if (err == nil) != tc.want {
				t.Fatalf("expected success=%t, got err=%v", tc.want, err)
			}
		})
	}
}

func TestLifecycle(t *testing.T) {
	now := time.Now()
	cases := []struct {
		name string
		b    Bookmark
		want BookmarkLifecycle
	}{
		{"fresh", Bookmark{CreatedAt: now, Privacy: PrivacyPersonal}, LifecycleFresh},
		{"persistent older than day", Bookmark{CreatedAt: now.Add(-25 * time.Hour), Privacy: PrivacyShared, Archived: false}, LifecyclePersistent},
		{"archived regardless of age", Bookmark{CreatedAt: now.Add(-25 * time.Hour), Archived: true, Privacy: PrivacyPersonal}, LifecycleArchived},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := Lifecycle(tc.b); got != tc.want {
				t.Fatalf("expected %s, got %s", tc.want, got)
			}
		})
	}
}
