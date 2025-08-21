package model

import "testing"

func TestMapNetexToGtfsRouteType_Air(t *testing.T) {
	tests := []struct {
		mode     string
		submode  string
		expected RouteType
	}{
		{"air", "domesticFlight", DomesticAirService},
		{"air", "helicopterService", HelicopterAirService},
		{"air", "internationalFlight", InternationalAirService},
		{"air", "", AirService},        // default air
		{"air", "unknown", AirService}, // fallback
	}

	for _, tt := range tests {
		t.Run(tt.mode+"_"+tt.submode, func(t *testing.T) {
			result := MapNetexToGtfsRouteType(tt.mode, tt.submode)
			if result != tt.expected {
				t.Errorf("MapNetexToGtfsRouteType(%s, %s) = %v, expected %v",
					tt.mode, tt.submode, result, tt.expected)
			}
		})
	}
}

func TestMapNetexToGtfsRouteType_Bus(t *testing.T) {
	tests := []struct {
		mode     string
		submode  string
		expected RouteType
	}{
		{"bus", "airportLinkBus", BusService},
		{"bus", "expressBus", ExpressBusService},
		{"bus", "localBus", LocalBusService},
		{"bus", "nightBus", NightBusService},
		{"bus", "railReplacementBus", RailReplacementBusService},
		{"bus", "regionalBus", RegionalBusService},
		{"bus", "schoolBus", SchoolBus},
		{"bus", "shuttleBus", ShuttleBus},
		{"bus", "sightseeingBus", SightseeingBus},
		{"bus", "", BusService},        // default bus
		{"bus", "unknown", BusService}, // fallback
	}

	for _, tt := range tests {
		t.Run(tt.mode+"_"+tt.submode, func(t *testing.T) {
			result := MapNetexToGtfsRouteType(tt.mode, tt.submode)
			if result != tt.expected {
				t.Errorf("MapNetexToGtfsRouteType(%s, %s) = %v, expected %v",
					tt.mode, tt.submode, result, tt.expected)
			}
		})
	}
}

func TestMapNetexToGtfsRouteType_Rail(t *testing.T) {
	tests := []struct {
		mode     string
		submode  string
		expected RouteType
	}{
		{"rail", "airportLinkRail", HighSpeedRailService},
		{"rail", "international", LongDistanceTrains},
		{"rail", "interregionalRail", InterRegionalRailService},
		{"rail", "local", RailwayService},
		{"rail", "longDistance", LongDistanceTrains},
		{"rail", "nightRail", SleeperRailService},
		{"rail", "regionalRail", RegionalRailService},
		{"rail", "touristRailway", TouristRailwayService},
		{"rail", "", RailwayService},        // default rail
		{"rail", "unknown", RailwayService}, // fallback
	}

	for _, tt := range tests {
		t.Run(tt.mode+"_"+tt.submode, func(t *testing.T) {
			result := MapNetexToGtfsRouteType(tt.mode, tt.submode)
			if result != tt.expected {
				t.Errorf("MapNetexToGtfsRouteType(%s, %s) = %v, expected %v",
					tt.mode, tt.submode, result, tt.expected)
			}
		})
	}
}

func TestMapNetexToGtfsRouteType_Water(t *testing.T) {
	tests := []struct {
		mode     string
		submode  string
		expected RouteType
	}{
		{"water", "highSpeedPassengerService", PassengerHighSpeedFerryService},
		{"water", "highSpeedVehicleService", CarHighSpeedFerryService},
		{"water", "internationalCarFerry", InternationalCarFerryService},
		{"water", "internationalPassengerFerry", InternationalPassengerFerryService},
		{"water", "localCarFerry", LocalCarFerryService},
		{"water", "localPassengerFerry", LocalPassengerFerryService},
		{"water", "nationalCarFerry", NationalCarFerryService},
		{"water", "sightseeingService", SightseeingBoatService},
		{"water", "", WaterTransportService},        // default water
		{"water", "unknown", WaterTransportService}, // fallback
	}

	for _, tt := range tests {
		t.Run(tt.mode+"_"+tt.submode, func(t *testing.T) {
			result := MapNetexToGtfsRouteType(tt.mode, tt.submode)
			if result != tt.expected {
				t.Errorf("MapNetexToGtfsRouteType(%s, %s) = %v, expected %v",
					tt.mode, tt.submode, result, tt.expected)
			}
		})
	}
}

func TestMapNetexToGtfsRouteType_SingleModes(t *testing.T) {
	tests := []struct {
		mode     string
		submode  string
		expected RouteType
	}{
		{"cableway", "", TelecabinService},
		{"ferry", "", FerryService},
		{"funicular", "", FunicularService},
		{"lift", "", TelecabinService},
		{"metro", "", MetroService},
		{"other", "", MiscellaneousService},
		{"taxi", "", TaxiService},
		{"trolleyBus", "", TrolleybusService},
	}

	for _, tt := range tests {
		t.Run(tt.mode, func(t *testing.T) {
			result := MapNetexToGtfsRouteType(tt.mode, tt.submode)
			if result != tt.expected {
				t.Errorf("MapNetexToGtfsRouteType(%s, %s) = %v, expected %v",
					tt.mode, tt.submode, result, tt.expected)
			}
		})
	}
}

func TestMapNetexToGtfsRouteType_Tram(t *testing.T) {
	tests := []struct {
		mode     string
		submode  string
		expected RouteType
	}{
		{"tram", "cityTram", CityTramService},
		{"tram", "localTram", LocalTramService},
		{"tram", "", TramService},        // default tram
		{"tram", "unknown", TramService}, // fallback
	}

	for _, tt := range tests {
		t.Run(tt.mode+"_"+tt.submode, func(t *testing.T) {
			result := MapNetexToGtfsRouteType(tt.mode, tt.submode)
			if result != tt.expected {
				t.Errorf("MapNetexToGtfsRouteType(%s, %s) = %v, expected %v",
					tt.mode, tt.submode, result, tt.expected)
			}
		})
	}
}

func TestMapNetexToGtfsRouteType_Coach(t *testing.T) {
	tests := []struct {
		mode     string
		submode  string
		expected RouteType
	}{
		{"coach", "internationalCoach", InternationalCoachService},
		{"coach", "nationalCoach", NationalCoachService},
		{"coach", "touristCoach", TouristCoachService},
		{"coach", "", CoachService},        // default coach
		{"coach", "unknown", CoachService}, // fallback
	}

	for _, tt := range tests {
		t.Run(tt.mode+"_"+tt.submode, func(t *testing.T) {
			result := MapNetexToGtfsRouteType(tt.mode, tt.submode)
			if result != tt.expected {
				t.Errorf("MapNetexToGtfsRouteType(%s, %s) = %v, expected %v",
					tt.mode, tt.submode, result, tt.expected)
			}
		})
	}
}

func TestMapNetexToGtfsRouteType_DefaultFallback(t *testing.T) {
	tests := []struct {
		mode     string
		submode  string
		expected RouteType
	}{
		{"", "", BusService},                       // empty mode
		{"unknown", "", BusService},                // unknown mode
		{"invalidMode", "someSubmode", BusService}, // invalid mode
	}

	for _, tt := range tests {
		t.Run("fallback_"+tt.mode+"_"+tt.submode, func(t *testing.T) {
			result := MapNetexToGtfsRouteType(tt.mode, tt.submode)
			if result != tt.expected {
				t.Errorf("MapNetexToGtfsRouteType(%s, %s) = %v, expected %v",
					tt.mode, tt.submode, result, tt.expected)
			}
		})
	}
}

func TestRouteType_String(t *testing.T) {
	tests := []struct {
		routeType RouteType
		expected  string
	}{
		{Tram, "Tram"},
		{Subway, "Subway"},
		{Rail, "Rail"},
		{Bus, "Bus"},
		{Ferry, "Ferry"},
		{Cable, "Cable"},
		{Gondola, "Gondola"},
		{Funicular, "Funicular"},
		{BusService, "Bus Service"},
		{ExpressBusService, "Express Bus Service"},
		{LocalBusService, "Local Bus Service"},
		{RailwayService, "Railway Service"},
		{HighSpeedRailService, "High Speed Rail Service"},
		{LongDistanceTrains, "Long Distance Trains"},
		{MetroService, "Metro Service"},
		{TramService, "Tram Service"},
		{FerryService, "Ferry Service"},
		{AirService, "Air Service"},
		{TaxiService, "Taxi Service"},
		{TelecabinService, "Telecabin Service"},
		{FunicularService, "Funicular Service"},
		{RouteType(9999), "Unknown"}, // Unknown route type
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.routeType.String()
			if result != tt.expected {
				t.Errorf("RouteType(%d).String() = %s, expected %s",
					tt.routeType, result, tt.expected)
			}
		})
	}
}

func TestRouteType_Value(t *testing.T) {
	tests := []struct {
		routeType RouteType
		expected  int
	}{
		{Tram, 0},
		{Subway, 1},
		{Rail, 2},
		{Bus, 3},
		{Ferry, 4},
		{Cable, 5},
		{Gondola, 6},
		{Funicular, 7},
		{RailwayService, 100},
		{BusService, 700},
		{TramService, 900},
		{AirService, 1100},
		{FerryService, 1200},
		{TaxiService, 1500},
	}

	for _, tt := range tests {
		t.Run(tt.routeType.String(), func(t *testing.T) {
			result := tt.routeType.Value()
			if result != tt.expected {
				t.Errorf("RouteType(%d).Value() = %d, expected %d",
					tt.routeType, result, tt.expected)
			}
		})
	}
}

func TestRouteType_Constants(t *testing.T) {
	// Test that constants have expected values
	if Tram != 0 {
		t.Errorf("Tram constant should be 0, got %d", Tram)
	}
	if Bus != 3 {
		t.Errorf("Bus constant should be 3, got %d", Bus)
	}
	if BusService != 700 {
		t.Errorf("BusService constant should be 700, got %d", BusService)
	}
	if RailwayService != 100 {
		t.Errorf("RailwayService constant should be 100, got %d", RailwayService)
	}
}

func TestRouteType_ExtendedTypes(t *testing.T) {
	// Test extended route types have correct ranges
	tests := []struct {
		name      string
		routeType RouteType
		minRange  int
		maxRange  int
	}{
		{"Rail services", RailwayService, 100, 199},
		{"Coach services", CoachService, 200, 299},
		{"Bus services", BusService, 700, 799},
		{"Tram services", TramService, 900, 999},
		{"Water services", WaterTransportService, 1000, 1099},
		{"Air services", AirService, 1100, 1199},
		{"Taxi services", TaxiService, 1500, 1599},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value := tt.routeType.Value()
			if value < tt.minRange || value > tt.maxRange {
				t.Errorf("%s should be in range %d-%d, got %d",
					tt.name, tt.minRange, tt.maxRange, value)
			}
		})
	}
}
