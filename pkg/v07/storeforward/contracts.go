package storeforward

import "time"

type TTLPolicy struct {
	Min time.Duration
	Max time.Duration
}

func (p TTLPolicy) IsValid(ttl time.Duration) bool {
	if ttl < 0 {
		return false
	}
	if p.Min == 0 && p.Max == 0 {
		return ttl >= 0
	}
	if p.Min > 0 && ttl < p.Min {
		return false
	}
	if p.Max > 0 && ttl > p.Max {
		return false
	}
	return true
}

type ReplicationStatus string

const (
	ReplicationStatusHealthy   ReplicationStatus = "replication.healthy"
	ReplicationStatusDegraded  ReplicationStatus = "replication.degraded"
	ReplicationStatusCritical  ReplicationStatus = "replication.critical"
	ReplicationStatusSuspended ReplicationStatus = "replication.suspended"
)

type ReplicationAssessment struct {
	Status       ReplicationStatus
	Reason       string
	ReplicaCount int
	Expected     int
}

func AssessReplication(replicaCount, expected int) ReplicationAssessment {
	status := ReplicationStatusHealthy
	reason := string(status)
	if expected <= 0 {
		status = ReplicationStatusSuspended
		reason = "replication.suspended.no-expected"
	} else if replicaCount < expected && replicaCount > 0 {
		status = ReplicationStatusDegraded
		reason = "replication.degraded.partial"
	} else if replicaCount == 0 {
		status = ReplicationStatusCritical
		reason = "replication.critical.missing"
	}
	return ReplicationAssessment{
		Status:       status,
		Reason:       reason,
		ReplicaCount: replicaCount,
		Expected:     expected,
	}
}

func NeedsRetry(assessment ReplicationAssessment) bool {
	switch assessment.Status {
	case ReplicationStatusHealthy:
		return false
	case ReplicationStatusDegraded:
		return true
	case ReplicationStatusCritical, ReplicationStatusSuspended:
		return true
	default:
		return false
	}
}

func BoundedRetryDelay(attempt int) time.Duration {
	if attempt <= 0 {
		return 0
	}
	delay := time.Duration(attempt) * 250 * time.Millisecond
	if delay > 5*time.Second {
		delay = 5 * time.Second
	}
	return delay
}
