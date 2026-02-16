package securitymode

type ChannelMode string

const (
	ModeTree    ChannelMode = "Tree"
	ModeCrowd   ChannelMode = "Crowd"
	ModeChannel ChannelMode = "Channel"
	ModeClear   ChannelMode = "Clear"
	ModeE2EE    ChannelMode = "E2EE"
)

var DefaultChannelModes = map[string]ChannelMode{
	"tree":    ModeTree,
	"crowd":   ModeCrowd,
	"channel": ModeChannel,
	"clear":   ModeClear,
	"e2ee":    ModeE2EE,
}

var AllowedTransitions = map[ChannelMode][]ChannelMode{
	ModeTree:    {ModeCrowd, ModeChannel},
	ModeCrowd:   {ModeChannel, ModeClear},
	ModeChannel: {ModeClear, ModeE2EE},
	ModeClear:   {ModeE2EE},
	ModeE2EE:    {ModeClear},
}

type DisclosureRequirement struct {
	Mode        ChannelMode
	Description string
}

func DefaultModeFor(channelType string) ChannelMode {
	if mode, ok := DefaultChannelModes[channelType]; ok {
		return mode
	}
	return ModeChannel
}

func AllowsTransition(from, to ChannelMode) bool {
	allowed, ok := AllowedTransitions[from]
	if !ok {
		return false
	}
	for _, candidate := range allowed {
		if candidate == to {
			return true
		}
	}
	return false
}

func DisclosureFor(mode ChannelMode) DisclosureRequirement {
	switch mode {
	case ModeTree:
		return DisclosureRequirement{Mode: mode, Description: "Topology-aware disclosure"}
	case ModeCrowd:
		return DisclosureRequirement{Mode: mode, Description: "Crowd policy mention"}
	case ModeChannel:
		return DisclosureRequirement{Mode: mode, Description: "Channel override disclosure"}
	case ModeClear:
		return DisclosureRequirement{Mode: mode, Description: "Clear channel audit-ready state"}
	case ModeE2EE:
		return DisclosureRequirement{Mode: mode, Description: "E2EE summary evidence"}
	default:
		return DisclosureRequirement{Mode: mode, Description: "Unknown mode"}
	}
}
