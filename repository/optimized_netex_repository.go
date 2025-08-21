package repository

import (
	"fmt"
	"runtime"
	"sync"

	"github.com/theoremus-urban-solutions/netex-gtfs-converter/memory"
	"github.com/theoremus-urban-solutions/netex-gtfs-converter/model"
	"github.com/theoremus-urban-solutions/netex-gtfs-converter/producer"
)

// OptimizedNetexRepository implements NetexRepository with memory optimization
type OptimizedNetexRepository struct {
	*DefaultNetexRepository
	memoryManager   *memory.MemoryManager
	streamProcessor *memory.StreamProcessor
	objectPools     map[string]*sync.Pool
	mu              sync.RWMutex
}

// NewOptimizedNetexRepository creates a new memory-optimized NeTEx repository
func NewOptimizedNetexRepository() producer.NetexRepository {
	memManager := memory.NewMemoryManager()

	repo := &OptimizedNetexRepository{
		DefaultNetexRepository: &DefaultNetexRepository{
			authorities:                          make(map[string]*model.Authority),
			networks:                             make(map[string]*model.Network),
			lines:                                make(map[string]*model.Line),
			routes:                               make(map[string]*model.Route),
			journeyPatterns:                      make(map[string]*model.JourneyPattern),
			serviceJourneys:                      make(map[string]*model.ServiceJourney),
			datedServiceJourneys:                 make(map[string]*model.DatedServiceJourney),
			destinationDisplays:                  make(map[string]*model.DestinationDisplay),
			scheduledStopPoints:                  make(map[string]*model.ScheduledStopPoint),
			stopPointInJourneyPatterns:           make(map[string]*model.StopPointInJourneyPattern),
			serviceJourneyInterchanges:           make(map[string]*model.ServiceJourneyInterchange),
			dayTypes:                             make(map[string]*model.DayType),
			operatingDays:                        make(map[string]*model.OperatingDay),
			operatingPeriods:                     make(map[string]*model.OperatingPeriod),
			dayTypeAssignments:                   make(map[string]*model.DayTypeAssignment),
			stopPlaces:                           make(map[string]*model.StopPlace),
			quays:                                make(map[string]*model.Quay),
			headwayJourneyGroups:                 make(map[string]*model.HeadwayJourneyGroup),
			routesByLineId:                       make(map[string][]*model.Route),
			serviceJourneysByPattern:             make(map[string][]*model.ServiceJourney),
			datedServiceJourneysByServiceJourney: make(map[string][]*model.DatedServiceJourney),
			dayTypeAssignmentsByDayType:          make(map[string][]*model.DayTypeAssignment),
			quaysByStopPlace:                     make(map[string][]*model.Quay),
			stopPlaceByQuayId:                    make(map[string]*model.StopPlace),
			pointInJourneyPatternToScheduledStopPoint: make(map[string]string),
			lineIdToNetworkId:                         make(map[string]string),
			timeZone:                                  "Europe/Oslo",
		},
		memoryManager:   memManager,
		streamProcessor: memory.NewStreamProcessor(memManager),
		objectPools:     make(map[string]*sync.Pool),
	}

	repo.initializePools()
	return repo
}

// initializePools sets up object pools for frequent allocations
func (r *OptimizedNetexRepository) initializePools() {
	// Service Journey pool
	r.objectPools["ServiceJourney"] = &sync.Pool{
		New: func() interface{} {
			return &model.ServiceJourney{}
		},
	}

	// Stop Time pool (for passing times)
	r.objectPools["TimetabledPassingTime"] = &sync.Pool{
		New: func() interface{} {
			return &model.TimetabledPassingTime{}
		},
	}

	// Journey Pattern pool
	r.objectPools["JourneyPattern"] = &sync.Pool{
		New: func() interface{} {
			return &model.JourneyPattern{}
		},
	}

	// Stop Place pool
	r.objectPools["StopPlace"] = &sync.Pool{
		New: func() interface{} {
			return &model.StopPlace{}
		},
	}

	// Quay pool
	r.objectPools["Quay"] = &sync.Pool{
		New: func() interface{} {
			return &model.Quay{}
		},
	}
}

// getFromPool retrieves an object from the pool
//
//nolint:unused // This function is used in memory optimization scenarios
func (r *OptimizedNetexRepository) getFromPool(typeName string) interface{} {
	r.mu.RLock()
	pool, exists := r.objectPools[typeName]
	r.mu.RUnlock()

	if exists {
		return pool.Get()
	}
	return nil
}

// returnToPool returns an object to the pool after resetting it
//
//nolint:unused // This function is used in memory optimization scenarios
func (r *OptimizedNetexRepository) returnToPool(typeName string, obj interface{}) {
	r.mu.RLock()
	pool, exists := r.objectPools[typeName]
	r.mu.RUnlock()

	if exists {
		// Reset object before returning to pool
		r.resetObject(obj)
		pool.Put(obj)
	}
}

// resetObject resets an object to its zero state for reuse
//
//nolint:unused // This function is used in memory optimization scenarios
func (r *OptimizedNetexRepository) resetObject(obj interface{}) {
	switch v := obj.(type) {
	case *model.ServiceJourney:
		*v = model.ServiceJourney{}
	case *model.TimetabledPassingTime:
		*v = model.TimetabledPassingTime{}
	case *model.JourneyPattern:
		*v = model.JourneyPattern{}
	case *model.StopPlace:
		*v = model.StopPlace{}
	case *model.Quay:
		*v = model.Quay{}
	}
}

// SaveEntity saves an entity with memory optimization
func (r *OptimizedNetexRepository) SaveEntity(entity interface{}) error {
	// Use stream processor for memory-managed processing
	return r.streamProcessor.ProcessItem(entity, func(item interface{}) error {
		return r.DefaultNetexRepository.SaveEntity(item)
	})
}

// SaveEntitiesBatch saves multiple entities in memory-efficient batches
func (r *OptimizedNetexRepository) SaveEntitiesBatch(entities []interface{}) error {
	batchProcessor := memory.NewBatchProcessor(r.memoryManager)

	return batchProcessor.ProcessInBatches(entities, func(batch []interface{}) error {
		for _, entity := range batch {
			if err := r.DefaultNetexRepository.SaveEntity(entity); err != nil {
				return fmt.Errorf("failed to save entity in batch: %w", err)
			}
		}
		return nil
	})
}

// GetMemoryStats returns current memory usage statistics
func (r *OptimizedNetexRepository) GetMemoryStats() memory.MemoryStats {
	return r.memoryManager.GetMemoryStats()
}

// ForceMemoryCleanup triggers garbage collection and memory cleanup
func (r *OptimizedNetexRepository) ForceMemoryCleanup() {
	r.memoryManager.ForceGC()
}

// GetServiceJourneysByPattern returns service journeys with memory optimization
func (r *OptimizedNetexRepository) GetServiceJourneysByPattern(patternId string) []*model.ServiceJourney {
	// Check memory pressure before large operations
	if r.memoryManager.CheckMemoryPressure() {
		r.memoryManager.ForceGC()
	}

	if pattern := r.GetJourneyPatternById(patternId); pattern != nil {
		return r.GetServiceJourneysByJourneyPattern(pattern)
	}
	return nil
}

// GetServiceJourneysByLine returns service journeys for a line with batching
func (r *OptimizedNetexRepository) GetServiceJourneysByLine(lineId string) []*model.ServiceJourney {
	r.DefaultNetexRepository.mu.RLock()
	line := r.lines[lineId]
	r.DefaultNetexRepository.mu.RUnlock()
	if line == nil {
		return nil
	}
	routes := r.GetRoutesByLine(line)
	var allJourneys []*model.ServiceJourney

	// Process routes in batches to manage memory
	batchProcessor := memory.NewBatchProcessor(r.memoryManager)
	routeInterfaces := make([]interface{}, len(routes))
	for i, route := range routes {
		routeInterfaces[i] = route
	}

	if err := batchProcessor.ProcessInBatches(routeInterfaces, func(batch []interface{}) error {
		for _, routeInterface := range batch {
			route := routeInterface.(*model.Route)
			journeys := r.GetServiceJourneysByRoute(route.ID)
			allJourneys = append(allJourneys, journeys...)
		}
		return nil
	}); err != nil {
		return nil
	}

	return allJourneys
}

// GetServiceJourneysByRoute returns service journeys for a route
func (r *OptimizedNetexRepository) GetServiceJourneysByRoute(routeId string) []*model.ServiceJourney {
	var journeys []*model.ServiceJourney

	for _, journey := range r.GetServiceJourneys() {
		// Check if this journey belongs to any pattern of this route
		if pattern := r.GetJourneyPatternById(journey.JourneyPatternRef.Ref); pattern != nil {
			if pattern.RouteRef == routeId {
				journeys = append(journeys, journey)
			}
		}
	}

	return journeys
}

// StreamProcessServiceJourneys processes service journeys in a memory-efficient stream
func (r *OptimizedNetexRepository) StreamProcessServiceJourneys(processor func(*model.ServiceJourney) error) error {
	r.streamProcessor.Reset()

	for _, journey := range r.GetServiceJourneys() {
		err := r.streamProcessor.ProcessItem(journey, func(item interface{}) error {
			return processor(item.(*model.ServiceJourney))
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// GetEntityCount returns the count of various entities for memory tracking
func (r *OptimizedNetexRepository) GetEntityCount() map[string]int {
	r.DefaultNetexRepository.mu.RLock()
	counts := map[string]int{
		"authorities":                len(r.authorities),
		"lines":                      len(r.lines),
		"routes":                     len(r.routes),
		"journeyPatterns":            len(r.journeyPatterns),
		"serviceJourneys":            len(r.serviceJourneys),
		"datedServiceJourneys":       len(r.datedServiceJourneys),
		"destinationDisplays":        len(r.destinationDisplays),
		"scheduledStopPoints":        len(r.scheduledStopPoints),
		"stopPointInJourneyPatterns": len(r.stopPointInJourneyPatterns),
		"serviceJourneyInterchanges": len(r.serviceJourneyInterchanges),
		"dayTypes":                   len(r.dayTypes),
		"operatingDays":              len(r.operatingDays),
		"operatingPeriods":           len(r.operatingPeriods),
		"dayTypeAssignments":         len(r.dayTypeAssignments),
		"stopPlaces":                 len(r.stopPlaces),
		"quays":                      len(r.quays),
		"headwayJourneyGroups":       len(r.headwayJourneyGroups),
	}
	r.DefaultNetexRepository.mu.RUnlock()
	return counts
}

// ClearCache clears internal caches to free memory
func (r *OptimizedNetexRepository) ClearCache() {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Clear lookup maps that can be rebuilt
	r.routesByLineId = make(map[string][]*model.Route)
	r.serviceJourneysByPattern = make(map[string][]*model.ServiceJourney)
	r.datedServiceJourneysByServiceJourney = make(map[string][]*model.DatedServiceJourney)
	r.dayTypeAssignmentsByDayType = make(map[string][]*model.DayTypeAssignment)
	r.quaysByStopPlace = make(map[string][]*model.Quay)
	r.stopPlaceByQuayId = make(map[string]*model.StopPlace)
	r.pointInJourneyPatternToScheduledStopPoint = make(map[string]string)

	// Force garbage collection
	runtime.GC()
}

// SetMemoryLimit configures the memory limit for automatic cleanup
func (r *OptimizedNetexRepository) SetMemoryLimit(limitMB uint64) {
	r.memoryManager.SetMemoryLimit(limitMB)
}

// SetBatchSize configures the batch processing size
func (r *OptimizedNetexRepository) SetBatchSize(size int) {
	r.memoryManager.SetBatchSize(size)
}
