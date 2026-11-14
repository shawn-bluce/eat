package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/log"
	"github.com/shirou/gopsutil/v4/process"
)

const statusInterval = 5 * time.Second

// reportStatus logs the CPU and memory usage of the whole system and of eat
// itself every interval until ctx is done.
func reportStatus(ctx context.Context, interval time.Duration) func() {
	self, err := process.NewProcess(int32(os.Getpid()))
	if err != nil {
		log.Warnf("Cannot inspect own process, status disabled: %v", err)
		return func() {}
	}
	// Prime the "since last call" counters so the first report covers the
	// first interval rather than the time since startup.
	meter := newCPUMeter()
	_, _ = self.Percent(0)

	done := make(chan struct{})
	go func() {
		defer close(done)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				log.Info(statusLine(meter, self))
			}
		}
	}()
	return func() { <-done }
}

func statusLine(meter *cpuMeter, self *process.Process) string {
	sysCPU, sysMem, eatCPU, eatMem := "?", "?", "?", "?"
	if percent, err := meter.percent(); err == nil {
		sysCPU = fmt.Sprintf("%.1f%%", percent)
	}
	if used, total, err := systemMemory(); err == nil && total > 0 {
		sysMem = fmt.Sprintf("%.1f%% (%s / %s)", float64(used)/float64(total)*100, formatBytes(used), formatBytes(total))
	}
	if percent, err := self.Percent(0); err == nil {
		eatCPU = fmt.Sprintf("%.2f cores", percent/100)
	}
	if info, err := self.MemoryInfo(); err == nil {
		eatMem = formatBytes(info.RSS)
	}
	return fmt.Sprintf("system: cpu %s, memory %s | eat: cpu %s, memory %s", sysCPU, sysMem, eatCPU, eatMem)
}

func formatBytes(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%dB", b)
	}
	value, suffix := float64(b), "KMGTPE"
	i := -1
	for value >= unit && i < len(suffix)-1 {
		value /= unit
		i++
	}
	return fmt.Sprintf("%.2f%cB", value, suffix[i])
}
