package cmd

import (
	"context"
	"fmt"
	"log"
	"math"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"eat/cmd/sysinfo"
	"eat/cmd/version"

	"github.com/pbnjay/memory"
	"github.com/spf13/cobra"
)

func getCPUAndMemory() (uint64, uint64) {
	cpuCount := uint64(runtime.NumCPU())
	memoryBytes := memory.TotalMemory()
	return cpuCount, memoryBytes
}

var RootCmd = &cobra.Command{
	Use:     "eat",
	Short:   "A monster that eats cpu and memory 🦕",
	Version: version.Version,
	Run:     eatFunction,
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
		rootCtx, cancel = context.WithCancel(context.Background())
	}
	return rootCtx, cancel
}

func eatFunction(cmd *cobra.Command, _ []string) {
	fmt.Printf("version: %s, build time: %s, build hash: %s\n", version.Version, version.BuildTime, version.BuildHash)
	cpuCount, memoryBytes := getCPUAndMemory()
	fmt.Printf("Have %dC%dG.\n", cpuCount, memoryBytes/1024/1024/1024)

	// Get the flags
	c, _ := cmd.Flags().GetString("cpu-usage")
	cMaintain, _ := cmd.Flags().GetString("cpu-maintain")
	cAff, _ := cmd.Flags().GetIntSlice("cpu-affinities")
	m, _ := cmd.Flags().GetString("memory-usage")
	dl, _ := cmd.Flags().GetString("time-deadline")
	r, _ := cmd.Flags().GetString("memory-refresh-interval")

	if c == "0" && m == "0m" && cMaintain == "" {
		fmt.Println("Error: no cpu or memory usage specified")
		return
	}
	if c == "0" && cMaintain != "" {
		c = cMaintain
	}

	cEat := parseEatCPUCount(c)
	cMaintainPercent := parseCPUMaintainPercent(cMaintain)
	phyCores := runtime.NumCPU()
	if int(math.Ceil(cEat)) > phyCores {
		fmt.Printf("Error: user specified cpu cores exceed system physical cores(%d)\n", phyCores)
		return
	}
	mEat := parseEatMemoryBytes(m)
	dlEat := parseTimeDuration(dl)
	mAteRenew := parseTimeDuration(r)
	cpuAffinitiesEat, err := parseCpuAffinity(cAff, cEat)
	if err != nil {
		fmt.Printf("Error: failed to parse cpu affinities, reason: %s\n", err.Error())
		return
	}

	var wg sync.WaitGroup
	rootCtx, cancel := getRootContext(dlEat)
	defer cancel()
	fmt.Printf("Want to eat %2.3fCPU, %s Memory\n", cEat, m)
	eatMemory(rootCtx, &wg, mEat, mAteRenew)
	if cMaintainPercent > 0 {
		maintainCpuUsage(rootCtx, &wg, cEat, cMaintainPercent, cpuAffinitiesEat, sysinfo.Monitor)
	} else {
		eatCPU(rootCtx, &wg, cEat, cpuAffinitiesEat)
	}
	// in case that all sub goroutines are dead due to runtime error like memory not enough.
	// so the main goroutine automatically quit as well, don't wait user ctrl+c or context deadline.
	go func(wgp *sync.WaitGroup) {
		wgp.Wait()
		cancel()
	}(&wg)
	waitUtil(rootCtx, &wg, cancel, dlEat)
}
