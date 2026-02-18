package capture

import "testing"

func TestParseSource(t *testing.T) {
	for _, tc := range []struct {
		name   string
		input  string
		expect SourceType
	}{
		{name: "default", input: "", expect: SourceDisplay},
		{name: "display", input: "DISPLAY", expect: SourceDisplay},
		{name: "window", input: "window", expect: SourceWindow},
	} {
		got, err := ParseSource(tc.input)
		if err != nil {
			t.Fatalf("%s: unexpected error: %v", tc.name, err)
		}
		if got != tc.expect {
			t.Fatalf("%s: expected %q got %q", tc.name, tc.expect, got)
		}
	}
}

func TestProfileForPreset(t *testing.T) {
	for _, tc := range []struct {
		preset      Preset
		encoder     string
		bitrate     int
		expectError bool
	}{
		{preset: PresetHigh, encoder: "vp9", bitrate: 4500},
		{preset: PresetMedium, encoder: "h264", bitrate: 2500},
		{preset: PresetLow, encoder: "av1", bitrate: 1200},
		{preset: "unknown", expectError: true},
	} {
		profile, err := ProfileForPreset(tc.preset)
		if tc.expectError {
			if err == nil {
				t.Fatalf("%s: expected error", tc.preset)
			}
			continue
		}
		if err != nil {
			t.Fatalf("%s: unexpected error %v", tc.preset, err)
		}
		if profile.Encoder != tc.encoder || profile.BitrateKbps != tc.bitrate {
			t.Fatalf("unexpected profile %v", profile)
		}
	}
}
