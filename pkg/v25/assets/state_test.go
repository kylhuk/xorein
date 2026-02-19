package assets

import "testing"

func TestPlanPreviewTransitions(t *testing.T) {
	plan := PlanPreview(KindAttachment, true)
	if plan.State != StatePreview {
		t.Fatalf("unexpected preview state: %v", plan.State)
	}
	if plan.Reason != ReasonReady {
		t.Fatalf("unexpected preview reason: %v", plan.Reason)
	}

	plan = PlanPreview(KindAvatar, false)
	if plan.State != StateDegraded {
		t.Fatalf("preview degrade state not set: %v", plan.State)
	}
	if plan.Reason != ReasonMissingBlob {
		t.Fatalf("preview degrade reason mismatch: %v", plan.Reason)
	}
}

func TestPlanDownloadTransitions(t *testing.T) {
	plan := PlanDownload(KindEmoji, true)
	if plan.State != StateDownload {
		t.Fatalf("unexpected download state: %v", plan.State)
	}
	if plan.Action != "download" {
		t.Fatalf("unexpected download action: %s", plan.Action)
	}

	plan = PlanDownload(KindAttachment, false)
	if plan.Placeholder != placeholderFor(ReasonMissingBlob) {
		t.Fatalf("placeholder mismatch: %s", plan.Placeholder)
	}
}

func TestPlanDegradedPlaceholderDeterminism(t *testing.T) {
	plan := PlanDegraded(KindAttachment, ReasonRateLimited)
	if plan.Placeholder != placeholderFor(ReasonRateLimited) {
		t.Fatalf("rate-limited placeholder mismatch: %s", plan.Placeholder)
	}
}

func TestTelemetryFieldsExcludesPayload(t *testing.T) {
	plan := PlanDownload(KindAttachment, true)
	fields := TelemetryFields(plan)
	if fields["asset.kind"] != string(KindAttachment) {
		t.Fatalf("unexpected kind field: %s", fields["asset.kind"])
	}
	for _, key := range []string{"asset.payload", "asset.text"} {
		if _, ok := fields[key]; ok {
			t.Fatalf("telemetry leaked plaintext field %q", key)
		}
	}
}
