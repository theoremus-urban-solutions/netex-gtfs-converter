package calendar

import (
	"fmt"
	"sort"
	"time"

	"github.com/theoremus-urban-solutions/netex-gtfs-converter/model"
)

const (
	unknownPatternType = "Unknown"
	defaultTimezone    = "Europe/Oslo"
)

// CalendarManager handles complex European transit calendar patterns
type CalendarManager struct {
	calendars        map[string]*model.Calendar
	calendarDates    map[string][]*model.CalendarDate
	servicePatterns  map[string]*ServicePattern
	holidayDetector  *HolidayDetector
	seasonalPatterns map[string]*SeasonalPattern
	operatingPeriods map[string]*OperatingPeriod
	config           CalendarConfig
}

// CalendarConfig controls calendar processing behavior
type CalendarConfig struct {
	EnableHolidayDetection   bool   `json:"enable_holiday_detection"`
	EnableSeasonalPatterns   bool   `json:"enable_seasonal_patterns"`
	EnableSchoolCalendar     bool   `json:"enable_school_calendar"`
	MaxServiceExceptions     int    `json:"max_service_exceptions"`
	DefaultOperatingDays     int    `json:"default_operating_days"`
	HolidayCountryCode       string `json:"holiday_country_code"`
	TimezoneName             string `json:"timezone_name"`
	EnableWeekendAdjustments bool   `json:"enable_weekend_adjustments"`
}

// ServicePattern represents a complex service operating pattern
type ServicePattern struct {
	ID                 string                 `json:"id"`
	Name               string                 `json:"name"`
	Type               ServicePatternType     `json:"type"`
	BaseCalendar       *model.Calendar        `json:"base_calendar"`
	Exceptions         []*ServiceException    `json:"exceptions"`
	SeasonalVariations []*SeasonalVariation   `json:"seasonal_variations"`
	SpecialDays        map[string]*SpecialDay `json:"special_days"`
	ValidityPeriod     *ValidityPeriod        `json:"validity_period"`
	OperatingDays      []time.Weekday         `json:"operating_days"`
	NonOperatingDays   []time.Weekday         `json:"non_operating_days"`
	HolidayBehavior    HolidayBehavior        `json:"holiday_behavior"`
}

// ServicePatternType defines different types of service patterns
type ServicePatternType int

const (
	PatternRegular ServicePatternType = iota
	PatternSeasonal
	PatternSpecialEvent
	PatternSchoolTerm
	PatternHoliday
	PatternWeekend
	PatternNightService
	PatternReplacementService
)

func (t ServicePatternType) String() string {
	switch t {
	case PatternRegular:
		return "Regular"
	case PatternSeasonal:
		return "Seasonal"
	case PatternSpecialEvent:
		return "SpecialEvent"
	case PatternSchoolTerm:
		return "SchoolTerm"
	case PatternHoliday:
		return "Holiday"
	case PatternWeekend:
		return "Weekend"
	case PatternNightService:
		return "NightService"
	case PatternReplacementService:
		return "ReplacementService"
	default:
		return unknownPatternType
	}
}

// ServiceException represents a service exception on specific dates
type ServiceException struct {
	Date        time.Time           `json:"date"`
	Type        ExceptionType       `json:"type"`
	Reason      string              `json:"reason"`
	Alternative *AlternativeService `json:"alternative,omitempty"`
}

// ExceptionType defines types of service exceptions
type ExceptionType int

const (
	ExceptionRemoved ExceptionType = iota
	ExceptionAdded
	ExceptionModified
	ExceptionReplaced
)

func (t ExceptionType) String() string {
	switch t {
	case ExceptionRemoved:
		return "Removed"
	case ExceptionAdded:
		return "Added"
	case ExceptionModified:
		return "Modified"
	case ExceptionReplaced:
		return "Replaced"
	default:
		return unknownPatternType
	}
}

// SeasonalVariation represents seasonal service changes
type SeasonalVariation struct {
	Season    Season          `json:"season"`
	StartDate time.Time       `json:"start_date"`
	EndDate   time.Time       `json:"end_date"`
	Changes   []ServiceChange `json:"changes"`
}

// Season represents different seasons for service patterns
type Season int

const (
	SeasonSpring Season = iota
	SeasonSummer
	SeasonAutumn
	SeasonWinter
	SeasonSchoolTerm
	SeasonSchoolHoliday
)

func (s Season) String() string {
	switch s {
	case SeasonSpring:
		return "Spring"
	case SeasonSummer:
		return "Summer"
	case SeasonAutumn:
		return "Autumn"
	case SeasonWinter:
		return "Winter"
	case SeasonSchoolTerm:
		return "SchoolTerm"
	case SeasonSchoolHoliday:
		return "SchoolHoliday"
	default:
		return unknownPatternType
	}
}

// ServiceChange represents a change to service during seasonal variations
type ServiceChange struct {
	Type        ChangeType             `json:"type"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// ChangeType defines types of service changes
type ChangeType int

const (
	ChangeFrequency ChangeType = iota
	ChangeRoute
	ChangeOperatingDays
	ChangeTimings
	ChangeCancellation
)

// SpecialDay represents special operating days
type SpecialDay struct {
	Date        time.Time      `json:"date"`
	Name        string         `json:"name"`
	Type        SpecialDayType `json:"type"`
	ServiceMode ServiceMode    `json:"service_mode"`
}

// SpecialDayType defines types of special days
type SpecialDayType int

const (
	SpecialHoliday SpecialDayType = iota
	SpecialEvent
	SpecialMaintenance
	SpecialStrike
	SpecialWeather
)

// ServiceMode defines how service operates on special days
type ServiceMode int

const (
	ServiceNormal ServiceMode = iota
	ServiceReduced
	ServiceHoliday
	ServiceSuspended
	ServiceReplacement
)

// ValidityPeriod represents the period during which a service pattern is valid
type ValidityPeriod struct {
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
}

// HolidayBehavior defines how service behaves on holidays
type HolidayBehavior int

const (
	HolidayAsWeekday HolidayBehavior = iota
	HolidayAsWeekend
	HolidaySpecialSchedule
	HolidayNoService
)

// AlternativeService represents alternative service during exceptions
type AlternativeService struct {
	ServiceID   string   `json:"service_id"`
	Description string   `json:"description"`
	Routes      []string `json:"routes"`
}

// SeasonalPattern represents complex seasonal operating patterns
type SeasonalPattern struct {
	ID          string              `json:"id"`
	Name        string              `json:"name"`
	Seasons     []*SeasonDefinition `json:"seasons"`
	Transitions []*SeasonTransition `json:"transitions"`
}

// SeasonDefinition defines a specific season with its characteristics
type SeasonDefinition struct {
	Season          Season                 `json:"season"`
	StartMonth      time.Month             `json:"start_month"`
	StartDay        int                    `json:"start_day"`
	EndMonth        time.Month             `json:"end_month"`
	EndDay          int                    `json:"end_day"`
	Characteristics map[string]interface{} `json:"characteristics"`
}

// SeasonTransition defines how to transition between seasons
type SeasonTransition struct {
	FromSeason    Season        `json:"from_season"`
	ToSeason      Season        `json:"to_season"`
	Duration      time.Duration `json:"duration"`
	GradualChange bool          `json:"gradual_change"`
}

// OperatingPeriod represents a complex operating period with multiple patterns
type OperatingPeriod struct {
	ID          string                     `json:"id"`
	Name        string                     `json:"name"`
	StartDate   time.Time                  `json:"start_date"`
	EndDate     time.Time                  `json:"end_date"`
	BasePattern *ServicePattern            `json:"base_pattern"`
	Overrides   map[string]*ServicePattern `json:"overrides"`
	Priority    int                        `json:"priority"`
}

// NewCalendarManager creates a new calendar manager
func NewCalendarManager(config CalendarConfig) *CalendarManager {
	if config.HolidayCountryCode == "" {
		config.HolidayCountryCode = "NO" // Default to Norway for European NeTEx
	}
	if config.TimezoneName == "" {
		config.TimezoneName = defaultTimezone
	}
	if config.MaxServiceExceptions == 0 {
		config.MaxServiceExceptions = 500
	}
	if config.DefaultOperatingDays == 0 {
		config.DefaultOperatingDays = 365
	}

	return &CalendarManager{
		calendars:        make(map[string]*model.Calendar),
		calendarDates:    make(map[string][]*model.CalendarDate),
		servicePatterns:  make(map[string]*ServicePattern),
		holidayDetector:  NewHolidayDetector(config.HolidayCountryCode),
		seasonalPatterns: make(map[string]*SeasonalPattern),
		operatingPeriods: make(map[string]*OperatingPeriod),
		config:           config,
	}
}

// AddServicePattern adds a service pattern to the manager
func (cm *CalendarManager) AddServicePattern(pattern *ServicePattern) {
	cm.servicePatterns[pattern.ID] = pattern
}

// AddSeasonalPattern adds a seasonal pattern to the manager
func (cm *CalendarManager) AddSeasonalPattern(pattern *SeasonalPattern) {
	cm.seasonalPatterns[pattern.ID] = pattern
}

// AddOperatingPeriod adds an operating period to the manager
func (cm *CalendarManager) AddOperatingPeriod(period *OperatingPeriod) {
	cm.operatingPeriods[period.ID] = period
}

// GenerateGTFSCalendar generates GTFS calendar entries from NeTEx service data
func (cm *CalendarManager) GenerateGTFSCalendar(netexServiceFrame interface{}) ([]*model.Calendar, []*model.CalendarDate, error) {
	calendars := make([]*model.Calendar, 0)
	calendarDates := make([]*model.CalendarDate, 0)

	// Process all service patterns
	for _, pattern := range cm.servicePatterns {
		calendar, dates, err := cm.processServicePattern(pattern)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to process service pattern %s: %v", pattern.ID, err)
		}

		if calendar != nil {
			calendars = append(calendars, calendar)
			cm.calendars[calendar.ServiceID] = calendar
		}

		if len(dates) > 0 {
			calendarDates = append(calendarDates, dates...)
			cm.calendarDates[pattern.ID] = dates
		}
	}

	// Process operating periods
	for _, period := range cm.operatingPeriods {
		periodCalendars, periodDates, err := cm.processOperatingPeriod(period)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to process operating period %s: %v", period.ID, err)
		}

		calendars = append(calendars, periodCalendars...)
		calendarDates = append(calendarDates, periodDates...)
	}

	// Apply seasonal patterns if enabled
	if cm.config.EnableSeasonalPatterns {
		seasonalDates, err := cm.generateSeasonalExceptions()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to generate seasonal exceptions: %v", err)
		}
		calendarDates = append(calendarDates, seasonalDates...)
	}

	// Apply holiday detection if enabled
	if cm.config.EnableHolidayDetection {
		holidayDates, err := cm.generateHolidayExceptions()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to generate holiday exceptions: %v", err)
		}
		calendarDates = append(calendarDates, holidayDates...)
	}

	// Sort and deduplicate calendar dates
	calendarDates = cm.consolidateCalendarDates(calendarDates)

	return calendars, calendarDates, nil
}

// processServicePattern processes a single service pattern
func (cm *CalendarManager) processServicePattern(pattern *ServicePattern) (*model.Calendar, []*model.CalendarDate, error) {
	if pattern.BaseCalendar == nil {
		return nil, nil, fmt.Errorf("service pattern %s has no base calendar", pattern.ID)
	}

	calendar := &model.Calendar{
		ServiceID: pattern.ID,
		Monday:    false,
		Tuesday:   false,
		Wednesday: false,
		Thursday:  false,
		Friday:    false,
		Saturday:  false,
		Sunday:    false,
		StartDate: pattern.ValidityPeriod.StartDate.Format("20060102"),
		EndDate:   pattern.ValidityPeriod.EndDate.Format("20060102"),
	}

	// Set operating days
	for _, day := range pattern.OperatingDays {
		switch day {
		case time.Monday:
			calendar.Monday = true
		case time.Tuesday:
			calendar.Tuesday = true
		case time.Wednesday:
			calendar.Wednesday = true
		case time.Thursday:
			calendar.Thursday = true
		case time.Friday:
			calendar.Friday = true
		case time.Saturday:
			calendar.Saturday = true
		case time.Sunday:
			calendar.Sunday = true
		}
	}

	// Generate calendar dates for exceptions
	calendarDates := make([]*model.CalendarDate, 0)

	for _, exception := range pattern.Exceptions {
		calendarDate := &model.CalendarDate{
			ServiceID:     pattern.ID,
			Date:          exception.Date.Format("20060102"),
			ExceptionType: cm.mapExceptionType(exception.Type),
		}
		calendarDates = append(calendarDates, calendarDate)
	}

	// Process special days
	for _, specialDay := range pattern.SpecialDays {
		exceptionType := cm.getSpecialDayExceptionType(specialDay)
		calendarDate := &model.CalendarDate{
			ServiceID:     pattern.ID,
			Date:          specialDay.Date.Format("20060102"),
			ExceptionType: exceptionType,
		}
		calendarDates = append(calendarDates, calendarDate)
	}

	return calendar, calendarDates, nil
}

// processOperatingPeriod processes an operating period with multiple patterns
func (cm *CalendarManager) processOperatingPeriod(period *OperatingPeriod) ([]*model.Calendar, []*model.CalendarDate, error) {
	calendars := make([]*model.Calendar, 0)
	allCalendarDates := make([]*model.CalendarDate, 0)

	// Process base pattern
	if period.BasePattern != nil {
		calendar, dates, err := cm.processServicePattern(period.BasePattern)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to process base pattern: %v", err)
		}

		if calendar != nil {
			// Update calendar ID to include period
			calendar.ServiceID = fmt.Sprintf("%s_%s", period.ID, period.BasePattern.ID)
			calendars = append(calendars, calendar)
		}

		// Update calendar dates service IDs
		for _, date := range dates {
			date.ServiceID = fmt.Sprintf("%s_%s", period.ID, period.BasePattern.ID)
		}
		allCalendarDates = append(allCalendarDates, dates...)
	}

	// Process override patterns
	for overrideID, overridePattern := range period.Overrides {
		calendar, dates, err := cm.processServicePattern(overridePattern)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to process override pattern %s: %v", overrideID, err)
		}

		if calendar != nil {
			calendar.ServiceID = fmt.Sprintf("%s_%s", period.ID, overrideID)
			calendars = append(calendars, calendar)
		}

		for _, date := range dates {
			date.ServiceID = fmt.Sprintf("%s_%s", period.ID, overrideID)
		}
		allCalendarDates = append(allCalendarDates, dates...)
	}

	return calendars, allCalendarDates, nil
}

// generateSeasonalExceptions generates calendar exceptions for seasonal patterns
func (cm *CalendarManager) generateSeasonalExceptions() ([]*model.CalendarDate, error) {
	calendarDates := make([]*model.CalendarDate, 0)

	for _, seasonalPattern := range cm.seasonalPatterns {
		for _, season := range seasonalPattern.Seasons {
			// Generate exceptions for seasonal transitions
			exceptions, err := cm.generateSeasonExceptions(season)
			if err != nil {
				return nil, fmt.Errorf("failed to generate exceptions for season %s: %v", season.Season.String(), err)
			}
			calendarDates = append(calendarDates, exceptions...)
		}
	}

	return calendarDates, nil
}

// generateHolidayExceptions generates calendar exceptions for holidays
func (cm *CalendarManager) generateHolidayExceptions() ([]*model.CalendarDate, error) {
	calendarDates := make([]*model.CalendarDate, 0)
	currentYear := time.Now().Year()

	// Generate holiday exceptions for current and next year
	for year := currentYear; year <= currentYear+1; year++ {
		holidays, err := cm.holidayDetector.GetHolidays(year)
		if err != nil {
			return nil, fmt.Errorf("failed to get holidays for year %d: %v", year, err)
		}

		for _, holiday := range holidays {
			// Create calendar exceptions for each service pattern
			for serviceID := range cm.servicePatterns {
				pattern := cm.servicePatterns[serviceID]
				exceptionType := cm.getHolidayExceptionType(pattern.HolidayBehavior, holiday)

				if exceptionType != 0 {
					calendarDate := &model.CalendarDate{
						ServiceID:     serviceID,
						Date:          holiday.Date.Format("20060102"),
						ExceptionType: exceptionType,
					}
					calendarDates = append(calendarDates, calendarDate)
				}
			}
		}
	}

	return calendarDates, nil
}

// generateSeasonExceptions generates exceptions for a specific season
func (cm *CalendarManager) generateSeasonExceptions(season *SeasonDefinition) ([]*model.CalendarDate, error) {
	calendarDates := make([]*model.CalendarDate, 0)

	// Implementation would depend on specific seasonal characteristics
	// This is a placeholder for complex seasonal logic

	return calendarDates, nil
}

// consolidateCalendarDates removes duplicates and sorts calendar dates
func (cm *CalendarManager) consolidateCalendarDates(dates []*model.CalendarDate) []*model.CalendarDate {
	// Create a map to track unique entries
	unique := make(map[string]*model.CalendarDate)

	for _, date := range dates {
		key := fmt.Sprintf("%s_%s", date.ServiceID, date.Date)
		if existing, exists := unique[key]; exists {
			// Prioritize removal over addition
			if date.ExceptionType == 2 && existing.ExceptionType == 1 {
				unique[key] = date
			}
		} else {
			unique[key] = date
		}
	}

	// Convert back to slice and sort
	result := make([]*model.CalendarDate, 0, len(unique))
	for _, date := range unique {
		result = append(result, date)
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].ServiceID != result[j].ServiceID {
			return result[i].ServiceID < result[j].ServiceID
		}
		return result[i].Date < result[j].Date
	})

	return result
}

// Helper methods

func (cm *CalendarManager) mapExceptionType(exceptionType ExceptionType) int {
	switch exceptionType {
	case ExceptionAdded:
		return 1
	case ExceptionRemoved:
		return 2
	default:
		return 2 // Default to removed
	}
}

func (cm *CalendarManager) getSpecialDayExceptionType(specialDay *SpecialDay) int {
	switch specialDay.ServiceMode {
	case ServiceNormal:
		return 1
	case ServiceSuspended:
		return 2
	default:
		return 2
	}
}

func (cm *CalendarManager) getHolidayExceptionType(behavior HolidayBehavior, holiday *Holiday) int {
	switch behavior {
	case HolidayAsWeekend:
		return 2 // Remove service
	case HolidayNoService:
		return 2 // Remove service
	case HolidaySpecialSchedule:
		return 1 // Add special service
	default:
		return 0 // No exception needed
	}
}

// GetServicePattern retrieves a service pattern by ID
func (cm *CalendarManager) GetServicePattern(id string) *ServicePattern {
	return cm.servicePatterns[id]
}

// GetCalendars returns all generated calendars
func (cm *CalendarManager) GetCalendars() map[string]*model.Calendar {
	return cm.calendars
}

// GetCalendarDates returns all generated calendar dates
func (cm *CalendarManager) GetCalendarDates() map[string][]*model.CalendarDate {
	return cm.calendarDates
}

// ValidateServicePattern validates a service pattern for consistency
func (cm *CalendarManager) ValidateServicePattern(pattern *ServicePattern) []string {
	issues := make([]string, 0)

	if pattern.ID == "" {
		issues = append(issues, "Service pattern missing ID")
	}

	if pattern.ValidityPeriod == nil {
		issues = append(issues, "Service pattern missing validity period")
	} else if pattern.ValidityPeriod.StartDate.After(pattern.ValidityPeriod.EndDate) {
		issues = append(issues, "Service pattern start date is after end date")
	}

	if len(pattern.OperatingDays) == 0 && len(pattern.Exceptions) == 0 {
		issues = append(issues, "Service pattern has no operating days or exceptions")
	}

	return issues
}

// GetEffectiveDates returns all dates when a service pattern is effective
func (cm *CalendarManager) GetEffectiveDates(patternID string, startDate, endDate time.Time) ([]time.Time, error) {
	pattern := cm.servicePatterns[patternID]
	if pattern == nil {
		return nil, fmt.Errorf("service pattern %s not found", patternID)
	}

	dates := make([]time.Time, 0)
	current := startDate

	for current.Before(endDate) || current.Equal(endDate) {
		if cm.isServiceOperating(pattern, current) {
			dates = append(dates, current)
		}
		current = current.AddDate(0, 0, 1)
	}

	return dates, nil
}

// isServiceOperating checks if service is operating on a specific date
func (cm *CalendarManager) isServiceOperating(pattern *ServicePattern, date time.Time) bool {
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

	// Check exceptions
	for _, exception := range pattern.Exceptions {
		if exception.Date.Equal(date) {
			switch exception.Type {
			case ExceptionAdded:
				return true
			case ExceptionRemoved:
				return false
			}
		}
	}

	// Check special days
	if specialDay, exists := pattern.SpecialDays[date.Format("2006-01-02")]; exists {
		switch specialDay.ServiceMode {
		case ServiceNormal, ServiceReduced, ServiceReplacement:
			return true
		case ServiceSuspended:
			return false
		}
	}

	return isOperatingDay
}
