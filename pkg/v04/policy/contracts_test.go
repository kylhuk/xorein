package policy

import "testing"

func TestPolicyVersionString(t *testing.T) {
	version := PolicyVersion{ID: "beta", Major: 2, Minor: 5}
	if got := version.String(); got != "2.5:beta" {
		t.Fatalf("String(): got %s want 2.5:beta", got)
	}
}

func TestPolicyVersionCanMigrateTo(t *testing.T) {
	tests := []struct {
		name   string
		base   PolicyVersion
		target PolicyVersion
		want   bool
	}{
		{name: "immutable false", base: PolicyVersion{Major: 1, Minor: 2, Immutable: false}, target: PolicyVersion{Major: 1, Minor: 3, Immutable: false}, want: false},
		{name: "major mismatch", base: PolicyVersion{Major: 1, Minor: 2, Immutable: true}, target: PolicyVersion{Major: 2, Minor: 0, Immutable: true}, want: false},
		{name: "minor downgrade", base: PolicyVersion{Major: 1, Minor: 2, Immutable: true}, target: PolicyVersion{Major: 1, Minor: 1, Immutable: true}, want: false},
		{name: "minor equal", base: PolicyVersion{Major: 1, Minor: 2, Immutable: true}, target: PolicyVersion{Major: 1, Minor: 2, Immutable: true}, want: true},
		{name: "minor upgrade", base: PolicyVersion{Major: 1, Minor: 2, Immutable: true}, target: PolicyVersion{Major: 1, Minor: 4, Immutable: true}, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.base.CanMigrateTo(tt.target); got != tt.want {
				t.Fatalf("CanMigrateTo %s: got %t want %t", tt.name, got, tt.want)
			}
		})
	}
}

func TestPolicyTraceRollbackAndAppend(t *testing.T) {
	current := PolicyVersion{ID: "current", Major: 2, Minor: 0, Immutable: true}
	trace := PolicyTrace{
		Current: current,
		History: []PolicyVersion{{ID: "first", Major: 1, Minor: 0}, {ID: "second", Major: 1, Minor: 1}},
	}

	if _, ok := trace.RollbackTarget(); !ok {
		t.Fatal("RollbackTarget should succeed when history exists")
	}

	empty := PolicyTrace{}
	if _, ok := empty.RollbackTarget(); ok {
		t.Fatal("RollbackTarget should fail when history is empty")
	}

	next := PolicyVersion{ID: "next", Major: 2, Minor: 1, Immutable: true}
	appended := trace.Append(next)

	if appended.Current != next {
		t.Fatalf("Append current: got %v want %v", appended.Current, next)
	}
	if len(appended.History) != len(trace.History)+1 {
		t.Fatalf("history length: got %d want %d", len(appended.History), len(trace.History)+1)
	}
	if got := appended.History[len(appended.History)-1]; got != trace.Current {
		t.Fatalf("history last entry: got %v want %v", got, trace.Current)
	}
	if len(trace.History) != 2 {
		t.Fatalf("original history mutated: got %d want 2", len(trace.History))
	}
}
