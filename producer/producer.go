package producer

import (
	"io"

	"github.com/theoremus-urban-solutions/netex-gtfs-converter/model"
)

// Producer interfaces for converting NeTEx entities to GTFS entities

// AgencyProducer converts NeTEx Authority to GTFS Agency
type AgencyProducer interface {
	Produce(authority *model.Authority) (*model.Agency, error)
}

// RouteProducer converts NeTEx Line to GTFS Route
type RouteProducer interface {
	Produce(line *model.Line) (*model.GtfsRoute, error)
}

// TripProducer converts NeTEx ServiceJourney to GTFS Trip
type TripProducer interface {
	Produce(input TripInput) (*model.Trip, error)
}

// TripInput contains all the data needed to produce a GTFS trip
type TripInput struct {
	ServiceJourney     *model.ServiceJourney
	NetexRoute         *model.Route
	GtfsRoute          *model.GtfsRoute
	ShapeID            string
	DestinationDisplay *model.DestinationDisplay
}

// StopProducer converts NeTEx Quay to GTFS Stop
type StopProducer interface {
	ProduceStopFromQuay(quay *model.Quay) (*model.Stop, error)
	ProduceStopFromStopPlace(stopPlace *model.StopPlace) (*model.Stop, error)
}

// StopTimeProducer converts NeTEx TimetabledPassingTime to GTFS StopTime
type StopTimeProducer interface {
	Produce(input StopTimeInput) (*model.StopTime, error)
}

// StopTimeInput contains all the data needed to produce a GTFS stop time
type StopTimeInput struct {
	TimetabledPassingTime *model.TimetabledPassingTime
	JourneyPattern        *model.JourneyPattern
	Trip                  *model.Trip
	Shape                 *model.Shape
	CurrentHeadSign       string
}

// ServiceCalendarProducer converts NeTEx service patterns to GTFS Calendar
type ServiceCalendarProducer interface {
	Produce(serviceID string, dayTypes []*model.DayType) (*model.Calendar, error)
}

// ServiceCalendarDateProducer converts NeTEx day type assignments to GTFS CalendarDate
type ServiceCalendarDateProducer interface {
	Produce(serviceID string, dayTypeAssignments []*model.DayTypeAssignment) ([]*model.CalendarDate, error)
}

// ShapeProducer converts NeTEx route geometry to GTFS Shape
type ShapeProducer interface {
	Produce(journeyPattern *model.JourneyPattern) (*model.Shape, error)
}

// TransferProducer converts NeTEx ServiceJourneyInterchange to GTFS Transfer
type TransferProducer interface {
	Produce(interchange *model.ServiceJourneyInterchange) (*model.Transfer, error)
}

// FeedInfoProducer creates GTFS FeedInfo
type FeedInfoProducer interface {
	ProduceFeedInfo() (*model.FeedInfo, error)
}

// PathwaysProducer converts NeTEx accessibility data to GTFS pathways and levels
type PathwaysProducer interface {
	ProducePathwaysFromStopPlace(stopPlace *model.StopPlace) ([]*model.Pathway, error)
	ProduceLevelsFromStopPlace(stopPlace *model.StopPlace) ([]*model.Level, error)
	ProduceAccessibilityPathways(from, to *model.Quay) (*model.Pathway, error)
}

// FrequencyProducer converts NeTEx frequency patterns to GTFS frequencies
type FrequencyProducer interface {
	ProduceFromHeadwayJourneyGroup(group *model.HeadwayJourneyGroup) ([]*model.Frequency, error)
	ProduceFromTimeBands(timeBands []model.TimeBand, tripID string) ([]*model.Frequency, error)
	ProduceFrequencyTrip(group *model.HeadwayJourneyGroup, route *model.GtfsRoute) (*model.Trip, error)
	CreateFrequencyBasedService(group *model.HeadwayJourneyGroup, line *model.Line) (*FrequencyService, error)
}

// Repository interfaces for data access

// NetexRepository provides access to NeTEx data
type NetexRepository interface {
	// Data loading
	SaveEntity(entity interface{}) error

	// Data retrieval
	GetLines() []*model.Line
	GetServiceJourneys() []*model.ServiceJourney
	GetAuthorityById(id string) *model.Authority
	GetQuayById(id string) *model.Quay
	GetStopPlaceByQuayId(quayId string) *model.StopPlace
	GetTimeZone() string
	GetJourneyPatternById(id string) *model.JourneyPattern
	GetRouteById(id string) *model.Route
	GetRoutesByLine(line *model.Line) []*model.Route
	GetServiceJourneysByJourneyPattern(pattern *model.JourneyPattern) []*model.ServiceJourney
	GetServiceJourneyInterchanges() []*model.ServiceJourneyInterchange
	GetDestinationDisplayById(id string) *model.DestinationDisplay
	GetAuthorityIdForLine(line *model.Line) string
	GetDatedServiceJourneysByServiceJourneyId(serviceJourneyId string) []*model.DatedServiceJourney
	GetDayTypeById(id string) *model.DayType
	GetOperatingDayById(id string) *model.OperatingDay
	GetDayTypeAssignmentsByDayType(dayType *model.DayType) []*model.DayTypeAssignment
	// Mapping from PointInJourneyPatternRef to ScheduledStopPointRef
	GetScheduledStopPointRefByPointInJourneyPatternRef(pjpRef string) string
	// Optional access to StopPointInJourneyPattern
	GetStopPointInJourneyPatternById(id string) *model.StopPointInJourneyPattern
	// Lookup ScheduledStopPoint by id
	GetScheduledStopPointById(id string) *model.ScheduledStopPoint
	// Access all stop places and quays (for exporting when stop-area repo is empty)
	GetAllStopPlaces() []*model.StopPlace
	GetAllQuays() []*model.Quay
	// Frequency-based services
	GetHeadwayJourneyGroups() []*model.HeadwayJourneyGroup
	GetHeadwayJourneyGroupById(id string) *model.HeadwayJourneyGroup
}

// GtfsRepository provides access to GTFS data
type GtfsRepository interface {
	SaveEntity(entity interface{}) error
	GetAgencyById(id string) *model.Agency
	GetTripById(id string) *model.Trip
	GetStopById(id string) *model.Stop
	GetDefaultAgency() *model.Agency
	WriteGtfs() (io.Reader, error)
}

// StopAreaRepository provides access to stop area data
type StopAreaRepository interface {
	GetQuayById(quayId string) *model.Quay
	GetStopPlaceByQuayId(quayId string) *model.StopPlace
	GetAllQuays() []*model.Quay
	LoadStopAreas(data []byte) error
}

// NetexDatasetLoader loads NeTEx data into a repository
type NetexDatasetLoader interface {
	Load(data io.Reader, repository NetexRepository) error
}

// These constructor functions are implemented in the repository package
// to avoid circular imports. The exporter should import repository directly.
