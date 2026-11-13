//go:build !linux

package cmd

func effectiveAvailableMemory(hostAvailable uint64) uint64 { return hostAvailable }

func cgroupMemoryUsage(uint64) (used, limit uint64, ok bool) { return 0, 0, false }
