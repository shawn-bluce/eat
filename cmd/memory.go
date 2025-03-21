package cmd

import (
	"github.com/charmbracelet/log"
	"syscall"
	"time"
)

func eatMemory(memoryBytes uint64) {
	log.Info("Starting to eat memory...")

	buf := make([]byte, memoryBytes)

	const pageSize = 4096
	for i := 0; i < len(buf); i += pageSize {
		buf[i] = 1
	}

	err := syscall.Mlock(buf)
	if err != nil {
		log.Warn("Failed to lock memory, the memory may be swapped out by OS.")
	} else {
		log.Info("Successfully locked memory, the memory can't be swapped out by OS.")
	}

	go func(b []byte) {
		for {
			for i := 0; i < len(b); i += pageSize {
				b[i] ^= 1
			}
			time.Sleep(50 * time.Millisecond)
		}
	}(buf)

	log.Info("Ending to eat memory...")
}
