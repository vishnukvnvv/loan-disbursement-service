package services

import (
	"math"
	"math/rand"
	"time"
)

const (
	MaxRetries      = 5
	InitialDelay    = 30 * time.Second
	MaxDelay        = 30 * time.Minute
	JitterPercent   = 0.2
	ChannelSwitchAt = 2
)

type RetryPolicy interface {
	CalculateBackoff(retryCount int) time.Duration
	NextRetryTime(retryCount int) time.Time
	IsRetryEligible(lastAttemptTime time.Time, retryCount int) bool
}

type RetryPolicyImpl struct{}

func NewRetryPolicy() RetryPolicy {
	return &RetryPolicyImpl{}
}

func (rp *RetryPolicyImpl) CalculateBackoff(retryCount int) time.Duration {
	if retryCount <= 0 {
		return InitialDelay
	}

	delay := float64(InitialDelay) * math.Pow(2, float64(retryCount))

	if delay > float64(MaxDelay) {
		delay = float64(MaxDelay)
	}

	jitter := delay * JitterPercent * (rand.Float64()*2 - 1)
	finalDelay := time.Duration(delay + jitter)

	return finalDelay
}

func (rp *RetryPolicyImpl) NextRetryTime(retryCount int) time.Time {
	backoff := rp.CalculateBackoff(retryCount)
	return time.Now().Add(backoff)
}

func (rp *RetryPolicyImpl) IsRetryEligible(lastAttemptTime time.Time, retryCount int) bool {
	nextRetryTime := lastAttemptTime.Add(rp.CalculateBackoff(retryCount))
	return time.Now().After(nextRetryTime)
}
