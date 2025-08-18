package producer

import (
	"fmt"
	"sync"
	"time"

	"github.com/theoremus-urban-solutions/netex-gtfs-converter/memory"
	"github.com/theoremus-urban-solutions/netex-gtfs-converter/model"
)

// OptimizedStopTimeProducer produces stop times with memory optimization
type OptimizedStopTimeProducer struct {
	netexRepository NetexRepository
	gtfsRepository  GtfsRepository
	memoryManager   *memory.MemoryManager
	objectPool      *sync.Pool
}

// NewOptimizedStopTimeProducer creates a new optimized stop time producer
func NewOptimizedStopTimeProducer(netexRepo NetexRepository, gtfsRepo GtfsRepository) *OptimizedStopTimeProducer {
	memManager := memory.NewMemoryManager()

	producer := &OptimizedStopTimeProducer{
		netexRepository: netexRepo,
		gtfsRepository:  gtfsRepo,
		memoryManager:   memManager,
		objectPool: &sync.Pool{
			New: func() interface{} {
				return &model.StopTime{}
			},
		},
	}

	return producer
}

// Produce implements the StopTimeProducer interface
func (p *OptimizedStopTimeProducer) Produce(input StopTimeInput) (*model.StopTime, error) {
	// Basic single stop time production
	return nil, fmt.Errorf("single stop time production not implemented in optimized producer")
}

// ProduceStopTimesForTrip produces stop times with memory optimization
func (p *OptimizedStopTimeProducer) ProduceStopTimesForTrip(
	serviceJourney *model.ServiceJourney,
	trip *model.Trip,
	shape *model.Shape,
	lineName string,
) ([]*model.StopTime, error) {

	if serviceJourney == nil || serviceJourney.PassingTimes == nil {
		return nil, fmt.Errorf("service journey or passing times are nil")
	}

	passingTimes := serviceJourney.PassingTimes.TimetabledPassingTime
	if len(passingTimes) == 0 {
		return nil, fmt.Errorf("no passing times found for service journey %s", serviceJourney.ID)
	}

	// Pre-allocate slice with known capacity for better memory efficiency
	stopTimes := make([]*model.StopTime, 0, len(passingTimes))

	// Use stream processor for memory-managed processing
	streamProcessor := memory.NewStreamProcessor(p.memoryManager)

	for i, passingTime := range passingTimes {
		err := streamProcessor.ProcessItem(&passingTime, func(item interface{}) error {
			pt := item.(*model.TimetabledPassingTime)

			// Get stop time from pool
			stopTime := p.objectPool.Get().(*model.StopTime)
			defer func() {
				// Reset and return to pool after processing
				*stopTime = model.StopTime{}
				p.objectPool.Put(stopTime)
			}()

			// Fill stop time data
			if err := p.fillStopTimeData(stopTime, pt, trip, i+1, shape, lineName); err != nil {
				return fmt.Errorf("failed to fill stop time data: %w", err)
			}

			// Create a new stop time for storage (since we're returning the pooled one)
			persistentStopTime := &model.StopTime{
				TripID:            stopTime.TripID,
				ArrivalTime:       stopTime.ArrivalTime,
				DepartureTime:     stopTime.DepartureTime,
				StopID:            stopTime.StopID,
				StopSequence:      stopTime.StopSequence,
				StopHeadsign:      stopTime.StopHeadsign,
				PickupType:        stopTime.PickupType,
				DropOffType:       stopTime.DropOffType,
				ShapeDistTraveled: stopTime.ShapeDistTraveled,
				Timepoint:         stopTime.Timepoint,
			}

			stopTimes = append(stopTimes, persistentStopTime)
			return nil
		})

		if err != nil {
			return nil, fmt.Errorf("failed to process passing time %d: %w", i, err)
		}
	}

	return stopTimes, nil
}

// fillStopTimeData fills stop time data from passing time with optimizations
func (p *OptimizedStopTimeProducer) fillStopTimeData(
	stopTime *model.StopTime,
	passingTime *model.TimetabledPassingTime,
	trip *model.Trip,
	sequence int,
	shape *model.Shape,
	lineName string,
) error {

	stopTime.TripID = trip.TripID
	stopTime.StopSequence = sequence

	// Convert times efficiently
	if passingTime.ArrivalTime != "" {
		if arrTime, err := time.Parse("15:04:05", passingTime.ArrivalTime); err == nil {
			stopTime.ArrivalTime = p.formatGTFSTime(arrTime, passingTime.DayOffset)
		}
	}

	if passingTime.DepartureTime != "" {
		if depTime, err := time.Parse("15:04:05", passingTime.DepartureTime); err == nil {
			stopTime.DepartureTime = p.formatGTFSTime(depTime, passingTime.DayOffset)
		}
	} else if passingTime.ArrivalTime != "" {
		// Use arrival time as departure time if departure is not specified
		stopTime.DepartureTime = stopTime.ArrivalTime
	}

	// Get stop information efficiently
	// TODO: Implement proper stop resolution
	stopTime.StopID = fmt.Sprintf("stop_%s", passingTime.PointInJourneyPatternRef)

	// Handle pickup and drop-off types for European transit patterns
	p.setEuropeanPickupDropoffTypes(stopTime, sequence, shape != nil, lineName)

	// Calculate shape distance if shape is available
	if shape != nil && sequence > 1 {
		stopTime.ShapeDistTraveled = p.calculateShapeDistance(shape, sequence-1)
	}

	// Set timepoint
	stopTime.Timepoint = "1" // Exact times

	return nil
}

// formatGTFSTime formats time with day offset handling, optimized for performance
func (p *OptimizedStopTimeProducer) formatGTFSTime(t time.Time, dayOffset int) string {
	// Add day offset
	adjustedTime := t.Add(time.Duration(dayOffset) * 24 * time.Hour)

	// Get hours, minutes, seconds
	hour := adjustedTime.Hour() + (dayOffset * 24)
	minute := adjustedTime.Minute()
	second := adjustedTime.Second()

	// Use faster string formatting for common cases
	if hour < 10 {
		if minute < 10 {
			if second < 10 {
				return fmt.Sprintf("0%d:0%d:0%d", hour, minute, second)
			}
			return fmt.Sprintf("0%d:0%d:%d", hour, minute, second)
		}
		if second < 10 {
			return fmt.Sprintf("0%d:%d:0%d", hour, minute, second)
		}
		return fmt.Sprintf("0%d:%d:%d", hour, minute, second)
	}

	if minute < 10 {
		if second < 10 {
			return fmt.Sprintf("%d:0%d:0%d", hour, minute, second)
		}
		return fmt.Sprintf("%d:0%d:%d", hour, minute, second)
	}

	if second < 10 {
		return fmt.Sprintf("%d:%d:0%d", hour, minute, second)
	}

	return fmt.Sprintf("%d:%d:%d", hour, minute, second)
}

// setEuropeanPickupDropoffTypes sets pickup/dropoff types optimized for memory access
func (p *OptimizedStopTimeProducer) setEuropeanPickupDropoffTypes(stopTime *model.StopTime, sequence int, hasShapes bool, lineName string) {
	// Default values
	stopTime.PickupType = "0"  // Regular pickup
	stopTime.DropOffType = "0" // Regular drop-off

	// European transit specific rules - optimized with early returns
	if sequence == 1 {
		// First stop - no pickup typically
		stopTime.PickupType = "1" // No pickup available
		return
	}

	// Check line type for specific rules (using string prefix for speed)
	if len(lineName) > 0 {
		switch lineName[0] {
		case 'N', 'n': // Night services
			if sequence == 1 {
				stopTime.PickupType = "1" // No pickup at first stop
			}
		case 'E', 'e': // Express services
			// Express services may have limited stops
			stopTime.PickupType = "0"
			stopTime.DropOffType = "0"
		}
	}
}

// calculateShapeDistance calculates distance along shape, optimized for performance
func (p *OptimizedStopTimeProducer) calculateShapeDistance(shape *model.Shape, stopIndex int) float64 {
	if shape == nil {
		return 0.0
	}

	// Use shape distance traveled if available
	if shape.ShapeDistTraveled > 0 {
		return shape.ShapeDistTraveled
	}

	// Simple estimation based on sequence for performance
	return float64(stopIndex) * 500.0 // Assume 500m between stops
}

// GetMemoryStats returns memory statistics for the producer
func (p *OptimizedStopTimeProducer) GetMemoryStats() memory.MemoryStats {
	return p.memoryManager.GetMemoryStats()
}

// SetMemoryLimit sets the memory limit for the producer
func (p *OptimizedStopTimeProducer) SetMemoryLimit(limitMB uint64) {
	p.memoryManager.SetMemoryLimit(limitMB)
}

// ForceMemoryCleanup triggers garbage collection
func (p *OptimizedStopTimeProducer) ForceMemoryCleanup() {
	p.memoryManager.ForceGC()
}
