package cmd

import (
	"context"
	"errors"
	"fmt"
	unimem "github.com/pbnjay/memory"
	"log"
	"sync"
	"time"
)

// ActiveMemoryManager struct to encapsulate buffers and related methods
type ActiveMemoryManager struct {
	buffers [][]byte
	size    uint64
	mu      sync.Mutex
}

// NewActiveMemoryManager creates a new ActiveMemoryManager
func NewActiveMemoryManager(size uint64) *ActiveMemoryManager {
	return &ActiveMemoryManager{size: size}
}

// AllocateMemory initializes the memory buffers
func (m *ActiveMemoryManager) AllocateMemory(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.buffers != nil {
		return errors.New("allocated, please free it before retry")
	}
	curFreeSize := unimem.FreeMemory()
	if m.size > curFreeSize {
		return fmt.Errorf("free memory not enough: %d > %d", m.size, curFreeSize)
	}
	//split request memory to multiple small-size chunks
	const unitChunk = chunkSizeMemoryWorkerEachAllocate
	nChunks := m.size / unitChunk
	remain := m.size % unitChunk
	bufSizes := []uint64{}
	for i := uint64(0); i < nChunks; i++ {
		bufSizes = append(bufSizes, unitChunk)
	}
	if remain > 0 {
		bufSizes = append(bufSizes, remain)
	}

	fmt.Printf("Eating %-12sStart\n", "memory...")
	// Because free memory not equal to contiguous free memory, `make([]byte, m.size)` may fail
	// So we change the direct allocation to "divide and conquer", each time we only allocate small chunk of memory.
	m.buffers = make([][]byte, len(bufSizes))
	for i := range m.buffers {
		// This also has the added benefit of checking that context is canceled before each small memory allocation,
		// which gives you a chance to gracefully shut down eat memory goroutine instead of brutally killing it
		// if you want to `eat` dozens of gigabytes of memory and the OS has swap partition turned on .
		select {
		case <-ctx.Done():
			m.buffers = nil
			return errors.New("cancel allocation")
		default:
			//
		}
		curSize := bufSizes[i]
		curFreeSize = unimem.FreeMemory()
		if curSize > curFreeSize {
			m.buffers = nil
			return fmt.Errorf("free memory not enough: %d > %d", curSize, curFreeSize)
		}

		buffer := make([]byte, curSize)
		for i := range buffer {
			buffer[i] = byte(i % 256)
		}
		m.buffers[i] = buffer
	}

	fmt.Printf("Ate %d bytes memory\n", m.size)
	return nil
}

// RefreshMemory touches the memory to keep it active
func (m *ActiveMemoryManager) RefreshMemory() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.buffers == nil {
		return
	}
	for i := range m.buffers {
		for j := range m.buffers[i] {
			// XOR with 0 keeps the value unchanged but touches the memory
			m.buffers[i][j] ^= 0
		}
	}
}

// FreeMemory let GC release allocated memory
func (m *ActiveMemoryManager) FreeMemory() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.buffers == nil {
		return
	}
	for i := range m.buffers {
		m.buffers[i] = nil
	}
	m.buffers = nil
}

func eatMemWork(ctx context.Context, size uint64, refreshInterval time.Duration) {
	m := NewActiveMemoryManager(size)
	err := m.AllocateMemory(ctx)
	if err != nil {
		log.Printf("failed to allocate memory due to %s\n", err.Error())
		return
	}

	ticker := time.NewTicker(refreshInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			m.RefreshMemory()
			log.Println("eatMemWork: Memory refreshed to keep it active")
		case <-ctx.Done():
			log.Println("eatMemWork: Quit due to context being cancelled")
			m.FreeMemory()
			log.Println("eatMemWork: Memory freed")
			return
		default:
			time.Sleep(durationEachSignCheck)
		}
	}
}

func eatMemory(
	ctx context.Context, wg *sync.WaitGroup,
	memoryBytes uint64, refreshInterval time.Duration,
) {
	if memoryBytes == 0 {
		return
	}
	if refreshInterval <= 0 {
		refreshInterval = durationMemoryWorkerDoRefresh
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		eatMemWork(ctx, memoryBytes, refreshInterval)
	}()
}
