package screenshare

import "testing"

func TestQualityPresetEncode(t *testing.T) {
	tests := []struct {
		preset   Preset
		expected int
	}{
		{PresetLow, 256},
		{PresetStandard, 512},
		{PresetHigh, 768},
		{PresetUltra, 1024},
		{PresetAuto, 600},
		{Preset("unknown"), 600},
	}

	for _, tt := range tests {
		if got := QualityPresetEncode(tt.preset); got != tt.expected {
			t.Fatalf("QualityPresetEncode(%q) = %d, want %d", tt.preset, got, tt.expected)
		}
	}
}

func TestSimulcastLayerLimit(t *testing.T) {
	if got := SimulcastLayerLimit(); got != 3 {
		t.Fatalf("SimulcastLayerLimit() = %d, want 3", got)
	}
}

func TestAllowedPresetsIncludesCore(t *testing.T) {
	seen := make(map[Preset]bool, len(AllowedPresets))
	for _, p := range AllowedPresets {
		seen[p] = true
	}
	for _, want := range []Preset{PresetLow, PresetStandard, PresetHigh, PresetUltra, PresetAuto} {
		if !seen[want] {
			t.Fatalf("AllowedPresets missing %q", want)
		}
	}
}

func TestViewerDegradationHint(t *testing.T) {
	tests := []struct {
		reason   ViewerControlReason
		expected string
	}{
		{ViewerReasonFullscreen, "Fullscreen locked; fallback to PiP if screen-delay exceeds 200ms"},
		{ViewerReasonPiP, "PiP maintains 1.0x frame rate unless bandwidth drops below 400kbps"},
		{ViewerReasonZoomPan, "Zoom/pan yields 1px-per-step deterministic interpolation"},
		{ViewerControlReason("other"), "Use Auto presets for adaptive fallback"},
	}

	for _, tt := range tests {
		if got := ViewerDegradationHint(tt.reason); got != tt.expected {
			t.Fatalf("ViewerDegradationHint(%q) = %q, want %q", tt.reason, got, tt.expected)
		}
	}
}

func TestResolveCaptureSource(t *testing.T) {
	tests := []struct {
		name           string
		preferWindow   bool
		windowAvailabe bool
		expected       CaptureSource
	}{
		{"display default", false, false, CaptureSourceDisplay},
		{"window unavailable", true, false, CaptureSourceDisplay},
		{"window preferred and available", true, true, CaptureSourceWindow},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ResolveCaptureSource(tt.preferWindow, tt.windowAvailabe); got != tt.expected {
				t.Fatalf("ResolveCaptureSource(%t, %t) = %q, want %q", tt.preferWindow, tt.windowAvailabe, got, tt.expected)
			}
		})
	}
}

func TestSelectEncoder(t *testing.T) {
	tests := []struct {
		name           string
		available      []Encoder
		preferHardware bool
		expected       Encoder
	}{
		{"empty fallback", nil, true, EncoderSoftware},
		{"prefer first when not hardware", []Encoder{EncoderVP8, EncoderH264}, false, EncoderVP8},
		{"prefer h264", []Encoder{EncoderVP8, EncoderH264}, true, EncoderH264},
		{"prefer vp9 when h264 missing", []Encoder{EncoderVP8, EncoderVP9}, true, EncoderVP9},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SelectEncoder(tt.available, tt.preferHardware); got != tt.expected {
				t.Fatalf("SelectEncoder(%v, %t) = %q, want %q", tt.available, tt.preferHardware, got, tt.expected)
			}
		})
	}
}

func TestScreenSecurityDisclosure(t *testing.T) {
	if got := ScreenSecurityDisclosure(true); got != "Media E2EE screen-share" {
		t.Fatalf("unexpected e2ee disclosure: %q", got)
	}
	if got := ScreenSecurityDisclosure(false); got != "Not E2EE screen-share (relay/server-readable)" {
		t.Fatalf("unexpected clear disclosure: %q", got)
	}
}
