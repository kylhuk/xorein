package assets

// Kind classifies frozen asset categories.
type Kind string

const (
	KindAttachment Kind = "attachment"
	KindAvatar     Kind = "avatar"
	KindEmoji      Kind = "emoji"
)

// State represents the renderer state for an asset.
type State string

const (
	StatePreview  State = "preview"
	StateDownload State = "download"
	StateDegraded State = "degraded"
)

// Reason codifies deterministic failure reasons visible to clients.
type Reason string

const (
	ReasonReady          Reason = "ready"
	ReasonMissingBlob    Reason = "missing-blob"
	ReasonRateLimited    Reason = "rate-limited"
	ReasonNetworkTimeout Reason = "network-timeout"
)

// RendererPlan describes how the renderer should behave for an asset request.
type RendererPlan struct {
	Kind        Kind
	Action      string
	State       State
	Reason      Reason
	Placeholder string
}

var placeholderByReason = map[Reason]string{
	ReasonMissingBlob:    "asset-placeholder-missing",
	ReasonRateLimited:    "asset-placeholder-rate-limited",
	ReasonNetworkTimeout: "asset-placeholder-network-timeout",
}

func placeholderFor(reason Reason) string {
	if ph, ok := placeholderByReason[reason]; ok {
		return ph
	}
	return "asset-placeholder-generic"
}

// PlanPreview returns a deterministic plan for a preview request.
func PlanPreview(kind Kind, blobAvailable bool) RendererPlan {
	if blobAvailable {
		return RendererPlan{
			Kind:        kind,
			Action:      "preview",
			State:       StatePreview,
			Reason:      ReasonReady,
			Placeholder: "",
		}
	}
	return PlanDegraded(kind, ReasonMissingBlob)
}

// PlanDownload returns a deterministic plan for a download request.
func PlanDownload(kind Kind, blobAvailable bool) RendererPlan {
	if blobAvailable {
		return RendererPlan{
			Kind:        kind,
			Action:      "download",
			State:       StateDownload,
			Reason:      ReasonReady,
			Placeholder: "",
		}
	}
	return PlanDegraded(kind, ReasonMissingBlob)
}

// PlanDegraded surfaces a placeholder for unavailable assets.
func PlanDegraded(kind Kind, reason Reason) RendererPlan {
	return RendererPlan{
		Kind:        kind,
		Action:      "degraded",
		State:       StateDegraded,
		Reason:      reason,
		Placeholder: placeholderFor(reason),
	}
}

// TelemetryFields builds a payload-safe telemetry map for the plan.
func TelemetryFields(plan RendererPlan) map[string]string {
	return map[string]string{
		"asset.kind":        string(plan.Kind),
		"asset.state":       string(plan.State),
		"asset.action":      plan.Action,
		"asset.reason":      string(plan.Reason),
		"asset.placeholder": plan.Placeholder,
	}
}
