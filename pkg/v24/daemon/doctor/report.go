package doctor

import (
	"fmt"
	"os"
	"time"

	"github.com/aether/code_aether/pkg/v24/daemon"
)

// HealthState enumerates deterministic doctor probe statuses.
type HealthState string

const (
	HealthStateUnknown       HealthState = "HEALTH_UNKNOWN"
	HealthStateRunning       HealthState = "HEALTH_RUNNING"
	HealthStateMissingSocket HealthState = "HEALTH_SOCKET_MISSING"
	HealthStateStaleSocket   HealthState = "HEALTH_SOCKET_STALE"
)

const defaultDaemonVersion = "v2.4-daemon"

// Report captures the deterministic doctor output for the daemon lifecycle.
type Report struct {
	SocketPath        string
	SocketExists      bool
	SocketPermission  os.FileMode
	SocketError       string
	LockOwner         string
	LockActive        bool
	StaleSocket       bool
	NextAction        string
	DiagnosticMessage string
	Version           string
	HealthState       HealthState
	HealthSummary     string
	LastCrashDetected bool
	LastCrashTime     time.Time
	LastCrashReason   string
}

// Doctor inspects socket and lock health information.
type Doctor struct {
	Manager    *daemon.LockManager
	SocketPath string
}

// New returns a doctor for the provided lock path and socket path.
func New(manager *daemon.LockManager, socketPath string) *Doctor {
	return &Doctor{Manager: manager, SocketPath: socketPath}
}

// Run executes checks and returns a deterministic report.
func (d *Doctor) Run() Report {
	report := Report{SocketPath: d.SocketPath}

	info, err := os.Stat(d.SocketPath)
	if err == nil {
		report.SocketExists = true
		report.SocketPermission = info.Mode().Perm()
		report.SocketError = ""
		if info.Mode()&os.ModeSocket == 0 {
			report.DiagnosticMessage = "path exists but is not a socket"
		}
	} else {
		report.SocketExists = false
		report.SocketPermission = 0
		report.SocketError = err.Error()
	}

	report.LockOwner = d.Manager.Owner()
	report.LockActive = d.Manager.IsLocked()

	if report.SocketExists && !report.LockActive {
		report.StaleSocket = true
	}

	report.evaluateHealth(info)

	return report
}

func (r *Report) evaluateHealth(info os.FileInfo) {
	r.Version = defaultDaemonVersion
	switch {
	case r.StaleSocket:
		r.HealthState = HealthStateStaleSocket
		r.HealthSummary = "socket exists without an active lock"
		r.LastCrashDetected = true
		r.LastCrashReason = "daemon likely crashed while holding the socket"
		if info != nil {
			r.LastCrashTime = info.ModTime()
		}
		r.NextAction = "remove stale socket and restart daemon"
	case !r.SocketExists:
		r.HealthState = HealthStateMissingSocket
		r.HealthSummary = "daemon socket missing"
		r.NextAction = "start daemon to create socket"
	case r.LockActive:
		owner := r.LockOwner
		if owner == "" {
			owner = "unknown"
		}
		r.HealthState = HealthStateRunning
		r.HealthSummary = fmt.Sprintf("daemon lock held by %s", owner)
		r.NextAction = fmt.Sprintf("lock held by %s", owner)
	default:
		r.HealthState = HealthStateUnknown
		r.HealthSummary = "daemon status requires monitoring"
		r.NextAction = "monitor daemon health"
	}
}
