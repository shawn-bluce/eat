package cmd

import (
	"context"
	"fmt"
	"math"
	"runtime"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
)

// cpuDutyPeriod is the busy/idle cycle length of each CPU worker.
const cpuDutyPeriod = 100 * time.Millisecond

// cpuCapacity is how many cores eat can use: the CPU limit of its cgroup
// (container) if there is one, otherwise all cores of the machine.
func cpuCapacity() float64 {
	if quota, _, ok := cgroupCPU(float64(runtime.NumCPU())); ok {
		return quota
	}
	return float64(runtime.NumCPU())
}

// cpuMeter measures CPU usage between calls to percent, of the cgroup when it
// has a CPU limit and of the whole machine otherwise.
type cpuMeter struct{ busy, total float64 }

func newCPUMeter() *cpuMeter {
	m := &cpuMeter{}
	_, _ = m.percent()
	return m
}

// percent returns the CPU usage since the previous call as a percentage of
// cpuCapacity.
func (m *cpuMeter) percent() (float64, error) {
	busy, total, err := cpuTimes()
	if err != nil {
		return 0, err
	}
	dBusy, dTotal := busy-m.busy, total-m.total
	m.busy, m.total = busy, total
	if dTotal <= 0 {
		return 0, nil
	}
	return math.Max(0, math.Min(100, dBusy/dTotal*100)), nil
}

var cpuTimesEpoch = time.Now()

// cpuTimes returns the busy and the available CPU seconds so far.
func cpuTimes() (busy, total float64, err error) {
	if quota, used, ok := cgroupCPU(float64(runtime.NumCPU())); ok {
		return used, time.Since(cpuTimesEpoch).Seconds() * quota, nil
	}
	times, err := cpu.Times(false)
	if err != nil || len(times) == 0 {
		return 0, 0, fmt.Errorf("cannot read CPU times: %v", err)
	}
	t := times[0]
	return t.Total() - t.Idle - t.Iowait, t.Total(), nil
}

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
