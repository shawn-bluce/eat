package cmd

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

func busyWork() {
	for {
		_ = 1 + 1
	}
}

func partialBusyWork(ratio float64) {
	busyDuration := time.Duration(ratio*10) * time.Millisecond
	idleDuration := time.Duration((1-ratio)*10) * time.Millisecond

	for {
		start := time.Now()
		for time.Since(start) < busyDuration {
			_ = 1 + 1
		}
		time.Sleep(idleDuration)
	}
}

func eatCPU(c float64) {
	fmt.Printf("Eating CPU...          ")

	runtime.GOMAXPROCS(runtime.NumCPU())

	var wg sync.WaitGroup
	fullCores := int(c)
	partialCoreRatio := c - float64(fullCores)

	// eat full cores
	for i := 0; i < fullCores; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			busyWork()
		}()
	}

	// eat partial core
	if partialCoreRatio > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			partialBusyWork(partialCoreRatio)
		}()
	}

	fmt.Printf("Ate %2.1f CPU cores\n", c)
}
