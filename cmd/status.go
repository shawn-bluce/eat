package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/log"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/process"
)

const statusInterval = 5 * time.Second

// reportStatus logs the CPU and memory usage of the whole system and of eat
// itself every interval until ctx is done.
func reportStatus(ctx context.Context, interval time.Duration) func() {
	done := make(chan struct{})
	self, err := process.NewProcess(int32(os.Getpid()))
	if err != nil {
		log.Warnf("Cannot inspect own process, status disabled: %v", err)
		close(done)
		return func() { <-done }
	}
	// Prime the "since last call" CPU counters so the first report covers
	// the first interval rather than the time since startup.
	_, _ = cpu.Percent(0, false)
	_, _ = self.Percent(0)

	go func() {
		defer close(done)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				log.Info(statusLine(self))
			}
		}
	}()
	return func() { <-done }
}

func statusLine(self *process.Process) string {
	sysCPU, sysMem, eatCPU, eatMem := "?", "?", "?", "?"
	if percents, err := cpu.Percent(0, false); err == nil && len(percents) > 0 {
		sysCPU = fmt.Sprintf("%.1f%%", percents[0])
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
