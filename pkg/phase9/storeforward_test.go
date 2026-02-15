package phase9

import (
	"strings"
	"testing"
	"time"
)

func TestStoreConfigNormalize(t *testing.T) {
	tests := []struct {
		name       string
		in         StoreConfig
		want       StoreConfig
		wantErrSub string
	}{
		{
			name: "defaults are applied",
			in:   StoreConfig{},
			want: StoreConfig{
				RetentionTTL:    24 * time.Hour,
				MaxMessages:     10_000,
				MaxBytes:        128 * 1024 * 1024,
				MaxPayloadBytes: 64 * 1024,
			},
		},
		{
			name: "custom valid values are preserved",
			in: StoreConfig{
				RetentionTTL:    5 * time.Minute,
				MaxMessages:     3,
				MaxBytes:        32,
				MaxPayloadBytes: 16,
			},
			want: StoreConfig{
				RetentionTTL:    5 * time.Minute,
				MaxMessages:     3,
				MaxBytes:        32,
				MaxPayloadBytes: 16,
			},
		},
		{
			name: "retention ttl lower bound",
			in: StoreConfig{
				RetentionTTL:    59 * time.Second,
				MaxMessages:     1,
				MaxBytes:        1,
				MaxPayloadBytes: 1,
			},
			wantErrSub: "retention ttl must be at least 1m",
		},
		{
			name: "max messages lower bound",
			in: StoreConfig{
				RetentionTTL:    time.Minute,
				MaxMessages:     -1,
				MaxBytes:        1,
				MaxPayloadBytes: 1,
			},
			wantErrSub: "max messages must be at least 1",
		},
		{
			name: "max bytes lower bound",
			in: StoreConfig{
				RetentionTTL:    time.Minute,
				MaxMessages:     1,
				MaxBytes:        -1,
				MaxPayloadBytes: 1,
			},
			wantErrSub: "max bytes must be at least 1",
		},
		{
			name: "max payload lower bound",
			in: StoreConfig{
				RetentionTTL:    time.Minute,
				MaxMessages:     1,
				MaxBytes:        1,
				MaxPayloadBytes: -1,
			},
			wantErrSub: "max payload bytes must be at least 1",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := tc.in.Normalize()
			if tc.wantErrSub != "" {
				if err == nil {
					t.Fatalf("Normalize() error = nil, want substring %q", tc.wantErrSub)
				}
				if !strings.Contains(err.Error(), tc.wantErrSub) {
					t.Fatalf("Normalize() error = %q, want substring %q", err.Error(), tc.wantErrSub)
				}
				return
			}
			if err != nil {
				t.Fatalf("Normalize() unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("Normalize() = %+v, want %+v", got, tc.want)
			}
		})
	}
}

func TestStoreServiceTTLAndQuotaDeterminism(t *testing.T) {
	now := time.Date(2026, time.January, 1, 12, 0, 0, 0, time.UTC)

	svc, err := NewStoreService(StoreConfig{
		RetentionTTL:    2 * time.Minute,
		MaxMessages:     2,
		MaxBytes:        5,
		MaxPayloadBytes: 4,
	})
	if err != nil {
		t.Fatalf("NewStoreService() unexpected error: %v", err)
	}

	if got := svc.Store(now, "alice", []byte("aa")); !got.Stored || got.Reason != "stored" {
		t.Fatalf("Store(first) = %+v, want Stored=true Reason=stored", got)
	}

	if got := svc.Store(now.Add(30*time.Second), "bob", []byte("bbb")); !got.Stored || got.DroppedByCap != 0 {
		t.Fatalf("Store(second) = %+v, want stored without cap drop", got)
	}

	third := svc.Store(now.Add(time.Minute), "alice", []byte("cccc"))
	if !third.Stored {
		t.Fatalf("Store(third) Stored = false, want true")
	}
	if third.DroppedByCap != 2 {
		t.Fatalf("Store(third) DroppedByCap = %d, want 2 (deterministic oldest-first eviction)", third.DroppedByCap)
	}

	snapAfterQuota := svc.Snapshot(now.Add(time.Minute))
	if snapAfterQuota.QueuedMessages != 1 {
		t.Fatalf("Snapshot().QueuedMessages after quota = %d, want 1", snapAfterQuota.QueuedMessages)
	}
	if snapAfterQuota.QueuedBytes != 4 {
		t.Fatalf("Snapshot().QueuedBytes after quota = %d, want 4", snapAfterQuota.QueuedBytes)
	}
	if snapAfterQuota.DroppedByCap != 2 {
		t.Fatalf("Snapshot().DroppedByCap after quota = %d, want 2", snapAfterQuota.DroppedByCap)
	}

	snapAfterTTL := svc.Snapshot(now.Add(3 * time.Minute))
	if snapAfterTTL.QueuedMessages != 0 {
		t.Fatalf("Snapshot().QueuedMessages after ttl = %d, want 0", snapAfterTTL.QueuedMessages)
	}
	if snapAfterTTL.DroppedByTTL != 1 {
		t.Fatalf("Snapshot().DroppedByTTL after ttl = %d, want 1", snapAfterTTL.DroppedByTTL)
	}
}

func TestStoreServiceRejectsInvalidWritesAndKeepsPayloadOpaque(t *testing.T) {
	now := time.Date(2026, time.January, 1, 12, 0, 0, 0, time.UTC)

	svc, err := NewStoreService(StoreConfig{
		RetentionTTL:    10 * time.Minute,
		MaxMessages:     4,
		MaxBytes:        16,
		MaxPayloadBytes: 4,
	})
	if err != nil {
		t.Fatalf("NewStoreService() unexpected error: %v", err)
	}

	rejects := []struct {
		name       string
		recipient  string
		ciphertext []byte
		wantReason string
	}{
		{name: "missing recipient", recipient: "", ciphertext: []byte("ab"), wantReason: "recipient required"},
		{name: "missing ciphertext", recipient: "alice", ciphertext: nil, wantReason: "ciphertext required"},
		{name: "payload too large", recipient: "alice", ciphertext: []byte("12345"), wantReason: "payload exceeds max"},
	}

	for _, tc := range rejects {
		t.Run(tc.name, func(t *testing.T) {
			got := svc.Store(now, tc.recipient, tc.ciphertext)
			if got.Stored {
				t.Fatalf("Store() Stored = true, want false for %s", tc.name)
			}
			if got.Reason != tc.wantReason {
				t.Fatalf("Store() Reason = %q, want %q", got.Reason, tc.wantReason)
			}
		})
	}

	before := []byte{0x01, 0x02, 0x03}
	stored := svc.Store(now.Add(time.Second), "alice", before)
	if !stored.Stored {
		t.Fatalf("Store(valid) = %+v, want Stored=true", stored)
	}
	before[0] = 0xFF

	drained := svc.DrainRecipient(now.Add(2*time.Second), "alice")
	if len(drained) != 1 {
		t.Fatalf("DrainRecipient() len = %d, want 1", len(drained))
	}
	if drained[0][0] != 0x01 {
		t.Fatalf("DrainRecipient() returned mutated payload: got first byte %#x, want %#x", drained[0][0], byte(0x01))
	}

	snap := svc.Snapshot(now.Add(2 * time.Second))
	if snap.RejectedWrites != uint64(len(rejects)) {
		t.Fatalf("Snapshot().RejectedWrites = %d, want %d", snap.RejectedWrites, len(rejects))
	}
}

func TestStoreServicePrivacyAuditSnapshotIsLogSafe(t *testing.T) {
	now := time.Date(2026, time.January, 1, 12, 0, 0, 0, time.UTC)

	svc, err := NewStoreService(StoreConfig{
		RetentionTTL:    2 * time.Minute,
		MaxMessages:     2,
		MaxBytes:        5,
		MaxPayloadBytes: 4,
	})
	if err != nil {
		t.Fatalf("NewStoreService() unexpected error: %v", err)
	}

	_ = svc.Store(now, "alice", []byte("aa"))
	_ = svc.Store(now.Add(30*time.Second), "bob", []byte("bbb"))
	_ = svc.Store(now.Add(time.Minute), "alice", []byte("cccc"))

	audit := svc.PrivacyAuditSnapshot(now.Add(time.Minute))
	if audit.Event != "storeforward_snapshot" {
		t.Fatalf("PrivacyAuditSnapshot().Event = %q, want storeforward_snapshot", audit.Event)
	}
	if audit.Sensitivity != StoreLogClassOperational {
		t.Fatalf("PrivacyAuditSnapshot().Sensitivity = %q, want %q", audit.Sensitivity, StoreLogClassOperational)
	}
	if audit.QueueSensitivity != StoreLogClassRestricted {
		t.Fatalf("PrivacyAuditSnapshot().QueueSensitivity = %q, want %q", audit.QueueSensitivity, StoreLogClassRestricted)
	}
	if audit.RetentionTTL != (2 * time.Minute).String() {
		t.Fatalf("PrivacyAuditSnapshot().RetentionTTL = %q, want %q", audit.RetentionTTL, (2 * time.Minute).String())
	}
	if audit.MaxMessages != 2 {
		t.Fatalf("PrivacyAuditSnapshot().MaxMessages = %d, want 2", audit.MaxMessages)
	}
	if audit.MaxBytes != 5 {
		t.Fatalf("PrivacyAuditSnapshot().MaxBytes = %d, want 5", audit.MaxBytes)
	}
	if audit.MaxPayloadBytes != 4 {
		t.Fatalf("PrivacyAuditSnapshot().MaxPayloadBytes = %d, want 4", audit.MaxPayloadBytes)
	}
	if audit.QueuedMessages != 1 {
		t.Fatalf("PrivacyAuditSnapshot().QueuedMessages = %d, want 1", audit.QueuedMessages)
	}
	if audit.QueuedBytes != 4 {
		t.Fatalf("PrivacyAuditSnapshot().QueuedBytes = %d, want 4", audit.QueuedBytes)
	}
	if audit.DroppedByCap != 2 {
		t.Fatalf("PrivacyAuditSnapshot().DroppedByCap = %d, want 2", audit.DroppedByCap)
	}
	if audit.DroppedByTTL != 0 {
		t.Fatalf("PrivacyAuditSnapshot().DroppedByTTL = %d, want 0", audit.DroppedByTTL)
	}
	if audit.RejectedWrites != 0 {
		t.Fatalf("PrivacyAuditSnapshot().RejectedWrites = %d, want 0", audit.RejectedWrites)
	}
	if len(audit.ResidualRiskIDs) != 1 || audit.ResidualRiskIDs[0] != "R6" {
		t.Fatalf("PrivacyAuditSnapshot().ResidualRiskIDs = %#v, want []string{\"R6\"}", audit.ResidualRiskIDs)
	}
}
