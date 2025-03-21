package cmd

import (
	"github.com/charmbracelet/log"
	"math"
	"runtime"
	"time"
)

func eatCPU(c float64) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	log.Info("Starting to eat CPU...")

	wholeCores := int(math.Floor(c))
	fraction := c - float64(wholeCores)

	for i := 0; i < wholeCores; i++ {
		go func(id int) {
			for {
				_ = math.Sqrt(12345.6789)
			}
		}(i + 1)
	}

	if fraction > 0 {
		go func() {
			interval := 100 * time.Millisecond
			busyTime := time.Duration(fraction * float64(interval))
			idleTime := interval - busyTime

			for {
				start := time.Now()
				for time.Since(start) < busyTime {
					_ = math.Sqrt(12345.6789)
				}
				time.Sleep(idleTime)
			}
		}()
	}
	log.Info("Ending to eat CPU...")
}
