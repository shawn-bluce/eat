//go:build linux

package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCgroupMemoryUsage(t *testing.T) {
	dir := t.TempDir()
	write := func(name, content string) string {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		return path
	}
	files := cgroupFiles{
		limit:           write("memory.max", "1073741824\n"),
		usage:           write("memory.current", "600000000\n"),
		stat:            write("memory.stat", "anon 400000000\ninactive_file 100000000\nactive_file 5\n"),
		inactiveFileKey: "inactive_file",
	}
	saved := cgroupMemoryFiles
	t.Cleanup(func() { cgroupMemoryFiles = saved })
	cgroupMemoryFiles = []cgroupFiles{files}

	used, limit, ok := cgroupMemoryUsage(16 << 30)
	if !ok || used != 500000000 || limit != 1073741824 {
		t.Fatalf("got used=%d limit=%d ok=%v, want 500000000 1073741824 true", used, limit, ok)
	}

	// A limit at or above the host total ("max" or the v1 sentinel) means
	// there is no effective cgroup limit.
	write("memory.max", "max\n")
	if _, _, ok := cgroupMemoryUsage(16 << 30); ok {
		t.Fatal("expected no cgroup limit for \"max\"")
	}
	write("memory.max", "9223372036854771712\n")
	if _, _, ok := cgroupMemoryUsage(16 << 30); ok {
		t.Fatal("expected no cgroup limit when the limit exceeds host memory")
	}
}
