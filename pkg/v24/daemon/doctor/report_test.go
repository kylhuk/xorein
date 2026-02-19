package doctor

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/aether/code_aether/pkg/v24/daemon"
)

func TestDoctorReportMissingSocket(t *testing.T) {
	mgr := daemon.NewLockManager("/tmp/lock-report")
	d := New(mgr, filepath.Join(t.TempDir(), "missing.sock"))

	report := d.Run()
	if report.SocketExists {
		t.Fatalf("expected socket missing")
	}
	if report.NextAction != "start daemon to create socket" {
		t.Fatalf("next action = %s", report.NextAction)
	}
	if report.Version != defaultDaemonVersion {
		t.Fatalf("unexpected version: %s", report.Version)
	}
	if report.HealthState != HealthStateMissingSocket {
		t.Fatalf("unexpected health state: %s", report.HealthState)
	}
	if report.HealthSummary != "daemon socket missing" {
		t.Fatalf("unexpected health summary: %s", report.HealthSummary)
	}
	if report.LastCrashDetected {
		t.Fatalf("did not expect a crash marker")
	}
	if !report.LastCrashTime.IsZero() {
		t.Fatalf("expected zero crash time; got %s", report.LastCrashTime)
	}
	if report.LastCrashReason != "" {
		t.Fatalf("unexpected crash reason: %s", report.LastCrashReason)
	}
}

func TestDoctorReportStaleSocket(t *testing.T) {
	tmp := t.TempDir()
	socketPath := filepath.Join(tmp, "daemon.sock")
	if err := os.WriteFile(socketPath, []byte(""), 0o600); err != nil {
		t.Fatalf("write socket: %v", err)
	}

	mgr := daemon.NewLockManager("/tmp/lock-report")
	lock, err := mgr.Acquire("doctor", daemon.StateStarting)
	if err != nil {
		t.Fatalf("acquire: %v", err)
	}
	if err := lock.Release(daemon.StateRunning); err != nil {
		t.Fatalf("release: %v", err)
	}

	d := New(mgr, socketPath)
	report := d.Run()
	if !report.SocketExists || !report.StaleSocket {
		t.Fatalf("expected stale socket report: %+v", report)
	}
	if report.NextAction != "remove stale socket and restart daemon" {
		t.Fatalf("next action = %s", report.NextAction)
	}
	if report.Version != defaultDaemonVersion {
		t.Fatalf("unexpected version: %s", report.Version)
	}
	if report.HealthState != HealthStateStaleSocket {
		t.Fatalf("unexpected health state: %s", report.HealthState)
	}
	if report.HealthSummary != "socket exists without an active lock" {
		t.Fatalf("unexpected health summary: %s", report.HealthSummary)
	}
	if !report.LastCrashDetected {
		t.Fatalf("expected crash marker")
	}
	if report.LastCrashReason != "daemon likely crashed while holding the socket" {
		t.Fatalf("unexpected crash reason: %s", report.LastCrashReason)
	}
	info, err := os.Stat(socketPath)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if !report.LastCrashTime.Equal(info.ModTime()) {
		t.Fatalf("unexpected crash time: %s", report.LastCrashTime)
	}
}

func TestDoctorReportActiveLock(t *testing.T) {
	tmp := t.TempDir()
	socketPath := filepath.Join(tmp, "daemon.sock")
	if err := os.WriteFile(socketPath, []byte(""), 0o600); err != nil {
		t.Fatalf("write socket: %v", err)
	}

	mgr := daemon.NewLockManager("/tmp/lock-report")
	lock, err := mgr.Acquire("doctor", daemon.StateStarting)
	if err != nil {
		t.Fatalf("acquire: %v", err)
	}

	d := New(mgr, socketPath)
	report := d.Run()
	if !report.SocketExists || !report.LockActive {
		t.Fatalf("unexpected report: %+v", report)
	}
	if report.NextAction != "lock held by doctor" {
		t.Fatalf("next action = %s", report.NextAction)
	}
	if report.Version != defaultDaemonVersion {
		t.Fatalf("unexpected version: %s", report.Version)
	}
	if report.HealthState != HealthStateRunning {
		t.Fatalf("unexpected health state: %s", report.HealthState)
	}
	if report.HealthSummary != "daemon lock held by doctor" {
		t.Fatalf("unexpected health summary: %s", report.HealthSummary)
	}
	if report.LastCrashDetected {
		t.Fatalf("did not expect crash marker")
	}
	if !report.LastCrashTime.IsZero() {
		t.Fatalf("expected zero crash time; got %s", report.LastCrashTime)
	}

	if err := lock.Release(daemon.StateRunning); err != nil {
		t.Fatalf("cleanup release: %v", err)
	}
}
