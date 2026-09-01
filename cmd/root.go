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
	Example: "  eat -c 2 -m 2g\n  eat -c 50% -m 50%\n  eat -c 1.5 -m 10%",
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
	<-ctx.Done()
	log.Infof("receive signal, will exit...")
	waitCPU()
	waitMemory()
	return nil
}
