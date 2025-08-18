package calendar

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/theoremus-urban-solutions/netex-gtfs-converter/model"
)

// NeTExCalendarProcessor processes NeTEx calendar data and converts it to service patterns
type NeTExCalendarProcessor struct {
	calendarManager *CalendarManager
	timezone        *time.Location
}

// NewNeTExCalendarProcessor creates a new NeTEx calendar processor
func NewNeTExCalendarProcessor(calendarManager *CalendarManager, timezoneName string) (*NeTExCalendarProcessor, error) {
	timezone, err := time.LoadLocation(timezoneName)
	if err != nil {
		timezone = time.UTC
	}

	return &NeTExCalendarProcessor{
		calendarManager: calendarManager,
		timezone:        timezone,
	}, nil
}

// ProcessServiceFrame processes a NeTEx ServiceFrame to extract calendar information
func (ncp *NeTExCalendarProcessor) ProcessServiceFrame(serviceFrame interface{}) error {
	// This would process the actual NeTEx ServiceFrame structure
	// For now, we'll create example service patterns

	// Create example patterns for different service types
	if _, err := ncp.createRegularServicePattern(); err != nil {
		return fmt.Errorf("failed to create regular service pattern: %v", err)
	}

	if _, err := ncp.createWeekendServicePattern(); err != nil {
		return fmt.Errorf("failed to create weekend service pattern: %v", err)
	}

	if _, err := ncp.createHolidayServicePattern(); err != nil {
		return fmt.Errorf("failed to create holiday service pattern: %v", err)
	}

	if _, err := ncp.createSeasonalServicePattern(); err != nil {
		return fmt.Errorf("failed to create seasonal service pattern: %v", err)
	}

	return nil
}

// ProcessDayType processes NeTEx DayType elements
func (ncp *NeTExCalendarProcessor) ProcessDayType(dayType interface{}) (*ServicePattern, error) {
	// Extract day type information from NeTEx structure
	// This is a simplified example

	pattern := &ServicePattern{
		ID:   "day_type_example",
		Name: "Example Day Type Pattern",
		Type: PatternRegular,
		ValidityPeriod: &ValidityPeriod{
			StartDate: time.Now(),
			EndDate:   time.Now().AddDate(1, 0, 0),
		},
		OperatingDays: []time.Weekday{
			time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday,
		},
		Exceptions:         make([]*ServiceException, 0),
		SeasonalVariations: make([]*SeasonalVariation, 0),
		SpecialDays:        make(map[string]*SpecialDay),
		HolidayBehavior:    HolidayAsWeekday,
	}

	return pattern, nil
}

// ProcessOperatingPeriod processes NeTEx OperatingPeriod elements
func (ncp *NeTExCalendarProcessor) ProcessOperatingPeriod(operatingPeriod interface{}) (*OperatingPeriod, error) {
	// Extract operating period information from NeTEx structure
	// This is a simplified example

	basePattern, err := ncp.createRegularServicePattern()
	if err != nil {
		return nil, fmt.Errorf("failed to create base pattern: %v", err)
	}

	period := &OperatingPeriod{
		ID:          "operating_period_example",
		Name:        "Example Operating Period",
		StartDate:   time.Now(),
		EndDate:     time.Now().AddDate(0, 6, 0), // 6 months
		BasePattern: basePattern,
		Overrides:   make(map[string]*ServicePattern),
		Priority:    1,
	}

	return period, nil
}

// ProcessServiceCalendar processes NeTEx ServiceCalendar elements
func (ncp *NeTExCalendarProcessor) ProcessServiceCalendar(serviceCalendar interface{}) error {
	// Process service calendar from NeTEx
	// This would extract calendar information and create appropriate patterns

	return nil
}

// ProcessDayTypeAssignment processes NeTEx DayTypeAssignment elements
func (ncp *NeTExCalendarProcessor) ProcessDayTypeAssignment(assignment interface{}) error {
	// Process day type assignments that link day types to specific dates
	// This creates exceptions in service patterns

	return nil
}

// ProcessUicOperatingPeriod processes UIC operating periods common in European rail
func (ncp *NeTExCalendarProcessor) ProcessUicOperatingPeriod(uicPeriod interface{}) (*OperatingPeriod, error) {
	// UIC (International Union of Railways) operating periods
	// These are standardized across European rail networks

	period := &OperatingPeriod{
		ID:        "uic_period_example",
		Name:      "UIC Operating Period",
		StartDate: time.Now(),
		EndDate:   time.Now().AddDate(0, 3, 0), // 3 months
		Priority:  2,
	}

	return period, nil
}

// Example service pattern creation methods

func (ncp *NeTExCalendarProcessor) createRegularServicePattern() (*ServicePattern, error) {
	pattern := &ServicePattern{
		ID:   "regular_weekday",
		Name: "Regular Weekday Service",
		Type: PatternRegular,
		ValidityPeriod: &ValidityPeriod{
			StartDate: time.Now(),
			EndDate:   time.Now().AddDate(1, 0, 0),
		},
		OperatingDays: []time.Weekday{
			time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday,
		},
		Exceptions:         make([]*ServiceException, 0),
		SeasonalVariations: make([]*SeasonalVariation, 0),
		SpecialDays:        make(map[string]*SpecialDay),
		HolidayBehavior:    HolidayAsWeekend,
	}

	// Add common holiday exceptions
	ncp.addCommonHolidayExceptions(pattern)

	ncp.calendarManager.AddServicePattern(pattern)
	return pattern, nil
}

func (ncp *NeTExCalendarProcessor) createWeekendServicePattern() (*ServicePattern, error) {
	pattern := &ServicePattern{
		ID:   "weekend_service",
		Name: "Weekend Service",
		Type: PatternWeekend,
		ValidityPeriod: &ValidityPeriod{
			StartDate: time.Now(),
			EndDate:   time.Now().AddDate(1, 0, 0),
		},
		OperatingDays: []time.Weekday{
			time.Saturday, time.Sunday,
		},
		Exceptions:         make([]*ServiceException, 0),
		SeasonalVariations: make([]*SeasonalVariation, 0),
		SpecialDays:        make(map[string]*SpecialDay),
		HolidayBehavior:    HolidayAsWeekday,
	}

	ncp.calendarManager.AddServicePattern(pattern)
	return pattern, nil
}

func (ncp *NeTExCalendarProcessor) createHolidayServicePattern() (*ServicePattern, error) {
	pattern := &ServicePattern{
		ID:   "holiday_service",
		Name: "Holiday Service",
		Type: PatternHoliday,
		ValidityPeriod: &ValidityPeriod{
			StartDate: time.Now(),
			EndDate:   time.Now().AddDate(1, 0, 0),
		},
		OperatingDays:      make([]time.Weekday, 0), // No regular operating days
		Exceptions:         make([]*ServiceException, 0),
		SeasonalVariations: make([]*SeasonalVariation, 0),
		SpecialDays:        make(map[string]*SpecialDay),
		HolidayBehavior:    HolidaySpecialSchedule,
	}

	// Add specific holiday dates as exceptions
	ncp.addHolidayExceptions(pattern)

	ncp.calendarManager.AddServicePattern(pattern)
	return pattern, nil
}

func (ncp *NeTExCalendarProcessor) createSeasonalServicePattern() (*ServicePattern, error) {
	pattern := &ServicePattern{
		ID:   "seasonal_service",
		Name: "Seasonal Service",
		Type: PatternSeasonal,
		ValidityPeriod: &ValidityPeriod{
			StartDate: time.Date(time.Now().Year(), time.June, 1, 0, 0, 0, 0, ncp.timezone),
			EndDate:   time.Date(time.Now().Year(), time.August, 31, 0, 0, 0, 0, ncp.timezone),
		},
		OperatingDays: []time.Weekday{
			time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday,
			time.Saturday, time.Sunday,
		},
		Exceptions:         make([]*ServiceException, 0),
		SeasonalVariations: make([]*SeasonalVariation, 0),
		SpecialDays:        make(map[string]*SpecialDay),
		HolidayBehavior:    HolidayAsWeekday,
	}

	// Add seasonal variations
	summerVariation := &SeasonalVariation{
		Season:    SeasonSummer,
		StartDate: time.Date(time.Now().Year(), time.June, 21, 0, 0, 0, 0, ncp.timezone),
		EndDate:   time.Date(time.Now().Year(), time.September, 21, 0, 0, 0, 0, ncp.timezone),
		Changes: []ServiceChange{
			{
				Type:        ChangeFrequency,
				Description: "Increased frequency during summer",
				Parameters: map[string]interface{}{
					"frequency_multiplier": 1.5,
				},
			},
		},
	}
	pattern.SeasonalVariations = append(pattern.SeasonalVariations, summerVariation)

	ncp.calendarManager.AddServicePattern(pattern)
	return pattern, nil
}

// Helper methods for adding exceptions

func (ncp *NeTExCalendarProcessor) addCommonHolidayExceptions(pattern *ServicePattern) {
	currentYear := time.Now().Year()

	// New Year's Day
	newYear := time.Date(currentYear, time.January, 1, 0, 0, 0, 0, ncp.timezone)
	pattern.Exceptions = append(pattern.Exceptions, &ServiceException{
		Date:   newYear,
		Type:   ExceptionRemoved,
		Reason: "New Year's Day",
	})

	// Christmas Day
	christmas := time.Date(currentYear, time.December, 25, 0, 0, 0, 0, ncp.timezone)
	pattern.Exceptions = append(pattern.Exceptions, &ServiceException{
		Date:   christmas,
		Type:   ExceptionRemoved,
		Reason: "Christmas Day",
	})

	// Boxing Day
	boxingDay := time.Date(currentYear, time.December, 26, 0, 0, 0, 0, ncp.timezone)
	pattern.Exceptions = append(pattern.Exceptions, &ServiceException{
		Date:   boxingDay,
		Type:   ExceptionRemoved,
		Reason: "Boxing Day",
	})
}

func (ncp *NeTExCalendarProcessor) addHolidayExceptions(pattern *ServicePattern) {
	// Get holidays from holiday detector
	currentYear := time.Now().Year()
	holidays, err := ncp.calendarManager.holidayDetector.GetHolidays(currentYear)
	if err != nil {
		return // Fail silently for now
	}

	for _, holiday := range holidays {
		if holiday.IsNational {
			pattern.Exceptions = append(pattern.Exceptions, &ServiceException{
				Date:   holiday.Date,
				Type:   ExceptionAdded,
				Reason: fmt.Sprintf("Holiday: %s", holiday.Name),
			})
		}
	}
}

// ConvertToGTFSCalendar converts processed service patterns to GTFS calendar format
func (ncp *NeTExCalendarProcessor) ConvertToGTFSCalendar() ([]*model.Calendar, []*model.CalendarDate, error) {
	return ncp.calendarManager.GenerateGTFSCalendar(nil)
}

// ParseNeTExDate parses a NeTEx date string to time.Time
func (ncp *NeTExCalendarProcessor) ParseNeTExDate(dateStr string) (time.Time, error) {
	// NeTEx dates are typically in ISO 8601 format
	layouts := []string{
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04:05",
		"2006-01-02",
		"20060102",
	}

	for _, layout := range layouts {
		if date, err := time.ParseInLocation(layout, dateStr, ncp.timezone); err == nil {
			return date, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}

// ParseNeTExTime parses a NeTEx time string to duration from midnight
func (ncp *NeTExCalendarProcessor) ParseNeTExTime(timeStr string) (time.Duration, error) {
	// NeTEx times can be in various formats
	if strings.Contains(timeStr, ":") {
		// HH:MM:SS format
		parts := strings.Split(timeStr, ":")
		if len(parts) < 2 {
			return 0, fmt.Errorf("invalid time format: %s", timeStr)
		}

		hours, err := strconv.Atoi(parts[0])
		if err != nil {
			return 0, fmt.Errorf("invalid hours: %s", parts[0])
		}

		minutes, err := strconv.Atoi(parts[1])
		if err != nil {
			return 0, fmt.Errorf("invalid minutes: %s", parts[1])
		}

		seconds := 0
		if len(parts) > 2 {
			seconds, err = strconv.Atoi(parts[2])
			if err != nil {
				return 0, fmt.Errorf("invalid seconds: %s", parts[2])
			}
		}

		return time.Duration(hours)*time.Hour + time.Duration(minutes)*time.Minute + time.Duration(seconds)*time.Second, nil
	}

	// ISO 8601 duration format (PT1H30M)
	if strings.HasPrefix(timeStr, "PT") {
		return ncp.parseISO8601Duration(timeStr)
	}

	return 0, fmt.Errorf("unsupported time format: %s", timeStr)
}

// parseISO8601Duration parses an ISO 8601 duration string
func (ncp *NeTExCalendarProcessor) parseISO8601Duration(durationStr string) (time.Duration, error) {
	// Remove PT prefix
	durationStr = strings.TrimPrefix(durationStr, "PT")

	var duration time.Duration
	var current strings.Builder

	for _, char := range durationStr {
		if char >= '0' && char <= '9' || char == '.' {
			current.WriteRune(char)
		} else {
			valueStr := current.String()
			if valueStr == "" {
				continue
			}

			value, err := strconv.ParseFloat(valueStr, 64)
			if err != nil {
				return 0, fmt.Errorf("invalid duration value: %s", valueStr)
			}

			switch char {
			case 'H':
				duration += time.Duration(value * float64(time.Hour))
			case 'M':
				duration += time.Duration(value * float64(time.Minute))
			case 'S':
				duration += time.Duration(value * float64(time.Second))
			default:
				return 0, fmt.Errorf("unsupported duration unit: %c", char)
			}

			current.Reset()
		}
	}

	return duration, nil
}

// ExtractServicePatternFromNeTEx extracts service pattern information from NeTEx elements
func (ncp *NeTExCalendarProcessor) ExtractServicePatternFromNeTEx(netexElement interface{}) (*ServicePattern, error) {
	// This would parse the actual NeTEx XML structure
	// For now, return a placeholder pattern

	pattern := &ServicePattern{
		ID:   "extracted_pattern",
		Name: "Pattern from NeTEx",
		Type: PatternRegular,
		ValidityPeriod: &ValidityPeriod{
			StartDate: time.Now(),
			EndDate:   time.Now().AddDate(1, 0, 0),
		},
		OperatingDays: []time.Weekday{
			time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday,
		},
		Exceptions:         make([]*ServiceException, 0),
		SeasonalVariations: make([]*SeasonalVariation, 0),
		SpecialDays:        make(map[string]*SpecialDay),
		HolidayBehavior:    HolidayAsWeekday,
	}

	return pattern, nil
}

// ValidateNeTExCalendarData validates NeTEx calendar data for consistency
func (ncp *NeTExCalendarProcessor) ValidateNeTExCalendarData() []string {
	issues := make([]string, 0)

	// Validate all service patterns
	for patternID, pattern := range ncp.calendarManager.servicePatterns {
		patternIssues := ncp.calendarManager.ValidateServicePattern(pattern)
		for _, issue := range patternIssues {
			issues = append(issues, fmt.Sprintf("Pattern %s: %s", patternID, issue))
		}
	}

	// Validate operating periods
	for periodID, period := range ncp.calendarManager.operatingPeriods {
		if period.StartDate.After(period.EndDate) {
			issues = append(issues, fmt.Sprintf("Operating period %s: start date is after end date", periodID))
		}

		if period.BasePattern == nil && len(period.Overrides) == 0 {
			issues = append(issues, fmt.Sprintf("Operating period %s: no base pattern or overrides defined", periodID))
		}
	}

	return issues
}
