//go:build !linux

package cmd

func effectiveAvailableMemory(hostAvailable uint64) uint64 { return hostAvailable }
