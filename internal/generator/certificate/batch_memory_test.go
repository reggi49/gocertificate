package certificate_test

import (
	"context"
	"os"
	"runtime"
	"runtime/debug"
	"sync/atomic"
	"testing"
	"time"

	"certificate-generator/internal/domain"
)

// TestBatchMemoryReleasedAfterGeneration renders a batch with the real Janaiz 300 DPI template
// and verifies that heap memory is returned after the job finishes (no retained leak).
// Run explicitly with: STRESS_BATCH=1 go test ./internal/generator/certificate -run MemoryReleased -v
func TestBatchMemoryReleasedAfterGeneration(t *testing.T) {
	if os.Getenv("STRESS_BATCH") == "" {
		t.Skip("set STRESS_BATCH=1 to run the batch memory stress test")
	}
	const count = 150
	generator, proj, _, _, _, cleanup := setupBatchTestEnvWithDoc(t, count, domain.DefaultJanaizDocument())
	defer cleanup()

	runtime.GC()
	var before runtime.MemStats
	runtime.ReadMemStats(&before)

	// Sample peak RSS-ish heap usage while the batch runs.
	var peakHeap atomic.Uint64
	stop := make(chan struct{})
	go func() {
		tk := time.NewTicker(20 * time.Millisecond)
		defer tk.Stop()
		for {
			select {
			case <-stop:
				return
			case <-tk.C:
				var m runtime.MemStats
				runtime.ReadMemStats(&m)
				if m.HeapInuse > peakHeap.Load() {
					peakHeap.Store(m.HeapInuse)
				}
			}
		}
	}()

	var events atomic.Int32
	start := time.Now()
	job, err := generator.ExecuteBatch(context.Background(), domain.BatchConfig{
		ProjectID:  proj.ID,
		TemplateID: "tmpl-batch-1",
		OutputDir:  t.TempDir(),
	}, func(domain.ProgressEvent) { events.Add(1) })
	close(stop)
	if err != nil {
		t.Fatalf("ExecuteBatch: %v", err)
	}
	elapsed := time.Since(start)

	var after runtime.MemStats
	runtime.ReadMemStats(&after)
	debug.FreeOSMemory()

	t.Logf("certs=%d status=%s elapsed=%s (%.1f ms/cert) progressEvents=%d",
		job.CompletedCount, job.Status, elapsed, float64(elapsed.Milliseconds())/float64(count), events.Load())
	t.Logf("heapInuse before=%.1fMB peak=%.1fMB after=%.1fMB heapReleasedToOS=%.1fMB numGoroutine=%d",
		mb(before.HeapInuse), mb(peakHeap.Load()), mb(after.HeapInuse), mb(after.HeapReleased), runtime.NumGoroutine())

	if after.HeapInuse > before.HeapInuse+64<<20 {
		t.Errorf("heap not released after batch: before=%.1fMB after=%.1fMB", mb(before.HeapInuse), mb(after.HeapInuse))
	}
}

func mb(v uint64) float64 { return float64(v) / (1 << 20) }
