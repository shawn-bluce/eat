//go:build !linux

package cmd

func cgroupMemoryUsage(uint64) (used, limit uint64, ok bool) { return 0, 0, false }
