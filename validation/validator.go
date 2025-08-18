package validation

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/theoremus-urban-solutions/netex-gtfs-converter/model"
)

// ValidationSeverity represents the severity level of validation issues
type ValidationSeverity int

const (
	SeverityInfo ValidationSeverity = iota
	SeverityWarning
	SeverityError
	SeverityCritical
)

func (s ValidationSeverity) String() string {
	switch s {
	case SeverityInfo:
		return "INFO"
	case SeverityWarning:
		return "WARNING"
	case SeverityError:
		return "ERROR"
	case SeverityCritical:
		return "CRITICAL"
	default:
		return "UNKNOWN"
	}
}

// ValidationIssue represents a single validation issue
type ValidationIssue struct {
	Severity   ValidationSeverity `json:"severity"`
	Code       string             `json:"code"`
	Message    string             `json:"message"`
	EntityType string             `json:"entity_type"`
	EntityID   string             `json:"entity_id,omitempty"`
	Field      string             `json:"field,omitempty"`
	Value      string             `json:"value,omitempty"`
	Suggestion string             `json:"suggestion,omitempty"`
	Location   string             `json:"location,omitempty"`
	Context    map[string]string  `json:"context,omitempty"`
}

// ValidationReport contains all validation issues and summary statistics
type ValidationReport struct {
	Issues          []ValidationIssue `json:"issues"`
	Summary         ValidationSummary `json:"summary"`
	Timestamp       time.Time         `json:"timestamp"`
	ProcessingStats ProcessingStats   `json:"processing_stats"`
}

// ValidationSummary provides a summary of validation results
type ValidationSummary struct {
	TotalIssues  int                        `json:"total_issues"`
	BySeverity   map[ValidationSeverity]int `json:"by_severity"`
	ByCategory   map[string]int             `json:"by_category"`
	ByEntityType map[string]int             `json:"by_entity_type"`
	IsValid      bool                       `json:"is_valid"`
	HasCritical  bool                       `json:"has_critical"`
	HasErrors    bool                       `json:"has_errors"`
}

// ProcessingStats tracks conversion processing statistics
type ProcessingStats struct {
	EntitiesProcessed  map[string]int `json:"entities_processed"`
	EntitiesConverted  map[string]int `json:"entities_converted"`
	EntitiesSkipped    map[string]int `json:"entities_skipped"`
	ConversionRate     float64        `json:"conversion_rate"`
	ProcessingDuration time.Duration  `json:"processing_duration"`
	MemoryUsageMB      float64        `json:"memory_usage_mb"`
}

// Validator provides comprehensive validation for NeTEx and GTFS data
type Validator struct {
	issues          []ValidationIssue
	processingStats ProcessingStats
	config          ValidationConfig
	patterns        *ValidationPatterns
}

// ValidationConfig controls validation behavior
type ValidationConfig struct {
	EnableStrictMode           bool               `json:"enable_strict_mode"`
	MaxIssuesPerType           int                `json:"max_issues_per_type"`
	EnableContextualValidation bool               `json:"enable_contextual_validation"`
	SeverityThreshold          ValidationSeverity `json:"severity_threshold"`
	EnableGTFSValidation       bool               `json:"enable_gtfs_validation"`
	EnableNeTExValidation      bool               `json:"enable_netex_validation"`
	ValidateReferences         bool               `json:"validate_references"`
	ValidateGeometry           bool               `json:"validate_geometry"`
	ValidateTiming             bool               `json:"validate_timing"`
	ValidateAccessibility      bool               `json:"validate_accessibility"`
}

// ValidationPatterns contains compiled regex patterns for validation
type ValidationPatterns struct {
	GTFSStopID    *regexp.Regexp
	GTFSRouteID   *regexp.Regexp
	GTFSTripID    *regexp.Regexp
	GTFSTime      *regexp.Regexp
	GTFSDate      *regexp.Regexp
	GTFSColor     *regexp.Regexp
	GTFSPhone     *regexp.Regexp
	GTFSEmail     *regexp.Regexp
	GTFSURL       *regexp.Regexp
	NeTExID       *regexp.Regexp
	NeTExDuration *regexp.Regexp
}

// NewValidator creates a new validator with default configuration
func NewValidator() *Validator {
	config := ValidationConfig{
		EnableStrictMode:           false,
		MaxIssuesPerType:           100,
		EnableContextualValidation: true,
		SeverityThreshold:          SeverityInfo,
		EnableGTFSValidation:       true,
		EnableNeTExValidation:      true,
		ValidateReferences:         true,
		ValidateGeometry:           true,
		ValidateTiming:             true,
		ValidateAccessibility:      true,
	}

	patterns := compileValidationPatterns()

	return &Validator{
		issues: make([]ValidationIssue, 0),
		processingStats: ProcessingStats{
			EntitiesProcessed: make(map[string]int),
			EntitiesConverted: make(map[string]int),
			EntitiesSkipped:   make(map[string]int),
		},
		config:   config,
		patterns: patterns,
	}
}

// compileValidationPatterns compiles all regex patterns used for validation
func compileValidationPatterns() *ValidationPatterns {
	return &ValidationPatterns{
		GTFSStopID:    regexp.MustCompile(`^[a-zA-Z0-9_-]+$`),
		GTFSRouteID:   regexp.MustCompile(`^[a-zA-Z0-9_-]+$`),
		GTFSTripID:    regexp.MustCompile(`^[a-zA-Z0-9_-]+$`),
		GTFSTime:      regexp.MustCompile(`^([0-9]{1,2}):([0-5][0-9]):([0-5][0-9])$`),
		GTFSDate:      regexp.MustCompile(`^[0-9]{8}$`),
		GTFSColor:     regexp.MustCompile(`^[0-9A-Fa-f]{6}$`),
		GTFSPhone:     regexp.MustCompile(`^[\+]?[\s\d\-\(\)]{7,20}$`),
		GTFSEmail:     regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`),
		GTFSURL:       regexp.MustCompile(`^https?://[^\s/$.?#].[^\s]*$`),
		NeTExID:       regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_:-]*$`),
		NeTExDuration: regexp.MustCompile(`^PT(?:(\d+)H)?(?:(\d+)M)?(?:(\d+(?:\.\d+)?)S)?$`),
	}
}

// SetConfig updates the validator configuration
func (v *Validator) SetConfig(config ValidationConfig) {
	v.config = config
}

// AddIssue adds a validation issue to the report
func (v *Validator) AddIssue(issue ValidationIssue) {
	// Check if we've exceeded the maximum issues for this type
	typeCount := 0
	for _, existingIssue := range v.issues {
		if existingIssue.Code == issue.Code {
			typeCount++
		}
	}

	if typeCount >= v.config.MaxIssuesPerType {
		return // Skip this issue to prevent spam
	}

	// Only add issues at or above the severity threshold
	if issue.Severity >= v.config.SeverityThreshold {
		v.issues = append(v.issues, issue)
	}
}

// ValidateGTFSAgency validates a GTFS agency entity
func (v *Validator) ValidateGTFSAgency(agency *model.Agency) {
	if agency == nil {
		v.AddIssue(ValidationIssue{
			Severity:   SeverityCritical,
			Code:       "AGENCY_NULL",
			Message:    "Agency entity is null",
			EntityType: "Agency",
		})
		return
	}

	// Validate required fields
	if agency.AgencyID == "" {
		v.AddIssue(ValidationIssue{
			Severity:   SeverityError,
			Code:       "AGENCY_MISSING_ID",
			Message:    "Agency ID is required but missing",
			EntityType: "Agency",
			EntityID:   agency.AgencyID,
			Field:      "agency_id",
		})
	}

	if agency.AgencyName == "" {
		v.AddIssue(ValidationIssue{
			Severity:   SeverityError,
			Code:       "AGENCY_MISSING_NAME",
			Message:    "Agency name is required but missing",
			EntityType: "Agency",
			EntityID:   agency.AgencyID,
			Field:      "agency_name",
		})
	}

	if agency.AgencyURL == "" {
		v.AddIssue(ValidationIssue{
			Severity:   SeverityError,
			Code:       "AGENCY_MISSING_URL",
			Message:    "Agency URL is required but missing",
			EntityType: "Agency",
			EntityID:   agency.AgencyID,
			Field:      "agency_url",
		})
	} else if !v.patterns.GTFSURL.MatchString(agency.AgencyURL) {
		v.AddIssue(ValidationIssue{
			Severity:   SeverityError,
			Code:       "AGENCY_INVALID_URL",
			Message:    "Agency URL format is invalid",
			EntityType: "Agency",
			EntityID:   agency.AgencyID,
			Field:      "agency_url",
			Value:      agency.AgencyURL,
			Suggestion: "URL should start with http:// or https://",
		})
	}

	if agency.AgencyTimezone == "" {
		v.AddIssue(ValidationIssue{
			Severity:   SeverityError,
			Code:       "AGENCY_MISSING_TIMEZONE",
			Message:    "Agency timezone is required but missing",
			EntityType: "Agency",
			EntityID:   agency.AgencyID,
			Field:      "agency_timezone",
		})
	} else {
		// Validate timezone format
		_, err := time.LoadLocation(agency.AgencyTimezone)
		if err != nil {
			v.AddIssue(ValidationIssue{
				Severity:   SeverityError,
				Code:       "AGENCY_INVALID_TIMEZONE",
				Message:    fmt.Sprintf("Invalid timezone: %s", agency.AgencyTimezone),
				EntityType: "Agency",
				EntityID:   agency.AgencyID,
				Field:      "agency_timezone",
				Value:      agency.AgencyTimezone,
				Suggestion: "Use IANA timezone database names (e.g., 'Europe/Oslo')",
			})
		}
	}

	// Validate optional fields
	if agency.AgencyPhone != "" && !v.patterns.GTFSPhone.MatchString(agency.AgencyPhone) {
		v.AddIssue(ValidationIssue{
			Severity:   SeverityWarning,
			Code:       "AGENCY_INVALID_PHONE",
			Message:    "Agency phone number format may be invalid",
			EntityType: "Agency",
			EntityID:   agency.AgencyID,
			Field:      "agency_phone",
			Value:      agency.AgencyPhone,
		})
	}

	if agency.AgencyEmail != "" && !v.patterns.GTFSEmail.MatchString(agency.AgencyEmail) {
		v.AddIssue(ValidationIssue{
			Severity:   SeverityWarning,
			Code:       "AGENCY_INVALID_EMAIL",
			Message:    "Agency email format is invalid",
			EntityType: "Agency",
			EntityID:   agency.AgencyID,
			Field:      "agency_email",
			Value:      agency.AgencyEmail,
		})
	}
}

// ValidateGTFSRoute validates a GTFS route entity
func (v *Validator) ValidateGTFSRoute(route *model.GtfsRoute) {
	if route == nil {
		v.AddIssue(ValidationIssue{
			Severity:   SeverityCritical,
			Code:       "ROUTE_NULL",
			Message:    "Route entity is null",
			EntityType: "Route",
		})
		return
	}

	// Validate required fields
	if route.RouteID == "" {
		v.AddIssue(ValidationIssue{
			Severity:   SeverityError,
			Code:       "ROUTE_MISSING_ID",
			Message:    "Route ID is required but missing",
			EntityType: "Route",
			Field:      "route_id",
		})
	} else if !v.patterns.GTFSRouteID.MatchString(route.RouteID) {
		v.AddIssue(ValidationIssue{
			Severity:   SeverityWarning,
			Code:       "ROUTE_INVALID_ID_FORMAT",
			Message:    "Route ID contains potentially problematic characters",
			EntityType: "Route",
			EntityID:   route.RouteID,
			Field:      "route_id",
			Value:      route.RouteID,
		})
	}

	// Validate route type
	validRouteTypes := map[int]string{
		0: "Tram", 1: "Subway", 2: "Rail", 3: "Bus", 4: "Ferry",
		5: "Cable tram", 6: "Aerial lift", 7: "Funicular", 11: "Trolleybus", 12: "Monorail",
	}
	if _, valid := validRouteTypes[route.RouteType]; !valid {
		v.AddIssue(ValidationIssue{
			Severity:   SeverityError,
			Code:       "ROUTE_INVALID_TYPE",
			Message:    fmt.Sprintf("Invalid route type: %d", route.RouteType),
			EntityType: "Route",
			EntityID:   route.RouteID,
			Field:      "route_type",
			Value:      fmt.Sprintf("%d", route.RouteType),
			Suggestion: "Use valid GTFS route types (0-12)",
		})
	}

	// Validate route name requirements
	if route.RouteShortName == "" && route.RouteLongName == "" {
		v.AddIssue(ValidationIssue{
			Severity:   SeverityError,
			Code:       "ROUTE_MISSING_NAME",
			Message:    "Either route_short_name or route_long_name must be specified",
			EntityType: "Route",
			EntityID:   route.RouteID,
			Suggestion: "Provide at least one name field",
		})
	}

	// Validate route colors
	if route.RouteColor != "" && !v.patterns.GTFSColor.MatchString(route.RouteColor) {
		v.AddIssue(ValidationIssue{
			Severity:   SeverityWarning,
			Code:       "ROUTE_INVALID_COLOR",
			Message:    "Route color should be a 6-character hex color",
			EntityType: "Route",
			EntityID:   route.RouteID,
			Field:      "route_color",
			Value:      route.RouteColor,
			Suggestion: "Use format RRGGBB (e.g., 'FF0000' for red)",
		})
	}

	if route.RouteTextColor != "" && !v.patterns.GTFSColor.MatchString(route.RouteTextColor) {
		v.AddIssue(ValidationIssue{
			Severity:   SeverityWarning,
			Code:       "ROUTE_INVALID_TEXT_COLOR",
			Message:    "Route text color should be a 6-character hex color",
			EntityType: "Route",
			EntityID:   route.RouteID,
			Field:      "route_text_color",
			Value:      route.RouteTextColor,
			Suggestion: "Use format RRGGBB (e.g., 'FFFFFF' for white)",
		})
	}
}

// ValidateGTFSStop validates a GTFS stop entity
func (v *Validator) ValidateGTFSStop(stop *model.Stop) {
	if stop == nil {
		v.AddIssue(ValidationIssue{
			Severity:   SeverityCritical,
			Code:       "STOP_NULL",
			Message:    "Stop entity is null",
			EntityType: "Stop",
		})
		return
	}

	// Validate required fields
	if stop.StopID == "" {
		v.AddIssue(ValidationIssue{
			Severity:   SeverityError,
			Code:       "STOP_MISSING_ID",
			Message:    "Stop ID is required but missing",
			EntityType: "Stop",
			Field:      "stop_id",
		})
	}

	if stop.StopName == "" {
		v.AddIssue(ValidationIssue{
			Severity:   SeverityError,
			Code:       "STOP_MISSING_NAME",
			Message:    "Stop name is required but missing",
			EntityType: "Stop",
			EntityID:   stop.StopID,
			Field:      "stop_name",
		})
	}

	// Validate coordinates
	if v.config.ValidateGeometry {
		if stop.StopLat < -90 || stop.StopLat > 90 {
			v.AddIssue(ValidationIssue{
				Severity:   SeverityError,
				Code:       "STOP_INVALID_LATITUDE",
				Message:    "Stop latitude must be between -90 and 90",
				EntityType: "Stop",
				EntityID:   stop.StopID,
				Field:      "stop_lat",
				Value:      fmt.Sprintf("%.6f", stop.StopLat),
			})
		}

		if stop.StopLon < -180 || stop.StopLon > 180 {
			v.AddIssue(ValidationIssue{
				Severity:   SeverityError,
				Code:       "STOP_INVALID_LONGITUDE",
				Message:    "Stop longitude must be between -180 and 180",
				EntityType: "Stop",
				EntityID:   stop.StopID,
				Field:      "stop_lon",
				Value:      fmt.Sprintf("%.6f", stop.StopLon),
			})
		}

		// Check for potentially incorrect coordinates (e.g., 0,0 or very imprecise)
		if stop.StopLat == 0.0 && stop.StopLon == 0.0 {
			v.AddIssue(ValidationIssue{
				Severity:   SeverityWarning,
				Code:       "STOP_SUSPICIOUS_COORDINATES",
				Message:    "Stop coordinates are at 0,0 which may indicate missing location data",
				EntityType: "Stop",
				EntityID:   stop.StopID,
				Field:      "coordinates",
			})
		}
	}

	// Validate accessibility information
	if v.config.ValidateAccessibility {
		if stop.WheelchairBoarding != "" {
			wheelchairValues := map[string]bool{"0": true, "1": true, "2": true}
			if !wheelchairValues[stop.WheelchairBoarding] {
				v.AddIssue(ValidationIssue{
					Severity:   SeverityWarning,
					Code:       "STOP_INVALID_WHEELCHAIR_BOARDING",
					Message:    "Invalid wheelchair boarding value",
					EntityType: "Stop",
					EntityID:   stop.StopID,
					Field:      "wheelchair_boarding",
					Value:      stop.WheelchairBoarding,
					Suggestion: "Use 0 (unknown), 1 (accessible), or 2 (not accessible)",
				})
			}
		}
	}
}

// ValidateGTFSStopTime validates a GTFS stop time entity
func (v *Validator) ValidateGTFSStopTime(stopTime *model.StopTime) {
	if stopTime == nil {
		v.AddIssue(ValidationIssue{
			Severity:   SeverityCritical,
			Code:       "STOPTIME_NULL",
			Message:    "StopTime entity is null",
			EntityType: "StopTime",
		})
		return
	}

	// Validate required fields
	if stopTime.TripID == "" {
		v.AddIssue(ValidationIssue{
			Severity:   SeverityError,
			Code:       "STOPTIME_MISSING_TRIP_ID",
			Message:    "Trip ID is required but missing",
			EntityType: "StopTime",
			Field:      "trip_id",
		})
	}

	if stopTime.StopID == "" {
		v.AddIssue(ValidationIssue{
			Severity:   SeverityError,
			Code:       "STOPTIME_MISSING_STOP_ID",
			Message:    "Stop ID is required but missing",
			EntityType: "StopTime",
			EntityID:   fmt.Sprintf("%s:%d", stopTime.TripID, stopTime.StopSequence),
			Field:      "stop_id",
		})
	}

	if stopTime.StopSequence <= 0 {
		v.AddIssue(ValidationIssue{
			Severity:   SeverityError,
			Code:       "STOPTIME_INVALID_SEQUENCE",
			Message:    "Stop sequence must be a positive integer",
			EntityType: "StopTime",
			EntityID:   fmt.Sprintf("%s:%d", stopTime.TripID, stopTime.StopSequence),
			Field:      "stop_sequence",
			Value:      fmt.Sprintf("%d", stopTime.StopSequence),
		})
	}

	// Validate time formats
	if v.config.ValidateTiming {
		if stopTime.ArrivalTime != "" && !v.patterns.GTFSTime.MatchString(stopTime.ArrivalTime) {
			v.AddIssue(ValidationIssue{
				Severity:   SeverityError,
				Code:       "STOPTIME_INVALID_ARRIVAL_TIME",
				Message:    "Arrival time format is invalid",
				EntityType: "StopTime",
				EntityID:   fmt.Sprintf("%s:%d", stopTime.TripID, stopTime.StopSequence),
				Field:      "arrival_time",
				Value:      stopTime.ArrivalTime,
				Suggestion: "Use HH:MM:SS format (e.g., '14:30:00')",
			})
		}

		if stopTime.DepartureTime != "" && !v.patterns.GTFSTime.MatchString(stopTime.DepartureTime) {
			v.AddIssue(ValidationIssue{
				Severity:   SeverityError,
				Code:       "STOPTIME_INVALID_DEPARTURE_TIME",
				Message:    "Departure time format is invalid",
				EntityType: "StopTime",
				EntityID:   fmt.Sprintf("%s:%d", stopTime.TripID, stopTime.StopSequence),
				Field:      "departure_time",
				Value:      stopTime.DepartureTime,
				Suggestion: "Use HH:MM:SS format (e.g., '14:30:00')",
			})
		}

		// Validate time logic: departure >= arrival
		if stopTime.ArrivalTime != "" && stopTime.DepartureTime != "" {
			arrivalMinutes := v.parseTimeToMinutes(stopTime.ArrivalTime)
			departureMinutes := v.parseTimeToMinutes(stopTime.DepartureTime)

			if arrivalMinutes > departureMinutes {
				v.AddIssue(ValidationIssue{
					Severity:   SeverityError,
					Code:       "STOPTIME_DEPARTURE_BEFORE_ARRIVAL",
					Message:    "Departure time cannot be before arrival time",
					EntityType: "StopTime",
					EntityID:   fmt.Sprintf("%s:%d", stopTime.TripID, stopTime.StopSequence),
					Context: map[string]string{
						"arrival_time":   stopTime.ArrivalTime,
						"departure_time": stopTime.DepartureTime,
					},
				})
			}
		}
	}

	// Validate pickup and drop-off types
	pickupDropoffValues := map[string]bool{"0": true, "1": true, "2": true, "3": true}
	if stopTime.PickupType != "" && !pickupDropoffValues[stopTime.PickupType] {
		v.AddIssue(ValidationIssue{
			Severity:   SeverityWarning,
			Code:       "STOPTIME_INVALID_PICKUP_TYPE",
			Message:    "Invalid pickup type value",
			EntityType: "StopTime",
			EntityID:   fmt.Sprintf("%s:%d", stopTime.TripID, stopTime.StopSequence),
			Field:      "pickup_type",
			Value:      stopTime.PickupType,
			Suggestion: "Use 0-3 (regular, none, phone, coordinate with driver)",
		})
	}

	if stopTime.DropOffType != "" && !pickupDropoffValues[stopTime.DropOffType] {
		v.AddIssue(ValidationIssue{
			Severity:   SeverityWarning,
			Code:       "STOPTIME_INVALID_DROPOFF_TYPE",
			Message:    "Invalid drop-off type value",
			EntityType: "StopTime",
			EntityID:   fmt.Sprintf("%s:%d", stopTime.TripID, stopTime.StopSequence),
			Field:      "drop_off_type",
			Value:      stopTime.DropOffType,
			Suggestion: "Use 0-3 (regular, none, phone, coordinate with driver)",
		})
	}
}

// parseTimeToMinutes converts GTFS time string to minutes since midnight
func (v *Validator) parseTimeToMinutes(timeStr string) int {
	parts := strings.Split(timeStr, ":")
	if len(parts) != 3 {
		return -1
	}

	hours := 0
	minutes := 0
	seconds := 0

	if h, err := fmt.Sscanf(parts[0], "%d", &hours); err != nil || h != 1 {
		return -1
	}
	if m, err := fmt.Sscanf(parts[1], "%d", &minutes); err != nil || m != 1 {
		return -1
	}
	if s, err := fmt.Sscanf(parts[2], "%d", &seconds); err != nil || s != 1 {
		return -1
	}

	return hours*60 + minutes
}

// UpdateProcessingStats updates processing statistics
func (v *Validator) UpdateProcessingStats(entityType string, processed, converted, skipped int) {
	v.processingStats.EntitiesProcessed[entityType] = processed
	v.processingStats.EntitiesConverted[entityType] = converted
	v.processingStats.EntitiesSkipped[entityType] = skipped
}

// GetReport generates a comprehensive validation report
func (v *Validator) GetReport() ValidationReport {
	summary := v.generateSummary()

	// Calculate conversion rate
	totalProcessed := 0
	totalConverted := 0
	for _, count := range v.processingStats.EntitiesProcessed {
		totalProcessed += count
	}
	for _, count := range v.processingStats.EntitiesConverted {
		totalConverted += count
	}

	if totalProcessed > 0 {
		v.processingStats.ConversionRate = float64(totalConverted) / float64(totalProcessed) * 100
	}

	return ValidationReport{
		Issues:          v.issues,
		Summary:         summary,
		Timestamp:       time.Now(),
		ProcessingStats: v.processingStats,
	}
}

// generateSummary creates a validation summary
func (v *Validator) generateSummary() ValidationSummary {
	summary := ValidationSummary{
		TotalIssues:  len(v.issues),
		BySeverity:   make(map[ValidationSeverity]int),
		ByCategory:   make(map[string]int),
		ByEntityType: make(map[string]int),
		IsValid:      true,
		HasCritical:  false,
		HasErrors:    false,
	}

	for _, issue := range v.issues {
		summary.BySeverity[issue.Severity]++
		summary.ByEntityType[issue.EntityType]++

		// Extract category from issue code
		parts := strings.Split(issue.Code, "_")
		if len(parts) > 0 {
			summary.ByCategory[parts[0]]++
		}

		// Update validation status
		if issue.Severity == SeverityCritical {
			summary.HasCritical = true
			summary.IsValid = false
		}
		if issue.Severity == SeverityError {
			summary.HasErrors = true
			summary.IsValid = false
		}
	}

	return summary
}

// Reset clears all validation issues and statistics
func (v *Validator) Reset() {
	v.issues = make([]ValidationIssue, 0)
	v.processingStats = ProcessingStats{
		EntitiesProcessed: make(map[string]int),
		EntitiesConverted: make(map[string]int),
		EntitiesSkipped:   make(map[string]int),
	}
}
