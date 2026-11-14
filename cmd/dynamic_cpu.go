package cmd

import (
	"context"
	"math"
	"sync/atomic"
	"time"

	"github.com/charmbracelet/log"
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

// dynamicCPU keeps the CPU usage of the system (or of the container eat runs
// in) at targetPercent (0-100).
func dynamicCPU(ctx context.Context, targetPercent float64) func() {
	capacity := cpuCapacity()
	var level atomic.Uint64 // math.Float64bits of the number of cores to burn
	loadLevel := func() float64 { return math.Float64frombits(level.Load()) }
	waitWorkers := startCPUWorkers(ctx, int(math.Ceil(capacity)), loadLevel)

	done := make(chan struct{})
	go func() {
		defer close(done)
		meter := newCPUMeter()
		ticker := time.NewTicker(dynamicSampleInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
			measured, err := meter.percent()
			if err != nil {
				log.Warnf("Cannot measure CPU usage: %v", err)
				continue
			}
			level.Store(math.Float64bits(nextCPULevel(loadLevel(), targetPercent, measured, capacity)))
		}
	}()
	return func() { waitWorkers(); <-done }
}
