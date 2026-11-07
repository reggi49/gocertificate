package certificate

import (
	"runtime"
	"runtime/debug"
	"sync"
	"time"
)

// progressEmitInterval bounds how often job progress is persisted / pushed to the UI.
const progressEmitInterval = 200 * time.Millisecond

// maxDefaultConcurrency caps automatic worker count. PDF rendering is CPU + disk bound and
// every worker holds a full copy of the 300 DPI background in memory, so more workers bring
// little throughput but a lot of RAM pressure and UI starvation on consumer laptops.
const maxDefaultConcurrency = 4

// DefaultConcurrency returns half the logical CPUs, clamped to [1, maxDefaultConcurrency].
func DefaultConcurrency() int {
	n := runtime.NumCPU() / 2
	if n < 1 {
		n = 1
	}
	if n > maxDefaultConcurrency {
		n = maxDefaultConcurrency
	}
	return n
}

// progressThrottle allows at most one event per interval (plus forced final events).
type progressThrottle struct {
	mu       sync.Mutex
	interval time.Duration
	last     time.Time
}

func newProgressThrottle(interval time.Duration) *progressThrottle {
	return &progressThrottle{interval: interval}
}

func (t *progressThrottle) allow(force bool) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	now := time.Now()
	if force || now.Sub(t.last) >= t.interval {
		t.last = now
		return true
	}
	return false
}

// releaseMemory forces a GC cycle and returns freed pages to the OS.
func releaseMemory() {
	runtime.GC()
	debug.FreeOSMemory()
}
