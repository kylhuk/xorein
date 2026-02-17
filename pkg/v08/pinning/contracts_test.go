package pinning

import (
	"reflect"
	"testing"
)

func TestValidatePinScope(t *testing.T) {
	cases := []struct {
		scope Scope
		want  bool
	}{
		{ScopePersonal, true},
		{ScopeTeam, true},
		{ScopeGlobal, true},
		{Scope("unknown"), false},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(string(tc.scope), func(t *testing.T) {
			t.Parallel()
			if got := ValidatePinScope(tc.scope); got != tc.want {
				t.Fatalf("scope %s expected %t, got %t", tc.scope, tc.want, got)
			}
		})
	}
}

func TestDeterministicOrder(t *testing.T) {
	input := []PinAuthority{
		{ID: "b", Scope: ScopeTeam, Priority: 2},
		{ID: "a", Scope: ScopeTeam, Priority: 1},
		{ID: "c", Scope: ScopePersonal, Priority: 5},
		{ID: "d", Scope: ScopeGlobal, Priority: 3},
	}
	want := []PinAuthority{
		{ID: "d", Scope: ScopeGlobal, Priority: 3},
		{ID: "c", Scope: ScopePersonal, Priority: 5},
		{ID: "a", Scope: ScopeTeam, Priority: 1},
		{ID: "b", Scope: ScopeTeam, Priority: 2},
	}

	got := DeterministicOrder(input)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ordering mismatch\nwant=%#v\ngot=%#v", want, got)
	}
}
