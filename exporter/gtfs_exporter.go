package exporter

import (
	"io"

	"github.com/theoremus-urban-solutions/netex-gtfs-converter/loader"
	"github.com/theoremus-urban-solutions/netex-gtfs-converter/model"
	"github.com/theoremus-urban-solutions/netex-gtfs-converter/producer"
	"github.com/theoremus-urban-solutions/netex-gtfs-converter/repository"
)

// GtfsExporter is the main interface for converting NeTEx datasets to GTFS
type GtfsExporter interface {
	// ConvertTimetablesToGtfs converts a NeTEx timetable dataset to GTFS format
	ConvertTimetablesToGtfs(netexData io.Reader) (io.Reader, error)

	// ConvertStopsToGtfs converts only stop data to GTFS format (no timetable data)
	ConvertStopsToGtfs() (io.Reader, error)

	// Extension points for customization
	SetAgencyProducer(producer producer.AgencyProducer)
	SetRouteProducer(producer producer.RouteProducer)
	SetTripProducer(producer producer.TripProducer)
	SetStopProducer(producer producer.StopProducer)
	SetStopTimeProducer(producer producer.StopTimeProducer)
	SetServiceCalendarProducer(producer producer.ServiceCalendarProducer)
	SetServiceCalendarDateProducer(producer producer.ServiceCalendarDateProducer)
	SetShapeProducer(producer producer.ShapeProducer)
	SetTransferProducer(producer producer.TransferProducer)
	SetFeedInfoProducer(producer producer.FeedInfoProducer)

	// Get repositories for access to data
	GetNetexRepository() producer.NetexRepository
	GetGtfsRepository() producer.GtfsRepository
	GetStopAreaRepository() producer.StopAreaRepository
}

// DefaultGtfsExporter implements the GtfsExporter interface
type DefaultGtfsExporter struct {
	codespace string
	// European profile only; no config
	netexRepository    producer.NetexRepository
	gtfsRepository     producer.GtfsRepository
	stopAreaRepository producer.StopAreaRepository

	// Producers
	agencyProducer              producer.AgencyProducer
	routeProducer               producer.RouteProducer
	tripProducer                producer.TripProducer
	stopProducer                producer.StopProducer
	stopTimeProducer            producer.StopTimeProducer
	serviceCalendarProducer     producer.ServiceCalendarProducer
	serviceCalendarDateProducer producer.ServiceCalendarDateProducer
	shapeProducer               producer.ShapeProducer
	transferProducer            producer.TransferProducer
	feedInfoProducer            producer.FeedInfoProducer

	// internal cache
	lineIdToGtfsRoute map[string]*model.GtfsRoute
}

// NewDefaultGtfsExporter creates a new default GTFS exporter
func NewDefaultGtfsExporter(codespace string, stopAreaRepository producer.StopAreaRepository) *DefaultGtfsExporter {
	netexRepo := repository.NewDefaultNetexRepository()
	gtfsRepo := repository.NewDefaultGtfsRepository()

	exporter := &DefaultGtfsExporter{
		codespace:          codespace,
		netexRepository:    netexRepo,
		gtfsRepository:     gtfsRepo,
		stopAreaRepository: stopAreaRepository,
		lineIdToGtfsRoute:  make(map[string]*model.GtfsRoute),
	}

	// Initialize default producers
	exporter.initializeDefaultProducers()

	return exporter
}

// initializeDefaultProducers sets up the default producer implementations
func (e *DefaultGtfsExporter) initializeDefaultProducers() {
	e.agencyProducer = producer.NewDefaultAgencyProducer(e.netexRepository)
	e.routeProducer = producer.NewDefaultRouteProducer(e.netexRepository, e.gtfsRepository)
	e.tripProducer = producer.NewDefaultTripProducer(e.netexRepository, e.gtfsRepository)
	e.stopProducer = producer.NewDefaultStopProducer(e.stopAreaRepository, e.gtfsRepository)
	e.stopTimeProducer = producer.NewDefaultStopTimeProducer(e.netexRepository, e.gtfsRepository)
	e.serviceCalendarProducer = producer.NewDefaultServiceCalendarProducer(e.gtfsRepository)
	e.serviceCalendarDateProducer = producer.NewDefaultServiceCalendarDateProducer(e.gtfsRepository)
	e.shapeProducer = producer.NewDefaultShapeProducer(e.netexRepository, e.gtfsRepository)
	e.transferProducer = producer.NewDefaultTransferProducer(e.netexRepository, e.gtfsRepository)
	e.feedInfoProducer = producer.NewDefaultFeedInfoProducer()
}

// ConvertTimetablesToGtfs implements the main conversion logic
func (e *DefaultGtfsExporter) ConvertTimetablesToGtfs(netexData io.Reader) (io.Reader, error) {
	if e.codespace == "" {
		return nil, ErrMissingCodespace
	}

	// Load NeTEx data
	if err := e.loadNetex(netexData); err != nil {
		return nil, err
	}

	// Convert to GTFS
	if err := e.convertNetexToGtfs(); err != nil {
		return nil, err
	}

	// Write GTFS archive
	return e.gtfsRepository.WriteGtfs()
}

// ConvertStopsToGtfs converts only stop data
func (e *DefaultGtfsExporter) ConvertStopsToGtfs() (io.Reader, error) {
	if err := e.convertStops(false); err != nil {
		return nil, err
	}

	if err := e.ensureDefaultAgency(); err != nil {
		return nil, err
	}

	if err := e.addFeedInfo(); err != nil {
		return nil, err
	}

	return e.gtfsRepository.WriteGtfs()
}

// loadNetex loads NeTEx data into the repository
func (e *DefaultGtfsExporter) loadNetex(netexData io.Reader) error {
	loaderImpl := loader.NewDefaultNetexDatasetLoader()
	return loaderImpl.Load(netexData, e.netexRepository)
}

// convertNetexToGtfs orchestrates the conversion process
func (e *DefaultGtfsExporter) convertNetexToGtfs() error {
	// Convert agencies
	if err := e.convertAgencies(); err != nil {
		return err
	}

	// Convert stops
	if err := e.convertStops(true); err != nil {
		return err
	}

	// Convert routes
	if err := e.convertRoutes(); err != nil {
		return err
	}

	// Convert services and trips
	if err := e.convertServices(); err != nil {
		return err
	}

	// Convert transfers
	if err := e.convertTransfers(); err != nil {
		return err
	}

	// Add feed info
	if err := e.ensureDefaultAgency(); err != nil {
		return err
	}
	return e.addFeedInfo()
}

// convertAgencies converts NeTEx authorities to GTFS agencies
func (e *DefaultGtfsExporter) convertAgencies() error {
	lines := e.netexRepository.GetLines()
	if len(lines) == 0 {
		return ConversionError{Stage: "agencies", Err: ErrNoDataFound}
	}

	authorityIDs := make(map[string]bool)

	// Collect unique authority IDs from lines
	for _, line := range lines {
		if line == nil {
			continue
		}
		authorityID := e.netexRepository.GetAuthorityIdForLine(line)
		if authorityID != "" {
			authorityIDs[authorityID] = true
		}
	}

	if len(authorityIDs) == 0 {
		return ConversionError{Stage: "agencies", Err: ErrNoDataFound}
	}

	// Convert each authority to GTFS agency
	for authorityID := range authorityIDs {
		authority := e.netexRepository.GetAuthorityById(authorityID)
		if authority == nil {
			continue // Skip missing authorities
		}

		agency, err := e.agencyProducer.Produce(authority)
		if err != nil {
			return ConversionError{Stage: "agencies", EntityID: authorityID, Err: err}
		}

		if agency == nil {
			continue // Skip if producer returns nil
		}

		// Validate required fields
		if agency.AgencyName == "" {
			return ValidationError{Field: "AgencyName", Value: agency.AgencyID, Message: "agency name is required"}
		}
		if agency.AgencyTimezone == "" {
			return ValidationError{Field: "AgencyTimezone", Value: agency.AgencyID, Message: "agency timezone is required"}
		}

		if err := e.gtfsRepository.SaveEntity(agency); err != nil {
			return ConversionError{Stage: "agencies", EntityID: authorityID, Err: err}
		}
	}

	return nil
}

// convertStops converts NeTEx stops to GTFS stops
func (e *DefaultGtfsExporter) convertStops(exportOnlyUsedStops bool) error {
	// Export stations (StopPlace) as parent stops where present, then quays as platforms
	// Use StopAreaRepository if provided; else fall back to NetexRepository
	seenStations := make(map[string]bool)
	quays := e.stopAreaRepository.GetAllQuays()
	if len(quays) == 0 {
		quays = e.netexRepository.GetAllQuays()
	}
	stopPlaces := make(map[string]*model.StopPlace)
	// try to populate stop places for parent station emission
	if sps := e.netexRepository.GetAllStopPlaces(); len(sps) > 0 {
		for _, sp := range sps {
			stopPlaces[sp.ID] = sp
		}
	}
	for _, quay := range quays {
		sp := e.stopAreaRepository.GetStopPlaceByQuayId(quay.ID)
		if sp == nil {
			// try netex repo if stop area repo cannot resolve
			for _, cand := range stopPlaces {
				if cand.Quays != nil {
					for i := range cand.Quays.Quay {
						if cand.Quays.Quay[i].ID == quay.ID {
							sp = cand
							break
						}
					}
				}
				if sp != nil {
					break
				}
			}
		}
		if sp != nil && !seenStations[sp.ID] {
			station, err := e.stopProducer.ProduceStopFromStopPlace(sp)
			if err != nil {
				return err
			}
			if station != nil {
				if err := e.gtfsRepository.SaveEntity(station); err != nil {
					return err
				}
			}
			seenStations[sp.ID] = true
		}
	}
	for _, quay := range quays {
		stop, err := e.stopProducer.ProduceStopFromQuay(quay)
		if err != nil {
			return err
		}
		if stop != nil {
			if err := e.gtfsRepository.SaveEntity(stop); err != nil {
				return err
			}
		}
	}
	return nil
}

// convertRoutes converts NeTEx lines to GTFS routes
func (e *DefaultGtfsExporter) convertRoutes() error {
	lines := e.netexRepository.GetLines()
	if len(lines) == 0 {
		return ConversionError{Stage: "routes", Err: ErrNoDataFound}
	}

	for _, line := range lines {
		if line == nil {
			continue
		}

		route, err := e.routeProducer.Produce(line)
		if err != nil {
			return ConversionError{Stage: "routes", EntityID: line.ID, Err: err}
		}

		if route == nil {
			continue // Skip if producer returns nil
		}

		// Validate required fields
		if route.RouteID == "" {
			return ValidationError{Field: "RouteID", Value: line.ID, Message: "route ID is required"}
		}
		if route.RouteType == 0 && line.TransportMode == "" {
			return ValidationError{Field: "RouteType", Value: line.ID, Message: "route type or transport mode is required"}
		}

		// Validate that either route short name or long name is provided
		if route.RouteShortName == "" && route.RouteLongName == "" {
			return ValidationError{Field: "RouteName", Value: line.ID, Message: "either route short name or long name is required"}
		}

		if err := e.gtfsRepository.SaveEntity(route); err != nil {
			return ConversionError{Stage: "routes", EntityID: line.ID, Err: err}
		}
		e.lineIdToGtfsRoute[line.ID] = route
	}

	return nil
}

// convertServices converts NeTEx service journeys to GTFS trips
func (e *DefaultGtfsExporter) convertServices() error {
	// For each JourneyPattern, produce shape, trips for ServiceJourneys, and stop_times
	serviceJourneys := e.netexRepository.GetServiceJourneys()
	for _, sj := range serviceJourneys {
		// resolve JourneyPattern and Line/Route
		jp := e.netexRepository.GetJourneyPatternById(sj.JourneyPatternRef.Ref)
		if jp == nil {
			continue
		}
		// Find GTFS route by LineRef
		gtfsRoute := e.lineIdToGtfsRoute[sj.LineRef.Ref]
		if gtfsRoute == nil {
			// fall back: produce on the fly
			for _, line := range e.netexRepository.GetLines() {
				if line.ID == sj.LineRef.Ref {
					r, err := e.routeProducer.Produce(line)
					if err != nil {
						return err
					}
					if err := e.gtfsRepository.SaveEntity(r); err != nil {
						return err
					}
					e.lineIdToGtfsRoute[line.ID] = r
					gtfsRoute = r
					break
				}
			}
		}
		if gtfsRoute == nil {
			continue
		}

		// Optional shape
		var shape *model.Shape
		if jp != nil && e.shapeProducer != nil {
			s, err := e.shapeProducer.Produce(jp)
			if err != nil {
				return err
			}
			shape = s
			if shape != nil {
				_ = e.gtfsRepository.SaveEntity(shape)
			}
		}

		// Trip headsign: from DestinationDisplay of first StopPoint in JP if present
		var headsign string
		if jp != nil && jp.PointsInSequence != nil {
			// Not fully typed; keep empty unless DestinationDisplay is resolvable
		}

		// Build Trip
		trip, err := e.tripProducer.Produce(producer.TripInput{
			ServiceJourney:     sj,
			NetexRoute:         nil,
			GtfsRoute:          gtfsRoute,
			ShapeID:            shapeID(shape),
			DestinationDisplay: nil,
		})
		if err != nil {
			return err
		}
		if headsign != "" {
			trip.TripHeadsign = headsign
		}
		if shape != nil {
			trip.ShapeID = shape.ShapeID
		}
		if err := e.gtfsRepository.SaveEntity(trip); err != nil {
			return err
		}

		// Stop times from PassingTimes
		if sj.PassingTimes != nil {
			seq := 1
			for _, pt := range sj.PassingTimes.TimetabledPassingTime {
				st, err := e.stopTimeProducer.Produce(producer.StopTimeInput{
					TimetabledPassingTime: &pt,
					JourneyPattern:        jp,
					Trip:                  trip,
					Shape:                 shape,
					CurrentHeadSign:       headsign,
				})
				if err != nil {
					return err
				}
				if st != nil {
					st.StopSequence = seq
					seq++
					if err := e.gtfsRepository.SaveEntity(st); err != nil {
						return err
					}
				}
			}
		}

		// Service calendar and dates
		if e.serviceCalendarProducer != nil {
			// Simplified: serviceID equals ServiceJourney ID
			serviceID := sj.ID
			var dayTypes []*model.DayType
			if sj.DayTypes != nil {
				for _, id := range sj.DayTypes.DayTypeRef {
					dayTypes = append(dayTypes, e.netexRepository.GetDayTypeById(id))
				}
			}
			cal, err := e.serviceCalendarProducer.Produce(serviceID, dayTypes)
			if err != nil {
				return err
			}
			if cal != nil {
				if err := e.gtfsRepository.SaveEntity(cal); err != nil {
					return err
				}
			}
		}
		if e.serviceCalendarDateProducer != nil && sj.DayTypes != nil {
			serviceID := sj.ID
			var assignments []*model.DayTypeAssignment
			for _, id := range sj.DayTypes.DayTypeRef {
				dt := e.netexRepository.GetDayTypeById(id)
				if dt != nil {
					as := e.netexRepository.GetDayTypeAssignmentsByDayType(dt)
					assignments = append(assignments, as...)
				}
			}
			cds, err := e.serviceCalendarDateProducer.Produce(serviceID, assignments)
			if err != nil {
				return err
			}
			for _, cd := range cds {
				if err := e.gtfsRepository.SaveEntity(cd); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// convertTransfers converts NeTEx interchanges to GTFS transfers
func (e *DefaultGtfsExporter) convertTransfers() error {
	interchanges := e.netexRepository.GetServiceJourneyInterchanges()

	for _, interchange := range interchanges {
		transfer, err := e.transferProducer.Produce(interchange)
		if err != nil {
			return err
		}
		if transfer != nil {
			if err := e.gtfsRepository.SaveEntity(transfer); err != nil {
				return err
			}
		}
	}

	return nil
}

// addFeedInfo adds feed information to the GTFS dataset
func (e *DefaultGtfsExporter) addFeedInfo() error {
	if e.feedInfoProducer != nil {
		feedInfo, err := e.feedInfoProducer.ProduceFeedInfo()
		if err != nil {
			return err
		}
		if feedInfo != nil {
			return e.gtfsRepository.SaveEntity(feedInfo)
		}
	}
	return nil
}

func shapeID(s *model.Shape) string {
	if s == nil {
		return ""
	}
	return s.ShapeID
}

// ensureDefaultAgency creates a default agency if none have been added
func (e *DefaultGtfsExporter) ensureDefaultAgency() error {
	// If at least one agency exists, nothing to do
	if e.gtfsRepository.GetDefaultAgency() != nil {
		return nil
	}
	tz := e.netexRepository.GetTimeZone()
	if tz == "" {
		tz = "UTC"
	}
	agency := &model.Agency{
		AgencyID:       "default",
		AgencyName:     "Default Agency",
		AgencyTimezone: tz,
	}
	return e.gtfsRepository.SaveEntity(agency)
}

// Setter methods for extension points
func (e *DefaultGtfsExporter) SetAgencyProducer(producer producer.AgencyProducer) {
	e.agencyProducer = producer
}

func (e *DefaultGtfsExporter) SetRouteProducer(producer producer.RouteProducer) {
	e.routeProducer = producer
}

func (e *DefaultGtfsExporter) SetTripProducer(producer producer.TripProducer) {
	e.tripProducer = producer
}

func (e *DefaultGtfsExporter) SetStopProducer(producer producer.StopProducer) {
	e.stopProducer = producer
}

func (e *DefaultGtfsExporter) SetStopTimeProducer(producer producer.StopTimeProducer) {
	e.stopTimeProducer = producer
}

func (e *DefaultGtfsExporter) SetServiceCalendarProducer(producer producer.ServiceCalendarProducer) {
	e.serviceCalendarProducer = producer
}

func (e *DefaultGtfsExporter) SetServiceCalendarDateProducer(producer producer.ServiceCalendarDateProducer) {
	e.serviceCalendarDateProducer = producer
}

func (e *DefaultGtfsExporter) SetShapeProducer(producer producer.ShapeProducer) {
	e.shapeProducer = producer
}

func (e *DefaultGtfsExporter) SetTransferProducer(producer producer.TransferProducer) {
	e.transferProducer = producer
}

func (e *DefaultGtfsExporter) SetFeedInfoProducer(producer producer.FeedInfoProducer) {
	e.feedInfoProducer = producer
}

// Getter methods for repositories
func (e *DefaultGtfsExporter) GetNetexRepository() producer.NetexRepository {
	return e.netexRepository
}

func (e *DefaultGtfsExporter) GetGtfsRepository() producer.GtfsRepository {
	return e.gtfsRepository
}

func (e *DefaultGtfsExporter) GetStopAreaRepository() producer.StopAreaRepository {
	return e.stopAreaRepository
}
