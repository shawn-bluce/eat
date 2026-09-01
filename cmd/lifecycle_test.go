package cmd

import (
	"context"
	"testing"
	"time"
)

func TestEatMemoryStopsOnCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	wait, err := eatMemory(ctx, 4096)
	if err != nil {
		t.Fatalf("eatMemory failed for small allocation: %v", err)
	}
	cancel()
	done := make(chan struct{})
	go func() { wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("memory worker did not stop after cancellation")
	}
}

func TestEatCPUStopsOnCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	wait := eatCPU(ctx, 0.01)
	cancel()
	done := make(chan struct{})
	go func() { wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("CPU worker did not stop after cancellation")
	}
}
