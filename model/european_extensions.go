package model

import "encoding/xml"

// European NeTEx profile extensions for advanced features
// These models extend the basic NeTEx models with European-specific features

// ServiceAlteration represents changes to regular service patterns
type ServiceAlteration struct {
	XMLName        xml.Name `xml:"ServiceAlteration"`
	ID             string   `xml:"id,attr"`
	Version        string   `xml:"version,attr"`
	Name           string   `xml:"Name,omitempty"`
	Description    string   `xml:"Description,omitempty"`
	Reason         string   `xml:"Reason,omitempty"`
	AlterationType string   `xml:"AlterationType,omitempty"` // cancellation, diversion, extraJourney, etc.

	// Date validity
	ValidFrom string `xml:"ValidFrom,omitempty"`
	ValidTo   string `xml:"ValidTo,omitempty"`

	// Affected services
	AffectedJourneyPatterns []string `xml:"journeyPatterns>JourneyPatternRef,omitempty"`
	AffectedServiceJourneys []string `xml:"serviceJourneys>ServiceJourneyRef,omitempty"`
	AffectedStopPoints      []string `xml:"stopPoints>StopPointRef,omitempty"`

	// Replacement service info
	ReplacementService *ReplacementService `xml:"ReplacementService,omitempty"`
}

// ReplacementService describes alternative transport arrangements
type ReplacementService struct {
	XMLName     xml.Name `xml:"ReplacementService"`
	ID          string   `xml:"id,attr"`
	Description string   `xml:"Description,omitempty"`

	// Transport details
	TransportMode    string `xml:"TransportMode,omitempty"`
	TransportSubmode string `xml:"TransportSubmode,omitempty"`

	// Affected stops
	FromStopPointRef string `xml:"FromStopPointRef,omitempty"`
	ToStopPointRef   string `xml:"ToStopPointRef,omitempty"`
}

// EuropeanAccessibilityFeatures represents extended accessibility information
// This extends the basic AccessibilityAssessment with European-specific features
type EuropeanAccessibilityFeatures struct {
	XMLName xml.Name `xml:"EuropeanAccessibilityFeatures"`

	// Additional accessibility tools
	AccessibilityTools *AccessibilityTools `xml:"accessibilityTools,omitempty"`
}

// SuitableFor describes what types of mobility equipment are supported
type SuitableFor struct {
	XMLName xml.Name `xml:"suitableFor"`

	SuitableForWheelchairs  string `xml:"SuitableForWheelchairs,omitempty"`  // true, false, unknown
	SuitableForHeavyLuggage string `xml:"SuitableForHeavyLuggage,omitempty"` // true, false, unknown
	SuitableForCycles       string `xml:"SuitableForCycles,omitempty"`       // true, false, unknown
	SuitableForPrams        string `xml:"SuitableForPrams,omitempty"`        // true, false, unknown
}

// AccessibilityTools describes available assistance tools
type AccessibilityTools struct {
	XMLName xml.Name `xml:"accessibilityTools"`

	AudioInformationAvailable  string `xml:"AudioInformationAvailable,omitempty"`  // true, false, unknown
	VisualInformationAvailable string `xml:"VisualInformationAvailable,omitempty"` // true, false, unknown
	IndoorNavigationAvailable  string `xml:"IndoorNavigationAvailable,omitempty"`  // true, false, unknown
}

// OperatorView represents operator-specific service information
type OperatorView struct {
	XMLName xml.Name `xml:"OperatorView"`
	ID      string   `xml:"id,attr"`
	Version string   `xml:"version,attr"`

	// Operator information
	OperatorRef string `xml:"OperatorRef,omitempty"`

	// Service visibility
	PrivateCode     string `xml:"PrivateCode,omitempty"`     // Internal operator code
	InternalLineRef string `xml:"InternalLineRef,omitempty"` // Internal line reference

	// Commercial information
	BrandingRef string `xml:"BrandingRef,omitempty"` // Brand/livery reference

	// Operational details
	OperationalContext string `xml:"OperationalContext,omitempty"` // commercial, school, works, etc.
}

// Notice represents informational messages about services
type Notice struct {
	XMLName xml.Name `xml:"Notice"`
	ID      string   `xml:"id,attr"`
	Version string   `xml:"version,attr"`

	// Notice content
	Text         string `xml:"Text,omitempty"`
	PublicCode   string `xml:"PublicCode,omitempty"`
	TypeOfNotice string `xml:"TypeOfNotice,omitempty"` // information, warning, disruption, etc.

	// Validity
	ValidFrom string `xml:"ValidFrom,omitempty"`
	ValidTo   string `xml:"ValidTo,omitempty"`

	// Delivery channels
	CanBeAdvertised  string                  `xml:"CanBeAdvertised,omitempty"` // true, false
	DeliveryVariants []NoticeDeliveryVariant `xml:"deliveryVariants>NoticeDeliveryVariant,omitempty"`
}

// NoticeDeliveryVariant represents different ways to deliver a notice
type NoticeDeliveryVariant struct {
	XMLName xml.Name `xml:"NoticeDeliveryVariant"`

	DeliveryVariantMediaType string `xml:"DeliveryVariantMediaType,omitempty"` // printedMedia, mobile, web, etc.
	NoticeText               string `xml:"NoticeText,omitempty"`
}

// FlexibleService represents demand-responsive transport services
type FlexibleService struct {
	XMLName xml.Name `xml:"FlexibleService"`
	ID      string   `xml:"id,attr"`
	Version string   `xml:"version,attr"`

	// Service type
	FlexibleServiceType string `xml:"FlexibleServiceType,omitempty"` // dynamicPassingTimes, fixedHeadwayFrequency, etc.

	// Booking arrangements
	BookingArrangements *BookingArrangements `xml:"BookingArrangements,omitempty"`

	// Service areas
	FlexibleArea *FlexibleArea `xml:"FlexibleArea,omitempty"`

	// Cancellation policy
	CancellationPolicy string `xml:"CancellationPolicy,omitempty"`
}

// BookingArrangements describes how to book flexible services
type BookingArrangements struct {
	XMLName xml.Name `xml:"BookingArrangements"`

	BookingContact       *BookingContact `xml:"BookingContact,omitempty"`
	BookingMethods       []string        `xml:"BookingMethods>BookingMethod,omitempty"` // online, phone, etc.
	LatestBookingTime    string          `xml:"LatestBookingTime,omitempty"`            // ISO duration before travel
	MinimumBookingPeriod string          `xml:"MinimumBookingPeriod,omitempty"`         // ISO duration
	BookingNote          string          `xml:"BookingNote,omitempty"`
}

// BookingContact provides contact information for bookings
type BookingContact struct {
	XMLName xml.Name `xml:"BookingContact"`

	ContactPerson  string `xml:"ContactPerson,omitempty"`
	Phone          string `xml:"Phone,omitempty"`
	Email          string `xml:"Email,omitempty"`
	Url            string `xml:"Url,omitempty"`
	FurtherDetails string `xml:"FurtherDetails,omitempty"`
}

// FlexibleArea defines the service area for flexible transport
type FlexibleArea struct {
	XMLName xml.Name `xml:"FlexibleArea"`
	ID      string   `xml:"id,attr"`

	Name        string   `xml:"Name,omitempty"`
	Description string   `xml:"Description,omitempty"`
	Polygon     *Polygon `xml:"Polygon,omitempty"`

	// Hail and ride areas
	FlexibleQuays []FlexibleQuay `xml:"FlexibleQuays>FlexibleQuay,omitempty"`
}

// FlexibleQuay represents a flexible boarding point
type FlexibleQuay struct {
	XMLName xml.Name `xml:"FlexibleQuay"`
	ID      string   `xml:"id,attr"`

	Name         string    `xml:"Name,omitempty"`
	Centroid     *Centroid `xml:"Centroid,omitempty"`
	FlexibleArea string    `xml:"FlexibleArea,omitempty"` // hailAndRide, demandAndResponseArea, etc.
}

// Polygon represents a geographic polygon for areas
type Polygon struct {
	XMLName xml.Name `xml:"Polygon"`

	Exterior *LinearRing `xml:"exterior>LinearRing,omitempty"`
}

// LinearRing represents a closed ring of coordinates
type LinearRing struct {
	XMLName xml.Name `xml:"LinearRing"`

	PosList string `xml:"posList,omitempty"` // Space-separated coordinate pairs
}

// VehicleType represents detailed vehicle specifications
type VehicleType struct {
	XMLName xml.Name `xml:"VehicleType"`
	ID      string   `xml:"id,attr"`
	Version string   `xml:"version,attr"`

	Name        string `xml:"Name,omitempty"`
	Description string `xml:"Description,omitempty"`

	// Physical characteristics
	VehicleTypeCapacity *VehicleTypeCapacity `xml:"capacities>VehicleTypeCapacity,omitempty"`
	Length              string               `xml:"Length,omitempty"`
	Width               string               `xml:"Width,omitempty"`
	Height              string               `xml:"Height,omitempty"`
	Weight              string               `xml:"Weight,omitempty"`

	// Accessibility
	LowFloor             string `xml:"LowFloor,omitempty"`             // true, false, partial
	HasLiftOrRamp        string `xml:"HasLiftOrRamp,omitempty"`        // true, false
	WheelchairAccessible string `xml:"WheelchairAccessible,omitempty"` // true, false, partial

	// Fuel and emissions
	FuelType      string `xml:"FuelType,omitempty"`      // diesel, electric, hybrid, etc.
	EmissionClass string `xml:"EmissionClass,omitempty"` // Euro 6, etc.

	// Passenger facilities
	HasAirConditioning string `xml:"HasAirConditioning,omitempty"` // true, false
	HasWifi            string `xml:"HasWifi,omitempty"`            // true, false
	HasUSBPower        string `xml:"HasUSBPower,omitempty"`        // true, false

	// Bike facilities
	CyclesAllowed         string                 `xml:"CyclesAllowed,omitempty"`    // true, false
	CycleStorageType      string                 `xml:"CycleStorageType,omitempty"` // racks, designated area, etc.
	CycleStorageEquipment *CycleStorageEquipment `xml:"CycleStorageEquipment,omitempty"`
}

// VehicleTypeCapacity describes passenger capacity details
type VehicleTypeCapacity struct {
	XMLName xml.Name `xml:"VehicleTypeCapacity"`

	// Seating capacity
	TotalCapacity    string `xml:"TotalCapacity,omitempty"`
	SeatingCapacity  string `xml:"SeatingCapacity,omitempty"`
	StandingCapacity string `xml:"StandingCapacity,omitempty"`

	// Special needs capacity
	WheelchairPlaces string `xml:"WheelchairPlaces,omitempty"`
	PramPlaces       string `xml:"PramPlaces,omitempty"`
	BicyclePlaces    string `xml:"BicyclePlaces,omitempty"`

	// Luggage capacity
	LuggageCapacity string `xml:"LuggageCapacity,omitempty"`
}

// CycleStorageEquipment describes bicycle storage facilities
type CycleStorageEquipment struct {
	XMLName xml.Name `xml:"CycleStorageEquipment"`
	ID      string   `xml:"id,attr"`

	CycleStorageType          string `xml:"CycleStorageType,omitempty"` // racks, hooks, lockers
	NumberOfSpaces            string `xml:"NumberOfSpaces,omitempty"`
	CycleStorageEquipmentType string `xml:"CycleStorageEquipmentType,omitempty"` // indoor, outdoor, covered
}

// PassengerInformation represents real-time passenger information systems
type PassengerInformation struct {
	XMLName xml.Name `xml:"PassengerInformation"`
	ID      string   `xml:"id,attr"`

	// Information types
	InfoChannelType string `xml:"InfoChannelType,omitempty"` // announcements, displays, mobile

	// Language support
	Languages []string `xml:"Languages>Language,omitempty"`

	// Content types
	ArrivalInfo    string `xml:"ArrivalInfo,omitempty"`    // true, false
	DepartureInfo  string `xml:"DepartureInfo,omitempty"`  // true, false
	ConnectionInfo string `xml:"ConnectionInfo,omitempty"` // true, false
	DisruptionInfo string `xml:"DisruptionInfo,omitempty"` // true, false

	// Accessibility features
	AudioInfo   string `xml:"AudioInfo,omitempty"`   // true, false
	VisualInfo  string `xml:"VisualInfo,omitempty"`  // true, false
	TactileInfo string `xml:"TactileInfo,omitempty"` // true, false
}
