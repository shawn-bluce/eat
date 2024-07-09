package cmd

import (
	"fmt"
	"time"
)

// contextKey is a value for use with context.WithValue.
// It's used as a pointer. so it fits in an interface{} without allocation.
type contextKey struct {
	name      string
	valueType string
}

func (k *contextKey) String() string {
	return fmt.Sprintf("worker context value: name %s, type %s", k.name, k.valueType)
}

const (
	intervalCpuWorkerCheckContextDone = 10000
	durationMemoryWorkerDoRefresh     = 5 * time.Minute
	durationEachSignCheck             = 100 * time.Millisecond
	chunkSizeMemoryWorkerEachAllocate = 128 * 1024 * 1024 // 128MB
)

var (
	cpuWorkerPartialCoreRatioContextKey = &contextKey{"partialCoreRatio", "float64"}
)
