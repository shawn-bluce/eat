package cmd

import (
	"context"
	"fmt"
	"log"
	"math"
	"math/rand/v2"
	"runtime"
	"sync"
	"time"

	"eat/cmd/cpu_affinity"
	"eat/cmd/sysinfo"
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

func maintainCpuUsage(ctx context.Context, wg *sync.WaitGroup, coreNum float64, usagePercent float64, cpuAffinitiesEat []uint, cpuMonitor sysinfo.SystemCPUMonitor) {
	if coreNum == 0 {
		coreNum = float64(runtime.NumCPU())
	}
	fmt.Printf("CPU usage will be maintaining at minimum %.3f%%, eating %.3f cores, be patient...\n", usagePercent, coreNum)

	wg.Add(1)
	go func() {
		defer wg.Done()
		MaintainCpuUsage(ctx, coreNum, usagePercent, cpuAffinitiesEat, cpuMonitor)
	}()
}

func MaintainCpuUsage(ctx context.Context, coreNum float64, usagePercent float64, cpuAffinitiesEat []uint, cpuMonitor sysinfo.SystemCPUMonitor) {
	runtime.GOMAXPROCS(runtime.NumCPU())

	fullCores := int(coreNum)
	partialCoreRatio := coreNum - float64(fullCores)

	const maxIdleDuration = 1 * time.Second
	const minIdleDuration = 1 * time.Millisecond
	const initIdleDurationAdjustRatio float64 = 0.1
	const minIdleDurationAdjustRatio = 0.002
	var idleDuration = maxIdleDuration
	var dynIdleDurationAdjustRatio float64 = initIdleDurationAdjustRatio
	var stopWork = false
	var cur float64 = 0

	var fixIdleDuration = func() {
		var err error
		cur, err = cpuMonitor.GetCPUUsage()
		if err != nil {
			log.Printf("MaintainCpuUsage: get cpu usage failed, reason: %s", err.Error())
			return
		}
		// When the cpu usage fluctuates greatly, increase idleDurationAdjustRatio to stabilize the cpu usage
		if dynIdleDurationAdjustRatio == minIdleDurationAdjustRatio {
			if cur > usagePercent+20 || cur < usagePercent-20 {
				dynIdleDurationAdjustRatio = initIdleDurationAdjustRatio
			} else if cur > usagePercent+10 || cur < usagePercent-10 {
				dynIdleDurationAdjustRatio = initIdleDurationAdjustRatio * 0.5
			}
		}
		if cur > usagePercent {
			idleDuration = time.Duration(float64(idleDuration) * (1 + dynIdleDurationAdjustRatio))
		} else if cur < usagePercent {
			idleDuration = time.Duration(float64(idleDuration) * (1 - dynIdleDurationAdjustRatio))
		}
		if idleDuration < minIdleDuration {
			idleDuration = minIdleDuration
		} else if idleDuration > maxIdleDuration {
			idleDuration = maxIdleDuration
			stopWork = true
		} else {
			stopWork = false
		}
		// gradually decrease the idle duration adjustment ratio, make the idle duration more stable
		dynIdleDurationAdjustRatio -= 0.001
		dynIdleDurationAdjustRatio = max(minIdleDurationAdjustRatio, dynIdleDurationAdjustRatio)
	}
	var worker = func(idx int, workerName string, work func()) {
		cleanup, err := setCpuAffWrapper(idx, cpuAffinitiesEat)
		if err != nil {
			fmt.Printf("Error: %s failed to set cpu affinities, reason: %s\n", workerName, err.Error())
			return
		}
		if cleanup != nil {
			fmt.Printf("Worker %s: CPU affinities set to %d\n", workerName, cpuAffinitiesEat[idx])
			defer cleanup()
		}
		for {
			if !stopWork {
				work()
			}
			// if idleDuration is less than 1ms, do not sleep, directly execute fixIdleDuration
			if idleDuration > time.Millisecond*1 {
				time.Sleep(idleDuration)
			}
		}
	}
	cpuIntensiveTask := GenerateCPUIntensiveTask(time.Microsecond * 2000) // 2ms is empirical data
	for i := 0; i < fullCores; i++ {
		workerName := fmt.Sprintf("%d@fullCore", i)
		go worker(i, workerName, cpuIntensiveTask)
	}
	if partialCoreRatio > 0 {
		workerName := fmt.Sprintf("%d@partCore", fullCores)
		go worker(fullCores, workerName, cpuIntensiveTask)
	}
	fmt.Print("\033[?25l")       // hide cursor
	defer fmt.Print("\033[?25h") // show cursor

	ticker := time.NewTicker(time.Millisecond * 300)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			fmt.Println("MaintainCpuUsage: quit due to context being cancelled")
			return
		case <-ticker.C:
			fixIdleDuration()
			printCpuInfo(cur, idleDuration, dynIdleDurationAdjustRatio)
		}
	}
}

func printCpuInfo(usagePercent float64, idleDuration time.Duration, ratio float64) {
	// clear current line
	fmt.Print("\033[2K")
	fmt.Printf("Idle: %s\n", idleDuration)
	fmt.Printf("Ratio: %.3f%%\n", ratio)
	fmt.Printf("CPU usage:\033[32m%6.2f%%\033[0m\n", usagePercent)
	fmt.Print("\033[3A\033[G")
}

// GenerateCPUIntensiveTask returns a function that performs a duration CPU-intensive task
func GenerateCPUIntensiveTask(duration time.Duration) func() {
	const N = 1000
	start := time.Now()
	iteration := 0
	var cnt int64 = 0
	var i int
	for time.Since(start) < duration {
		for i = 0; i < N; i++ {
			cnt = cnt * rand.Int64N(100)
			iteration++
		}
	}
	return func() {
		var cnt2 int64 = cnt
		for i = 0; i < iteration; i++ {
			cnt2 = cnt2 * rand.Int64N(100)
			cnt++
		}
		// KeepAlive ensures that the variable is not optimized away by the compiler
		runtime.KeepAlive(cnt)
		runtime.KeepAlive(cnt2)
	}
}
