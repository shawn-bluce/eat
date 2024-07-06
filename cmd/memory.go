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

// ActiveMemoryManager struct to encapsulate buffer and related methods
type ActiveMemoryManager struct {
	buffer  []byte
	size    uint64
	mu      sync.Mutex
}

// NewActiveMemoryManager creates a new ActiveMemoryManager
func NewActiveMemoryManager(size uint64) *ActiveMemoryManager {
	return &ActiveMemoryManager{size: size}
}

// AllocateMemory initializes the memory buffer
func (m *ActiveMemoryManager) AllocateMemory(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.buffer != nil {
		return errors.New("allocated, please free it before retry")
	}
	curFreeSize := unimem.FreeMemory()
	if m.size > curFreeSize {
		return fmt.Errorf("free memory not enough: %d > %d", m.size, curFreeSize)
	}
	fmt.Printf("Eating %-12s", "memory...")
	m.buffer = make([]byte, m.size)
	for i := range m.buffer {
		m.buffer[i] = byte(i % 256)
	}
	fmt.Printf("Ate %d bytes memory\n", m.size)
	return nil
}

// RefreshMemory touches the memory to keep it active
func (m *ActiveMemoryManager) RefreshMemory() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.buffer == nil {
		return
	}
	for i := range m.buffer {
		// XOR with 0 keeps the value unchanged but touches the memory
		m.buffer[i] ^= 0
	}
}

// FreeMemory let GC release allocated memory
func (m *ActiveMemoryManager) FreeMemory() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.buffer == nil {
		return
	}
	m.buffer = nil
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
			log.Println("eatMemWork: quit due to context being cancelled")
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
