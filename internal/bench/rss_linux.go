//go:build linux

package bench

import "syscall"

func readPeakRSSBytes() uint64 {
	var usage syscall.Rusage
	if err := syscall.Getrusage(syscall.RUSAGE_SELF, &usage); err != nil || usage.Maxrss <= 0 {
		return 0
	}
	// Linux reports ru_maxrss in KiB.
	return uint64(usage.Maxrss) * 1024
}
