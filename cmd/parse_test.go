package cmd

import (
	"context"
	"runtime"
	"testing"
)

func TestParserRejectsUnsafeValues(t *testing.T) {
	for _, value := range []string{"Inf", "-1", "101%", "bad"} {
		if _, err := parserCPUEatCount(value); err == nil {
			t.Errorf("parserCPUEatCount(%q) accepted unsafe value", value)
		}
	}
	for _, value := range []string{"Inf", "-1m", "101%", "20xyz"} {
		if _, err := parserMemory(value); err == nil {
			t.Errorf("parserMemory(%q) accepted unsafe value", value)
		}
	}
}

func TestParserAcceptsSmallValues(t *testing.T) {
	cpu, err := parserCPUEatCount("0.5")
	if err != nil || cpu != 0.5 {
		t.Fatalf("unexpected CPU parse result: %v, %v", cpu, err)
	}
	memory, err := parserMemory("1k")
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
		got, err := parserMemory(tt.input)
		if err != nil || got != tt.want {
			t.Errorf("parserMemory(%q) = %d, %v; want %d", tt.input, got, err, tt.want)
		}
	}
	cpu, err := parserCPUEatCount("50%")
	if err != nil || cpu != float64(runtime.NumCPU())/2 {
		t.Fatalf("parserCPUEatCount(50%%) = %v, %v", cpu, err)
	}
}

func TestEatMemoryRejectsHugeRequestBeforeAllocation(t *testing.T) {
	wait, err := eatMemory(context.Background(), ^uint64(0))
	if err == nil || wait != nil {
		t.Fatalf("huge memory request was not rejected: wait-nil=%t err=%v", wait == nil, err)
	}
}
