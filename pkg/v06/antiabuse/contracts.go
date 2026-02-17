package antiabuse

import (
	"fmt"
	"math"
	"time"

	"github.com/aether/code_aether/pkg/v06/conformance"
)

type AntiAbuseReasonClass string

const (
	AntiAbuseReasonPoW     AntiAbuseReasonClass = "VA-A1:antiabuse.pow"
	AntiAbuseReasonLimiter AntiAbuseReasonClass = "VA-A2:antiabuse.limiter"
	AntiAbuseReasonScore   AntiAbuseReasonClass = "VA-A3:antiabuse.score"
)

type AntiAbuseContract struct {
	Policy   string
	TaskID   string
	Artifact string
	Gate     conformance.GateID
	Reason   AntiAbuseReasonClass
}

func NewAntiAbuseContract(policy, task, artifact string, gate conformance.GateID, reason AntiAbuseReasonClass) AntiAbuseContract {
	return AntiAbuseContract{
		Policy:   policy,
		TaskID:   task,
		Artifact: artifact,
		Gate:     gate,
		Reason:   reason,
	}
}

func (c AntiAbuseContract) EvidenceAnchor() string {
	return fmt.Sprintf("%s#%s", c.Artifact, c.TaskID)
}

func (c AntiAbuseContract) ReasonLabel() string {
	return string(c.Reason)
}

func AntiAbuseReasonClasses() []AntiAbuseReasonClass {
	return []AntiAbuseReasonClass{
		AntiAbuseReasonPoW,
		AntiAbuseReasonLimiter,
		AntiAbuseReasonScore,
	}
}

type PowProfile struct {
	Difficulty  uint
	EnvelopeMin uint
	EnvelopeMax uint
	ReasonLabel string
}

func (p PowProfile) Adapt(desired uint) uint {
	minBound := p.EnvelopeMin
	maxBound := p.EnvelopeMax
	if minBound == 0 {
		minBound = 1
	}
	if desired < minBound {
		return minBound
	}
	if maxBound > 0 && desired > maxBound {
		return maxBound
	}
	return desired
}

type LimiterDecision struct {
	Status     string
	RetryAfter time.Duration
	Reason     string
}

func DecideLimiter(currentTokens, threshold, burstAllowance int) LimiterDecision {
	status := "antiabuse.limiter.ok"
	reason := status
	retryAfter := time.Duration(0)
	if currentTokens > threshold+burstAllowance {
		status = "antiabuse.limiter.throttled"
		reason = status
		retryAfter = 5 * time.Second
	} else if currentTokens > threshold {
		status = "antiabuse.limiter.throttled"
		reason = status
		retryAfter = 2 * time.Second
	}
	return LimiterDecision{Status: status, RetryAfter: retryAfter, Reason: reason}
}

type PeerScoreDecision struct {
	Score         float64
	Reintegration bool
	Reason        string
	SincePenalty  time.Duration
}

func EvaluatePeerScore(score, threshold float64, penalized bool, lastPenalty time.Time, window time.Duration) PeerScoreDecision {
	reason := "antiabuse.score.allow"
	reintegrate := false
	if penalized {
		reason = "antiabuse.score.penalize"
		if time.Since(lastPenalty) >= window {
			reintegrate = true
			reason = "antiabuse.score.reintegrate"
		}
	}
	return PeerScoreDecision{
		Score:         math.Max(-1, math.Min(score, 1)),
		Reintegration: reintegrate,
		Reason:        reason,
		SincePenalty:  time.Since(lastPenalty),
	}
}
