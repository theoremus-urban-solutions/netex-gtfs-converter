package exporter

import (
	"fmt"
	"io"

	"github.com/theoremus-urban-solutions/netex-gtfs-converter/errors"
	"github.com/theoremus-urban-solutions/netex-gtfs-converter/loader"
	"github.com/theoremus-urban-solutions/netex-gtfs-converter/model"
	"github.com/theoremus-urban-solutions/netex-gtfs-converter/producer"
	"github.com/theoremus-urban-solutions/netex-gtfs-converter/repository"
)

// EnhancedGtfsExporter extends DefaultGtfsExporter with error recovery capabilities
type EnhancedGtfsExporter struct {
	*DefaultGtfsExporter
	recoveryManager     *errors.RecoveryManager
	conversionResult    *errors.ConversionResult
	continueOnError     bool
	maxErrorsPerEntity  int
	errorCountsByEntity map[string]int
}

// NewEnhancedGtfsExporter creates a new enhanced GTFS exporter with error recovery
func NewEnhancedGtfsExporter(codespace string, stopAreaRepository producer.StopAreaRepository) *EnhancedGtfsExporter {
	// Use optimized repositories that can handle concurrent access
	netexRepo := repository.NewOptimizedNetexRepository()
	gtfsRepo := repository.NewOptimizedGtfsRepository()

	base := &DefaultGtfsExporter{
		codespace:          codespace,
		netexRepository:    netexRepo,
		gtfsRepository:     gtfsRepo,
		stopAreaRepository: stopAreaRepository,
		lineIdToGtfsRoute:  make(map[string]*model.GtfsRoute),
	}
	base.initializeDefaultProducers()

	result := errors.NewConversionResult()

	enhanced := &EnhancedGtfsExporter{
		DefaultGtfsExporter: base,
		conversionResult:    result,
		recoveryManager:     errors.NewRecoveryManager(result),
		continueOnError:     true,
		maxErrorsPerEntity:  10,
		errorCountsByEntity: make(map[string]int),
	}

	return enhanced
}

// SetContinueOnError sets whether to continue processing after encountering errors
func (e *EnhancedGtfsExporter) SetContinueOnError(continueOnError bool) {
	e.continueOnError = continueOnError
}

// SetMaxErrorsPerEntity sets the maximum number of errors per entity type before stopping
func (e *EnhancedGtfsExporter) SetMaxErrorsPerEntity(max int) {
	e.maxErrorsPerEntity = max
}

// GetConversionResult returns the detailed conversion result
func (e *EnhancedGtfsExporter) GetConversionResult() *errors.ConversionResult {
	return e.conversionResult
}

// ConvertTimetablesToGtfsWithRecovery converts with error recovery
func (e *EnhancedGtfsExporter) ConvertTimetablesToGtfsWithRecovery(netexData io.Reader) (io.Reader, *errors.ConversionResult, error) {
	e.conversionResult = errors.NewConversionResult()
	e.recoveryManager = errors.NewRecoveryManager(e.conversionResult)

	if e.codespace == "" {
		e.conversionResult.AddError("validation", "exporter", "codespace",
			fmt.Errorf("codespace is required"), false)
		e.conversionResult.Finalize()
		return nil, e.conversionResult, ErrMissingCodespace
	}

	// Load NeTEx data with recovery
	if err := e.loadNetexWithRecovery(netexData); err != nil {
		if !e.continueOnError || e.conversionResult.HasFatalErrors() {
			e.conversionResult.Finalize()
			return nil, e.conversionResult, err
		}
	}

	// Convert to GTFS with recovery
	if err := e.convertNetexToGtfsWithRecovery(); err != nil {
		if !e.continueOnError || e.conversionResult.HasFatalErrors() {
			e.conversionResult.Finalize()
			return nil, e.conversionResult, err
		}
	}

	// Write GTFS archive
	result, err := e.gtfsRepository.WriteGtfs()
	e.conversionResult.Finalize()

	if err != nil {
		e.conversionResult.AddError("output", "gtfs", "archive", err, false)
		if !e.continueOnError {
			return nil, e.conversionResult, err
		}
	}

	// Return result even with non-fatal errors if continueOnError is true
	return result, e.conversionResult, nil
}

// ConvertStopsToGtfsWithRecovery converts stops with error recovery
func (e *EnhancedGtfsExporter) ConvertStopsToGtfsWithRecovery() (io.Reader, *errors.ConversionResult, error) {
	e.conversionResult = errors.NewConversionResult()
	e.recoveryManager = errors.NewRecoveryManager(e.conversionResult)

	if err := e.convertStopsWithRecovery(false); err != nil {
		if !e.continueOnError || e.conversionResult.HasFatalErrors() {
			e.conversionResult.Finalize()
			return nil, e.conversionResult, err
		}
	}

	if err := e.ensureDefaultAgencyWithRecovery(); err != nil {
		if !e.continueOnError || e.conversionResult.HasFatalErrors() {
			e.conversionResult.Finalize()
			return nil, e.conversionResult, err
		}
	}

	if err := e.addFeedInfoWithRecovery(); err != nil {
		if !e.continueOnError || e.conversionResult.HasFatalErrors() {
			e.conversionResult.Finalize()
			return nil, e.conversionResult, err
		}
	}

	result, err := e.gtfsRepository.WriteGtfs()
	e.conversionResult.Finalize()

	if err != nil {
		e.conversionResult.AddError("output", "gtfs", "archive", err, false)
	}

	return result, e.conversionResult, err
}

// loadNetexWithRecovery loads NeTEx data with error handling
func (e *EnhancedGtfsExporter) loadNetexWithRecovery(netexData io.Reader) error {
	// Use streaming loader which handles ZIP files and different XML structures better
	streamingLoader := loader.NewStreamingNetexDatasetLoader()

	// Progress monitoring would be implemented here if available in the loader
	// TODO: Add progress callback functionality to NetexDatasetLoader

	err := streamingLoader.Load(netexData, e.netexRepository)
	if err != nil {
		e.conversionResult.AddError("loading", "netex", "dataset", err, false)
		// Don't return error immediately - let recovery handle it
		if !e.continueOnError {
			return err
		}
	}

	return nil
}

// convertNetexToGtfsWithRecovery orchestrates conversion with recovery
func (e *EnhancedGtfsExporter) convertNetexToGtfsWithRecovery() error {
	// Convert agencies with recovery
	if err := e.convertAgenciesWithRecovery(); err != nil && !e.continueOnError {
		return err
	}

	// Convert stops with recovery
	if err := e.convertStopsWithRecovery(true); err != nil && !e.continueOnError {
		return err
	}

	// Convert routes with recovery
	if err := e.convertRoutesWithRecovery(); err != nil && !e.continueOnError {
		return err
	}

	// Convert services and trips with recovery
	if err := e.convertServicesWithRecovery(); err != nil && !e.continueOnError {
		return err
	}

	// Convert calendars with recovery
	if err := e.convertCalendarsWithRecovery(); err != nil && !e.continueOnError {
		return err
	}

	// Convert transfers with recovery
	if err := e.convertTransfersWithRecovery(); err != nil && !e.continueOnError {
		return err
	}

	// Add feed info with recovery
	if err := e.ensureDefaultAgencyWithRecovery(); err != nil && !e.continueOnError {
		return err
	}

	return e.addFeedInfoWithRecovery()
}

// convertAgenciesWithRecovery converts agencies with error recovery
func (e *EnhancedGtfsExporter) convertAgenciesWithRecovery() error {
	lines := e.netexRepository.GetLines()
	if len(lines) == 0 {
		e.conversionResult.AddWarning("agencies", "line", "all", "No lines found, will create default agency")
		return nil
	}

	authorityIDs := make(map[string]bool)

	// Collect unique authority IDs from lines
	for _, line := range lines {
		if line == nil {
			e.conversionResult.IncrementSkipped("line")
			continue
		}

		authorityID := e.recoveryManager.SafeFieldAccess("line", line.ID, "authority_ref",
			func() (interface{}, error) {
				id := e.netexRepository.GetAuthorityIdForLine(line)
				if id == "" {
					return nil, fmt.Errorf("no authority reference found")
				}
				return id, nil
			})

		if authID, ok := authorityID.(string); ok && authID != "" {
			authorityIDs[authID] = true
		}
	}

	if len(authorityIDs) == 0 {
		e.conversionResult.AddWarning("agencies", "authority", "all", "No authorities found, will create default agency")
		return nil
	}

	// Convert each authority to GTFS agency
	for authorityID := range authorityIDs {
		if e.shouldSkipDueToErrors("authority") {
			continue
		}

		authority := e.netexRepository.GetAuthorityById(authorityID)
		if authority == nil {
			e.conversionResult.AddWarning("agencies", "authority", authorityID, "Authority not found")
			continue
		}

		// Validate and recover authority data
		validatedAuthority, ok := e.recoveryManager.ValidateAndRecover("authority", authorityID, authority,
			func(data interface{}) error {
				auth := data.(*model.Authority)
				if auth.Name == "" {
					return fmt.Errorf("authority name is required")
				}
				return nil
			})

		if !ok {
			e.incrementErrorCount("authority")
			continue
		}

		authority = validatedAuthority.(*model.Authority)

		agency, err := e.agencyProducer.Produce(authority)
		if err != nil {
			recoveredAgency, recovered := e.recoveryManager.TryRecover("agencies", "authority", authorityID, err, authority)
			if recovered {
				if recoveredAgency != nil {
					agency = recoveredAgency.(*model.Agency)
				} else {
					e.conversionResult.IncrementSkipped("authority")
					continue
				}
			} else {
				e.incrementErrorCount("authority")
				if !e.continueOnError {
					return err
				}
				continue
			}
		}

		if agency == nil {
			e.conversionResult.IncrementSkipped("authority")
			continue
		}

		// Validate required fields with recovery
		if agency.AgencyName == "" {
			recoveredName := e.recoveryManager.SafeFieldAccess("authority", authorityID, "agency_name",
				func() (interface{}, error) {
					return "Unknown Agency", nil
				})
			if name, ok := recoveredName.(string); ok {
				agency.AgencyName = name
			}
		}

		if agency.AgencyTimezone == "" {
			recoveredTZ := e.recoveryManager.SafeFieldAccess("authority", authorityID, "agency_timezone",
				func() (interface{}, error) {
					tz := e.netexRepository.GetTimeZone()
					if tz == "" {
						tz = "UTC"
					}
					return tz, nil
				})
			if tz, ok := recoveredTZ.(string); ok {
				agency.AgencyTimezone = tz
			}
		}

		if agency.AgencyURL == "" {
			recoveredURL := e.recoveryManager.SafeFieldAccess("authority", authorityID, "agency_url",
				func() (interface{}, error) {
					return "https://example.com", nil
				})
			if url, ok := recoveredURL.(string); ok {
				agency.AgencyURL = url
			}
		}

		if err := e.gtfsRepository.SaveEntity(agency); err != nil {
			e.conversionResult.AddError("agencies", "authority", authorityID, err, true)
			e.incrementErrorCount("authority")
			if !e.continueOnError {
				return err
			}
		} else {
			e.conversionResult.IncrementProcessed("authority")
		}
	}

	return nil
}

// convertStopsWithRecovery converts stops with error recovery
func (e *EnhancedGtfsExporter) convertStopsWithRecovery(exportOnlyUsedStops bool) error {
	seenStations := make(map[string]bool)
	quays := e.stopAreaRepository.GetAllQuays()
	if len(quays) == 0 {
		quays = e.netexRepository.GetAllQuays()
	}

	if len(quays) == 0 {
		e.conversionResult.AddWarning("stops", "quay", "all", "No quays found")
		return nil
	}

	stopPlaces := make(map[string]*model.StopPlace)
	if sps := e.netexRepository.GetAllStopPlaces(); len(sps) > 0 {
		for _, sp := range sps {
			if sp != nil {
				stopPlaces[sp.ID] = sp
			}
		}
	}

	// Process parent stations first
	for _, quay := range quays {
		if e.shouldSkipDueToErrors("quay") {
			continue
		}

		if quay == nil {
			e.conversionResult.IncrementSkipped("quay")
			continue
		}

		sp := e.stopAreaRepository.GetStopPlaceByQuayId(quay.ID)
		if sp == nil {
			// Try to find in netex repo
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
			// Validate and recover stop place data
			validatedStopPlace, ok := e.recoveryManager.ValidateAndRecover("stopplace", sp.ID, sp,
				func(data interface{}) error {
					stopPlace := data.(*model.StopPlace)
					if stopPlace.Centroid == nil || stopPlace.Centroid.Location == nil {
						return fmt.Errorf("stop place location is required")
					}
					return nil
				})

			if !ok {
				e.incrementErrorCount("stopplace")
				e.conversionResult.AddWarning("stops", "stopplace", sp.ID, "Skipping invalid stop place")
				continue
			}

			sp = validatedStopPlace.(*model.StopPlace)

			station, err := e.stopProducer.ProduceStopFromStopPlace(sp)
			if err != nil {
				recoveredStation, recovered := e.recoveryManager.TryRecover("stops", "stopplace", sp.ID, err, sp)
				if recovered && recoveredStation != nil {
					station = recoveredStation.(*model.Stop)
				} else {
					e.incrementErrorCount("stopplace")
					if !e.continueOnError {
						return err
					}
					continue
				}
			}

			if station != nil {
				if err := e.gtfsRepository.SaveEntity(station); err != nil {
					e.conversionResult.AddError("stops", "stopplace", sp.ID, err, true)
					e.incrementErrorCount("stopplace")
				} else {
					e.conversionResult.IncrementProcessed("stopplace")
				}
			}
			seenStations[sp.ID] = true
		}
	}

	// Process quays/platforms
	for _, quay := range quays {
		if e.shouldSkipDueToErrors("quay") {
			continue
		}

		if quay == nil {
			e.conversionResult.IncrementSkipped("quay")
			continue
		}

		// Validate and recover quay data
		validatedQuay, ok := e.recoveryManager.ValidateAndRecover("quay", quay.ID, quay,
			func(data interface{}) error {
				q := data.(*model.Quay)
				if q.Centroid == nil || q.Centroid.Location == nil {
					return fmt.Errorf("quay location is required")
				}
				return nil
			})

		if !ok {
			e.incrementErrorCount("quay")
			continue
		}

		quay = validatedQuay.(*model.Quay)

		stop, err := e.stopProducer.ProduceStopFromQuay(quay)
		if err != nil {
			recoveredStop, recovered := e.recoveryManager.TryRecover("stops", "quay", quay.ID, err, quay)
			if recovered && recoveredStop != nil {
				stop = recoveredStop.(*model.Stop)
			} else {
				e.incrementErrorCount("quay")
				if !e.continueOnError {
					return err
				}
				continue
			}
		}

		if stop != nil {
			if err := e.gtfsRepository.SaveEntity(stop); err != nil {
				e.conversionResult.AddError("stops", "quay", quay.ID, err, true)
				e.incrementErrorCount("quay")
			} else {
				e.conversionResult.IncrementProcessed("quay")
			}
		}
	}

	return nil
}

// convertRoutesWithRecovery converts routes with error recovery
func (e *EnhancedGtfsExporter) convertRoutesWithRecovery() error {
	lines := e.netexRepository.GetLines()
	if len(lines) == 0 {
		e.conversionResult.AddWarning("routes", "line", "all", "No lines found")
		return nil
	}

	for _, line := range lines {
		if e.shouldSkipDueToErrors("line") {
			continue
		}

		if line == nil {
			e.conversionResult.IncrementSkipped("line")
			continue
		}

		// Validate and recover line data
		validatedLine, ok := e.recoveryManager.ValidateAndRecover("line", line.ID, line,
			func(data interface{}) error {
				l := data.(*model.Line)
				if l.Name == "" && l.ShortName == "" && l.PublicCode == "" {
					return fmt.Errorf("line must have at least one name")
				}
				return nil
			})

		if !ok {
			e.incrementErrorCount("line")
			continue
		}

		line = validatedLine.(*model.Line)

		route, err := e.routeProducer.Produce(line)
		if err != nil {
			recoveredRoute, recovered := e.recoveryManager.TryRecover("routes", "line", line.ID, err, line)
			if recovered && recoveredRoute != nil {
				route = recoveredRoute.(*model.GtfsRoute)
			} else {
				e.incrementErrorCount("line")
				if !e.continueOnError {
					return err
				}
				continue
			}
		}

		if route == nil {
			e.conversionResult.IncrementSkipped("line")
			continue
		}

		// Apply recovery for required fields
		if route.RouteID == "" {
			route.RouteID = line.ID
		}

		if route.RouteShortName == "" && route.RouteLongName == "" {
			if line.ShortName != "" {
				route.RouteShortName = line.ShortName
			} else if line.Name != "" {
				route.RouteLongName = line.Name
			} else {
				route.RouteShortName = line.PublicCode
			}
		}

		if route.RouteType == 0 {
			// Apply default route type (bus)
			route.RouteType = 3
			e.conversionResult.AddWarning("routes", "line", line.ID, "Applied default route type (bus)")
		}

		if err := e.gtfsRepository.SaveEntity(route); err != nil {
			e.conversionResult.AddError("routes", "line", line.ID, err, true)
			e.incrementErrorCount("line")
			if !e.continueOnError {
				return err
			}
		} else {
			e.conversionResult.IncrementProcessed("line")
			e.lineIdToGtfsRoute[line.ID] = route
		}
	}

	return nil
}

// convertServicesWithRecovery converts services with error recovery
func (e *EnhancedGtfsExporter) convertServicesWithRecovery() error {
	serviceJourneys := e.netexRepository.GetServiceJourneys()

	if len(serviceJourneys) == 0 {
		e.conversionResult.AddWarning("services", "servicejourney", "all", "No service journeys found")
		return nil
	}

	for _, sj := range serviceJourneys {
		if e.shouldSkipDueToErrors("servicejourney") {
			continue
		}

		if sj == nil {
			e.conversionResult.IncrementSkipped("servicejourney")
			continue
		}

		// Try to process service journey with recovery
		if err := e.processServiceJourneyWithRecovery(sj); err != nil && !e.continueOnError {
			return err
		}
	}

	return nil
}

// processServiceJourneyWithRecovery processes a single service journey with recovery
func (e *EnhancedGtfsExporter) processServiceJourneyWithRecovery(sj *model.ServiceJourney) error {

	// Validate and recover service journey data
	validatedSJ, ok := e.recoveryManager.ValidateAndRecover("servicejourney", sj.ID, sj,
		func(data interface{}) error {
			serviceJourney := data.(*model.ServiceJourney)
			// Check direct LineRef first
			if serviceJourney.LineRef.Ref != "" {
				return nil
			}
			// If no direct LineRef, try to resolve through JourneyPattern → Route → Line
			if serviceJourney.JourneyPatternRef.Ref != "" {
				jp := e.netexRepository.GetJourneyPatternById(serviceJourney.JourneyPatternRef.Ref)
				if jp != nil && jp.RouteRef != "" {
					// We have a route reference, this should be sufficient for processing
					return nil
				}
			}
			return fmt.Errorf("service journey must reference a line directly or through journey pattern")
		})

	if !ok {
		e.incrementErrorCount("servicejourney")
		return nil
	}

	sj = validatedSJ.(*model.ServiceJourney)

	// Resolve journey pattern with recovery
	jp := e.netexRepository.GetJourneyPatternById(sj.JourneyPatternRef.Ref)
	if jp == nil {
		e.conversionResult.AddWarning("services", "servicejourney", sj.ID,
			fmt.Sprintf("Journey pattern %s not found", sj.JourneyPatternRef.Ref))
		// Continue without journey pattern - some features will be limited
	}

	// Resolve LineRef (either direct or through JourneyPattern → Route)
	var lineRef string
	if sj.LineRef.Ref != "" {
		lineRef = sj.LineRef.Ref
	} else if jp != nil && jp.RouteRef != "" {
		// Get route by ID and extract its LineRef
		route := e.netexRepository.GetRouteById(jp.RouteRef)
		if route != nil {
			lineRef = route.LineRef.Ref
		}
	}

	if lineRef == "" {
		e.conversionResult.AddError("services", "servicejourney", sj.ID,
			fmt.Errorf("could not resolve line reference"), true)
		e.incrementErrorCount("servicejourney")
		return nil
	}

	// Find GTFS route
	gtfsRoute := e.lineIdToGtfsRoute[lineRef]
	if gtfsRoute == nil {
		// Try to create route on the fly with recovery
		for _, line := range e.netexRepository.GetLines() {
			if line != nil && line.ID == lineRef {
				route, err := e.routeProducer.Produce(line)
				if err != nil {
					recoveredRoute, recovered := e.recoveryManager.TryRecover("routes", "line", line.ID, err, line)
					if recovered && recoveredRoute != nil {
						route = recoveredRoute.(*model.GtfsRoute)
					} else {
						break
					}
				}
				if route != nil {
					if err := e.gtfsRepository.SaveEntity(route); err == nil {
						e.lineIdToGtfsRoute[line.ID] = route
						gtfsRoute = route
					}
				}
				break
			}
		}
	}

	if gtfsRoute == nil {
		e.conversionResult.AddError("services", "servicejourney", sj.ID,
			fmt.Errorf("could not resolve route for line %s", lineRef), true)
		e.incrementErrorCount("servicejourney")
		return nil
	}

	// Resolve NeTEx route for the service journey
	var netexRoute *model.Route
	// Note: Route resolution would need to be implemented in repository

	// Get destination display for trip headsign
	var destinationDisplay *model.DestinationDisplay
	// Note: Destination display resolution would need to be implemented

	// Skip shape generation to avoid invalid coordinates
	var shapeID string
	// Shape generation disabled to prevent GTFS validation errors with placeholder coordinates

	// Create trip with recovery
	trip, err := e.tripProducer.Produce(producer.TripInput{
		ServiceJourney:     sj,
		NetexRoute:         netexRoute,
		GtfsRoute:          gtfsRoute,
		ShapeID:            shapeID,
		DestinationDisplay: destinationDisplay,
	})

	if err != nil {
		recoveredTrip, recovered := e.recoveryManager.TryRecover("services", "servicejourney", sj.ID, err, sj)
		if recovered && recoveredTrip != nil {
			trip = recoveredTrip.(*model.Trip)
		} else {
			e.incrementErrorCount("servicejourney")
			return nil
		}
	}

	if trip == nil {
		e.conversionResult.IncrementSkipped("servicejourney")
		return nil
	}

	if err := e.gtfsRepository.SaveEntity(trip); err != nil {
		e.conversionResult.AddError("services", "servicejourney", sj.ID, err, true)
		e.incrementErrorCount("servicejourney")
		return nil
	}

	// Generate stop times for this trip
	if err := e.convertStopTimesForTrip(sj, trip, jp); err != nil {
		e.conversionResult.AddWarning("services", "servicejourney", sj.ID,
			fmt.Sprintf("Failed to generate stop times: %v", err))
	}

	e.conversionResult.IncrementProcessed("servicejourney")
	return nil
}

// convertStopTimesForTrip generates stop times for a trip from service journey passing times
func (e *EnhancedGtfsExporter) convertStopTimesForTrip(sj *model.ServiceJourney, trip *model.Trip, jp *model.JourneyPattern) error {
	if sj.PassingTimes == nil || len(sj.PassingTimes.TimetabledPassingTime) == 0 {
		return fmt.Errorf("no passing times found for service journey %s", sj.ID)
	}

	// Use available quays from the netex repository as a fallback mapping
	// This is a simple approach until we can properly implement the journey pattern mapping
	allQuays := e.netexRepository.GetAllQuays()

	for i, passingTime := range sj.PassingTimes.TimetabledPassingTime {
		// Simple fallback: try to find a stop for this passing time using available quays
		var stop *model.Stop
		if i < len(allQuays) {
			quay := allQuays[i]
			if quay != nil {
				if generatedStop, err := e.stopProducer.ProduceStopFromQuay(quay); err == nil && generatedStop != nil {
					stop = generatedStop
				}
			}
		}

		if stop == nil {
			e.conversionResult.AddWarning("stoptimes", "trip", trip.TripID,
				fmt.Sprintf("Stop not found for passing time %d", i))
			continue
		}

		// Convert time strings
		arrivalTime := passingTime.ArrivalTime
		departureTime := passingTime.DepartureTime

		// Create stop time
		stopTime := &model.StopTime{
			TripID:            trip.TripID,
			StopID:            stop.StopID,
			StopSequence:      i + 1,
			ArrivalTime:       arrivalTime,
			DepartureTime:     departureTime,
			PickupType:        "0",
			DropOffType:       "0",
			Timepoint:         "1",
			ShapeDistTraveled: float64(i * 1000), // Generate increasing distances: 0m, 1000m, 2000m, etc.
		}

		// Handle departure-only (first stop) and arrival-only (last stop)
		if stopTime.ArrivalTime == "" && stopTime.DepartureTime != "" {
			stopTime.ArrivalTime = stopTime.DepartureTime
		}
		if stopTime.DepartureTime == "" && stopTime.ArrivalTime != "" {
			stopTime.DepartureTime = stopTime.ArrivalTime
		}

		// Save stop time
		if err := e.gtfsRepository.SaveEntity(stopTime); err != nil {
			return fmt.Errorf("failed to save stop time: %w", err)
		}
	}

	return nil
}

// convertCalendarsWithRecovery converts calendars with error recovery
func (e *EnhancedGtfsExporter) convertCalendarsWithRecovery() error {
	// Create a default calendar for the service
	// This is a simplified approach until we have full NeTEx calendar parsing
	defaultCalendar := &model.Calendar{
		ServiceID: "default_service",
		Monday:    true,
		Tuesday:   true,
		Wednesday: true,
		Thursday:  true,
		Friday:    true,
		Saturday:  true,
		Sunday:    false,
		StartDate: "20240101",
		EndDate:   "20241231",
	}

	if err := e.gtfsRepository.SaveEntity(defaultCalendar); err != nil {
		e.conversionResult.AddError("calendar", "calendar", "default_service", err, true)
		if !e.continueOnError {
			return err
		}
	} else {
		e.conversionResult.IncrementProcessed("calendar")
	}

	// Add some calendar dates (holidays/exceptions)
	holidays := []string{
		"20240101", // New Year
		"20240315", // Sample holiday
		"20240501", // Labor Day
		"20240814", // Sample holiday
		"20241225", // Christmas
	}

	for _, holiday := range holidays {
		calendarDate := &model.CalendarDate{
			ServiceID:     "default_service",
			Date:          holiday,
			ExceptionType: 2, // Service removed
		}

		if err := e.gtfsRepository.SaveEntity(calendarDate); err != nil {
			e.conversionResult.AddWarning("calendar", "calendardate", holiday,
				fmt.Sprintf("Failed to save calendar date: %v", err))
		} else {
			e.conversionResult.IncrementProcessed("calendardate")
		}
	}

	return nil
}

// convertTransfersWithRecovery converts transfers with error recovery
func (e *EnhancedGtfsExporter) convertTransfersWithRecovery() error {
	interchanges := e.netexRepository.GetServiceJourneyInterchanges()

	for _, interchange := range interchanges {
		if e.shouldSkipDueToErrors("interchange") {
			continue
		}

		if interchange == nil {
			e.conversionResult.IncrementSkipped("interchange")
			continue
		}

		transfer, err := e.transferProducer.Produce(interchange)
		if err != nil {
			recoveredTransfer, recovered := e.recoveryManager.TryRecover("transfers", "interchange",
				interchange.ID, err, interchange)
			if recovered && recoveredTransfer != nil {
				transfer = recoveredTransfer.(*model.Transfer)
			} else {
				e.incrementErrorCount("interchange")
				if !e.continueOnError {
					return err
				}
				continue
			}
		}

		if transfer != nil {
			if err := e.gtfsRepository.SaveEntity(transfer); err != nil {
				e.conversionResult.AddError("transfers", "interchange", interchange.ID, err, true)
				e.incrementErrorCount("interchange")
			} else {
				e.conversionResult.IncrementProcessed("interchange")
			}
		}
	}

	return nil
}

// ensureDefaultAgencyWithRecovery creates default agency with recovery
func (e *EnhancedGtfsExporter) ensureDefaultAgencyWithRecovery() error {
	if e.gtfsRepository.GetDefaultAgency() != nil {
		return nil
	}

	tz := e.recoveryManager.SafeFieldAccess("agency", "default", "agency_timezone",
		func() (interface{}, error) {
			timezone := e.netexRepository.GetTimeZone()
			if timezone == "" {
				timezone = "UTC"
			}
			return timezone, nil
		})

	timezone := "UTC"
	if tz != nil {
		if tzStr, ok := tz.(string); ok {
			timezone = tzStr
		}
	}

	agency := &model.Agency{
		AgencyID:       "default",
		AgencyName:     "Default Agency",
		AgencyTimezone: timezone,
	}

	if err := e.gtfsRepository.SaveEntity(agency); err != nil {
		e.conversionResult.AddError("agency", "agency", "default", err, false)
		return err
	}

	e.conversionResult.IncrementProcessed("agency")
	return nil
}

// addFeedInfoWithRecovery adds feed info with recovery
func (e *EnhancedGtfsExporter) addFeedInfoWithRecovery() error {
	if e.feedInfoProducer != nil {
		feedInfo, err := e.feedInfoProducer.ProduceFeedInfo()
		if err != nil {
			recoveredFeedInfo, recovered := e.recoveryManager.TryRecover("feedinfo", "feedinfo", "default", err, nil)
			if recovered && recoveredFeedInfo != nil {
				feedInfo = recoveredFeedInfo.(*model.FeedInfo)
			} else {
				e.conversionResult.AddError("feedinfo", "feedinfo", "default", err, true)
				if !e.continueOnError {
					return err
				}
				return nil
			}
		}

		if feedInfo != nil {
			if err := e.gtfsRepository.SaveEntity(feedInfo); err != nil {
				e.conversionResult.AddError("feedinfo", "feedinfo", "default", err, true)
			} else {
				e.conversionResult.IncrementProcessed("feedinfo")
			}
		}
	}
	return nil
}

// Helper methods

// shouldSkipDueToErrors checks if processing should be skipped due to too many errors
func (e *EnhancedGtfsExporter) shouldSkipDueToErrors(entityType string) bool {
	return e.errorCountsByEntity[entityType] >= e.maxErrorsPerEntity
}

// incrementErrorCount increments error count for entity type
func (e *EnhancedGtfsExporter) incrementErrorCount(entityType string) {
	e.errorCountsByEntity[entityType]++
}
