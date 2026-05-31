// Package darwin provides Darwin-specific process metrics via libproc.
package darwin

import "sync"

var (
	clkTckOnce sync.Once
	clkTckVal  float64
)

func getClkTck() float64 {
	clkTckOnce.Do(func() {
		clkTckVal = 100.0 // Darwin uses mach_absolute_time, CLK_TCK not needed
	})
	return clkTckVal
}