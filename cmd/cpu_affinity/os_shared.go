package cpu_affinity

type CpuAffinitySysCall interface {
	GetProcessId() uint
	GetThreadId() uint
	IsImplemented() bool
	SetCpuAffinities(pid uint, cpus ...uint) error
	GetCpuAffinities(pid uint) (map[uint]bool, error)
}
