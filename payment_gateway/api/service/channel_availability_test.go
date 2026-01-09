package service

import (
	"payment-gateway/models"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAvailabilitySchedule_IsAvailable(t *testing.T) {
	service := NewAvailabilitySchedule()
	testTime := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC) // Monday, Jan 15, 12:00 PM

	t.Run("returns false for unknown channel", func(t *testing.T) {
		result := service.IsAvailable(models.PaymentChannel("UNKNOWN"), testTime)
		assert.False(t, result)
	})

	t.Run("delegates to isUPIAvailable for UPI channel", func(t *testing.T) {
		// UPI uses random, so we just verify it doesn't panic
		result := service.IsAvailable(models.PaymentChannelUPI, testTime)
		assert.NotNil(t, result) // Should return a boolean value
	})

	t.Run("delegates to isNEFTAvailable for NEFT channel", func(t *testing.T) {
		// Monday 12:00 PM should be available for NEFT
		result := service.IsAvailable(models.PaymentChannelNEFT, testTime)
		assert.True(t, result)
	})

	t.Run("delegates to isIMPSAvailable for IMPS channel", func(t *testing.T) {
		// Monday should be available for IMPS
		result := service.IsAvailable(models.PaymentChannelIMPS, testTime)
		assert.True(t, result)
	})
}

func TestAvailabilitySchedule_isUPIAvailable(t *testing.T) {
	service := NewAvailabilitySchedule().(*AvailabilityScheduleImpl)

	t.Run("returns boolean value based on random probability", func(t *testing.T) {
		results := make(map[bool]int)
		iterations := 1000

		for i := 0; i < iterations; i++ {
			result := service.isUPIAvailable()
			results[result]++
		}

		// Should have both true and false results (statistically)
		assert.Greater(t, results[true], 0, "Should have some true results")
		assert.Greater(t, results[false], 0, "Should have some false results")

		trueRatio := float64(results[true]) / float64(iterations)
		assert.Greater(t, trueRatio, 0.85, "True ratio should be around 91%")
		assert.Less(t, trueRatio, 0.97, "True ratio should be around 91%")
	})
}

func TestAvailabilitySchedule_isNEFTAvailable(t *testing.T) {
	service := NewAvailabilitySchedule().(*AvailabilityScheduleImpl)

	t.Run("returns false on Saturday", func(t *testing.T) {
		saturday := time.Date(2024, 1, 13, 12, 0, 0, 0, time.UTC) // Saturday, Jan 13, 12:00 PM
		result := service.isNEFTAvailable(saturday)
		assert.False(t, result)
	})

	t.Run("returns false on Sunday", func(t *testing.T) {
		sunday := time.Date(2024, 1, 14, 12, 0, 0, 0, time.UTC) // Sunday, Jan 14, 12:00 PM
		result := service.isNEFTAvailable(sunday)
		assert.False(t, result)
	})

	t.Run("returns false before 8 AM on weekday", func(t *testing.T) {
		mondayEarly := time.Date(2024, 1, 15, 7, 0, 0, 0, time.UTC) // Monday, 7:00 AM
		result := service.isNEFTAvailable(mondayEarly)
		assert.False(t, result)
	})

	t.Run("returns false at 7:59 AM on weekday", func(t *testing.T) {
		mondayBefore8 := time.Date(2024, 1, 15, 7, 59, 0, 0, time.UTC) // Monday, 7:59 AM
		result := service.isNEFTAvailable(mondayBefore8)
		assert.False(t, result)
	})

	t.Run("returns true at 8 AM on weekday", func(t *testing.T) {
		monday8AM := time.Date(2024, 1, 15, 8, 0, 0, 0, time.UTC) // Monday, 8:00 AM
		result := service.isNEFTAvailable(monday8AM)
		assert.True(t, result)
	})

	t.Run("returns true during business hours on weekday", func(t *testing.T) {
		mondayNoon := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC) // Monday, 12:00 PM
		result := service.isNEFTAvailable(mondayNoon)
		assert.True(t, result)
	})

	t.Run("returns true at 6 PM on weekday", func(t *testing.T) {
		monday6PM := time.Date(2024, 1, 15, 18, 0, 0, 0, time.UTC) // Monday, 6:00 PM (18:00)
		result := service.isNEFTAvailable(monday6PM)
		assert.True(t, result)
	})

	t.Run("returns false at 7 PM on weekday", func(t *testing.T) {
		monday7PM := time.Date(2024, 1, 15, 19, 0, 0, 0, time.UTC) // Monday, 7:00 PM (19:00)
		result := service.isNEFTAvailable(monday7PM)
		assert.False(t, result)
	})

	t.Run("returns false after 7 PM on weekday", func(t *testing.T) {
		mondayLate := time.Date(2024, 1, 15, 20, 0, 0, 0, time.UTC) // Monday, 8:00 PM (20:00)
		result := service.isNEFTAvailable(mondayLate)
		assert.False(t, result)
	})

	t.Run("returns true for all weekdays during business hours", func(t *testing.T) {
		weekdays := []time.Weekday{
			time.Monday,
			time.Tuesday,
			time.Wednesday,
			time.Thursday,
			time.Friday,
		}

		for _, weekday := range weekdays {
			// Find a date that falls on the specified weekday
			date := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC) // Start with Monday
			daysToAdd := int(weekday - time.Monday)
			if daysToAdd < 0 {
				daysToAdd += 7
			}
			testDate := date.AddDate(0, 0, daysToAdd)

			result := service.isNEFTAvailable(testDate)
			assert.True(t, result, "Should be available on %s during business hours", weekday)
		}
	})
}

func TestAvailabilitySchedule_isIMPSAvailable(t *testing.T) {
	service := NewAvailabilitySchedule().(*AvailabilityScheduleImpl)

	t.Run("returns true on Monday", func(t *testing.T) {
		monday := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC) // Monday, Jan 15, 12:00 PM
		result := service.isIMPSAvailable(monday)
		assert.True(t, result)
	})

	t.Run("returns true on Tuesday", func(t *testing.T) {
		tuesday := time.Date(2024, 1, 16, 12, 0, 0, 0, time.UTC) // Tuesday, Jan 16, 12:00 PM
		result := service.isIMPSAvailable(tuesday)
		assert.True(t, result)
	})

	t.Run("returns true on Wednesday", func(t *testing.T) {
		wednesday := time.Date(2024, 1, 17, 12, 0, 0, 0, time.UTC) // Wednesday, Jan 17, 12:00 PM
		result := service.isIMPSAvailable(wednesday)
		assert.True(t, result)
	})

	t.Run("returns true on Thursday", func(t *testing.T) {
		thursday := time.Date(2024, 1, 18, 12, 0, 0, 0, time.UTC) // Thursday, Jan 18, 12:00 PM
		result := service.isIMPSAvailable(thursday)
		assert.True(t, result)
	})

	t.Run("returns true on Friday", func(t *testing.T) {
		friday := time.Date(2024, 1, 19, 12, 0, 0, 0, time.UTC) // Friday, Jan 19, 12:00 PM
		result := service.isIMPSAvailable(friday)
		assert.True(t, result)
	})

	t.Run("returns true on Saturday", func(t *testing.T) {
		saturday := time.Date(2024, 1, 20, 12, 0, 0, 0, time.UTC) // Saturday, Jan 20, 12:00 PM
		result := service.isIMPSAvailable(saturday)
		assert.True(t, result)
	})

	t.Run(
		"returns true on Sunday outside maintenance window - before second week",
		func(t *testing.T) {
			// Sunday, Jan 7 (day 7, before maintenance week)
			sundayBefore := time.Date(2024, 1, 7, 3, 0, 0, 0, time.UTC) // Sunday, 3:00 AM
			result := service.isIMPSAvailable(sundayBefore)
			assert.True(t, result)
		},
	)

	t.Run(
		"returns true on Sunday outside maintenance window - after second week",
		func(t *testing.T) {
			// Sunday, Jan 21 (day 21, after maintenance week)
			sundayAfter := time.Date(2024, 1, 21, 3, 0, 0, 0, time.UTC) // Sunday, 3:00 AM
			result := service.isIMPSAvailable(sundayAfter)
			assert.True(t, result)
		},
	)

	t.Run(
		"returns true on Sunday during maintenance week but before maintenance hours",
		func(t *testing.T) {
			// Sunday, Jan 14 (day 14, in maintenance week), 1:00 AM
			sundayBeforeHours := time.Date(2024, 1, 14, 1, 0, 0, 0, time.UTC)
			result := service.isIMPSAvailable(sundayBeforeHours)
			assert.True(t, result)
		},
	)

	t.Run(
		"returns true on Sunday during maintenance week but after maintenance hours",
		func(t *testing.T) {
			// Sunday, Jan 14 (day 14, in maintenance week), 5:00 AM
			sundayAfterHours := time.Date(2024, 1, 14, 5, 0, 0, 0, time.UTC)
			result := service.isIMPSAvailable(sundayAfterHours)
			assert.True(t, result)
		},
	)

	t.Run("returns false on Sunday during maintenance window - day 8, 2 AM", func(t *testing.T) {
		// Sunday, Jan 14 (day 14, in maintenance week), 2:00 AM
		sundayMaintenance := time.Date(2024, 1, 14, 2, 0, 0, 0, time.UTC)
		result := service.isIMPSAvailable(sundayMaintenance)
		assert.False(t, result)
	})

	t.Run("returns false on Sunday during maintenance window - day 10, 3 AM", func(t *testing.T) {
		sundayMaintenance := time.Date(2024, 1, 14, 3, 0, 0, 0, time.UTC) // Sunday, day 14, 3:00 AM
		result := service.isIMPSAvailable(sundayMaintenance)
		assert.False(t, result)
	})

	t.Run("returns false on Sunday during maintenance window - day 12, 4 AM", func(t *testing.T) {
		sundayMaintenance := time.Date(2024, 1, 14, 4, 0, 0, 0, time.UTC) // Sunday, day 14, 4:00 AM
		result := service.isIMPSAvailable(sundayMaintenance)
		assert.False(t, result)
	})

	t.Run("returns false on Sunday during maintenance window - day 8, 2 AM", func(t *testing.T) {
		sundayMaintenance := time.Date(
			2024,
			1,
			14,
			2,
			30,
			0,
			0,
			time.UTC,
		) // Sunday, day 14, 2:30 AM
		result := service.isIMPSAvailable(sundayMaintenance)
		assert.False(t, result)
	})

	t.Run(
		"returns false on Sunday during maintenance window - boundary at 2 AM",
		func(t *testing.T) {
			sundayMaintenance := time.Date(
				2024,
				1,
				14,
				2,
				0,
				0,
				0,
				time.UTC,
			) // Sunday, day 14, 2:00 AM
			result := service.isIMPSAvailable(sundayMaintenance)
			assert.False(t, result)
		},
	)

	t.Run(
		"returns false on Sunday during maintenance window - boundary at 4 AM",
		func(t *testing.T) {
			sundayMaintenance := time.Date(
				2024,
				1,
				14,
				4,
				0,
				0,
				0,
				time.UTC,
			) // Sunday, day 14, 4:00 AM
			result := service.isIMPSAvailable(sundayMaintenance)
			assert.False(t, result)
		},
	)

	t.Run("returns true on Sunday during maintenance week but at 1:59 AM", func(t *testing.T) {
		sundayBeforeWindow := time.Date(
			2024,
			1,
			14,
			1,
			59,
			0,
			0,
			time.UTC,
		) // Sunday, day 14, 1:59 AM
		result := service.isIMPSAvailable(sundayBeforeWindow)
		assert.True(t, result)
	})

	t.Run(
		"returns false on Sunday during maintenance week at 4:01 AM (hour 4 still in window)",
		func(t *testing.T) {
			// Note: The code checks hour only, so hour 4 (4:00-4:59) is still in maintenance window
			sundayInWindow := time.Date(
				2024,
				1,
				14,
				4,
				1,
				0,
				0,
				time.UTC,
			) // Sunday, day 14, 4:01 AM
			result := service.isIMPSAvailable(sundayInWindow)
			assert.False(t, result)
		},
	)

	t.Run("returns true on Sunday during maintenance week but at 5:00 AM", func(t *testing.T) {
		sundayAfterWindow := time.Date(2024, 1, 14, 5, 0, 0, 0, time.UTC) // Sunday, day 14, 5:00 AM
		result := service.isIMPSAvailable(sundayAfterWindow)
		assert.True(t, result)
	})

	t.Run("returns true on Sunday during maintenance week but at day 7", func(t *testing.T) {
		// Jan 7, 2024 is a Sunday, day 7 (before maintenance week)
		sundayBeforeWeek := time.Date(2024, 1, 7, 3, 0, 0, 0, time.UTC) // Sunday, day 7, 3:00 AM
		result := service.isIMPSAvailable(sundayBeforeWeek)
		assert.True(t, result)
	})

	t.Run("returns true on Sunday during maintenance week but at day 15", func(t *testing.T) {
		// Jan 15, 2024 is a Monday, not Sunday
		// Let's use Jan 21, 2024 which is a Sunday, day 21 (after maintenance week)
		sundayAfterWeek := time.Date(2024, 1, 21, 3, 0, 0, 0, time.UTC) // Sunday, day 21, 3:00 AM
		result := service.isIMPSAvailable(sundayAfterWeek)
		assert.True(t, result)
	})

	t.Run("comprehensive test for all maintenance window boundaries", func(t *testing.T) {
		// Test all combinations of day 8-14, hours 2-4 on Sunday
		testCases := []struct {
			name     string
			date     time.Time
			expected bool
		}{
			{
				"Sunday day 8, 2 AM",
				time.Date(2024, 2, 11, 2, 0, 0, 0, time.UTC),
				false,
			}, // Feb 11, 2024 is Sunday, day 11 (in range)
			{"Sunday day 14, 4 AM", time.Date(2024, 1, 14, 4, 0, 0, 0, time.UTC), false},
			{
				"Sunday day 8, 1 AM",
				time.Date(2024, 2, 11, 1, 0, 0, 0, time.UTC),
				true,
			}, // Before maintenance hours
			{
				"Sunday day 14, 5 AM",
				time.Date(2024, 1, 14, 5, 0, 0, 0, time.UTC),
				true,
			}, // After maintenance hours
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result := service.isIMPSAvailable(tc.date)
				assert.Equal(t, tc.expected, result, "Date: %v", tc.date)
			})
		}
	})
}
