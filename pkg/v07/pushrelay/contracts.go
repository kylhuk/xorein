package pushrelay

import "time"

type Metadata map[string]string

func IsMetadataMinimal(meta Metadata) bool {
	return len(meta) <= 3
}

type TokenStatus string

const (
	TokenStatusActive   TokenStatus = "pushrelay.token.active"
	TokenStatusStale    TokenStatus = "pushrelay.token.stale"
	TokenStatusRotating TokenStatus = "pushrelay.token.rotating"
)

type TokenLifecycle struct {
	Status   TokenStatus
	Age      time.Duration
	Rotating bool
}

func EvaluateTokenLifecycle(age time.Duration) TokenLifecycle {
	status := TokenStatusActive
	rotating := false
	if age > time.Hour {
		status = TokenStatusStale
	} else if age > 30*time.Minute {
		status = TokenStatusRotating
		rotating = true
	}
	return TokenLifecycle{Status: status, Age: age, Rotating: rotating}
}

func BackoffClassification(attempt int) string {
	switch {
	case attempt <= 0:
		return "backoff.immediate"
	case attempt <= 3:
		return "backoff.short"
	case attempt <= 6:
		return "backoff.medium"
	default:
		return "backoff.long"
	}
}
