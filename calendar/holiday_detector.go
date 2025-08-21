package calendar

import (
	"fmt"
	"time"
)

// HolidayDetector detects holidays for European countries
type HolidayDetector struct {
	countryCode    string
	customHolidays map[string]*Holiday
}

// Holiday represents a holiday with its characteristics
type Holiday struct {
	Date       time.Time   `json:"date"`
	Name       string      `json:"name"`
	Type       HolidayType `json:"type"`
	IsNational bool        `json:"is_national"`
	IsRegional bool        `json:"is_regional"`
	Regions    []string    `json:"regions,omitempty"`
	Observance Observance  `json:"observance"`
}

// HolidayType defines different types of holidays
type HolidayType int

const (
	HolidayPublic HolidayType = iota
	HolidayReligious
	HolidayCultural
	HolidayCommercial
	HolidaySchool
	HolidayBank
)

func (t HolidayType) String() string {
	switch t {
	case HolidayPublic:
		return "Public"
	case HolidayReligious:
		return "Religious"
	case HolidayCultural:
		return "Cultural"
	case HolidayCommercial:
		return "Commercial"
	case HolidaySchool:
		return "School"
	case HolidayBank:
		return "Bank"
	default:
		return "Unknown"
	}
}

// Observance defines how a holiday is observed
type Observance int

const (
	ObservanceActual Observance = iota
	ObservanceMonday
	ObservanceFriday
	ObservanceNearest
)

func (o Observance) String() string {
	switch o {
	case ObservanceActual:
		return "Actual"
	case ObservanceMonday:
		return "Monday"
	case ObservanceFriday:
		return "Friday"
	case ObservanceNearest:
		return "Nearest"
	default:
		return "Unknown"
	}
}

// NewHolidayDetector creates a new holiday detector for a specific country
func NewHolidayDetector(countryCode string) *HolidayDetector {
	return &HolidayDetector{
		countryCode:    countryCode,
		customHolidays: make(map[string]*Holiday),
	}
}

// AddCustomHoliday adds a custom holiday
func (hd *HolidayDetector) AddCustomHoliday(holiday *Holiday) {
	key := fmt.Sprintf("%s_%s", holiday.Date.Format("2006-01-02"), holiday.Name)
	hd.customHolidays[key] = holiday
}

// GetHolidays returns all holidays for a given year
func (hd *HolidayDetector) GetHolidays(year int) ([]*Holiday, error) {
	holidays := make([]*Holiday, 0)

	// Get fixed holidays
	fixedHolidays := hd.getFixedHolidays(year)
	holidays = append(holidays, fixedHolidays...)

	// Get Easter-based holidays
	easterHolidays, err := hd.getEasterBasedHolidays(year)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate Easter-based holidays: %v", err)
	}
	holidays = append(holidays, easterHolidays...)

	// Get variable holidays
	variableHolidays := hd.getVariableHolidays(year)
	holidays = append(holidays, variableHolidays...)

	// Add custom holidays
	for _, holiday := range hd.customHolidays {
		if holiday.Date.Year() == year {
			holidays = append(holidays, holiday)
		}
	}

	// Apply observance rules
	holidays = hd.applyObservanceRules(holidays)

	return holidays, nil
}

// getFixedHolidays returns holidays that occur on the same date each year
func (hd *HolidayDetector) getFixedHolidays(year int) []*Holiday {
	holidays := make([]*Holiday, 0)

	switch hd.countryCode {
	case "NO": // Norway
		holidays = append(holidays, &Holiday{
			Date:       time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC),
			Name:       "New Year's Day",
			Type:       HolidayPublic,
			IsNational: true,
			Observance: ObservanceActual,
		})
		holidays = append(holidays, &Holiday{
			Date:       time.Date(year, time.May, 1, 0, 0, 0, 0, time.UTC),
			Name:       "Labour Day",
			Type:       HolidayPublic,
			IsNational: true,
			Observance: ObservanceActual,
		})
		holidays = append(holidays, &Holiday{
			Date:       time.Date(year, time.May, 17, 0, 0, 0, 0, time.UTC),
			Name:       "Constitution Day",
			Type:       HolidayPublic,
			IsNational: true,
			Observance: ObservanceActual,
		})
		holidays = append(holidays, &Holiday{
			Date:       time.Date(year, time.December, 25, 0, 0, 0, 0, time.UTC),
			Name:       "Christmas Day",
			Type:       HolidayReligious,
			IsNational: true,
			Observance: ObservanceActual,
		})
		holidays = append(holidays, &Holiday{
			Date:       time.Date(year, time.December, 26, 0, 0, 0, 0, time.UTC),
			Name:       "Boxing Day",
			Type:       HolidayReligious,
			IsNational: true,
			Observance: ObservanceActual,
		})

	case "SE": // Sweden
		holidays = append(holidays, &Holiday{
			Date:       time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC),
			Name:       "New Year's Day",
			Type:       HolidayPublic,
			IsNational: true,
			Observance: ObservanceActual,
		})
		holidays = append(holidays, &Holiday{
			Date:       time.Date(year, time.January, 6, 0, 0, 0, 0, time.UTC),
			Name:       "Epiphany",
			Type:       HolidayReligious,
			IsNational: true,
			Observance: ObservanceActual,
		})
		holidays = append(holidays, &Holiday{
			Date:       time.Date(year, time.May, 1, 0, 0, 0, 0, time.UTC),
			Name:       "Labour Day",
			Type:       HolidayPublic,
			IsNational: true,
			Observance: ObservanceActual,
		})
		holidays = append(holidays, &Holiday{
			Date:       time.Date(year, time.June, 6, 0, 0, 0, 0, time.UTC),
			Name:       "National Day",
			Type:       HolidayPublic,
			IsNational: true,
			Observance: ObservanceActual,
		})
		holidays = append(holidays, &Holiday{
			Date:       time.Date(year, time.December, 25, 0, 0, 0, 0, time.UTC),
			Name:       "Christmas Day",
			Type:       HolidayReligious,
			IsNational: true,
			Observance: ObservanceActual,
		})
		holidays = append(holidays, &Holiday{
			Date:       time.Date(year, time.December, 26, 0, 0, 0, 0, time.UTC),
			Name:       "Boxing Day",
			Type:       HolidayReligious,
			IsNational: true,
			Observance: ObservanceActual,
		})

	case "DK": // Denmark
		holidays = append(holidays, &Holiday{
			Date:       time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC),
			Name:       "New Year's Day",
			Type:       HolidayPublic,
			IsNational: true,
			Observance: ObservanceActual,
		})
		holidays = append(holidays, &Holiday{
			Date:       time.Date(year, time.December, 25, 0, 0, 0, 0, time.UTC),
			Name:       "Christmas Day",
			Type:       HolidayReligious,
			IsNational: true,
			Observance: ObservanceActual,
		})
		holidays = append(holidays, &Holiday{
			Date:       time.Date(year, time.December, 26, 0, 0, 0, 0, time.UTC),
			Name:       "Boxing Day",
			Type:       HolidayReligious,
			IsNational: true,
			Observance: ObservanceActual,
		})

	case "FI": // Finland
		holidays = append(holidays, &Holiday{
			Date:       time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC),
			Name:       "New Year's Day",
			Type:       HolidayPublic,
			IsNational: true,
			Observance: ObservanceActual,
		})
		holidays = append(holidays, &Holiday{
			Date:       time.Date(year, time.January, 6, 0, 0, 0, 0, time.UTC),
			Name:       "Epiphany",
			Type:       HolidayReligious,
			IsNational: true,
			Observance: ObservanceActual,
		})
		holidays = append(holidays, &Holiday{
			Date:       time.Date(year, time.May, 1, 0, 0, 0, 0, time.UTC),
			Name:       "May Day",
			Type:       HolidayPublic,
			IsNational: true,
			Observance: ObservanceActual,
		})
		holidays = append(holidays, &Holiday{
			Date:       time.Date(year, time.December, 6, 0, 0, 0, 0, time.UTC),
			Name:       "Independence Day",
			Type:       HolidayPublic,
			IsNational: true,
			Observance: ObservanceActual,
		})
		holidays = append(holidays, &Holiday{
			Date:       time.Date(year, time.December, 25, 0, 0, 0, 0, time.UTC),
			Name:       "Christmas Day",
			Type:       HolidayReligious,
			IsNational: true,
			Observance: ObservanceActual,
		})
		holidays = append(holidays, &Holiday{
			Date:       time.Date(year, time.December, 26, 0, 0, 0, 0, time.UTC),
			Name:       "Boxing Day",
			Type:       HolidayReligious,
			IsNational: true,
			Observance: ObservanceActual,
		})

	default:
		// Generic European holidays
		holidays = append(holidays, &Holiday{
			Date:       time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC),
			Name:       "New Year's Day",
			Type:       HolidayPublic,
			IsNational: true,
			Observance: ObservanceActual,
		})
		holidays = append(holidays, &Holiday{
			Date:       time.Date(year, time.December, 25, 0, 0, 0, 0, time.UTC),
			Name:       "Christmas Day",
			Type:       HolidayReligious,
			IsNational: true,
			Observance: ObservanceActual,
		})
	}

	return holidays
}

// getEasterBasedHolidays returns holidays that are calculated relative to Easter
func (hd *HolidayDetector) getEasterBasedHolidays(year int) ([]*Holiday, error) {
	holidays := make([]*Holiday, 0)

	easter, err := hd.calculateEaster(year)
	if err != nil {
		return nil, err
	}

	// Common Easter-based holidays across European countries
	holidays = append(holidays, &Holiday{
		Date:       easter.AddDate(0, 0, -3), // Maundy Thursday
		Name:       "Maundy Thursday",
		Type:       HolidayReligious,
		IsNational: true,
		Observance: ObservanceActual,
	})

	holidays = append(holidays, &Holiday{
		Date:       easter.AddDate(0, 0, -2), // Good Friday
		Name:       "Good Friday",
		Type:       HolidayReligious,
		IsNational: true,
		Observance: ObservanceActual,
	})

	holidays = append(holidays, &Holiday{
		Date:       easter, // Easter Sunday
		Name:       "Easter Sunday",
		Type:       HolidayReligious,
		IsNational: true,
		Observance: ObservanceActual,
	})

	holidays = append(holidays, &Holiday{
		Date:       easter.AddDate(0, 0, 1), // Easter Monday
		Name:       "Easter Monday",
		Type:       HolidayReligious,
		IsNational: true,
		Observance: ObservanceActual,
	})

	holidays = append(holidays, &Holiday{
		Date:       easter.AddDate(0, 0, 39), // Ascension Day
		Name:       "Ascension Day",
		Type:       HolidayReligious,
		IsNational: true,
		Observance: ObservanceActual,
	})

	holidays = append(holidays, &Holiday{
		Date:       easter.AddDate(0, 0, 49), // Whit Sunday
		Name:       "Whit Sunday",
		Type:       HolidayReligious,
		IsNational: true,
		Observance: ObservanceActual,
	})

	holidays = append(holidays, &Holiday{
		Date:       easter.AddDate(0, 0, 50), // Whit Monday
		Name:       "Whit Monday",
		Type:       HolidayReligious,
		IsNational: true,
		Observance: ObservanceActual,
	})

	return holidays, nil
}

// getVariableHolidays returns holidays that vary each year based on other rules
func (hd *HolidayDetector) getVariableHolidays(year int) []*Holiday {
	holidays := make([]*Holiday, 0)

	switch hd.countryCode {
	case "SE": // Sweden
		// Midsummer Eve (Friday between June 19-25)
		midsummerEve := hd.findFirstFridayInRange(year, time.June, 19, 25)
		holidays = append(holidays, &Holiday{
			Date:       midsummerEve,
			Name:       "Midsummer Eve",
			Type:       HolidayCultural,
			IsNational: true,
			Observance: ObservanceActual,
		})

		// Midsummer Day (Saturday between June 20-26)
		midsummerDay := midsummerEve.AddDate(0, 0, 1)
		holidays = append(holidays, &Holiday{
			Date:       midsummerDay,
			Name:       "Midsummer Day",
			Type:       HolidayCultural,
			IsNational: true,
			Observance: ObservanceActual,
		})

		// All Saints' Day (Saturday between October 31 - November 6)
		allSaints := hd.findFirstSaturdayInRange(year, time.October, 31, time.November, 6)
		holidays = append(holidays, &Holiday{
			Date:       allSaints,
			Name:       "All Saints' Day",
			Type:       HolidayReligious,
			IsNational: true,
			Observance: ObservanceActual,
		})

	case "NO": // Norway
		// Some variable Norwegian holidays could be added here
		break

	case "DK": // Denmark
		// Great Prayer Day (4th Friday after Easter)
		easter, _ := hd.calculateEaster(year)
		greatPrayerDay := easter.AddDate(0, 0, 26) // 4th Friday after Easter
		holidays = append(holidays, &Holiday{
			Date:       greatPrayerDay,
			Name:       "Great Prayer Day",
			Type:       HolidayReligious,
			IsNational: true,
			Observance: ObservanceActual,
		})
	}

	return holidays
}

// calculateEaster calculates Easter Sunday for a given year using the Western algorithm
func (hd *HolidayDetector) calculateEaster(year int) (time.Time, error) {
	// Using the algorithm for calculating Easter (Gregorian calendar)
	a := year % 19
	b := year / 100
	c := year % 100
	d := b / 4
	e := b % 4
	f := (b + 8) / 25
	g := (b - f + 1) / 3
	h := (19*a + b - d - g + 15) % 30
	i := c / 4
	k := c % 4
	l := (32 + 2*e + 2*i - h - k) % 7
	m := (a + 11*h + 22*l) / 451
	month := (h + l - 7*m + 114) / 31
	day := ((h + l - 7*m + 114) % 31) + 1

	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC), nil
}

// findFirstFridayInRange finds the first Friday within a date range in a single month
func (hd *HolidayDetector) findFirstFridayInRange(year int, month time.Month, startDay, endDay int) time.Time {
	for day := startDay; day <= endDay; day++ {
		date := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
		if date.Weekday() == time.Friday {
			return date
		}
	}
	// Fallback to the last possible day if no Friday found
	return time.Date(year, month, endDay, 0, 0, 0, 0, time.UTC)
}

// findFirstSaturdayInRange finds the first Saturday within a date range that may span two months
func (hd *HolidayDetector) findFirstSaturdayInRange(year int, startMonth time.Month, startDay int, endMonth time.Month, endDay int) time.Time {
	// Start from the start date
	current := time.Date(year, startMonth, startDay, 0, 0, 0, 0, time.UTC)
	end := time.Date(year, endMonth, endDay, 0, 0, 0, 0, time.UTC)

	for current.Before(end) || current.Equal(end) {
		if current.Weekday() == time.Saturday {
			return current
		}
		current = current.AddDate(0, 0, 1)
	}

	// Fallback to the end date if no Saturday found
	return end
}

// applyObservanceRules applies observance rules to holidays
func (hd *HolidayDetector) applyObservanceRules(holidays []*Holiday) []*Holiday {
	result := make([]*Holiday, 0, len(holidays))

	for _, holiday := range holidays {
		observedDate := hd.getObservedDate(holiday.Date, holiday.Observance)

		if !observedDate.Equal(holiday.Date) {
			// Create a new holiday for the observed date
			observedHoliday := &Holiday{
				Date:       observedDate,
				Name:       holiday.Name + " (Observed)",
				Type:       holiday.Type,
				IsNational: holiday.IsNational,
				IsRegional: holiday.IsRegional,
				Regions:    holiday.Regions,
				Observance: holiday.Observance,
			}
			result = append(result, observedHoliday)
		} else {
			result = append(result, holiday)
		}
	}

	return result
}

// getObservedDate calculates the observed date based on observance rules
func (hd *HolidayDetector) getObservedDate(actualDate time.Time, observance Observance) time.Time {
	switch observance {
	case ObservanceMonday:
		// If holiday falls on weekend, observe on Monday
		if actualDate.Weekday() == time.Saturday {
			return actualDate.AddDate(0, 0, 2)
		} else if actualDate.Weekday() == time.Sunday {
			return actualDate.AddDate(0, 0, 1)
		}
	case ObservanceFriday:
		// If holiday falls on weekend, observe on Friday
		if actualDate.Weekday() == time.Saturday {
			return actualDate.AddDate(0, 0, -1)
		} else if actualDate.Weekday() == time.Sunday {
			return actualDate.AddDate(0, 0, -2)
		}
	case ObservanceNearest:
		// Observe on nearest weekday
		if actualDate.Weekday() == time.Saturday {
			return actualDate.AddDate(0, 0, -1) // Friday
		} else if actualDate.Weekday() == time.Sunday {
			return actualDate.AddDate(0, 0, 1) // Monday
		}
	}
	return actualDate
}

// IsHoliday checks if a given date is a holiday
func (hd *HolidayDetector) IsHoliday(date time.Time) (*Holiday, bool) {
	holidays, err := hd.GetHolidays(date.Year())
	if err != nil {
		return nil, false
	}

	for _, holiday := range holidays {
		if holiday.Date.Equal(date) {
			return holiday, true
		}
	}

	return nil, false
}

// GetHolidaysInRange returns all holidays within a date range
func (hd *HolidayDetector) GetHolidaysInRange(startDate, endDate time.Time) ([]*Holiday, error) {
	allHolidays := make([]*Holiday, 0)

	for year := startDate.Year(); year <= endDate.Year(); year++ {
		yearHolidays, err := hd.GetHolidays(year)
		if err != nil {
			return nil, fmt.Errorf("failed to get holidays for year %d: %v", year, err)
		}

		for _, holiday := range yearHolidays {
			if (holiday.Date.After(startDate) || holiday.Date.Equal(startDate)) &&
				(holiday.Date.Before(endDate) || holiday.Date.Equal(endDate)) {
				allHolidays = append(allHolidays, holiday)
			}
		}
	}

	return allHolidays, nil
}
