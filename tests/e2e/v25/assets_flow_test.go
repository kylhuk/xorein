package v25

import (
	"testing"

	"github.com/aether/code_aether/pkg/v25/assets"
)

func TestAssetsPreviewDownloadFlow(t *testing.T) {
	preview := assets.PlanPreview(assets.KindAttachment, true)
	if preview.State != assets.StatePreview {
		t.Fatalf("preview state %s", preview.State)
	}
	download := assets.PlanDownload(assets.KindAttachment, true)
	if download.State != assets.StateDownload {
		t.Fatalf("download state %s", download.State)
	}
	if download.Action != "download" {
		t.Fatalf("download action unexpected: %s", download.Action)
	}
}

func TestAssetsDegradedUnavailableFlow(t *testing.T) {
	plan := assets.PlanPreview(assets.KindEmoji, false)
	if plan.State != assets.StateDegraded {
		t.Fatalf("expected degraded state, got %s", plan.State)
	}
	if plan.Reason != assets.ReasonMissingBlob {
		t.Fatalf("unexpected degrade reason %s", plan.Reason)
	}
	if plan.Placeholder == "" {
		t.Fatalf("expected deterministic placeholder for degraded state")
	}
}

func TestAssetsTelemetryNoPlaintextLeakage(t *testing.T) {
	plan := assets.PlanDownload(assets.KindAvatar, true)
	fields := assets.TelemetryFields(plan)
	if _, ok := fields["asset.blob"]; ok {
		t.Fatalf("telemetry leaked blob field")
	}
	if _, ok := fields["asset.content"]; ok {
		t.Fatalf("telemetry leaked content field")
	}
	if len(fields) != 5 {
		t.Fatalf("unexpected telemetry field count %d", len(fields))
	}
}
