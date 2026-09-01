//go:build linux || darwin || freebsd

package cmd

import "syscall"

func lockMemory(buf []byte) error   { return syscall.Mlock(buf) }
func unlockMemory(buf []byte) error { return syscall.Munlock(buf) }
