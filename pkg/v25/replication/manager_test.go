package replication

import (
	"testing"
	"time"
)

func TestProviderSelectionHonorsRegionAndASN(t *testing.T) {
	now := time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)
	providers := []Provider{
		{ID: "eu-100", Region: "eu-west", ASN: "as-100", Healthy: true},
		{ID: "eu-200", Region: "eu-west", ASN: "as-200", Healthy: true},
		{ID: "us-100", Region: "us-east", ASN: "as-100", Healthy: true},
	}
	mgr, err := NewManager(providers, Config{TargetReplicas: 2, PreferredRegion: "eu-west", AvoidSingleASN: true})
	if err != nil {
		t.Fatalf("manager init: %v", err)
	}
	mgr.SetTimeSource(func() time.Time { return now })
	if err := mgr.Publish("blob-region-asn"); err != nil {
		t.Fatalf("publish failed: %v", err)
	}
	meta, ok := mgr.ReplicaMetadata("blob-region-asn")
	if !ok {
		t.Fatalf("metadata missing")
	}
	if len(meta.Providers) != 2 {
		t.Fatalf("expected 2 replicas, got %d", len(meta.Providers))
	}
	if meta.Providers[0].Region != "eu-west" {
		t.Fatalf("first provider region %q, want eu-west", meta.Providers[0].Region)
	}
	if meta.Providers[0].ASN == meta.Providers[1].ASN {
		t.Fatalf("expected ASN diversity, got %q", meta.Providers)
	}
}

func TestPublishRecordsReplicaMetadata(t *testing.T) {
	now := time.Date(2025, time.February, 2, 0, 0, 0, 0, time.UTC)
	providers := []Provider{
		{ID: "p1", Region: "us-east", ASN: "as-1", Healthy: true},
		{ID: "p2", Region: "us-west", ASN: "as-2", Healthy: true},
	}
	mgr, err := NewManager(providers, Config{TargetReplicas: 2})
	if err != nil {
		t.Fatalf("manager init: %v", err)
	}
	mgr.SetTimeSource(func() time.Time { return now })
	if err := mgr.Publish("blob-metadata"); err != nil {
		t.Fatalf("publish failed: %v", err)
	}
	meta, ok := mgr.ReplicaMetadata("blob-metadata")
	if !ok {
		t.Fatalf("metadata missing")
	}
	if meta.PublishedAt != now || meta.LastVerifiedAt != now {
		t.Fatalf("timestamps not set")
	}
}

func TestVerifyAndRepairRecoversFromChurn(t *testing.T) {
	providers := []Provider{
		{ID: "p1", Region: "us-east", ASN: "as-1", Healthy: true},
		{ID: "p2", Region: "us-west", ASN: "as-2", Healthy: true},
		{ID: "p3", Region: "us-west", ASN: "as-3", Healthy: true},
	}
	mgr, err := NewManager(providers, Config{TargetReplicas: 2})
	if err != nil {
		t.Fatalf("manager init: %v", err)
	}
	if err := mgr.Publish("blob-churn"); err != nil {
		t.Fatalf("publish failed: %v", err)
	}
	// Simulate churn by marking the first replica unhealthy.
	mgr.SetProvider(Provider{ID: "p1", Healthy: false})
	if err := mgr.VerifyAndRepair("blob-churn"); err != nil {
		t.Fatalf("repair failed: %v", err)
	}
	meta, _ := mgr.ReplicaMetadata("blob-churn")
	if len(meta.Providers) != 2 {
		t.Fatalf("expected 2 replicas after repair, got %d", len(meta.Providers))
	}
	for _, record := range meta.Providers {
		if record.ProviderID == "p1" {
			t.Fatalf("stale provider still present")
		}
	}
}

func TestDeterministicRefusalCodes(t *testing.T) {
	mgr, err := NewManager([]Provider{{ID: "solo", Healthy: true}}, Config{TargetReplicas: 2})
	if err != nil {
		t.Fatalf("manager init: %v", err)
	}
	publishErr := mgr.Publish("blob-failure")
	if publishErr == nil {
		t.Fatalf("expected publish refusal")
	}
	if re, ok := publishErr.(RefusalError); !ok || re.Code != RefusalCodeInsufficientProviders {
		t.Fatalf("unexpected refusal %+v", publishErr)
	}
	// Prepare metadata for repair failure.
	mgr, _ = NewManager([]Provider{{ID: "r1", Healthy: true}, {ID: "r2", Healthy: true}}, Config{TargetReplicas: 2})
	if err := mgr.Publish("blob-repair"); err != nil {
		t.Fatalf("publish: %v", err)
	}
	mgr.SetProvider(Provider{ID: "r1", Healthy: false})
	mgr.SetProvider(Provider{ID: "r2", Healthy: false})
	err = mgr.VerifyAndRepair("blob-repair")
	if err == nil {
		t.Fatalf("expected repair refusal")
	}
	if re, ok := err.(RefusalError); !ok || re.Code != RefusalCodeRepairFailure {
		t.Fatalf("unexpected repair refusal %+v", err)
	}
}
