package audit

import (
	"sort"

	"github.com/aether/code_aether/pkg/v17/moderation"
)

// Role represents the visibility tier for the audit log.
type Role string

const (
	RoleObserver  Role = "observer"
	RoleModerator Role = "moderator"
	RoleAdmin     Role = "admin"
)

// Entry is an append-only audit row.
type Entry struct {
	Sequence   int
	Event      moderation.SignedEvent
	Visibility []Role
}

// AuditLog is an append-only store for moderation events.
type AuditLog struct {
	entries      []Entry
	nextSequence int
}

// NewAuditLog builds a fresh audit log.
func NewAuditLog() *AuditLog {
	return &AuditLog{}
}

// Append adds an entry with deterministic role visibility.
func (l *AuditLog) Append(event moderation.SignedEvent, visibility []Role) Entry {
	cleaned := uniqueRoles(visibility)
	entry := Entry{Sequence: l.nextSequence, Event: event, Visibility: cleaned}
	l.entries = append(l.entries, entry)
	l.nextSequence++
	return entry
}

// Query returns entries visible to the requested role.
func (l *AuditLog) Query(role Role) []Entry {
	result := make([]Entry, 0, len(l.entries))
	for _, entry := range l.entries {
		if entry.visibleTo(role) {
			result = append(result, entry)
		}
	}
	return result
}

// Entries returns a copy of all rows in order.
func (l *AuditLog) Entries() []Entry {
	copyEntries := make([]Entry, len(l.entries))
	copy(copyEntries, l.entries)
	return copyEntries
}

func uniqueRoles(in []Role) []Role {
	seen := make(map[Role]struct{})
	out := make([]Role, 0, len(in))
	for _, role := range in {
		if _, ok := seen[role]; ok {
			continue
		}
		seen[role] = struct{}{}
		out = append(out, role)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

func (e Entry) visibleTo(role Role) bool {
	if len(e.Visibility) == 0 {
		return true
	}
	for _, candidate := range e.Visibility {
		if candidate == role {
			return true
		}
	}
	return false
}
