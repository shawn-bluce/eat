package cmd

import (
	"context"
	"fmt"
	"runtime"
	"testing"
)

func TestParserRejectsUnsafeValues(t *testing.T) {
	for _, value := range []string{"Inf", "-1", "101%", "bad"} {
		if _, err := parseCPU(value); err == nil {
			t.Errorf("parseCPU(%q) accepted unsafe value", value)
		}
	}
	if _, err := parseCPU(fmt.Sprintf("%d", runtime.NumCPU()+1)); err == nil {
		t.Fatal("parseCPU accepted more CPUs than the device has")
	}
	for _, value := range []string{"Inf", "-1m", "101%", "20xyz"} {
		if _, err := parseMemory(value); err == nil {
			t.Errorf("parseMemory(%q) accepted unsafe value", value)
		}
	}
}

func TestParserAcceptsSmallValues(t *testing.T) {
	cpu, err := parseCPU("0.5")
	if err != nil || cpu != 0.5 {
		t.Fatalf("unexpected CPU parse result: %v, %v", cpu, err)
	}
	memory, err := parseMemory("1k")
	if err != nil || memory != 1024 {
		t.Fatalf("unexpected memory parse result: %d, %v", memory, err)
	}
}

func TestParserMemoryUnitsAndPercent(t *testing.T) {
	tests := []struct {
		input string
		want  uint64
	}{
		{"1b", 1},
		{"2k", 2 * 1024},
		{"3m", 3 * 1024 * 1024},
		{"1.5g", 3 * 512 * 1024 * 1024},
	}
	for _, tt := range tests {
		got, err := parseMemory(tt.input)
		if err != nil || got != tt.want {
			t.Errorf("parseMemory(%q) = %d, %v; want %d", tt.input, got, err, tt.want)
		}
	}
	cpu, err := parseCPU("50%")
	if err != nil || cpu != float64(runtime.NumCPU())/2 {
		t.Fatalf("parseCPU(50%%) = %v, %v", cpu, err)
	}
}

func TestEatMemoryRejectsHugeRequestBeforeAllocation(t *testing.T) {
	wait, err := eatMemory(context.Background(), ^uint64(0))
	if err == nil || wait != nil {
		t.Fatalf("huge memory request was not rejected: wait-nil=%t err=%v", wait == nil, err)
	}
}
