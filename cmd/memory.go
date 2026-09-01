package cmd

import (
	"context"
	"fmt"
	"github.com/charmbracelet/log"
	"github.com/shirou/gopsutil/v4/mem"
	"time"
)

func eatMemory(ctx context.Context, memoryBytes uint64) (wait func(), err error) {
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

	const chunkSize = 32 * 1024 * 1024
	var buffers [][]byte
	defer func() {
		if r := recover(); r != nil {
			buffers = nil
			wait = nil
			err = fmt.Errorf("allocation failed: %v", r)
		}
	}()
	for remaining := memoryBytes; remaining > 0; {
		select {
		case <-ctx.Done():
			return func() {}, ctx.Err()
		default:
		}
		current := uint64(chunkSize)
		if remaining < current {
			current = remaining
		}
		vm, err = mem.VirtualMemory()
		if err != nil {
			return func() {}, err
		}
		if current >= effectiveAvailableMemory(vm.Available) {
			return func() {}, fmt.Errorf("allocation stopped: only %d bytes available", vm.Available)
		}
		buf := make([]byte, current)
		for i := 0; i < len(buf); i += 4096 {
			buf[i] = 1
		}
		if err := lockMemory(buf); err != nil {
			log.Warn("Failed to lock memory, the memory may be swapped out by OS.")
		}
		buffers = append(buffers, buf)
		remaining -= current
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			select {
			case <-ctx.Done():
				for _, b := range buffers {
					_ = unlockMemory(b)
				}
				return
			default:
			}
			for _, b := range buffers {
				for i := 0; i < len(b); i += 4096 {
					b[i] ^= 1
				}
			}
			select {
			case <-ctx.Done():
			case <-time.After(50 * time.Millisecond):
			}
		}
	}()

	log.Info("Ending to eat memory...")
	return func() { <-done }, nil
}
