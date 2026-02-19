package limits

import (
	"fmt"
	"sync/atomic"
	"time"
)

// Scope identifies the request context a budget is counting against.
type Scope string

const (
	// ScopeBackfillVerification applies to the verification sweep that validates backfill segments.
	ScopeBackfillVerification Scope = "backfill_verification"

	// ScopeIndexing applies to in-memory index maintenance for search/backfill coverage.
	ScopeIndexing Scope = "indexing"

	// ScopeDiskGrowth represents disk usage guardrails for Archivist stores.
	ScopeDiskGrowth Scope = "disk_growth"
)

// ScopeLimits describes the CPU/IO ceilings for a request scope.
type ScopeLimits struct {
	CPULimit     time.Duration
	IOLimitBytes int64
}

// Predefined budgets used by the history/search plane.
var (
	// BackfillVerificationBudget keeps verification runs bounded to avoid runaway CPU or IO.
	BackfillVerificationBudget = ScopeLimits{
		CPULimit:     250 * time.Millisecond,
		IOLimitBytes: 16 * 1024 * 1024,
	}

	// IndexingBudget keeps index maintenance bounded for ordinary search updates.
	IndexingBudget = ScopeLimits{
		CPULimit:     200 * time.Millisecond,
		IOLimitBytes: 12 * 1024 * 1024,
	}
)

// ReasonCode enumerates deterministic refusal codes.
type ReasonCode string

const (
	CodeCPUExceeded   ReasonCode = "CPU_LIMIT_EXCEEDED"
	CodeIOExceeded    ReasonCode = "IO_LIMIT_EXCEEDED"
	CodeDiskHardLimit ReasonCode = "DISK_HARD_LIMIT"
)

const (
	remediationCPU  = "Reduce concurrent work or request a larger verification budget before retrying."
	remediationIO   = "Break the workload into smaller IO batches or re-run once IO pressure subsides."
	remediationDisk = "Clean up or archive data to return below the disk hard limit before writing more."
)

var remediationByCode = map[ReasonCode]string{
	CodeCPUExceeded:   remediationCPU,
	CodeIOExceeded:    remediationIO,
	CodeDiskHardLimit: remediationDisk,
}

// Refusal is returned when a resource guard rejects work.
type Refusal struct {
	Code        ReasonCode
	Scope       Scope
	Message     string
	Remediation string
}

// Error implements error.
func (r Refusal) Error() string {
	return fmt.Sprintf("%s[%s]: %s", r.Scope, r.Code, r.Message)
}

// RequestBudget tracks CPU/IO consumption for a scope.
type RequestBudget struct {
	scope         Scope
	cpuLimitNanos int64
	ioLimitBytes  int64
	cpuUsedNanos  atomic.Int64
	ioUsedBytes   atomic.Int64
}

// NewRequestBudget creates a budget for the provided scope and limits.
func NewRequestBudget(scope Scope, limits ScopeLimits) *RequestBudget {
	var cpuLimitNano int64
	if limits.CPULimit > 0 {
		cpuLimitNano = int64(limits.CPULimit)
	}
	return &RequestBudget{
		scope:         scope,
		cpuLimitNanos: cpuLimitNano,
		ioLimitBytes:  limits.IOLimitBytes,
	}
}

// ConsumeCPU records CPU usage and rejects if the budget is exhausted.
func (b *RequestBudget) ConsumeCPU(delta time.Duration) error {
	if delta <= 0 || b.cpuLimitNanos <= 0 {
		return nil
	}
	newUsed := b.cpuUsedNanos.Add(int64(delta))
	if newUsed > b.cpuLimitNanos {
		return b.refuse(CodeCPUExceeded, fmt.Sprintf("consumed %s of %s", time.Duration(newUsed), time.Duration(b.cpuLimitNanos)))
	}
	return nil
}

// ConsumeIO records IO usage and rejects requests that exceed the configured cap.
func (b *RequestBudget) ConsumeIO(bytes int64) error {
	if bytes <= 0 || b.ioLimitBytes <= 0 {
		return nil
	}
	newUsed := b.ioUsedBytes.Add(bytes)
	if newUsed > b.ioLimitBytes {
		return b.refuse(CodeIOExceeded, fmt.Sprintf("consumed %d of %d bytes", newUsed, b.ioLimitBytes))
	}
	return nil
}

func (b *RequestBudget) refuse(code ReasonCode, detail string) error {
	return &Refusal{
		Code:        code,
		Scope:       b.scope,
		Message:     detail,
		Remediation: remediationByCode[code],
	}
}

// DiskGuard ensures safe behavior as disk usage approaches the configured limits.
type DiskGuard struct {
	alarmThreshold int64
	hardLimit      int64
	usage          atomic.Int64
}

// AlarmState describes whether the guard has triggered an alarm.
type AlarmState struct {
	Alarmed    bool
	UsageBytes int64
}

// Disk thresholds used by Archivist measurement.
const (
	DiskAlarmThresholdBytes = 200 * 1024 * 1024 * 1024
	DiskHardLimitBytes      = 220 * 1024 * 1024 * 1024
)

// NewDiskGuard builds a guard populated with the provided thresholds.
func NewDiskGuard(alarmThresholdBytes, hardLimitBytes, initialUsageBytes int64) (*DiskGuard, error) {
	if alarmThresholdBytes < 0 || hardLimitBytes <= 0 {
		return nil, fmt.Errorf("disk thresholds must be positive")
	}
	if alarmThresholdBytes >= hardLimitBytes {
		return nil, fmt.Errorf("alarm threshold must be below hard limit")
	}
	if initialUsageBytes < 0 || initialUsageBytes > hardLimitBytes {
		return nil, fmt.Errorf("starting usage must be between 0 and hard limit")
	}
	guard := &DiskGuard{
		alarmThreshold: alarmThresholdBytes,
		hardLimit:      hardLimitBytes,
	}
	guard.usage.Store(initialUsageBytes)
	return guard, nil
}

// AddUsage increments tracked usage, triggers alarms, and refuses writes above the hard limit.
func (d *DiskGuard) AddUsage(deltaBytes int64) (AlarmState, error) {
	if deltaBytes <= 0 {
		current := d.usage.Load()
		return AlarmState{Alarmed: current >= d.alarmThreshold, UsageBytes: current}, nil
	}
	for {
		current := d.usage.Load()
		if current >= d.hardLimit {
			return AlarmState{Alarmed: true, UsageBytes: current}, d.diskRefusal(current)
		}
		newUsage := current + deltaBytes
		if newUsage >= d.hardLimit {
			if d.usage.CompareAndSwap(current, d.hardLimit) {
				return AlarmState{Alarmed: true, UsageBytes: d.hardLimit}, d.diskRefusal(d.hardLimit)
			}
			continue
		}
		if d.usage.CompareAndSwap(current, newUsage) {
			return AlarmState{Alarmed: newUsage >= d.alarmThreshold, UsageBytes: newUsage}, nil
		}
	}
}

func (d *DiskGuard) diskRefusal(usage int64) error {
	return &Refusal{
		Code:        CodeDiskHardLimit,
		Scope:       ScopeDiskGrowth,
		Message:     fmt.Sprintf("disk usage %d reached hard limit %d", usage, d.hardLimit),
		Remediation: remediationByCode[CodeDiskHardLimit],
	}
}
