package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/pbnjay/memory"
	"github.com/spf13/cobra"
)

const SleepDurationEachIteration = 100 * time.Millisecond

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

func getConsoleHelpTips(deadline time.Duration) string {
	var helpTips = []string{"Press Ctrl + C to exit"}
	if deadline > 0 {
		eta := time.Now().Add(deadline)
		helpTips = append(helpTips, fmt.Sprintf("or wait it util deadline %s", eta.Format(time.DateTime)))
	}
	helpTips = append(helpTips, "...")
	return strings.Join(helpTips, " ")
}

func waitUtil(ctx context.Context, ctxCancel context.CancelFunc, deadline time.Duration) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	log.Println(getConsoleHelpTips(deadline))

	for {
		select {
		case <-ctx.Done():
			return
		case sig := <-sigs:
			log.Printf("\nReceive exit signal: %v\n", sig)
			ctxCancel()
			return
		default:
			time.Sleep(SleepDurationEachIteration)
		}
	}
}

func getRootContext(dlEat time.Duration) (context.Context, context.CancelFunc) {
	var (
		cancel  context.CancelFunc
		rootCtx context.Context
	)
	if dlEat > 0 {
		deadline := time.Now().Add(dlEat)
		rootCtx, cancel = context.WithDeadline(context.Background(), deadline)
	} else {
		rootCtx = context.Background()
		cancel = func() {}
	}
	return rootCtx, cancel
}

func eatFunction(cmd *cobra.Command, _ []string) {
	cpuCount, memoryBytes := getCPUAndMemory()
	fmt.Printf("Have %dC%dG.\n", cpuCount, memoryBytes/1024/1024/1024)

	// Get the flags
	c, _ := cmd.Flags().GetString("cpu_usage")
	m, _ := cmd.Flags().GetString("memory_usage")
	dl, _ := cmd.Flags().GetString("time_deadline")

	if c == "0" && m == "0m" {
		fmt.Println("Error: no cpu or memory usage specified")
		return
	}

	cEat := parseEatCPUCount(c)
	mEat := parseEatMemoryBytes(m)
	dlEat := parseEatDeadline(dl)

	rootCtx, cancel := getRootContext(dlEat)
	fmt.Printf("Want to eat %2.3fCPU, %s Memory\n", cEat, m)
	eatMemory(mEat)
	eatCPU(rootCtx, cEat)
	waitUtil(rootCtx, cancel, dlEat)
}
