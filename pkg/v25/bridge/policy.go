package bridge

// TransferMode describes how a bridge asset request is delivered.
type TransferMode string

const (
	TransferModeMetadata TransferMode = "metadata"
	TransferModeRaw      TransferMode = "raw"
)

// ProviderCapabilities describes what a provider advertises for bridge-
// oriented asset requests.
type ProviderCapabilities struct {
	SupportsAssetBridges bool
}

// BotCapabilities describes whether a caller is allowed to exercise
// bridge asset requests.
type BotCapabilities struct {
	AllowsAssetBridge bool
}

// RefusalReason encodes deterministic refusal codes for denied bridge
// asset requests.
type RefusalReason string

const (
	ReasonBridgeAllowed             RefusalReason = "bridge-allowed"
	ReasonRawBlobNotAllowed         RefusalReason = "bridge-raw-blob-not-allowed"
	ReasonProviderBridgeUnsupported RefusalReason = "bridge-provider-unsupported"
	ReasonBotBridgeDenied           RefusalReason = "bridge-bot-capability-denied"
)

// Predefined messages describe the safe user-facing text attached to
// each refusal reason. They are intentionally short, deterministic, and
// free of sensitive details.
const (
	MessageBridgeAllowed             = "Bridge asset policy approved metadata-only transfer."
	MessageRawBlobNotAllowed         = "Bridges receive metadata tokens only; raw blob bytes are rejected."
	MessageProviderBridgeUnsupported = "The chosen provider is not enabled for bridge asset paths."
	MessageBotBridgeDenied           = "This bridge or bot lacks the capability to request assets."
)

// AssetBridgeDecision describes whether a bridge asset request may
// proceed and which payload types the decision forwarding path may use.
type AssetBridgeDecision struct {
	Allowed              bool
	Reason               RefusalReason
	Message              string
	ForwardMetadataToken bool
	ForwardRawBytes      bool
}

// EvaluateAssetBridgeRequest enforces the deterministic bridge asset
// policy described in ST1-ST3. It never allows raw blob bytes to leave
// the bridge path, it consults provider and bot capabilities, and it
// surfaces deterministic refusal reasons and user-safe messages.
func EvaluateAssetBridgeRequest(provider ProviderCapabilities, bot BotCapabilities, mode TransferMode) AssetBridgeDecision {
	if mode == TransferModeRaw {
		return AssetBridgeDecision{
			Allowed:              false,
			Reason:               ReasonRawBlobNotAllowed,
			Message:              MessageRawBlobNotAllowed,
			ForwardMetadataToken: false,
			ForwardRawBytes:      false,
		}
	}
	if !provider.SupportsAssetBridges {
		return AssetBridgeDecision{
			Allowed:              false,
			Reason:               ReasonProviderBridgeUnsupported,
			Message:              MessageProviderBridgeUnsupported,
			ForwardMetadataToken: false,
			ForwardRawBytes:      false,
		}
	}
	if !bot.AllowsAssetBridge {
		return AssetBridgeDecision{
			Allowed:              false,
			Reason:               ReasonBotBridgeDenied,
			Message:              MessageBotBridgeDenied,
			ForwardMetadataToken: false,
			ForwardRawBytes:      false,
		}
	}
	return AssetBridgeDecision{
		Allowed:              true,
		Reason:               ReasonBridgeAllowed,
		Message:              MessageBridgeAllowed,
		ForwardMetadataToken: true,
		ForwardRawBytes:      false,
	}
}
