package cmd

import (
	"context"
	"eat/cmd/version"
	"fmt"
	"github.com/charmbracelet/log"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"runtime"
	"syscall"
)

var RootCmd = &cobra.Command{
	Use:     "eat",
	Short:   "A monster that eats cpu and memory 🦕",
	Version: version.Version,
	Example: "  eat -c 2 -m 2g\n  eat -c 50% -m 50%\n  eat -c 1.5 -m 10%\n  eat -c 70% -m 60% --dynamic",
	RunE:    eatFunction,
}

func displaySystemInfo() {
	vm, err := mem.VirtualMemory()
	if err != nil {
		fmt.Printf("Unable to read total memory: %v\n", err)
		return
	}
	fmt.Printf("\nThis system has %d logical CPUs and %.2fG physical memory\n\n", runtime.NumCPU(), float64(vm.Total)/(1024*1024*1024))
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

	log.Infof("version: %s, build time: %s, build hash: %s", version.Version, version.BuildTime, version.BuildHash)
	displaySystemInfo()
	eatCpuCount, err := parserCPUEatCount(c)
	if err != nil {
		return err
	}
	eatMemoryCount, err := parserMemory(m)
	if err != nil {
		return err
	}
	if dynamic {
		return runDynamic(eatCpuCount, eatMemoryCount)
	}
	vm, err := mem.VirtualMemory()
	if err != nil {
		return fmt.Errorf("cannot read device memory specification: %w", err)
	}
	if eatMemoryCount > vm.Total {
		return fmt.Errorf("requested memory %d bytes exceeds device total memory %d bytes", eatMemoryCount, vm.Total)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Infof("Will eating %2.3f CPU cores and %d bytes memory", eatCpuCount, eatMemoryCount)
	waitMemory, err := eatMemory(ctx, eatMemoryCount)
	if err != nil {
		return fmt.Errorf("cannot allocate memory: %w", err)
	}
	waitCPU := eatCPU(ctx, eatCpuCount)
	log.Infof("This monster is eating %2.3f CPU cores and %d bytes memory", eatCpuCount, eatMemoryCount)
	waitStatus := reportStatus(ctx, statusInterval)
	<-ctx.Done()
	log.Infof("receive signal, will exit...")
	waitCPU()
	waitMemory()
	waitStatus()
	return nil
}

// runDynamic keeps the system-wide CPU and memory usage at the requested
// levels. cpuCores and memoryBytes are converted into shares of the machine,
// so both "-c 70%" and "-c 4" (on an 8-core machine, 50%) work.
func runDynamic(cpuCores float64, memoryBytes uint64) error {
	cpuTarget := cpuCores / float64(runtime.NumCPU()) * 100
	_, totalMemory, err := systemMemory()
	if err != nil {
		return fmt.Errorf("cannot read device memory specification: %w", err)
	}
	memoryRatio := float64(memoryBytes) / float64(totalMemory)
	if memoryRatio > 1 {
		return fmt.Errorf("requested memory %d bytes exceeds device total memory %d bytes", memoryBytes, totalMemory)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	waitCPU, waitMemory := func() {}, func() {}
	if cpuTarget > 0 {
		log.Infof("Dynamic mode: keeping system CPU usage at %.1f%%", cpuTarget)
		waitCPU = dynamicCPU(ctx, cpuTarget)
	}
	if memoryRatio > 0 {
		log.Infof("Dynamic mode: keeping system memory usage at %.1f%%", memoryRatio*100)
		waitMemory = dynamicMemory(ctx, memoryRatio)
	}
	waitStatus := reportStatus(ctx, statusInterval)
	<-ctx.Done()
	log.Infof("receive signal, will exit...")
	waitCPU()
	waitMemory()
	waitStatus()
	return nil
}
