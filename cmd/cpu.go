package cmd

import (
	"context"
	"fmt"
	"log"
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
	busyDuration := time.Duration(ratio*10) * time.Millisecond
	idleDuration := time.Duration((1-ratio)*10) * time.Millisecond
	cnt := 0
	for {
		// Busy period
		busyStart := time.Now()
		for time.Since(busyStart) < busyDuration {
			cnt += 1 // Simulate work
		}
		// Idle period
		time.Sleep(idleDuration)

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
}

func eatCPU(ctx context.Context, c float64) {
	fmt.Printf("Eating %-12s", "CPU...")

	runtime.GOMAXPROCS(runtime.NumCPU())

	var wg sync.WaitGroup
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

	fmt.Printf("Ate %2.1f CPU cores\n", c)
}
