package phase9

import (
	"strings"
	"sync"
	"testing"
	"time"
)

func TestConfigNormalize(t *testing.T) {
	tests := []struct {
		name       string
		in         Config
		want       Config
		wantErrSub string
	}{
		{
			name: "defaults are applied",
			in:   Config{},
			want: Config{
				ReservationLimit: 256,
				SessionTimeout:   2 * time.Minute,
				MaxBytesPerSec:   1_000_000,
			},
		},
		{
			name: "custom valid values are preserved",
			in: Config{
				ReservationLimit: 10,
				SessionTimeout:   45 * time.Second,
				MaxBytesPerSec:   64_000,
			},
			want: Config{
				ReservationLimit: 10,
				SessionTimeout:   45 * time.Second,
				MaxBytesPerSec:   64_000,
			},
		},
		{
			name: "reservation limit lower bound",
			in: Config{
				ReservationLimit: -1,
				SessionTimeout:   30 * time.Second,
				MaxBytesPerSec:   64_000,
			},
			wantErrSub: "reservation limit must be at least 1",
		},
		{
			name: "session timeout lower bound",
			in: Config{
				ReservationLimit: 1,
				SessionTimeout:   10 * time.Second,
				MaxBytesPerSec:   64_000,
			},
			wantErrSub: "session timeout must be at least 15s",
		},
		{
			name: "bandwidth lower bound",
			in: Config{
				ReservationLimit: 1,
				SessionTimeout:   30 * time.Second,
				MaxBytesPerSec:   1024,
			},
			wantErrSub: "max bytes per second must be at least 16384",
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

func TestServiceReservationAndSnapshotCounters(t *testing.T) {
	svc, err := NewService(Config{ReservationLimit: 2, SessionTimeout: 30 * time.Second, MaxBytesPerSec: 64_000})
	if err != nil {
		t.Fatalf("NewService() unexpected error: %v", err)
	}

	if ok := svc.Reserve(); !ok {
		t.Fatalf("Reserve() first call = false, want true")
	}
	if ok := svc.Reserve(); !ok {
		t.Fatalf("Reserve() second call = false, want true")
	}
	if ok := svc.Reserve(); ok {
		t.Fatalf("Reserve() third call = true, want false when at reservation limit")
	}

	svc.ExpireOne()
	svc.Release()
	svc.Release()

	snap := svc.Snapshot()
	if snap.ReservationLimit != 2 {
		t.Fatalf("Snapshot().ReservationLimit = %d, want 2", snap.ReservationLimit)
	}
	if snap.SessionTimeout != 30*time.Second {
		t.Fatalf("Snapshot().SessionTimeout = %s, want %s", snap.SessionTimeout, 30*time.Second)
	}
	if snap.MaxBytesPerSec != 64_000 {
		t.Fatalf("Snapshot().MaxBytesPerSec = %d, want 64000", snap.MaxBytesPerSec)
	}
	if snap.Active != 0 {
		t.Fatalf("Snapshot().Active = %d, want 0", snap.Active)
	}
	if snap.Rejected != 1 {
		t.Fatalf("Snapshot().Rejected = %d, want 1", snap.Rejected)
	}
	if snap.TimedOut != 1 {
		t.Fatalf("Snapshot().TimedOut = %d, want 1", snap.TimedOut)
	}
	if snap.Established != 2 {
		t.Fatalf("Snapshot().Established = %d, want 2", snap.Established)
	}
}

func TestServiceReserveHonorsLimitUnderConcurrency(t *testing.T) {
	const limit = 3
	svc, err := NewService(Config{ReservationLimit: limit, SessionTimeout: 30 * time.Second, MaxBytesPerSec: 64_000})
	if err != nil {
		t.Fatalf("NewService() unexpected error: %v", err)
	}
	const attempts = limit * 4
	var wg sync.WaitGroup
	results := make(chan bool, attempts)
	startCh := make(chan struct{})
	readyCh := make(chan struct{}, attempts)
	for i := 0; i < attempts; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			readyCh <- struct{}{}
			<-startCh
			results <- svc.Reserve()
		}()
	}
	for i := 0; i < attempts; i++ {
		select {
		case <-readyCh:
		case <-time.After(2 * time.Second):
			t.Fatalf("reserve goroutine not ready")
		}
	}
	close(startCh)
	wg.Wait()
	close(results)

	success := 0
	for ok := range results {
		if ok {
			success++
		}
	}
	snap := svc.Snapshot()
	if success != limit {
		t.Fatalf("expected %d successful reservations, got %d", limit, success)
	}
	if snap.Active > snap.ReservationLimit {
		t.Fatalf("Snapshot().Active = %d, want <= %d", snap.Active, snap.ReservationLimit)
	}
	if snap.Established != uint64(success) {
		t.Fatalf("Snapshot().Established = %d, want %d", snap.Established, success)
	}
	if snap.Rejected != uint64(attempts-success) {
		t.Fatalf("Snapshot().Rejected = %d, want %d", snap.Rejected, attempts-success)
	}
	if uint64(snap.Active) != snap.Established {
		t.Fatalf("Snapshot().Active (%d) and Established (%d) mismatch", snap.Active, snap.Established)
	}
}

func TestServicePrivacyAuditSnapshotIsLogSafe(t *testing.T) {
	svc, err := NewService(Config{ReservationLimit: 2, SessionTimeout: 30 * time.Second, MaxBytesPerSec: 64_000})
	if err != nil {
		t.Fatalf("NewService() unexpected error: %v", err)
	}

	_ = svc.Reserve()
	_ = svc.Reserve()
	_ = svc.Reserve()
	svc.ExpireOne()

	audit := svc.PrivacyAuditSnapshot()
	if audit.Event != "relay_snapshot" {
		t.Fatalf("PrivacyAuditSnapshot().Event = %q, want relay_snapshot", audit.Event)
	}
	if audit.Sensitivity != RelayLogClassOperational {
		t.Fatalf("PrivacyAuditSnapshot().Sensitivity = %q, want %q", audit.Sensitivity, RelayLogClassOperational)
	}
	if audit.ReservationCap != 2 {
		t.Fatalf("PrivacyAuditSnapshot().ReservationCap = %d, want 2", audit.ReservationCap)
	}
	if audit.SessionTimeout != (30 * time.Second).String() {
		t.Fatalf("PrivacyAuditSnapshot().SessionTimeout = %q, want %q", audit.SessionTimeout, (30 * time.Second).String())
	}
	if audit.RateLimitBps != 64_000 {
		t.Fatalf("PrivacyAuditSnapshot().RateLimitBps = %d, want 64000", audit.RateLimitBps)
	}
	if audit.Active != 1 {
		t.Fatalf("PrivacyAuditSnapshot().Active = %d, want 1", audit.Active)
	}
	if audit.Rejected != 1 {
		t.Fatalf("PrivacyAuditSnapshot().Rejected = %d, want 1", audit.Rejected)
	}
	if audit.TimedOut != 1 {
		t.Fatalf("PrivacyAuditSnapshot().TimedOut = %d, want 1", audit.TimedOut)
	}
	if audit.Established != 2 {
		t.Fatalf("PrivacyAuditSnapshot().Established = %d, want 2", audit.Established)
	}
	if len(audit.ResidualRiskIDs) != 1 || audit.ResidualRiskIDs[0] != "R6" {
		t.Fatalf("PrivacyAuditSnapshot().ResidualRiskIDs = %#v, want []string{\"R6\"}", audit.ResidualRiskIDs)
	}
}
