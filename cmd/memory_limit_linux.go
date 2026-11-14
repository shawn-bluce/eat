//go:build linux

package cmd

import (
	"os"
	"strconv"
	"strings"
)

// cgroupFiles describes where a cgroup version exposes its memory accounting.
type cgroupFiles struct {
	limit, usage, stat, inactiveFileKey string
}

var cgroupMemoryFiles = []cgroupFiles{
	{"/sys/fs/cgroup/memory.max", "/sys/fs/cgroup/memory.current", "/sys/fs/cgroup/memory.stat", "inactive_file"},
	{"/sys/fs/cgroup/memory/memory.limit_in_bytes", "/sys/fs/cgroup/memory/memory.usage_in_bytes", "/sys/fs/cgroup/memory/memory.stat", "total_inactive_file"},
}

// cgroupMemoryUsage returns the working set and the limit of the current
// cgroup when it has a memory limit lower than hostTotal. The working set is
// usage minus inactive page cache, the same figure `kubectl top` reports.
func cgroupMemoryUsage(hostTotal uint64) (used, limit uint64, ok bool) {
	for _, f := range cgroupMemoryFiles {
		limit, err := readUintFile(f.limit)
		if err != nil || limit == 0 || limit >= hostTotal {
			continue
		}
		usage, err := readUintFile(f.usage)
		if err != nil {
			continue
		}
		if inactive, ok := readStatValue(f.stat, f.inactiveFileKey); ok && inactive < usage {
			usage -= inactive
		}
		return usage, limit, true
	}
	return 0, 0, false
}

func readUintFile(path string) (uint64, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	return strconv.ParseUint(strings.TrimSpace(string(data)), 10, 64)
}

func readStatValue(path, key string) (uint64, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, false
	}
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 2 && fields[0] == key {
			v, err := strconv.ParseUint(fields[1], 10, 64)
			return v, err == nil
		}
	}
	return 0, false
}
