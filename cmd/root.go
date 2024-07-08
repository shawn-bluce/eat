package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/pbnjay/memory"
	"github.com/spf13/cobra"
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

func getConsoleHelpTips(deadline time.Duration) string {
	var helpTips = []string{"Press Ctrl + C to exit"}
	if deadline > 0 {
		eta := time.Now().Add(deadline)
		helpTips = append(helpTips, fmt.Sprintf("or wait it util deadline %s", eta.Format(time.DateTime)))
	}
	helpTips = append(helpTips, "...")
	return strings.Join(helpTips, " ")
}

func gracefulExit(ctx context.Context, ctxCancel context.CancelFunc) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case <-ctx.Done():
			return
		case sig := <-sigs:
			log.Printf("\nReceive exit signal: %v\n", sig)
			ctxCancel()
			return
		default:
			time.Sleep(durationEachSignCheck)
		}
	}
}

func waitUtil(ctx context.Context, wg *sync.WaitGroup, ctxCancel context.CancelFunc, deadline time.Duration) {
	log.Println(getConsoleHelpTips(deadline))
	gracefulExit(ctx, ctxCancel)
	wg.Wait()
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
	r, _ := cmd.Flags().GetString("memory_refresh_interval")

	if c == "0" && m == "0m" {
		fmt.Println("Error: no cpu or memory usage specified")
		return
	}

	cEat := parseEatCPUCount(c)
	mEat := parseEatMemoryBytes(m)
	dlEat := parseTimeDuration(dl)
	mAteRenew := parseTimeDuration(r)

	var wg sync.WaitGroup
	rootCtx, cancel := getRootContext(dlEat)
	defer cancel()
	fmt.Printf("Want to eat %2.3fCPU, %s Memory\n", cEat, m)
	eatMemory(rootCtx, &wg, mEat, mAteRenew)
	eatCPU(rootCtx, &wg, cEat)
	// in case that all sub goroutines are dead due to runtime error like memory not enough.
	// so the main gooutine automaticlly quit as well, don't wait user ctrl+c or context deadline.
	go func(wgp *sync.WaitGroup) {
		wgp.Wait()
		cancel()
	}(&wg)
	waitUtil(rootCtx, &wg, cancel, dlEat)
}
