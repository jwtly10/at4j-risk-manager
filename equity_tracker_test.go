package main

import (
	"testing"
	"time"
)

func TestIsUpdateTime(t *testing.T) {
	tests := []struct {
		name         string
		currentTime  time.Time
		targetHour   int
		targetMinute int
		expected     bool
	}{
		{
			name:         "Exactly at update time",
			currentTime:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			targetHour:   0,
			targetMinute: 0,
			expected:     true,
		},
		{
			name:         "One minute before update time",
			currentTime:  time.Date(2024, 1, 1, 23, 59, 0, 0, time.UTC),
			targetHour:   0,
			targetMinute: 0,
			expected:     false,
		},
		{
			name:         "One minute after update window",
			currentTime:  time.Date(2024, 1, 1, 0, 1, 0, 0, time.UTC),
			targetHour:   0,
			targetMinute: 0,
			expected:     false,
		},
		{
			name:         "Custom update time - 15:30",
			currentTime:  time.Date(2024, 1, 1, 15, 30, 0, 0, time.UTC),
			targetHour:   15,
			targetMinute: 30,
			expected:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isUpdateTime(tt.currentTime, tt.targetHour, tt.targetMinute)
			if result != tt.expected {
				t.Errorf("isUpdateTime() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TODO: Test actually updating the equity
// We may implement some logic in future to only track based on if
// the equity has actually changed, so this is something to think about in future...
