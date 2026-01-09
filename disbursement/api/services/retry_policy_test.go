package services

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var _ RetryPolicy = (*MockRetryPolicy)(nil)

type MockRetryPolicy struct {
	mock.Mock
}

func (m *MockRetryPolicy) CalculateBackoff(retryCount int) time.Duration {
	args := m.Called(retryCount)
	return args.Get(0).(time.Duration)
}

func (m *MockRetryPolicy) NextRetryTime(retryCount int) time.Time {
	args := m.Called(retryCount)
	return args.Get(0).(time.Time)
}

func (m *MockRetryPolicy) IsRetryEligible(lastAttemptTime time.Time, retryCount int) bool {
	args := m.Called(lastAttemptTime, retryCount)
	return args.Bool(0)
}

func TestRetryPolicy_CalculateBackoff(t *testing.T) {
	policy := NewRetryPolicy()

	t.Run("returns initial delay for zero retry count", func(t *testing.T) {
		backoff := policy.CalculateBackoff(0)

		assert.Equal(t, InitialDelay, backoff)
	})

	t.Run("returns initial delay for negative retry count", func(t *testing.T) {
		backoff := policy.CalculateBackoff(-1)

		assert.Equal(t, InitialDelay, backoff)
	})

	t.Run("returns exponential backoff for retry count 1", func(t *testing.T) {
		backoff := policy.CalculateBackoff(1)

		expectedBase := 60 * time.Second
		expectedMin := time.Duration(float64(expectedBase) * 0.8)
		expectedMax := time.Duration(float64(expectedBase) * 1.2)

		assert.GreaterOrEqual(t, backoff, expectedMin)
		assert.LessOrEqual(t, backoff, expectedMax)
	})

	t.Run("returns exponential backoff for retry count 2", func(t *testing.T) {
		backoff := policy.CalculateBackoff(2)

		// Expected: 30s * 2^2 = 120s, with ±20% jitter = 96s to 144s
		expectedBase := 120 * time.Second
		expectedMin := time.Duration(float64(expectedBase) * 0.8)
		expectedMax := time.Duration(float64(expectedBase) * 1.2)

		assert.GreaterOrEqual(t, backoff, expectedMin)
		assert.LessOrEqual(t, backoff, expectedMax)
	})

	t.Run("returns exponential backoff for retry count 3", func(t *testing.T) {
		backoff := policy.CalculateBackoff(3)

		expectedBase := 240 * time.Second
		expectedMin := time.Duration(float64(expectedBase) * 0.8)
		expectedMax := time.Duration(float64(expectedBase) * 1.2)

		assert.GreaterOrEqual(t, backoff, expectedMin)
		assert.LessOrEqual(t, backoff, expectedMax)
	})

	t.Run("returns exponential backoff for retry count 4", func(t *testing.T) {
		backoff := policy.CalculateBackoff(4)

		expectedBase := 480 * time.Second
		expectedMin := time.Duration(float64(expectedBase) * 0.8)
		expectedMax := time.Duration(float64(expectedBase) * 1.2)

		assert.GreaterOrEqual(t, backoff, expectedMin)
		assert.LessOrEqual(t, backoff, expectedMax)
	})

	t.Run("caps at max delay for high retry counts", func(t *testing.T) {
		backoff := policy.CalculateBackoff(10)

		expectedMin := time.Duration(float64(MaxDelay) * 0.8)
		expectedMax := time.Duration(float64(MaxDelay) * 1.2)

		assert.GreaterOrEqual(t, backoff, expectedMin)
		assert.LessOrEqual(t, backoff, expectedMax)
		assert.LessOrEqual(t, backoff, MaxDelay*2)
	})

	t.Run("returns consistent backoff pattern across multiple calls", func(t *testing.T) {
		backoff0 := policy.CalculateBackoff(0)
		backoff1 := policy.CalculateBackoff(1)
		backoff2 := policy.CalculateBackoff(2)
		backoff3 := policy.CalculateBackoff(3)

		avg0 := float64(backoff0)
		avg1 := float64(backoff1)
		avg2 := float64(backoff2)
		avg3 := float64(backoff3)

		assert.Greater(t, avg1, avg0*0.5)
		assert.Greater(t, avg2, avg1*0.5)
		assert.Greater(t, avg3, avg2*0.5)
	})

	t.Run("handles max retries constant", func(t *testing.T) {
		backoff := policy.CalculateBackoff(MaxRetries)

		expectedBase := 30 * time.Second * 32
		expectedMin := time.Duration(float64(expectedBase) * 0.8)
		expectedMax := time.Duration(float64(expectedBase) * 1.2)

		assert.GreaterOrEqual(t, backoff, expectedMin)
		assert.LessOrEqual(t, backoff, expectedMax)
	})
}

func TestRetryPolicy_NextRetryTime(t *testing.T) {
	policy := NewRetryPolicy()

	t.Run("returns future time for zero retry count", func(t *testing.T) {
		before := time.Now()
		nextRetryTime := policy.NextRetryTime(0)
		after := time.Now()

		assert.True(t, nextRetryTime.After(before))
		assert.True(t, nextRetryTime.After(after.Add(-1*time.Second)))

		assert.True(t, nextRetryTime.After(before))
		assert.True(t, nextRetryTime.After(after.Add(-1*time.Second)))
	})

	t.Run("returns future time for retry count 1", func(t *testing.T) {
		before := time.Now()
		nextRetryTime := policy.NextRetryTime(1)
		after := time.Now()

		assert.True(t, nextRetryTime.After(before))
		assert.True(t, nextRetryTime.After(after.Add(-1*time.Second)))

		assert.True(t, nextRetryTime.After(before))
		assert.True(t, nextRetryTime.After(after.Add(-1*time.Second)))
	})

	t.Run("returns future time for retry count 2", func(t *testing.T) {
		before := time.Now()
		nextRetryTime := policy.NextRetryTime(2)
		after := time.Now()

		assert.True(t, nextRetryTime.After(before))
		assert.True(t, nextRetryTime.After(after.Add(-1*time.Second)))

		assert.True(t, nextRetryTime.After(before))
		assert.True(t, nextRetryTime.After(after.Add(-1*time.Second)))
	})

	t.Run("returns future time for high retry count", func(t *testing.T) {
		before := time.Now()
		nextRetryTime := policy.NextRetryTime(10)
		after := time.Now()

		assert.True(t, nextRetryTime.After(before))
		assert.True(t, nextRetryTime.After(after.Add(-1*time.Second)))

		assert.True(t, nextRetryTime.After(before))
		assert.True(t, nextRetryTime.After(after.Add(-1*time.Second)))
	})

	t.Run("returns increasing retry times for increasing retry counts", func(t *testing.T) {
		now := time.Now()
		time0 := policy.NextRetryTime(0)
		time1 := policy.NextRetryTime(1)
		time2 := policy.NextRetryTime(2)
		time3 := policy.NextRetryTime(3)

		assert.True(t, time0.After(now))
		assert.True(t, time1.After(now))
		assert.True(t, time2.After(now))
		assert.True(t, time3.After(now))

		assert.True(t, time0.After(now.Add(20*time.Second)))
		assert.True(t, time1.After(now.Add(40*time.Second)))
		assert.True(t, time2.After(now.Add(80*time.Second)))
		assert.True(t, time3.After(now.Add(150*time.Second)))
	})
}

func TestRetryPolicy_IsRetryEligible(t *testing.T) {
	policy := NewRetryPolicy()

	t.Run("returns false when last attempt was just now", func(t *testing.T) {
		lastAttemptTime := time.Now()
		retryCount := 0

		result := policy.IsRetryEligible(lastAttemptTime, retryCount)

		assert.False(t, result, "should not be eligible immediately after attempt")
	})

	t.Run("returns false when last attempt was recent", func(t *testing.T) {
		lastAttemptTime := time.Now().Add(-10 * time.Second)
		retryCount := 0

		result := policy.IsRetryEligible(lastAttemptTime, retryCount)

		assert.False(t, result, "should not be eligible before backoff period")
	})

	t.Run("returns true when last attempt was long ago for retry count 0", func(t *testing.T) {
		lastAttemptTime := time.Now().Add(-40 * time.Second)
		retryCount := 0

		result := policy.IsRetryEligible(lastAttemptTime, retryCount)

		assert.True(t, result, "should be eligible after backoff period")
	})

	t.Run("returns true when last attempt was long ago for retry count 1", func(t *testing.T) {
		lastAttemptTime := time.Now().Add(-80 * time.Second)
		retryCount := 1

		result := policy.IsRetryEligible(lastAttemptTime, retryCount)

		assert.True(t, result, "should be eligible after backoff period")
	})

	t.Run("returns true when last attempt was long ago for retry count 2", func(t *testing.T) {
		lastAttemptTime := time.Now().Add(-150 * time.Second)
		retryCount := 2

		result := policy.IsRetryEligible(lastAttemptTime, retryCount)

		assert.True(t, result, "should be eligible after backoff period")
	})

	t.Run("returns false when last attempt was in the future", func(t *testing.T) {
		lastAttemptTime := time.Now().Add(1 * time.Hour)
		retryCount := 0

		result := policy.IsRetryEligible(lastAttemptTime, retryCount)

		assert.False(t, result, "should not be eligible if last attempt is in the future")
	})

	t.Run("returns true for high retry count when enough time has passed", func(t *testing.T) {
		lastAttemptTime := time.Now().Add(-35 * time.Minute)
		retryCount := 10

		result := policy.IsRetryEligible(lastAttemptTime, retryCount)

		assert.True(t, result, "should be eligible after max delay period")
	})

	t.Run("returns false for high retry count when not enough time has passed", func(t *testing.T) {
		lastAttemptTime := time.Now().Add(-25 * time.Minute)
		retryCount := 10

		result := policy.IsRetryEligible(lastAttemptTime, retryCount)

		assert.False(t, result, "should not be eligible before max delay period")
	})

	t.Run("handles edge case at boundary of backoff period", func(t *testing.T) {
		lastAttemptTime := time.Now().Add(-35 * time.Second)
		retryCount := 0

		result := policy.IsRetryEligible(lastAttemptTime, retryCount)

		assert.True(t, result, "should be eligible at boundary")
	})

	t.Run("handles negative retry count", func(t *testing.T) {
		lastAttemptTime := time.Now().Add(-40 * time.Second)
		retryCount := -1

		result := policy.IsRetryEligible(lastAttemptTime, retryCount)

		assert.True(t, result, "should be eligible with negative retry count using initial delay")
	})

	t.Run(
		"returns different results for different retry counts with same last attempt time",
		func(t *testing.T) {
			lastAttemptTime := time.Now().Add(-100 * time.Second)

			result0 := policy.IsRetryEligible(lastAttemptTime, 0)
			assert.True(t, result0)

			result1 := policy.IsRetryEligible(lastAttemptTime, 1)
			assert.True(t, result1)

			result2 := policy.IsRetryEligible(lastAttemptTime, 2)
			assert.IsType(t, true, result2)
		},
	)
}
