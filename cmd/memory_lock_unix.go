//go:build linux || darwin

package cmd

import (
	"fmt"
	"syscall"
)

func lockMemory(buf []byte) error {
	if err := syscall.Mlock(buf); err != nil {
		return fmt.Errorf("%w; run eat as root or raise the limit with `ulimit -l` to lock it", err)
	}
	return nil
}

func unlockMemory(buf []byte) error { return syscall.Munlock(buf) }
