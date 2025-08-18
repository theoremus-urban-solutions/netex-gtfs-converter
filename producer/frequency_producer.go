package producer

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/theoremus-urban-solutions/netex-gtfs-converter/model"
)

// DefaultFrequencyProducer implements frequency conversion
type DefaultFrequencyProducer struct {
	netexRepo NetexRepository
	gtfsRepo  GtfsRepository
}

// NewDefaultFrequencyProducer creates a new frequency producer
func NewDefaultFrequencyProducer(netexRepo NetexRepository, gtfsRepo GtfsRepository) *DefaultFrequencyProducer {
	return &DefaultFrequencyProducer{
		netexRepo: netexRepo,
		gtfsRepo:  gtfsRepo,
	}
}

// ProduceFromHeadwayJourneyGroup converts a NeTEx HeadwayJourneyGroup to GTFS frequencies
func (p *DefaultFrequencyProducer) ProduceFromHeadwayJourneyGroup(group *model.HeadwayJourneyGroup) ([]*model.Frequency, error) {
	if group == nil {
		return nil, nil
	}

	var frequencies []*model.Frequency

	// Get the journey pattern to understand the route structure
	journeyPattern := p.netexRepo.GetJourneyPatternById(group.JourneyPatternRef)
	if journeyPattern == nil {
		return nil, fmt.Errorf("journey pattern not found: %s", group.JourneyPatternRef)
	}

	// Create a representative trip ID for this frequency group
	tripID := fmt.Sprintf("freq_trip_%s", group.ID)

	// Convert headway interval to seconds
	headwaySeconds, err := p.convertDurationToSeconds(group.ScheduledHeadwayInterval)
	if err != nil {
		return nil, fmt.Errorf("invalid headway interval: %w", err)
	}

	// Parse start and end times
	startTime, err := p.convertTimeToSeconds(group.FirstDepartureTime)
	if err != nil {
		return nil, fmt.Errorf("invalid first departure time: %w", err)
	}

	endTime, err := p.convertTimeToSeconds(group.LastDepartureTime)
	if err != nil {
		return nil, fmt.Errorf("invalid last departure time: %w", err)
	}

	// Create GTFS frequency entry
	frequency := &model.Frequency{
		TripID:      tripID,
		StartTime:   p.formatTimeForGTFS(startTime),
		EndTime:     p.formatTimeForGTFS(endTime),
		HeadwaySecs: headwaySeconds,
		ExactTimes:  "0", // Frequency-based, not exact times
	}

	frequencies = append(frequencies, frequency)

	return frequencies, nil
}

// ProduceFromTimeBands converts NeTEx TimeBands to GTFS frequencies
func (p *DefaultFrequencyProducer) ProduceFromTimeBands(timeBands []model.TimeBand, tripID string) ([]*model.Frequency, error) {
	if len(timeBands) == 0 {
		return nil, nil
	}

	var frequencies []*model.Frequency

	for _, band := range timeBands {
		// Convert headway interval to seconds
		headwaySeconds, err := p.convertDurationToSeconds(band.ScheduledHeadwayInterval)
		if err != nil {
			continue // Skip invalid time bands
		}

		// Parse start and end times
		startTime, err := p.convertTimeToSeconds(band.StartTime)
		if err != nil {
			continue // Skip invalid time bands
		}

		endTime, err := p.convertTimeToSeconds(band.EndTime)
		if err != nil {
			continue // Skip invalid time bands
		}

		// Create GTFS frequency entry for this time band
		frequency := &model.Frequency{
			TripID:      tripID,
			StartTime:   p.formatTimeForGTFS(startTime),
			EndTime:     p.formatTimeForGTFS(endTime),
			HeadwaySecs: headwaySeconds,
			ExactTimes:  "0", // Frequency-based
		}

		frequencies = append(frequencies, frequency)
	}

	return frequencies, nil
}

// ProduceFrequencyTrip creates a template trip for frequency-based services
func (p *DefaultFrequencyProducer) ProduceFrequencyTrip(group *model.HeadwayJourneyGroup, route *model.GtfsRoute) (*model.Trip, error) {
	if group == nil || route == nil {
		return nil, fmt.Errorf("group and route are required")
	}

	// Create template trip ID
	tripID := fmt.Sprintf("freq_trip_%s", group.ID)

	// Determine service ID from day types
	serviceID, err := p.determineServiceID(group.DayTypes)
	if err != nil {
		return nil, fmt.Errorf("failed to determine service ID: %w", err)
	}

	trip := &model.Trip{
		RouteID:      route.RouteID,
		ServiceID:    serviceID,
		TripID:       tripID,
		TripHeadsign: group.Name, // Use group name as headsign
		// DirectionID will be determined from journey pattern
		// ShapeID will be determined from journey pattern
	}

	return trip, nil
}

// ProduceFrequencyStopTimes creates template stop times for frequency-based services
func (p *DefaultFrequencyProducer) ProduceFrequencyStopTimes(group *model.HeadwayJourneyGroup, trip *model.Trip) ([]*model.StopTime, error) {
	if group == nil || trip == nil {
		return nil, fmt.Errorf("group and trip are required")
	}

	// Get the journey pattern
	journeyPattern := p.netexRepo.GetJourneyPatternById(group.JourneyPatternRef)
	if journeyPattern == nil {
		return nil, fmt.Errorf("journey pattern not found: %s", group.JourneyPatternRef)
	}

	var stopTimes []*model.StopTime

	// For frequency-based services, we create template stop times with relative timing
	if journeyPattern.PointsInSequence != nil {
		baseTime := 0 // Start at time 0 (will be offset by frequency schedule)

		for i, pointInterface := range journeyPattern.PointsInSequence.PointInJourneyPatternOrStopPointInJourneyPatternOrTimingPointInJourneyPattern {
			if stopPoint, ok := pointInterface.(*model.StopPointInJourneyPattern); ok {
				// Resolve stop ID
				stopID := p.resolveStopID(stopPoint)
				if stopID == "" {
					continue // Skip if we can't resolve the stop
				}

				// Calculate arrival and departure times based on pattern
				// For frequency services, we use relative times from the start
				arrivalTime := baseTime + (i * 60) // Assume 1 minute between stops as default
				departureTime := arrivalTime + 30  // Assume 30-second stop time

				stopTime := &model.StopTime{
					TripID:        trip.TripID,
					ArrivalTime:   p.formatTimeForGTFS(arrivalTime),
					DepartureTime: p.formatTimeForGTFS(departureTime),
					StopID:        stopID,
					StopSequence:  i + 1,
					PickupType:    "0", // Regular pickup
					DropOffType:   "0", // Regular drop off
				}

				stopTimes = append(stopTimes, stopTime)
			}
		}
	}

	return stopTimes, nil
}

// Helper methods

// convertDurationToSeconds converts ISO 8601 duration to seconds
func (p *DefaultFrequencyProducer) convertDurationToSeconds(isoDuration string) (int, error) {
	if isoDuration == "" {
		return 0, fmt.Errorf("empty duration")
	}

	// Simple ISO 8601 duration parsing
	// Format: PT15M (15 minutes), PT1H30M (1 hour 30 minutes), etc.
	duration := strings.ToUpper(isoDuration)

	if !strings.HasPrefix(duration, "PT") {
		return 0, fmt.Errorf("invalid duration format: %s", isoDuration)
	}

	duration = duration[2:] // Remove "PT"
	totalSeconds := 0

	// Parse hours
	if hIndex := strings.Index(duration, "H"); hIndex > 0 {
		hoursStr := duration[:hIndex]
		if hours, err := strconv.Atoi(hoursStr); err == nil {
			totalSeconds += hours * 3600
		}
		duration = duration[hIndex+1:]
	}

	// Parse minutes
	if mIndex := strings.Index(duration, "M"); mIndex > 0 {
		minutesStr := duration[:mIndex]
		if minutes, err := strconv.Atoi(minutesStr); err == nil {
			totalSeconds += minutes * 60
		}
		duration = duration[mIndex+1:]
	}

	// Parse seconds
	if sIndex := strings.Index(duration, "S"); sIndex > 0 {
		secondsStr := duration[:sIndex]
		if seconds, err := strconv.Atoi(secondsStr); err == nil {
			totalSeconds += seconds
		}
	}

	if totalSeconds == 0 {
		return 0, fmt.Errorf("could not parse duration: %s", isoDuration)
	}

	return totalSeconds, nil
}

// convertTimeToSeconds converts HH:MM:SS to seconds since midnight
func (p *DefaultFrequencyProducer) convertTimeToSeconds(timeStr string) (int, error) {
	if timeStr == "" {
		return 0, fmt.Errorf("empty time string")
	}

	parts := strings.Split(timeStr, ":")
	if len(parts) != 3 {
		return 0, fmt.Errorf("invalid time format: %s", timeStr)
	}

	hours, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, fmt.Errorf("invalid hours: %s", parts[0])
	}

	minutes, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, fmt.Errorf("invalid minutes: %s", parts[1])
	}

	seconds, err := strconv.Atoi(parts[2])
	if err != nil {
		return 0, fmt.Errorf("invalid seconds: %s", parts[2])
	}

	return hours*3600 + minutes*60 + seconds, nil
}

// formatTimeForGTFS formats seconds since midnight as HH:MM:SS for GTFS
func (p *DefaultFrequencyProducer) formatTimeForGTFS(seconds int) string {
	hours := seconds / 3600
	minutes := (seconds % 3600) / 60
	secs := seconds % 60

	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, secs)
}

// determineServiceID determines the service ID from day types
func (p *DefaultFrequencyProducer) determineServiceID(dayTypes *model.DayTypeRefs) (string, error) {
	if dayTypes == nil || len(dayTypes.DayTypeRef) == 0 {
		return "default_service", nil
	}

	// Use the first day type reference as the service ID
	// In practice, this might need more sophisticated logic
	return dayTypes.DayTypeRef[0], nil
}

// resolveStopID resolves stop ID from stop point in journey pattern
func (p *DefaultFrequencyProducer) resolveStopID(stopPoint *model.StopPointInJourneyPattern) string {
	if stopPoint == nil {
		return ""
	}

	// Try to resolve through scheduled stop point
	if stopPoint.ScheduledStopPointRef != "" {
		if ssp := p.netexRepo.GetScheduledStopPointById(stopPoint.ScheduledStopPointRef); ssp != nil {
			if ssp.QuayRef != "" {
				return ssp.QuayRef
			}
		}
	}

	return ""
}

// ValidateFrequencyTrip validates that a trip is suitable for frequency-based service
func (p *DefaultFrequencyProducer) ValidateFrequencyTrip(trip *model.Trip, frequencies []*model.Frequency) error {
	if trip == nil {
		return fmt.Errorf("trip cannot be nil")
	}

	if len(frequencies) == 0 {
		return fmt.Errorf("at least one frequency entry is required")
	}

	// Check that all frequencies reference the same trip
	for _, freq := range frequencies {
		if freq.TripID != trip.TripID {
			return fmt.Errorf("frequency trip ID %s does not match trip ID %s", freq.TripID, trip.TripID)
		}
	}

	// Validate frequency parameters
	for i, freq := range frequencies {
		if freq.HeadwaySecs <= 0 {
			return fmt.Errorf("frequency %d has invalid headway: %d", i, freq.HeadwaySecs)
		}

		// Parse times to validate format
		if _, err := p.convertTimeToSeconds(freq.StartTime); err != nil {
			return fmt.Errorf("frequency %d has invalid start time: %s", i, freq.StartTime)
		}

		if _, err := p.convertTimeToSeconds(freq.EndTime); err != nil {
			return fmt.Errorf("frequency %d has invalid end time: %s", i, freq.EndTime)
		}
	}

	return nil
}

// CreateFrequencyBasedService creates a complete frequency-based service from NeTEx data
func (p *DefaultFrequencyProducer) CreateFrequencyBasedService(group *model.HeadwayJourneyGroup, line *model.Line) (*FrequencyService, error) {
	if group == nil || line == nil {
		return nil, fmt.Errorf("group and line are required")
	}

	// Create GTFS route (if needed)
	var route *model.GtfsRoute
	// In practice, this would use a route producer
	route = &model.GtfsRoute{
		RouteID:        line.ID,
		RouteShortName: line.ShortName,
		RouteLongName:  line.Name,
		RouteType:      3, // Default to bus
	}

	// Create template trip
	trip, err := p.ProduceFrequencyTrip(group, route)
	if err != nil {
		return nil, fmt.Errorf("failed to create frequency trip: %w", err)
	}

	// Create frequencies
	frequencies, err := p.ProduceFromHeadwayJourneyGroup(group)
	if err != nil {
		return nil, fmt.Errorf("failed to create frequencies: %w", err)
	}

	// Create template stop times
	stopTimes, err := p.ProduceFrequencyStopTimes(group, trip)
	if err != nil {
		return nil, fmt.Errorf("failed to create stop times: %w", err)
	}

	// Validate the complete service
	if err := p.ValidateFrequencyTrip(trip, frequencies); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	service := &FrequencyService{
		Route:       route,
		Trip:        trip,
		Frequencies: frequencies,
		StopTimes:   stopTimes,
		Group:       group,
	}

	return service, nil
}

// FrequencyService represents a complete frequency-based service
type FrequencyService struct {
	Route       *model.GtfsRoute
	Trip        *model.Trip
	Frequencies []*model.Frequency
	StopTimes   []*model.StopTime
	Group       *model.HeadwayJourneyGroup
}

// GetServiceID returns the service ID for this frequency service
func (fs *FrequencyService) GetServiceID() string {
	if fs.Trip != nil {
		return fs.Trip.ServiceID
	}
	return ""
}

// GetTotalServiceHours returns the total hours of service per day
func (fs *FrequencyService) GetTotalServiceHours() float64 {
	if len(fs.Frequencies) == 0 {
		return 0
	}

	var totalHours float64
	producer := &DefaultFrequencyProducer{} // For helper methods

	for _, freq := range fs.Frequencies {
		startSeconds, _ := producer.convertTimeToSeconds(freq.StartTime)
		endSeconds, _ := producer.convertTimeToSeconds(freq.EndTime)

		if endSeconds > startSeconds {
			hours := float64(endSeconds-startSeconds) / 3600.0
			totalHours += hours
		}
	}

	return totalHours
}

// GetAverageHeadway returns the average headway in minutes
func (fs *FrequencyService) GetAverageHeadway() float64 {
	if len(fs.Frequencies) == 0 {
		return 0
	}

	var totalHeadway int
	for _, freq := range fs.Frequencies {
		totalHeadway += freq.HeadwaySecs
	}

	averageSeconds := float64(totalHeadway) / float64(len(fs.Frequencies))
	return averageSeconds / 60.0 // Convert to minutes
}
