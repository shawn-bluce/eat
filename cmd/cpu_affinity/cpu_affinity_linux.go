//go:build linux
// +build linux

package cpu_affinity

import (
	"runtime"
	"syscall"

	"golang.org/x/sys/unix"
)

type CpuAffinityDeputy struct{}

func (CpuAffinityDeputy) GetProcessId() uint {
	return uint(syscall.Getpid())
}

func (CpuAffinityDeputy) GetThreadId() uint {
	return uint(syscall.Gettid())
}

func (CpuAffinityDeputy) SetCpuAffinities(pid uint, cpus ...uint) error {
	if len(cpus) == 0 {
		return nil
	}
	mask := new(unix.CPUSet)
	mask.Zero()
	for _, c := range cpus {
		mask.Set(int(c))
	}
	return unix.SchedSetaffinity(int(pid), mask)
}

func (CpuAffinityDeputy) GetCpuAffinities(pid uint) (map[uint]bool, error) {
	mask := new(unix.CPUSet)
	mask.Zero()
	err := unix.SchedGetaffinity(int(pid), mask)
	if err != nil {
		return nil, err
	}
	var res = make(map[uint]bool)
	for i := 0; i < runtime.NumCPU(); i++ {
		res[uint(i)] = mask.IsSet(i)
	}
	return res, nil
}

func (CpuAffinityDeputy) IsImplemented() bool {
	return true
}
