//go:build linux

package cmd

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// cgroupRoot is where the cgroup of this process is mounted. Containers
// (Docker, Kubernetes) mount their own cgroup here.
var cgroupRoot = "/sys/fs/cgroup"

// cgroupMemoryUsage returns the working set and the limit of the current
// cgroup when it has a memory limit lower than hostTotal. The working set is
// usage minus inactive page cache, the same figure `kubectl top` reports.
func cgroupMemoryUsage(hostTotal uint64) (used, limit uint64, ok bool) {
	for _, v := range []struct{ limit, usage, stat, inactiveKey string }{
		{"memory.max", "memory.current", "memory.stat", "inactive_file"},                                              // v2
		{"memory/memory.limit_in_bytes", "memory/memory.usage_in_bytes", "memory/memory.stat", "total_inactive_file"}, // v1
	} {
		limit, err := readCgroupUint(v.limit)
		if err != nil || limit == 0 || limit >= hostTotal {
			continue
		}
		usage, err := readCgroupUint(v.usage)
		if err != nil {
			continue
		}
		if inactive, ok := readCgroupStat(v.stat, v.inactiveKey); ok && inactive < usage {
			usage -= inactive
		}
		return usage, limit, true
	}
	return 0, 0, false
}

// cgroupCPU returns the CPU limit of the current cgroup in cores and the CPU
// time it has used in seconds, when it has a limit lower than hostCPUs.
func cgroupCPU(hostCPUs float64) (quota, usedSeconds float64, ok bool) {
	// v2: cpu.max is "<quota> <period>" or "max <period>".
	if fields := strings.Fields(readCgroupFile("cpu.max")); len(fields) == 2 {
		q, err1 := strconv.ParseFloat(fields[0], 64)
		p, err2 := strconv.ParseFloat(fields[1], 64)
		usec, ok := readCgroupStat("cpu.stat", "usage_usec")
		if err1 == nil && err2 == nil && ok && q > 0 && p > 0 && q/p < hostCPUs {
			return q / p, float64(usec) / 1e6, true
		}
		return 0, 0, false
	}
	// v1: a quota of -1 means no limit.
	q, err1 := strconv.ParseFloat(readCgroupFile("cpu/cpu.cfs_quota_us"), 64)
	p, err2 := strconv.ParseFloat(readCgroupFile("cpu/cpu.cfs_period_us"), 64)
	nsec, err3 := readCgroupUint("cpuacct/cpuacct.usage")
	if err1 == nil && err2 == nil && err3 == nil && q > 0 && p > 0 && q/p < hostCPUs {
		return q / p, float64(nsec) / 1e9, true
	}
	return 0, 0, false
}

func readCgroupFile(name string) string {
	data, _ := os.ReadFile(filepath.Join(cgroupRoot, name))
	return strings.TrimSpace(string(data))
}

func readCgroupUint(name string) (uint64, error) {
	return strconv.ParseUint(readCgroupFile(name), 10, 64)
}

// readCgroupStat reads the value of key from a "key value" per line file.
func readCgroupStat(name, key string) (uint64, bool) {
	for _, line := range strings.Split(readCgroupFile(name), "\n") {
		if k, v, found := strings.Cut(line, " "); found && k == key {
			n, err := strconv.ParseUint(v, 10, 64)
			return n, err == nil
		}
	}
	return 0, false
}
