package model

import "testing"

const (
	trip1ID = "trip1"
)

func TestAgency_Validation(t *testing.T) {
	tests := []struct {
		name    string
		agency  Agency
		isValid bool
	}{
		{
			name: "valid agency",
			agency: Agency{
				AgencyID:       "agency1",
				AgencyName:     "Test Agency",
				AgencyURL:      "https://example.com",
				AgencyTimezone: "Europe/Oslo",
			},
			isValid: true,
		},
		{
			name: "agency with all fields",
			agency: Agency{
				AgencyID:       "agency2",
				AgencyName:     "Full Agency",
				AgencyURL:      "https://example.com",
				AgencyTimezone: "America/New_York",
				AgencyLang:     "en",
				AgencyPhone:    "+1-555-123-4567",
				AgencyFareURL:  "https://example.com/fares",
				AgencyEmail:    "contact@example.com",
			},
			isValid: true,
		},
		{
			name: "agency with minimal fields",
			agency: Agency{
				AgencyName:     "Minimal Agency",
				AgencyURL:      "https://minimal.com",
				AgencyTimezone: "UTC",
			},
			isValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test that agency struct can be created with different field combinations
			if tt.agency.AgencyName == "" && tt.isValid {
				t.Error("Valid agency should have non-empty name")
			}
			if tt.agency.AgencyURL == "" && tt.isValid {
				t.Error("Valid agency should have non-empty URL")
			}
			if tt.agency.AgencyTimezone == "" && tt.isValid {
				t.Error("Valid agency should have non-empty timezone")
			}
		})
	}
}

func TestGtfsRoute_FieldMapping(t *testing.T) {
	route := GtfsRoute{
		RouteID:           "route1",
		AgencyID:          "agency1",
		RouteShortName:    "R1",
		RouteLongName:     "Route One",
		RouteDesc:         "First route",
		RouteType:         3, // Bus
		RouteURL:          "https://example.com/route1",
		RouteColor:        "FF0000",
		RouteTextColor:    "FFFFFF",
		RouteSortOrder:    1,
		ContinuousPickup:  "0",
		ContinuousDropOff: "0",
	}

	// Test field access
	if route.RouteID != "route1" {
		t.Errorf("Expected RouteID 'route1', got '%s'", route.RouteID)
	}
	if route.RouteType != 3 {
		t.Errorf("Expected RouteType 3, got %d", route.RouteType)
	}
	if route.RouteColor != "FF0000" {
		t.Errorf("Expected RouteColor 'FF0000', got '%s'", route.RouteColor)
	}
}

func TestTrip_FieldMapping(t *testing.T) {
	trip := Trip{
		RouteID:              "route1",
		ServiceID:            "service1",
		TripID:               "trip1",
		TripHeadsign:         "Downtown",
		TripShortName:        "T1",
		DirectionID:          "0",
		BlockID:              "block1",
		ShapeID:              "shape1",
		WheelchairAccessible: "1",
		BikesAllowed:         "2",
	}

	// Test field access
	if trip.TripID != trip1ID {
		t.Errorf("Expected TripID 'trip1', got '%s'", trip.TripID)
	}
	if trip.TripHeadsign != "Downtown" {
		t.Errorf("Expected TripHeadsign 'Downtown', got '%s'", trip.TripHeadsign)
	}
	if trip.WheelchairAccessible != "1" {
		t.Errorf("Expected WheelchairAccessible '1', got '%s'", trip.WheelchairAccessible)
	}
}

func TestStop_FieldMapping(t *testing.T) {
	stop := Stop{
		StopID:   "stop1",
		StopCode: "S001",
	}

	// Test field access
	if stop.StopID != "stop1" {
		t.Errorf("Expected StopID 'stop1', got '%s'", stop.StopID)
	}
	if stop.StopCode != "S001" {
		t.Errorf("Expected StopCode 'S001', got '%s'", stop.StopCode)
	}
}

func TestStopTime_FieldMapping(t *testing.T) {
	stopTime := StopTime{
		TripID:            "trip1",
		ArrivalTime:       "08:30:00",
		DepartureTime:     "08:31:00",
		StopID:            "stop1",
		StopSequence:      1,
		StopHeadsign:      "City Center",
		PickupType:        "0",
		DropOffType:       "0",
		ContinuousPickup:  "0",
		ContinuousDropOff: "0",
		ShapeDistTraveled: 1500.5,
		Timepoint:         "1",
	}

	// Test field access
	if stopTime.TripID != "trip1" {
		t.Errorf("Expected TripID 'trip1', got '%s'", stopTime.TripID)
	}
	if stopTime.ArrivalTime != "08:30:00" {
		t.Errorf("Expected ArrivalTime '08:30:00', got '%s'", stopTime.ArrivalTime)
	}
	if stopTime.StopSequence != 1 {
		t.Errorf("Expected StopSequence 1, got %d", stopTime.StopSequence)
	}
	if stopTime.ShapeDistTraveled != 1500.5 {
		t.Errorf("Expected ShapeDistTraveled 1500.5, got %f", stopTime.ShapeDistTraveled)
	}
}

func TestCalendar_FieldMapping(t *testing.T) {
	calendar := Calendar{
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
	}

	// Test field access
	if calendar.ServiceID != "service1" {
		t.Errorf("Expected ServiceID 'service1', got '%s'", calendar.ServiceID)
	}
	if calendar.Monday != true {
		t.Errorf("Expected Monday true, got %v", calendar.Monday)
	}
	if calendar.StartDate != "20240101" {
		t.Errorf("Expected StartDate '20240101', got '%s'", calendar.StartDate)
	}
}

func TestCalendarDate_FieldMapping(t *testing.T) {
	calendarDate := CalendarDate{
		ServiceID:     "service1",
		Date:          "20240701",
		ExceptionType: 2,
	}

	// Test field access
	if calendarDate.ServiceID != "service1" {
		t.Errorf("Expected ServiceID 'service1', got '%s'", calendarDate.ServiceID)
	}
	if calendarDate.Date != "20240701" {
		t.Errorf("Expected Date '20240701', got '%s'", calendarDate.Date)
	}
	if calendarDate.ExceptionType != 2 {
		t.Errorf("Expected ExceptionType 2, got %d", calendarDate.ExceptionType)
	}
}

func TestShape_FieldMapping(t *testing.T) {
	shape := Shape{
		ShapeID:           "shape1",
		ShapePtLat:        60.123456,
		ShapePtLon:        10.654321,
		ShapePtSequence:   1,
		ShapeDistTraveled: 0.0,
	}

	// Test field access
	if shape.ShapeID != "shape1" {
		t.Errorf("Expected ShapeID 'shape1', got '%s'", shape.ShapeID)
	}
	if shape.ShapePtLat != 60.123456 {
		t.Errorf("Expected ShapePtLat 60.123456, got %f", shape.ShapePtLat)
	}
	if shape.ShapePtSequence != 1 {
		t.Errorf("Expected ShapePtSequence 1, got %d", shape.ShapePtSequence)
	}
}

func TestTransfer_FieldMapping(t *testing.T) {
	transfer := Transfer{
		FromStopID:      "stop1",
		ToStopID:        "stop2",
		FromRouteID:     "route1",
		ToRouteID:       "route2",
		FromTripID:      "trip1",
		ToTripID:        "trip2",
		TransferType:    2,
		MinTransferTime: 300,
	}

	// Test field access
	if transfer.FromStopID != "stop1" {
		t.Errorf("Expected FromStopID 'stop1', got '%s'", transfer.FromStopID)
	}
	if transfer.TransferType != 2 {
		t.Errorf("Expected TransferType 2, got %d", transfer.TransferType)
	}
	if transfer.MinTransferTime != 300 {
		t.Errorf("Expected MinTransferTime 300, got %d", transfer.MinTransferTime)
	}
}

func TestFeedInfo_FieldMapping(t *testing.T) {
	feedInfo := FeedInfo{
		FeedPublisherName: "Test Transit Agency",
		FeedPublisherURL:  "https://example.com",
		FeedLang:          "en",
		FeedStartDate:     "20240101",
		FeedEndDate:       "20241231",
		FeedVersion:       "1.0",
		FeedContactEmail:  "contact@example.com",
		FeedContactURL:    "https://example.com/contact",
	}

	// Test field access
	if feedInfo.FeedPublisherName != "Test Transit Agency" {
		t.Errorf("Expected FeedPublisherName 'Test Transit Agency', got '%s'", feedInfo.FeedPublisherName)
	}
	if feedInfo.FeedLang != "en" {
		t.Errorf("Expected FeedLang 'en', got '%s'", feedInfo.FeedLang)
	}
	if feedInfo.FeedVersion != "1.0" {
		t.Errorf("Expected FeedVersion '1.0', got '%s'", feedInfo.FeedVersion)
	}
}

func TestFrequency_FieldMapping(t *testing.T) {
	frequency := Frequency{
		TripID:      "trip1",
		StartTime:   "06:00:00",
		EndTime:     "22:00:00",
		HeadwaySecs: 600,
		ExactTimes:  "0",
	}

	// Test field access
	if frequency.TripID != "trip1" {
		t.Errorf("Expected TripID 'trip1', got '%s'", frequency.TripID)
	}
	if frequency.HeadwaySecs != 600 {
		t.Errorf("Expected HeadwaySecs 600, got %d", frequency.HeadwaySecs)
	}
	if frequency.ExactTimes != "0" {
		t.Errorf("Expected ExactTimes '0', got '%s'", frequency.ExactTimes)
	}
}

func TestPathway_FieldMapping(t *testing.T) {
	pathway := Pathway{
		PathwayID:            "pathway1",
		FromStopID:           "stop1",
		ToStopID:             "stop2",
		PathwayMode:          1,
		IsBidirectional:      1,
		Length:               100.5,
		TraversalTime:        120,
		StairCount:           0,
		MaxSlope:             0.05,
		MinWidth:             1.2,
		SignpostedAs:         "Exit A",
		ReversedSignpostedAs: "Platform 1",
	}

	// Test field access
	if pathway.PathwayID != "pathway1" {
		t.Errorf("Expected PathwayID 'pathway1', got '%s'", pathway.PathwayID)
	}
	if pathway.Length != 100.5 {
		t.Errorf("Expected Length 100.5, got %f", pathway.Length)
	}
	if pathway.TraversalTime != 120 {
		t.Errorf("Expected TraversalTime 120, got %d", pathway.TraversalTime)
	}
}
