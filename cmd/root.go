package cmd

import (
	"fmt"
	"github.com/pbnjay/memory"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"runtime"
	"syscall"
)

func getCPUAndMemory() (uint64, uint64) {
	cpuCount := uint64(runtime.NumCPU())
	memoryBytes := memory.TotalMemory()
	return cpuCount, memoryBytes
}

var RootCmd = &cobra.Command{
	Use:   "eat",
	Short: "A monster that eats cpu and memory 🦕",
	Run:   eatFunction,
}

func waitForever() {
	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	done := make(chan bool, 1)

	go func() {
		sig := <-sigs
		fmt.Println()
		fmt.Println(sig)
		done <- true
	}()

	fmt.Println("Press Ctrl + C to exit...")
	<-done
}

func eatFunction(cmd *cobra.Command, args []string) {
	cpuCount, memoryBytes := getCPUAndMemory()
	fmt.Printf("Have %dC%dG.\n", cpuCount, memoryBytes/1024/1024/1024)

	// Get the flags
	c, _ := cmd.Flags().GetString("cpu_usage")
	m, _ := cmd.Flags().GetString("memory_usage")

	if c == "0" && m == "0m" {
		fmt.Println("Error: no cpu or memory usage specified")
		return
	}

	cEat := parseEatCPUCount(c)
	mEat := parseEatMemoryBytes(m)

	fmt.Printf("Want to eat %2.1fCPU, %s Memory\n", cEat, m)
	eatMemory(mEat)
	eatCPU(cEat)
	waitForever()
}
