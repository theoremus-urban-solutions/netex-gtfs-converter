package calendar

import (
	"fmt"
	"time"

	"github.com/theoremus-urban-solutions/netex-gtfs-converter/model"
)

// CalendarService provides high-level calendar conversion services
type CalendarService struct {
	manager   *CalendarManager
	processor *NeTExCalendarProcessor
	validator *CalendarValidator
	generator *GTFSCalendarGenerator
	config    CalendarServiceConfig
}

// CalendarServiceConfig controls the calendar service behavior
type CalendarServiceConfig struct {
	DefaultTimezoneName        string          `json:"default_timezone_name"`
	HolidayCountryCode         string          `json:"holiday_country_code"`
	EnableHolidayDetection     bool            `json:"enable_holiday_detection"`
	EnableSeasonalPatterns     bool            `json:"enable_seasonal_patterns"`
	EnableSchoolCalendar       bool            `json:"enable_school_calendar"`
	EnableWeekendAdjustments   bool            `json:"enable_weekend_adjustments"`
	MaxServiceExceptions       int             `json:"max_service_exceptions"`
	ValidationLevel            ValidationLevel `json:"validation_level"`
	OptimizeCalendarDates      bool            `json:"optimize_calendar_dates"`
	ConsolidateSimilarPatterns bool            `json:"consolidate_similar_patterns"`
}

// ValidationLevel defines the level of calendar validation
type ValidationLevel int

const (
	ValidationMinimal ValidationLevel = iota
	ValidationStandard
	ValidationStrict
	ValidationDetailed
)

func (v ValidationLevel) String() string {
	switch v {
	case ValidationMinimal:
		return "Minimal"
	case ValidationStandard:
		return "Standard"
	case ValidationStrict:
		return "Strict"
	case ValidationDetailed:
		return "Detailed"
	default:
		return unknownPatternType
	}
}

// ConversionResult holds the results of calendar conversion
type ConversionResult struct {
	Calendars          []*model.Calendar     `json:"calendars"`
	CalendarDates      []*model.CalendarDate `json:"calendar_dates"`
	ServicePatterns    []*ServicePattern     `json:"service_patterns"`
	OperatingPeriods   []*OperatingPeriod    `json:"operating_periods"`
	ConversionStats    ConversionStats       `json:"conversion_stats"`
	ValidationIssues   []ValidationIssue     `json:"validation_issues"`
	ProcessingDuration time.Duration         `json:"processing_duration"`
}

// ConversionStats tracks statistics during calendar conversion
type ConversionStats struct {
	TotalServicePatterns  int                      `json:"total_service_patterns"`
	TotalOperatingPeriods int                      `json:"total_operating_periods"`
	TotalCalendars        int                      `json:"total_calendars"`
	TotalCalendarDates    int                      `json:"total_calendar_dates"`
	TotalExceptions       int                      `json:"total_exceptions"`
	HolidaysDetected      int                      `json:"holidays_detected"`
	SeasonalVariations    int                      `json:"seasonal_variations"`
	PatternsByType        map[string]int           `json:"patterns_by_type"`
	ProcessingTimeByStage map[string]time.Duration `json:"processing_time_by_stage"`
}

// ValidationIssue represents an issue found during calendar validation
type ValidationIssue struct {
	Level      ValidationLevel   `json:"level"`
	Code       string            `json:"code"`
	Message    string            `json:"message"`
	Entity     string            `json:"entity"`
	EntityID   string            `json:"entity_id"`
	Suggestion string            `json:"suggestion,omitempty"`
	Context    map[string]string `json:"context,omitempty"`
}

// NewCalendarService creates a new calendar service
func NewCalendarService(config CalendarServiceConfig) (*CalendarService, error) {
	// Set defaults
	if config.DefaultTimezoneName == "" {
		config.DefaultTimezoneName = "Europe/Oslo"
	}
	if config.HolidayCountryCode == "" {
		config.HolidayCountryCode = "NO"
	}
	if config.MaxServiceExceptions == 0 {
		config.MaxServiceExceptions = 500
	}
	if config.ValidationLevel == 0 {
		config.ValidationLevel = ValidationStandard
	}

	// Create calendar manager
	managerConfig := CalendarConfig{
		EnableHolidayDetection:   config.EnableHolidayDetection,
		EnableSeasonalPatterns:   config.EnableSeasonalPatterns,
		EnableSchoolCalendar:     config.EnableSchoolCalendar,
		EnableWeekendAdjustments: config.EnableWeekendAdjustments,
		MaxServiceExceptions:     config.MaxServiceExceptions,
		HolidayCountryCode:       config.HolidayCountryCode,
		TimezoneName:             config.DefaultTimezoneName,
	}
	manager := NewCalendarManager(managerConfig)

	// Create NeTEx processor
	processor, err := NewNeTExCalendarProcessor(manager, config.DefaultTimezoneName)
	if err != nil {
		return nil, fmt.Errorf("failed to create NeTEx processor: %v", err)
	}

	// Create validator
	validator := NewCalendarValidator(config.ValidationLevel)

	// Create GTFS generator
	generator := NewGTFSCalendarGenerator(manager)

	return &CalendarService{
		manager:   manager,
		processor: processor,
		validator: validator,
		generator: generator,
		config:    config,
	}, nil
}

// ConvertNeTExToGTFS converts NeTEx calendar data to GTFS format
func (cs *CalendarService) ConvertNeTExToGTFS(netexData interface{}) (*ConversionResult, error) {
	startTime := time.Now()

	result := &ConversionResult{
		ConversionStats: ConversionStats{
			PatternsByType:        make(map[string]int),
			ProcessingTimeByStage: make(map[string]time.Duration),
		},
		ValidationIssues: make([]ValidationIssue, 0),
	}

	// Stage 1: Process NeTEx calendar data
	stageStart := time.Now()
	if err := cs.processor.ProcessServiceFrame(netexData); err != nil {
		return nil, fmt.Errorf("failed to process NeTEx service frame: %v", err)
	}
	result.ConversionStats.ProcessingTimeByStage["netex_processing"] = time.Since(stageStart)

	// Stage 2: Validate service patterns
	stageStart = time.Now()
	if cs.config.ValidationLevel >= ValidationStandard {
		validationIssues := cs.validateAllServicePatterns()
		result.ValidationIssues = append(result.ValidationIssues, validationIssues...)
	}
	result.ConversionStats.ProcessingTimeByStage["validation"] = time.Since(stageStart)

	// Stage 3: Generate GTFS calendars
	stageStart = time.Now()
	calendars, calendarDates, err := cs.generator.GenerateGTFSCalendars()
	if err != nil {
		return nil, fmt.Errorf("failed to generate GTFS calendars: %v", err)
	}
	result.ConversionStats.ProcessingTimeByStage["gtfs_generation"] = time.Since(stageStart)

	// Stage 4: Optimize and consolidate if enabled
	stageStart = time.Now()
	if cs.config.OptimizeCalendarDates {
		calendarDates = cs.optimizeCalendarDates(calendarDates)
	}
	if cs.config.ConsolidateSimilarPatterns {
		calendars = cs.consolidateSimilarCalendars(calendars)
	}
	result.ConversionStats.ProcessingTimeByStage["optimization"] = time.Since(stageStart)

	// Populate results
	result.Calendars = calendars
	result.CalendarDates = calendarDates
	result.ServicePatterns = cs.getServicePatterns()
	result.OperatingPeriods = cs.getOperatingPeriods()
	result.ProcessingDuration = time.Since(startTime)

	// Calculate statistics
	cs.calculateConversionStats(result)

	return result, nil
}

// ProcessNeTExServiceFrame processes a NeTEx ServiceFrame
func (cs *CalendarService) ProcessNeTExServiceFrame(serviceFrame interface{}) error {
	return cs.processor.ProcessServiceFrame(serviceFrame)
}

// AddCustomServicePattern adds a custom service pattern
func (cs *CalendarService) AddCustomServicePattern(pattern *ServicePattern) error {
	// Validate the pattern
	if cs.config.ValidationLevel >= ValidationStandard {
		issues := cs.manager.ValidateServicePattern(pattern)
		if len(issues) > 0 {
			return fmt.Errorf("service pattern validation failed: %v", issues)
		}
	}

	cs.manager.AddServicePattern(pattern)
	return nil
}

// AddCustomOperatingPeriod adds a custom operating period
func (cs *CalendarService) AddCustomOperatingPeriod(period *OperatingPeriod) error {
	// Validate the operating period
	if cs.config.ValidationLevel >= ValidationStandard {
		if period.StartDate.After(period.EndDate) {
			return fmt.Errorf("operating period start date cannot be after end date")
		}
		if period.BasePattern == nil && len(period.Overrides) == 0 {
			return fmt.Errorf("operating period must have either a base pattern or overrides")
		}
	}

	cs.manager.AddOperatingPeriod(period)
	return nil
}

// GetServiceDates returns all service dates for a given service pattern
func (cs *CalendarService) GetServiceDates(patternID string, startDate, endDate time.Time) ([]time.Time, error) {
	return cs.manager.GetEffectiveDates(patternID, startDate, endDate)
}

// IsServiceOperating checks if a service is operating on a specific date
func (cs *CalendarService) IsServiceOperating(patternID string, date time.Time) (bool, error) {
	pattern := cs.manager.GetServicePattern(patternID)
	if pattern == nil {
		return false, fmt.Errorf("service pattern %s not found", patternID)
	}

	return cs.manager.isServiceOperating(pattern, date), nil
}

// GetHolidays returns holidays for a specific year and country
func (cs *CalendarService) GetHolidays(year int) ([]*Holiday, error) {
	return cs.manager.holidayDetector.GetHolidays(year)
}

// ValidateConfiguration validates the calendar service configuration
func (cs *CalendarService) ValidateConfiguration() []ValidationIssue {
	issues := make([]ValidationIssue, 0)

	// Validate timezone
	if _, err := time.LoadLocation(cs.config.DefaultTimezoneName); err != nil {
		issues = append(issues, ValidationIssue{
			Level:   ValidationStandard,
			Code:    "INVALID_TIMEZONE",
			Message: fmt.Sprintf("Invalid timezone: %s", cs.config.DefaultTimezoneName),
			Entity:  "Configuration",
		})
	}

	// Validate country code
	if len(cs.config.HolidayCountryCode) != 2 {
		issues = append(issues, ValidationIssue{
			Level:   ValidationStandard,
			Code:    "INVALID_COUNTRY_CODE",
			Message: fmt.Sprintf("Invalid country code: %s", cs.config.HolidayCountryCode),
			Entity:  "Configuration",
		})
	}

	// Validate limits
	if cs.config.MaxServiceExceptions < 0 {
		issues = append(issues, ValidationIssue{
			Level:   ValidationStandard,
			Code:    "INVALID_MAX_EXCEPTIONS",
			Message: "Max service exceptions cannot be negative",
			Entity:  "Configuration",
		})
	}

	return issues
}

// Private helper methods

func (cs *CalendarService) validateAllServicePatterns() []ValidationIssue {
	issues := make([]ValidationIssue, 0)

	for patternID, pattern := range cs.manager.servicePatterns {
		patternIssues := cs.validator.ValidateServicePattern(pattern)
		for _, issue := range patternIssues {
			validationIssue := ValidationIssue{
				Level:    cs.config.ValidationLevel,
				Code:     issue,
				Message:  issue,
				Entity:   "ServicePattern",
				EntityID: patternID,
			}
			issues = append(issues, validationIssue)
		}
	}

	return issues
}

func (cs *CalendarService) optimizeCalendarDates(calendarDates []*model.CalendarDate) []*model.CalendarDate {
	// Remove redundant calendar dates and optimize storage
	// This is a simplified optimization - could be more sophisticated

	uniqueDates := make(map[string]*model.CalendarDate)

	for _, date := range calendarDates {
		key := fmt.Sprintf("%s_%s", date.ServiceID, date.Date)
		if existing, exists := uniqueDates[key]; exists {
			// Prioritize removal over addition
			if date.ExceptionType == 2 && existing.ExceptionType == 1 {
				uniqueDates[key] = date
			}
		} else {
			uniqueDates[key] = date
		}
	}

	result := make([]*model.CalendarDate, 0, len(uniqueDates))
	for _, date := range uniqueDates {
		result = append(result, date)
	}

	return result
}

func (cs *CalendarService) consolidateSimilarCalendars(calendars []*model.Calendar) []*model.Calendar {
	// Consolidate calendars with identical operating patterns
	// This is a simplified consolidation - could be more sophisticated

	patternMap := make(map[string]*model.Calendar)

	for _, calendar := range calendars {
		pattern := fmt.Sprintf("%t_%t_%t_%t_%t_%t_%t_%s_%s",
			calendar.Monday, calendar.Tuesday, calendar.Wednesday, calendar.Thursday,
			calendar.Friday, calendar.Saturday, calendar.Sunday,
			calendar.StartDate, calendar.EndDate)

		if _, exists := patternMap[pattern]; !exists {
			patternMap[pattern] = calendar
		}
		// If pattern already exists, we could merge the service IDs or handle differently
	}

	result := make([]*model.Calendar, 0, len(patternMap))
	for _, calendar := range patternMap {
		result = append(result, calendar)
	}

	return result
}

func (cs *CalendarService) getServicePatterns() []*ServicePattern {
	patterns := make([]*ServicePattern, 0, len(cs.manager.servicePatterns))
	for _, pattern := range cs.manager.servicePatterns {
		patterns = append(patterns, pattern)
	}
	return patterns
}

func (cs *CalendarService) getOperatingPeriods() []*OperatingPeriod {
	periods := make([]*OperatingPeriod, 0, len(cs.manager.operatingPeriods))
	for _, period := range cs.manager.operatingPeriods {
		periods = append(periods, period)
	}
	return periods
}

func (cs *CalendarService) calculateConversionStats(result *ConversionResult) {
	stats := &result.ConversionStats

	stats.TotalServicePatterns = len(result.ServicePatterns)
	stats.TotalOperatingPeriods = len(result.OperatingPeriods)
	stats.TotalCalendars = len(result.Calendars)
	stats.TotalCalendarDates = len(result.CalendarDates)

	// Count exceptions
	for _, pattern := range result.ServicePatterns {
		stats.TotalExceptions += len(pattern.Exceptions)
		stats.SeasonalVariations += len(pattern.SeasonalVariations)

		// Count patterns by type
		typeStr := pattern.Type.String()
		stats.PatternsByType[typeStr]++
	}

	// Count holidays if holiday detection is enabled
	if cs.config.EnableHolidayDetection {
		currentYear := time.Now().Year()
		holidays, err := cs.manager.holidayDetector.GetHolidays(currentYear)
		if err == nil {
			stats.HolidaysDetected = len(holidays)
		}
	}
}

// GetConversionSummary returns a summary of the calendar conversion
func (cs *CalendarService) GetConversionSummary() map[string]interface{} {
	summary := make(map[string]interface{})

	summary["service_patterns_count"] = len(cs.manager.servicePatterns)
	summary["operating_periods_count"] = len(cs.manager.operatingPeriods)
	summary["calendars_count"] = len(cs.manager.calendars)
	summary["calendar_dates_count"] = len(cs.manager.calendarDates)
	summary["holiday_detection_enabled"] = cs.config.EnableHolidayDetection
	summary["seasonal_patterns_enabled"] = cs.config.EnableSeasonalPatterns
	summary["validation_level"] = cs.config.ValidationLevel.String()

	return summary
}
