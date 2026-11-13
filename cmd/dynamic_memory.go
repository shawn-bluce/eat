package cmd

import (
	"context"
	"runtime/debug"
	"time"

	"github.com/charmbracelet/log"
)

// memoryReserveMin is the least memory dynamic mode always leaves free, so a
// high target cannot push the machine into the OOM killer.
const memoryReserveMin = 64 << 20

// memoryTarget converts a target ratio (0-1) into bytes, always leaving
// max(2% of total, memoryReserveMin) free.
func memoryTarget(ratio float64, total uint64) uint64 {
	reserve := max(total/50, memoryReserveMin)
	if total <= reserve {
		return 0
	}
	return min(uint64(ratio*float64(total)), total-reserve)
}

// memoryChunkDelta returns how many chunks to allocate (positive) or release
// (negative). It only allocates whole chunks that fit under the target and
// releases enough to drop back under it, so usage settles within one chunk
// below the target instead of oscillating around it.
func memoryChunkDelta(target, used uint64, held int, chunk uint64) int {
	if used <= target {
		return int((target - used) / chunk)
	}
	release := int((used - target + chunk - 1) / chunk)
	return -min(release, held)
}

// dynamicMemory keeps the memory usage of the system at targetRatio (0-1).
func dynamicMemory(ctx context.Context, targetRatio float64) func() {
	done := make(chan struct{})
	go func() {
		defer close(done)
		var chunks []memoryChunk
		defer func() { freeChunks(chunks) }()

		adjust := func() {
			used, total, err := systemMemory()
			if err != nil {
				log.Warnf("Cannot measure memory usage: %v", err)
				return
			}
			delta := memoryChunkDelta(memoryTarget(targetRatio, total), used, len(chunks), memoryChunkSize)
			for ; delta > 0 && ctx.Err() == nil; delta-- {
				chunks = append(chunks, allocateChunk(memoryChunkSize))
			}
			if delta < 0 {
				keep := len(chunks) + delta
				freeChunks(chunks[keep:])
				clear(chunks[keep:]) // drop references so the GC can reclaim them
				chunks = chunks[:keep]
				debug.FreeOSMemory()
			}
		}

		adjust()
		adjustTicker := time.NewTicker(dynamicSampleInterval)
		defer adjustTicker.Stop()
		keepAliveTicker := time.NewTicker(memoryKeepAliveInterval)
		defer keepAliveTicker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-adjustTicker.C:
				adjust()
			case <-keepAliveTicker.C:
				keepChunksAlive(chunks)
			}
		}
	}()
	return func() { <-done }
}
