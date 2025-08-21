package model

import (
	"encoding/xml"
)

// NeTEx XML structures

// Line represents a NeTEx Line
type Line struct {
	XMLName          xml.Name      `xml:"Line"`
	ID               string        `xml:"id,attr"`
	Version          string        `xml:"version,attr"`
	Name             string        `xml:"Name"`
	ShortName        string        `xml:"ShortName"`
	PublicCode       string        `xml:"PublicCode"`
	Description      string        `xml:"Description"`
	URL              string        `xml:"Url"`
	TransportMode    string        `xml:"TransportMode"`
	TransportSubmode string        `xml:"TransportSubmode"`
	AuthorityRef     string        `xml:"AuthorityRef"`
	OperatorRef      string        `xml:"OperatorRef"`
	NetworkRef       string        `xml:"NetworkRef"`
	BrandingRef      string        `xml:"BrandingRef"`
	Presentation     *Presentation `xml:"Presentation"`
}

// Presentation represents NeTEx presentation information
type Presentation struct {
	XMLName    xml.Name `xml:"Presentation"`
	Colour     string   `xml:"Colour"`
	TextColour string   `xml:"TextColour"`
}

// Network represents a NeTEx Network
type Network struct {
	XMLName      xml.Name            `xml:"Network"`
	ID           string              `xml:"id,attr"`
	Version      string              `xml:"version,attr"`
	Name         string              `xml:"Name"`
	ShortName    string              `xml:"ShortName"`
	Description  string              `xml:"Description"`
	PrivateCode  string              `xml:"PrivateCode"`
	Members      *NetworkMembers     `xml:"members"`
	AuthorityRef NetworkAuthorityRef `xml:"AuthorityRef"`
}

// NetworkAuthorityRef represents an authority reference in a network
type NetworkAuthorityRef struct {
	XMLName    xml.Name `xml:"AuthorityRef"`
	Ref        string   `xml:"ref,attr"`
	VersionRef string   `xml:"versionRef,attr"`
}

// NetworkMembers represents the members of a Network
type NetworkMembers struct {
	XMLName xml.Name         `xml:"members"`
	LineRef []NetworkLineRef `xml:"LineRef"`
}

// NetworkLineRef represents a line reference in a network
type NetworkLineRef struct {
	XMLName    xml.Name `xml:"LineRef"`
	Ref        string   `xml:"ref,attr"`
	VersionRef string   `xml:"versionRef,attr"`
}

// Authority represents a NeTEx Authority
type Authority struct {
	XMLName        xml.Name        `xml:"Authority"`
	ID             string          `xml:"id,attr"`
	Version        string          `xml:"version,attr"`
	Name           string          `xml:"Name"`
	ShortName      string          `xml:"ShortName"`
	Description    string          `xml:"Description"`
	URL            string          `xml:"Url"`
	ContactDetails *ContactDetails `xml:"ContactDetails"`
}

// ContactDetails represents contact information
type ContactDetails struct {
	XMLName xml.Name `xml:"ContactDetails"`
	Phone   string   `xml:"Phone"`
	Email   string   `xml:"Email"`
	URL     string   `xml:"Url"`
}

// ServiceJourney represents a NeTEx ServiceJourney
type ServiceJourney struct {
	XMLName           xml.Name                 `xml:"ServiceJourney"`
	ID                string                   `xml:"id,attr"`
	Version           string                   `xml:"version,attr"`
	JourneyPatternRef ServiceJourneyPatternRef `xml:"JourneyPatternRef"`
	LineRef           ServiceJourneyLineRef    `xml:"LineRef"`
	OperatorRef       string                   `xml:"OperatorRef"`
	ServiceAlteration string                   `xml:"ServiceAlteration"`
	Monitored         bool                     `xml:"Monitored"`
	PassingTimes      *PassingTimes            `xml:"passingTimes"`
	DayTypes          *DayTypes                `xml:"dayTypes"`
	NoticeAssignments *NoticeAssignments       `xml:"NoticeAssignments"`
}

// ServiceJourneyPatternRef represents a journey pattern reference in a service journey
type ServiceJourneyPatternRef struct {
	XMLName    xml.Name `xml:"JourneyPatternRef"`
	Ref        string   `xml:"ref,attr"`
	VersionRef string   `xml:"versionRef,attr"`
}

// ServiceJourneyLineRef represents a line reference in a service journey
type ServiceJourneyLineRef struct {
	XMLName    xml.Name `xml:"LineRef"`
	Ref        string   `xml:"ref,attr"`
	VersionRef string   `xml:"versionRef,attr"`
}

// PassingTimes represents the passing times of a service journey
type PassingTimes struct {
	XMLName               xml.Name                `xml:"passingTimes"`
	TimetabledPassingTime []TimetabledPassingTime `xml:"TimetabledPassingTime"`
}

// TimetabledPassingTime represents a scheduled passing time
type TimetabledPassingTime struct {
	XMLName                  xml.Name           `xml:"TimetabledPassingTime"`
	ID                       string             `xml:"id,attr"`
	Version                  string             `xml:"version,attr"`
	PointInJourneyPatternRef string             `xml:"PointInJourneyPatternRef"`
	ArrivalTime              string             `xml:"ArrivalTime"`
	DepartureTime            string             `xml:"DepartureTime"`
	EarliestTime             string             `xml:"EarliestTime"`
	LatestTime               string             `xml:"LatestTime"`
	DayOffset                int                `xml:"DayOffset"`
	NoticeAssignments        *NoticeAssignments `xml:"NoticeAssignments"`
}

// DayTypes represents day type assignments
type DayTypes struct {
	XMLName    xml.Name `xml:"dayTypes"`
	DayTypeRef []string `xml:"DayTypeRef"`
}

// NoticeAssignments represents notice assignments
type NoticeAssignments struct {
	XMLName          xml.Name           `xml:"NoticeAssignments"`
	NoticeAssignment []NoticeAssignment `xml:"NoticeAssignment"`
}

// NoticeAssignment represents a notice assignment
type NoticeAssignment struct {
	XMLName   xml.Name `xml:"NoticeAssignment"`
	ID        string   `xml:"id,attr"`
	NoticeRef string   `xml:"NoticeRef"`
}

// JourneyPattern represents a NeTEx JourneyPattern
type JourneyPattern struct {
	XMLName               xml.Name          `xml:"JourneyPattern"`
	ID                    string            `xml:"id,attr"`
	Version               string            `xml:"version,attr"`
	Name                  string            `xml:"Name"`
	Description           string            `xml:"Description"`
	PrivateCode           string            `xml:"PrivateCode"`
	RouteRef              string            `xml:"RouteRef"`
	DirectionType         string            `xml:"DirectionType"`
	PointsInSequence      *PointsInSequence `xml:"pointsInSequence"`
	DestinationDisplayRef string            `xml:"DestinationDisplayRef"`
}

// ServiceJourneyPattern represents a NeTEx ServiceJourneyPattern (same structure as JourneyPattern)
type ServiceJourneyPattern struct {
	XMLName               xml.Name                                   `xml:"ServiceJourneyPattern"`
	ID                    string                                     `xml:"id,attr"`
	Version               string                                     `xml:"version,attr"`
	Name                  string                                     `xml:"Name"`
	Description           string                                     `xml:"Description"`
	PrivateCode           string                                     `xml:"PrivateCode"`
	RouteRef              ServiceJourneyPatternRouteRef              `xml:"RouteRef"`
	DirectionType         string                                     `xml:"DirectionType"`
	PointsInSequence      *PointsInSequence                          `xml:"pointsInSequence"`
	DestinationDisplayRef ServiceJourneyPatternDestinationDisplayRef `xml:"DestinationDisplayRef"`
}

// ServiceJourneyPatternRouteRef represents a route reference in a service journey pattern
type ServiceJourneyPatternRouteRef struct {
	XMLName    xml.Name `xml:"RouteRef"`
	Ref        string   `xml:"ref,attr"`
	VersionRef string   `xml:"versionRef,attr"`
}

// ServiceJourneyPatternDestinationDisplayRef represents a destination display reference
type ServiceJourneyPatternDestinationDisplayRef struct {
	XMLName    xml.Name `xml:"DestinationDisplayRef"`
	Ref        string   `xml:"ref,attr"`
	VersionRef string   `xml:"versionRef,attr"`
}

// ToJourneyPattern converts ServiceJourneyPattern to JourneyPattern
func (sjp *ServiceJourneyPattern) ToJourneyPattern() *JourneyPattern {
	return &JourneyPattern{
		ID:                    sjp.ID,
		Version:               sjp.Version,
		Name:                  sjp.Name,
		Description:           sjp.Description,
		PrivateCode:           sjp.PrivateCode,
		RouteRef:              sjp.RouteRef.Ref,
		DirectionType:         sjp.DirectionType,
		PointsInSequence:      sjp.PointsInSequence,
		DestinationDisplayRef: sjp.DestinationDisplayRef.Ref,
	}
}

// PointsInSequence represents the sequence of points in a journey pattern
type PointsInSequence struct {
	XMLName                                                                       xml.Name      `xml:"pointsInSequence"`
	PointInJourneyPatternOrStopPointInJourneyPatternOrTimingPointInJourneyPattern []interface{} `xml:",any"`
}

// ScheduledStopPoint represents a scheduled stop point (may link to Quay/StopPlace elsewhere)
type ScheduledStopPoint struct {
	XMLName      xml.Name `xml:"ScheduledStopPoint"`
	ID           string   `xml:"id,attr"`
	Version      string   `xml:"version,attr"`
	Name         string   `xml:"Name"`
	StopPlaceRef string   `xml:"StopPlaceRef,attr"`
	QuayRef      string   `xml:"QuayRef,attr"`
}

// StopPointInJourneyPattern represents a stop point in a journey pattern
type StopPointInJourneyPattern struct {
	XMLName               xml.Name           `xml:"StopPointInJourneyPattern"`
	ID                    string             `xml:"id,attr"`
	Version               string             `xml:"version,attr"`
	Order                 int                `xml:"Order"`
	ScheduledStopPointRef string             `xml:"ScheduledStopPointRef"`
	DestinationDisplayRef string             `xml:"DestinationDisplayRef"`
	ForAlighting          bool               `xml:"ForAlighting"`
	ForBoarding           bool               `xml:"ForBoarding"`
	IsWaitPoint           bool               `xml:"IsWaitPoint"`
	WaitTime              string             `xml:"WaitTime"`
	NoticeAssignments     *NoticeAssignments `xml:"NoticeAssignments"`
}

// Route represents a NeTEx Route
type Route struct {
	XMLName          xml.Name          `xml:"Route"`
	ID               string            `xml:"id,attr"`
	Version          string            `xml:"version,attr"`
	Name             string            `xml:"Name"`
	ShortName        string            `xml:"ShortName"`
	Description      string            `xml:"Description"`
	LineRef          RouteLineRef      `xml:"LineRef"`
	DirectionType    string            `xml:"DirectionType"`
	PointsInSequence *PointsInSequence `xml:"pointsInSequence"`
}

// RouteLineRef represents a line reference in a route
type RouteLineRef struct {
	XMLName    xml.Name `xml:"LineRef"`
	Ref        string   `xml:"ref,attr"`
	VersionRef string   `xml:"versionRef,attr"`
}

// StopPlace represents a NeTEx StopPlace
type StopPlace struct {
	XMLName                 xml.Name                 `xml:"StopPlace"`
	ID                      string                   `xml:"id,attr"`
	Version                 string                   `xml:"version,attr"`
	Name                    string                   `xml:"Name"`
	ShortName               string                   `xml:"ShortName"`
	Description             string                   `xml:"Description"`
	TransportMode           string                   `xml:"TransportMode"`
	TransportSubmode        string                   `xml:"TransportSubmode"`
	Centroid                *Centroid                `xml:"Centroid"`
	Quays                   *Quays                   `xml:"Quays"`
	AccessibilityAssessment *AccessibilityAssessment `xml:"AccessibilityAssessment"`
	NoticeAssignments       *NoticeAssignments       `xml:"NoticeAssignments"`
}

// Quays represents a collection of quays
type Quays struct {
	XMLName xml.Name `xml:"Quays"`
	Quay    []Quay   `xml:"Quay"`
}

// Quay represents a NeTEx Quay
type Quay struct {
	XMLName                 xml.Name                 `xml:"Quay"`
	ID                      string                   `xml:"id,attr"`
	Version                 string                   `xml:"version,attr"`
	Name                    string                   `xml:"Name"`
	ShortName               string                   `xml:"ShortName"`
	Description             string                   `xml:"Description"`
	PublicCode              string                   `xml:"PublicCode"`
	Centroid                *Centroid                `xml:"Centroid"`
	AccessibilityAssessment *AccessibilityAssessment `xml:"AccessibilityAssessment"`
	NoticeAssignments       *NoticeAssignments       `xml:"NoticeAssignments"`
}

// Centroid represents a geometric centroid
type Centroid struct {
	XMLName  xml.Name  `xml:"Centroid"`
	Location *Location `xml:"Location"`
}

// Location represents a geographic location
type Location struct {
	XMLName   xml.Name `xml:"Location"`
	Longitude float64  `xml:"Longitude"`
	Latitude  float64  `xml:"Latitude"`
}

// AccessibilityAssessment represents accessibility information
type AccessibilityAssessment struct {
	XMLName     xml.Name     `xml:"AccessibilityAssessment"`
	Limitations *Limitations `xml:"Limitations"`
}

// Limitations represents accessibility limitations
type Limitations struct {
	XMLName                 xml.Name                 `xml:"Limitations"`
	AccessibilityLimitation *AccessibilityLimitation `xml:"AccessibilityLimitation"`
}

// AccessibilityLimitation represents specific accessibility limitations
type AccessibilityLimitation struct {
	XMLName                 xml.Name `xml:"AccessibilityLimitation"`
	WheelchairAccess        string   `xml:"WheelchairAccess"`
	StepFreeAccess          string   `xml:"StepFreeAccess"`
	EscalatorFreeAccess     string   `xml:"EscalatorFreeAccess"`
	LiftFreeAccess          string   `xml:"LiftFreeAccess"`
	AudibleSignalsAvailable string   `xml:"AudibleSignalsAvailable"`
	VisualSignalsAvailable  string   `xml:"VisualSignalsAvailable"`
}

// DestinationDisplay represents a NeTEx DestinationDisplay
type DestinationDisplay struct {
	XMLName   xml.Name `xml:"DestinationDisplay"`
	ID        string   `xml:"id,attr"`
	Version   string   `xml:"version,attr"`
	FrontText string   `xml:"FrontText"`
	SideText  string   `xml:"SideText"`
	Vias      *Vias    `xml:"Vias"`
}

// Vias represents via information
type Vias struct {
	XMLName xml.Name `xml:"Vias"`
	Via     []Via    `xml:"Via"`
}

// Via represents a via point
type Via struct {
	XMLName               xml.Name `xml:"Via"`
	DestinationDisplayRef string   `xml:"DestinationDisplayRef"`
}

// ServiceJourneyInterchange represents a service journey interchange
type ServiceJourneyInterchange struct {
	XMLName             xml.Name `xml:"ServiceJourneyInterchange"`
	ID                  string   `xml:"id,attr"`
	Version             string   `xml:"version,attr"`
	FromJourneyRef      string   `xml:"FromJourneyRef"`
	ToJourneyRef        string   `xml:"ToJourneyRef"`
	FromPointRef        string   `xml:"FromPointRef"`
	ToPointRef          string   `xml:"ToPointRef"`
	StaySeated          bool     `xml:"StaySeated"`
	Guaranteed          bool     `xml:"Guaranteed"`
	MinimumTransferTime string   `xml:"MinimumTransferTime"`
	Priority            int      `xml:"Priority"`
}

// DayType represents a NeTEx DayType
type DayType struct {
	XMLName    xml.Name    `xml:"DayType"`
	ID         string      `xml:"id,attr"`
	Version    string      `xml:"version,attr"`
	Name       string      `xml:"Name"`
	Properties *Properties `xml:"Properties"`
}

// Properties represents day type properties
type Properties struct {
	XMLName       xml.Name        `xml:"Properties"`
	PropertyOfDay []PropertyOfDay `xml:"PropertyOfDay"`
}

// PropertyOfDay represents a property of a day
type PropertyOfDay struct {
	XMLName    xml.Name `xml:"PropertyOfDay"`
	DaysOfWeek string   `xml:"DaysOfWeek"`
}

// OperatingDay represents a NeTEx OperatingDay
type OperatingDay struct {
	XMLName      xml.Name `xml:"OperatingDay"`
	ID           string   `xml:"id,attr"`
	Version      string   `xml:"version,attr"`
	CalendarDate string   `xml:"CalendarDate"`
}

// OperatingPeriod represents a NeTEx OperatingPeriod
type OperatingPeriod struct {
	XMLName  xml.Name `xml:"OperatingPeriod"`
	ID       string   `xml:"id,attr"`
	Version  string   `xml:"version,attr"`
	FromDate string   `xml:"FromDate"`
	ToDate   string   `xml:"ToDate"`
}

// DayTypeAssignment represents a day type assignment
type DayTypeAssignment struct {
	XMLName            xml.Name `xml:"DayTypeAssignment"`
	ID                 string   `xml:"id,attr"`
	Version            string   `xml:"version,attr"`
	DayTypeRef         string   `xml:"DayTypeRef"`
	OperatingDayRef    string   `xml:"OperatingDayRef"`
	OperatingPeriodRef string   `xml:"OperatingPeriodRef"`
	IsAvailable        bool     `xml:"IsAvailable"`
}

// DatedServiceJourney represents a dated service journey
type DatedServiceJourney struct {
	XMLName           xml.Name           `xml:"DatedServiceJourney"`
	ID                string             `xml:"id,attr"`
	Version           string             `xml:"version,attr"`
	ServiceJourneyRef string             `xml:"ServiceJourneyRef"`
	OperatingDayRef   string             `xml:"OperatingDayRef"`
	ServiceAlteration string             `xml:"ServiceAlteration"`
	NoticeAssignments *NoticeAssignments `xml:"NoticeAssignments"`
}

// NeTEx root structures for XML parsing

// PublicationDelivery represents the root NeTEx XML structure
type PublicationDelivery struct {
	XMLName        xml.Name        `xml:"PublicationDelivery"`
	Version        string          `xml:"version,attr"`
	DataObjects    *DataObjects    `xml:"DataObjects"`
	CompositeFrame *CompositeFrame `xml:"CompositeFrame"`
}

// DataObjects contains the main data structures
type DataObjects struct {
	XMLName        xml.Name        `xml:"DataObjects"`
	CompositeFrame *CompositeFrame `xml:"CompositeFrame"`
}

// CompositeFrame contains frames with different types of data
type CompositeFrame struct {
	XMLName xml.Name `xml:"CompositeFrame"`
	ID      string   `xml:"id,attr"`
	Version string   `xml:"version,attr"`
	Frames  *Frames  `xml:"Frames"`
}

// Frames contains different frame types
type Frames struct {
	XMLName              xml.Name              `xml:"Frames"`
	ResourceFrame        *ResourceFrame        `xml:"ResourceFrame"`
	ServiceFrame         *ServiceFrame         `xml:"ServiceFrame"`
	ServiceCalendarFrame *ServiceCalendarFrame `xml:"ServiceCalendarFrame"`
	TimetableFrame       *TimetableFrame       `xml:"TimetableFrame"`
	SiteFrame            *SiteFrame            `xml:"SiteFrame"`
}

// ResourceFrame contains authorities and other resources
type ResourceFrame struct {
	XMLName     xml.Name     `xml:"ResourceFrame"`
	ID          string       `xml:"id,attr"`
	Version     string       `xml:"version,attr"`
	Authorities *Authorities `xml:"Authorities"`
}

// Authorities contains authority definitions
type Authorities struct {
	XMLName   xml.Name    `xml:"Authorities"`
	Authority []Authority `xml:"Authority"`
}

// ServiceFrame contains lines, routes, and journey patterns
type ServiceFrame struct {
	XMLName                    xml.Name                    `xml:"ServiceFrame"`
	ID                         string                      `xml:"id,attr"`
	Version                    string                      `xml:"version,attr"`
	Lines                      *Lines                      `xml:"Lines"`
	Routes                     *Routes                     `xml:"Routes"`
	JourneyPatterns            *JourneyPatterns            `xml:"JourneyPatterns"`
	DestinationDisplays        *DestinationDisplays        `xml:"DestinationDisplays"`
	ScheduledStopPoints        *ScheduledStopPoints        `xml:"ScheduledStopPoints"`
	ServiceJourneyInterchanges *ServiceJourneyInterchanges `xml:"ServiceJourneyInterchanges"`
}

// Lines contains line definitions
type Lines struct {
	XMLName xml.Name `xml:"Lines"`
	Line    []Line   `xml:"Line"`
}

// Routes contains route definitions
type Routes struct {
	XMLName xml.Name `xml:"Routes"`
	Route   []Route  `xml:"Route"`
}

// JourneyPatterns contains journey pattern definitions
type JourneyPatterns struct {
	XMLName        xml.Name         `xml:"JourneyPatterns"`
	JourneyPattern []JourneyPattern `xml:"JourneyPattern"`
}

// DestinationDisplays contains destination display definitions
type DestinationDisplays struct {
	XMLName            xml.Name             `xml:"DestinationDisplays"`
	DestinationDisplay []DestinationDisplay `xml:"DestinationDisplay"`
}

// ScheduledStopPoints contains scheduled stop point definitions
type ScheduledStopPoints struct {
	XMLName            xml.Name             `xml:"ScheduledStopPoints"`
	ScheduledStopPoint []ScheduledStopPoint `xml:"ScheduledStopPoint"`
}

// ServiceJourneyInterchanges contains service journey interchange definitions
type ServiceJourneyInterchanges struct {
	XMLName                   xml.Name                    `xml:"ServiceJourneyInterchanges"`
	ServiceJourneyInterchange []ServiceJourneyInterchange `xml:"ServiceJourneyInterchange"`
}

// ServiceCalendarFrame contains day types and operating periods
type ServiceCalendarFrame struct {
	XMLName            xml.Name            `xml:"ServiceCalendarFrame"`
	ID                 string              `xml:"id,attr"`
	Version            string              `xml:"version,attr"`
	DayTypes           *DayTypesFrame      `xml:"DayTypes"`
	OperatingDays      *OperatingDays      `xml:"OperatingDays"`
	OperatingPeriods   *OperatingPeriods   `xml:"OperatingPeriods"`
	DayTypeAssignments *DayTypeAssignments `xml:"DayTypeAssignments"`
}

// DayTypesFrame contains day type definitions
type DayTypesFrame struct {
	XMLName xml.Name  `xml:"DayTypes"`
	DayType []DayType `xml:"DayType"`
}

// OperatingDays contains operating day definitions
type OperatingDays struct {
	XMLName      xml.Name       `xml:"OperatingDays"`
	OperatingDay []OperatingDay `xml:"OperatingDay"`
}

// OperatingPeriods contains operating period definitions
type OperatingPeriods struct {
	XMLName         xml.Name          `xml:"OperatingPeriods"`
	OperatingPeriod []OperatingPeriod `xml:"OperatingPeriod"`
}

// DayTypeAssignments contains day type assignment definitions
type DayTypeAssignments struct {
	XMLName           xml.Name            `xml:"DayTypeAssignments"`
	DayTypeAssignment []DayTypeAssignment `xml:"DayTypeAssignment"`
}

// TimetableFrame contains service journeys and timetable data
type TimetableFrame struct {
	XMLName              xml.Name              `xml:"TimetableFrame"`
	ID                   string                `xml:"id,attr"`
	Version              string                `xml:"version,attr"`
	ServiceJourneys      *ServiceJourneys      `xml:"ServiceJourneys"`
	DatedServiceJourneys *DatedServiceJourneys `xml:"DatedServiceJourneys"`
	HeadwayJourneyGroups *HeadwayJourneyGroups `xml:"HeadwayJourneyGroups"`
}

// ServiceJourneys contains service journey definitions
type ServiceJourneys struct {
	XMLName        xml.Name         `xml:"ServiceJourneys"`
	ServiceJourney []ServiceJourney `xml:"ServiceJourney"`
}

// DatedServiceJourneys contains dated service journey definitions
type DatedServiceJourneys struct {
	XMLName             xml.Name              `xml:"DatedServiceJourneys"`
	DatedServiceJourney []DatedServiceJourney `xml:"DatedServiceJourney"`
}

// SiteFrame contains stop places and quays
type SiteFrame struct {
	XMLName    xml.Name    `xml:"SiteFrame"`
	ID         string      `xml:"id,attr"`
	Version    string      `xml:"version,attr"`
	StopPlaces *StopPlaces `xml:"stopPlacesGroup"` //nolint:staticcheck // XML tag conflict is intentional for NeTEx compatibility
}

// StopPlaces contains stop place definitions
type StopPlaces struct {
	XMLName   xml.Name    `xml:"StopPlace"`
	StopPlace []StopPlace `xml:"StopPlace"`
}

// HeadwayJourneyGroup represents frequency-based service patterns
type HeadwayJourneyGroup struct {
	XMLName     xml.Name `xml:"HeadwayJourneyGroup"`
	ID          string   `xml:"id,attr"`
	Version     string   `xml:"version,attr"`
	Name        string   `xml:"Name,omitempty"`
	Description string   `xml:"Description,omitempty"`

	// Frequency information
	ScheduledHeadwayInterval string `xml:"ScheduledHeadwayInterval,omitempty"` // ISO 8601 duration
	MaximumHeadway           string `xml:"MaximumHeadway,omitempty"`           // ISO 8601 duration
	MinimumHeadway           string `xml:"MinimumHeadway,omitempty"`           // ISO 8601 duration

	// Operating periods
	FirstDepartureTime string `xml:"FirstDepartureTime,omitempty"` // HH:MM:SS
	LastDepartureTime  string `xml:"LastDepartureTime,omitempty"`  // HH:MM:SS

	// Reference to journey pattern
	JourneyPatternRef string `xml:"JourneyPatternRef,omitempty"`

	// Day types for frequency operation
	DayTypes *DayTypeRefs `xml:"dayTypes,omitempty"`
}

// HeadwayJourneyGroups contains headway journey group definitions
type HeadwayJourneyGroups struct {
	XMLName             xml.Name              `xml:"HeadwayJourneyGroups"`
	HeadwayJourneyGroup []HeadwayJourneyGroup `xml:"HeadwayJourneyGroup"`
}

// RhythmicalJourneyGroup represents rhythmic service patterns (regular intervals)
type RhythmicalJourneyGroup struct {
	XMLName xml.Name `xml:"RhythmicalJourneyGroup"`
	ID      string   `xml:"id,attr"`
	Version string   `xml:"version,attr"`
	Name    string   `xml:"Name,omitempty"`

	// Rhythm information
	RhythmicInterval   string `xml:"RhythmicInterval,omitempty"` // ISO 8601 duration
	FirstDepartureTime string `xml:"FirstDepartureTime,omitempty"`
	LastDepartureTime  string `xml:"LastDepartureTime,omitempty"`

	// Reference to journey pattern
	JourneyPatternRef string `xml:"JourneyPatternRef,omitempty"`

	// Template journey for the group
	TemplateServiceJourney *ServiceJourney `xml:"templateJourneyRef,omitempty"` //nolint:staticcheck // XML tag conflict is intentional for NeTEx compatibility
}

// TemplateServiceJourney represents a template for frequency-based services
type TemplateServiceJourney struct {
	XMLName xml.Name `xml:"TemplateServiceJourney"`
	ID      string   `xml:"id,attr"`
	Version string   `xml:"version,attr"`
	Name    string   `xml:"Name,omitempty"`

	// Basic service journey properties
	LineRef           string `xml:"LineRef,omitempty"`
	JourneyPatternRef string `xml:"JourneyPatternRef,omitempty"`

	// Template passing times (without specific times)
	PassingTimes *PassingTimes `xml:"passingTimes,omitempty"`

	// Day types
	DayTypes *DayTypeRefs `xml:"dayTypes,omitempty"`
}

// FrequencyGroup represents general frequency-based service group
type FrequencyGroup struct {
	XMLName     xml.Name `xml:"FrequencyGroup"`
	ID          string   `xml:"id,attr"`
	Version     string   `xml:"version,attr"`
	Name        string   `xml:"Name,omitempty"`
	Description string   `xml:"Description,omitempty"`

	// Frequency specification
	Frequency *NetexFrequency `xml:"Frequency,omitempty"`

	// Time bands for different frequencies
	TimeBands []TimeBand `xml:"timeBands>TimeBand,omitempty"`

	// Reference to journey pattern
	JourneyPatternRef string `xml:"JourneyPatternRef,omitempty"`
}

// NetexFrequency represents NeTEx frequency information
type NetexFrequency struct {
	XMLName xml.Name `xml:"Frequency"`

	// Frequency interval
	ScheduledHeadwayInterval string `xml:"ScheduledHeadwayInterval,omitempty"` // ISO 8601 duration
	MinimumHeadwayInterval   string `xml:"MinimumHeadwayInterval,omitempty"`   // ISO 8601 duration
	MaximumHeadwayInterval   string `xml:"MaximumHeadwayInterval,omitempty"`   // ISO 8601 duration

	// Exact times or frequency-based
	ExactTime string `xml:"ExactTime,omitempty"` // true/false
}

// DayTypeRefs represents references to day types for frequency services
type DayTypeRefs struct {
	XMLName    xml.Name `xml:"dayTypes"`
	DayTypeRef []string `xml:"DayTypeRef,omitempty"`
}

// TimeBand represents time periods with different frequencies
type TimeBand struct {
	XMLName xml.Name `xml:"TimeBand"`
	ID      string   `xml:"id,attr"`

	// Time period
	StartTime string `xml:"StartTime,omitempty"` // HH:MM:SS
	EndTime   string `xml:"EndTime,omitempty"`   // HH:MM:SS

	// Frequency for this time band
	ScheduledHeadwayInterval string `xml:"ScheduledHeadwayInterval,omitempty"` // ISO 8601 duration
}
