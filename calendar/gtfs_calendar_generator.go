package calendar

import (
	"fmt"
	"sort"
	"time"

	"github.com/theoremus-urban-solutions/netex-gtfs-converter/model"
)

// GTFSCalendarGenerator generates GTFS calendar and calendar_dates files
type GTFSCalendarGenerator struct {
	calendarManager *CalendarManager
	optimizations   GeneratorOptimizations
}

// GeneratorOptimizations controls calendar generation optimizations
type GeneratorOptimizations struct {
	MergeCompatibleCalendars     bool `json:"merge_compatible_calendars"`
	MinimizeCalendarDates        bool `json:"minimize_calendar_dates"`
	UseCalendarForRegularService bool `json:"use_calendar_for_regular_service"`
	MaxCalendarDatesPerService   int  `json:"max_calendar_dates_per_service"`
	PreferCalendarOverDates      bool `json:"prefer_calendar_over_dates"`
}

// NewGTFSCalendarGenerator creates a new GTFS calendar generator
func NewGTFSCalendarGenerator(manager *CalendarManager) *GTFSCalendarGenerator {
	return &GTFSCalendarGenerator{
		calendarManager: manager,
		optimizations: GeneratorOptimizations{
			MergeCompatibleCalendars:     true,
			MinimizeCalendarDates:        true,
			UseCalendarForRegularService: true,
			MaxCalendarDatesPerService:   100,
			PreferCalendarOverDates:      true,
		},
	}
}

// SetOptimizations updates the generator optimizations
func (gcg *GTFSCalendarGenerator) SetOptimizations(opts GeneratorOptimizations) {
	gcg.optimizations = opts
}

// GenerateGTFSCalendars generates GTFS calendar and calendar_dates from service patterns
func (gcg *GTFSCalendarGenerator) GenerateGTFSCalendars() ([]*model.Calendar, []*model.CalendarDate, error) {
	calendars := make([]*model.Calendar, 0)
	calendarDates := make([]*model.CalendarDate, 0)

	// Process each service pattern
	for patternID, pattern := range gcg.calendarManager.servicePatterns {
		calendar, dates, err := gcg.processServicePattern(patternID, pattern)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to process service pattern %s: %v", patternID, err)
		}

		if calendar != nil {
			calendars = append(calendars, calendar)
		}

		calendarDates = append(calendarDates, dates...)
	}

	// Process operating periods
	for periodID, period := range gcg.calendarManager.operatingPeriods {
		periodCalendars, periodDates, err := gcg.processOperatingPeriod(periodID, period)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to process operating period %s: %v", periodID, err)
		}

		calendars = append(calendars, periodCalendars...)
		calendarDates = append(calendarDates, periodDates...)
	}

	// Apply optimizations
	if gcg.optimizations.MergeCompatibleCalendars {
		calendars = gcg.mergeCompatibleCalendars(calendars)
	}

	if gcg.optimizations.MinimizeCalendarDates {
		calendarDates = gcg.minimizeCalendarDates(calendarDates, calendars)
	}

	// Sort results for consistency
	gcg.sortCalendars(calendars)
	gcg.sortCalendarDates(calendarDates)

	return calendars, calendarDates, nil
}

// processServicePattern converts a service pattern to GTFS calendar format
func (gcg *GTFSCalendarGenerator) processServicePattern(patternID string, pattern *ServicePattern) (*model.Calendar, []*model.CalendarDate, error) {
	var calendar *model.Calendar
	calendarDates := make([]*model.CalendarDate, 0)

	// Determine if we should use calendar.txt or calendar_dates.txt
	useCalendar := gcg.shouldUseCalendar(pattern)

	if useCalendar {
		calendar = &model.Calendar{
			ServiceID: patternID,
			Monday:    gcg.dayToBool(time.Monday, pattern.OperatingDays),
			Tuesday:   gcg.dayToBool(time.Tuesday, pattern.OperatingDays),
			Wednesday: gcg.dayToBool(time.Wednesday, pattern.OperatingDays),
			Thursday:  gcg.dayToBool(time.Thursday, pattern.OperatingDays),
			Friday:    gcg.dayToBool(time.Friday, pattern.OperatingDays),
			Saturday:  gcg.dayToBool(time.Saturday, pattern.OperatingDays),
			Sunday:    gcg.dayToBool(time.Sunday, pattern.OperatingDays),
			StartDate: pattern.ValidityPeriod.StartDate.Format("20060102"),
			EndDate:   pattern.ValidityPeriod.EndDate.Format("20060102"),
		}
	}

	// Process exceptions
	for _, exception := range pattern.Exceptions {
		calendarDate := &model.CalendarDate{
			ServiceID:     patternID,
			Date:          exception.Date.Format("20060102"),
			ExceptionType: gcg.mapExceptionType(exception.Type),
		}
		calendarDates = append(calendarDates, calendarDate)
	}

	// Process special days
	for _, specialDay := range pattern.SpecialDays {
		exceptionType := gcg.getSpecialDayExceptionType(specialDay)
		if exceptionType != 0 {
			calendarDate := &model.CalendarDate{
				ServiceID:     patternID,
				Date:          specialDay.Date.Format("20060102"),
				ExceptionType: exceptionType,
			}
			calendarDates = append(calendarDates, calendarDate)
		}
	}

	// If not using calendar.txt, generate calendar_dates for all service days
	if !useCalendar {
		serviceDates, err := gcg.generateServiceDates(pattern)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to generate service dates: %v", err)
		}

		for _, date := range serviceDates {
			calendarDate := &model.CalendarDate{
				ServiceID:     patternID,
				Date:          date.Format("20060102"),
				ExceptionType: 1, // Service added
			}
			calendarDates = append(calendarDates, calendarDate)
		}
	}

	return calendar, calendarDates, nil
}

// processOperatingPeriod converts an operating period to GTFS calendar format
func (gcg *GTFSCalendarGenerator) processOperatingPeriod(periodID string, period *OperatingPeriod) ([]*model.Calendar, []*model.CalendarDate, error) {
	calendars := make([]*model.Calendar, 0)
	allCalendarDates := make([]*model.CalendarDate, 0)

	// Process base pattern if present
	if period.BasePattern != nil {
		serviceID := fmt.Sprintf("%s_base", periodID)
		calendar, dates, err := gcg.processServicePatternWithID(serviceID, period.BasePattern)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to process base pattern: %v", err)
		}

		if calendar != nil {
			calendars = append(calendars, calendar)
		}
		allCalendarDates = append(allCalendarDates, dates...)
	}

	// Process override patterns
	for overrideID, overridePattern := range period.Overrides {
		serviceID := fmt.Sprintf("%s_%s", periodID, overrideID)
		calendar, dates, err := gcg.processServicePatternWithID(serviceID, overridePattern)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to process override pattern %s: %v", overrideID, err)
		}

		if calendar != nil {
			calendars = append(calendars, calendar)
		}
		allCalendarDates = append(allCalendarDates, dates...)
	}

	return calendars, allCalendarDates, nil
}

// processServicePatternWithID processes a service pattern with a specific service ID
func (gcg *GTFSCalendarGenerator) processServicePatternWithID(serviceID string, pattern *ServicePattern) (*model.Calendar, []*model.CalendarDate, error) {
	// Temporarily change the pattern ID for processing
	originalID := pattern.ID
	pattern.ID = serviceID

	calendar, dates, err := gcg.processServicePattern(serviceID, pattern)

	// Restore original ID
	pattern.ID = originalID

	return calendar, dates, err
}

// shouldUseCalendar determines whether to use calendar.txt or calendar_dates.txt
func (gcg *GTFSCalendarGenerator) shouldUseCalendar(pattern *ServicePattern) bool {
	if !gcg.optimizations.UseCalendarForRegularService {
		return false
	}

	// Use calendar.txt if:
	// 1. There are regular operating days
	// 2. The number of exceptions is reasonable
	// 3. The pattern covers a significant period

	hasRegularDays := len(pattern.OperatingDays) > 0
	exceptionCount := len(pattern.Exceptions) + len(pattern.SpecialDays)
	isLongPeriod := pattern.ValidityPeriod.EndDate.Sub(pattern.ValidityPeriod.StartDate) > 30*24*time.Hour

	return hasRegularDays &&
		exceptionCount <= gcg.optimizations.MaxCalendarDatesPerService &&
		isLongPeriod
}

// generateServiceDates generates all service dates for a pattern
func (gcg *GTFSCalendarGenerator) generateServiceDates(pattern *ServicePattern) ([]time.Time, error) {
	dates := make([]time.Time, 0)
	current := pattern.ValidityPeriod.StartDate

	for current.Before(pattern.ValidityPeriod.EndDate) || current.Equal(pattern.ValidityPeriod.EndDate) {
		if gcg.isServiceOperatingOnDate(pattern, current) {
			dates = append(dates, current)
		}
		current = current.AddDate(0, 0, 1)
	}

	return dates, nil
}

// isServiceOperatingOnDate checks if service operates on a specific date
func (gcg *GTFSCalendarGenerator) isServiceOperatingOnDate(pattern *ServicePattern, date time.Time) bool {
	// Check if date is in validity period
	if date.Before(pattern.ValidityPeriod.StartDate) || date.After(pattern.ValidityPeriod.EndDate) {
		return false
	}

	// Check operating days
	weekday := date.Weekday()
	isOperatingDay := false
	for _, day := range pattern.OperatingDays {
		if day == weekday {
			isOperatingDay = true
			break
		}
	}

	// Check exceptions (exceptions override operating days)
	for _, exception := range pattern.Exceptions {
		if exception.Date.Equal(date) {
			return exception.Type == ExceptionAdded
		}
	}

	// Check special days
	dateStr := date.Format("2006-01-02")
	if specialDay, exists := pattern.SpecialDays[dateStr]; exists {
		switch specialDay.ServiceMode {
		case ServiceNormal, ServiceReduced, ServiceReplacement:
			return true
		case ServiceSuspended:
			return false
		}
	}

	return isOperatingDay
}

// Helper methods

func (gcg *GTFSCalendarGenerator) dayToBool(day time.Weekday, operatingDays []time.Weekday) bool {
	for _, operatingDay := range operatingDays {
		if operatingDay == day {
			return true
		}
	}
	return false
}

func (gcg *GTFSCalendarGenerator) mapExceptionType(exceptionType ExceptionType) int {
	switch exceptionType {
	case ExceptionAdded:
		return 1
	case ExceptionRemoved:
		return 2
	default:
		return 2
	}
}

func (gcg *GTFSCalendarGenerator) getSpecialDayExceptionType(specialDay *SpecialDay) int {
	switch specialDay.ServiceMode {
	case ServiceNormal, ServiceReduced, ServiceReplacement:
		return 1
	case ServiceSuspended:
		return 2
	default:
		return 0
	}
}

// Optimization methods

func (gcg *GTFSCalendarGenerator) mergeCompatibleCalendars(calendars []*model.Calendar) []*model.Calendar {
	if !gcg.optimizations.MergeCompatibleCalendars {
		return calendars
	}

	// Group calendars by their operating pattern
	patternGroups := make(map[string][]*model.Calendar)

	for _, calendar := range calendars {
		pattern := fmt.Sprintf("%t_%t_%t_%t_%t_%t_%t_%s_%s",
			calendar.Monday, calendar.Tuesday, calendar.Wednesday, calendar.Thursday,
			calendar.Friday, calendar.Saturday, calendar.Sunday,
			calendar.StartDate, calendar.EndDate)

		patternGroups[pattern] = append(patternGroups[pattern], calendar)
	}

	// Keep one calendar per pattern (could be enhanced to merge service IDs)
	result := make([]*model.Calendar, 0, len(patternGroups))
	for _, group := range patternGroups {
		result = append(result, group[0]) // Take the first one
	}

	return result
}

func (gcg *GTFSCalendarGenerator) minimizeCalendarDates(calendarDates []*model.CalendarDate, calendars []*model.Calendar) []*model.CalendarDate {
	if !gcg.optimizations.MinimizeCalendarDates {
		return calendarDates
	}

	// Remove redundant calendar dates that are already covered by calendar.txt
	calendarServiceIDs := make(map[string]bool)
	for _, calendar := range calendars {
		calendarServiceIDs[calendar.ServiceID] = true
	}

	// Keep all calendar dates - the logic is the same regardless of calendar.txt presence
	result := make([]*model.CalendarDate, 0, len(calendarDates))
	result = append(result, calendarDates...)

	return result
}

func (gcg *GTFSCalendarGenerator) sortCalendars(calendars []*model.Calendar) {
	sort.Slice(calendars, func(i, j int) bool {
		return calendars[i].ServiceID < calendars[j].ServiceID
	})
}

func (gcg *GTFSCalendarGenerator) sortCalendarDates(calendarDates []*model.CalendarDate) {
	sort.Slice(calendarDates, func(i, j int) bool {
		if calendarDates[i].ServiceID != calendarDates[j].ServiceID {
			return calendarDates[i].ServiceID < calendarDates[j].ServiceID
		}
		return calendarDates[i].Date < calendarDates[j].Date
	})
}

// GenerateCalendarSummary generates a summary of calendar generation results
func (gcg *GTFSCalendarGenerator) GenerateCalendarSummary(calendars []*model.Calendar, calendarDates []*model.CalendarDate) map[string]interface{} {
	summary := make(map[string]interface{})

	summary["total_calendars"] = len(calendars)
	summary["total_calendar_dates"] = len(calendarDates)

	// Count services by type
	servicesByType := make(map[string]int)
	for _, calendar := range calendars {
		operatingDays := 0
		if calendar.Monday {
			operatingDays++
		}
		if calendar.Tuesday {
			operatingDays++
		}
		if calendar.Wednesday {
			operatingDays++
		}
		if calendar.Thursday {
			operatingDays++
		}
		if calendar.Friday {
			operatingDays++
		}
		if calendar.Saturday {
			operatingDays++
		}
		if calendar.Sunday {
			operatingDays++
		}

		switch {
		case operatingDays == 5 && !calendar.Saturday && !calendar.Sunday:
			servicesByType["weekday"]++
		case operatingDays == 2 && calendar.Saturday && calendar.Sunday:
			servicesByType["weekend"]++
		case operatingDays == 7:
			servicesByType["daily"]++
		default:
			servicesByType["other"]++
		}
	}
	summary["services_by_type"] = servicesByType

	// Count exception types
	exceptionsByType := make(map[string]int)
	for _, date := range calendarDates {
		switch date.ExceptionType {
		case 1:
			exceptionsByType["added"]++
		case 2:
			exceptionsByType["removed"]++
		}
	}
	summary["exceptions_by_type"] = exceptionsByType

	return summary
}
