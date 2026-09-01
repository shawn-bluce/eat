package cmd

import (
	"context"
	"github.com/charmbracelet/log"
	"math"
	"runtime"
	"sync"
	"time"
)

func eatCPU(ctx context.Context, c float64) func() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	log.Info("Starting to eat CPU...")

	wholeCores := int(math.Floor(c))
	fraction := c - float64(wholeCores)

	var workers sync.WaitGroup
	for i := 0; i < wholeCores; i++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for {
				for i := 0; i < intervalCpuWorkerCheckContextDone; i++ {
					_ = math.Sqrt(12345.6789)
				}
				select {
				case <-ctx.Done():
					return
				default:
				}
			}
		}()
	}

	if fraction > 0 {
		workers.Add(1)
		go func() {
			defer workers.Done()
			interval := 100 * time.Millisecond
			busyTime := time.Duration(fraction * float64(interval))
			idleTime := interval - busyTime

			for {
				start := time.Now()
				for time.Since(start) < busyTime {
					_ = math.Sqrt(12345.6789)
					select {
					case <-ctx.Done():
						return
					default:
					}
				}
				select {
				case <-ctx.Done():
					return
				case <-time.After(idleTime):
				}
			}
		}()
	}
	log.Info("Ending to eat CPU...")
	return workers.Wait
}
