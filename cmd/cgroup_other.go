//go:build !linux

package cmd

// cgroups only exist on Linux.

func cgroupMemoryUsage(uint64) (used, limit uint64, ok bool) { return 0, 0, false }

func cgroupCPU(float64) (quota, usedSeconds float64, ok bool) { return 0, 0, false }
