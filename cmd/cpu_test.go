package cmd

import (
	"testing"
	"time"
)

func TestGenerateCPUIntensiveTask(t *testing.T) {

	tests := []struct {
		name     string
		duration time.Duration
	}{
		{"2ms", time.Millisecond * 2},
		{"5ms", time.Millisecond * 5},
		{"20ms", time.Millisecond * 20},
		{"100ms", time.Millisecond * 100},
		{"500ms", time.Millisecond * 500},
		{"1s", time.Second},
		{"2s", time.Second * 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateCPUIntensiveTask(tt.duration)
			now := time.Now()
			got()
			after := time.Now()
			if (after.Sub(now).Abs() - tt.duration) > max(tt.duration*8/10, time.Millisecond*2) {
				t.Errorf("GenerateCPUIntensiveTask() = %v, want %v", after.Sub(now), max(tt.duration*8/10, time.Millisecond*2))
			}
		})
	}
}
