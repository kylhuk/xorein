package v25

import (
	"testing"

	"github.com/aether/code_aether/pkg/v25/replication"
)

func TestReplicationPolicyPublishPrefersLocalAndDiversifies(t *testing.T) {
	providers := []replication.Provider{
		{ID: "local-1", Region: "eu-west", ASN: "asn-1", Healthy: true},
		{ID: "local-2", Region: "eu-west", ASN: "asn-2", Healthy: true},
		{ID: "remote", Region: "us-east", ASN: "asn-1", Healthy: true},
	}
	mgr, err := replication.NewManager(providers, replication.Config{TargetReplicas: 2, PreferredRegion: "eu-west", AvoidSingleASN: true})
	if err != nil {
		t.Fatalf("manager init: %v", err)
	}
	if err := mgr.Publish("blob-e2e-policy"); err != nil {
		t.Fatalf("publish failed: %v", err)
	}
	meta, ok := mgr.ReplicaMetadata("blob-e2e-policy")
	if !ok {
		t.Fatalf("metadata missing")
	}
	if len(meta.Providers) != 2 {
		t.Fatalf("unexpected replica count %d", len(meta.Providers))
	}
	if meta.Providers[0].Region != "eu-west" {
		t.Fatalf("expected first replica local, got %s", meta.Providers[0].Region)
	}
	if meta.Providers[0].ASN == meta.Providers[1].ASN {
		t.Fatalf("expected AS diversity, got %s", meta.Providers)
	}
}

func TestReplicationPolicyRepairMaintainsSet(t *testing.T) {
	providers := []replication.Provider{
		{ID: "a", Region: "us-east", ASN: "asn-1", Healthy: true},
		{ID: "b", Region: "us-west", ASN: "asn-2", Healthy: true},
		{ID: "c", Region: "us-west", ASN: "asn-3", Healthy: true},
	}
	mgr, err := replication.NewManager(providers, replication.Config{TargetReplicas: 2})
	if err != nil {
		t.Fatalf("manager init: %v", err)
	}
	if err := mgr.Publish("blob-e2e-repair"); err != nil {
		t.Fatalf("publish failed: %v", err)
	}
	mgr.SetProvider(replication.Provider{ID: "a", Healthy: false})
	if err := mgr.VerifyAndRepair("blob-e2e-repair"); err != nil {
		t.Fatalf("repair failed: %v", err)
	}
	meta, ok := mgr.ReplicaMetadata("blob-e2e-repair")
	if !ok {
		t.Fatalf("metadata missing")
	}
	if len(meta.Providers) != 2 {
		t.Fatalf("unexpected replica count after repair %d", len(meta.Providers))
	}
	for _, record := range meta.Providers {
		if record.ProviderID == "a" {
			t.Fatalf("churned provider still listed")
		}
	}
}
