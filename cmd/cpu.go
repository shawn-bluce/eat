package cmd

import (
	"context"
	"fmt"
	"log"
	"math"
	"runtime"
	"sync"
	"time"

	"eat/cmd/cpu_affinity"
)

func busyWork(ctx context.Context) {
	cnt := 0
	for {
		cnt += 1
		if cnt%intervalCpuWorkerCheckContextDone == 0 {
			cnt = 0
			select {
			case <-ctx.Done():
				log.Printf("busyWork: quit due to context be cancelled")
				return
			default:
			}
		}
	}
}

func partialBusyWork(ctx context.Context) {
	const (
		oneCycle  = 10 * time.Microsecond
		precision = 1000
	)
	ratio, ok := ctx.Value(cpuWorkerPartialCoreRatioContextKey).(float64)
	if !ok {
		log.Printf("partialBusyWork: partial core ratio context key not set or type ")
		return
	}
	// round busy and idle percent
	// case 1: ratio 0.8
	//   busy 0.8                     idle 0.19999999999999996
	//   busyRound 8ms                idleRound 2ms
	//
	// case 2: ratio 0.16
	//   busy 0.16000000000000014     idle 0.8399999999999999
	//   buseRound 1.6ms              idleRound 8.4ms
	busyDuration := time.Duration(math.Floor(ratio*precision)) * oneCycle
	idleDuration := time.Duration(math.Ceil((1-ratio)*precision)) * oneCycle
	cnt := 0
	for {
		// Busy period
		busyStart := time.Now()
		for time.Since(busyStart) < busyDuration {
			cnt += 1 // Simulate work
			if cnt%intervalCpuWorkerCheckContextDone == 0 {
				cnt = 0
				select {
				case <-ctx.Done():
					log.Printf("partialBusyWork: quit due to context being cancelled")
					return
				default:
					//
				}
			}
		}
		// Idle period
		time.Sleep(idleDuration)
	}
}

func startEatCpuWorker(ctx context.Context, wg *sync.WaitGroup, workerName string, idx int, workerFunc func(ctx context.Context), cpuAffinitiesEat []uint) {
	defer wg.Done()
	cleanup, err := setCpuAffWrapper(idx, cpuAffinitiesEat)
	if err != nil {
		fmt.Printf("Error: %s failed to set cpu affinities, reason: %s\n", workerName, err.Error())
		return
	}
	if cleanup != nil {
		fmt.Printf("Worker %s: CPU affinities set to %d\n", workerName, cpuAffinitiesEat[idx])
		defer cleanup()
	}
	workerFunc(ctx)
}

func setCpuAffWrapper(index int, cpuAffinitiesEat []uint) (func(), error) {
	if len(cpuAffinitiesEat) == 0 { // user not set cpu affinities, skip...
		return nil, nil
	}
	if len(cpuAffinitiesEat) <= index { // index error
		return nil, fmt.Errorf("cpuAffinities: index out of range")
	}
	// LockOSThread wires the calling goroutine to its current operating system thread.
	// The calling goroutine will **always execute** in that thread, and no other goroutine will execute in it,
	// until the calling goroutine has made as many calls to [UnlockOSThread] as to LockOSThread.
	// If the calling goroutine exits without unlocking the thread, the thread will be terminated.
	//
	// All init functions are run on the startup thread. Calling LockOSThread
	// from an init function will cause the main function to be invoked on
	// that thread.
	//
	// A goroutine should **call LockOSThread before** calling OS services or non-Go library functions
	// that depend on per-thread state.
	runtime.LockOSThread() // IMPORTANT!! Only limit the system thread affinity, not the whole go program process
	var cpuAffDeputy = cpu_affinity.NewCpuAffinityDeputy()
	if !cpuAffDeputy.IsImplemented() {
		return nil, fmt.Errorf("SetCpuAffinities currently not support in this os: %s", runtime.GOOS)
	}
	tid := cpuAffDeputy.GetThreadId()
	err := cpuAffDeputy.SetCpuAffinities(uint(tid), cpuAffinitiesEat[index])
	if err != nil {
		return nil, err
	}
	return func() {
		runtime.UnlockOSThread()
	}, nil
}

func eatCPU(ctx context.Context, wg *sync.WaitGroup, c float64, cpuAffinitiesEat []uint) {
	fmt.Printf("Eating %-12s", "CPU...")
	runtime.GOMAXPROCS(runtime.NumCPU())

	fullCores := int(c)
	partialCoreRatio := c - float64(fullCores)

	// eat full cores
	for i := 0; i < fullCores; i++ {
		wg.Add(1)
		workerName := fmt.Sprintf("%d@fullCore", i)
		go startEatCpuWorker(ctx, wg, workerName, i, busyWork, cpuAffinitiesEat)
	}

	// eat partial core
	if partialCoreRatio > 0 {
		i := fullCores // the last core affinity
		wg.Add(1)
		workerName := fmt.Sprintf("%d@partCore", i)
		childCtx := context.WithValue(ctx, cpuWorkerPartialCoreRatioContextKey, partialCoreRatio)
		go startEatCpuWorker(childCtx, wg, workerName, i, partialBusyWork, cpuAffinitiesEat)
	}

	fmt.Printf("Ate %2.3f CPU cores\n", c)
}
