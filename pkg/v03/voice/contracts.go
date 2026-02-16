package voice

type AudioSecurityMode string

const (
	AudioSecurityE2EE  AudioSecurityMode = "e2ee"
	AudioSecurityClear AudioSecurityMode = "clear"
)

type VoiceQualityReason string

const (
	VoiceQualityReasonRNNoise      VoiceQualityReason = "rnnoise"       // baseline noise suppression
	VoiceQualityReasonBitrateClamp VoiceQualityReason = "bitrate-clamp" // ABR guardrails
	VoiceQualityReasonJitterRange  VoiceQualityReason = "jitter-range"  // jitter buffer boundary
	VoiceQualityReasonFecDtx       VoiceQualityReason = "fec-dtx"       // forward correction/dtx toggles
	VoiceQualityReasonSecurity     VoiceQualityReason = "security"      // mode disclosure
	VoiceQualityReasonSFUElection  VoiceQualityReason = "sfu-election"  // 9+ participant election path
	VoiceQualityReasonRelayMode    VoiceQualityReason = "relay-mode"    // relay sfu enable/disable state
)

type SFUCandidate struct {
	PeerID   string
	Score    int
	RTTms    int
	Eligible bool
}

func ShouldElectPeerSFU(participants int) bool {
	return participants >= 9
}

func ElectPeerSFU(candidates []SFUCandidate) (SFUCandidate, bool) {
	var winner SFUCandidate
	found := false
	for _, candidate := range candidates {
		if !candidate.Eligible {
			continue
		}
		if !found || candidate.Score > winner.Score ||
			(candidate.Score == winner.Score && candidate.RTTms < winner.RTTms) ||
			(candidate.Score == winner.Score && candidate.RTTms == winner.RTTms && candidate.PeerID < winner.PeerID) {
			winner = candidate
			found = true
		}
	}
	return winner, found
}

func RelaySFUModeDisclosure(enabled bool, participants int) string {
	if !enabled {
		return "Relay SFU disabled; remain on peer voice topology"
	}
	if !ShouldElectPeerSFU(participants) {
		return "Relay SFU enabled; threshold not met, remain on peer voice topology"
	}
	return "Relay SFU enabled; transition to relay-assisted SFU"
}

func EffectiveBitrate(availableKbps int) int {
	if availableKbps < 16 {
		return 16
	}
	if availableKbps > 128 {
		return 128
	}
	return availableKbps
}

func JitterWindow(ms int) string {
	switch {
	case ms < 20:
		return "underflow" // jitter buffer refilled
	case ms <= 60:
		return "tight"
	case ms <= 120:
		return "bounded"
	case ms <= 200:
		return "wide"
	default:
		return "overflow"
	}
}

func MediaSecurityDisclosure(mode AudioSecurityMode) string {
	if mode == AudioSecurityClear {
		return "Clear (server-readable) with FEC+DTX fallback in relay scenarios"
	}
	return "Media E2EE (SFrame/Opus) with RNNoise/ABR enforcement"
}

func FECEnabled(fec bool) string {
	if fec {
		return "FEC enabled"
	}
	return "FEC disabled"
}
