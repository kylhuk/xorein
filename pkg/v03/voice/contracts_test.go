package voice

import "testing"

func TestEffectiveBitrateBounds(t *testing.T) {
	tests := []struct {
		name         string
		available    int
		expectedRate int
	}{
		{"floor", 5, 16},
		{"lower bound", 16, 16},
		{"mid range", 64, 64},
		{"upper bound", 128, 128},
		{"cap", 300, 128},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EffectiveBitrate(tt.available); got != tt.expectedRate {
				t.Fatalf("EffectiveBitrate(%d) = %d, want %d", tt.available, got, tt.expectedRate)
			}
		})
	}
}

func TestJitterWindowRanges(t *testing.T) {
	tests := []struct {
		ms       int
		expected string
	}{
		{10, "underflow"},
		{20, "tight"},
		{60, "tight"},
		{90, "bounded"},
		{150, "wide"},
		{250, "overflow"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := JitterWindow(tt.ms); got != tt.expected {
				t.Fatalf("JitterWindow(%d) = %q, want %q", tt.ms, got, tt.expected)
			}
		})
	}
}

func TestMediaSecurityDisclosure(t *testing.T) {
	if got := MediaSecurityDisclosure(AudioSecurityE2EE); got == MediaSecurityDisclosure(AudioSecurityClear) {
		t.Fatalf("media security disclosure should differ between modes: got %q", got)
	}
	wantClear := "Clear (server-readable) with FEC+DTX fallback in relay scenarios"
	if got := MediaSecurityDisclosure(AudioSecurityClear); got != wantClear {
		t.Fatalf("unexpected clear disclosure: got %q", got)
	}
}

func TestFECEnabled(t *testing.T) {
	tests := []struct {
		fec      bool
		expected string
	}{
		{true, "FEC enabled"},
		{false, "FEC disabled"},
	}

	for _, tt := range tests {
		if got := FECEnabled(tt.fec); got != tt.expected {
			t.Fatalf("FECEnabled(%t) = %q, want %q", tt.fec, got, tt.expected)
		}
	}
}

func TestShouldElectPeerSFU(t *testing.T) {
	if ShouldElectPeerSFU(8) {
		t.Fatal("participants below threshold should not elect SFU")
	}
	if !ShouldElectPeerSFU(9) {
		t.Fatal("participants at threshold should elect SFU")
	}
}

func TestElectPeerSFU(t *testing.T) {
	candidates := []SFUCandidate{
		{PeerID: "peer-c", Score: 90, RTTms: 40, Eligible: true},
		{PeerID: "peer-a", Score: 90, RTTms: 30, Eligible: true},
		{PeerID: "peer-b", Score: 90, RTTms: 30, Eligible: true},
		{PeerID: "peer-z", Score: 99, RTTms: 100, Eligible: false},
	}

	winner, ok := ElectPeerSFU(candidates)
	if !ok {
		t.Fatal("expected winner for eligible candidates")
	}
	if winner.PeerID != "peer-a" {
		t.Fatalf("winner = %q, want %q", winner.PeerID, "peer-a")
	}
}

func TestElectPeerSFUEmpty(t *testing.T) {
	if _, ok := ElectPeerSFU(nil); ok {
		t.Fatal("expected no winner for empty candidate list")
	}
}

func TestRelaySFUModeDisclosure(t *testing.T) {
	tests := []struct {
		name         string
		enabled      bool
		participants int
		expected     string
	}{
		{"disabled", false, 20, "Relay SFU disabled; remain on peer voice topology"},
		{"enabled threshold not met", true, 4, "Relay SFU enabled; threshold not met, remain on peer voice topology"},
		{"enabled threshold met", true, 12, "Relay SFU enabled; transition to relay-assisted SFU"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RelaySFUModeDisclosure(tt.enabled, tt.participants); got != tt.expected {
				t.Fatalf("RelaySFUModeDisclosure(%t, %d) = %q, want %q", tt.enabled, tt.participants, got, tt.expected)
			}
		})
	}
}
