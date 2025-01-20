package sysinfo

import (
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
)

type SystemResourceMonitor interface {
	SystemMemoryMonitor
	SystemCPUMonitor
}

type SystemMemoryMonitor interface {
	GetTotalMemory() (uint64, error)
	GetFreeMemory() (uint64, error)
	GetCurrentProcessMemory() (uint64, error)
}

type SystemCPUMonitor interface {
	// GetCPUUsage returns the CPU usage as a percentage (0-100).
	// To ensure that frequent calls do not consume a lot of CPU, consider using caching in the implementation.
	GetCPUUsage() (float64, error)
}

var Monitor SystemResourceMonitor

func init() {
	Monitor = NewGopsutilMonitor(300 * time.Millisecond)
}

type GopsutilMonitor struct {
	SystemCPUMonitor
	SystemMemoryMonitor
}

func NewGopsutilMonitor(refreshInterval time.Duration) *GopsutilMonitor {
	return &GopsutilMonitor{
		SystemCPUMonitor:    NewGopsutilCpuMonitor(refreshInterval),
		SystemMemoryMonitor: NewGopsutilMemoryMonitor(refreshInterval),
	}
}

type GopsutilCpuMonitor struct {
	cpuUsage          float64
	latestRefreshTime time.Time
	refreshInterval   time.Duration
}

func NewGopsutilCpuMonitor(refreshInterval time.Duration) *GopsutilCpuMonitor {
	return &GopsutilCpuMonitor{
		refreshInterval: refreshInterval,
	}
}

func (m *GopsutilCpuMonitor) Refresh() error {
	cpuUsage, err := cpu.Percent(time.Millisecond*1000, false)
	if err != nil {
		return err
	}
	if runtime.GOOS == "windows" {
		// The gopsutil library returns a CPU usage rate that is 10% lower than the actual usage rate on Windows.
		// see: https://github.com/shirou/gopsutil/issues/1744
		m.cpuUsage += 8
	}
	m.cpuUsage = cpuUsage[0]
	if m.cpuUsage > 100 {
		m.cpuUsage = 100
	}
	if m.cpuUsage < 0 {
		m.cpuUsage = 0
	}

	// fmt.Println("PercentWithContext", cpuUsage)
	m.latestRefreshTime = time.Now()
	return nil
}

func (m *GopsutilCpuMonitor) refresh() error {
	if time.Since(m.latestRefreshTime) < m.refreshInterval {
		return nil
	}
	return m.Refresh()
}

func (m *GopsutilCpuMonitor) GetCPUUsage() (float64, error) {
	err := m.refresh()
	if err != nil {
		return 0, err
	}
	return m.cpuUsage, nil
}

// TODO: implement SystemMemoryMonitor
type GopsutilMemoryMonitor struct {
	// totalMemory          uint64
	// freeMemory           uint64
	// currentProcessMemory uint64
	// latestRefreshTime    time.Time
	refreshInterval time.Duration
}

func NewGopsutilMemoryMonitor(refreshInterval time.Duration) *GopsutilMemoryMonitor {
	return &GopsutilMemoryMonitor{
		refreshInterval: refreshInterval,
	}
}

func (m *GopsutilMemoryMonitor) GetTotalMemory() (uint64, error) {
	panic("not implemented")
}

func (m *GopsutilMemoryMonitor) GetFreeMemory() (uint64, error) {
	panic("not implemented")
}

func (m *GopsutilMemoryMonitor) GetCurrentProcessMemory() (uint64, error) {
	panic("not implemented")
}
