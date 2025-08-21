package model

// RouteType represents GTFS route types
type RouteType int

const (
	// Basic types
	Tram      RouteType = 0
	Subway    RouteType = 1
	Rail      RouteType = 2
	Bus       RouteType = 3
	Ferry     RouteType = 4
	Cable     RouteType = 5
	Gondola   RouteType = 6
	Funicular RouteType = 7

	// Extended types : Rail
	RailwayService              RouteType = 100
	HighSpeedRailService        RouteType = 101
	LongDistanceTrains          RouteType = 102
	InterRegionalRailService    RouteType = 103
	CarTransportRailService     RouteType = 104
	SleeperRailService          RouteType = 105
	RegionalRailService         RouteType = 106
	TouristRailwayService       RouteType = 107
	RailShuttleWithinComplex    RouteType = 108
	SuburbanRailway             RouteType = 109
	ReplacementRailService      RouteType = 110
	SpecialRailService          RouteType = 111
	LorryTransportRailService   RouteType = 112
	AllRailServices             RouteType = 113
	CrossCountryRailService     RouteType = 114
	VehicleTransportRailService RouteType = 115
	RackAndPinionRailway        RouteType = 116
	AdditionalRailService       RouteType = 117

	// Extended types : Coach
	CoachService              RouteType = 200
	InternationalCoachService RouteType = 201
	NationalCoachService      RouteType = 202
	ShuttleCoachService       RouteType = 203
	RegionalCoachService      RouteType = 204
	SpecialCoachService       RouteType = 205
	SightseeingCoachService   RouteType = 206
	TouristCoachService       RouteType = 207
	CommuterCoachService      RouteType = 208
	AllCoachServices          RouteType = 209

	// Extended types : Suburban Rail
	SuburbanRailwayService RouteType = 300

	// Extended types : Urban Rail
	UrbanRailwayService     RouteType = 400
	MetroService            RouteType = 401
	UndergroundService      RouteType = 402
	UrbanRailwayService2    RouteType = 403
	AllUrbanRailwayServices RouteType = 404
	Monorail                RouteType = 405

	// Extended types : Metro
	MetroService2 RouteType = 500

	// Extended types : Underground
	UndergroundService2 RouteType = 600

	// Extended types : Bus
	BusService                       RouteType = 700
	RegionalBusService               RouteType = 701
	ExpressBusService                RouteType = 702
	StoppingBusService               RouteType = 703
	LocalBusService                  RouteType = 704
	NightBusService                  RouteType = 705
	PostBusService                   RouteType = 706
	SpecialNeedsBus                  RouteType = 707
	MobilityBusService               RouteType = 708
	MobilityBusForRegisteredDisabled RouteType = 709
	SightseeingBus                   RouteType = 710
	ShuttleBus                       RouteType = 711
	SchoolBus                        RouteType = 712
	SchoolAndPublicServiceBus        RouteType = 713
	RailReplacementBusService        RouteType = 714
	AllBusServices                   RouteType = 716

	// Extended types : Trolleybus
	TrolleybusService RouteType = 800

	// Extended types : Tram
	TramService            RouteType = 900
	CityTramService        RouteType = 901
	LocalTramService       RouteType = 902
	RegionalTramService    RouteType = 903
	SightseeingTramService RouteType = 904
	ShuttleTramService     RouteType = 905
	AllTramServices        RouteType = 906

	// Extended types : Water
	WaterTransportService              RouteType = 1000
	InternationalCarFerryService       RouteType = 1001
	NationalCarFerryService            RouteType = 1002
	RegionalCarFerryService            RouteType = 1003
	LocalCarFerryService               RouteType = 1004
	InternationalPassengerFerryService RouteType = 1005
	NationalPassengerFerryService      RouteType = 1006
	RegionalPassengerFerryService      RouteType = 1007
	LocalPassengerFerryService         RouteType = 1008
	PostBoatService                    RouteType = 1009
	TrainFerryService                  RouteType = 1010
	RoadLinkFerryService               RouteType = 1011
	AirportLinkFerryService            RouteType = 1012
	CarHighSpeedFerryService           RouteType = 1013
	PassengerHighSpeedFerryService     RouteType = 1014
	SightseeingBoatService             RouteType = 1015
	SchoolBoat                         RouteType = 1016
	CableDrawnBoatService              RouteType = 1017
	RiverBusService                    RouteType = 1018
	ScheduledFerryService              RouteType = 1019
	ShuttleFerryService                RouteType = 1020
	AllWaterTransportServices          RouteType = 1021

	// Extended types : Air
	AirService                        RouteType = 1100
	InternationalAirService           RouteType = 1101
	DomesticAirService                RouteType = 1102
	IntercontinentalAirService        RouteType = 1103
	DomesticScheduledAirService       RouteType = 1104
	ShuttleAirService                 RouteType = 1105
	IntercontinentalCharterAirService RouteType = 1106
	InternationalCharterAirService    RouteType = 1107
	RoundTripCharterAirService        RouteType = 1108
	SightseeingAirService             RouteType = 1109
	HelicopterAirService              RouteType = 1110
	DomesticCharterAirService         RouteType = 1111
	SchengenAreaAirService            RouteType = 1112
	AirshipService                    RouteType = 1113
	AllAirServices                    RouteType = 1114

	// Extended types : Ferry
	FerryService RouteType = 1200

	// Extended types : Telecabin
	TelecabinService      RouteType = 1300
	TelecabinService2     RouteType = 1301
	CableCarService       RouteType = 1302
	ElevatorService       RouteType = 1303
	ChairLiftService      RouteType = 1304
	DragLiftService       RouteType = 1305
	SmallTelecabinService RouteType = 1306
	AllTelecabinServices  RouteType = 1307

	// Extended types : Funicular
	FunicularService    RouteType = 1400
	FunicularService2   RouteType = 1401
	AllFunicularService RouteType = 1402

	// Extended types : Taxi
	TaxiService               RouteType = 1500
	CommunalTaxiService       RouteType = 1501
	WaterTaxiService          RouteType = 1502
	RailTaxiService           RouteType = 1503
	BikeTaxiService           RouteType = 1504
	LicensedTaxiService       RouteType = 1505
	PrivateHireServiceVehicle RouteType = 1506
	AllTaxiServices           RouteType = 1507

	// Extended types : Self-drive
	SelfDrive     RouteType = 1600
	HireCar       RouteType = 1601
	HireVan       RouteType = 1602
	HireMotorbike RouteType = 1603
	HireCycle     RouteType = 1604

	// Extended types : Miscellaneous
	MiscellaneousService RouteType = 1700
	CableCar             RouteType = 1701
	HorseDrawnCarriage   RouteType = 1702
)

// MapNetexToGtfsRouteType converts NeTEx transport mode and submode to GTFS route type
func MapNetexToGtfsRouteType(mode, submode string) RouteType {
	return MapNetexToGtfsRouteTypeWithConfig(mode, submode, false)
}

// MapNetexToGtfsRouteTypeWithConfig converts NeTEx transport mode and submode to GTFS route type
// with option to use basic GTFS types for compatibility
func MapNetexToGtfsRouteTypeWithConfig(mode, submode string, useBasicTypes bool) RouteType {
	switch mode {
	case "air":
		switch submode {
		case "domesticFlight":
			return DomesticAirService
		case "helicopterService":
			return HelicopterAirService
		case "internationalFlight":
			return InternationalAirService
		default:
			return AirService
		}
	case "bus":
		if useBasicTypes {
			// Use basic GTFS route type for maximum compatibility
			return Bus
		}
		switch submode {
		case "airportLinkBus":
			return BusService
		case "expressBus":
			return ExpressBusService
		case "localBus":
			return LocalBusService
		case "nightBus":
			return NightBusService
		case "railReplacementBus":
			return RailReplacementBusService
		case "regionalBus":
			return RegionalBusService
		case "schoolBus":
			return SchoolBus
		case "shuttleBus":
			return ShuttleBus
		case "sightseeingBus":
			return SightseeingBus
		default:
			return BusService
		}
	case "cableway":
		return TelecabinService
	case "coach":
		switch submode {
		case "internationalCoach":
			return InternationalCoachService
		case "nationalCoach":
			return NationalCoachService
		case "touristCoach":
			return TouristCoachService
		default:
			return CoachService
		}
	case "ferry":
		return FerryService
	case "funicular":
		return FunicularService
	case "lift":
		return TelecabinService
	case "metro":
		return MetroService
	case "other":
		return MiscellaneousService
	case "rail":
		switch submode {
		case "airportLinkRail":
			return HighSpeedRailService
		case "international":
			return LongDistanceTrains
		case "interregionalRail":
			return InterRegionalRailService
		case "local":
			return RailwayService
		case "longDistance":
			return LongDistanceTrains
		case "nightRail":
			return SleeperRailService
		case "regionalRail":
			return RegionalRailService
		case "touristRailway":
			return TouristRailwayService
		default:
			return RailwayService
		}
	case "taxi":
		return TaxiService
	case "tram":
		switch submode {
		case "cityTram":
			return CityTramService
		case "localTram":
			return LocalTramService
		default:
			return TramService
		}
	case "trolleyBus":
		return TrolleybusService
	case "water":
		switch submode {
		case "highSpeedPassengerService":
			return PassengerHighSpeedFerryService
		case "highSpeedVehicleService":
			return CarHighSpeedFerryService
		case "internationalCarFerry":
			return InternationalCarFerryService
		case "internationalPassengerFerry":
			return InternationalPassengerFerryService
		case "localCarFerry":
			return LocalCarFerryService
		case "localPassengerFerry":
			return LocalPassengerFerryService
		case "nationalCarFerry":
			return NationalCarFerryService
		case "sightseeingService":
			return SightseeingBoatService
		default:
			return WaterTransportService
		}
	default:
		if useBasicTypes {
			return Bus // basic GTFS bus type for compatibility
		}
		return BusService // default fallback
	}
}

// String returns the string representation of the route type
func (rt RouteType) String() string {
	switch rt {
	case Tram:
		return "Tram"
	case Subway:
		return "Subway"
	case Rail:
		return "Rail"
	case Bus:
		return "Bus"
	case Ferry:
		return "Ferry"
	case Cable:
		return "Cable"
	case Gondola:
		return "Gondola"
	case Funicular:
		return "Funicular"
	case BusService:
		return "Bus Service"
	case ExpressBusService:
		return "Express Bus Service"
	case LocalBusService:
		return "Local Bus Service"
	case RailwayService:
		return "Railway Service"
	case HighSpeedRailService:
		return "High Speed Rail Service"
	case LongDistanceTrains:
		return "Long Distance Trains"
	case MetroService:
		return "Metro Service"
	case TramService:
		return "Tram Service"
	case FerryService:
		return "Ferry Service"
	case AirService:
		return "Air Service"
	case TaxiService:
		return "Taxi Service"
	case TelecabinService:
		return "Telecabin Service"
	case FunicularService:
		return "Funicular Service"
	default:
		return "Unknown"
	}
}

// Value returns the integer value of the route type
func (rt RouteType) Value() int {
	return int(rt)
}
