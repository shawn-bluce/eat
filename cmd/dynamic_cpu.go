package cmd

import (
	"context"
	"math"
	"sync/atomic"
	"time"

	"github.com/charmbracelet/log"
	"github.com/shirou/gopsutil/v4/cpu"
)

const (
	// dynamicSampleInterval is how often the controllers measure the system.
	dynamicSampleInterval = time.Second
	// cpuControlGain is the share of the measured error corrected per sample.
	// Values in (0, 2) converge; smaller values are slower but less jumpy.
	cpuControlGain = 0.5
)

// nextCPULevel returns how many cores eat should burn next, given how many it
// burns now and the target and measured system-wide CPU usage in percent.
// Correcting by the error alone converges without knowing how much CPU eat
// itself uses, so other processes are accounted for automatically.
func nextCPULevel(level, targetPercent, measuredPercent, capacity float64) float64 {
	next := level + cpuControlGain*(targetPercent-measuredPercent)/100*capacity
	return math.Max(0, math.Min(capacity, next))
}

// dynamicCPU keeps system-wide CPU usage at targetPercent (0-100).
func dynamicCPU(ctx context.Context, targetPercent float64) func() {
	capacity := cpuCapacity()
	var level atomic.Uint64 // math.Float64bits of the number of cores to burn
	loadLevel := func() float64 { return math.Float64frombits(level.Load()) }
	waitWorkers := startCPUWorkers(ctx, int(math.Ceil(capacity)), loadLevel)

	done := make(chan struct{})
	go func() {
		defer close(done)
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
			level.Store(math.Float64bits(nextCPULevel(loadLevel(), targetPercent, percents[0], capacity)))
		}
	}()
	return func() { waitWorkers(); <-done }
}
