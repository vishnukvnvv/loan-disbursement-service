package service

import (
	"math/rand/v2"
	"payment-gateway/models"
	"time"
)

type AvailabilitySchedule interface {
	IsAvailable(channel models.PaymentChannel, time time.Time) bool
}

type AvailabilityScheduleImpl struct{}

func NewAvailabilitySchedule() AvailabilitySchedule {
	return &AvailabilityScheduleImpl{}
}

func (s AvailabilityScheduleImpl) IsAvailable(channel models.PaymentChannel, time time.Time) bool {
	switch channel {
	case models.PaymentChannelUPI:
		return s.isUPIAvailable()
	case models.PaymentChannelNEFT:
		return s.isNEFTAvailable(time)
	case models.PaymentChannelIMPS:
		return s.isIMPSAvailable(time)
	}
	return false
}

func (s AvailabilityScheduleImpl) isUPIAvailable() bool {
	return rand.Float64() > 0.09
}

func (s AvailabilityScheduleImpl) isNEFTAvailable(at time.Time) bool {
	if at.Weekday() == time.Saturday || at.Weekday() == time.Sunday {
		return false
	}
	return at.Hour() >= 8 && at.Hour() < 19
}

func (s AvailabilityScheduleImpl) isIMPSAvailable(at time.Time) bool {
	if at.Weekday() != time.Sunday {
		return true
	}
	if at.Day() < 8 || at.Day() > 14 {
		return true
	}
	if at.Hour() < 2 || at.Hour() > 4 {
		return true
	}
	return false
}
