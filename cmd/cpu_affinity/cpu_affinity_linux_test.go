//go:build linux
// +build linux

package cpu_affinity

import (
	"math"
	"math/rand"
	"os"
	"runtime"
	"slices"
	"testing"

	"golang.org/x/sys/unix"
)

func TestSchedGetAffinity(t *testing.T) {
	pid := os.Getpid()
	cpuAffDeputy := CpuAffinityDeputy{}
	res, err := cpuAffDeputy.GetCpuAffinities(uint(pid))
	if err != nil {
		t.Errorf("schedGetAffinity failed: %v", err)
		return
	}
	var firstMask bool
	for i := 0; i < runtime.NumCPU(); i++ {
		val, ok := res[uint(i)]
		if !ok {
			t.Errorf("core index %d not found in GetCpuAffinities result", i)
			return
		}
		if i == 0 {
			firstMask = val
			continue
		}

		// it should be all true or all false
		if firstMask != val {
			t.Errorf("cpu %d mask should be %v(not set), but it is %v", i, firstMask, !firstMask)
		}
	}
}

func genRandomCpuCore(num int) []uint {
	numCpu := runtime.NumCPU()
	if num > (numCpu - 1) {
		num = numCpu - 1
	}
	var uniqueCores []uint
	for len(uniqueCores) != num {
		core := uint(math.Floor(rand.Float64() * float64(numCpu)))
		if slices.Contains(uniqueCores, core) {
			continue
		}
		uniqueCores = append(uniqueCores, core)
	}
	return uniqueCores
}

func TestSchedSetAffinity(t *testing.T) {
	pid := os.Getpid()
	mask := new(unix.CPUSet)
	mask.Zero()
	var modCpuCores []uint
	if runtime.NumCPU() > 2 {
		modCpuCores = genRandomCpuCore(2)
	} else {
		modCpuCores = []uint{0}
	}
	cpuAffDeputy := CpuAffinityDeputy{}
	err := cpuAffDeputy.SetCpuAffinities(uint(pid), modCpuCores...)
	if err != nil {
		t.Errorf("SetCpuAffinities failed: %v", err)
		return
	}

	resMap, errGet := cpuAffDeputy.GetCpuAffinities(uint(pid))
	if errGet != nil {
		t.Errorf("schedGetAffinity failed: %v", errGet)
		return
	}

	for i := 0; i < runtime.NumCPU(); i++ {
		expect := slices.Contains(modCpuCores, uint(i))
		val, ok := resMap[uint(i)]
		if !ok {
			t.Errorf("core index %d not found in GetCpuAffinities result", i)
			return
		}
		if expect != val {
			t.Errorf("cpu %d affinities not equal expect: %v", i, expect)
		}
	}
}
