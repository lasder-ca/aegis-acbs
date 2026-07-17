//go:build !linux && !darwin

package bench

func readPeakRSSBytes() uint64 { return 0 }
