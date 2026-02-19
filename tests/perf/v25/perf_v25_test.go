package v25

import (
	"testing"
	"time"

	"github.com/aether/code_aether/pkg/v25/blobref"
	"github.com/aether/code_aether/pkg/v25/replication"
)

func TestPerfBlobRefValidationStability(t *testing.T) {
	cases := []struct {
		name string
		ref  blobref.BlobRef
		want blobref.RefusalCode
	}{
		{
			name: "missing algorithm",
			ref:  blobref.BlobRef{},
			want: blobref.RefusalCodeMissingField,
		},
		{
			name: "unsupported algorithm",
			ref:  blobref.BlobRef{HashAlgorithm: "MD5", ContentHash: "c", MimeType: "a", ChunkSize: 1, ChunkProfile: "p"},
			want: blobref.RefusalCodeUnsupportedAlgorithm,
		},
		{
			name: "negative size",
			ref:  blobref.BlobRef{HashAlgorithm: "BLAKE3", ContentHash: "c", MimeType: "a", ChunkSize: 1, ChunkProfile: "p", Size: -10},
			want: blobref.RefusalCodeInvalidSize,
		},
		{
			name: "missing mime",
			ref:  blobref.BlobRef{HashAlgorithm: "BLAKE3", ContentHash: "c", Size: 1, ChunkSize: 1, ChunkProfile: "p"},
			want: blobref.RefusalCodeMissingField,
		},
		{
			name: "invalid chunk size",
			ref:  blobref.BlobRef{HashAlgorithm: "BLAKE3", ContentHash: "c", MimeType: "a", ChunkSize: 0, ChunkProfile: "p"},
			want: blobref.RefusalCodeInvalidChunkSize,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			var observed blobref.RefusalCode
			for iter := 0; iter < 6; iter++ {
				err := blobref.ValidateBlobRef(tc.ref)
				if err == nil {
					t.Fatalf("iteration %d: expected refusal", iter)
				}
				re, ok := err.(*blobref.RefusalError)
				if !ok {
					t.Fatalf("iteration %d: expected RefusalError, got %T", iter, err)
				}
				if iter == 0 {
					observed = re.Code
					continue
				}
				if re.Code != observed {
					t.Fatalf("iteration %d: code drifted %s -> %s", iter, observed, re.Code)
				}
			}
			if observed != tc.want {
				t.Fatalf("expected %s, got %s", tc.want, observed)
			}
		})
	}
}

func TestPerfReplicationSelectionDeterminism(t *testing.T) {
	const iterations = 4
	t.Run("publish-stable-order", func(t *testing.T) {
		now := time.Date(2025, time.March, 5, 1, 0, 0, 0, time.UTC)
		providers := []replication.Provider{
			{ID: "eu-1", Region: "eu-west", ASN: "asn-eu-1", Healthy: true},
			{ID: "eu-2", Region: "eu-west", ASN: "asn-eu-2", Healthy: true},
			{ID: "us-1", Region: "us-east", ASN: "asn-us-1", Healthy: true},
		}
		mgr, err := replication.NewManager(providers, replication.Config{TargetReplicas: 2, PreferredRegion: "eu-west", AvoidSingleASN: true})
		if err != nil {
			t.Fatalf("manager init: %v", err)
		}
		mgr.SetTimeSource(func() time.Time { return now })
		var expected []string
		for iter := 0; iter < iterations; iter++ {
			if err := mgr.Publish("perf-blob"); err != nil {
				t.Fatalf("publish %d: %v", iter, err)
			}
			meta, ok := mgr.ReplicaMetadata("perf-blob")
			if !ok {
				t.Fatalf("iteration %d: metadata missing", iter)
			}
			ids := providerIDs(meta.Providers)
			if iter == 0 {
				expected = append([]string(nil), ids...)
				continue
			}
			if !equalStrings(expected, ids) {
				t.Fatalf("provider order changed: expected %v got %v", expected, ids)
			}
		}
	})

	t.Run("repair-refusal-stable", func(t *testing.T) {
		now := time.Date(2025, time.April, 1, 2, 0, 0, 0, time.UTC)
		providers := []replication.Provider{
			{ID: "p1", Region: "us-east", ASN: "asn-a", Healthy: true},
			{ID: "p2", Region: "us-west", ASN: "asn-b", Healthy: true},
		}
		mgr, err := replication.NewManager(providers, replication.Config{TargetReplicas: 2})
		if err != nil {
			t.Fatalf("manager init: %v", err)
		}
		mgr.SetTimeSource(func() time.Time { return now })
		if err := mgr.Publish("perf-repair"); err != nil {
			t.Fatalf("publish: %v", err)
		}
		for _, id := range []string{"p1", "p2"} {
			mgr.SetProvider(replication.Provider{ID: id, Healthy: false})
		}
		for iter := 0; iter < 5; iter++ {
			err := mgr.VerifyAndRepair("perf-repair")
			if err == nil {
				t.Fatalf("iteration %d: expected refusal", iter)
			}
			re, ok := err.(replication.RefusalError)
			if !ok {
				t.Fatalf("iteration %d: unexpected error type %T", iter, err)
			}
			if re.Code != replication.RefusalCodeRepairFailure {
				t.Fatalf("iteration %d: expected %s got %s", iter, replication.RefusalCodeRepairFailure, re.Code)
			}
		}
	})
}

func providerIDs(records []replication.ReplicaProviderRecord) []string {
	ids := make([]string, len(records))
	for i, record := range records {
		ids[i] = record.ProviderID
	}
	return ids
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
