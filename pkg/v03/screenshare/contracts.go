package screenshare

type Preset string

const (
	PresetLow      Preset = "low"
	PresetStandard Preset = "standard"
	PresetHigh     Preset = "high"
	PresetUltra    Preset = "ultra"
	PresetAuto     Preset = "auto"
)

var AllowedPresets = []Preset{PresetLow, PresetStandard, PresetHigh, PresetUltra, PresetAuto}

type CaptureSource string

const (
	CaptureSourceDisplay CaptureSource = "display"
	CaptureSourceWindow  CaptureSource = "window"
)

func ResolveCaptureSource(preferWindow bool, windowAvailable bool) CaptureSource {
	if preferWindow && windowAvailable {
		return CaptureSourceWindow
	}
	return CaptureSourceDisplay
}

type Encoder string

const (
	EncoderH264     Encoder = "h264"
	EncoderVP9      Encoder = "vp9"
	EncoderVP8      Encoder = "vp8"
	EncoderSoftware Encoder = "software"
)

func SelectEncoder(available []Encoder, preferHardware bool) Encoder {
	if len(available) == 0 {
		return EncoderSoftware
	}
	if !preferHardware {
		return available[0]
	}
	priority := []Encoder{EncoderH264, EncoderVP9, EncoderVP8}
	for _, expected := range priority {
		for _, offered := range available {
			if offered == expected {
				return offered
			}
		}
	}
	return available[0]
}

func ScreenSecurityDisclosure(e2ee bool) string {
	if e2ee {
		return "Media E2EE screen-share"
	}
	return "Not E2EE screen-share (relay/server-readable)"
}

func SimulcastLayerLimit() int {
	return 3
}

type ViewerControlReason string

const (
	ViewerReasonFullscreen ViewerControlReason = "fullscreen"
	ViewerReasonPiP        ViewerControlReason = "pip"
	ViewerReasonZoomPan    ViewerControlReason = "zoom-pan"
)

func QualityPresetEncode(p Preset) int {
	switch p {
	case PresetLow:
		return 256
	case PresetStandard:
		return 512
	case PresetHigh:
		return 768
	case PresetUltra:
		return 1024
	default:
		return 600
	}
}

func ViewerDegradationHint(reason ViewerControlReason) string {
	switch reason {
	case ViewerReasonFullscreen:
		return "Fullscreen locked; fallback to PiP if screen-delay exceeds 200ms"
	case ViewerReasonPiP:
		return "PiP maintains 1.0x frame rate unless bandwidth drops below 400kbps"
	case ViewerReasonZoomPan:
		return "Zoom/pan yields 1px-per-step deterministic interpolation"
	default:
		return "Use Auto presets for adaptive fallback"
	}
}
