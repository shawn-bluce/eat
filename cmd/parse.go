package cmd

import (
	"fmt"
	"github.com/pbnjay/memory"
	"runtime"
	"strconv"
)

func parseEatCPUCount(c string) float64 {
	if c == "100%" {
		return float64(runtime.NumCPU())
	} else {
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
	}
	return 0
}
