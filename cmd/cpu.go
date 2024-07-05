package cmd

import (
	"context"
	"fmt"
	"log"
	"math"
	"runtime"
	"sync"
	"time"
)

const interval = 10000

func busyWork(ctx context.Context) {
	cnt := 0
	for {
		cnt += 1
		if cnt%interval == 0 {
			cnt = 0
			select {
			case <-ctx.Done():
				log.Printf("busyWork: quit due to context be cancelled")
				return
			default:
			}
		}
	}
}

func partialBusyWork(ctx context.Context, ratio float64) {
	const (
		oneCycle  = 10 * time.Microsecond
		precision = 1000
	)
	// round busy and idle percent
	// case 1: ratio 0.8
	//   busy 0.8                     idle 0.19999999999999996
	//   busyRound 8ms                idleRound 2ms
	//
	// case 2: ratio 0.2
	//   busy 0.16000000000000014     idle 0.8399999999999999
	//   buseRound 1.6ms              idleRound 8.4ms
	busyDuration := time.Duration(math.Floor(ratio*precision)) * oneCycle
	idleDuration := time.Duration(math.Ceil((1-ratio)*precision)) * oneCycle
	cnt := 0
	for {
		// Busy period
		busyStart := time.Now()
		for time.Since(busyStart) < busyDuration {
			cnt += 1 // Simulate work
			if cnt%interval == 0 {
				cnt = 0
				select {
				case <-ctx.Done():
					log.Printf("partialBusyWork: quit due to context being cancelled")
					return
				default:
					//
				}
			}
		}
		// Idle period
		time.Sleep(idleDuration)
	}
}

func eatCPU(ctx context.Context, wg *sync.WaitGroup, c float64) {
	fmt.Printf("Eating %-12s", "CPU...")

	runtime.GOMAXPROCS(runtime.NumCPU())

	fullCores := int(c)
	partialCoreRatio := c - float64(fullCores)

	// eat full cores
	for i := 0; i < fullCores; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			busyWork(ctx)
		}()
	}

	// eat partial core
	if partialCoreRatio > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			partialBusyWork(ctx, partialCoreRatio)
		}()
	}

	fmt.Printf("Ate %2.3f CPU cores\n", c)
}
