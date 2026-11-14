package cmd

import "errors"

// Locking memory on Windows needs the process working set to be enlarged
// first; eat keeps it simple and relies on keepChunksAlive instead.

func lockMemory([]byte) error   { return errors.New("not supported on Windows") }
func unlockMemory([]byte) error { return nil }
