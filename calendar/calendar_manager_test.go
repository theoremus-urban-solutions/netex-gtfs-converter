package calendar

import (
	"testing"
	"time"

	"github.com/theoremus-urban-solutions/netex-gtfs-converter/model"
)

func TestNewCalendarManager(t *testing.T) {
	config := CalendarConfig{
		EnableHolidayDetection: true,
		HolidayCountryCode:     "NO",
		TimezoneName:           "Europe/Oslo",
	}

	manager := NewCalendarManager(config)

	if manager == nil {
		t.Error("Expected non-nil calendar manager")
		return
	}

	if manager.config.HolidayCountryCode != "NO" {
		t.Errorf("Expected country code NO, got %s", manager.config.HolidayCountryCode)
	}

	if manager.holidayDetector == nil {
		t.Error("Expected non-nil holiday detector")
	}
}

func TestAddServicePattern(t *testing.T) {
	manager := NewCalendarManager(CalendarConfig{})

	pattern := &ServicePattern{
		ID:   "test_pattern",
		Name: "Test Pattern",
		Type: PatternRegular,
		ValidityPeriod: &ValidityPeriod{
			StartDate: time.Now(),
			EndDate:   time.Now().AddDate(1, 0, 0),
		},
		OperatingDays: []time.Weekday{time.Monday, time.Tuesday},
	}

	manager.AddServicePattern(pattern)

	retrieved := manager.GetServicePattern("test_pattern")
	if retrieved == nil {
		t.Error("Expected to retrieve added service pattern")
		return
	}

	if retrieved.ID != "test_pattern" {
		t.Errorf("Expected pattern ID test_pattern, got %s", retrieved.ID)
	}
}

func TestValidateServicePattern(t *testing.T) {
	manager := NewCalendarManager(CalendarConfig{})

	tests := []struct {
		name           string
		pattern        *ServicePattern
		expectedIssues int
	}{
		{
			name: "valid pattern",
			pattern: &ServicePattern{
				ID:   "valid",
				Name: "Valid Pattern",
				ValidityPeriod: &ValidityPeriod{
					StartDate: time.Now(),
					EndDate:   time.Now().AddDate(1, 0, 0),
				},
				OperatingDays: []time.Weekday{time.Monday},
			},
			expectedIssues: 0,
		},
		{
			name: "missing ID",
			pattern: &ServicePattern{
				Name: "No ID Pattern",
				ValidityPeriod: &ValidityPeriod{
					StartDate: time.Now(),
					EndDate:   time.Now().AddDate(1, 0, 0),
				},
			},
			expectedIssues: 2, // Missing ID and no operating days
		},
		{
			name: "invalid date range",
			pattern: &ServicePattern{
				ID:   "invalid_dates",
				Name: "Invalid Dates",
				ValidityPeriod: &ValidityPeriod{
					StartDate: time.Now().AddDate(1, 0, 0),
					EndDate:   time.Now(),
				},
				OperatingDays: []time.Weekday{time.Monday},
			},
			expectedIssues: 1, // Start date after end date
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues := manager.ValidateServicePattern(tt.pattern)
			if len(issues) != tt.expectedIssues {
				t.Errorf("Expected %d issues, got %d: %v", tt.expectedIssues, len(issues), issues)
			}
		})
	}
}

func TestGetEffectiveDates(t *testing.T) {
	manager := NewCalendarManager(CalendarConfig{})

	// Create a pattern that operates on weekdays
	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC) // Monday
	endDate := time.Date(2024, 1, 7, 0, 0, 0, 0, time.UTC)   // Sunday

	pattern := &ServicePattern{
		ID:   "weekday_pattern",
		Name: "Weekday Pattern",
		ValidityPeriod: &ValidityPeriod{
			StartDate: startDate,
			EndDate:   endDate,
		},
		OperatingDays: []time.Weekday{
			time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday,
		},
		Exceptions:      make([]*ServiceException, 0),
		SpecialDays:     make(map[string]*SpecialDay),
		HolidayBehavior: HolidayAsWeekday,
	}

	manager.AddServicePattern(pattern)

	dates, err := manager.GetEffectiveDates("weekday_pattern", startDate, endDate)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Should have 5 weekdays (Mon-Fri)
	expectedDates := 5
	if len(dates) != expectedDates {
		t.Errorf("Expected %d effective dates, got %d", expectedDates, len(dates))
	}

	// Check that all dates are weekdays
	for _, date := range dates {
		weekday := date.Weekday()
		if weekday == time.Saturday || weekday == time.Sunday {
			t.Errorf("Found weekend date %s in weekday pattern", date.Format("2006-01-02"))
		}
	}
}

func TestServiceOperatingWithExceptions(t *testing.T) {
	manager := NewCalendarManager(CalendarConfig{})

	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)
	exceptionDate := time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC) // Wednesday

	pattern := &ServicePattern{
		ID:   "exception_pattern",
		Name: "Exception Pattern",
		ValidityPeriod: &ValidityPeriod{
			StartDate: startDate,
			EndDate:   endDate,
		},
		OperatingDays: []time.Weekday{
			time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday,
		},
		Exceptions: []*ServiceException{
			{
				Date:   exceptionDate,
				Type:   ExceptionRemoved,
				Reason: "Holiday",
			},
		},
		SpecialDays:     make(map[string]*SpecialDay),
		HolidayBehavior: HolidayAsWeekday,
	}

	manager.AddServicePattern(pattern)

	// Check that the exception date is not operating
	isOperating := manager.isServiceOperating(pattern, exceptionDate)
	if isOperating {
		t.Error("Expected service not to be operating on exception date")
	}

	// Check that other weekdays are still operating
	normalDate := time.Date(2024, 1, 4, 0, 0, 0, 0, time.UTC) // Thursday
	isOperating = manager.isServiceOperating(pattern, normalDate)
	if !isOperating {
		t.Error("Expected service to be operating on normal weekday")
	}
}

func TestGenerateGTFSCalendar(t *testing.T) {
	manager := NewCalendarManager(CalendarConfig{})

	pattern := &ServicePattern{
		ID:   "gtfs_test",
		Name: "GTFS Test Pattern",
		BaseCalendar: &model.Calendar{
			ServiceID: "gtfs_test",
			Monday:    true,
			Tuesday:   true,
			StartDate: "20240101",
			EndDate:   "20241231",
		},
		ValidityPeriod: &ValidityPeriod{
			StartDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:   time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC),
		},
		OperatingDays: []time.Weekday{time.Monday, time.Tuesday},
		Exceptions:    make([]*ServiceException, 0),
		SpecialDays:   make(map[string]*SpecialDay),
	}

	manager.AddServicePattern(pattern)

	calendars, calendarDates, err := manager.GenerateGTFSCalendar(nil)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(calendars) == 0 {
		t.Error("Expected at least one calendar to be generated")
	}

	// Check that the calendar has correct service ID
	found := false
	for _, calendar := range calendars {
		if calendar.ServiceID == "gtfs_test" {
			found = true

			// Check operating days
			if !calendar.Monday {
				t.Error("Expected Monday to be operating")
			}
			if !calendar.Tuesday {
				t.Error("Expected Tuesday to be operating")
			}
			if calendar.Wednesday {
				t.Error("Expected Wednesday not to be operating")
			}

			break
		}
	}

	if !found {
		t.Error("Expected to find calendar with service ID 'gtfs_test'")
	}

	// Calendar dates might be empty for this simple pattern
	t.Logf("Generated %d calendars and %d calendar dates", len(calendars), len(calendarDates))
}

func TestOperatingPeriodProcessing(t *testing.T) {
	manager := NewCalendarManager(CalendarConfig{})

	basePattern := &ServicePattern{
		ID:   "base_pattern",
		Name: "Base Pattern",
		BaseCalendar: &model.Calendar{
			ServiceID: "base_pattern",
			Monday:    true,
			Tuesday:   true,
			Wednesday: true,
			Thursday:  true,
			Friday:    true,
			StartDate: "20240101",
			EndDate:   "20240630",
		},
		ValidityPeriod: &ValidityPeriod{
			StartDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:   time.Date(2024, 6, 30, 0, 0, 0, 0, time.UTC),
		},
		OperatingDays: []time.Weekday{time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday},
		Exceptions:    make([]*ServiceException, 0),
		SpecialDays:   make(map[string]*SpecialDay),
	}

	overridePattern := &ServicePattern{
		ID:   "override_pattern",
		Name: "Override Pattern",
		BaseCalendar: &model.Calendar{
			ServiceID: "override_pattern",
			Monday:    true,
			Tuesday:   true,
			Wednesday: true,
			Thursday:  true,
			Friday:    true,
			Saturday:  true,
			Sunday:    true,
			StartDate: "20240701",
			EndDate:   "20241231",
		},
		ValidityPeriod: &ValidityPeriod{
			StartDate: time.Date(2024, 7, 1, 0, 0, 0, 0, time.UTC),
			EndDate:   time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC),
		},
		OperatingDays: []time.Weekday{time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday, time.Saturday, time.Sunday},
		Exceptions:    make([]*ServiceException, 0),
		SpecialDays:   make(map[string]*SpecialDay),
	}

	period := &OperatingPeriod{
		ID:          "test_period",
		Name:        "Test Operating Period",
		StartDate:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:     time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC),
		BasePattern: basePattern,
		Overrides: map[string]*ServicePattern{
			"summer": overridePattern,
		},
		Priority: 1,
	}

	manager.AddOperatingPeriod(period)

	calendars, calendarDates, err := manager.GenerateGTFSCalendar(nil)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Should generate calendars for both base and override patterns
	expectedCalendars := 2 // base + override
	if len(calendars) < expectedCalendars {
		t.Errorf("Expected at least %d calendars, got %d", expectedCalendars, len(calendars))
	}

	t.Logf("Generated %d calendars and %d calendar dates for operating period", len(calendars), len(calendarDates))
}
