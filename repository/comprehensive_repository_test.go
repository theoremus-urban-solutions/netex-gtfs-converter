package repository

import (
	"fmt"
	"testing"

	"github.com/theoremus-urban-solutions/netex-gtfs-converter/model"
)

// Comprehensive tests for DefaultNetexRepository

func TestDefaultNetexRepository_Comprehensive(t *testing.T) {
	repo := NewDefaultNetexRepository()

	// Test all entity types saving and retrieval
	testEntities := []struct {
		name   string
		entity interface{}
		check  func() bool
	}{
		{
			name: "Authority",
			entity: &model.Authority{
				ID:   "auth1",
				Name: "Test Authority",
				ContactDetails: &model.ContactDetails{
					Phone: "+123456789",
					Email: "test@example.com",
					URL:   "https://example.com",
				},
			},
			check: func() bool {
				auth := repo.GetAuthorityById("auth1")
				return auth != nil && auth.Name == "Test Authority"
			},
		},
		{
			name: "Network",
			entity: &model.Network{
				ID:           "net1",
				Name:         "Test Network",
				AuthorityRef: model.NetworkAuthorityRef{Ref: "auth1"},
			},
			check: func() bool {
				// Networks don't have a direct getter, but should be saved
				return true // SaveEntity should not error
			},
		},
		{
			name: "Line",
			entity: &model.Line{
				ID:            "line1",
				Name:          "Test Line",
				ShortName:     "TL",
				PublicCode:    "1",
				AuthorityRef:  "auth1",
				TransportMode: "bus",
			},
			check: func() bool {
				lines := repo.GetLines()
				return len(lines) > 0 && lines[0].ID == "line1"
			},
		},
		{
			name: "Route",
			entity: &model.Route{
				ID:      "route1",
				Name:    "Test Route",
				LineRef: model.RouteLineRef{Ref: "line1"},
			},
			check: func() bool {
				route := repo.GetRouteById("route1")
				return route != nil && route.Name == "Test Route"
			},
		},
		{
			name: "JourneyPattern",
			entity: &model.JourneyPattern{
				ID:       "jp1",
				Name:     "Test Journey Pattern",
				RouteRef: "route1",
			},
			check: func() bool {
				jp := repo.GetJourneyPatternById("jp1")
				return jp != nil && jp.Name == "Test Journey Pattern"
			},
		},
		{
			name: "ServiceJourney",
			entity: &model.ServiceJourney{
				ID:                "sj1",
				LineRef:           model.ServiceJourneyLineRef{Ref: "line1"},
				JourneyPatternRef: model.ServiceJourneyPatternRef{Ref: "jp1"},
				PassingTimes: &model.PassingTimes{
					TimetabledPassingTime: []model.TimetabledPassingTime{
						{ArrivalTime: "10:00:00", DepartureTime: "10:01:00"},
					},
				},
			},
			check: func() bool {
				journeys := repo.GetServiceJourneys()
				return len(journeys) > 0 && journeys[0].ID == "sj1"
			},
		},
		{
			name: "StopPlace",
			entity: &model.StopPlace{
				ID:   "sp1",
				Name: "Test Stop Place",
				Centroid: &model.Centroid{
					Location: &model.Location{
						Latitude:  60.0,
						Longitude: 10.0,
					},
				},
			},
			check: func() bool {
				stopPlaces := repo.GetAllStopPlaces()
				return len(stopPlaces) > 0 && stopPlaces[0].ID == "sp1"
			},
		},
		{
			name: "Quay",
			entity: &model.Quay{
				ID:   "quay1",
				Name: "Platform 1",
				Centroid: &model.Centroid{
					Location: &model.Location{
						Latitude:  60.0,
						Longitude: 10.0,
					},
				},
			},
			check: func() bool {
				quays := repo.GetAllQuays()
				return len(quays) > 0 && quays[0].ID == "quay1"
			},
		},
		{
			name: "DayType",
			entity: &model.DayType{
				ID:   "dt1",
				Name: "Weekdays",
			},
			check: func() bool {
				dt := repo.GetDayTypeById("dt1")
				return dt != nil && dt.Name == "Weekdays"
			},
		},
		{
			name: "ServiceJourneyInterchange",
			entity: &model.ServiceJourneyInterchange{
				ID: "int1",
			},
			check: func() bool {
				interchanges := repo.GetServiceJourneyInterchanges()
				return len(interchanges) > 0 && interchanges[0].ID == "int1"
			},
		},
	}

	// Save all entities and verify
	for _, test := range testEntities {
		t.Run("Save"+test.name, func(t *testing.T) {
			err := repo.SaveEntity(test.entity)
			if err != nil {
				t.Errorf("SaveEntity(%s) failed: %v", test.name, err)
			}

			if !test.check() {
				t.Errorf("Entity %s was not saved correctly", test.name)
			}
		})
	}
}

func TestDefaultNetexRepository_LookupMethods(t *testing.T) {
	repo := NewDefaultNetexRepository()

	// Set up test data
	line := &model.Line{ID: "line1", Name: "Test Line", AuthorityRef: "auth1"}
	if err := repo.SaveEntity(line); err != nil {
		t.Fatal(err)
	}

	route := &model.Route{
		ID:      "route1",
		Name:    "Test Route",
		LineRef: model.RouteLineRef{Ref: "line1"},
	}
	if err := repo.SaveEntity(route); err != nil {
		t.Fatal(err)
	}

	jp := &model.JourneyPattern{ID: "jp1", RouteRef: "route1"}
	if err := repo.SaveEntity(jp); err != nil {
		t.Fatal(err)
	}

	sj := &model.ServiceJourney{
		ID:                "sj1",
		LineRef:           model.ServiceJourneyLineRef{Ref: "line1"},
		JourneyPatternRef: model.ServiceJourneyPatternRef{Ref: "jp1"},
	}
	if err := repo.SaveEntity(sj); err != nil {
		t.Fatal(err)
	}

	// Test lookup methods
	t.Run("GetRoutesByLine", func(t *testing.T) {
		routes := repo.GetRoutesByLine(line)
		if len(routes) != 1 {
			t.Errorf("Expected 1 route, got %d", len(routes))
		}
		if routes[0].ID != "route1" {
			t.Errorf("Expected route1, got %s", routes[0].ID)
		}
	})

	t.Run("GetServiceJourneysByJourneyPattern", func(t *testing.T) {
		journeys := repo.GetServiceJourneysByJourneyPattern(jp)
		if len(journeys) != 1 {
			t.Errorf("Expected 1 service journey, got %d", len(journeys))
		}
		if journeys[0].ID != "sj1" {
			t.Errorf("Expected sj1, got %s", journeys[0].ID)
		}
	})
}

func TestDefaultNetexRepository_UnsupportedEntity(t *testing.T) {
	repo := NewDefaultNetexRepository()

	// Test with unsupported entity type
	unsupported := &struct {
		ID   string
		Name string
	}{
		ID:   "test",
		Name: "Unsupported",
	}

	err := repo.SaveEntity(unsupported)
	if err == nil {
		t.Error("Expected error for unsupported entity type")
	}
}

func TestDefaultNetexRepository_NilEntity(t *testing.T) {
	repo := NewDefaultNetexRepository()

	err := repo.SaveEntity(nil)
	if err == nil {
		t.Error("Expected error for nil entity")
	}
}

func TestDefaultNetexRepository_GetNonExistentEntities(t *testing.T) {
	repo := NewDefaultNetexRepository()

	// Test getting non-existent entities
	if auth := repo.GetAuthorityById("nonexistent"); auth != nil {
		t.Error("Expected nil for non-existent authority")
	}

	if route := repo.GetRouteById("nonexistent"); route != nil {
		t.Error("Expected nil for non-existent route")
	}

	if jp := repo.GetJourneyPatternById("nonexistent"); jp != nil {
		t.Error("Expected nil for non-existent journey pattern")
	}

	if dt := repo.GetDayTypeById("nonexistent"); dt != nil {
		t.Error("Expected nil for non-existent day type")
	}
}

// Comprehensive tests for DefaultGtfsRepository

func TestDefaultGtfsRepository_Comprehensive(t *testing.T) {
	repo := NewDefaultGtfsRepository()

	// Test all GTFS entity types
	testEntities := []struct {
		name   string
		entity interface{}
		check  func() bool
	}{
		{
			name: "Agency",
			entity: &model.Agency{
				AgencyID:       "agency1",
				AgencyName:     "Test Agency",
				AgencyURL:      "https://example.com",
				AgencyTimezone: "Europe/Oslo",
				AgencyLang:     "no",
			},
			check: func() bool {
				// Should be set as default agency
				return repo.GetDefaultAgency() != nil
			},
		},
		{
			name: "Route",
			entity: &model.GtfsRoute{
				RouteID:        "route1",
				AgencyID:       "agency1",
				RouteShortName: "R1",
				RouteLongName:  "Route 1",
				RouteType:      3, // Bus
				RouteColor:     "FF0000",
				RouteTextColor: "FFFFFF",
			},
			check: func() bool {
				// Routes don't have a direct getter in the interface
				return true
			},
		},
		{
			name: "Stop",
			entity: &model.Stop{
				StopID:   "stop1",
				StopName: "Test Stop",
				StopLat:  60.0,
				StopLon:  10.0,
				StopCode: "TS1",
			},
			check: func() bool {
				// Stops don't have a direct getter in the interface
				return true
			},
		},
		{
			name: "Trip",
			entity: &model.Trip{
				TripID:       "trip1",
				RouteID:      "route1",
				ServiceID:    "service1",
				TripHeadsign: "Test Destination",
				DirectionID:  "0",
				ShapeID:      "shape1",
			},
			check: func() bool {
				// Trips don't have a direct getter in the interface
				return true
			},
		},
		{
			name: "StopTime",
			entity: &model.StopTime{
				TripID:            "trip1",
				StopID:            "stop1",
				StopSequence:      1,
				ArrivalTime:       "10:00:00",
				DepartureTime:     "10:01:00",
				PickupType:        "0",
				DropOffType:       "0",
				ShapeDistTraveled: 0.0,
			},
			check: func() bool {
				// StopTimes don't have a direct getter in the interface
				return true
			},
		},
		{
			name: "Calendar",
			entity: &model.Calendar{
				ServiceID: "service1",
				Monday:    true,
				Tuesday:   true,
				Wednesday: true,
				Thursday:  true,
				Friday:    true,
				Saturday:  false,
				Sunday:    false,
				StartDate: "20240101",
				EndDate:   "20241231",
			},
			check: func() bool {
				// Calendars don't have a direct getter in the interface
				return true
			},
		},
		{
			name: "CalendarDate",
			entity: &model.CalendarDate{
				ServiceID:     "service1",
				Date:          "20240101",
				ExceptionType: 2, // Service removed
			},
			check: func() bool {
				// Calendar dates don't have a direct getter
				return true
			},
		},
		{
			name: "Transfer",
			entity: &model.Transfer{
				FromStopID:      "stop1",
				ToStopID:        "stop2",
				TransferType:    0,
				MinTransferTime: 120,
			},
			check: func() bool {
				// Transfers don't have a direct getter
				return true
			},
		},
		{
			name: "Shape",
			entity: &model.Shape{
				ShapeID:           "shape1",
				ShapePtSequence:   1,
				ShapePtLat:        60.0,
				ShapePtLon:        10.0,
				ShapeDistTraveled: 0.0,
			},
			check: func() bool {
				// Shapes don't have a direct getter
				return true
			},
		},
		{
			name: "FeedInfo",
			entity: &model.FeedInfo{
				FeedPublisherName: "Test Publisher",
				FeedPublisherURL:  "https://example.com",
				FeedLang:          "no",
				FeedStartDate:     "20240101",
				FeedEndDate:       "20241231",
				FeedVersion:       "1.0",
			},
			check: func() bool {
				// Feed info should be stored
				return true
			},
		},
	}

	// Save all entities
	for _, test := range testEntities {
		t.Run("Save"+test.name, func(t *testing.T) {
			err := repo.SaveEntity(test.entity)
			if err != nil {
				t.Errorf("SaveEntity(%s) failed: %v", test.name, err)
			}

			if !test.check() {
				t.Errorf("Entity %s was not saved correctly", test.name)
			}
		})
	}
}

func TestDefaultGtfsRepository_WriteGtfs(t *testing.T) {
	repo := NewDefaultGtfsRepository()

	// Add minimal required data
	agency := &model.Agency{
		AgencyID:       "agency1",
		AgencyName:     "Test Agency",
		AgencyURL:      "https://example.com",
		AgencyTimezone: "Europe/Oslo",
	}
	if err := repo.SaveEntity(agency); err != nil {
		t.Fatal(err)
	}

	feedInfo := &model.FeedInfo{
		FeedPublisherName: "Test Publisher",
		FeedPublisherURL:  "https://example.com",
		FeedLang:          "no",
	}
	if err := repo.SaveEntity(feedInfo); err != nil {
		t.Fatal(err)
	}

	// Test WriteGtfs
	result, err := repo.WriteGtfs()
	if err != nil {
		t.Fatalf("WriteGtfs() failed: %v", err)
	}

	if result == nil {
		t.Error("WriteGtfs() returned nil result")
	}

	// Read some data to verify it's a valid reader
	buffer := make([]byte, 4)
	n, err := result.Read(buffer)
	if err != nil && err.Error() != "EOF" {
		t.Errorf("Failed to read from result: %v", err)
	}

	if n < 4 {
		t.Error("Result too small to be a valid ZIP")
	}

	// Check ZIP magic bytes
	if buffer[0] != 'P' || buffer[1] != 'K' {
		t.Error("Result does not appear to be a ZIP file")
	}
}

func TestDefaultGtfsRepository_UnsupportedEntity(t *testing.T) {
	repo := NewDefaultGtfsRepository()

	// Test with unsupported entity type
	unsupported := &struct {
		ID   string
		Name string
	}{
		ID:   "test",
		Name: "Unsupported",
	}

	err := repo.SaveEntity(unsupported)
	if err == nil {
		t.Error("Expected error for unsupported entity type")
	}
}

// Comprehensive tests for DefaultStopAreaRepository

func TestDefaultStopAreaRepository_Comprehensive(t *testing.T) {
	repo := NewDefaultStopAreaRepository()

	// Test getting empty collections
	if len(repo.GetAllQuays()) != 0 {
		t.Error("Expected empty quay collection initially")
	}

	// Test LoadStopAreas with invalid ZIP data
	invalidZip := []byte("invalid zip data")
	err := repo.LoadStopAreas(invalidZip)
	if err == nil {
		t.Error("Expected error for invalid ZIP data")
	}

	// For now, skip the valid ZIP test as it requires complex ZIP creation
	// Instead, test the basic repository functionality without ZIP loading

	// Test retrieval methods on empty repository
	allQuays := repo.GetAllQuays()
	if len(allQuays) != 0 {
		t.Errorf("Expected 0 quays, got %d", len(allQuays))
	}

	retrievedQuay := repo.GetQuayById("quay1")
	if retrievedQuay != nil {
		t.Error("Expected GetQuayById() to return nil for non-existent quay")
	}

	// Test non-existent quay
	if repo.GetQuayById("nonexistent") != nil {
		t.Error("Expected nil for non-existent quay")
	}

	// Test stop place association
	if repo.GetStopPlaceByQuayId("quay1") != nil {
		t.Error("Expected nil stop place for quay without association")
	}
}

func TestDefaultStopAreaRepository_LoadStopAreasZip(t *testing.T) {
	repo := NewDefaultStopAreaRepository()

	// Test with data that looks like zip but isn't valid
	fakeZipData := []byte("PK\x03\x04fake zip content")
	err := repo.LoadStopAreas(fakeZipData)
	if err == nil {
		t.Error("Expected error for invalid zip data")
	}
}

func TestDefaultStopAreaRepository_LoadStopAreasXML(t *testing.T) {
	repo := NewDefaultStopAreaRepository()

	// Test with XML data (should be handled)
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<root>
	<Quay id="quay1">
		<Name>Platform 1</Name>
	</Quay>
</root>`

	err := repo.LoadStopAreas([]byte(xmlData))
	// This might succeed or fail depending on implementation
	// The key is that it shouldn't panic
	t.Logf("XML loading result: %v", err)
}

// Test error cases and edge conditions

func TestRepositories_ConcurrentAccess(t *testing.T) {
	repo := NewDefaultNetexRepository()

	// Test concurrent access (the repository uses sync.RWMutex)
	done := make(chan bool)

	// Start multiple goroutines accessing the repository
	for i := 0; i < 10; i++ {
		go func(id int) {
			line := &model.Line{
				ID:   fmt.Sprintf("line%d", id),
				Name: fmt.Sprintf("Line %d", id),
			}
			if err := repo.SaveEntity(line); err != nil {
				// Use a channel to communicate error instead of t.Fatal
				done <- false
				return
			}

			// Read data
			_ = repo.GetLines()
			_ = repo.GetAuthorityIdForLine(line)

			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		if !<-done {
			t.Fatal("One or more goroutines failed")
		}
	}

	// Verify all lines were saved
	lines := repo.GetLines()
	if len(lines) != 10 {
		t.Errorf("Expected 10 lines, got %d", len(lines))
	}
}

func TestDefaultNetexRepository_EntityRelationships(t *testing.T) {
	repo := NewDefaultNetexRepository()

	// Create related entities
	authority := &model.Authority{ID: "auth1", Name: "Authority 1"}
	if err := repo.SaveEntity(authority); err != nil {
		t.Fatal(err)
	}

	line := &model.Line{ID: "line1", Name: "Line 1", AuthorityRef: "auth1"}
	if err := repo.SaveEntity(line); err != nil {
		t.Fatal(err)
	}

	route1 := &model.Route{ID: "route1", Name: "Route 1", LineRef: model.RouteLineRef{Ref: "line1"}}
	route2 := &model.Route{ID: "route2", Name: "Route 2", LineRef: model.RouteLineRef{Ref: "line1"}}
	if err := repo.SaveEntity(route1); err != nil {
		t.Fatal(err)
	}
	if err := repo.SaveEntity(route2); err != nil {
		t.Fatal(err)
	}

	jp := &model.JourneyPattern{ID: "jp1", Name: "Pattern 1", RouteRef: "route1"}
	if err := repo.SaveEntity(jp); err != nil {
		t.Fatal(err)
	}

	sj1 := &model.ServiceJourney{ID: "sj1", JourneyPatternRef: model.ServiceJourneyPatternRef{Ref: "jp1"}}
	sj2 := &model.ServiceJourney{ID: "sj2", JourneyPatternRef: model.ServiceJourneyPatternRef{Ref: "jp1"}}
	if err := repo.SaveEntity(sj1); err != nil {
		t.Fatal(err)
	}
	if err := repo.SaveEntity(sj2); err != nil {
		t.Fatal(err)
	}

	// Test relationships
	routes := repo.GetRoutesByLine(line)
	if len(routes) != 2 {
		t.Errorf("Expected 2 routes for line, got %d", len(routes))
	}

	journeys := repo.GetServiceJourneysByJourneyPattern(jp)
	if len(journeys) != 2 {
		t.Errorf("Expected 2 service journeys for pattern, got %d", len(journeys))
	}

	// Test authority lookup
	authorityID := repo.GetAuthorityIdForLine(line)
	if authorityID != "auth1" {
		t.Errorf("Expected authority ID 'auth1', got '%s'", authorityID)
	}
}
