//go:build linux

package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

// fakeCgroup points cgroupRoot at a temporary directory holding files.
func fakeCgroup(t *testing.T, files map[string]string) {
	t.Helper()
	dir := t.TempDir()
	for name, content := range files {
		path := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	saved := cgroupRoot
	t.Cleanup(func() { cgroupRoot = saved })
	cgroupRoot = dir
}

func TestCgroupMemoryUsage(t *testing.T) {
	const host = 16 << 30
	tests := []struct {
		name        string
		files       map[string]string
		used, limit uint64
		ok          bool
	}{
		{"v2 working set excludes inactive cache", map[string]string{
			"memory.max": "1073741824\n", "memory.current": "600000000\n",
			"memory.stat": "anon 400000000\ninactive_file 100000000\nactive_file 5\n",
		}, 500000000, 1073741824, true},
		{"v1", map[string]string{
			"memory/memory.limit_in_bytes": "1073741824\n", "memory/memory.usage_in_bytes": "300000000\n",
			"memory/memory.stat": "total_inactive_file 100000000\n",
		}, 200000000, 1073741824, true},
		{"v2 without limit", map[string]string{"memory.max": "max\n", "memory.current": "1\n"}, 0, 0, false},
		{"v1 without limit", map[string]string{"memory/memory.limit_in_bytes": "9223372036854771712\n", "memory/memory.usage_in_bytes": "1\n"}, 0, 0, false},
		{"no cgroup", nil, 0, 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeCgroup(t, tt.files)
			used, limit, ok := cgroupMemoryUsage(host)
			if used != tt.used || limit != tt.limit || ok != tt.ok {
				t.Fatalf("got %d, %d, %v; want %d, %d, %v", used, limit, ok, tt.used, tt.limit, tt.ok)
			}
		})
	}
}

func TestCgroupCPU(t *testing.T) {
	tests := []struct {
		name        string
		files       map[string]string
		quota, used float64
		ok          bool
	}{
		{"v2 limit", map[string]string{"cpu.max": "150000 100000\n", "cpu.stat": "usage_usec 2500000\nuser_usec 1\n"}, 1.5, 2.5, true},
		{"v2 without limit", map[string]string{"cpu.max": "max 100000\n", "cpu.stat": "usage_usec 1\n"}, 0, 0, false},
		{"v2 limit above host", map[string]string{"cpu.max": "1600000 100000\n", "cpu.stat": "usage_usec 1\n"}, 0, 0, false},
		{"v1 limit", map[string]string{"cpu/cpu.cfs_quota_us": "200000\n", "cpu/cpu.cfs_period_us": "100000\n", "cpuacct/cpuacct.usage": "3000000000\n"}, 2, 3, true},
		{"v1 without limit", map[string]string{"cpu/cpu.cfs_quota_us": "-1\n", "cpu/cpu.cfs_period_us": "100000\n", "cpuacct/cpuacct.usage": "1\n"}, 0, 0, false},
		{"no cgroup", nil, 0, 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeCgroup(t, tt.files)
			quota, used, ok := cgroupCPU(8)
			if quota != tt.quota || used != tt.used || ok != tt.ok {
				t.Fatalf("got %v, %v, %v; want %v, %v, %v", quota, used, ok, tt.quota, tt.used, tt.ok)
			}
		})
	}
}
