package cmd

import (
	"context"
	"math"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/charmbracelet/log"
	"github.com/shirou/gopsutil/v4/cpu"
)

const (
	// dynamicSampleInterval is how often the controllers measure the system.
	dynamicSampleInterval = time.Second
	// cpuDutyPeriod is the busy/idle cycle length of each CPU worker.
	cpuDutyPeriod = 100 * time.Millisecond
	// cpuControlGain is the share of the measured error corrected per sample.
	// Values in (0, 2) converge; smaller values are slower but less jumpy.
	cpuControlGain = 0.5
)

// nextCPULevel returns how many cores eat should burn next, given how many it
// burns now and the target and measured system-wide CPU usage in percent.
// Correcting by the error alone converges without knowing how much CPU eat
// itself uses, so other processes are accounted for automatically.
func nextCPULevel(level, targetPercent, measuredPercent float64, numCPU int) float64 {
	next := level + cpuControlGain*(targetPercent-measuredPercent)/100*float64(numCPU)
	return math.Max(0, math.Min(float64(numCPU), next))
}

// workerDuty is the busy share of worker index when eat burns level cores:
// workers below the level spin fully, the next one spins part of the time.
func workerDuty(level float64, index int) float64 {
	return math.Max(0, math.Min(1, level-float64(index)))
}

// dynamicCPU keeps system-wide CPU usage at targetPercent (0-100).
func dynamicCPU(ctx context.Context, targetPercent float64) func() {
	numCPU := runtime.NumCPU()
	var level atomic.Uint64 // math.Float64bits of the number of cores to burn
	loadLevel := func() float64 { return math.Float64frombits(level.Load()) }

	var wg sync.WaitGroup
	for i := range numCPU {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			cpuDutyWorker(ctx, func() float64 { return workerDuty(loadLevel(), index) })
		}(i)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			percents, err := cpu.PercentWithContext(ctx, dynamicSampleInterval, false)
			if ctx.Err() != nil {
				return
			}
			if err != nil || len(percents) == 0 {
				log.Warnf("Cannot measure CPU usage: %v", err)
				select {
				case <-ctx.Done():
					return
				case <-time.After(dynamicSampleInterval):
				}
				continue
			}
			level.Store(math.Float64bits(nextCPULevel(loadLevel(), targetPercent, percents[0], numCPU)))
		}
	}()
	return wg.Wait
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
