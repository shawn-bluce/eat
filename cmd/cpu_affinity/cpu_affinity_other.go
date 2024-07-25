//go:build !linux

package cpu_affinity

import (
	"fmt"
	"runtime"
)

type CpuAffinityDeputy struct{}

func (CpuAffinityDeputy) GetProcessId() uint {
	return 0
}

func (CpuAffinityDeputy) GetThreadId() uint {
	return 0
}

func (CpuAffinityDeputy) SetCpuAffinities(pid uint, cpus ...uint) error {
	return fmt.Errorf("SetCpuAffinities currently not support in this os: %s", runtime.GOOS)
}

func (CpuAffinityDeputy) GetCpuAffinities(pid uint) (map[uint]bool, error) {
	return nil, fmt.Errorf("GetCpuAffinities currently not support in this os: %s", runtime.GOOS)
}

func (CpuAffinityDeputy) IsImplemented() bool {
	return false
}
