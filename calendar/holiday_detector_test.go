package calendar

import (
	"fmt"
	"testing"
	"time"
)

func TestNewHolidayDetector(t *testing.T) {
	detector := NewHolidayDetector("NO")

	if detector.countryCode != "NO" {
		t.Errorf("Expected country code NO, got %s", detector.countryCode)
	}

	if detector.customHolidays == nil {
		t.Error("Expected non-nil custom holidays map")
	}
}

func TestCalculateEaster(t *testing.T) {
	detector := NewHolidayDetector("NO")

	tests := []struct {
		year     int
		expected time.Time
	}{
		{2024, time.Date(2024, time.March, 31, 0, 0, 0, 0, time.UTC)},
		{2025, time.Date(2025, time.April, 20, 0, 0, 0, 0, time.UTC)},
		{2026, time.Date(2026, time.April, 5, 0, 0, 0, 0, time.UTC)},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("Easter_%d", tt.year), func(t *testing.T) {
			easter, err := detector.calculateEaster(tt.year)
			if err != nil {
				t.Errorf("Unexpected error calculating Easter for %d: %v", tt.year, err)
			}

			if !easter.Equal(tt.expected) {
				t.Errorf("Expected Easter %s for year %d, got %s",
					tt.expected.Format("2006-01-02"), tt.year, easter.Format("2006-01-02"))
			}
		})
	}
}

func TestGetHolidaysNorway(t *testing.T) {
	detector := NewHolidayDetector("NO")
	year := 2024

	holidays, err := detector.GetHolidays(year)
	if err != nil {
		t.Errorf("Unexpected error getting holidays: %v", err)
	}

	if len(holidays) == 0 {
		t.Error("Expected some holidays to be returned")
	}

	// Check for specific Norwegian holidays
	expectedHolidays := map[string]time.Time{
		"New Year's Day":   time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC),
		"Labour Day":       time.Date(2024, time.May, 1, 0, 0, 0, 0, time.UTC),
		"Constitution Day": time.Date(2024, time.May, 17, 0, 0, 0, 0, time.UTC),
		"Christmas Day":    time.Date(2024, time.December, 25, 0, 0, 0, 0, time.UTC),
		"Boxing Day":       time.Date(2024, time.December, 26, 0, 0, 0, 0, time.UTC),
	}

	foundHolidays := make(map[string]bool)
	for _, holiday := range holidays {
		foundHolidays[holiday.Name] = true

		if expectedDate, exists := expectedHolidays[holiday.Name]; exists {
			if !holiday.Date.Equal(expectedDate) {
				t.Errorf("Holiday %s: expected %s, got %s",
					holiday.Name, expectedDate.Format("2006-01-02"), holiday.Date.Format("2006-01-02"))
			}
		}
	}

	for holidayName := range expectedHolidays {
		if !foundHolidays[holidayName] {
			t.Errorf("Expected to find holiday: %s", holidayName)
		}
	}
}

func TestGetHolidaysSweden(t *testing.T) {
	detector := NewHolidayDetector("SE")
	year := 2024

	holidays, err := detector.GetHolidays(year)
	if err != nil {
		t.Errorf("Unexpected error getting holidays: %v", err)
	}

	// Check for specific Swedish holidays
	expectedHolidays := []string{
		"New Year's Day",
		"Epiphany",
		"Labour Day",
		"National Day",
		"Midsummer Eve",
		"Midsummer Day",
		"All Saints' Day",
		"Christmas Day",
		"Boxing Day",
	}

	foundHolidays := make(map[string]bool)
	for _, holiday := range holidays {
		foundHolidays[holiday.Name] = true
	}

	for _, expectedHoliday := range expectedHolidays {
		if !foundHolidays[expectedHoliday] {
			t.Errorf("Expected to find Swedish holiday: %s", expectedHoliday)
		}
	}
}

func TestEasterBasedHolidays(t *testing.T) {
	detector := NewHolidayDetector("NO")
	year := 2024
	easter := time.Date(2024, time.March, 31, 0, 0, 0, 0, time.UTC)

	holidays, err := detector.getEasterBasedHolidays(year)
	if err != nil {
		t.Errorf("Unexpected error getting Easter-based holidays: %v", err)
	}

	expectedHolidays := map[string]time.Time{
		"Maundy Thursday": easter.AddDate(0, 0, -3),
		"Good Friday":     easter.AddDate(0, 0, -2),
		"Easter Sunday":   easter,
		"Easter Monday":   easter.AddDate(0, 0, 1),
		"Ascension Day":   easter.AddDate(0, 0, 39),
		"Whit Sunday":     easter.AddDate(0, 0, 49),
		"Whit Monday":     easter.AddDate(0, 0, 50),
	}

	foundHolidays := make(map[string]bool)
	for _, holiday := range holidays {
		foundHolidays[holiday.Name] = true

		if expectedDate, exists := expectedHolidays[holiday.Name]; exists {
			if !holiday.Date.Equal(expectedDate) {
				t.Errorf("Holiday %s: expected %s, got %s",
					holiday.Name, expectedDate.Format("2006-01-02"), holiday.Date.Format("2006-01-02"))
			}
		}
	}

	for holidayName := range expectedHolidays {
		if !foundHolidays[holidayName] {
			t.Errorf("Expected to find Easter-based holiday: %s", holidayName)
		}
	}
}

func TestAddCustomHoliday(t *testing.T) {
	detector := NewHolidayDetector("NO")

	customHoliday := &Holiday{
		Date:       time.Date(2024, time.June, 15, 0, 0, 0, 0, time.UTC),
		Name:       "Custom Holiday",
		Type:       HolidayPublic,
		IsNational: false,
		IsRegional: true,
		Regions:    []string{"Oslo"},
		Observance: ObservanceActual,
	}

	detector.AddCustomHoliday(customHoliday)

	holidays, err := detector.GetHolidays(2024)
	if err != nil {
		t.Errorf("Unexpected error getting holidays: %v", err)
	}

	found := false
	for _, holiday := range holidays {
		if holiday.Name == "Custom Holiday" {
			found = true
			if !holiday.Date.Equal(customHoliday.Date) {
				t.Errorf("Custom holiday date mismatch: expected %s, got %s",
					customHoliday.Date.Format("2006-01-02"), holiday.Date.Format("2006-01-02"))
			}
			break
		}
	}

	if !found {
		t.Error("Expected to find custom holiday in results")
	}
}

func TestIsHoliday(t *testing.T) {
	detector := NewHolidayDetector("NO")

	// Test a known holiday
	christmasDate := time.Date(2024, time.December, 25, 0, 0, 0, 0, time.UTC)
	holiday, isHoliday := detector.IsHoliday(christmasDate)

	if !isHoliday {
		t.Error("Expected Christmas Day to be identified as a holiday")
	}

	if holiday == nil || holiday.Name != "Christmas Day" {
		t.Error("Expected Christmas Day holiday to be returned")
	}

	// Test a non-holiday
	regularDate := time.Date(2024, time.June, 15, 0, 0, 0, 0, time.UTC)
	_, isHoliday = detector.IsHoliday(regularDate)

	if isHoliday {
		t.Error("Expected regular date not to be identified as a holiday")
	}
}

func TestGetHolidaysInRange(t *testing.T) {
	detector := NewHolidayDetector("NO")

	startDate := time.Date(2024, time.December, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, time.December, 31, 0, 0, 0, 0, time.UTC)

	holidays, err := detector.GetHolidaysInRange(startDate, endDate)
	if err != nil {
		t.Errorf("Unexpected error getting holidays in range: %v", err)
	}

	// Should find Christmas and Boxing Day in December
	expectedHolidays := []string{"Christmas Day", "Boxing Day"}
	foundHolidays := make(map[string]bool)

	for _, holiday := range holidays {
		foundHolidays[holiday.Name] = true

		// Verify all holidays are within the range
		if holiday.Date.Before(startDate) || holiday.Date.After(endDate) {
			t.Errorf("Holiday %s (%s) is outside the requested range",
				holiday.Name, holiday.Date.Format("2006-01-02"))
		}
	}

	for _, expectedHoliday := range expectedHolidays {
		if !foundHolidays[expectedHoliday] {
			t.Errorf("Expected to find holiday %s in December range", expectedHoliday)
		}
	}
}

func TestObservanceRules(t *testing.T) {
	detector := NewHolidayDetector("US") // Use US for observance rule testing

	tests := []struct {
		name         string
		actualDate   time.Time
		observance   Observance
		expectedDate time.Time
	}{
		{
			name:         "Saturday to Monday",
			actualDate:   time.Date(2024, time.June, 15, 0, 0, 0, 0, time.UTC), // Saturday
			observance:   ObservanceMonday,
			expectedDate: time.Date(2024, time.June, 17, 0, 0, 0, 0, time.UTC), // Monday
		},
		{
			name:         "Sunday to Monday",
			actualDate:   time.Date(2024, time.June, 16, 0, 0, 0, 0, time.UTC), // Sunday
			observance:   ObservanceMonday,
			expectedDate: time.Date(2024, time.June, 17, 0, 0, 0, 0, time.UTC), // Monday
		},
		{
			name:         "Saturday to Friday",
			actualDate:   time.Date(2024, time.June, 15, 0, 0, 0, 0, time.UTC), // Saturday
			observance:   ObservanceFriday,
			expectedDate: time.Date(2024, time.June, 14, 0, 0, 0, 0, time.UTC), // Friday
		},
		{
			name:         "Weekday no change",
			actualDate:   time.Date(2024, time.June, 17, 0, 0, 0, 0, time.UTC), // Monday
			observance:   ObservanceMonday,
			expectedDate: time.Date(2024, time.June, 17, 0, 0, 0, 0, time.UTC), // Monday
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			observedDate := detector.getObservedDate(tt.actualDate, tt.observance)

			if !observedDate.Equal(tt.expectedDate) {
				t.Errorf("Expected observed date %s, got %s",
					tt.expectedDate.Format("2006-01-02"), observedDate.Format("2006-01-02"))
			}
		})
	}
}

func TestHolidayTypes(t *testing.T) {
	types := []HolidayType{
		HolidayPublic,
		HolidayReligious,
		HolidayCultural,
		HolidayCommercial,
		HolidaySchool,
		HolidayBank,
	}

	expectedStrings := []string{
		"Public",
		"Religious",
		"Cultural",
		"Commercial",
		"School",
		"Bank",
	}

	for i, holidayType := range types {
		if holidayType.String() != expectedStrings[i] {
			t.Errorf("Expected %s, got %s", expectedStrings[i], holidayType.String())
		}
	}
}
