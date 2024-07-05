package cmd

import (
	"fmt"
	"runtime"
	"strconv"
	"time"

	"github.com/pbnjay/memory"
)

func parseEatCPUCount(c string) float64 {
	if c == "100%" {
		return float64(runtime.NumCPU())
	} else {
		if len(c) > 1 && (c[len(c)-1] == '%') {
			cEat, err := strconv.ParseFloat(c[:len(c)-1], 32)
			if err != nil {
				fmt.Println("Error: invalid cpu count")
				return 0
			} else {
				return cEat / 100 * float64(runtime.NumCPU())
			}
		}

		cEat, err := strconv.ParseFloat(c, 32)
		if err != nil {
			fmt.Println("Error: invalid cpu count")
			return 0
		} else {
			return cEat
		}
	}
}

func parseEatMemoryBytes(m string) uint64 {
	// allow g/G, m/M, k/K suffixes
	// 1G = 1024M = 1048576K
	if m == "100%" {
		return memory.TotalMemory()
	} else {
		// process k, m, g suffixes
		if len(m) > 1 && (m[len(m)-1] == 'g' || m[len(m)-1] == 'G') {
			mEatBytes, err := strconv.ParseUint(m[:len(m)-1], 10, 64)
			if err == nil {
				return mEatBytes * 1024 * 1024 * 1024
			}
		}
		if len(m) > 1 && (m[len(m)-1] == 'm' || m[len(m)-1] == 'M') {
			mEatBytes, err := strconv.ParseUint(m[:len(m)-1], 10, 64)
			if err == nil {
				return mEatBytes * 1024 * 1024
			}
		}
		if len(m) > 1 && (m[len(m)-1] == 'k' || m[len(m)-1] == 'K') {
			mEatBytes, err := strconv.ParseUint(m[:len(m)-1], 10, 64)
			if err == nil {
				return mEatBytes * 1024
			}
		}

		// process percent
		if len(m) > 1 && m[len(m)-1] == '%' {
			mEatPercent, err := strconv.ParseFloat(m[:len(m)-1], 32)
			if err == nil {
				return uint64(float64(memory.TotalMemory()) * mEatPercent / 100)
			}
		}
	}
	return 0
}

func parseEatDeadline(eta string) time.Duration {
	duration, err := time.ParseDuration(eta)
	if err != nil {
		return time.Duration(0)
	}
	if duration <= 0 {
		return time.Duration(0)
	}
	return duration
}
