package producer

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/theoremus-urban-solutions/netex-gtfs-converter/model"
)

// AdvancedStopTimeProducer implements advanced stop time calculations and interpolation
type AdvancedStopTimeProducer struct {
	netexRepository NetexRepository
	gtfsRepository  GtfsRepository

	// Configuration
	defaultTravelSpeed  float64 // km/h for interpolation
	minStopDuration     int     // minimum seconds at stop
	maxInterpolationGap int     // maximum gap in seconds to interpolate
}

// NewAdvancedStopTimeProducer creates a new advanced stop time producer
func NewAdvancedStopTimeProducer(netexRepo NetexRepository, gtfsRepo GtfsRepository) *AdvancedStopTimeProducer {
	return &AdvancedStopTimeProducer{
		netexRepository:     netexRepo,
		gtfsRepository:      gtfsRepo,
		defaultTravelSpeed:  25.0, // 25 km/h default speed
		minStopDuration:     30,   // 30 seconds minimum at stop
		maxInterpolationGap: 3600, // 1 hour maximum gap to interpolate
	}
}

// SetDefaultTravelSpeed sets the default travel speed for interpolation (km/h)
func (p *AdvancedStopTimeProducer) SetDefaultTravelSpeed(speed float64) {
	p.defaultTravelSpeed = speed
}

// SetMinStopDuration sets the minimum duration at stops (seconds)
func (p *AdvancedStopTimeProducer) SetMinStopDuration(duration int) {
	p.minStopDuration = duration
}

// ProduceAdvanced produces stop times with advanced calculations and interpolation
func (p *AdvancedStopTimeProducer) ProduceAdvanced(tripInput TripStopTimeInput) ([]*model.StopTime, error) {
	if tripInput.ServiceJourney == nil {
		return nil, fmt.Errorf("service journey is required")
	}

	// Get journey pattern for stop sequence
	journeyPattern := p.netexRepository.GetJourneyPatternById(tripInput.ServiceJourney.JourneyPatternRef.Ref)
	if journeyPattern == nil {
		return nil, fmt.Errorf("journey pattern not found: %s", tripInput.ServiceJourney.JourneyPatternRef.Ref)
	}

	// Build stop sequence from journey pattern
	stopSequence, err := p.buildStopSequence(journeyPattern, tripInput.ServiceJourney)
	if err != nil {
		return nil, fmt.Errorf("failed to build stop sequence: %w", err)
	}

	// Apply timetabled passing times
	err = p.applyTimetabledTimes(stopSequence, tripInput.ServiceJourney.PassingTimes)
	if err != nil {
		return nil, fmt.Errorf("failed to apply timetabled times: %w", err)
	}

	// Interpolate missing times
	err = p.interpolateMissingTimes(stopSequence, tripInput.Shape)
	if err != nil {
		return nil, fmt.Errorf("failed to interpolate missing times: %w", err)
	}

	// Calculate shape distances
	if tripInput.Shape != nil {
		err = p.calculateShapeDistances(stopSequence, tripInput.Shape)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate shape distances: %w", err)
		}
	}

	// Convert to GTFS stop times
	stopTimes := make([]*model.StopTime, len(stopSequence))
	for i, stop := range stopSequence {
		stopTimes[i] = p.convertToGtfsStopTime(stop, tripInput.Trip, tripInput.CurrentHeadSign)
	}

	return stopTimes, nil
}

// TripStopTimeInput contains all data needed for advanced stop time production
type TripStopTimeInput struct {
	ServiceJourney  *model.ServiceJourney
	Trip            *model.Trip
	Shape           *model.Shape
	CurrentHeadSign string
}

// StopInSequence represents a stop with calculated timing information
type StopInSequence struct {
	StopID            string
	StopSequence      int
	ArrivalTime       *time.Time
	DepartureTime     *time.Time
	ArrivalSeconds    int
	DepartureSeconds  int
	ShapeDistTraveled float64
	PickupType        string
	DropOffType       string
	StopHeadsign      string
	HasTimetabledTime bool
	IsInterpolated    bool

	// NeTEx references for debugging
	PointInJourneyPatternRef string
	ScheduledStopPointRef    string
}

// buildStopSequence creates the basic stop sequence from journey pattern
func (p *AdvancedStopTimeProducer) buildStopSequence(pattern *model.JourneyPattern, journey *model.ServiceJourney) ([]*StopInSequence, error) {
	if pattern.PointsInSequence == nil {
		return nil, fmt.Errorf("no points in sequence found")
	}

	var sequence []*StopInSequence
	stopCounter := 1

	for _, pointInterface := range pattern.PointsInSequence.PointInJourneyPatternOrStopPointInJourneyPatternOrTimingPointInJourneyPattern {
		if stopPoint, ok := pointInterface.(*model.StopPointInJourneyPattern); ok {
			// Resolve stop ID
			stopID := p.resolveStopIDAdvanced(stopPoint)
			if stopID == "" {
				continue // Skip unresolvable stops
			}

			stop := &StopInSequence{
				StopID:                   stopID,
				StopSequence:             stopCounter,
				PointInJourneyPatternRef: stopPoint.ID,
				ScheduledStopPointRef:    stopPoint.ScheduledStopPointRef,
				PickupType:               "0", // Default regular pickup
				DropOffType:              "0", // Default regular drop-off
				HasTimetabledTime:        false,
				IsInterpolated:           false,
			}

			// Apply boarding and alighting restrictions from stop point
			p.applyBoardingRestrictions(stop, stopPoint)

			sequence = append(sequence, stop)
			stopCounter++
		}
	}

	if len(sequence) == 0 {
		return nil, fmt.Errorf("no valid stops found in journey pattern")
	}

	return sequence, nil
}

// applyTimetabledTimes applies actual timetabled times to the sequence
func (p *AdvancedStopTimeProducer) applyTimetabledTimes(sequence []*StopInSequence, passingTimes *model.PassingTimes) error {
	if passingTimes == nil {
		return nil
	}

	// Create mapping from point reference to timetabled passing time
	timeMap := make(map[string]*model.TimetabledPassingTime)
	for i, tpt := range passingTimes.TimetabledPassingTime {
		if tpt.PointInJourneyPatternRef != "" {
			timeMap[tpt.PointInJourneyPatternRef] = &passingTimes.TimetabledPassingTime[i]
		}
	}

	// Apply times to sequence
	for _, stop := range sequence {
		if tpt, exists := timeMap[stop.PointInJourneyPatternRef]; exists {
			stop.HasTimetabledTime = true

			if tpt.ArrivalTime != "" {
				// Parse time string and create time.Time for further processing
				if arrTime, err := p.parseTimeString(tpt.ArrivalTime); err == nil {
					stop.ArrivalTime = &arrTime
					stop.ArrivalSeconds = p.timeToSeconds(arrTime, tpt.DayOffset)
				}
			}

			if tpt.DepartureTime != "" {
				// Parse time string and create time.Time for further processing
				if depTime, err := p.parseTimeString(tpt.DepartureTime); err == nil {
					stop.DepartureTime = &depTime
					stop.DepartureSeconds = p.timeToSeconds(depTime, tpt.DayOffset)
				}
			}

			// If only one time available, use for both
			if stop.ArrivalTime == nil && stop.DepartureTime != nil {
				stop.ArrivalTime = stop.DepartureTime
				stop.ArrivalSeconds = stop.DepartureSeconds
			}
			if stop.DepartureTime == nil && stop.ArrivalTime != nil {
				stop.DepartureTime = stop.ArrivalTime
				stop.DepartureSeconds = stop.ArrivalSeconds
			}
		}
	}

	return nil
}

// interpolateMissingTimes fills in missing arrival/departure times using interpolation
func (p *AdvancedStopTimeProducer) interpolateMissingTimes(sequence []*StopInSequence, shape *model.Shape) error {
	if len(sequence) <= 1 {
		return nil
	}

	// Find segments that need interpolation
	for i := 0; i < len(sequence); i++ {
		if sequence[i].HasTimetabledTime {
			continue
		}

		// Find the next stop with timetabled time
		nextTimetabledIndex := -1
		for j := i + 1; j < len(sequence); j++ {
			if sequence[j].HasTimetabledTime {
				nextTimetabledIndex = j
				break
			}
		}

		// Find the previous stop with timetabled time
		prevTimetabledIndex := -1
		for j := i - 1; j >= 0; j-- {
			if sequence[j].HasTimetabledTime {
				prevTimetabledIndex = j
				break
			}
		}

		// Interpolate based on available anchor points
		if prevTimetabledIndex >= 0 && nextTimetabledIndex >= 0 {
			// Interpolate between two known times
			err := p.interpolateBetweenTimes(sequence, prevTimetabledIndex, nextTimetabledIndex, shape)
			if err != nil {
				return err
			}
		} else if prevTimetabledIndex >= 0 {
			// Extrapolate forward from previous time
			err := p.extrapolateForward(sequence, prevTimetabledIndex, i, shape)
			if err != nil {
				return err
			}
		} else if nextTimetabledIndex >= 0 {
			// Extrapolate backward from next time
			err := p.extrapolateBackward(sequence, i, nextTimetabledIndex, shape)
			if err != nil {
				return err
			}
		}

		if sequence[i].ArrivalSeconds > 0 {
			sequence[i].IsInterpolated = true
		}
	}

	return nil
}

// interpolateBetweenTimes interpolates times between two known time points
func (p *AdvancedStopTimeProducer) interpolateBetweenTimes(sequence []*StopInSequence, startIdx, endIdx int, shape *model.Shape) error {
	startTime := sequence[startIdx].DepartureSeconds
	endTime := sequence[endIdx].ArrivalSeconds
	totalDuration := endTime - startTime

	if totalDuration <= 0 || totalDuration > p.maxInterpolationGap {
		return nil // Skip invalid or too large gaps
	}

	// Calculate total distance if shape is available
	var totalDistance float64
	var distances []float64

	if shape != nil {
		distances = make([]float64, endIdx-startIdx+1)
		for i := startIdx; i <= endIdx; i++ {
			if i == startIdx {
				distances[i-startIdx] = 0
			} else {
				// Calculate distance from previous stop
				dist := p.estimateDistanceBetweenStops(sequence[i-1], sequence[i], shape)
				distances[i-startIdx] = distances[i-1-startIdx] + dist
			}
		}
		totalDistance = distances[len(distances)-1]
	}

	// Interpolate times for stops between start and end
	for i := startIdx + 1; i < endIdx; i++ {
		var ratio float64

		if totalDistance > 0 && len(distances) > i-startIdx {
			// Distance-based interpolation
			ratio = distances[i-startIdx] / totalDistance
		} else {
			// Linear interpolation based on stop sequence
			ratio = float64(i-startIdx) / float64(endIdx-startIdx)
		}

		// Calculate interpolated time
		interpolatedSeconds := startTime + int(float64(totalDuration)*ratio)

		// Add minimum stop duration
		arrivalSeconds := interpolatedSeconds
		departureSeconds := interpolatedSeconds + p.minStopDuration

		// Convert to time objects
		arrivalTime := p.secondsToTime(arrivalSeconds)
		departureTime := p.secondsToTime(departureSeconds)

		sequence[i].ArrivalSeconds = arrivalSeconds
		sequence[i].DepartureSeconds = departureSeconds
		sequence[i].ArrivalTime = &arrivalTime
		sequence[i].DepartureTime = &departureTime
		sequence[i].IsInterpolated = true
	}

	return nil
}

// extrapolateForward estimates times forward from a known time point
func (p *AdvancedStopTimeProducer) extrapolateForward(sequence []*StopInSequence, fromIdx, toIdx int, shape *model.Shape) error {
	if fromIdx >= toIdx {
		return nil
	}

	baseTime := sequence[fromIdx].DepartureSeconds

	for i := fromIdx + 1; i <= toIdx; i++ {
		// Estimate travel time based on distance and default speed
		distance := p.estimateDistanceBetweenStops(sequence[i-1], sequence[i], shape)
		travelTimeSeconds := int((distance / p.defaultTravelSpeed) * 3600) // Convert km/h to seconds

		if travelTimeSeconds < 60 {
			travelTimeSeconds = 60 // Minimum 1 minute travel time
		}

		arrivalSeconds := baseTime + travelTimeSeconds
		departureSeconds := arrivalSeconds + p.minStopDuration

		arrivalTime := p.secondsToTime(arrivalSeconds)
		departureTime := p.secondsToTime(departureSeconds)

		sequence[i].ArrivalSeconds = arrivalSeconds
		sequence[i].DepartureSeconds = departureSeconds
		sequence[i].ArrivalTime = &arrivalTime
		sequence[i].DepartureTime = &departureTime
		sequence[i].IsInterpolated = true

		baseTime = departureSeconds
	}

	return nil
}

// extrapolateBackward estimates times backward from a known time point
func (p *AdvancedStopTimeProducer) extrapolateBackward(sequence []*StopInSequence, fromIdx, toIdx int, shape *model.Shape) error {
	if fromIdx >= toIdx {
		return nil
	}

	baseTime := sequence[toIdx].ArrivalSeconds

	for i := toIdx - 1; i >= fromIdx; i-- {
		// Estimate travel time based on distance and default speed
		distance := p.estimateDistanceBetweenStops(sequence[i], sequence[i+1], shape)
		travelTimeSeconds := int((distance / p.defaultTravelSpeed) * 3600)

		if travelTimeSeconds < 60 {
			travelTimeSeconds = 60
		}

		departureSeconds := baseTime - travelTimeSeconds
		arrivalSeconds := departureSeconds - p.minStopDuration

		arrivalTime := p.secondsToTime(arrivalSeconds)
		departureTime := p.secondsToTime(departureSeconds)

		sequence[i].ArrivalSeconds = arrivalSeconds
		sequence[i].DepartureSeconds = departureSeconds
		sequence[i].ArrivalTime = &arrivalTime
		sequence[i].DepartureTime = &departureTime
		sequence[i].IsInterpolated = true

		baseTime = arrivalSeconds
	}

	return nil
}

// calculateShapeDistances calculates the distance traveled along the shape for each stop
func (p *AdvancedStopTimeProducer) calculateShapeDistances(sequence []*StopInSequence, shape *model.Shape) error {
	if shape == nil {
		return nil
	}

	totalDistance := 0.0

	for i, stop := range sequence {
		if i == 0 {
			stop.ShapeDistTraveled = 0.0
		} else {
			// For now, estimate distance between consecutive stops
			// In a real implementation, this would project stops onto the shape
			// and calculate actual distance along the shape
			distance := p.estimateDistanceBetweenStops(sequence[i-1], stop, shape)
			totalDistance += distance
			stop.ShapeDistTraveled = totalDistance
		}
	}

	return nil
}

// Helper methods

func (p *AdvancedStopTimeProducer) resolveStopIDAdvanced(stopPoint *model.StopPointInJourneyPattern) string {
	if stopPoint == nil {
		return ""
	}

	// Try scheduled stop point reference
	if stopPoint.ScheduledStopPointRef != "" {
		if ssp := p.netexRepository.GetScheduledStopPointById(stopPoint.ScheduledStopPointRef); ssp != nil {
			if ssp.QuayRef != "" {
				// Verify quay exists
				if p.netexRepository.GetQuayById(ssp.QuayRef) != nil {
					return ssp.QuayRef
				}
			}
			if ssp.StopPlaceRef != "" {
				// Verify stop place exists
				if stopPlace := p.netexRepository.GetStopPlaceByQuayId(ssp.StopPlaceRef); stopPlace != nil {
					return ssp.StopPlaceRef
				}
			}
			// Fallback to scheduled stop point ID
			return ssp.ID
		}
	}

	// Final fallback
	return stopPoint.ID
}

func (p *AdvancedStopTimeProducer) applyBoardingRestrictions(stop *StopInSequence, stopPoint *model.StopPointInJourneyPattern) {
	// Check for boarding and alighting restrictions
	// This would be enhanced based on NeTEx boarding/alighting conditions

	// For now, keep defaults unless specific restrictions are found
	// In a full implementation, this would check:
	// - stopPoint.BoardingUse
	// - stopPoint.AlightingUse
	// - Time-based restrictions
	// - Accessibility restrictions
}

func (p *AdvancedStopTimeProducer) timeToSeconds(t time.Time, dayOffset int) int {
	seconds := t.Hour()*3600 + t.Minute()*60 + t.Second()
	return seconds + (dayOffset * 24 * 3600)
}

func (p *AdvancedStopTimeProducer) secondsToTime(seconds int) time.Time {
	hours := seconds / 3600
	minutes := (seconds % 3600) / 60
	secs := seconds % 60

	// Handle day overflow for GTFS (24:00:00+)
	day := hours / 24
	hours = hours % 24

	baseDate := time.Date(2000, 1, 1+day, hours, minutes, secs, 0, time.UTC)
	return baseDate
}

func (p *AdvancedStopTimeProducer) formatTimeForGTFS(t time.Time, dayOffset int) string {
	hours := t.Hour() + (dayOffset * 24)
	return fmt.Sprintf("%02d:%02d:%02d", hours, t.Minute(), t.Second())
}

func (p *AdvancedStopTimeProducer) estimateDistanceBetweenStops(stop1, stop2 *StopInSequence, shape *model.Shape) float64 {
	// In a real implementation, this would:
	// 1. Get the actual coordinates of the stops
	// 2. Project them onto the shape
	// 3. Calculate the distance along the shape

	// For now, return a default estimate based on typical stop spacing
	return 0.5 // 500 meters default spacing
}

func (p *AdvancedStopTimeProducer) convertToGtfsStopTime(stop *StopInSequence, trip *model.Trip, headsign string) *model.StopTime {
	gtfsStop := &model.StopTime{
		TripID:            trip.TripID,
		StopID:            stop.StopID,
		StopSequence:      stop.StopSequence,
		PickupType:        stop.PickupType,
		DropOffType:       stop.DropOffType,
		ShapeDistTraveled: stop.ShapeDistTraveled,
	}

	// Format times
	if stop.ArrivalTime != nil {
		gtfsStop.ArrivalTime = p.formatTimeForGTFS(*stop.ArrivalTime, 0)
	}
	if stop.DepartureTime != nil {
		gtfsStop.DepartureTime = p.formatTimeForGTFS(*stop.DepartureTime, 0)
	}

	// Set headsign
	if stop.StopHeadsign != "" {
		gtfsStop.StopHeadsign = stop.StopHeadsign
	} else if headsign != "" {
		gtfsStop.StopHeadsign = headsign
	}

	return gtfsStop
}

// ValidateStopTimeSequence validates and potentially corrects stop time sequences
func (p *AdvancedStopTimeProducer) ValidateStopTimeSequence(stopTimes []*model.StopTime) error {
	if len(stopTimes) <= 1 {
		return nil
	}

	// Sort by stop sequence to ensure order
	sort.Slice(stopTimes, func(i, j int) bool {
		return stopTimes[i].StopSequence < stopTimes[j].StopSequence
	})

	// Validate time progression
	for i := 1; i < len(stopTimes); i++ {
		prevTime := p.parseGTFSTime(stopTimes[i-1].DepartureTime)
		currTime := p.parseGTFSTime(stopTimes[i].ArrivalTime)

		if prevTime > 0 && currTime > 0 && currTime < prevTime {
			return fmt.Errorf("stop time sequence violation at stop %d: arrival %s is before previous departure %s",
				stopTimes[i].StopSequence, stopTimes[i].ArrivalTime, stopTimes[i-1].DepartureTime)
		}
	}

	return nil
}

func (p *AdvancedStopTimeProducer) parseGTFSTime(timeStr string) int {
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

// parseTimeString parses a time string like "10:25:00" into a time.Time
func (p *AdvancedStopTimeProducer) parseTimeString(timeStr string) (time.Time, error) {
	// Parse time in format "HH:MM:SS"
	return time.Parse("15:04:05", timeStr)
}
