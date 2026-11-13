package cmd

import (
	"context"
	"fmt"
	"github.com/charmbracelet/log"
	"github.com/shirou/gopsutil/v4/mem"
	"sync/atomic"
	"time"
)

const (
	memoryChunkSize  = 32 * 1024 * 1024
	memoryPageStride = 4096
	// memoryKeepAliveInterval is how often pages that could not be locked are
	// rewritten to keep them from being swapped out. Rewriting is linear in
	// the memory held, so doing it too often burns noticeable CPU.
	memoryKeepAliveInterval = 10 * time.Second
)

// memoryChunk is a block of memory held by eat. Every page of it is backed by
// physical memory, and it is locked in RAM when the OS allows it.
type memoryChunk struct {
	buf    []byte
	locked bool
}

func allocateChunk(size uint64) memoryChunk {
	c := memoryChunk{buf: make([]byte, size)}
	touchPages(c.buf)
	if err := lockMemory(c.buf); err != nil {
		warnLockFailure(err)
	} else {
		c.locked = true
	}
	return c
}

// lockWarned makes sure a failing mlock is reported once instead of once
// per chunk, which would print hundreds of lines when eating gigabytes.
var lockWarned atomic.Bool

func warnLockFailure(err error) {
	if lockWarned.CompareAndSwap(false, true) {
		log.Warnf("Cannot lock memory in RAM (%v), so it may be swapped out by the OS. %s", err, memoryLockHint)
	}
}

func (c memoryChunk) free() {
	if c.locked {
		_ = unlockMemory(c.buf)
	}
}

// touchPages writes one byte in every page so the OS has to keep it resident.
func touchPages(buf []byte) {
	for i := 0; i < len(buf); i += memoryPageStride {
		buf[i] ^= 1
	}
}

// keepChunksAlive rewrites the pages of chunks that are not locked. Locked
// chunks are pinned in RAM by the OS and need no help.
func keepChunksAlive(chunks []memoryChunk) {
	for _, c := range chunks {
		if !c.locked {
			touchPages(c.buf)
		}
	}
}

func freeChunks(chunks []memoryChunk) {
	for _, c := range chunks {
		c.free()
	}
}

// systemMemory returns how much memory is in use and the total amount. Inside a
// Linux cgroup with a memory limit it reports the cgroup instead of the host.
func systemMemory() (used, total uint64, err error) {
	vm, err := mem.VirtualMemory()
	if err != nil {
		return 0, 0, err
	}
	if used, limit, ok := cgroupMemoryUsage(vm.Total); ok {
		return used, limit, nil
	}
	return vm.Total - vm.Available, vm.Total, nil
}

func eatMemory(ctx context.Context, memoryBytes uint64) (func(), error) {
	if memoryBytes == 0 {
		return func() {}, nil
	}
	vm, err := mem.VirtualMemory()
	if err != nil {
		return nil, fmt.Errorf("read available memory: %w", err)
	}
	if memoryBytes >= effectiveAvailableMemory(vm.Available) {
		return nil, fmt.Errorf("requested %d bytes, only %d bytes available", memoryBytes, effectiveAvailableMemory(vm.Available))
	}
	log.Info("Starting to eat memory...")

	var chunks []memoryChunk
	for remaining := memoryBytes; remaining > 0; {
		select {
		case <-ctx.Done():
			freeChunks(chunks)
			return func() {}, ctx.Err()
		default:
		}
		current := uint64(memoryChunkSize)
		if remaining < current {
			current = remaining
		}
		vm, err = mem.VirtualMemory()
		if err != nil {
			freeChunks(chunks)
			return func() {}, err
		}
		if current >= effectiveAvailableMemory(vm.Available) {
			freeChunks(chunks)
			return func() {}, fmt.Errorf("allocation stopped: only %d bytes available", vm.Available)
		}
		chunks = append(chunks, allocateChunk(current))
		remaining -= current
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		defer freeChunks(chunks)
		ticker := time.NewTicker(memoryKeepAliveInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				keepChunksAlive(chunks)
			}
		}
	}()

	log.Info("Ending to eat memory...")
	return func() { <-done }, nil
}
