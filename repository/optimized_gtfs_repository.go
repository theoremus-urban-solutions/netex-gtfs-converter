package repository

import (
	"runtime"
	"sync"

	"github.com/theoremus-urban-solutions/netex-gtfs-converter/memory"
	"github.com/theoremus-urban-solutions/netex-gtfs-converter/model"
	"github.com/theoremus-urban-solutions/netex-gtfs-converter/producer"
)

// OptimizedGtfsRepository implements GtfsRepository with memory optimization
type OptimizedGtfsRepository struct {
	*DefaultGtfsRepository
	memoryManager   *memory.MemoryManager
	streamProcessor *memory.StreamProcessor
	objectPools     map[string]*sync.Pool
	mu              sync.RWMutex
}

// NewOptimizedGtfsRepository creates a new memory-optimized GTFS repository
func NewOptimizedGtfsRepository() producer.GtfsRepository {
	memManager := memory.NewMemoryManager()

	repo := &OptimizedGtfsRepository{
		DefaultGtfsRepository: &DefaultGtfsRepository{
			agencies:       make(map[string]*model.Agency),
			routes:         make(map[string]*model.GtfsRoute),
			trips:          make(map[string]*model.Trip),
			stops:          make(map[string]*model.Stop),
			stopTimes:      make([]*model.StopTime, 0),
			calendars:      make(map[string]*model.Calendar),
			calendarDates:  make([]*model.CalendarDate, 0),
			transfers:      make([]*model.Transfer, 0),
			shapes:         make([]*model.Shape, 0),
			frequencies:    make([]*model.Frequency, 0),
			fareAttributes: make(map[string]*model.FareAttribute),
			fareRules:      make([]*model.FareRule, 0),
			pathways:       make([]*model.Pathway, 0),
			levels:         make([]*model.Level, 0),
		},
		memoryManager:   memManager,
		streamProcessor: memory.NewStreamProcessor(memManager),
		objectPools:     make(map[string]*sync.Pool),
	}

	repo.initializePools()
	return repo
}

// initializePools sets up object pools for frequent GTFS allocations
func (r *OptimizedGtfsRepository) initializePools() {
	// Stop Time pool (most frequently allocated)
	r.objectPools["StopTime"] = &sync.Pool{
		New: func() interface{} {
			return &model.StopTime{}
		},
	}

	// Trip pool
	r.objectPools["Trip"] = &sync.Pool{
		New: func() interface{} {
			return &model.Trip{}
		},
	}

	// Stop pool
	r.objectPools["Stop"] = &sync.Pool{
		New: func() interface{} {
			return &model.Stop{}
		},
	}

	// Shape pool
	r.objectPools["Shape"] = &sync.Pool{
		New: func() interface{} {
			return &model.Shape{}
		},
	}

	// Transfer pool
	r.objectPools["Transfer"] = &sync.Pool{
		New: func() interface{} {
			return &model.Transfer{}
		},
	}

	// Frequency pool
	r.objectPools["Frequency"] = &sync.Pool{
		New: func() interface{} {
			return &model.Frequency{}
		},
	}

	// Pathway pool
	r.objectPools["Pathway"] = &sync.Pool{
		New: func() interface{} {
			return &model.Pathway{}
		},
	}
}

// GetFromPool retrieves an object from the pool
func (r *OptimizedGtfsRepository) GetFromPool(typeName string) interface{} {
	r.mu.RLock()
	pool, exists := r.objectPools[typeName]
	r.mu.RUnlock()

	if exists {
		return pool.Get()
	}
	return nil
}

// ReturnToPool returns an object to the pool after resetting it
func (r *OptimizedGtfsRepository) ReturnToPool(typeName string, obj interface{}) {
	r.mu.RLock()
	pool, exists := r.objectPools[typeName]
	r.mu.RUnlock()

	if exists {
		r.resetGtfsObject(obj)
		pool.Put(obj)
	}
}

// resetGtfsObject resets a GTFS object to its zero state for reuse
func (r *OptimizedGtfsRepository) resetGtfsObject(obj interface{}) {
	switch v := obj.(type) {
	case *model.StopTime:
		*v = model.StopTime{}
	case *model.Trip:
		*v = model.Trip{}
	case *model.Stop:
		*v = model.Stop{}
	case *model.Shape:
		*v = model.Shape{}
	case *model.Transfer:
		*v = model.Transfer{}
	case *model.Frequency:
		*v = model.Frequency{}
	case *model.Pathway:
		*v = model.Pathway{}
	}
}

// SaveEntity saves an entity with memory optimization
func (r *OptimizedGtfsRepository) SaveEntity(entity interface{}) error {
	return r.streamProcessor.ProcessItem(entity, func(item interface{}) error {
		return r.DefaultGtfsRepository.SaveEntity(item)
	})
}

// SaveStopTimesBatch saves stop times in memory-efficient batches
func (r *OptimizedGtfsRepository) SaveStopTimesBatch(stopTimes []*model.StopTime) error {
	batchProcessor := memory.NewBatchProcessor(r.memoryManager)

	// Convert to interface slice for batch processing
	entities := make([]interface{}, len(stopTimes))
	for i, st := range stopTimes {
		entities[i] = st
	}

	return batchProcessor.ProcessInBatches(entities, func(batch []interface{}) error {
		for _, entity := range batch {
			stopTime := entity.(*model.StopTime)
			r.stopTimes = append(r.stopTimes, stopTime)
		}
		return nil
	})
}

// GetStopTimesByTrip returns stop times for a trip with memory management
func (r *OptimizedGtfsRepository) GetStopTimesByTrip(tripId string) []*model.StopTime {
	// Check memory pressure before processing
	if r.memoryManager.CheckMemoryPressure() {
		r.memoryManager.ForceGC()
	}

	var result []*model.StopTime
	for _, stopTime := range r.stopTimes {
		if stopTime.TripID == tripId {
			result = append(result, stopTime)
		}
	}
	return result
}

// StreamProcessStopTimes processes stop times in a memory-efficient stream
func (r *OptimizedGtfsRepository) StreamProcessStopTimes(processor func(*model.StopTime) error) error {
	r.streamProcessor.Reset()

	for _, stopTime := range r.stopTimes {
		err := r.streamProcessor.ProcessItem(stopTime, func(item interface{}) error {
			return processor(item.(*model.StopTime))
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// OptimizeStopTimes sorts and optimizes stop times storage for better memory access patterns
func (r *OptimizedGtfsRepository) OptimizeStopTimes() {
	// Sort stop times by trip ID for better memory locality
	// This can significantly improve cache performance

	// Group by trip ID
	tripStopTimes := make(map[string][]*model.StopTime)
	for _, st := range r.stopTimes {
		tripStopTimes[st.TripID] = append(tripStopTimes[st.TripID], st)
	}

	// Rebuild the slice with better locality
	r.stopTimes = r.stopTimes[:0] // Clear but keep capacity
	for _, stopTimes := range tripStopTimes {
		r.stopTimes = append(r.stopTimes, stopTimes...)
	}

	// Force GC to clean up temporary maps
	_ = tripStopTimes // Suppress unused variable warning
	runtime.GC()
}

// GetMemoryStats returns current memory usage statistics
func (r *OptimizedGtfsRepository) GetMemoryStats() memory.MemoryStats {
	return r.memoryManager.GetMemoryStats()
}

// ForceMemoryCleanup triggers garbage collection and memory cleanup
func (r *OptimizedGtfsRepository) ForceMemoryCleanup() {
	r.memoryManager.ForceGC()
}

// GetEntityCount returns the count of various GTFS entities for memory tracking
func (r *OptimizedGtfsRepository) GetEntityCount() map[string]int {
	return map[string]int{
		"agencies":       len(r.agencies),
		"routes":         len(r.routes),
		"trips":          len(r.trips),
		"stops":          len(r.stops),
		"stopTimes":      len(r.stopTimes),
		"calendars":      len(r.calendars),
		"calendarDates":  len(r.calendarDates),
		"transfers":      len(r.transfers),
		"shapes":         len(r.shapes),
		"frequencies":    len(r.frequencies),
		"fareAttributes": len(r.fareAttributes),
		"fareRules":      len(r.fareRules),
		"pathways":       len(r.pathways),
		"levels":         len(r.levels),
	}
}

// ClearLargeCollections clears large collections to free memory
func (r *OptimizedGtfsRepository) ClearLargeCollections() {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Clear the largest collections that can be rebuilt if needed
	r.stopTimes = r.stopTimes[:0]
	r.calendarDates = r.calendarDates[:0]
	r.transfers = r.transfers[:0]
	r.shapes = r.shapes[:0]
	r.frequencies = r.frequencies[:0]
	r.fareRules = r.fareRules[:0]
	r.pathways = r.pathways[:0]
	r.levels = r.levels[:0]

	runtime.GC()
}

// SetMemoryLimit configures the memory limit for automatic cleanup
func (r *OptimizedGtfsRepository) SetMemoryLimit(limitMB uint64) {
	r.memoryManager.SetMemoryLimit(limitMB)
}

// SetBatchSize configures the batch processing size
func (r *OptimizedGtfsRepository) SetBatchSize(size int) {
	r.memoryManager.SetBatchSize(size)
}

// GetStopTimesCount returns the number of stop times for memory monitoring
func (r *OptimizedGtfsRepository) GetStopTimesCount() int {
	return len(r.stopTimes)
}

// StreamWriteStopTimes writes stop times in a memory-efficient streaming manner
func (r *OptimizedGtfsRepository) StreamWriteStopTimes(writer func([]*model.StopTime) error, batchSize int) error {
	if len(r.stopTimes) == 0 {
		return nil
	}

	batchProcessor := memory.NewBatchProcessor(r.memoryManager)
	if err := batchProcessor.ProcessInBatches(make([]interface{}, len(r.stopTimes)), func(batch []interface{}) error {
		start := len(batch) * (batchSize)
		end := start + len(batch)
		if end > len(r.stopTimes) {
			end = len(r.stopTimes)
		}

		if start < len(r.stopTimes) {
			batchStopTimes := r.stopTimes[start:end]
			return writer(batchStopTimes)
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}
