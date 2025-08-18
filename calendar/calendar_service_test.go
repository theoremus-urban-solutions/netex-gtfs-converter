package calendar

import (
	"fmt"
	"testing"
	"time"
)

func TestNewCalendarService(t *testing.T) {
	config := CalendarServiceConfig{
		DefaultTimezoneName:     "Europe/Oslo",
		HolidayCountryCode:     "NO",
		EnableHolidayDetection: true,
		ValidationLevel:        ValidationStandard,
	}

	service, err := NewCalendarService(config)
	if err != nil {
		t.Errorf("Unexpected error creating calendar service: %v", err)
	}

	if service == nil {
		t.Error("Expected non-nil calendar service")
	}

	if service.config.DefaultTimezoneName != "Europe/Oslo" {
		t.Errorf("Expected timezone Europe/Oslo, got %s", service.config.DefaultTimezoneName)
	}
}

func TestCalendarServiceDefaults(t *testing.T) {
	config := CalendarServiceConfig{} // Empty config

	service, err := NewCalendarService(config)
	if err != nil {
		t.Errorf("Unexpected error creating calendar service: %v", err)
	}

	// Check defaults
	if service.config.DefaultTimezoneName != "Europe/Oslo" {
		t.Errorf("Expected default timezone Europe/Oslo, got %s", service.config.DefaultTimezoneName)
	}

	if service.config.HolidayCountryCode != "NO" {
		t.Errorf("Expected default country code NO, got %s", service.config.HolidayCountryCode)
	}

	if service.config.MaxServiceExceptions != 500 {
		t.Errorf("Expected default max exceptions 500, got %d", service.config.MaxServiceExceptions)
	}

	if service.config.ValidationLevel != ValidationStandard {
		t.Errorf("Expected default validation level Standard, got %s", service.config.ValidationLevel.String())
	}
}

func TestAddCustomServicePattern(t *testing.T) {
	service, _ := NewCalendarService(CalendarServiceConfig{})

	pattern := &ServicePattern{
		ID:   "custom_pattern",
		Name: "Custom Pattern",
		Type: PatternRegular,
		ValidityPeriod: &ValidityPeriod{
			StartDate: time.Now(),
			EndDate:   time.Now().AddDate(1, 0, 0),
		},
		OperatingDays:   []time.Weekday{time.Monday, time.Tuesday},
		Exceptions:      make([]*ServiceException, 0),
		SpecialDays:     make(map[string]*SpecialDay),
		HolidayBehavior: HolidayAsWeekday,
	}

	err := service.AddCustomServicePattern(pattern)
	if err != nil {
		t.Errorf("Unexpected error adding custom pattern: %v", err)
	}

	// Verify pattern was added
	retrievedPattern := service.manager.GetServicePattern("custom_pattern")
	if retrievedPattern == nil {
		t.Error("Expected to retrieve added custom pattern")
	}
}

func TestAddInvalidServicePattern(t *testing.T) {
	config := CalendarServiceConfig{
		ValidationLevel: ValidationStrict,
	}
	service, _ := NewCalendarService(config)

	// Create invalid pattern (missing ID)
	pattern := &ServicePattern{
		Name: "Invalid Pattern",
		Type: PatternRegular,
		ValidityPeriod: &ValidityPeriod{
			StartDate: time.Now().AddDate(1, 0, 0), // Start after end
			EndDate:   time.Now(),
		},
	}

	err := service.AddCustomServicePattern(pattern)
	if err == nil {
		t.Error("Expected error when adding invalid pattern")
	}
}

func TestConvertNeTExToGTFS(t *testing.T) {
	service, _ := NewCalendarService(CalendarServiceConfig{
		EnableHolidayDetection: false, // Disable for simpler testing
		ValidationLevel:        ValidationMinimal,
	})

	// Add some test patterns first
	weekdayPattern := &ServicePattern{
		ID:   "weekday_service",
		Name: "Weekday Service",
		Type: PatternRegular,
		ValidityPeriod: &ValidityPeriod{
			StartDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:   time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC),
		},
		OperatingDays: []time.Weekday{
			time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday,
		},
		Exceptions:      make([]*ServiceException, 0),
		SpecialDays:     make(map[string]*SpecialDay),
		HolidayBehavior: HolidayAsWeekday,
	}

	service.AddCustomServicePattern(weekdayPattern)

	// Convert to GTFS
	result, err := service.ConvertNeTExToGTFS(nil)
	if err != nil {
		t.Errorf("Unexpected error converting to GTFS: %v", err)
	}

	if result == nil {
		t.Error("Expected non-nil conversion result")
	}

	if len(result.Calendars) == 0 {
		t.Error("Expected at least one calendar to be generated")
	}

	if result.ConversionStats.TotalServicePatterns == 0 {
		t.Error("Expected service patterns to be counted in stats")
	}

	// Check processing duration
	if result.ProcessingDuration <= 0 {
		t.Error("Expected positive processing duration")
	}

	t.Logf("Conversion result: %d calendars, %d calendar dates, %d patterns",
		len(result.Calendars), len(result.CalendarDates), len(result.ServicePatterns))
}

func TestGetServiceDates(t *testing.T) {
	service, _ := NewCalendarService(CalendarServiceConfig{})

	pattern := &ServicePattern{
		ID:   "test_dates",
		Name: "Test Dates Pattern",
		Type: PatternRegular,
		ValidityPeriod: &ValidityPeriod{
			StartDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), // Monday
			EndDate:   time.Date(2024, 1, 7, 0, 0, 0, 0, time.UTC),  // Sunday
		},
		OperatingDays: []time.Weekday{time.Monday, time.Wednesday, time.Friday},
		Exceptions:    make([]*ServiceException, 0),
		SpecialDays:   make(map[string]*SpecialDay),
	}

	service.AddCustomServicePattern(pattern)

	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 1, 7, 0, 0, 0, 0, time.UTC)

	dates, err := service.GetServiceDates("test_dates", startDate, endDate)
	if err != nil {
		t.Errorf("Unexpected error getting service dates: %v", err)
	}

	// Should have 3 dates: Mon, Wed, Fri
	expectedDates := 3
	if len(dates) != expectedDates {
		t.Errorf("Expected %d service dates, got %d", expectedDates, len(dates))
	}

	// Verify the specific dates
	expectedWeekdays := map[time.Weekday]bool{
		time.Monday:    true,
		time.Wednesday: true,
		time.Friday:    true,
	}

	for _, date := range dates {
		weekday := date.Weekday()
		if !expectedWeekdays[weekday] {
			t.Errorf("Unexpected service date weekday: %s", weekday.String())
		}
	}
}

func TestIsServiceOperating(t *testing.T) {
	service, _ := NewCalendarService(CalendarServiceConfig{})

	pattern := &ServicePattern{
		ID:   "operating_test",
		Name: "Operating Test Pattern",
		Type: PatternRegular,
		ValidityPeriod: &ValidityPeriod{
			StartDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:   time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC),
		},
		OperatingDays: []time.Weekday{time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday},
		Exceptions: []*ServiceException{
			{
				Date:   time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC), // Wednesday
				Type:   ExceptionRemoved,
				Reason: "Holiday",
			},
		},
		SpecialDays: make(map[string]*SpecialDay),
	}

	service.AddCustomServicePattern(pattern)

	tests := []struct {
		name     string
		date     time.Time
		expected bool
	}{
		{
			name:     "Normal weekday",
			date:     time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC), // Tuesday
			expected: true,
		},
		{
			name:     "Exception day",
			date:     time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC), // Wednesday with exception
			expected: false,
		},
		{
			name:     "Weekend",
			date:     time.Date(2024, 1, 6, 0, 0, 0, 0, time.UTC), // Saturday
			expected: false,
		},
		{
			name:     "Outside validity period",
			date:     time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC), // Before start
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isOperating, err := service.IsServiceOperating("operating_test", tt.date)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if isOperating != tt.expected {
				t.Errorf("Expected service operating %t for %s, got %t",
					tt.expected, tt.date.Format("2006-01-02"), isOperating)
			}
		})
	}
}

func TestValidateConfiguration(t *testing.T) {
	tests := []struct {
		name           string
		config         CalendarServiceConfig
		expectedIssues int
	}{
		{
			name: "valid config",
			config: CalendarServiceConfig{
				DefaultTimezoneName:  "Europe/Oslo",
				HolidayCountryCode:  "NO",
				MaxServiceExceptions: 100,
			},
			expectedIssues: 0,
		},
		{
			name: "invalid timezone",
			config: CalendarServiceConfig{
				DefaultTimezoneName:  "Invalid/Timezone",
				HolidayCountryCode:  "NO",
				MaxServiceExceptions: 100,
			},
			expectedIssues: 1,
		},
		{
			name: "invalid country code",
			config: CalendarServiceConfig{
				DefaultTimezoneName:  "Europe/Oslo",
				HolidayCountryCode:  "INVALID",
				MaxServiceExceptions: 100,
			},
			expectedIssues: 1,
		},
		{
			name: "negative max exceptions",
			config: CalendarServiceConfig{
				DefaultTimezoneName:  "Europe/Oslo",
				HolidayCountryCode:  "NO",
				MaxServiceExceptions: -1,
			},
			expectedIssues: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, _ := NewCalendarService(tt.config)
			issues := service.ValidateConfiguration()

			if len(issues) != tt.expectedIssues {
				t.Errorf("Expected %d validation issues, got %d: %v",
					tt.expectedIssues, len(issues), issues)
			}
		})
	}
}

func TestGetHolidays(t *testing.T) {
	config := CalendarServiceConfig{
		EnableHolidayDetection: true,
		HolidayCountryCode:    "NO",
	}
	service, _ := NewCalendarService(config)

	year := 2024
	holidays, err := service.GetHolidays(year)
	if err != nil {
		t.Errorf("Unexpected error getting holidays: %v", err)
	}

	if len(holidays) == 0 {
		t.Error("Expected some holidays to be returned")
	}

	// Check for Norwegian national day
	foundNationalDay := false
	for _, holiday := range holidays {
		if holiday.Name == "Constitution Day" {
			foundNationalDay = true
			expectedDate := time.Date(2024, time.May, 17, 0, 0, 0, 0, time.UTC)
			if !holiday.Date.Equal(expectedDate) {
				t.Errorf("Expected Constitution Day on %s, got %s",
					expectedDate.Format("2006-01-02"), holiday.Date.Format("2006-01-02"))
			}
			break
		}
	}

	if !foundNationalDay {
		t.Error("Expected to find Norwegian Constitution Day")
	}
}

func TestConversionWithOptimizations(t *testing.T) {
	config := CalendarServiceConfig{
		OptimizeCalendarDates:      true,
		ConsolidateSimilarPatterns: true,
		ValidationLevel:            ValidationMinimal,
	}
	service, _ := NewCalendarService(config)

	// Add similar patterns
	for i := 0; i < 3; i++ {
		pattern := &ServicePattern{
			ID:   fmt.Sprintf("similar_pattern_%d", i),
			Name: fmt.Sprintf("Similar Pattern %d", i),
			Type: PatternRegular,
			ValidityPeriod: &ValidityPeriod{
				StartDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				EndDate:   time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC),
			},
			OperatingDays:   []time.Weekday{time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday},
			Exceptions:      make([]*ServiceException, 0),
			SpecialDays:     make(map[string]*SpecialDay),
			HolidayBehavior: HolidayAsWeekday,
		}
		service.AddCustomServicePattern(pattern)
	}

	result, err := service.ConvertNeTExToGTFS(nil)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	t.Logf("Conversion with optimizations: %d calendars, %d calendar dates",
		len(result.Calendars), len(result.CalendarDates))

	// Should have processed at least our custom patterns (may include defaults)
	if result.ConversionStats.TotalServicePatterns < 3 {
		t.Errorf("Expected at least 3 service patterns, got %d", result.ConversionStats.TotalServicePatterns)
	}
}

func TestGetConversionSummary(t *testing.T) {
	service, _ := NewCalendarService(CalendarServiceConfig{})

	pattern := &ServicePattern{
		ID:   "summary_test",
		Name: "Summary Test Pattern",
		Type: PatternRegular,
		ValidityPeriod: &ValidityPeriod{
			StartDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:   time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC),
		},
		OperatingDays:   []time.Weekday{time.Monday},
		Exceptions:      make([]*ServiceException, 0),
		SpecialDays:     make(map[string]*SpecialDay),
		HolidayBehavior: HolidayAsWeekday,
	}

	service.AddCustomServicePattern(pattern)

	// Generate GTFS to populate internal structures
	service.ConvertNeTExToGTFS(nil)

	summary := service.GetConversionSummary()

	if summary == nil {
		t.Error("Expected non-nil summary")
	}

	expectedKeys := []string{
		"service_patterns_count",
		"operating_periods_count",
		"calendars_count",
		"validation_level",
	}

	for _, key := range expectedKeys {
		if _, exists := summary[key]; !exists {
			t.Errorf("Expected summary to contain key: %s", key)
		}
	}

	if summary["service_patterns_count"].(int) == 0 {
		t.Error("Expected non-zero service patterns count in summary")
	}

	t.Logf("Conversion summary: %+v", summary)
}