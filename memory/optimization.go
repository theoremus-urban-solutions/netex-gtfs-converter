package memory

import (
	"runtime"
	"runtime/debug"
	"sync"
	"time"
)

// MemoryPool manages object pooling for frequent allocations
type MemoryPool struct {
	pools map[string]*sync.Pool
	mu    sync.RWMutex
}

// NewMemoryPool creates a new memory pool manager
func NewMemoryPool() *MemoryPool {
	return &MemoryPool{
		pools: make(map[string]*sync.Pool),
	}
}

// GetPool returns or creates a pool for the given type
func (mp *MemoryPool) GetPool(typeName string, newFunc func() interface{}) *sync.Pool {
	mp.mu.RLock()
	pool, exists := mp.pools[typeName]
	mp.mu.RUnlock()

	if !exists {
		mp.mu.Lock()
		// Double-check pattern
		if pool, exists = mp.pools[typeName]; !exists {
			pool = &sync.Pool{
				New: newFunc,
			}
			mp.pools[typeName] = pool
		}
		mp.mu.Unlock()
	}

	return pool
}

// MemoryManager provides memory optimization utilities
type MemoryManager struct {
	pool               *MemoryPool
	gcInterval         time.Duration
	memoryLimit        uint64
	lastGC             time.Time
	compressionEnabled bool
	batchSize          int
	mu                 sync.RWMutex
}

// NewMemoryManager creates a new memory manager with optimization features
func NewMemoryManager() *MemoryManager {
	return &MemoryManager{
		pool:               NewMemoryPool(),
		gcInterval:         time.Minute * 2,   // GC every 2 minutes
		memoryLimit:        500 * 1024 * 1024, // 500MB default limit
		compressionEnabled: true,
		batchSize:          1000, // Process in batches of 1000
	}
}

// SetMemoryLimit sets the memory limit for automatic GC triggers
func (mm *MemoryManager) SetMemoryLimit(limitMB uint64) {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	mm.memoryLimit = limitMB * 1024 * 1024
}

// SetGCInterval sets how often to check for memory pressure
func (mm *MemoryManager) SetGCInterval(interval time.Duration) {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	mm.gcInterval = interval
}

// SetBatchSize sets the processing batch size for large datasets
func (mm *MemoryManager) SetBatchSize(size int) {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	mm.batchSize = size
}

// GetBatchSize returns the current batch processing size
func (mm *MemoryManager) GetBatchSize() int {
	mm.mu.RLock()
	defer mm.mu.RUnlock()
	return mm.batchSize
}

// CheckMemoryPressure checks if memory usage is approaching limits
func (mm *MemoryManager) CheckMemoryPressure() bool {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// Read the limit without holding lock to avoid deadlock
	limit := mm.getMemoryLimit()

	return memStats.Alloc > limit
}

// getMemoryLimit safely returns the memory limit
func (mm *MemoryManager) getMemoryLimit() uint64 {
	mm.mu.RLock()
	defer mm.mu.RUnlock()
	return mm.memoryLimit
}

// ForceGC triggers garbage collection if needed
func (mm *MemoryManager) ForceGC() {
	now := time.Now()

	// Check conditions without holding lock to avoid deadlock
	shouldGC := false

	mm.mu.Lock()
	if now.Sub(mm.lastGC) > mm.gcInterval {
		shouldGC = true
		mm.lastGC = now
	}
	mm.mu.Unlock()

	// Check memory pressure separately (doesn't need lock)
	if !shouldGC && mm.CheckMemoryPressure() {
		shouldGC = true
		mm.mu.Lock()
		mm.lastGC = now
		mm.mu.Unlock()
	}

	if shouldGC {
		runtime.GC()
		debug.FreeOSMemory() // Return memory to OS
	}
}

// GetMemoryStats returns current memory statistics
func (mm *MemoryManager) GetMemoryStats() MemoryStats {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	return MemoryStats{
		AllocatedMB:   float64(memStats.Alloc) / (1024 * 1024),
		TotalAllocMB:  float64(memStats.TotalAlloc) / (1024 * 1024),
		SystemMB:      float64(memStats.Sys) / (1024 * 1024),
		HeapAllocMB:   float64(memStats.HeapAlloc) / (1024 * 1024),
		HeapSysMB:     float64(memStats.HeapSys) / (1024 * 1024),
		StackInUseMB:  float64(memStats.StackInuse) / (1024 * 1024),
		GCCycles:      memStats.NumGC,
		LastGCTime:    time.Unix(0, int64(memStats.LastGC)), //nolint:gosec // intentional conversion of nanoseconds to int64
		GCCPUFraction: memStats.GCCPUFraction,
	}
}

// MemoryStats holds memory usage statistics
type MemoryStats struct {
	AllocatedMB   float64   `json:"allocated_mb"`
	TotalAllocMB  float64   `json:"total_alloc_mb"`
	SystemMB      float64   `json:"system_mb"`
	HeapAllocMB   float64   `json:"heap_alloc_mb"`
	HeapSysMB     float64   `json:"heap_sys_mb"`
	StackInUseMB  float64   `json:"stack_inuse_mb"`
	GCCycles      uint32    `json:"gc_cycles"`
	LastGCTime    time.Time `json:"last_gc_time"`
	GCCPUFraction float64   `json:"gc_cpu_fraction"`
}

// BatchProcessor provides memory-efficient batch processing
type BatchProcessor struct {
	manager   *MemoryManager
	batchSize int
}

// NewBatchProcessor creates a new batch processor
func NewBatchProcessor(manager *MemoryManager) *BatchProcessor {
	return &BatchProcessor{
		manager:   manager,
		batchSize: manager.GetBatchSize(),
	}
}

// ProcessInBatches processes a large slice in memory-efficient batches
func (bp *BatchProcessor) ProcessInBatches(items []interface{}, processor func(batch []interface{}) error) error {
	if len(items) == 0 {
		return nil
	}

	for i := 0; i < len(items); i += bp.batchSize {
		end := i + bp.batchSize
		if end > len(items) {
			end = len(items)
		}

		batch := items[i:end]

		// Process batch
		if err := processor(batch); err != nil {
			return err
		}

		// Check memory pressure and force GC if needed
		bp.manager.ForceGC()
	}

	return nil
}

// StreamProcessor processes items one at a time with memory management
type StreamProcessor struct {
	manager        *MemoryManager
	processedCount int
	gcThreshold    int
}

// NewStreamProcessor creates a new stream processor
func NewStreamProcessor(manager *MemoryManager) *StreamProcessor {
	return &StreamProcessor{
		manager:     manager,
		gcThreshold: 10000, // GC every 10k items
	}
}

// ProcessItem processes a single item with memory management
func (sp *StreamProcessor) ProcessItem(item interface{}, processor func(interface{}) error) error {
	err := processor(item)
	if err != nil {
		return err
	}

	sp.processedCount++

	// Periodic memory management
	if sp.processedCount%sp.gcThreshold == 0 {
		sp.manager.ForceGC()
	}

	return nil
}

// Reset resets the processor counters
func (sp *StreamProcessor) Reset() {
	sp.processedCount = 0
}

// GetProcessedCount returns the number of processed items
func (sp *StreamProcessor) GetProcessedCount() int {
	return sp.processedCount
}

// MemoryOptimizedBuffer provides a memory-efficient buffer with automatic compression
type MemoryOptimizedBuffer struct {
	buffer          []byte
	compressionPool *sync.Pool
	compressed      bool
	maxSize         int
	mu              sync.RWMutex
}

// NewMemoryOptimizedBuffer creates a new memory-optimized buffer
func NewMemoryOptimizedBuffer(maxSize int) *MemoryOptimizedBuffer {
	return &MemoryOptimizedBuffer{
		buffer:  make([]byte, 0, maxSize/4), // Start with 1/4 capacity
		maxSize: maxSize,
		compressionPool: &sync.Pool{
			New: func() interface{} {
				return make([]byte, 0, maxSize)
			},
		},
	}
}

// Write appends data to the buffer
func (mob *MemoryOptimizedBuffer) Write(data []byte) (int, error) {
	mob.mu.Lock()
	defer mob.mu.Unlock()

	// Check if buffer would exceed max size
	if len(mob.buffer)+len(data) > mob.maxSize {
		// Implement compression or chunking logic here
		return 0, ErrBufferFull
	}

	mob.buffer = append(mob.buffer, data...)
	return len(data), nil
}

// Read returns the buffered data
func (mob *MemoryOptimizedBuffer) Read() []byte {
	mob.mu.RLock()
	defer mob.mu.RUnlock()

	// Return copy to prevent external modification
	result := make([]byte, len(mob.buffer))
	copy(result, mob.buffer)
	return result
}

// Reset clears the buffer
func (mob *MemoryOptimizedBuffer) Reset() {
	mob.mu.Lock()
	defer mob.mu.Unlock()

	mob.buffer = mob.buffer[:0] // Keep capacity, reset length
	mob.compressed = false
}

// Size returns the current buffer size
func (mob *MemoryOptimizedBuffer) Size() int {
	mob.mu.RLock()
	defer mob.mu.RUnlock()
	return len(mob.buffer)
}

// Error types
var (
	ErrBufferFull = &MemoryError{Message: "buffer is full"}
)

// MemoryError represents a memory-related error
type MemoryError struct {
	Message string
}

func (e *MemoryError) Error() string {
	return e.Message
}

// MemoryMonitor provides continuous memory monitoring
type MemoryMonitor struct {
	manager   *MemoryManager
	interval  time.Duration
	stopCh    chan struct{}
	running   bool
	mu        sync.Mutex
	callbacks []func(MemoryStats)
}

// NewMemoryMonitor creates a new memory monitor
func NewMemoryMonitor(manager *MemoryManager) *MemoryMonitor {
	return &MemoryMonitor{
		manager:   manager,
		interval:  time.Second * 30, // Monitor every 30 seconds
		stopCh:    make(chan struct{}),
		callbacks: make([]func(MemoryStats), 0),
	}
}

// AddCallback adds a callback for memory statistics
func (mm *MemoryMonitor) AddCallback(callback func(MemoryStats)) {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	mm.callbacks = append(mm.callbacks, callback)
}

// Start begins memory monitoring
func (mm *MemoryMonitor) Start() {
	mm.mu.Lock()
	if mm.running {
		mm.mu.Unlock()
		return
	}
	mm.running = true
	mm.mu.Unlock()

	go mm.monitor()
}

// Stop ends memory monitoring
func (mm *MemoryMonitor) Stop() {
	mm.mu.Lock()
	if !mm.running {
		mm.mu.Unlock()
		return
	}
	mm.running = false
	mm.mu.Unlock()

	close(mm.stopCh)
}

func (mm *MemoryMonitor) monitor() {
	ticker := time.NewTicker(mm.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			stats := mm.manager.GetMemoryStats()

			// Check for memory pressure
			if mm.manager.CheckMemoryPressure() {
				mm.manager.ForceGC()
			}

			// Call callbacks
			mm.mu.Lock()
			for _, callback := range mm.callbacks {
				callback(stats)
			}
			mm.mu.Unlock()

		case <-mm.stopCh:
			return
		}
	}
}
