//go:build linux

package cmd

import (
	"os"
	"strconv"
	"strings"
)

func effectiveAvailableMemory(hostAvailable uint64) uint64 {
	for _, path := range []string{"/sys/fs/cgroup/memory.max", "/sys/fs/cgroup/memory/memory.limit_in_bytes"} {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		value := strings.TrimSpace(string(data))
		if value == "max" {
			continue
		}
		limit, err := strconv.ParseUint(value, 10, 64)
		if err == nil && limit > 0 && limit < hostAvailable {
			return limit
		}
	}
	return hostAvailable
}
