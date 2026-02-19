package security

import (
	"fmt"
	"sync"
	"time"
)

const (
	// RefusalReasonRateLimitExceeded signals too many requests in the configured window.
	RefusalReasonRateLimitExceeded RefusalReason = "rate_limit_exceeded"
	// RefusalReasonResponseSizeExceeded signals a response would exceed configured bounds.
	RefusalReasonResponseSizeExceeded RefusalReason = "response_size_exceeded"
)

// RateLimitPolicy defines the deterministic rate and response size constraints.
type RateLimitPolicy struct {
	Limit            int           // maximum requests per window
	Window           time.Duration // sliding window duration
	MaxResponseBytes int           // optional response size limit
}

// RateLimiter applies the RateLimitPolicy in a deterministic, single-token window.
type RateLimiter struct {
	policy      RateLimitPolicy
	mu          sync.Mutex
	windowStart time.Time
	count       int
}

// NewRateLimiter builds a RateLimiter guarded by the supplied policy.
func NewRateLimiter(policy RateLimitPolicy) *RateLimiter {
	return &RateLimiter{policy: policy}
}

// CheckRequest enforces the request rate policy for the provided timestamp.
func (r *RateLimiter) CheckRequest(now time.Time) error {
	if r == nil || r.policy.Limit <= 0 || r.policy.Window <= 0 {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.windowStart.IsZero() || now.Sub(r.windowStart) >= r.policy.Window {
		r.windowStart = now
		r.count = 0
	}
	if r.count >= r.policy.Limit {
		return &RefusalError{
			Reason:  RefusalReasonRateLimitExceeded,
			Details: fmt.Sprintf("window started at %s reached %d requests", r.windowStart.Format(time.RFC3339), r.policy.Limit),
		}
	}
	r.count++
	return nil
}

// ValidateResponseSize enforces response payload bounds.
func (r *RateLimiter) ValidateResponseSize(size int) error {
	if r == nil || r.policy.MaxResponseBytes <= 0 {
		return nil
	}
	if size > r.policy.MaxResponseBytes {
		return &RefusalError{
			Reason:  RefusalReasonResponseSizeExceeded,
			Details: fmt.Sprintf("response %d bytes exceeds limit %d", size, r.policy.MaxResponseBytes),
		}
	}
	return nil
}
