package cmd

import (
	"context"
	"math"
	"runtime"
	"sync"
	"time"
)

// cpuDutyPeriod is the busy/idle cycle length of each CPU worker.
const cpuDutyPeriod = 100 * time.Millisecond

// cpuCapacity is how many cores eat can use.
func cpuCapacity() float64 { return float64(runtime.NumCPU()) }

// eatCPU burns a fixed number of cores until ctx is done.
func eatCPU(ctx context.Context, cores float64) func() {
	return startCPUWorkers(ctx, int(math.Ceil(cores)), func() float64 { return cores })
}

// startCPUWorkers starts n workers that together burn level() cores. level
// may change at any time; workers pick it up within one cycle.
func startCPUWorkers(ctx context.Context, n int, level func() float64) func() {
	var wg sync.WaitGroup
	for i := range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cpuDutyWorker(ctx, func() float64 { return workerDuty(level(), i) })
		}()
	}
	return wg.Wait
}

// workerDuty is the busy share of worker index when eat burns level cores:
// workers below the level spin fully, the next one spins part of the time.
func workerDuty(level float64, index int) float64 {
	return math.Max(0, math.Min(1, level-float64(index)))
}

// cpuDutyWorker spins for duty() of every cpuDutyPeriod and sleeps the rest.
func cpuDutyWorker(ctx context.Context, duty func() float64) {
	for {
		start := time.Now()
		busy := time.Duration(duty() * float64(cpuDutyPeriod))
		for time.Since(start) < busy {
			for range 1000 {
				_ = math.Sqrt(12345.6789)
			}
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(cpuDutyPeriod - time.Since(start)):
		}
	}
}
