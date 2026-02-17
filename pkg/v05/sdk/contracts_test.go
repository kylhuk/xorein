package sdk

import "testing"

func TestCommunityProfileContractExpectations(t *testing.T) {
	contract := CommunityProfileContract()

	if contract.ProfileName != ProfileCommunity {
		t.Fatalf("expected community profile, got %s", contract.ProfileName)
	}

	required := contract.RequiredNames()
	if len(required) != 2 {
		t.Fatalf("expected two required expectations, got %d", len(required))
	}

	for _, name := range []string{"context-aware clients", "deterministic serialization"} {
		if !contract.HasExpectation(name) {
			t.Fatalf("expected expectation %q to be present", name)
		}
	}

	if !contract.HasExpectation("traceable event hooks") {
		t.Fatal("expected optional expectation to be discoverable")
	}
}

func TestProfileMeetsExpectations(t *testing.T) {
	contract := CommunityProfileContract()
	tests := []struct {
		name         string
		expectations []GoSDKExpectation
		want         bool
	}{
		{
			name: "meets required",
			expectations: []GoSDKExpectation{
				{Name: "context-aware clients"},
				{Name: "deterministic serialization"},
			},
			want: true,
		},
		{
			name: "missing required",
			expectations: []GoSDKExpectation{
				{Name: "context-aware clients"},
			},
			want: false,
		},
		{
			name: "extra expectation",
			expectations: []GoSDKExpectation{
				{Name: "context-aware clients"},
				{Name: "deterministic serialization"},
				{Name: "traceable event hooks"},
			},
			want: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if got := contract.Meets(tc.expectations); got != tc.want {
				t.Fatalf("Meets() => %v, want %v", got, tc.want)
			}
		})
	}
}
