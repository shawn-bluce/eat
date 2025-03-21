package cmd

import (
	"eat/cmd/version"
	"github.com/charmbracelet/log"
	"github.com/pbnjay/memory"
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
	Run:     eatFunction,
}

func displaySystemInfo() {
	log.Infof("This system has %d CPUs and %d bytes memory (%dC%dG)", runtime.NumCPU(), memory.TotalMemory(), runtime.NumCPU(), memory.TotalMemory()/1024/1024/1024)
}

func waiting4exit() {
	sigChan := make(chan os.Signal, 1)

	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	sig := <-sigChan

	log.Infof("receive signal %v, will exit...", sig)
	os.Exit(0)
}

func eatFunction(cmd *cobra.Command, _ []string) {
	c, _ := cmd.Flags().GetString("cpu")
	m, _ := cmd.Flags().GetString("memory")

	if c == "0" && m == "0m" {
		_ = cmd.Help()
		return
	}

	log.Infof("version: %s, build time: %s, build hash: %s", version.Version, version.BuildTime, version.BuildHash)
	displaySystemInfo()
	eatCpuCount := parserCPUEatCount(c)
	eatMemoryCount := parserMemory(m)
	log.Infof("Will eating %2.3f CPU cores and %d bytes memory", eatCpuCount, eatMemoryCount)
	eatCPU(eatCpuCount)
	eatMemory(eatMemoryCount)
	log.Infof("This monster is eating %2.3f CPU cores and %d bytes memory", eatCpuCount, eatMemoryCount)
	waiting4exit()
}
