package ui

import "testing"

func TestPlanUploadProgressClampsAndCancellable(t *testing.T) {
	plan := PlanUploadProgress(120, true)
	if plan.Stage != ProgressUploading {
		t.Fatalf("unexpected stage: %v", plan.Stage)
	}
	if plan.ProgressPct != 100 {
		t.Fatalf("percent was not clamped: %d", plan.ProgressPct)
	}
	if !plan.Cancellable {
		t.Fatalf("upload progress should expose cancel")
	}
}

func TestPlanUploadCancelled(t *testing.T) {
	plan := PlanUploadCancelled()
	if plan.Stage != ProgressCancelled || plan.Reason != ReasonCancelled {
		t.Fatalf("cancel state not set, got %+v", plan)
	}
	if plan.OfflineBadge != "badge-cancelled" {
		t.Fatalf("badge missing: %s", plan.OfflineBadge)
	}
}

func TestPlanDownloadRequestToggleBehavior(t *testing.T) {
	controls := Controls{TapToDownload: true, DownloadOnWiFiOnly: true}

	plan := PlanDownloadRequest(NetworkOffline, controls, false)
	if plan.Reason != ReasonOffline || plan.OfflineBadge != "badge-offline" {
		t.Fatalf("offline plan wrong: %+v", plan)
	}

	plan = PlanDownloadRequest(NetworkCellular, controls, false)
	if plan.Reason != ReasonWiFiOnly || plan.Stage != ProgressIdle {
		t.Fatalf("wifi-only plan wrong: %+v", plan)
	}

	plan = PlanDownloadRequest(NetworkWiFi, controls, false)
	if plan.Reason != ReasonTapToDownload || plan.Stage != ProgressIdle {
		t.Fatalf("tap-to-download plan wrong: %+v", plan)
	}

	plan = PlanDownloadRequest(NetworkWiFi, controls, true)
	if plan.Stage != ProgressDownloading || plan.Reason != ReasonReady {
		t.Fatalf("download should start when toggles satisfied: %+v", plan)
	}
}

func TestPlanDownloadNetworkError(t *testing.T) {
	plan := PlanDownloadNetworkError()
	if plan.Reason != ReasonNetworkError || plan.OfflineBadge != "badge-network-error" {
		t.Fatalf("network error plan wrong: %+v", plan)
	}
}
