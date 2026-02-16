package governance

import "sort"

type CompatibilityMetadata struct {
	ProtocolID      string
	ProtocolMajor   int
	ProtocolMinor   int
	RequiredFlags   []string
	AdvertisedFlags []string
}

func (m CompatibilityMetadata) Normalized() CompatibilityMetadata {
	out := CompatibilityMetadata{
		ProtocolID:      m.ProtocolID,
		ProtocolMajor:   m.ProtocolMajor,
		ProtocolMinor:   m.ProtocolMinor,
		RequiredFlags:   append([]string(nil), m.RequiredFlags...),
		AdvertisedFlags: append([]string(nil), m.AdvertisedFlags...),
	}
	sort.Strings(out.RequiredFlags)
	sort.Strings(out.AdvertisedFlags)
	return out
}

type SecurityMode string

const (
	SecurityModeSeal SecurityMode = "seal"
	SecurityModeTree SecurityMode = "tree"
)

type ModeMetadata struct {
	SecurityMode  SecurityMode
	ModeEpochID   string
	PolicyVersion uint64
}
