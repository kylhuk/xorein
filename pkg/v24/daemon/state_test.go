package daemon

import "testing"

func TestValidateTransition(t *testing.T) {
	tests := []struct {
		name    string
		from    State
		to      State
		wantErr bool
	}{
		{"stopped->starting", StateStopped, StateStarting, false},
		{"running->stopping", StateRunning, StateStopping, false},
		{"stopping->stopped", StateStopping, StateStopped, false},
		{"starting->running", StateStarting, StateRunning, false},
		{"running->starting", StateRunning, StateStarting, true},
		{"stopped->running", StateStopped, StateRunning, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTransition(tt.from, tt.to)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ValidateTransition() error = %v, wantErr=%v", err, tt.wantErr)
			}
		})
	}
}
