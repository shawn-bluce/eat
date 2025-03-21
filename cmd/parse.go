package cmd

import (
	"github.com/shirou/gopsutil/v4/mem"
	"runtime"
	"strconv"
	"strings"
)

func parserCPUEatCount(cpuArg string) float64 {
	totalCPUCount := float64(runtime.NumCPU())

	if cpuArg == "" {
		return 0
	}

	if strings.HasSuffix(cpuArg, "%") {
		percentStr := strings.TrimSuffix(cpuArg, "%")
		parsedVal, err := strconv.ParseFloat(percentStr, 64)
		if err != nil {
			return 0
		}
		return (parsedVal / 100.0) * totalCPUCount
	} else {
		parsedVal, err := strconv.ParseFloat(cpuArg, 64)
		if err != nil {
			return 0
		}
		return parsedVal
	}
}

func parserMemory(memArg string) uint64 {
	if memArg == "" {
		return 0
	}
	if strings.HasSuffix(memArg, "%") {
		percentStr := strings.TrimSuffix(memArg, "%")
		percentage, err := strconv.ParseFloat(percentStr, 64)
		if err != nil {
			return 0
		}
		vmStat, err := mem.VirtualMemory()
		if err != nil {
			return 0
		}
		totalMemory := float64(vmStat.Total)
		return uint64((percentage / 100.0) * totalMemory)
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
			return 0
		}
		return uint64(value * multiplier)
	}
}
