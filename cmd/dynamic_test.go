package cmd

import (
	"context"
	"math"
	"testing"
	"time"
)

func TestNextCPULevel(t *testing.T) {
	tests := []struct {
		name                    string
		level, target, measured float64
		numCPU                  int
		want                    float64
	}{
		{"below target adds load", 0, 50, 10, 8, 1.6},
		{"above target removes load", 4, 50, 70, 8, 3.2},
		{"at target keeps load", 2.5, 60, 60, 8, 2.5},
		{"never below zero", 0.5, 30, 90, 8, 0},
		{"never above all cores", 3.9, 100, 0, 4, 4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := nextCPULevel(tt.level, tt.target, tt.measured, tt.numCPU)
			if math.Abs(got-tt.want) > 1e-9 {
				t.Errorf("nextCPULevel(%v, %v, %v, %d) = %v, want %v", tt.level, tt.target, tt.measured, tt.numCPU, got, tt.want)
			}
		})
	}
}

func TestNextCPULevelConverges(t *testing.T) {
	// Other processes use 1.5 of 4 cores; eat should settle where the whole
	// machine is at 75%, i.e. eat itself uses 1.5 cores.
	const others, numCPU = 1.5, 4
	level := 0.0
	for range 30 {
		measured := (others + level) / numCPU * 100
		level = nextCPULevel(level, 75, measured, numCPU)
	}
	if math.Abs(level-1.5) > 0.01 {
		t.Fatalf("level = %v, want about 1.5", level)
	}
}

func TestWorkerDuty(t *testing.T) {
	tests := []struct {
		level float64
		index int
		want  float64
	}{
		{2.5, 0, 1}, {2.5, 1, 1}, {2.5, 2, 0.5}, {2.5, 3, 0}, {0, 0, 0},
	}
	for _, tt := range tests {
		if got := workerDuty(tt.level, tt.index); got != tt.want {
			t.Errorf("workerDuty(%v, %d) = %v, want %v", tt.level, tt.index, got, tt.want)
		}
	}
}

func TestMemoryChunkDelta(t *testing.T) {
	const chunk = 100
	tests := []struct {
		name         string
		target, used uint64
		held, want   int
	}{
		{"far below target allocates whole chunks", 1000, 650, 0, 3},
		{"within one chunk below target does nothing", 1000, 950, 2, 0},
		{"exactly at target does nothing", 1000, 1000, 2, 0},
		{"above target releases enough to drop below", 1000, 1150, 5, -2},
		{"slightly above target releases one chunk", 1000, 1010, 5, -1},
		{"cannot release more than held", 1000, 1500, 2, -2},
		{"nothing held, nothing to release", 1000, 1500, 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := memoryChunkDelta(tt.target, tt.used, tt.held, chunk); got != tt.want {
				t.Errorf("memoryChunkDelta(%d, %d, %d) = %d, want %d", tt.target, tt.used, tt.held, got, tt.want)
			}
		})
	}
}

func TestMemoryTarget(t *testing.T) {
	const gib = 1 << 30
	tests := []struct {
		name  string
		ratio float64
		total uint64
		want  uint64
	}{
		{"plain ratio", 0.5, 16 * gib, 8 * gib},
		{"keeps a 2% reserve", 1, 16 * gib, 16*gib - 16*gib/50},
		{"reserve is at least 64MB", 1, 1 * gib, 1*gib - 64<<20},
		{"tiny total leaves nothing", 1, 32 << 20, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := memoryTarget(tt.ratio, tt.total); got != tt.want {
				t.Errorf("memoryTarget(%v, %d) = %d, want %d", tt.ratio, tt.total, got, tt.want)
			}
		})
	}
}

func TestDynamicCPUStopsOnCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	wait := dynamicCPU(ctx, 1)
	time.Sleep(50 * time.Millisecond)
	cancel()
	waitWithTimeout(t, wait, "dynamic CPU")
}

func TestDynamicMemoryStopsOnCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	wait := dynamicMemory(ctx, 0.001)
	time.Sleep(50 * time.Millisecond)
	cancel()
	waitWithTimeout(t, wait, "dynamic memory")
}

func waitWithTimeout(t *testing.T, wait func(), name string) {
	t.Helper()
	done := make(chan struct{})
	go func() { wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatalf("%s worker did not stop after cancellation", name)
	}
}

func TestFormatBytes(t *testing.T) {
	tests := map[uint64]string{
		512:           "512B",
		1536:          "1.50KB",
		32 << 20:      "32.00MB",
		3 << 30:       "3.00GB",
		5<<40 + 1<<39: "5.50TB",
	}
	for in, want := range tests {
		if got := formatBytes(in); got != want {
			t.Errorf("formatBytes(%d) = %q, want %q", in, got, want)
		}
	}
}
