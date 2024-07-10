package cmd

import (
	"github.com/pbnjay/memory"
	"runtime"
	"testing"
	"time"
)

func TestParseEatCPUCount(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"100%", float64(runtime.NumCPU())},
		{"50%", float64(runtime.NumCPU()) * 0.5},
		{"1", 1.0},
		{"2", 2.0},
		{"invalid", 0.0}, // Invalid input should return 0
	}

	for _, tt := range tests {
		result := parseEatCPUCount(tt.input)
		if result != tt.expected {
			t.Errorf("Expected %f for input %s, got %f", tt.expected, tt.input, result)
		}
	}
}

func TestParseEatMemoryBytes(t *testing.T) {
	totalMemory := memory.TotalMemory()

	tests := []struct {
		input  string
		expect uint64
	}{
		{"100%", totalMemory},
		{"50%", totalMemory / 2},
		{"1G", 1 * 1024 * 1024 * 1024},
		{"1M", 1 * 1024 * 1024},
		{"1K", 1 * 1024},
		{"invalid", 0}, // invalid input should return 0
	}

	for _, tt := range tests {
		result := parseEatMemoryBytes(tt.input)
		if result != tt.expect {
			t.Errorf("Expected %d for input %s, got %d", tt.expect, tt.input, result)
		}
	}
}

func TestParesTimeDuration(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Duration
	}{
		{"1s", 1 * time.Second},
		{"1m", 1 * time.Minute},
		{"1h", 1 * time.Hour},
		{"invalid", 0}, // Invalid input should return 0
		{"-1s", 0},     // Negative duration should return 0
	}

	for _, tt := range tests {
		result := parseTimeDuration(tt.input)
		if result != tt.expected {
			t.Errorf("Expected %d for input %s, got %d", tt.expected, tt.input, result)
		}
	}
}
