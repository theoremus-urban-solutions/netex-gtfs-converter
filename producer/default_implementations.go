package producer

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/theoremus-urban-solutions/netex-gtfs-converter/model"
)

// Default implementations for producer interfaces

// DefaultAgencyProducer implements AgencyProducer
type DefaultAgencyProducer struct {
	netexRepository NetexRepository
}

func NewDefaultAgencyProducer(netexRepository NetexRepository) *DefaultAgencyProducer {
	return &DefaultAgencyProducer{
		netexRepository: netexRepository,
	}
}

func (p *DefaultAgencyProducer) Produce(authority *model.Authority) (*model.Agency, error) {
	var phone, email, url string
	if authority.ContactDetails != nil {
		phone = authority.ContactDetails.Phone
		email = authority.ContactDetails.Email
		if url == "" {
			url = authority.ContactDetails.URL
		}
	}
	if url == "" {
		url = authority.URL
	}
	// timezone from repository if available
	tz := p.netexRepository.GetTimeZone()
	if tz == "" {
		tz = "UTC"
	}
	return &model.Agency{
		AgencyID:       authority.ID,
		AgencyName:     firstNonEmpty(authority.Name, authority.ShortName),
		AgencyURL:      url,
		AgencyTimezone: tz,
		AgencyLang:     "",
		AgencyPhone:    phone,
		AgencyEmail:    email,
	}, nil
}

// DefaultRouteProducer implements RouteProducer
type DefaultRouteProducer struct {
	netexRepository NetexRepository
	gtfsRepository  GtfsRepository
}

func NewDefaultRouteProducer(netexRepository NetexRepository, gtfsRepository GtfsRepository) *DefaultRouteProducer {
	return &DefaultRouteProducer{
		netexRepository: netexRepository,
		gtfsRepository:  gtfsRepository,
	}
}

func (p *DefaultRouteProducer) Produce(line *model.Line) (*model.GtfsRoute, error) {
	// Fill fields from NeTEx line
	gtfsType := model.MapNetexToGtfsRouteType(line.TransportMode, line.TransportSubmode).Value()

	// For route names, prefer PublicCode for short name (e.g., "1", "2")
	// and use Name/ShortName for long name (e.g., "Downtown Express")
	shortName := line.PublicCode
	longName := line.Name

	// If PublicCode is not available, try to extract from ShortName
	if shortName == "" {
		shortName = line.ShortName
		// If ShortName and Name are the same, try to extract a number
		if shortName == longName && len(shortName) > 0 {
			// Try to extract line number from name like "Ligne 1 ..."
			if parts := strings.Fields(shortName); len(parts) > 1 {
				for _, part := range parts {
					if _, err := strconv.Atoi(part); err == nil {
						shortName = part
						break
					}
				}
			}
		}
	}

	// Ensure short and long names are different
	if shortName == longName && longName != "" {
		// If they're still the same, clear the short name to avoid duplication
		shortName = ""
	}

	return &model.GtfsRoute{
		RouteID:        line.ID,
		AgencyID:       p.netexRepository.GetAuthorityIdForLine(line),
		RouteShortName: shortName,
		RouteLongName:  longName,
		RouteDesc:      line.Description,
		RouteType:      gtfsType,
		RouteURL:       line.URL,
		RouteColor:     valueOrEmpty(line.Presentation, func(p *model.Presentation) string { return p.Colour }),
		RouteTextColor: valueOrEmpty(line.Presentation, func(p *model.Presentation) string { return p.TextColour }),
	}, nil
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

func valueOrEmpty[T any](ptr *T, getter func(*T) string) string {
	if ptr == nil {
		return ""
	}
	return getter(ptr)
}

// DefaultTripProducer implements TripProducer
type DefaultTripProducer struct {
	netexRepository NetexRepository
	gtfsRepository  GtfsRepository
}

func NewDefaultTripProducer(netexRepository NetexRepository, gtfsRepository GtfsRepository) *DefaultTripProducer {
	return &DefaultTripProducer{
		netexRepository: netexRepository,
		gtfsRepository:  gtfsRepository,
	}
}

func (p *DefaultTripProducer) Produce(input TripInput) (*model.Trip, error) {
	trip := &model.Trip{
		TripID:  input.ServiceJourney.ID,
		RouteID: input.GtfsRoute.RouteID,
	}

	// Use default service ID to match calendar.txt
	trip.ServiceID = "default_service"

	// Set headsign from destination display
	if input.DestinationDisplay != nil {
		trip.TripHeadsign = firstNonEmpty(input.DestinationDisplay.FrontText, input.DestinationDisplay.SideText)
	}

	// Set shape ID if provided
	if input.ShapeID != "" {
		trip.ShapeID = input.ShapeID
	}

	// Determine direction based on route direction
	if input.NetexRoute != nil && input.NetexRoute.DirectionType != "" {
		switch input.NetexRoute.DirectionType {
		case "inbound", "return":
			trip.DirectionID = "1"
		case "outbound", "outward":
			trip.DirectionID = "0"
		}
	}

	// Handle service alterations (cancelled trips)
	if input.ServiceJourney.ServiceAlteration == "cancelled" {
		// Skip cancelled trips
		return nil, nil
	}

	return trip, nil
}

// DefaultStopProducer implements StopProducer
type DefaultStopProducer struct {
	stopAreaRepository StopAreaRepository
	gtfsRepository     GtfsRepository
}

func NewDefaultStopProducer(stopAreaRepository StopAreaRepository, gtfsRepository GtfsRepository) *DefaultStopProducer {
	return &DefaultStopProducer{
		stopAreaRepository: stopAreaRepository,
		gtfsRepository:     gtfsRepository,
	}
}

func (p *DefaultStopProducer) ProduceStopFromQuay(quay *model.Quay) (*model.Stop, error) {
	stopPlace := p.stopAreaRepository.GetStopPlaceByQuayId(quay.ID)
	name := firstNonEmpty(quay.Name, quay.ShortName, quay.PublicCode)
	if name == "" && stopPlace != nil {
		name = firstNonEmpty(stopPlace.Name, stopPlace.ShortName)
	}
	var lat, lon float64
	if quay.Centroid != nil && quay.Centroid.Location != nil {
		lat = quay.Centroid.Location.Latitude
		lon = quay.Centroid.Location.Longitude
	}
	parentStation := ""
	if stopPlace != nil {
		parentStation = stopPlace.ID
	}
	return &model.Stop{
		StopID:        quay.ID,
		StopCode:      quay.PublicCode,
		StopName:      name,
		StopLat:       lat,
		StopLon:       lon,
		LocationType:  "0",
		ParentStation: parentStation,
		PlatformCode:  quay.PublicCode,
	}, nil
}

func (p *DefaultStopProducer) ProduceStopFromStopPlace(stopPlace *model.StopPlace) (*model.Stop, error) {
	var lat, lon float64
	if stopPlace.Centroid != nil && stopPlace.Centroid.Location != nil {
		lat = stopPlace.Centroid.Location.Latitude
		lon = stopPlace.Centroid.Location.Longitude
	}
	return &model.Stop{
		StopID:       stopPlace.ID,
		StopName:     stopPlace.Name,
		StopLat:      lat,
		StopLon:      lon,
		LocationType: "1", // station
	}, nil
}

// DefaultStopTimeProducer implements StopTimeProducer
type DefaultStopTimeProducer struct {
	netexRepository NetexRepository
	gtfsRepository  GtfsRepository
}

func NewDefaultStopTimeProducer(netexRepository NetexRepository, gtfsRepository GtfsRepository) *DefaultStopTimeProducer {
	return &DefaultStopTimeProducer{
		netexRepository: netexRepository,
		gtfsRepository:  gtfsRepository,
	}
}

func (p *DefaultStopTimeProducer) Produce(input StopTimeInput) (*model.StopTime, error) {
	st := &model.StopTime{
		TripID: input.Trip.TripID,
	}

	// Handle times as strings, applying day offset if needed
	if input.TimetabledPassingTime.ArrivalTime != "" {
		arrivalTime := input.TimetabledPassingTime.ArrivalTime
		// Handle day offset (add 24 hours for each day)
		if input.TimetabledPassingTime.DayOffset > 0 {
			if parsedTime, err := time.Parse("15:04:05", arrivalTime); err == nil {
				arrivalTime = fmt.Sprintf("%02d:%s",
					parsedTime.Hour()+(input.TimetabledPassingTime.DayOffset*24),
					parsedTime.Format("04:05"))
			}
		}
		st.ArrivalTime = arrivalTime
	}

	if input.TimetabledPassingTime.DepartureTime != "" {
		departureTime := input.TimetabledPassingTime.DepartureTime
		// Handle day offset
		if input.TimetabledPassingTime.DayOffset > 0 {
			if parsedTime, err := time.Parse("15:04:05", departureTime); err == nil {
				departureTime = fmt.Sprintf("%02d:%s",
					parsedTime.Hour()+(input.TimetabledPassingTime.DayOffset*24),
					parsedTime.Format("04:05"))
			}
		}
		st.DepartureTime = departureTime
	}

	// If only one time is available, use it for both
	if st.ArrivalTime == "" && st.DepartureTime != "" {
		st.ArrivalTime = st.DepartureTime
	}
	if st.DepartureTime == "" && st.ArrivalTime != "" {
		st.DepartureTime = st.ArrivalTime
	}

	// Resolve stop_id from PointInJourneyPatternRef → ScheduledStopPointRef → QuayRef/StopPlaceRef
	if input.TimetabledPassingTime.PointInJourneyPatternRef != "" {
		pjpRef := input.TimetabledPassingTime.PointInJourneyPatternRef
		sspRef := p.netexRepository.GetScheduledStopPointRefByPointInJourneyPatternRef(pjpRef)
		if sspRef != "" {
			// Look up ScheduledStopPoint to get QuayRef or StopPlaceRef
			if ssp := p.netexRepository.GetScheduledStopPointById(sspRef); ssp != nil {
				if ssp.QuayRef != "" {
					st.StopID = ssp.QuayRef
				} else if ssp.StopPlaceRef != "" {
					st.StopID = ssp.StopPlaceRef
				} else {
					st.StopID = sspRef
				}
			} else {
				// Fallback: use reference directly
				st.StopID = sspRef
			}
		}
	}

	// Set headsign if provided
	if input.CurrentHeadSign != "" {
		st.StopHeadsign = input.CurrentHeadSign
	}

	// Set pickup and drop-off types (default to regular)
	st.PickupType = "0"  // Regular pickup
	st.DropOffType = "0" // Regular drop-off

	// Shape distance is handled in Enhanced GTFS Exporter
	st.ShapeDistTraveled = 0

	return st, nil
}

// DefaultServiceCalendarProducer implements ServiceCalendarProducer
type DefaultServiceCalendarProducer struct {
	gtfsRepository GtfsRepository
}

func NewDefaultServiceCalendarProducer(gtfsRepository GtfsRepository) *DefaultServiceCalendarProducer {
	return &DefaultServiceCalendarProducer{
		gtfsRepository: gtfsRepository,
	}
}

func (p *DefaultServiceCalendarProducer) Produce(serviceID string, dayTypes []*model.DayType) (*model.Calendar, error) {
	if len(dayTypes) == 0 {
		return nil, nil
	}

	calendar := &model.Calendar{
		ServiceID: serviceID,
		Monday:    false,
		Tuesday:   false,
		Wednesday: false,
		Thursday:  false,
		Friday:    false,
		Saturday:  false,
		Sunday:    false,
		StartDate: "20240101", // Default start date
		EndDate:   "20251231", // Default end date
	}

	// Analyze DayType properties to determine service days
	for _, dayType := range dayTypes {
		if dayType != nil && dayType.Properties != nil {
			for _, prop := range dayType.Properties.PropertyOfDay {
				daysOfWeek := prop.DaysOfWeek
				switch daysOfWeek {
				case "Monday", "monday", "1":
					calendar.Monday = true
				case "Tuesday", "tuesday", "2":
					calendar.Tuesday = true
				case "Wednesday", "wednesday", "3":
					calendar.Wednesday = true
				case "Thursday", "thursday", "4":
					calendar.Thursday = true
				case "Friday", "friday", "5":
					calendar.Friday = true
				case "Saturday", "saturday", "6":
					calendar.Saturday = true
				case "Sunday", "sunday", "7":
					calendar.Sunday = true
				case "Weekdays", "weekdays":
					calendar.Monday = true
					calendar.Tuesday = true
					calendar.Wednesday = true
					calendar.Thursday = true
					calendar.Friday = true
				case "Weekend", "weekend":
					calendar.Saturday = true
					calendar.Sunday = true
				case "Everyday", "everyday", "Daily", "daily":
					calendar.Monday = true
					calendar.Tuesday = true
					calendar.Wednesday = true
					calendar.Thursday = true
					calendar.Friday = true
					calendar.Saturday = true
					calendar.Sunday = true
				}
			}
		}
	}

	// If no specific days are set, assume it runs every day
	if !calendar.Monday && !calendar.Tuesday && !calendar.Wednesday &&
		!calendar.Thursday && !calendar.Friday && !calendar.Saturday && !calendar.Sunday {
		calendar.Monday = true
		calendar.Tuesday = true
		calendar.Wednesday = true
		calendar.Thursday = true
		calendar.Friday = true
		calendar.Saturday = true
		calendar.Sunday = true
	}

	return calendar, nil
}

// DefaultServiceCalendarDateProducer implements ServiceCalendarDateProducer
type DefaultServiceCalendarDateProducer struct {
	gtfsRepository GtfsRepository
}

func NewDefaultServiceCalendarDateProducer(gtfsRepository GtfsRepository) *DefaultServiceCalendarDateProducer {
	return &DefaultServiceCalendarDateProducer{
		gtfsRepository: gtfsRepository,
	}
}

func (p *DefaultServiceCalendarDateProducer) Produce(serviceID string, dayTypeAssignments []*model.DayTypeAssignment) ([]*model.CalendarDate, error) {
	result := make([]*model.CalendarDate, 0)

	for _, assignment := range dayTypeAssignments {
		if assignment == nil {
			continue
		}

		// Handle operating day references
		if assignment.OperatingDayRef != "" {
			// The operating day ref might need to be resolved to an actual date
			// For now, assume it's in YYYYMMDD format or can be converted
			dateStr := assignment.OperatingDayRef

			// If it's not in YYYYMMDD format, try to parse and convert
			if len(dateStr) == 10 && (dateStr[4] == '-' || dateStr[4] == '/') {
				// Convert YYYY-MM-DD or YYYY/MM/DD to YYYYMMDD
				dateStr = dateStr[0:4] + dateStr[5:7] + dateStr[8:10]
			}

			calendarDate := &model.CalendarDate{
				ServiceID:     serviceID,
				Date:          dateStr,
				ExceptionType: 1, // Service added
			}

			// If service is not available, it's an exception (service removed)
			if !assignment.IsAvailable {
				calendarDate.ExceptionType = 2
			}

			result = append(result, calendarDate)
		}

		// Handle operating period references
		if assignment.OperatingPeriodRef != "" {
			// Operating periods represent date ranges
			// This would typically need to be expanded into individual dates
			// For now, we skip this complex logic
		}
	}

	return result, nil
}

func ternaryInt(b bool, t, f int) int {
	if b {
		return t
	}
	return f
}

// DefaultShapeProducer implements ShapeProducer
type DefaultShapeProducer struct {
	netexRepository NetexRepository
	gtfsRepository  GtfsRepository
}

func NewDefaultShapeProducer(netexRepository NetexRepository, gtfsRepository GtfsRepository) *DefaultShapeProducer {
	return &DefaultShapeProducer{
		netexRepository: netexRepository,
		gtfsRepository:  gtfsRepository,
	}
}

func (p *DefaultShapeProducer) Produce(journeyPattern *model.JourneyPattern) (*model.Shape, error) {
	if journeyPattern == nil {
		return nil, nil
	}

	// Skip shape generation to avoid invalid (0,0) coordinates
	// This prevents GTFS validation errors with placeholder coordinates
	// In a full implementation, this would extract actual route geometry
	return nil, nil
}

// DefaultTransferProducer implements TransferProducer
type DefaultTransferProducer struct {
	netexRepository NetexRepository
	gtfsRepository  GtfsRepository
}

func NewDefaultTransferProducer(netexRepository NetexRepository, gtfsRepository GtfsRepository) *DefaultTransferProducer {
	return &DefaultTransferProducer{
		netexRepository: netexRepository,
		gtfsRepository:  gtfsRepository,
	}
}

func (p *DefaultTransferProducer) Produce(interchange *model.ServiceJourneyInterchange) (*model.Transfer, error) {
	if interchange == nil {
		return nil, nil
	}

	transfer := &model.Transfer{
		FromStopID: interchange.FromPointRef,
		ToStopID:   interchange.ToPointRef,
	}

	// Set transfer type based on NeTEx properties
	if interchange.StaySeated {
		transfer.TransferType = 0 // Recommended transfer point
	} else if interchange.Guaranteed {
		transfer.TransferType = 1 // Timed transfer point
	} else {
		transfer.TransferType = 2 // Minimum transfer time required
	}

	// Parse minimum transfer time if available
	if interchange.MinimumTransferTime != "" {
		// NeTEx uses ISO 8601 duration format (PT5M for 5 minutes)
		// Simple parsing for minutes - in full implementation would use proper ISO 8601 parser
		if len(interchange.MinimumTransferTime) > 3 &&
			interchange.MinimumTransferTime[:2] == "PT" &&
			interchange.MinimumTransferTime[len(interchange.MinimumTransferTime)-1:] == "M" {

			// Extract number of minutes
			minutesStr := interchange.MinimumTransferTime[2 : len(interchange.MinimumTransferTime)-1]
			if minutes := parseIntSafe(minutesStr); minutes > 0 {
				transfer.MinTransferTime = minutes * 60 // Convert to seconds
			}
		}
	}

	// Set default transfer time if none specified but guaranteed
	if transfer.MinTransferTime == 0 && interchange.Guaranteed {
		transfer.MinTransferTime = 120 // 2 minutes default
	}

	// Resolve from/to journey references to route/trip if needed
	// This would require additional lookup logic in a full implementation
	transfer.FromRouteID = ""
	transfer.ToRouteID = ""
	transfer.FromTripID = interchange.FromJourneyRef
	transfer.ToTripID = interchange.ToJourneyRef

	return transfer, nil
}

// Helper function to safely parse integers
func parseIntSafe(s string) int {
	result := 0
	for _, r := range s {
		if r >= '0' && r <= '9' {
			result = result*10 + int(r-'0')
		} else {
			return 0 // Invalid number
		}
	}
	return result
}

// DefaultFeedInfoProducer implements FeedInfoProducer
type DefaultFeedInfoProducer struct{}

func NewDefaultFeedInfoProducer() *DefaultFeedInfoProducer {
	return &DefaultFeedInfoProducer{}
}

func (p *DefaultFeedInfoProducer) ProduceFeedInfo() (*model.FeedInfo, error) {
	feedInfo := &model.FeedInfo{
		FeedPublisherName: "NeTEx to GTFS Converter (Go)",
		FeedPublisherURL:  "https://github.com/entur/netex-gtfs-converter-java",
		FeedLang:          "en",
		FeedVersion:       "1.0.0",
		FeedContactEmail:  "",
		FeedContactURL:    "",
	}

	// Set date range - would typically be calculated from actual service dates
	// For now, use reasonable defaults
	feedInfo.FeedStartDate = "20240101"
	feedInfo.FeedEndDate = "20251231"

	return feedInfo, nil
}
