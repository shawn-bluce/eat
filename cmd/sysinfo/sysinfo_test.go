package sysinfo

import (
	"testing"
	"time"
)

func TestGopsutilMemoryMonitor(t *testing.T) {
	m := NewGopsutilMemoryMonitor(time.Second)
	total, err := m.GetTotalMemory()
	if err != nil || total == 0 {
		t.Fatalf("unexpected total memory: %d, %v", total, err)
	}
	free, err := m.GetFreeMemory()
	if err != nil || free > total {
		t.Fatalf("unexpected free memory: %d, %v", free, err)
	}
	rss, err := m.GetCurrentProcessMemory()
	if err != nil || rss == 0 {
		t.Fatalf("unexpected process RSS: %d, %v", rss, err)
	}
}

func TestGopsutilCPUConcurrentAccess(t *testing.T) {
	m := NewGopsutilCpuMonitor(time.Hour)
	results := make(chan float64, 3)
	for i := 0; i < 3; i++ {
		go func() {
			usage, err := m.GetCPUUsage()
			if err == nil {
				results <- usage
			} else {
				results <- -1
			}
		}()
	}
	for i := 0; i < 3; i++ {
		usage := <-results
		if usage < 0 || usage > 100 {
			t.Fatalf("invalid CPU usage: %v", usage)
		}
	}
}
