package memory

import (
	"runtime"
	"sync"
	"testing"
	"time"
)

const (
	testString = "test"
)

func TestNewMemoryPool(t *testing.T) {
	pool := NewMemoryPool()
	if pool == nil {
		t.Fatal("NewMemoryPool() returned nil")
	}
	if pool.pools == nil {
		t.Error("MemoryPool.pools is nil")
	}
}

func TestMemoryPool_GetPool(t *testing.T) {
	mp := NewMemoryPool()

	// Test getting a new pool
	newFunc := func() interface{} { return testString }
	pool1 := mp.GetPool("test", newFunc)
	if pool1 == nil {
		t.Error("GetPool() returned nil")
	}

	// Test getting the same pool again
	pool2 := mp.GetPool("test", newFunc)
	if pool1 != pool2 {
		t.Error("GetPool() should return the same pool for the same type")
	}

	// Test that the pool works
	obj := pool1.Get()
	if obj != "test" {
		t.Errorf("Pool returned wrong object: expected 'test', got %v", obj)
	}
}

func TestMemoryPool_ConcurrentAccess(t *testing.T) {
	mp := NewMemoryPool()
	var wg sync.WaitGroup
	numGoroutines := 10
	numOperations := 100

	// Test concurrent access to GetPool
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			newFunc := func() interface{} { return id }

			for j := 0; j < numOperations; j++ {
				pool := mp.GetPool("concurrent", newFunc)
				if pool == nil {
					t.Errorf("GetPool() returned nil in goroutine %d", id)
				}
			}
		}(i)
	}

	wg.Wait()

	// Verify only one pool was created
	if len(mp.pools) != 1 {
		t.Errorf("Expected 1 pool, got %d", len(mp.pools))
	}
}

func TestNewMemoryManager(t *testing.T) {
	mm := NewMemoryManager()
	if mm == nil {
		t.Fatal("NewMemoryManager() returned nil")
	}

	// Check default values
	if mm.gcInterval != time.Minute*2 {
		t.Errorf("Expected default GC interval 2 minutes, got %v", mm.gcInterval)
	}
	if mm.memoryLimit != 500*1024*1024 {
		t.Errorf("Expected default memory limit 500MB, got %d", mm.memoryLimit)
	}
	if mm.batchSize != 1000 {
		t.Errorf("Expected default batch size 1000, got %d", mm.batchSize)
	}
}

func TestMemoryManager_Configuration(t *testing.T) {
	mm := NewMemoryManager()

	// Test SetMemoryLimit
	mm.SetMemoryLimit(100)
	expected := uint64(100 * 1024 * 1024)
	if mm.getMemoryLimit() != expected {
		t.Errorf("Expected memory limit %d, got %d", expected, mm.getMemoryLimit())
	}

	// Test SetGCInterval
	interval := time.Minute * 5
	mm.SetGCInterval(interval)
	if mm.gcInterval != interval {
		t.Errorf("Expected GC interval %v, got %v", interval, mm.gcInterval)
	}

	// Test SetBatchSize
	batchSize := 500
	mm.SetBatchSize(batchSize)
	if mm.GetBatchSize() != batchSize {
		t.Errorf("Expected batch size %d, got %d", batchSize, mm.GetBatchSize())
	}
}

func TestMemoryManager_CheckMemoryPressure(t *testing.T) {
	mm := NewMemoryManager()

	// Set very low memory limit to trigger pressure
	mm.SetMemoryLimit(1) // 1MB

	pressure := mm.CheckMemoryPressure()
	// This should typically be true as Go programs usually use more than 1MB
	if !pressure {
		// This is not necessarily an error as it depends on system state
		t.Logf("Memory pressure is false with 1MB limit")
	}

	// Set very high memory limit
	mm.SetMemoryLimit(10000) // 10GB
	pressure = mm.CheckMemoryPressure()
	if pressure {
		t.Error("Memory pressure should be false with high limit")
	}
}

func TestMemoryManager_ForceGC(t *testing.T) {
	mm := NewMemoryManager()

	// Get initial GC count
	var initialStats runtime.MemStats
	runtime.ReadMemStats(&initialStats)
	initialGCCount := initialStats.NumGC

	// Force GC with time condition
	mm.SetGCInterval(0) // Force immediate GC
	mm.ForceGC()

	// Give time for GC to potentially complete
	time.Sleep(10 * time.Millisecond)

	var finalStats runtime.MemStats
	runtime.ReadMemStats(&finalStats)

	// GC should have been called (though Go's GC behavior is not deterministic)
	if finalStats.NumGC < initialGCCount {
		t.Error("GC count should not decrease")
	}
}

func TestMemoryManager_GetMemoryStats(t *testing.T) {
	mm := NewMemoryManager()

	stats := mm.GetMemoryStats()

	// Check that stats are populated with reasonable values
	if stats.AllocatedMB < 0 {
		t.Error("AllocatedMB should be non-negative")
	}
	if stats.TotalAllocMB < stats.AllocatedMB {
		t.Error("TotalAllocMB should be >= AllocatedMB")
	}
	if stats.SystemMB < 0 {
		t.Error("SystemMB should be non-negative")
	}
	if stats.HeapAllocMB < 0 {
		t.Error("HeapAllocMB should be non-negative")
	}
}

func TestNewBatchProcessor(t *testing.T) {
	mm := NewMemoryManager()
	bp := NewBatchProcessor(mm)

	if bp == nil {
		t.Fatal("NewBatchProcessor() returned nil")
	}
	if bp.manager != mm {
		t.Error("BatchProcessor should reference the memory manager")
	}
	if bp.batchSize != mm.GetBatchSize() {
		t.Error("BatchProcessor should use manager's batch size")
	}
}

func TestBatchProcessor_ProcessInBatches(t *testing.T) {
	mm := NewMemoryManager()
	mm.SetBatchSize(3) // Small batch size for testing
	bp := NewBatchProcessor(mm)

	// Create test items
	items := make([]interface{}, 10)
	for i := 0; i < 10; i++ {
		items[i] = i
	}

	var processedBatches [][]interface{}
	processor := func(batch []interface{}) error {
		// Make a copy to avoid slice issues
		batchCopy := make([]interface{}, len(batch))
		copy(batchCopy, batch)
		processedBatches = append(processedBatches, batchCopy)
		return nil
	}

	err := bp.ProcessInBatches(items, processor)
	if err != nil {
		t.Fatalf("ProcessInBatches() failed: %v", err)
	}

	// Check number of batches
	expectedBatches := 4 // 10 items / 3 batch size = 3 full + 1 partial
	if len(processedBatches) != expectedBatches {
		t.Errorf("Expected %d batches, got %d", expectedBatches, len(processedBatches))
	}

	// Check batch sizes
	if len(processedBatches[0]) != 3 {
		t.Errorf("First batch should have 3 items, got %d", len(processedBatches[0]))
	}
	if len(processedBatches[3]) != 1 { // Last batch should have 1 item
		t.Errorf("Last batch should have 1 item, got %d", len(processedBatches[3]))
	}

	// Test with empty items
	err = bp.ProcessInBatches([]interface{}{}, processor)
	if err != nil {
		t.Errorf("ProcessInBatches() with empty slice should not error: %v", err)
	}
}

func TestNewStreamProcessor(t *testing.T) {
	mm := NewMemoryManager()
	sp := NewStreamProcessor(mm)

	if sp == nil {
		t.Fatal("NewStreamProcessor() returned nil")
	}
	if sp.manager != mm {
		t.Error("StreamProcessor should reference the memory manager")
	}
	if sp.gcThreshold != 10000 {
		t.Errorf("Expected GC threshold 10000, got %d", sp.gcThreshold)
	}
}

func TestStreamProcessor_ProcessItem(t *testing.T) {
	mm := NewMemoryManager()
	sp := NewStreamProcessor(mm)

	var processedItems []interface{}
	processor := func(item interface{}) error {
		processedItems = append(processedItems, item)
		return nil
	}

	// Process several items
	for i := 0; i < 5; i++ {
		err := sp.ProcessItem(i, processor)
		if err != nil {
			t.Fatalf("ProcessItem() failed: %v", err)
		}
	}

	// Check processed count
	if sp.GetProcessedCount() != 5 {
		t.Errorf("Expected processed count 5, got %d", sp.GetProcessedCount())
	}

	// Check processed items
	if len(processedItems) != 5 {
		t.Errorf("Expected 5 processed items, got %d", len(processedItems))
	}

	// Test Reset
	sp.Reset()
	if sp.GetProcessedCount() != 0 {
		t.Error("Reset() should set processed count to 0")
	}
}

func TestNewMemoryOptimizedBuffer(t *testing.T) {
	maxSize := 1024
	mob := NewMemoryOptimizedBuffer(maxSize)

	if mob == nil {
		t.Fatal("NewMemoryOptimizedBuffer() returned nil")
	}
	if mob.maxSize != maxSize {
		t.Errorf("Expected max size %d, got %d", maxSize, mob.maxSize)
	}
	if cap(mob.buffer) != maxSize/4 {
		t.Errorf("Expected initial capacity %d, got %d", maxSize/4, cap(mob.buffer))
	}
}

func TestMemoryOptimizedBuffer_WriteRead(t *testing.T) {
	mob := NewMemoryOptimizedBuffer(1024)

	// Test write
	data := []byte("test data")
	n, err := mob.Write(data)
	if err != nil {
		t.Fatalf("Write() failed: %v", err)
	}
	if n != len(data) {
		t.Errorf("Expected to write %d bytes, wrote %d", len(data), n)
	}

	// Test read
	result := mob.Read()
	if string(result) != string(data) {
		t.Errorf("Expected to read %s, got %s", string(data), string(result))
	}

	// Test size
	if mob.Size() != len(data) {
		t.Errorf("Expected size %d, got %d", len(data), mob.Size())
	}
}

func TestMemoryOptimizedBuffer_BufferFull(t *testing.T) {
	mob := NewMemoryOptimizedBuffer(10) // Very small buffer

	// Fill buffer to capacity
	data := make([]byte, 15) // Larger than max size
	_, err := mob.Write(data)
	if err != ErrBufferFull {
		t.Errorf("Expected ErrBufferFull, got %v", err)
	}
}

func TestMemoryOptimizedBuffer_Reset(t *testing.T) {
	mob := NewMemoryOptimizedBuffer(1024)

	// Add some data
	data := []byte("test data")
	if _, err := mob.Write(data); err != nil {
		t.Fatal(err)
	}

	// Check initial state
	if mob.Size() == 0 {
		t.Error("Buffer should not be empty after writing")
	}

	// Reset
	mob.Reset()

	// Check reset state
	if mob.Size() != 0 {
		t.Error("Buffer should be empty after reset")
	}
}

func TestMemoryError(t *testing.T) {
	err := &MemoryError{Message: "test error"}
	if err.Error() != "test error" {
		t.Errorf("Expected error message 'test error', got '%s'", err.Error())
	}
}

func TestNewMemoryMonitor(t *testing.T) {
	mm := NewMemoryManager()
	monitor := NewMemoryMonitor(mm)

	if monitor == nil {
		t.Fatal("NewMemoryMonitor() returned nil")
	}
	if monitor.manager != mm {
		t.Error("MemoryMonitor should reference the memory manager")
	}
	if monitor.interval != time.Second*30 {
		t.Errorf("Expected interval 30 seconds, got %v", monitor.interval)
	}
}

func TestMemoryMonitor_AddCallback(t *testing.T) {
	mm := NewMemoryManager()
	monitor := NewMemoryMonitor(mm)

	callbackCalled := false
	callback := func(stats MemoryStats) {
		callbackCalled = true
	}

	monitor.AddCallback(callback)

	// Check that callback was added
	if len(monitor.callbacks) != 1 {
		t.Errorf("Expected 1 callback, got %d", len(monitor.callbacks))
	}

	// Test callback execution
	stats := MemoryStats{}
	monitor.callbacks[0](stats)
	if !callbackCalled {
		t.Error("Callback was not called")
	}
}

func TestMemoryMonitor_StartStop(t *testing.T) {
	mm := NewMemoryManager()
	monitor := NewMemoryMonitor(mm)

	// Test start
	if monitor.running {
		t.Error("Monitor should not be running initially")
	}

	monitor.Start()
	if !monitor.running {
		t.Error("Monitor should be running after Start()")
	}

	// Test double start (should not cause issues)
	monitor.Start()
	if !monitor.running {
		t.Error("Monitor should still be running after double Start()")
	}

	// Test stop
	monitor.Stop()

	// Give time for goroutine to stop
	time.Sleep(10 * time.Millisecond)

	if monitor.running {
		t.Error("Monitor should not be running after Stop()")
	}

	// Test double stop (should not cause issues)
	monitor.Stop()
}

func TestMemoryMonitor_WithCallback(t *testing.T) {
	mm := NewMemoryManager()
	monitor := NewMemoryMonitor(mm)

	// Use very short interval for testing
	monitor.interval = time.Millisecond * 10

	callbackCount := 0
	monitor.AddCallback(func(stats MemoryStats) {
		callbackCount++
	})

	monitor.Start()

	// Wait for a few callbacks
	time.Sleep(50 * time.Millisecond)

	monitor.Stop()

	// Should have received at least one callback
	if callbackCount == 0 {
		t.Error("Expected at least one callback to be called")
	}
}

func TestMemoryStats_Fields(t *testing.T) {
	mm := NewMemoryManager()
	stats := mm.GetMemoryStats()

	// Test that all fields are accessible and reasonable
	fields := []struct {
		name  string
		value float64
	}{
		{"AllocatedMB", stats.AllocatedMB},
		{"TotalAllocMB", stats.TotalAllocMB},
		{"SystemMB", stats.SystemMB},
		{"HeapAllocMB", stats.HeapAllocMB},
		{"HeapSysMB", stats.HeapSysMB},
		{"StackInUseMB", stats.StackInUseMB},
	}

	for _, field := range fields {
		if field.value < 0 {
			t.Errorf("Field %s should be non-negative, got %f", field.name, field.value)
		}
	}

	// GCCycles should be reasonable
	if stats.GCCycles > 1000000 {
		t.Errorf("GCCycles seems unreasonably high: %d", stats.GCCycles)
	}
}

// Benchmark tests
func BenchmarkMemoryPool_GetPool(b *testing.B) {
	mp := NewMemoryPool()
	newFunc := func() interface{} { return "test" }

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mp.GetPool("benchmark", newFunc)
	}
}

func BenchmarkBatchProcessor(b *testing.B) {
	mm := NewMemoryManager()
	bp := NewBatchProcessor(mm)

	items := make([]interface{}, 1000)
	for i := 0; i < 1000; i++ {
		items[i] = i
	}

	processor := func(batch []interface{}) error {
		return nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := bp.ProcessInBatches(items, processor); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMemoryOptimizedBuffer_Write(b *testing.B) {
	mob := NewMemoryOptimizedBuffer(1024 * 1024) // 1MB buffer
	data := []byte("benchmark data")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := mob.Write(data); err != nil {
			b.Fatal(err)
		}
		if mob.Size() > 1024*512 { // Reset when half full
			mob.Reset()
		}
	}
}
