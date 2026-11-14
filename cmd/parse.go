package cmd

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// parseCPU parses -c: a number of cores ("2.5") or a share of all cores ("50%").
func parseCPU(arg string) (float64, error) {
	if arg == "" {
		return 0, nil
	}
	capacity := cpuCapacity()
	if p, ok := strings.CutSuffix(arg, "%"); ok {
		value, err := parseNonNegative(p)
		if err != nil || value > 100 {
			return 0, fmt.Errorf("invalid CPU value %q: percentage must be between 0 and 100", arg)
		}
		return value / 100 * capacity, nil
	}
	value, err := parseNonNegative(arg)
	if err != nil || value > capacity {
		return 0, fmt.Errorf("invalid CPU value %q: must be between 0 and %g cores", arg, capacity)
	}
	return value, nil
}

// parseMemory parses -m: a size with an optional b/k/m/g unit ("1.5g") or a
// share of all memory ("50%").
func parseMemory(arg string) (uint64, error) {
	if arg == "" {
		return 0, nil
	}
	if p, ok := strings.CutSuffix(arg, "%"); ok {
		value, err := parseNonNegative(p)
		if err != nil || value > 100 {
			return 0, fmt.Errorf("invalid memory value %q: percentage must be between 0 and 100", arg)
		}
		_, total, err := systemMemory()
		if err != nil {
			return 0, fmt.Errorf("cannot read memory size: %w", err)
		}
		return uint64(value / 100 * float64(total)), nil
	}
	number, multiplier := strings.ToLower(arg), 1.0
	if unit := strings.IndexByte("bkmg", number[len(number)-1]); unit >= 0 {
		number, multiplier = number[:len(number)-1], math.Pow(1024, float64(unit))
	}
	value, err := parseNonNegative(number)
	if err != nil || value*multiplier >= math.MaxUint64 {
		return 0, fmt.Errorf("invalid memory value %q", arg)
	}
	return uint64(value * multiplier), nil
}

func parseNonNegative(s string) (float64, error) {
	v, err := strconv.ParseFloat(s, 64)
	if err != nil || math.IsNaN(v) || math.IsInf(v, 0) || v < 0 {
		return 0, fmt.Errorf("invalid number %q", s)
	}
	return v, nil
}
