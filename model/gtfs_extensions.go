package model

// GTFS extensions for European features and GTFS-Flex support
// These models extend the basic GTFS specification with additional capabilities

// GTFS-Flex extensions for demand-responsive transport

// LocationGroup represents a group of locations for flexible services
type LocationGroup struct {
	LocationGroupID   string `csv:"location_group_id"`
	LocationGroupName string `csv:"location_group_name,omitempty"`
}

// LocationGroupStop links location groups to stops
type LocationGroupStop struct {
	LocationGroupID string `csv:"location_group_id"`
	StopID          string `csv:"stop_id"`
}

// StopArea represents a geographic area for flexible services  
type StopArea struct {
	AreaID   string  `csv:"area_id"`
	AreaName string  `csv:"area_name,omitempty"`
	AreaLat  float64 `csv:"area_lat,omitempty"`
	AreaLon  float64 `csv:"area_lon,omitempty"`
}

// BookingRule defines booking requirements for flexible services
type BookingRule struct {
	BookingRuleID          string `csv:"booking_rule_id"`
	BookingType            int    `csv:"booking_type"`                      // 0=real-time, 1=up to same day, 2=prior day(s)
	PriorNoticeDurationMin int    `csv:"prior_notice_duration_min,omitempty"` // minutes before departure
	PriorNoticeDurationMax int    `csv:"prior_notice_duration_max,omitempty"` // minutes before departure
	PriorNoticeLastDay     int    `csv:"prior_notice_last_day,omitempty"`     // days before service date
	PriorNoticeLastTime    string `csv:"prior_notice_last_time,omitempty"`    // HH:MM:SS
	PriorNoticeStartDay    int    `csv:"prior_notice_start_day,omitempty"`    // days before service date
	PriorNoticeStartTime   string `csv:"prior_notice_start_time,omitempty"`   // HH:MM:SS
	PriorNoticeServiceID   string `csv:"prior_notice_service_id,omitempty"`
	Message                string `csv:"message,omitempty"`
	PickupMessage          string `csv:"pickup_message,omitempty"`
	DropOffMessage         string `csv:"drop_off_message,omitempty"`
	PhoneNumber            string `csv:"phone_number,omitempty"`
	InfoURL                string `csv:"info_url,omitempty"`
}

// European-specific extensions

// VehicleTypeGTFS represents detailed vehicle information (custom extension)
// Named differently to avoid conflict with NeTEx VehicleType
type VehicleTypeGTFS struct {
	VehicleTypeID          string `csv:"vehicle_type_id"`
	VehicleTypeName        string `csv:"vehicle_type_name,omitempty"`
	VehicleDescription     string `csv:"vehicle_description,omitempty"`
	
	// Accessibility
	WheelchairAccessible   string `csv:"wheelchair_accessible,omitempty"`   // 0, 1, 2
	WheelchairCapacity     int    `csv:"wheelchair_capacity,omitempty"`
	LowFloor               int    `csv:"low_floor,omitempty"`               // 0 or 1
	HasRamp                int    `csv:"has_ramp,omitempty"`                // 0 or 1
	
	// Capacity
	SeatingCapacity        int    `csv:"seating_capacity,omitempty"`
	StandingCapacity       int    `csv:"standing_capacity,omitempty"`
	TotalCapacity          int    `csv:"total_capacity,omitempty"`
	
	// Bicycle support
	BikesAllowed           string `csv:"bikes_allowed,omitempty"`           // 0, 1, 2
	BikeCapacity           int    `csv:"bike_capacity,omitempty"`
	BikeStorageType        string `csv:"bike_storage_type,omitempty"`       // racks, designated_area, etc.
	
	// Environmental
	FuelType               string `csv:"fuel_type,omitempty"`               // diesel, electric, hybrid, etc.
	EmissionClass          string `csv:"emission_class,omitempty"`          // Euro 6, etc.
	
	// Passenger amenities
	HasAirConditioning     int    `csv:"has_air_conditioning,omitempty"`    // 0 or 1
	HasWiFi                int    `csv:"has_wifi,omitempty"`                // 0 or 1
	HasUSBPower            int    `csv:"has_usb_power,omitempty"`           // 0 or 1
	HasAudioSystem         int    `csv:"has_audio_system,omitempty"`        // 0 or 1
	HasVisualDisplay       int    `csv:"has_visual_display,omitempty"`      // 0 or 1
	
	// Physical characteristics
	VehicleLength          float64 `csv:"vehicle_length,omitempty"`         // meters
	VehicleWidth           float64 `csv:"vehicle_width,omitempty"`          // meters
	VehicleHeight          float64 `csv:"vehicle_height,omitempty"`         // meters
	VehicleWeight          float64 `csv:"vehicle_weight,omitempty"`         // kg
}

// TripVehicle links trips to vehicle types
type TripVehicle struct {
	TripID        string `csv:"trip_id"`
	VehicleTypeID string `csv:"vehicle_type_id"`
}

// StopAccessibility provides detailed accessibility information for stops
type StopAccessibility struct {
	StopID                    string `csv:"stop_id"`
	WheelchairBoarding        string `csv:"wheelchair_boarding"`           // 0, 1, 2 (override from stops.txt)
	StepFreeAccess            int    `csv:"step_free_access,omitempty"`    // 0 or 1
	LiftAccess                int    `csv:"lift_access,omitempty"`         // 0 or 1
	EscalatorAccess           int    `csv:"escalator_access,omitempty"`    // 0 or 1
	TactileGuidance           int    `csv:"tactile_guidance,omitempty"`    // 0 or 1
	AudioAnnouncements        int    `csv:"audio_announcements,omitempty"` // 0 or 1
	VisualInformation         int    `csv:"visual_information,omitempty"`  // 0 or 1
	InductionLoop             int    `csv:"induction_loop,omitempty"`      // 0 or 1
	
	// Platform specifications
	PlatformHeight            int    `csv:"platform_height,omitempty"`     // cm above rail level
	PlatformLength            int    `csv:"platform_length,omitempty"`     // meters
	PlatformWidth             int    `csv:"platform_width,omitempty"`      // meters
	
	// Passenger facilities
	ShelterAvailable          int    `csv:"shelter_available,omitempty"`   // 0 or 1
	SeatingAvailable          int    `csv:"seating_available,omitempty"`   // 0 or 1
	TicketMachine             int    `csv:"ticket_machine,omitempty"`      // 0 or 1
	
	// Accessibility equipment
	AccessibleParkingSpaces   int    `csv:"accessible_parking_spaces,omitempty"`
	AccessibleToilets         int    `csv:"accessible_toilets,omitempty"`   // 0 or 1
	ChangingTable             int    `csv:"changing_table,omitempty"`       // 0 or 1
}

// RouteAccessibility provides route-level accessibility information
type RouteAccessibility struct {
	RouteID                   string `csv:"route_id"`
	WheelchairAccessible      string `csv:"wheelchair_accessible,omitempty"` // 0, 1, 2
	LowFloorVehicles          int    `csv:"low_floor_vehicles,omitempty"`    // 0 or 1
	AudioAnnouncements        int    `csv:"audio_announcements,omitempty"`   // 0 or 1
	VisualInformation         int    `csv:"visual_information,omitempty"`    // 0 or 1
	StopAnnouncements         int    `csv:"stop_announcements,omitempty"`    // 0 or 1
	BikesAllowed              string `csv:"bikes_allowed,omitempty"`         // 0, 1, 2
	WheelchairAccessibleNote  string `csv:"wheelchair_accessible_note,omitempty"`
}

// ServiceAlterationGTFS represents planned service changes (custom extension)
// Named differently to avoid conflict with NeTEx ServiceAlteration
type ServiceAlterationGTFS struct {
	AlterationID     string `csv:"alteration_id"`
	ServiceID        string `csv:"service_id,omitempty"`
	RouteID          string `csv:"route_id,omitempty"`
	TripID           string `csv:"trip_id,omitempty"`
	StopID           string `csv:"stop_id,omitempty"`
	
	AlterationType   string `csv:"alteration_type"`       // cancellation, diversion, extra_service
	ValidFrom        string `csv:"valid_from"`            // YYYYMMDD
	ValidTo          string `csv:"valid_to"`              // YYYYMMDD
	
	Description      string `csv:"description,omitempty"`
	Reason           string `csv:"reason,omitempty"`
	AlternativeInfo  string `csv:"alternative_info,omitempty"`
	
	// Replacement service
	ReplacementRouteID string `csv:"replacement_route_id,omitempty"`
	ReplacementTripID  string `csv:"replacement_trip_id,omitempty"`
}

// MultimodalInfo represents connections between different transport modes
type MultimodalInfo struct {
	ConnectionID           string  `csv:"connection_id"`
	FromStopID            string  `csv:"from_stop_id"`
	ToStopID              string  `csv:"to_stop_id"`
	ConnectionType        string  `csv:"connection_type"`        // walking, transfer, interchange
	MinTransferTime       int     `csv:"min_transfer_time,omitempty"` // seconds
	WalkingDistance       int     `csv:"walking_distance,omitempty"`  // meters
	WalkingTime           int     `csv:"walking_time,omitempty"`      // seconds
	IsStepFree            int     `csv:"is_step_free,omitempty"`      // 0 or 1
	IsWheelchairAccessible int    `csv:"is_wheelchair_accessible,omitempty"` // 0 or 1
	HasSignage            int     `csv:"has_signage,omitempty"`       // 0 or 1
	IsCovered             int     `csv:"is_covered,omitempty"`        // 0 or 1
	
	// Geographic path
	PathCoordinates       string  `csv:"path_coordinates,omitempty"`  // WKT LineString
}

// OperatorInfo provides detailed operator information beyond basic agency
type OperatorInfo struct {
	OperatorID            string `csv:"operator_id"`
	OperatorName          string `csv:"operator_name"`
	OperatorShortName     string `csv:"operator_short_name,omitempty"`
	OperatorURL           string `csv:"operator_url,omitempty"`
	OperatorPhone         string `csv:"operator_phone,omitempty"`
	OperatorEmail         string `csv:"operator_email,omitempty"`
	
	// Legal entity information
	LegalName             string `csv:"legal_name,omitempty"`
	CompanyNumber         string `csv:"company_number,omitempty"`
	VATNumber             string `csv:"vat_number,omitempty"`
	
	// Licensing information
	LicenseNumber         string `csv:"license_number,omitempty"`
	LicenseAuthority      string `csv:"license_authority,omitempty"`
	LicenseValidFrom      string `csv:"license_valid_from,omitempty"` // YYYYMMDD
	LicenseValidTo        string `csv:"license_valid_to,omitempty"`   // YYYYMMDD
	
	// Service area
	ServiceArea           string `csv:"service_area,omitempty"`
	CountryCode           string `csv:"country_code,omitempty"` // ISO 3166-1 alpha-2
	RegionCode            string `csv:"region_code,omitempty"`
}

// RouteOperator links routes to detailed operator information
type RouteOperator struct {
	RouteID    string `csv:"route_id"`
	OperatorID string `csv:"operator_id"`
	ValidFrom  string `csv:"valid_from,omitempty"` // YYYYMMDD
	ValidTo    string `csv:"valid_to,omitempty"`   // YYYYMMDD
}

// QualityIndicator represents service quality metrics (custom extension)
type QualityIndicator struct {
	RouteID               string  `csv:"route_id,omitempty"`
	TripID                string  `csv:"trip_id,omitempty"`
	StopID                string  `csv:"stop_id,omitempty"`
	
	// Punctuality metrics
	OnTimePerformance     float64 `csv:"on_time_performance,omitempty"`     // percentage
	AverageDelay          int     `csv:"average_delay,omitempty"`           // seconds
	
	// Reliability metrics  
	ServiceReliability    float64 `csv:"service_reliability,omitempty"`     // percentage
	CancellationRate      float64 `csv:"cancellation_rate,omitempty"`       // percentage
	
	// Customer satisfaction
	CustomerSatisfaction  float64 `csv:"customer_satisfaction,omitempty"`   // score 1-10
	
	// Accessibility compliance
	AccessibilityCompliance float64 `csv:"accessibility_compliance,omitempty"` // percentage
	
	// Environmental impact
	CO2Emissions          float64 `csv:"co2_emissions,omitempty"`           // grams per passenger-km
	EnergyConsumption     float64 `csv:"energy_consumption,omitempty"`      // kWh per km
	
	// Data quality
	DataCompleteness      float64 `csv:"data_completeness,omitempty"`       // percentage
	DataAccuracy          float64 `csv:"data_accuracy,omitempty"`           // percentage
	
	ValidFrom             string  `csv:"valid_from,omitempty"` // YYYYMMDD
	ValidTo               string  `csv:"valid_to,omitempty"`   // YYYYMMDD
}