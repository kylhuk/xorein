package bookmarks

import (
	"fmt"
	"time"
)

// Privacy labels the allowed visibility for a bookmark.
type Privacy string

const (
	PrivacyPersonal Privacy = "personal"
	PrivacyShared   Privacy = "shared"
)

// BookmarkLifecycle describes deterministic states for bookmarks.
type BookmarkLifecycle string

const (
	LifecycleFresh      BookmarkLifecycle = "fresh"
	LifecyclePersistent BookmarkLifecycle = "persistent"
	LifecycleArchived   BookmarkLifecycle = "archived"
)

// Bookmark represents the contract around a user-supplied bookmark.
type Bookmark struct {
	ID        string
	Owner     string
	Privacy   Privacy
	CreatedAt time.Time
	Archived  bool
}

// ValidatePrivacy ensures a bookmark always uses a deterministic privacy label.
func ValidatePrivacy(b Bookmark) error {
	if b.ID == "" {
		return fmt.Errorf("bookmark missing ID")
	}
	if b.Privacy != PrivacyPersonal && b.Privacy != PrivacyShared {
		return fmt.Errorf("bookmark %s has invalid privacy %q", b.ID, b.Privacy)
	}
	return nil
}

// Lifecycle returns the deterministic lifecycle label for bookmarks.
func Lifecycle(b Bookmark) BookmarkLifecycle {
	if b.Archived {
		return LifecycleArchived
	}
	if time.Since(b.CreatedAt) > 24*time.Hour {
		return LifecyclePersistent
	}
	return LifecycleFresh
}
