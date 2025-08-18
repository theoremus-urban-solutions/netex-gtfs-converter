package model

// GTFS data models

// Agency represents a GTFS agency
type Agency struct {
	AgencyID       string
	AgencyName     string
	AgencyURL      string
	AgencyTimezone string
	AgencyLang     string
	AgencyPhone    string
	AgencyFareURL  string
	AgencyEmail    string
}

// GtfsRoute represents a GTFS route
type GtfsRoute struct {
	RouteID           string
	AgencyID          string
	RouteShortName    string
	RouteLongName     string
	RouteDesc         string
	RouteType         int
	RouteURL          string
	RouteColor        string
	RouteTextColor    string
	RouteSortOrder    int
	ContinuousPickup  string
	ContinuousDropOff string
}

// Trip represents a GTFS trip
type Trip struct {
	RouteID              string
	ServiceID            string
	TripID               string
	TripHeadsign         string
	TripShortName        string
	DirectionID          string
	BlockID              string
	ShapeID              string
	WheelchairAccessible string
	BikesAllowed         string
}

// Stop represents a GTFS stop
type Stop struct {
	StopID             string
	StopCode           string
	StopName           string
	StopDesc           string
	StopLat            float64
	StopLon            float64
	ZoneID             string
	StopURL            string
	LocationType       string
	ParentStation      string
	WheelchairBoarding string
	LevelID            string
	PlatformCode       string
}

// StopTime represents a GTFS stop time
type StopTime struct {
	TripID            string
	ArrivalTime       string
	DepartureTime     string
	StopID            string
	StopSequence      int
	StopHeadsign      string
	PickupType        string
	DropOffType       string
	ContinuousPickup  string
	ContinuousDropOff string
	ShapeDistTraveled float64
	Timepoint         string
}

// Calendar represents a GTFS calendar
type Calendar struct {
	ServiceID string
	Monday    bool
	Tuesday   bool
	Wednesday bool
	Thursday  bool
	Friday    bool
	Saturday  bool
	Sunday    bool
	StartDate string
	EndDate   string
}

// CalendarDate represents a GTFS calendar date
type CalendarDate struct {
	ServiceID     string
	Date          string
	ExceptionType int
}

// Transfer represents a GTFS transfer
type Transfer struct {
	FromStopID      string
	ToStopID        string
	TransferType    int
	MinTransferTime int
	FromRouteID     string
	ToRouteID       string
	FromTripID      string
	ToTripID        string
}

// Shape represents a GTFS shape
type Shape struct {
	ShapeID           string
	ShapePtLat        float64
	ShapePtLon        float64
	ShapePtSequence   int
	ShapeDistTraveled float64
}

// FeedInfo represents GTFS feed information
type FeedInfo struct {
	FeedPublisherName string
	FeedPublisherURL  string
	FeedLang          string
	FeedStartDate     string
	FeedEndDate       string
	FeedVersion       string
	FeedContactEmail  string
	FeedContactURL    string
}

// Frequency represents a GTFS frequency
type Frequency struct {
	TripID      string
	StartTime   string
	EndTime     string
	HeadwaySecs int
	ExactTimes  string
}

// FareAttribute represents a GTFS fare attribute
type FareAttribute struct {
	AgencyID         string
	FareID           string
	Price            float64
	CurrencyType     string
	PaymentMethod    int
	Transfers        int
	TransferDuration int
}

// FareRule represents a GTFS fare rule
type FareRule struct {
	FareID        string
	RouteID       string
	OriginID      string
	DestinationID string
	ContainsID    string
}

// Level represents a GTFS level
type Level struct {
	LevelID    string
	LevelIndex float64
	LevelName  string
}

// Pathway represents a GTFS pathway
type Pathway struct {
	PathwayID            string
	FromStopID           string
	ToStopID             string
	PathwayMode          int
	IsBidirectional      int
	Length               float64
	TraversalTime        int
	StairCount           int
	MaxSlope             float64
	MinWidth             float64
	SignpostedAs         string
	ReversedSignpostedAs string
}

// Translation represents a GTFS translation
type Translation struct {
	TableName   string
	FieldName   string
	Language    string
	Translation string
	RecordID    string
	RecordSubID string
	FieldValue  string
}

// Attribution represents a GTFS attribution
type Attribution struct {
	AttributionID    string
	AgencyID         string
	RouteID          string
	TripID           string
	OrganizationName string
	IsProducer       int
	IsOperator       int
	IsAuthority      int
	AttributionURL   string
	AttributionEmail string
	AttributionPhone string
}
