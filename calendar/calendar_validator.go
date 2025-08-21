package calendar

import (
	"fmt"
	"time"
)

// CalendarValidator validates calendar patterns and operating periods
type CalendarValidator struct {
	validationLevel ValidationLevel
}

// NewCalendarValidator creates a new calendar validator
func NewCalendarValidator(level ValidationLevel) *CalendarValidator {
	return &CalendarValidator{
		validationLevel: level,
	}
}

// ValidateServicePattern validates a service pattern for consistency and completeness
func (cv *CalendarValidator) ValidateServicePattern(pattern *ServicePattern) []string {
	issues := make([]string, 0)

	// Basic validation
	if pattern.ID == "" {
		issues = append(issues, "Service pattern missing ID")
	}

	if pattern.Name == "" && cv.validationLevel >= ValidationStandard {
		issues = append(issues, "Service pattern missing name")
	}

	// Validity period validation
	if pattern.ValidityPeriod == nil {
		issues = append(issues, "Service pattern missing validity period")
	} else {
		if pattern.ValidityPeriod.StartDate.IsZero() {
			issues = append(issues, "Service pattern validity period missing start date")
		}
		if pattern.ValidityPeriod.EndDate.IsZero() {
			issues = append(issues, "Service pattern validity period missing end date")
		}
		if pattern.ValidityPeriod.StartDate.After(pattern.ValidityPeriod.EndDate) {
			issues = append(issues, "Service pattern start date is after end date")
		}

		// Check if validity period is reasonable
		if cv.validationLevel >= ValidationStrict {
			duration := pattern.ValidityPeriod.EndDate.Sub(pattern.ValidityPeriod.StartDate)
			if duration > 5*365*24*time.Hour { // More than 5 years
				issues = append(issues, "Service pattern validity period is unusually long (>5 years)")
			}
			if duration < 24*time.Hour { // Less than 1 day
				issues = append(issues, "Service pattern validity period is very short (<1 day)")
			}
		}
	}

	// Operating days validation
	if len(pattern.OperatingDays) == 0 && len(pattern.Exceptions) == 0 {
		issues = append(issues, "Service pattern has no operating days or exceptions")
	}

	if cv.validationLevel >= ValidationStrict {
		// Check for conflicting operating days and non-operating days
		operatingDaysMap := make(map[time.Weekday]bool)
		for _, day := range pattern.OperatingDays {
			operatingDaysMap[day] = true
		}

		for _, day := range pattern.NonOperatingDays {
			if operatingDaysMap[day] {
				issues = append(issues, fmt.Sprintf("Day %s is both operating and non-operating", day.String()))
			}
		}
	}

	// Exceptions validation
	issues = append(issues, cv.validateExceptions(pattern.Exceptions)...)

	// Seasonal variations validation
	if cv.validationLevel >= ValidationStandard {
		issues = append(issues, cv.validateSeasonalVariations(pattern.SeasonalVariations)...)
	}

	// Special days validation
	if cv.validationLevel >= ValidationDetailed {
		issues = append(issues, cv.validateSpecialDays(pattern.SpecialDays)...)
	}

	return issues
}

// ValidateOperatingPeriod validates an operating period
func (cv *CalendarValidator) ValidateOperatingPeriod(period *OperatingPeriod) []string {
	issues := make([]string, 0)

	// Basic validation
	if period.ID == "" {
		issues = append(issues, "Operating period missing ID")
	}

	if period.Name == "" && cv.validationLevel >= ValidationStandard {
		issues = append(issues, "Operating period missing name")
	}

	// Date validation
	if period.StartDate.IsZero() {
		issues = append(issues, "Operating period missing start date")
	}

	if period.EndDate.IsZero() {
		issues = append(issues, "Operating period missing end date")
	}

	if period.StartDate.After(period.EndDate) {
		issues = append(issues, "Operating period start date is after end date")
	}

	// Pattern validation
	if period.BasePattern == nil && len(period.Overrides) == 0 {
		issues = append(issues, "Operating period has no base pattern or overrides")
	}

	// Validate base pattern if present
	if period.BasePattern != nil {
		baseIssues := cv.ValidateServicePattern(period.BasePattern)
		for _, issue := range baseIssues {
			issues = append(issues, fmt.Sprintf("Base pattern: %s", issue))
		}
	}

	// Validate override patterns
	if cv.validationLevel >= ValidationStandard {
		for overrideID, override := range period.Overrides {
			overrideIssues := cv.ValidateServicePattern(override)
			for _, issue := range overrideIssues {
				issues = append(issues, fmt.Sprintf("Override pattern %s: %s", overrideID, issue))
			}
		}
	}

	// Priority validation
	if cv.validationLevel >= ValidationStrict {
		if period.Priority < 0 {
			issues = append(issues, "Operating period priority cannot be negative")
		}
	}

	return issues
}

// ValidateSeasonalPattern validates a seasonal pattern
func (cv *CalendarValidator) ValidateSeasonalPattern(pattern *SeasonalPattern) []string {
	issues := make([]string, 0)

	if pattern.ID == "" {
		issues = append(issues, "Seasonal pattern missing ID")
	}

	if pattern.Name == "" && cv.validationLevel >= ValidationStandard {
		issues = append(issues, "Seasonal pattern missing name")
	}

	if len(pattern.Seasons) == 0 {
		issues = append(issues, "Seasonal pattern has no seasons defined")
	}

	// Validate season definitions
	for i, season := range pattern.Seasons {
		seasonIssues := cv.validateSeasonDefinition(season)
		for _, issue := range seasonIssues {
			issues = append(issues, fmt.Sprintf("Season %d: %s", i, issue))
		}
	}

	// Validate transitions
	if cv.validationLevel >= ValidationDetailed {
		for i, transition := range pattern.Transitions {
			transitionIssues := cv.validateSeasonTransition(transition)
			for _, issue := range transitionIssues {
				issues = append(issues, fmt.Sprintf("Transition %d: %s", i, issue))
			}
		}
	}

	return issues
}

// validateExceptions validates service exceptions
func (cv *CalendarValidator) validateExceptions(exceptions []*ServiceException) []string {
	issues := make([]string, 0)

	dateMap := make(map[string]*ServiceException)

	for i, exception := range exceptions {
		if exception.Date.IsZero() {
			issues = append(issues, fmt.Sprintf("Exception %d missing date", i))
			continue
		}

		dateStr := exception.Date.Format("2006-01-02")

		// Check for duplicate dates
		if existing, exists := dateMap[dateStr]; exists {
			if existing.Type != exception.Type {
				issues = append(issues, fmt.Sprintf("Conflicting exceptions for date %s", dateStr))
			} else if cv.validationLevel >= ValidationStandard {
				issues = append(issues, fmt.Sprintf("Duplicate exception for date %s", dateStr))
			}
		}
		dateMap[dateStr] = exception

		if exception.Reason == "" && cv.validationLevel >= ValidationStandard {
			issues = append(issues, fmt.Sprintf("Exception %d missing reason", i))
		}

		// Validate alternative service if present
		if exception.Alternative != nil && cv.validationLevel >= ValidationDetailed {
			if exception.Alternative.ServiceID == "" {
				issues = append(issues, fmt.Sprintf("Exception %d alternative missing service ID", i))
			}
		}
	}

	return issues
}

// validateSeasonalVariations validates seasonal variations
func (cv *CalendarValidator) validateSeasonalVariations(variations []*SeasonalVariation) []string {
	issues := make([]string, 0)

	for i, variation := range variations {
		if variation.StartDate.IsZero() {
			issues = append(issues, fmt.Sprintf("Seasonal variation %d missing start date", i))
		}

		if variation.EndDate.IsZero() {
			issues = append(issues, fmt.Sprintf("Seasonal variation %d missing end date", i))
		}

		if variation.StartDate.After(variation.EndDate) {
			issues = append(issues, fmt.Sprintf("Seasonal variation %d start date after end date", i))
		}

		if len(variation.Changes) == 0 {
			issues = append(issues, fmt.Sprintf("Seasonal variation %d has no changes defined", i))
		}

		// Validate changes
		for j, change := range variation.Changes {
			if change.Description == "" && cv.validationLevel >= ValidationStandard {
				issues = append(issues, fmt.Sprintf("Seasonal variation %d change %d missing description", i, j))
			}
		}
	}

	return issues
}

// validateSpecialDays validates special days
func (cv *CalendarValidator) validateSpecialDays(specialDays map[string]*SpecialDay) []string {
	issues := make([]string, 0)

	for dateStr, specialDay := range specialDays {
		if specialDay.Date.IsZero() {
			issues = append(issues, fmt.Sprintf("Special day %s has zero date", dateStr))
		}

		if specialDay.Name == "" && cv.validationLevel >= ValidationStandard {
			issues = append(issues, fmt.Sprintf("Special day %s missing name", dateStr))
		}

		// Validate date string matches actual date
		if specialDay.Date.Format("2006-01-02") != dateStr {
			issues = append(issues, fmt.Sprintf("Special day %s date mismatch", dateStr))
		}
	}

	return issues
}

// validateSeasonDefinition validates a season definition
func (cv *CalendarValidator) validateSeasonDefinition(season *SeasonDefinition) []string {
	issues := make([]string, 0)

	if season.StartMonth < time.January || season.StartMonth > time.December {
		issues = append(issues, "Invalid start month")
	}

	if season.EndMonth < time.January || season.EndMonth > time.December {
		issues = append(issues, "Invalid end month")
	}

	if season.StartDay < 1 || season.StartDay > 31 {
		issues = append(issues, "Invalid start day")
	}

	if season.EndDay < 1 || season.EndDay > 31 {
		issues = append(issues, "Invalid end day")
	}

	// Validate day ranges for specific months
	if cv.validationLevel >= ValidationStrict {
		if season.StartMonth == time.February && season.StartDay > 29 {
			issues = append(issues, "Start day too high for February")
		}

		if season.EndMonth == time.February && season.EndDay > 29 {
			issues = append(issues, "End day too high for February")
		}

		shortMonths := map[time.Month]bool{
			time.April: true, time.June: true, time.September: true, time.November: true,
		}

		if shortMonths[season.StartMonth] && season.StartDay > 30 {
			issues = append(issues, fmt.Sprintf("Start day too high for %s", season.StartMonth.String()))
		}

		if shortMonths[season.EndMonth] && season.EndDay > 30 {
			issues = append(issues, fmt.Sprintf("End day too high for %s", season.EndMonth.String()))
		}
	}

	return issues
}

// validateSeasonTransition validates a season transition
func (cv *CalendarValidator) validateSeasonTransition(transition *SeasonTransition) []string {
	issues := make([]string, 0)

	if transition.Duration < 0 {
		issues = append(issues, "Transition duration cannot be negative")
	}

	if transition.Duration > 30*24*time.Hour && cv.validationLevel >= ValidationStandard {
		issues = append(issues, "Transition duration is unusually long (>30 days)")
	}

	return issues
}

// ValidateCalendarConsistency validates consistency across multiple patterns and periods
func (cv *CalendarValidator) ValidateCalendarConsistency(patterns []*ServicePattern, periods []*OperatingPeriod) []string {
	issues := make([]string, 0)

	if cv.validationLevel < ValidationDetailed {
		return issues
	}

	// Check for overlapping operating periods
	for i, period1 := range periods {
		for j, period2 := range periods {
			if i >= j {
				continue
			}

			if cv.periodsOverlap(period1, period2) {
				issues = append(issues, fmt.Sprintf("Operating periods %s and %s overlap", period1.ID, period2.ID))
			}
		}
	}

	// Check for unreferenced service patterns
	referencedPatterns := make(map[string]bool)
	for _, period := range periods {
		if period.BasePattern != nil {
			referencedPatterns[period.BasePattern.ID] = true
		}
		for _, override := range period.Overrides {
			referencedPatterns[override.ID] = true
		}
	}

	for _, pattern := range patterns {
		if !referencedPatterns[pattern.ID] {
			issues = append(issues, fmt.Sprintf("Service pattern %s is not referenced by any operating period", pattern.ID))
		}
	}

	return issues
}

// periodsOverlap checks if two operating periods overlap in time
func (cv *CalendarValidator) periodsOverlap(period1, period2 *OperatingPeriod) bool {
	return period1.StartDate.Before(period2.EndDate) && period2.StartDate.Before(period1.EndDate)
}

// ValidateAgainstGTFSRules validates calendar data against GTFS specification rules
func (cv *CalendarValidator) ValidateAgainstGTFSRules(patterns []*ServicePattern) []string {
	issues := make([]string, 0)

	for _, pattern := range patterns {
		// GTFS rule: service_id must be unique
		// (This would typically be checked at a higher level)

		// GTFS rule: At least one day of the week must be set to 1
		if len(pattern.OperatingDays) == 0 && len(pattern.Exceptions) == 0 {
			issues = append(issues, fmt.Sprintf("Pattern %s has no operating days or exceptions (violates GTFS)", pattern.ID))
		}

		// GTFS rule: start_date and end_date must be in YYYYMMDD format
		if pattern.ValidityPeriod != nil {
			startStr := pattern.ValidityPeriod.StartDate.Format("20060102")
			endStr := pattern.ValidityPeriod.EndDate.Format("20060102")

			if len(startStr) != 8 || len(endStr) != 8 {
				issues = append(issues, fmt.Sprintf("Pattern %s date format issue", pattern.ID))
			}
		}

		// GTFS rule: exception_type must be 1 or 2
		for _, exception := range pattern.Exceptions {
			if exception.Type != ExceptionAdded && exception.Type != ExceptionRemoved {
				issues = append(issues, fmt.Sprintf("Pattern %s has invalid exception type", pattern.ID))
			}
		}
	}

	return issues
}

// SetValidationLevel updates the validation level
func (cv *CalendarValidator) SetValidationLevel(level ValidationLevel) {
	cv.validationLevel = level
}
