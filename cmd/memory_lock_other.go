//go:build !linux && !darwin && !freebsd

package cmd

func lockMemory([]byte) error   { return errMemoryLockUnsupported{} }
func unlockMemory([]byte) error { return nil }

type errMemoryLockUnsupported struct{}

func (errMemoryLockUnsupported) Error() string {
	return "memory locking is unsupported on this platform"
}
