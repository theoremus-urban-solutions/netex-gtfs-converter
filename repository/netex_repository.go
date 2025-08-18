package repository

import (
	"fmt"
	"sync"

	"github.com/theoremus-urban-solutions/netex-gtfs-converter/model"
	"github.com/theoremus-urban-solutions/netex-gtfs-converter/producer"
)

// DefaultNetexRepository implements NetexRepository
type DefaultNetexRepository struct {
	mu sync.RWMutex
	// Entity storage maps
	authorities                map[string]*model.Authority
	networks                   map[string]*model.Network
	lines                      map[string]*model.Line
	routes                     map[string]*model.Route
	journeyPatterns            map[string]*model.JourneyPattern
	serviceJourneys            map[string]*model.ServiceJourney
	datedServiceJourneys       map[string]*model.DatedServiceJourney
	destinationDisplays        map[string]*model.DestinationDisplay
	scheduledStopPoints        map[string]*model.ScheduledStopPoint
	stopPointInJourneyPatterns map[string]*model.StopPointInJourneyPattern
	serviceJourneyInterchanges map[string]*model.ServiceJourneyInterchange
	dayTypes                   map[string]*model.DayType
	operatingDays              map[string]*model.OperatingDay
	operatingPeriods           map[string]*model.OperatingPeriod
	dayTypeAssignments         map[string]*model.DayTypeAssignment
	stopPlaces                 map[string]*model.StopPlace
	quays                      map[string]*model.Quay
	headwayJourneyGroups       map[string]*model.HeadwayJourneyGroup

	// Lookup maps for efficient querying
	routesByLineId                            map[string][]*model.Route
	serviceJourneysByPattern                  map[string][]*model.ServiceJourney
	datedServiceJourneysByServiceJourney      map[string][]*model.DatedServiceJourney
	dayTypeAssignmentsByDayType               map[string][]*model.DayTypeAssignment
	quaysByStopPlace                          map[string][]*model.Quay
	stopPlaceByQuayId                         map[string]*model.StopPlace
	pointInJourneyPatternToScheduledStopPoint map[string]string
	lineIdToNetworkId                         map[string]string

	// Default timezone
	timeZone string
}

// NewDefaultNetexRepository creates a new DefaultNetexRepository
func NewDefaultNetexRepository() producer.NetexRepository {
	return &DefaultNetexRepository{
		authorities:                make(map[string]*model.Authority),
		networks:                   make(map[string]*model.Network),
		lines:                      make(map[string]*model.Line),
		routes:                     make(map[string]*model.Route),
		journeyPatterns:            make(map[string]*model.JourneyPattern),
		serviceJourneys:            make(map[string]*model.ServiceJourney),
		datedServiceJourneys:       make(map[string]*model.DatedServiceJourney),
		destinationDisplays:        make(map[string]*model.DestinationDisplay),
		scheduledStopPoints:        make(map[string]*model.ScheduledStopPoint),
		stopPointInJourneyPatterns: make(map[string]*model.StopPointInJourneyPattern),
		serviceJourneyInterchanges: make(map[string]*model.ServiceJourneyInterchange),
		dayTypes:                   make(map[string]*model.DayType),
		operatingDays:              make(map[string]*model.OperatingDay),
		operatingPeriods:           make(map[string]*model.OperatingPeriod),
		dayTypeAssignments:         make(map[string]*model.DayTypeAssignment),
		stopPlaces:                 make(map[string]*model.StopPlace),
		quays:                      make(map[string]*model.Quay),
		headwayJourneyGroups:       make(map[string]*model.HeadwayJourneyGroup),

		routesByLineId:                            make(map[string][]*model.Route),
		serviceJourneysByPattern:                  make(map[string][]*model.ServiceJourney),
		datedServiceJourneysByServiceJourney:      make(map[string][]*model.DatedServiceJourney),
		dayTypeAssignmentsByDayType:               make(map[string][]*model.DayTypeAssignment),
		quaysByStopPlace:                          make(map[string][]*model.Quay),
		stopPlaceByQuayId:                         make(map[string]*model.StopPlace),
		pointInJourneyPatternToScheduledStopPoint: make(map[string]string),
		lineIdToNetworkId:                         make(map[string]string),

		timeZone: "Europe/Oslo", // Default timezone
	}
}

// SaveEntity saves an entity to the repository
func (r *DefaultNetexRepository) SaveEntity(entity interface{}) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	switch e := entity.(type) {
	case *model.Authority:
		r.authorities[e.ID] = e
	case *model.Network:
		r.networks[e.ID] = e
		r.buildLineToNetworkMappings(e)
	case *model.Line:
		r.lines[e.ID] = e
	case *model.Route:
		r.routes[e.ID] = e
		r.addToRoutesByLine(e)
	case *model.JourneyPattern:
		r.journeyPatterns[e.ID] = e
		r.buildPointInJourneyPatternMappings(e)
	case *model.ServiceJourney:
		r.serviceJourneys[e.ID] = e
		r.addToServiceJourneysByPattern(e)
	case *model.DatedServiceJourney:
		r.datedServiceJourneys[e.ID] = e
		r.addToDatedServiceJourneysByServiceJourney(e)
	case *model.DestinationDisplay:
		r.destinationDisplays[e.ID] = e
	case *model.ScheduledStopPoint:
		r.scheduledStopPoints[e.ID] = e
	case *model.StopPointInJourneyPattern:
		r.stopPointInJourneyPatterns[e.ID] = e
	case *model.ServiceJourneyInterchange:
		r.serviceJourneyInterchanges[e.ID] = e
	case *model.DayType:
		r.dayTypes[e.ID] = e
	case *model.OperatingDay:
		r.operatingDays[e.ID] = e
	case *model.OperatingPeriod:
		r.operatingPeriods[e.ID] = e
	case *model.DayTypeAssignment:
		r.dayTypeAssignments[e.ID] = e
		r.addToDayTypeAssignmentsByDayType(e)
	case *model.StopPlace:
		r.stopPlaces[e.ID] = e
	case *model.Quay:
		r.quays[e.ID] = e
		r.addQuayToStopPlace(e)
	case *model.HeadwayJourneyGroup:
		r.headwayJourneyGroups[e.ID] = e
	default:
		return fmt.Errorf("unknown entity type: %T", entity)
	}
	return nil
}

// GetLines returns all lines
func (r *DefaultNetexRepository) GetLines() []*model.Line {
	r.mu.RLock()
	defer r.mu.RUnlock()
	lines := make([]*model.Line, 0, len(r.lines))
	for _, line := range r.lines {
		lines = append(lines, line)
	}
	return lines
}

// GetServiceJourneys returns all service journeys
func (r *DefaultNetexRepository) GetServiceJourneys() []*model.ServiceJourney {
	r.mu.RLock()
	defer r.mu.RUnlock()
	journeys := make([]*model.ServiceJourney, 0, len(r.serviceJourneys))
	for _, journey := range r.serviceJourneys {
		journeys = append(journeys, journey)
	}
	return journeys
}

// GetAuthorityById returns an authority by ID
func (r *DefaultNetexRepository) GetAuthorityById(id string) *model.Authority {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.authorities[id]
}

// GetQuayById returns a quay by ID
func (r *DefaultNetexRepository) GetQuayById(id string) *model.Quay {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.quays[id]
}

// GetStopPlaceByQuayId returns the stop place for a given quay ID
func (r *DefaultNetexRepository) GetStopPlaceByQuayId(quayId string) *model.StopPlace {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.stopPlaceByQuayId[quayId]
}

// GetTimeZone returns the default timezone
func (r *DefaultNetexRepository) GetTimeZone() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.timeZone
}

// GetJourneyPatternById returns a journey pattern by ID
func (r *DefaultNetexRepository) GetJourneyPatternById(id string) *model.JourneyPattern {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.journeyPatterns[id]
}

// GetRouteById returns a route by ID
func (r *DefaultNetexRepository) GetRouteById(id string) *model.Route {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.routes[id]
}

// GetRoutesByLine returns all routes for a given line
func (r *DefaultNetexRepository) GetRoutesByLine(line *model.Line) []*model.Route {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if routes, exists := r.routesByLineId[line.ID]; exists {
		return routes
	}
	return []*model.Route{}
}

// GetServiceJourneysByJourneyPattern returns all service journeys for a journey pattern
func (r *DefaultNetexRepository) GetServiceJourneysByJourneyPattern(pattern *model.JourneyPattern) []*model.ServiceJourney {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if journeys, exists := r.serviceJourneysByPattern[pattern.ID]; exists {
		return journeys
	}
	return []*model.ServiceJourney{}
}

// GetServiceJourneyInterchanges returns all service journey interchanges
func (r *DefaultNetexRepository) GetServiceJourneyInterchanges() []*model.ServiceJourneyInterchange {
	r.mu.RLock()
	defer r.mu.RUnlock()
	interchanges := make([]*model.ServiceJourneyInterchange, 0, len(r.serviceJourneyInterchanges))
	for _, interchange := range r.serviceJourneyInterchanges {
		interchanges = append(interchanges, interchange)
	}
	return interchanges
}

// GetDestinationDisplayById returns a destination display by ID
func (r *DefaultNetexRepository) GetDestinationDisplayById(id string) *model.DestinationDisplay {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.destinationDisplays[id]
}

// GetAuthorityIdForLine returns the authority ID for a line
func (r *DefaultNetexRepository) GetAuthorityIdForLine(line *model.Line) string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// First check if line has direct authority reference
	if line.AuthorityRef != "" {
		return line.AuthorityRef
	}

	// If no direct authority, look up through network mapping
	if networkId, exists := r.lineIdToNetworkId[line.ID]; exists {
		if network, networkExists := r.networks[networkId]; networkExists {
			return network.AuthorityRef.Ref
		}
	}

	// Fallback: check if line has NetworkRef (alternative NeTEx structure)
	if line.NetworkRef != "" {
		if network, exists := r.networks[line.NetworkRef]; exists {
			return network.AuthorityRef.Ref
		}
	}

	return ""
}

// GetDatedServiceJourneysByServiceJourneyId returns all dated service journeys for a service journey
func (r *DefaultNetexRepository) GetDatedServiceJourneysByServiceJourneyId(serviceJourneyId string) []*model.DatedServiceJourney {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if journeys, exists := r.datedServiceJourneysByServiceJourney[serviceJourneyId]; exists {
		return journeys
	}
	return []*model.DatedServiceJourney{}
}

// GetDayTypeById returns a day type by ID
func (r *DefaultNetexRepository) GetDayTypeById(id string) *model.DayType {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.dayTypes[id]
}

// GetOperatingDayById returns an operating day by ID
func (r *DefaultNetexRepository) GetOperatingDayById(id string) *model.OperatingDay {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.operatingDays[id]
}

// GetDayTypeAssignmentsByDayType returns all day type assignments for a day type
func (r *DefaultNetexRepository) GetDayTypeAssignmentsByDayType(dayType *model.DayType) []*model.DayTypeAssignment {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if assignments, exists := r.dayTypeAssignmentsByDayType[dayType.ID]; exists {
		return assignments
	}
	return []*model.DayTypeAssignment{}
}

// GetScheduledStopPointRefByPointInJourneyPatternRef returns the scheduled stop point ref for a point in journey pattern ref
func (r *DefaultNetexRepository) GetScheduledStopPointRefByPointInJourneyPatternRef(pjpRef string) string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.pointInJourneyPatternToScheduledStopPoint[pjpRef]
}

// GetStopPointInJourneyPatternById returns a stop point in journey pattern by ID
func (r *DefaultNetexRepository) GetStopPointInJourneyPatternById(id string) *model.StopPointInJourneyPattern {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.stopPointInJourneyPatterns[id]
}

// GetScheduledStopPointById returns a scheduled stop point by ID
func (r *DefaultNetexRepository) GetScheduledStopPointById(id string) *model.ScheduledStopPoint {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.scheduledStopPoints[id]
}

// GetAllStopPlaces returns all stop places
func (r *DefaultNetexRepository) GetAllStopPlaces() []*model.StopPlace {
	r.mu.RLock()
	defer r.mu.RUnlock()
	stopPlaces := make([]*model.StopPlace, 0, len(r.stopPlaces))
	for _, stopPlace := range r.stopPlaces {
		stopPlaces = append(stopPlaces, stopPlace)
	}
	return stopPlaces
}

// GetAllQuays returns all quays
func (r *DefaultNetexRepository) GetAllQuays() []*model.Quay {
	r.mu.RLock()
	defer r.mu.RUnlock()
	quays := make([]*model.Quay, 0, len(r.quays))
	for _, quay := range r.quays {
		quays = append(quays, quay)
	}
	return quays
}

// Helper methods for building lookup maps

func (r *DefaultNetexRepository) addToRoutesByLine(route *model.Route) {
	if route.LineRef.Ref == "" {
		return
	}
	r.routesByLineId[route.LineRef.Ref] = append(r.routesByLineId[route.LineRef.Ref], route)
}

func (r *DefaultNetexRepository) addToServiceJourneysByPattern(journey *model.ServiceJourney) {
	if journey.JourneyPatternRef.Ref == "" {
		return
	}
	r.serviceJourneysByPattern[journey.JourneyPatternRef.Ref] = append(r.serviceJourneysByPattern[journey.JourneyPatternRef.Ref], journey)
}

func (r *DefaultNetexRepository) addToDatedServiceJourneysByServiceJourney(datedJourney *model.DatedServiceJourney) {
	if datedJourney.ServiceJourneyRef == "" {
		return
	}
	r.datedServiceJourneysByServiceJourney[datedJourney.ServiceJourneyRef] = append(r.datedServiceJourneysByServiceJourney[datedJourney.ServiceJourneyRef], datedJourney)
}

func (r *DefaultNetexRepository) addToDayTypeAssignmentsByDayType(assignment *model.DayTypeAssignment) {
	if assignment.DayTypeRef == "" {
		return
	}
	r.dayTypeAssignmentsByDayType[assignment.DayTypeRef] = append(r.dayTypeAssignmentsByDayType[assignment.DayTypeRef], assignment)
}

func (r *DefaultNetexRepository) addQuayToStopPlace(quay *model.Quay) {
	// Find the stop place that contains this quay
	for _, stopPlace := range r.stopPlaces {
		if stopPlace.Quays != nil {
			for _, spQuay := range stopPlace.Quays.Quay {
				if spQuay.ID == quay.ID {
					r.stopPlaceByQuayId[quay.ID] = stopPlace
					r.quaysByStopPlace[stopPlace.ID] = append(r.quaysByStopPlace[stopPlace.ID], quay)
					return
				}
			}
		}
	}
}

func (r *DefaultNetexRepository) buildPointInJourneyPatternMappings(journeyPattern *model.JourneyPattern) {
	if journeyPattern.PointsInSequence == nil {
		return
	}

	// Build mapping from PointInJourneyPatternRef to ScheduledStopPointRef
	for _, pointInterface := range journeyPattern.PointsInSequence.PointInJourneyPatternOrStopPointInJourneyPatternOrTimingPointInJourneyPattern {
		if stopPoint, ok := pointInterface.(*model.StopPointInJourneyPattern); ok {
			r.pointInJourneyPatternToScheduledStopPoint[stopPoint.ID] = stopPoint.ScheduledStopPointRef
		}
	}
}

func (r *DefaultNetexRepository) buildLineToNetworkMappings(network *model.Network) {
	if network.Members == nil {
		return
	}

	// Build mapping from Line ID to Network ID
	for _, lineRef := range network.Members.LineRef {
		if lineRef.Ref != "" {
			r.lineIdToNetworkId[lineRef.Ref] = network.ID
		}
	}
}

// GetHeadwayJourneyGroups returns all headway journey groups
func (r *DefaultNetexRepository) GetHeadwayJourneyGroups() []*model.HeadwayJourneyGroup {
	groups := make([]*model.HeadwayJourneyGroup, 0, len(r.headwayJourneyGroups))
	for _, group := range r.headwayJourneyGroups {
		groups = append(groups, group)
	}
	return groups
}

// GetHeadwayJourneyGroupById returns a headway journey group by ID
func (r *DefaultNetexRepository) GetHeadwayJourneyGroupById(id string) *model.HeadwayJourneyGroup {
	return r.headwayJourneyGroups[id]
}
