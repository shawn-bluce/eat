package cmd

import (
	"fmt"
	"github.com/shirou/gopsutil/v4/mem"
	"math"
	"runtime"
	"strconv"
	"strings"
)

func parserCPUEatCount(cpuArg string) (float64, error) {
	totalCPUCount := float64(runtime.NumCPU())

	if cpuArg == "" {
		return 0, nil
	}

	if strings.HasSuffix(cpuArg, "%") {
		percentStr := strings.TrimSuffix(cpuArg, "%")
		parsedVal, err := strconv.ParseFloat(percentStr, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid CPU value %q", cpuArg)
		}
		if !finiteNonNegative(parsedVal) || parsedVal > 100 {
			return 0, fmt.Errorf("CPU percentage must be between 0 and 100")
		}
		return (parsedVal / 100.0) * totalCPUCount, nil
	} else {
		parsedVal, err := strconv.ParseFloat(cpuArg, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid CPU value %q", cpuArg)
		}
		if !finiteNonNegative(parsedVal) || parsedVal > totalCPUCount {
			return 0, fmt.Errorf("CPU count must be between 0 and %d", runtime.NumCPU())
		}
		return parsedVal, nil
	}
}

func finiteNonNegative(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) && v >= 0 }

func parserMemory(memArg string) (uint64, error) {
	if memArg == "" {
		return 0, nil
	}
	if strings.HasSuffix(memArg, "%") {
		percentStr := strings.TrimSuffix(memArg, "%")
		percentage, err := strconv.ParseFloat(percentStr, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid memory value %q", memArg)
		}
		vmStat, err := mem.VirtualMemory()
		if err != nil {
			return 0, err
		}
		totalMemory := float64(effectiveAvailableMemory(vmStat.Total))
		if !finiteNonNegative(percentage) || percentage > 100 {
			return 0, fmt.Errorf("memory percentage must be between 0 and 100")
		}
		return uint64((percentage / 100.0) * totalMemory), nil
	} else {
		memArgLower := strings.ToLower(memArg)

		multiplier := float64(1)
		var numericPart string

		lastChar := memArgLower[len(memArgLower)-1]
		switch lastChar {
		case 'b':
			numericPart = memArgLower[:len(memArgLower)-1]
		case 'k':
			multiplier = 1024
			numericPart = memArgLower[:len(memArgLower)-1]
		case 'm':
			multiplier = 1024 * 1024
			numericPart = memArgLower[:len(memArgLower)-1]
		case 'g':
			multiplier = 1024 * 1024 * 1024
			numericPart = memArgLower[:len(memArgLower)-1]
		default:
			numericPart = memArgLower
		}

		value, err := strconv.ParseFloat(numericPart, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid memory value %q", memArg)
		}
		if !finiteNonNegative(value) || value*multiplier >= float64(^uint64(0)) {
			return 0, fmt.Errorf("memory value is out of range")
		}
		return uint64(value * multiplier), nil
	}
}
