package cmd

import "fmt"

func eatMemory(memoryBytes uint64) {
	if memoryBytes == 0 {
		return
	}

	memoryBlock := make([]byte, memoryBytes)
	fmt.Printf("Eating memory...          ")
	for i := range memoryBlock {
		memoryBlock[i] = byte(i % 256)
	}
	fmt.Printf("Ate %d bytes memory\n", memoryBytes)
}
