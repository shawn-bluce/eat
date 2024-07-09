//go:build linux
// +build linux

package cpu_affinity

import (
	"math/bits"
	"runtime"
	"syscall"
	"unsafe"
)

const (
	cpuSetSize = 0x400
	nCpuBits   = 0x40
	cpuSetLen  = cpuSetSize / nCpuBits
)

type cpuMaskT uint64

// cpuSet use array to represents a CPU affinity mask.
type cpuSet [cpuSetLen]cpuMaskT

const (
	enoAGAIN = syscall.Errno(0xb)
	enoINVAL = syscall.Errno(0x16)
	enoNOENT = syscall.Errno(0x2)
)

// Do the interface allocations only once for common
// Errno values.
var (
	errEAGAIN error = syscall.EAGAIN
	errEINVAL error = syscall.EINVAL
	errENOENT error = syscall.ENOENT
)

// errnoErr returns common boxed Errno values, to prevent allocations at runtime.
func errnoErr(e syscall.Errno) error {
	switch e {
	case 0:
		return nil
	case enoAGAIN:
		return errEAGAIN
	case enoINVAL:
		return errEINVAL
	case enoNOENT:
		return errENOENT
	}
	return e
}

func schedAffinity(trap uintptr, pid uint, set *cpuSet) error {
	_, _, e := syscall.RawSyscall(trap, uintptr(pid), unsafe.Sizeof(*set), uintptr(unsafe.Pointer(set)))
	if e != 0 {
		return errnoErr(e)
	}
	return nil
}

// schedGetAffinity gets the CPU affinity mask of the thread specified by pid.
// If pid is 0 the calling thread is used.
func schedGetAffinity(pid uint, set *cpuSet) error {
	return schedAffinity(syscall.SYS_SCHED_GETAFFINITY, pid, set)
}

// schedSetAffinity sets the CPU affinity mask of the thread specified by pid.
// If pid is 0 the calling thread is used.
func schedSetAffinity(pid uint, set *cpuSet) error {
	return schedAffinity(syscall.SYS_SCHED_SETAFFINITY, pid, set)
}

// Zero clears the set s, so that it contains no CPUs.
func (s *cpuSet) Zero() {
	for i := range s {
		s[i] = 0
	}
}

func cpuBitsIndex(cpu uint) uint {
	return cpu / nCpuBits
}

func cpuBitsMask(cpu uint) cpuMaskT {
	return cpuMaskT(1 << (uint(cpu) % nCpuBits))
}

// Set adds cpu to the set s.
func (s *cpuSet) Set(cpu uint) {
	i := cpuBitsIndex(cpu)
	if int(i) < len(s) {
		s[i] |= cpuBitsMask(cpu)
	}
}

// Clear removes cpu from the set s.
func (s *cpuSet) Clear(cpu uint) {
	i := cpuBitsIndex(cpu)
	if int(i) < len(s) {
		s[i] &^= cpuBitsMask(cpu)
	}
}

// IsSet reports whether cpu is in the set s.
func (s *cpuSet) IsSet(cpu uint) bool {
	i := cpuBitsIndex(cpu)
	if int(i) < len(s) {
		return s[i]&cpuBitsMask(cpu) != 0
	}
	return false
}

// Count returns the number of CPUs in the set s.
func (s *cpuSet) Count() uint {
	var c uint = 0
	for _, b := range s {
		c += uint(bits.OnesCount64(uint64(b)))
	}
	return c
}

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
	mask := new(cpuSet)
	mask.Zero()
	for _, c := range cpus {
		mask.Set(c)
	}
	return schedSetAffinity(pid, mask)
}

func (CpuAffinityDeputy) GetCpuAffinities(pid uint) (map[uint]bool, error) {
	mask := new(cpuSet)
	mask.Zero()
	err := schedGetAffinity(pid, mask)
	if err != nil {
		return nil, err
	}
	var res = make(map[uint]bool)
	for i := 0; i < runtime.NumCPU(); i++ {
		res[uint(i)] = mask.IsSet(uint(i))
	}
	return res, nil
}

func (CpuAffinityDeputy) IsImplemented() bool {
	return true
}
