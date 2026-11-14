package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"eat/cmd/version"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:     "eat",
	Short:   "A monster that eats cpu and memory 🦕",
	Version: version.Version,
	Example: "  eat -c 2 -m 2g\n  eat -c 50% -m 50%\n  eat -c 1.5 -m 10%\n  eat -c 70% -m 60% --dynamic",
	RunE:    eatFunction,
	// Print errors once, without the usage text, from main.
	SilenceErrors: true,
	SilenceUsage:  true,
}

func init() {
	flags := RootCmd.Flags()
	flags.StringP("cpu", "c", "0", "CPU to eat: cores (2.5) or a share of all cores (50%)")
	flags.StringP("memory", "m", "0m", "memory to eat: size with b/k/m/g unit (1.5g) or a share of all memory (50%)")
	flags.BoolP("dynamic", "d", false, "keep the whole system's CPU/memory usage at -c/-m instead of eating a fixed amount")
}

func displaySystemInfo() {
	_, total, err := systemMemory()
	if err != nil {
		fmt.Printf("Unable to read total memory: %v\n", err)
		return
	}
	fmt.Printf("\nThis system has %g CPUs and %s memory\n\n", cpuCapacity(), formatBytes(total))
}

func eatFunction(cmd *cobra.Command, _ []string) error {
	c, err := cmd.Flags().GetString("cpu")
	if err != nil {
		return err
	}
	m, err := cmd.Flags().GetString("memory")
	if err != nil {
		return err
	}
	dynamic, err := cmd.Flags().GetBool("dynamic")
	if err != nil {
		return err
	}

	if c == "0" && m == "0m" {
		fmt.Println(cmd.Short)
		displaySystemInfo()
		fmt.Print(cmd.UsageString())
		return nil
	}

	cpuCores, err := parseCPU(c)
	if err != nil {
		return err
	}
	memoryBytes, err := parseMemory(m)
	if err != nil {
		return err
	}
	log.Infof("version: %s, build time: %s, build hash: %s", version.Version, version.BuildTime, version.BuildHash)
	displaySystemInfo()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	start := startFixed
	if dynamic {
		start = startDynamic
	}
	waits, err := start(ctx, cpuCores, memoryBytes)
	if err != nil {
		return err
	}
	waits = append(waits, reportStatus(ctx, statusInterval))
	<-ctx.Done()
	log.Info("Received signal, releasing CPU and memory...")
	for _, wait := range waits {
		wait()
	}
	return nil
}

// startFixed eats exactly cpuCores cores and memoryBytes of memory.
func startFixed(ctx context.Context, cpuCores float64, memoryBytes uint64) ([]func(), error) {
	log.Infof("Eating %g CPU cores and %s memory...", cpuCores, formatBytes(memoryBytes))
	waitMemory, err := eatMemory(ctx, memoryBytes)
	if err != nil {
		return nil, fmt.Errorf("cannot allocate memory: %w", err)
	}
	return []func(){waitMemory, eatCPU(ctx, cpuCores)}, nil
}

// startDynamic keeps the CPU and memory usage of the whole system at the
// levels given. cpuCores and memoryBytes are converted into shares of the
// machine, so both "-c 70%" and "-c 4" (on an 8-core machine, 50%) work.
func startDynamic(ctx context.Context, cpuCores float64, memoryBytes uint64) ([]func(), error) {
	_, totalMemory, err := systemMemory()
	if err != nil {
		return nil, fmt.Errorf("cannot read memory size: %w", err)
	}
	cpuPercent := cpuCores / cpuCapacity() * 100
	memoryRatio := float64(memoryBytes) / float64(totalMemory)
	if memoryRatio > 1 {
		return nil, fmt.Errorf("requested %s memory, but there is only %s", formatBytes(memoryBytes), formatBytes(totalMemory))
	}

	var waits []func()
	if cpuPercent > 0 {
		log.Infof("Dynamic mode: keeping system CPU usage at %.1f%%", cpuPercent)
		waits = append(waits, dynamicCPU(ctx, cpuPercent))
	}
	if memoryRatio > 0 {
		log.Infof("Dynamic mode: keeping system memory usage at %.1f%%", memoryRatio*100)
		waits = append(waits, dynamicMemory(ctx, memoryRatio))
	}
	return waits, nil
}
