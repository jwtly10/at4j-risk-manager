package utils

import "time"

// TimeProvider is an interface that provides the current time
// This is abstracted for testing purposes
// as some of the logic is time critical
type TimeProvider interface {
	Now() time.Time
}

type RealTimeProvider struct{}

func (RealTimeProvider) Now() time.Time {
	return time.Now()
}