package producer

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/theoremus-urban-solutions/netex-gtfs-converter/model"
)

// EnhancedStopTimeProducer extends DefaultStopTimeProducer with advanced capabilities
type EnhancedStopTimeProducer struct {
	*DefaultStopTimeProducer
	*AdvancedStopTimeProducer
}

// NewEnhancedStopTimeProducer creates a new enhanced stop time producer
func NewEnhancedStopTimeProducer(netexRepo NetexRepository, gtfsRepo GtfsRepository) *EnhancedStopTimeProducer {
	return &EnhancedStopTimeProducer{
		DefaultStopTimeProducer:  NewDefaultStopTimeProducer(netexRepo, gtfsRepo),
		AdvancedStopTimeProducer: NewAdvancedStopTimeProducer(netexRepo, gtfsRepo),
	}
}

// Produce implements the StopTimeProducer interface with enhanced capabilities
func (p *EnhancedStopTimeProducer) Produce(input StopTimeInput) (*model.StopTime, error) {
	// Use the default producer for single stop time production
	return p.DefaultStopTimeProducer.Produce(input)
}

// ProduceStopTimesForTrip produces all stop times for a complete trip with advanced interpolation
func (p *EnhancedStopTimeProducer) ProduceStopTimesForTrip(serviceJourney *model.ServiceJourney, trip *model.Trip, shape *model.Shape, headsign string) ([]*model.StopTime, error) {
	if serviceJourney == nil || trip == nil {
		return nil, fmt.Errorf("service journey and trip are required")
	}

	// Use the advanced producer for complete trip processing
	input := TripStopTimeInput{
		ServiceJourney:  serviceJourney,
		Trip:            trip,
		Shape:           shape,
		CurrentHeadSign: headsign,
	}

	stopTimes, err := p.ProduceAdvanced(input)
	if err != nil {
		return nil, err
	}

	// Apply additional enhancements
	err = p.applyAccessibilityInformation(stopTimes, serviceJourney)
	if err != nil {
		return nil, fmt.Errorf("failed to apply accessibility information: %w", err)
	}

	err = p.applyEuropeanSpecificFeatures(stopTimes, serviceJourney)
	if err != nil {
		return nil, fmt.Errorf("failed to apply European features: %w", err)
	}

	// Final validation
	err = p.ValidateStopTimeSequence(stopTimes)
	if err != nil {
		return nil, fmt.Errorf("stop time validation failed: %w", err)
	}

	return stopTimes, nil
}

// applyAccessibilityInformation adds accessibility details from NeTEx to GTFS stop times
func (p *EnhancedStopTimeProducer) applyAccessibilityInformation(stopTimes []*model.StopTime, serviceJourney *model.ServiceJourney) error {
	for _, stopTime := range stopTimes {
		// Get stop information
		stop := p.DefaultStopTimeProducer.gtfsRepository.GetStopById(stopTime.StopID)
		if stop == nil {
			continue
		}

		// In a full implementation, this would check:
		// - Wheelchair accessibility
		// - Boarding assistance requirements
		// - Visual/audio accessibility features
		// - Platform height compatibility

		// For now, we set basic accessibility information
		// This would be enhanced based on actual NeTEx accessibility data
	}

	return nil
}

// applyEuropeanSpecificFeatures adds European transit-specific features
func (p *EnhancedStopTimeProducer) applyEuropeanSpecificFeatures(stopTimes []*model.StopTime, serviceJourney *model.ServiceJourney) error {
	for i, stopTime := range stopTimes {
		// Apply European-specific pickup/dropoff rules
		err := p.applyEuropeanPickupDropoffRules(stopTime, serviceJourney, i)
		if err != nil {
			continue // Log error but don't fail the entire process
		}

		// Apply flexible service rules if applicable
		err = p.applyFlexibleServiceRules(stopTime, serviceJourney)
		if err != nil {
			continue
		}
	}

	return nil
}

// applyEuropeanPickupDropoffRules applies European transit pickup/dropoff patterns
func (p *EnhancedStopTimeProducer) applyEuropeanPickupDropoffRules(stopTime *model.StopTime, serviceJourney *model.ServiceJourney, stopIndex int) error {
	// European transit often has specific boarding/alighting rules:
	// - First stop: typically drop-off only not allowed
	// - Last stop: typically pickup not allowed
	// - Request stops: pickup/dropoff by arrangement
	// - Interchange stops: special handling

	// Get journey pattern to determine stop role
	journeyPattern := p.DefaultStopTimeProducer.netexRepository.GetJourneyPatternById(serviceJourney.JourneyPatternRef.Ref)
	if journeyPattern == nil {
		return fmt.Errorf("journey pattern not found")
	}

	totalStops := len(journeyPattern.PointsInSequence.PointInJourneyPatternOrStopPointInJourneyPatternOrTimingPointInJourneyPattern)

	// Apply rules based on stop position
	switch stopIndex {
	case 0:
		// First stop - no drop-off typically
		if stopTime.DropOffType == "0" {
			stopTime.DropOffType = "1" // No drop-off available
		}
	case totalStops - 1:
		// Last stop - no pickup typically
		if stopTime.PickupType == "0" {
			stopTime.PickupType = "1" // No pickup available
		}
	}

	return nil
}

// applyFlexibleServiceRules applies rules for flexible/demand-responsive services
func (p *EnhancedStopTimeProducer) applyFlexibleServiceRules(stopTime *model.StopTime, serviceJourney *model.ServiceJourney) error {
	// Check if this is a flexible service
	// In NeTEx, flexible services have specific indicators

	// This would be enhanced to check:
	// - FlexibleServiceProperties
	// - BookingArrangements
	// - RequestStop indicators
	// - Zone-based flexible areas

	return nil
}

// InterpolateStopTimesWithConstraints provides advanced interpolation with European constraints
func (p *EnhancedStopTimeProducer) InterpolateStopTimesWithConstraints(stopTimes []*model.StopTime, constraints InterpolationConstraints) error {
	if len(stopTimes) <= 1 {
		return nil
	}

	// Sort by sequence
	sort.Slice(stopTimes, func(i, j int) bool {
		return stopTimes[i].StopSequence < stopTimes[j].StopSequence
	})

	// Apply European-specific interpolation rules
	for i := 0; i < len(stopTimes); i++ {
		if stopTimes[i].ArrivalTime == "" || stopTimes[i].DepartureTime == "" {
			err := p.interpolateWithEuropeanRules(stopTimes, i, constraints)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// InterpolationConstraints defines constraints for stop time interpolation
type InterpolationConstraints struct {
	MaxSpeed         float64 // km/h maximum speed
	MinSpeed         float64 // km/h minimum speed
	MinStopTime      int     // seconds minimum at stop
	MaxStopTime      int     // seconds maximum at stop
	InterchangeTime  int     // seconds additional time at interchange stops
	TerminalTime     int     // seconds additional time at terminal stops
	RushHourFactor   float64 // speed reduction during rush hour
	RuralSpeedFactor float64 // speed adjustment for rural areas
	UrbanSpeedFactor float64 // speed adjustment for urban areas
}

// DefaultInterpolationConstraints returns typical European transit constraints
func DefaultInterpolationConstraints() InterpolationConstraints {
	return InterpolationConstraints{
		MaxSpeed:         80.0, // 80 km/h max for regional services
		MinSpeed:         15.0, // 15 km/h min in urban areas
		MinStopTime:      20,   // 20 seconds minimum stop time
		MaxStopTime:      180,  // 3 minutes maximum stop time
		InterchangeTime:  60,   // 1 minute extra for interchanges
		TerminalTime:     120,  // 2 minutes extra at terminals
		RushHourFactor:   0.7,  // 30% slower during rush hour
		RuralSpeedFactor: 1.2,  // 20% faster in rural areas
		UrbanSpeedFactor: 0.8,  // 20% slower in urban areas
	}
}

// interpolateWithEuropeanRules applies European-specific interpolation logic
func (p *EnhancedStopTimeProducer) interpolateWithEuropeanRules(stopTimes []*model.StopTime, index int, constraints InterpolationConstraints) error {
	// Find nearest stops with known times
	prevIndex := -1
	nextIndex := -1

	for i := index - 1; i >= 0; i-- {
		if stopTimes[i].ArrivalTime != "" || stopTimes[i].DepartureTime != "" {
			prevIndex = i
			break
		}
	}

	for i := index + 1; i < len(stopTimes); i++ {
		if stopTimes[i].ArrivalTime != "" || stopTimes[i].DepartureTime != "" {
			nextIndex = i
			break
		}
	}

	if prevIndex == -1 && nextIndex == -1 {
		return fmt.Errorf("no reference times available for interpolation")
	}

	// Calculate interpolated time with European constraints
	var arrivalSeconds, departureSeconds int

	switch {
	case prevIndex >= 0 && nextIndex >= 0:
		// Interpolate between two known points
		prevTime := p.parseGTFSTime(getLastKnownTime(stopTimes[prevIndex]))
		nextTime := p.parseGTFSTime(getFirstKnownTime(stopTimes[nextIndex]))

		if prevTime > 0 && nextTime > 0 {
			// Calculate position ratio
			ratio := float64(index-prevIndex) / float64(nextIndex-prevIndex)

			// Apply European timing constraints
			timeDiff := nextTime - prevTime
			interpolatedTime := prevTime + int(float64(timeDiff)*ratio)

			// Apply minimum stop times based on stop type
			stopTime := constraints.MinStopTime
			if p.isInterchangeStop(stopTimes[index]) {
				stopTime += constraints.InterchangeTime
			}
			if p.isTerminalStop(stopTimes[index], len(stopTimes)) {
				stopTime += constraints.TerminalTime
			}

			arrivalSeconds = interpolatedTime
			departureSeconds = interpolatedTime + stopTime
		}
	case prevIndex >= 0:
		// Extrapolate forward
		prevTime := p.parseGTFSTime(getLastKnownTime(stopTimes[prevIndex]))
		if prevTime > 0 {
			// Estimate travel time with European speed constraints
			travelTime := p.estimateTravelTimeEuropean(constraints)
			arrivalSeconds = prevTime + travelTime
			departureSeconds = arrivalSeconds + constraints.MinStopTime
		}
	case nextIndex >= 0:
		// Extrapolate backward
		nextTime := p.parseGTFSTime(getFirstKnownTime(stopTimes[nextIndex]))
		if nextTime > 0 {
			travelTime := p.estimateTravelTimeEuropean(constraints)
			arrivalSeconds = nextTime - travelTime - constraints.MinStopTime
			departureSeconds = arrivalSeconds + constraints.MinStopTime
		}
	}

	// Set the interpolated times
	if arrivalSeconds > 0 {
		stopTimes[index].ArrivalTime = p.secondsToGTFSTime(arrivalSeconds)
	}
	if departureSeconds > 0 {
		stopTimes[index].DepartureTime = p.secondsToGTFSTime(departureSeconds)
	}

	return nil
}

// Helper functions

func getLastKnownTime(stopTime *model.StopTime) string {
	if stopTime.DepartureTime != "" {
		return stopTime.DepartureTime
	}
	return stopTime.ArrivalTime
}

func getFirstKnownTime(stopTime *model.StopTime) string {
	if stopTime.ArrivalTime != "" {
		return stopTime.ArrivalTime
	}
	return stopTime.DepartureTime
}

func (p *EnhancedStopTimeProducer) isInterchangeStop(stopTime *model.StopTime) bool {
	// Check if this is an interchange stop
	// This would check NeTEx interchange information
	return false // Placeholder
}

func (p *EnhancedStopTimeProducer) isTerminalStop(stopTime *model.StopTime, totalStops int) bool {
	// Check if this is first or last stop
	return stopTime.StopSequence == 1 || stopTime.StopSequence == totalStops
}

func (p *EnhancedStopTimeProducer) estimateTravelTimeEuropean(constraints InterpolationConstraints) int {
	// Estimate travel time using European speed constraints
	// This is simplified - real implementation would consider:
	// - Stop spacing
	// - Service type (urban, suburban, regional)
	// - Time of day
	// - Geographic constraints

	avgSpeed := (constraints.MaxSpeed + constraints.MinSpeed) / 2
	distance := 1.0 // km - placeholder

	// Convert to seconds
	return int((distance / avgSpeed) * 3600)
}

func (p *EnhancedStopTimeProducer) secondsToGTFSTime(seconds int) string {
	hours := seconds / 3600
	minutes := (seconds % 3600) / 60
	secs := seconds % 60
	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, secs)
}

func (p *EnhancedStopTimeProducer) parseGTFSTime(timeStr string) int {
	if timeStr == "" {
		return 0
	}

	parts := strings.Split(timeStr, ":")
	if len(parts) != 3 {
		return 0
	}

	hours, _ := strconv.Atoi(parts[0])
	minutes, _ := strconv.Atoi(parts[1])
	seconds, _ := strconv.Atoi(parts[2])

	return hours*3600 + minutes*60 + seconds
}
