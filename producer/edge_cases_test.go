package producer

import (
	"testing"

	"github.com/theoremus-urban-solutions/netex-gtfs-converter/model"
)

// TestRouteProducerEdgeCases tests edge cases similar to Java version
func TestRouteProducerEdgeCases(t *testing.T) {
	netexRepo := &mockNetexRepository{}
	gtfsRepo := &mockGtfsRepository{}
	producer := NewDefaultRouteProducer(netexRepo, gtfsRepo)

	t.Run("Route without short name", func(t *testing.T) {
		line := &model.Line{
			ID:               "test-line",
			Name:             "Test Line",
			ShortName:        "", // No short name
			PublicCode:       "1",
			TransportMode:    "bus",
			TransportSubmode: "localBus",
			AuthorityRef:     "test-authority",
		}

		route, err := producer.Produce(line)
		if err != nil {
			t.Fatalf("Produce() failed: %v", err)
		}

		if route == nil {
			t.Fatal("Produce() returned nil route")
		}

		if route.RouteShortName != line.PublicCode {
			t.Errorf("Expected RouteShortName %s, got %s", line.PublicCode, route.RouteShortName)
		}

		if route.RouteLongName != line.Name {
			t.Errorf("Expected RouteLongName %s when short name is missing, got %s", line.Name, route.RouteLongName)
		}
	})

	t.Run("Route with identical PublicCode and ShortName", func(t *testing.T) {
		line := &model.Line{
			ID:               "test-line",
			Name:             "Test Line",
			ShortName:        "Line-Short-Name",
			PublicCode:       "Line-Short-Name", // Same as short name
			TransportMode:    "bus",
			TransportSubmode: "localBus",
			AuthorityRef:     "test-authority",
		}

		route, err := producer.Produce(line)
		if err != nil {
			t.Fatalf("Produce() failed: %v", err)
		}

		if route == nil {
			t.Fatal("Produce() returned nil route")
		}

		if route.RouteShortName != line.PublicCode {
			t.Errorf("Expected RouteShortName %s, got %s", line.PublicCode, route.RouteShortName)
		}

		// When PublicCode and ShortName are identical, short name should be cleared
		// to avoid duplication in GTFS output
		if route.RouteLongName == route.RouteShortName {
			t.Error("RouteLongName should be different from RouteShortName to avoid duplication")
		}
	})

	t.Run("Route with basic GTFS types", func(t *testing.T) {
		line := &model.Line{
			ID:               "test-line",
			Name:             "Test Line",
			ShortName:        "TL",
			PublicCode:       "1",
			TransportMode:    "bus",
			TransportSubmode: "localBus",
			AuthorityRef:     "test-authority",
		}

		route, err := producer.Produce(line)
		if err != nil {
			t.Fatalf("Produce() failed: %v", err)
		}

		// Should use basic GTFS route type (3) not extended (700+)
		expectedRouteType := model.Bus.Value() // 3
		if route.RouteType != expectedRouteType {
			t.Errorf("Expected basic GTFS RouteType %d (Bus), got %d", expectedRouteType, route.RouteType)
		}
	})

	t.Run("Route name extraction from complex names", func(t *testing.T) {
		line := &model.Line{
			ID:            "test-line",
			Name:          "Ligne 42 Express Downtown",
			ShortName:     "Ligne 42 Express Downtown", // Same as name
			PublicCode:    "",                          // No public code
			TransportMode: "bus",
			AuthorityRef:  "test-authority",
		}

		route, err := producer.Produce(line)
		if err != nil {
			t.Fatalf("Produce() failed: %v", err)
		}

		// Should extract number from complex name when no public code
		if route.RouteShortName != "42" {
			t.Errorf("Expected RouteShortName to extract number '42' from complex name, got %s", route.RouteShortName)
		}

		if route.RouteLongName != line.Name {
			t.Errorf("Expected RouteLongName %s, got %s", line.Name, route.RouteLongName)
		}
	})
}

// TestStopTimeProducerEdgeCases tests edge cases similar to Java version
func TestStopTimeProducerEdgeCases(t *testing.T) {
	netexRepo := &mockNetexRepository{}
	gtfsRepo := &mockGtfsRepository{}
	producer := NewDefaultStopTimeProducer(netexRepo, gtfsRepo)

	t.Run("StopTime with day offset", func(t *testing.T) {
		input := StopTimeInput{
			Trip: &model.Trip{TripID: "test-trip"},
			TimetabledPassingTime: &model.TimetabledPassingTime{
				ArrivalTime:              "01:30:00",
				DepartureTime:            "01:35:00",
				DayOffset:                1, // Next day
				PointInJourneyPatternRef: "test-pjp",
			},
			CurrentHeadSign: "Test Destination",
		}

		stopTime, err := producer.Produce(input)
		if err != nil {
			t.Fatalf("Produce() failed: %v", err)
		}

		if stopTime == nil {
			t.Fatal("Produce() returned nil stopTime")
		}

		// Day offset should add 24 hours
		expectedArrival := "25:30:00"   // 01:30:00 + 24 hours
		expectedDeparture := "25:35:00" // 01:35:00 + 24 hours

		if stopTime.ArrivalTime != expectedArrival {
			t.Errorf("Expected ArrivalTime %s with day offset, got %s", expectedArrival, stopTime.ArrivalTime)
		}

		if stopTime.DepartureTime != expectedDeparture {
			t.Errorf("Expected DepartureTime %s with day offset, got %s", expectedDeparture, stopTime.DepartureTime)
		}
	})

	t.Run("StopTime with only arrival time", func(t *testing.T) {
		input := StopTimeInput{
			Trip: &model.Trip{TripID: "test-trip"},
			TimetabledPassingTime: &model.TimetabledPassingTime{
				ArrivalTime:              "14:30:00",
				DepartureTime:            "", // No departure time
				PointInJourneyPatternRef: "test-pjp",
			},
		}

		stopTime, err := producer.Produce(input)
		if err != nil {
			t.Fatalf("Produce() failed: %v", err)
		}

		// Should use arrival time for both arrival and departure when departure is missing
		if stopTime.ArrivalTime != "14:30:00" {
			t.Errorf("Expected ArrivalTime 14:30:00, got %s", stopTime.ArrivalTime)
		}

		if stopTime.DepartureTime != "14:30:00" {
			t.Errorf("Expected DepartureTime to use ArrivalTime when missing, got %s", stopTime.DepartureTime)
		}
	})

	t.Run("StopTime with only departure time", func(t *testing.T) {
		input := StopTimeInput{
			Trip: &model.Trip{TripID: "test-trip"},
			TimetabledPassingTime: &model.TimetabledPassingTime{
				ArrivalTime:              "", // No arrival time
				DepartureTime:            "15:45:00",
				PointInJourneyPatternRef: "test-pjp",
			},
		}

		stopTime, err := producer.Produce(input)
		if err != nil {
			t.Fatalf("Produce() failed: %v", err)
		}

		// Should use departure time for both when arrival is missing
		if stopTime.ArrivalTime != "15:45:00" {
			t.Errorf("Expected ArrivalTime to use DepartureTime when missing, got %s", stopTime.ArrivalTime)
		}

		if stopTime.DepartureTime != "15:45:00" {
			t.Errorf("Expected DepartureTime 15:45:00, got %s", stopTime.DepartureTime)
		}
	})

	t.Run("StopTime with complex day offset", func(t *testing.T) {
		input := StopTimeInput{
			Trip: &model.Trip{TripID: "test-trip"},
			TimetabledPassingTime: &model.TimetabledPassingTime{
				ArrivalTime:              "23:45:00",
				DepartureTime:            "23:50:00",
				DayOffset:                2, // Two days later
				PointInJourneyPatternRef: "test-pjp",
			},
		}

		stopTime, err := producer.Produce(input)
		if err != nil {
			t.Fatalf("Produce() failed: %v", err)
		}

		// Two day offset should add 48 hours to the hour component
		expectedArrival := "71:45:00"   // 23 + (2*24) = 71 hours
		expectedDeparture := "71:50:00" // 23 + (2*24) = 71 hours

		if stopTime.ArrivalTime != expectedArrival {
			t.Errorf("Expected ArrivalTime %s with 2-day offset, got %s", expectedArrival, stopTime.ArrivalTime)
		}

		if stopTime.DepartureTime != expectedDeparture {
			t.Errorf("Expected DepartureTime %s with 2-day offset, got %s", expectedDeparture, stopTime.DepartureTime)
		}
	})
}

// TestServiceCalendarProducerEdgeCases tests service calendar edge cases
func TestServiceCalendarProducerEdgeCases(t *testing.T) {
	gtfsRepo := &mockGtfsRepository{}
	producer := NewDefaultServiceCalendarProducer(gtfsRepo)

	t.Run("Service calendar with no day types", func(t *testing.T) {
		serviceID := "test-service"
		dayTypes := []*model.DayType{} // Empty

		calendar, err := producer.Produce(serviceID, dayTypes)
		if err != nil {
			t.Fatalf("Produce() failed: %v", err)
		}

		if calendar != nil {
			t.Error("Expected nil calendar for empty day types")
		}
	})

	t.Run("Service calendar with weekdays pattern", func(t *testing.T) {
		serviceID := "weekday-service"
		dayType := &model.DayType{
			ID: "weekdays",
			Properties: &model.Properties{
				PropertyOfDay: []model.PropertyOfDay{
					{DaysOfWeek: "Weekdays"},
				},
			},
		}
		dayTypes := []*model.DayType{dayType}

		calendar, err := producer.Produce(serviceID, dayTypes)
		if err != nil {
			t.Fatalf("Produce() failed: %v", err)
		}

		if calendar == nil {
			t.Fatal("Expected non-nil calendar")
		}

		if calendar.ServiceID != serviceID {
			t.Errorf("Expected ServiceID %s, got %s", serviceID, calendar.ServiceID)
		}

		// Should set weekdays to true
		if !calendar.Monday || !calendar.Tuesday || !calendar.Wednesday || !calendar.Thursday || !calendar.Friday {
			t.Error("Expected weekdays (Mon-Fri) to be true")
		}

		// Should set weekend to false
		if calendar.Saturday || calendar.Sunday {
			t.Error("Expected weekend (Sat-Sun) to be false for weekdays pattern")
		}
	})

	t.Run("Service calendar with weekend pattern", func(t *testing.T) {
		serviceID := "weekend-service"
		dayType := &model.DayType{
			ID: "weekend",
			Properties: &model.Properties{
				PropertyOfDay: []model.PropertyOfDay{
					{DaysOfWeek: "Weekend"},
				},
			},
		}
		dayTypes := []*model.DayType{dayType}

		calendar, err := producer.Produce(serviceID, dayTypes)
		if err != nil {
			t.Fatalf("Produce() failed: %v", err)
		}

		if calendar == nil {
			t.Fatal("Expected non-nil calendar")
		}

		// Should set weekend to true
		if !calendar.Saturday || !calendar.Sunday {
			t.Error("Expected weekend (Sat-Sun) to be true")
		}

		// Should set weekdays to false
		if calendar.Monday || calendar.Tuesday || calendar.Wednesday || calendar.Thursday || calendar.Friday {
			t.Error("Expected weekdays (Mon-Fri) to be false for weekend pattern")
		}
	})

	t.Run("Service calendar with no specific days defaults to everyday", func(t *testing.T) {
		serviceID := "default-service"
		dayType := &model.DayType{
			ID: "default",
			Properties: &model.Properties{
				PropertyOfDay: []model.PropertyOfDay{
					{DaysOfWeek: "Unknown"}, // Unrecognized pattern
				},
			},
		}
		dayTypes := []*model.DayType{dayType}

		calendar, err := producer.Produce(serviceID, dayTypes)
		if err != nil {
			t.Fatalf("Produce() failed: %v", err)
		}

		if calendar == nil {
			t.Fatal("Expected non-nil calendar")
		}

		// Should default to everyday when no specific days are set
		if !calendar.Monday || !calendar.Tuesday || !calendar.Wednesday || !calendar.Thursday || !calendar.Friday || !calendar.Saturday || !calendar.Sunday {
			t.Error("Expected all days to be true when no specific pattern is recognized")
		}
	})
}

// TestTransferProducerEdgeCases tests transfer/interchange edge cases
func TestTransferProducerEdgeCases(t *testing.T) {
	netexRepo := &mockNetexRepository{}
	gtfsRepo := &mockGtfsRepository{}
	producer := NewDefaultTransferProducer(netexRepo, gtfsRepo)

	t.Run("Transfer with stay seated", func(t *testing.T) {
		interchange := &model.ServiceJourneyInterchange{
			FromPointRef:   "stop1",
			ToPointRef:     "stop2",
			StaySeated:     true,
			Guaranteed:     false,
			FromJourneyRef: "trip1",
			ToJourneyRef:   "trip2",
		}

		transfer, err := producer.Produce(interchange)
		if err != nil {
			t.Fatalf("Produce() failed: %v", err)
		}

		if transfer == nil {
			t.Fatal("Expected non-nil transfer")
		}

		// Stay seated should result in recommended transfer (type 0)
		if transfer.TransferType != 0 {
			t.Errorf("Expected TransferType 0 for stay seated, got %d", transfer.TransferType)
		}
	})

	t.Run("Transfer with guaranteed", func(t *testing.T) {
		interchange := &model.ServiceJourneyInterchange{
			FromPointRef:        "stop1",
			ToPointRef:          "stop2",
			StaySeated:          false,
			Guaranteed:          true,
			MinimumTransferTime: "PT5M", // 5 minutes
			FromJourneyRef:      "trip1",
			ToJourneyRef:        "trip2",
		}

		transfer, err := producer.Produce(interchange)
		if err != nil {
			t.Fatalf("Produce() failed: %v", err)
		}

		if transfer == nil {
			t.Fatal("Expected non-nil transfer")
		}

		// Guaranteed should result in timed transfer (type 1)
		if transfer.TransferType != 1 {
			t.Errorf("Expected TransferType 1 for guaranteed, got %d", transfer.TransferType)
		}

		// Should parse minimum transfer time (5 minutes = 300 seconds)
		if transfer.MinTransferTime != 300 {
			t.Errorf("Expected MinTransferTime 300 seconds, got %d", transfer.MinTransferTime)
		}
	})

	t.Run("Transfer with minimum time only", func(t *testing.T) {
		interchange := &model.ServiceJourneyInterchange{
			FromPointRef:        "stop1",
			ToPointRef:          "stop2",
			StaySeated:          false,
			Guaranteed:          false,
			MinimumTransferTime: "PT10M", // 10 minutes
			FromJourneyRef:      "trip1",
			ToJourneyRef:        "trip2",
		}

		transfer, err := producer.Produce(interchange)
		if err != nil {
			t.Fatalf("Produce() failed: %v", err)
		}

		if transfer == nil {
			t.Fatal("Expected non-nil transfer")
		}

		// Neither stay seated nor guaranteed should result in minimum time required (type 2)
		if transfer.TransferType != 2 {
			t.Errorf("Expected TransferType 2 for minimum time only, got %d", transfer.TransferType)
		}

		// Should parse minimum transfer time (10 minutes = 600 seconds)
		if transfer.MinTransferTime != 600 {
			t.Errorf("Expected MinTransferTime 600 seconds, got %d", transfer.MinTransferTime)
		}
	})

	t.Run("Transfer with guaranteed but no minimum time", func(t *testing.T) {
		interchange := &model.ServiceJourneyInterchange{
			FromPointRef:   "stop1",
			ToPointRef:     "stop2",
			StaySeated:     false,
			Guaranteed:     true,
			FromJourneyRef: "trip1",
			ToJourneyRef:   "trip2",
		}

		transfer, err := producer.Produce(interchange)
		if err != nil {
			t.Fatalf("Produce() failed: %v", err)
		}

		// Should set default transfer time for guaranteed transfers
		if transfer.MinTransferTime != 120 { // 2 minutes default
			t.Errorf("Expected default MinTransferTime 120 seconds for guaranteed transfer, got %d", transfer.MinTransferTime)
		}
	})
}
