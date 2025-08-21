package producer

import (
	"io"
	"testing"

	"github.com/theoremus-urban-solutions/netex-gtfs-converter/model"
)

// Mock implementations for testing

type mockNetexRepository struct {
	timeZone string
}

func (m *mockNetexRepository) SaveEntity(entity interface{}) error                 { return nil }
func (m *mockNetexRepository) GetLines() []*model.Line                             { return nil }
func (m *mockNetexRepository) GetServiceJourneys() []*model.ServiceJourney         { return nil }
func (m *mockNetexRepository) GetAuthorityById(id string) *model.Authority         { return nil }
func (m *mockNetexRepository) GetQuayById(id string) *model.Quay                   { return nil }
func (m *mockNetexRepository) GetStopPlaceByQuayId(quayId string) *model.StopPlace { return nil }
func (m *mockNetexRepository) GetTimeZone() string {
	if m.timeZone != "" {
		return m.timeZone
	}
	return "UTC"
}
func (m *mockNetexRepository) GetJourneyPatternById(id string) *model.JourneyPattern { return nil }
func (m *mockNetexRepository) GetRouteById(id string) *model.Route                   { return nil }
func (m *mockNetexRepository) GetRoutesByLine(line *model.Line) []*model.Route       { return nil }
func (m *mockNetexRepository) GetServiceJourneysByJourneyPattern(pattern *model.JourneyPattern) []*model.ServiceJourney {
	return nil
}
func (m *mockNetexRepository) GetServiceJourneyInterchanges() []*model.ServiceJourneyInterchange {
	return nil
}
func (m *mockNetexRepository) GetDestinationDisplayById(id string) *model.DestinationDisplay {
	return nil
}
func (m *mockNetexRepository) GetAuthorityIdForLine(line *model.Line) string {
	return line.AuthorityRef
}
func (m *mockNetexRepository) GetDatedServiceJourneysByServiceJourneyId(serviceJourneyId string) []*model.DatedServiceJourney {
	return nil
}
func (m *mockNetexRepository) GetDayTypeById(id string) *model.DayType           { return nil }
func (m *mockNetexRepository) GetOperatingDayById(id string) *model.OperatingDay { return nil }
func (m *mockNetexRepository) GetDayTypeAssignmentsByDayType(dayType *model.DayType) []*model.DayTypeAssignment {
	return nil
}
func (m *mockNetexRepository) GetScheduledStopPointRefByPointInJourneyPatternRef(pjpRef string) string {
	return ""
}
func (m *mockNetexRepository) GetStopPointInJourneyPatternById(id string) *model.StopPointInJourneyPattern {
	return nil
}
func (m *mockNetexRepository) GetScheduledStopPointById(id string) *model.ScheduledStopPoint {
	return nil
}
func (m *mockNetexRepository) GetAllStopPlaces() []*model.StopPlace                  { return nil }
func (m *mockNetexRepository) GetAllQuays() []*model.Quay                            { return nil }
func (m *mockNetexRepository) GetHeadwayJourneyGroups() []*model.HeadwayJourneyGroup { return nil }
func (m *mockNetexRepository) GetHeadwayJourneyGroupById(id string) *model.HeadwayJourneyGroup {
	return nil
}

type mockGtfsRepository struct{}

func (m *mockGtfsRepository) SaveEntity(entity interface{}) error   { return nil }
func (m *mockGtfsRepository) GetAgencyById(id string) *model.Agency { return nil }
func (m *mockGtfsRepository) GetTripById(id string) *model.Trip     { return nil }
func (m *mockGtfsRepository) GetStopById(id string) *model.Stop     { return nil }
func (m *mockGtfsRepository) GetDefaultAgency() *model.Agency       { return nil }
func (m *mockGtfsRepository) WriteGtfs() (io.Reader, error)         { return nil, nil }

type mockStopAreaRepository struct{}

func (m *mockStopAreaRepository) GetQuayById(quayId string) *model.Quay               { return nil }
func (m *mockStopAreaRepository) GetStopPlaceByQuayId(quayId string) *model.StopPlace { return nil }
func (m *mockStopAreaRepository) GetAllQuays() []*model.Quay                          { return nil }
func (m *mockStopAreaRepository) LoadStopAreas(data []byte) error                     { return nil }

func TestDefaultAgencyProducer_Produce(t *testing.T) {
	netexRepo := &mockNetexRepository{}
	producer := NewDefaultAgencyProducer(netexRepo)

	authority := &model.Authority{
		ID:        "test-authority",
		Name:      "Test Authority",
		URL:       "https://example.com",
		ShortName: "TA",
		ContactDetails: &model.ContactDetails{
			Phone: "+1234567890",
			Email: "test@example.com",
		},
	}

	agency, err := producer.Produce(authority)
	if err != nil {
		t.Fatalf("Produce() failed: %v", err)
	}

	if agency == nil {
		t.Fatal("Produce() returned nil agency")
	}

	if agency.AgencyID != authority.ID {
		t.Errorf("Expected AgencyID %s, got %s", authority.ID, agency.AgencyID)
	}

	if agency.AgencyName != authority.Name {
		t.Errorf("Expected AgencyName %s, got %s", authority.Name, agency.AgencyName)
	}

	if agency.AgencyURL != authority.URL {
		t.Errorf("Expected AgencyURL %s, got %s", authority.URL, agency.AgencyURL)
	}

	if agency.AgencyPhone != authority.ContactDetails.Phone {
		t.Errorf("Expected AgencyPhone %s, got %s", authority.ContactDetails.Phone, agency.AgencyPhone)
	}
}

func TestDefaultRouteProducer_Produce(t *testing.T) {
	netexRepo := &mockNetexRepository{}
	gtfsRepo := &mockGtfsRepository{}
	producer := NewDefaultRouteProducer(netexRepo, gtfsRepo)

	line := &model.Line{
		ID:               "test-line",
		Name:             "Test Line",
		ShortName:        "TL",
		PublicCode:       "1",
		TransportMode:    "bus",
		TransportSubmode: "localBus",
		AuthorityRef:     "test-authority",
		Presentation: &model.Presentation{
			Colour:     "FF0000",
			TextColour: "FFFFFF",
		},
	}

	route, err := producer.Produce(line)
	if err != nil {
		t.Fatalf("Produce() failed: %v", err)
	}

	if route == nil {
		t.Fatal("Produce() returned nil route")
	}

	if route.RouteID != line.ID {
		t.Errorf("Expected RouteID %s, got %s", line.ID, route.RouteID)
	}

	if route.RouteShortName != line.PublicCode {
		t.Errorf("Expected RouteShortName %s, got %s", line.PublicCode, route.RouteShortName)
	}

	if route.RouteLongName != line.Name {
		t.Errorf("Expected RouteLongName %s, got %s", line.Name, route.RouteLongName)
	}

	// Check route type mapping - should use basic GTFS types now
	expectedRouteType := model.Bus.Value() // 3 instead of 704
	if route.RouteType != expectedRouteType {
		t.Errorf("Expected RouteType %d, got %d", expectedRouteType, route.RouteType)
	}

	if route.RouteColor != line.Presentation.Colour {
		t.Errorf("Expected RouteColor %s, got %s", line.Presentation.Colour, route.RouteColor)
	}
}

func TestDefaultStopProducer_ProduceStopFromQuay(t *testing.T) {
	stopAreaRepo := &mockStopAreaRepository{}
	gtfsRepo := &mockGtfsRepository{}
	producer := NewDefaultStopProducer(stopAreaRepo, gtfsRepo)

	quay := &model.Quay{
		ID:         "test-quay",
		Name:       "Test Quay",
		PublicCode: "A1",
		Centroid: &model.Centroid{
			Location: &model.Location{
				Latitude:  59.9139,
				Longitude: 10.7522,
			},
		},
	}

	stop, err := producer.ProduceStopFromQuay(quay)
	if err != nil {
		t.Fatalf("ProduceStopFromQuay() failed: %v", err)
	}

	if stop == nil {
		t.Fatal("ProduceStopFromQuay() returned nil stop")
	}

	if stop.StopID != quay.ID {
		t.Errorf("Expected StopID %s, got %s", quay.ID, stop.StopID)
	}

	if stop.StopName != quay.Name {
		t.Errorf("Expected StopName %s, got %s", quay.Name, stop.StopName)
	}

	if stop.StopCode != quay.PublicCode {
		t.Errorf("Expected StopCode %s, got %s", quay.PublicCode, stop.StopCode)
	}

	if stop.StopLat != quay.Centroid.Location.Latitude {
		t.Errorf("Expected StopLat %f, got %f", quay.Centroid.Location.Latitude, stop.StopLat)
	}

	if stop.StopLon != quay.Centroid.Location.Longitude {
		t.Errorf("Expected StopLon %f, got %f", quay.Centroid.Location.Longitude, stop.StopLon)
	}

	// Should be a platform (location_type = 0)
	if stop.LocationType != "0" {
		t.Errorf("Expected LocationType '0', got '%s'", stop.LocationType)
	}
}

func TestDefaultStopProducer_ProduceStopFromStopPlace(t *testing.T) {
	stopAreaRepo := &mockStopAreaRepository{}
	gtfsRepo := &mockGtfsRepository{}
	producer := NewDefaultStopProducer(stopAreaRepo, gtfsRepo)

	stopPlace := &model.StopPlace{
		ID:   "test-stopplace",
		Name: "Test Station",
		Centroid: &model.Centroid{
			Location: &model.Location{
				Latitude:  59.9139,
				Longitude: 10.7522,
			},
		},
	}

	stop, err := producer.ProduceStopFromStopPlace(stopPlace)
	if err != nil {
		t.Fatalf("ProduceStopFromStopPlace() failed: %v", err)
	}

	if stop == nil {
		t.Fatal("ProduceStopFromStopPlace() returned nil stop")
	}

	if stop.StopID != stopPlace.ID {
		t.Errorf("Expected StopID %s, got %s", stopPlace.ID, stop.StopID)
	}

	if stop.StopName != stopPlace.Name {
		t.Errorf("Expected StopName %s, got %s", stopPlace.Name, stop.StopName)
	}

	// Should be a station (location_type = 1)
	if stop.LocationType != "1" {
		t.Errorf("Expected LocationType '1', got '%s'", stop.LocationType)
	}
}

func TestDefaultFeedInfoProducer_ProduceFeedInfo(t *testing.T) {
	producer := NewDefaultFeedInfoProducer()

	feedInfo, err := producer.ProduceFeedInfo()
	if err != nil {
		t.Fatalf("ProduceFeedInfo() failed: %v", err)
	}

	if feedInfo == nil {
		t.Fatal("ProduceFeedInfo() returned nil feedInfo")
	}

	if feedInfo.FeedPublisherName == "" {
		t.Error("FeedPublisherName should not be empty")
	}

	if feedInfo.FeedVersion == "" {
		t.Error("FeedVersion should not be empty")
	}

	if feedInfo.FeedLang == "" {
		t.Error("FeedLang should not be empty")
	}

	if feedInfo.FeedStartDate == "" {
		t.Error("FeedStartDate should not be empty")
	}

	if feedInfo.FeedEndDate == "" {
		t.Error("FeedEndDate should not be empty")
	}
}
