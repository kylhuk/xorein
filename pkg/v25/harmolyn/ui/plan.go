package ui

// ProgressStage describes where the asset flow currently stands.
type ProgressStage string

const (
	ProgressIdle        ProgressStage = "idle"
	ProgressUploading   ProgressStage = "uploading"
	ProgressDownloading ProgressStage = "downloading"
	ProgressComplete    ProgressStage = "complete"
	ProgressCancelled   ProgressStage = "cancelled"
	ProgressError       ProgressStage = "error"
)

// UXReason codifies deterministic error and informational surfaces.
type UXReason string

const (
	ReasonReady         UXReason = "ready"
	ReasonOffline       UXReason = "offline"
	ReasonTapToDownload UXReason = "tap-to-download"
	ReasonWiFiOnly      UXReason = "wifi-only"
	ReasonCancelled     UXReason = "cancelled"
	ReasonNetworkError  UXReason = "network-error"
)

var badgeByReason = map[UXReason]string{
	ReasonOffline:       "badge-offline",
	ReasonTapToDownload: "badge-tap-to-download",
	ReasonWiFiOnly:      "badge-wifi-only",
	ReasonCancelled:     "badge-cancelled",
	ReasonNetworkError:  "badge-network-error",
}

// Controls describes the user-facing toggles in harmolyn.
type Controls struct {
	TapToDownload      bool
	DownloadOnWiFiOnly bool
}

// NetworkType describes the last known network connectivity.
type NetworkType string

const (
	NetworkUnknown  NetworkType = "unknown"
	NetworkWiFi     NetworkType = "wifi"
	NetworkCellular NetworkType = "cellular"
	NetworkOffline  NetworkType = "offline"
)

// AssetUXPlan represents the deterministic renderer contract for harmolyn.
type AssetUXPlan struct {
	Stage        ProgressStage
	Action       string
	Reason       UXReason
	ProgressPct  int
	Cancellable  bool
	OfflineBadge string
	Controls     Controls
}

// PlanUploadProgress surfaces upload progress and cancellation state.
func PlanUploadProgress(percent int, cancellable bool) AssetUXPlan {
	return AssetUXPlan{
		Stage:       ProgressUploading,
		Action:      "upload",
		Reason:      ReasonReady,
		ProgressPct: clampPercent(percent),
		Cancellable: cancellable,
	}
}

// PlanUploadCancelled surfaces the deterministic cancelled state.
func PlanUploadCancelled() AssetUXPlan {
	return AssetUXPlan{
		Stage:        ProgressCancelled,
		Action:       "upload",
		Reason:       ReasonCancelled,
		ProgressPct:  0,
		Cancellable:  false,
		OfflineBadge: badgeFor(ReasonCancelled),
	}
}

// PlanDownloadProgress surfaces download progress with optional cancellation.
func PlanDownloadProgress(percent int, cancellable bool) AssetUXPlan {
	return AssetUXPlan{
		Stage:       ProgressDownloading,
		Action:      "download",
		Reason:      ReasonReady,
		ProgressPct: clampPercent(percent),
		Cancellable: cancellable,
	}
}

// PlanDownloadRequest returns the deterministic plan for a download intent.
func PlanDownloadRequest(network NetworkType, controls Controls, userTapped bool) AssetUXPlan {
	plan := AssetUXPlan{
		Stage:    ProgressIdle,
		Action:   "download",
		Reason:   ReasonReady,
		Controls: controls,
	}

	switch {
	case network == NetworkOffline:
		plan.Reason = ReasonOffline
		plan.OfflineBadge = badgeFor(ReasonOffline)
	case controls.DownloadOnWiFiOnly && network != NetworkWiFi:
		plan.Reason = ReasonWiFiOnly
		plan.OfflineBadge = badgeFor(ReasonWiFiOnly)
	case controls.TapToDownload && !userTapped:
		plan.Reason = ReasonTapToDownload
		plan.OfflineBadge = badgeFor(ReasonTapToDownload)
	default:
		plan.Stage = ProgressDownloading
		plan.Reason = ReasonReady
		plan.ProgressPct = 0
		plan.Cancellable = true
	}

	return plan
}

// PlanDownloadNetworkError surfaces deterministic network failures.
func PlanDownloadNetworkError() AssetUXPlan {
	return AssetUXPlan{
		Stage:        ProgressError,
		Action:       "download",
		Reason:       ReasonNetworkError,
		OfflineBadge: badgeFor(ReasonNetworkError),
	}
}

func badgeFor(reason UXReason) string {
	if badge, ok := badgeByReason[reason]; ok {
		return badge
	}
	return ""
}

func clampPercent(percent int) int {
	switch {
	case percent < 0:
		return 0
	case percent > 100:
		return 100
	default:
		return percent
	}
}
