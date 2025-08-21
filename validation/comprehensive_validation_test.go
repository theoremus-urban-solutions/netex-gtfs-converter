package validation

import (
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/theoremus-urban-solutions/netex-gtfs-converter/model"
)

// Comprehensive tests for Validator - Additional Coverage

func TestValidator_AddIssue_Comprehensive(t *testing.T) {
	validator := NewValidator()

	t.Run("Issue deduplication by type", func(t *testing.T) {
		validator.Reset()

		// Configure low max issues to test limit
		config := validator.config
		config.MaxIssuesPerType = 3
		validator.SetConfig(config)

		// Add more issues than the limit
		for i := 0; i < 5; i++ {
			validator.AddIssue(ValidationIssue{
				Severity:   SeverityError,
				Code:       "TEST_DUPLICATE",
				Message:    fmt.Sprintf("Duplicate issue %d", i),
				EntityType: "TestEntity",
				EntityID:   fmt.Sprintf("entity_%d", i),
			})
		}

		report := validator.GetReport()
		if len(report.Issues) != 3 {
			t.Errorf("Expected 3 issues due to MaxIssuesPerType limit, got %d", len(report.Issues))
		}
	})

	t.Run("Severity threshold filtering", func(t *testing.T) {
		validator.Reset()

		config := validator.config
		config.SeverityThreshold = SeverityWarning
		validator.SetConfig(config)

		// Add issues below threshold
		validator.AddIssue(ValidationIssue{
			Severity:   SeverityInfo,
			Code:       "TEST_INFO",
			Message:    "Info message should be filtered",
			EntityType: "TestEntity",
		})

		// Add issues at/above threshold
		validator.AddIssue(ValidationIssue{
			Severity:   SeverityWarning,
			Code:       "TEST_WARNING",
			Message:    "Warning message should be included",
			EntityType: "TestEntity",
		})

		validator.AddIssue(ValidationIssue{
			Severity:   SeverityError,
			Code:       "TEST_ERROR",
			Message:    "Error message should be included",
			EntityType: "TestEntity",
		})

		report := validator.GetReport()
		if len(report.Issues) != 2 {
			t.Errorf("Expected 2 issues above threshold, got %d", len(report.Issues))
		}

		for _, issue := range report.Issues {
			if issue.Severity < SeverityWarning {
				t.Errorf("Issue below threshold was included: %s", issue.Code)
			}
		}
	})

	t.Run("Issue context and metadata", func(t *testing.T) {
		validator.Reset()

		issue := ValidationIssue{
			Severity:   SeverityError,
			Code:       "TEST_CONTEXT",
			Message:    "Test issue with context",
			EntityType: "TestEntity",
			EntityID:   "test_entity_1",
			Field:      "test_field",
			Value:      "test_value",
			Suggestion: "This is a test suggestion",
			Location:   "test_location",
			Context: map[string]string{
				"context_key_1": "context_value_1",
				"context_key_2": "context_value_2",
			},
		}

		validator.AddIssue(issue)
		report := validator.GetReport()

		if len(report.Issues) != 1 {
			t.Fatalf("Expected 1 issue, got %d", len(report.Issues))
		}

		foundIssue := report.Issues[0]
		if foundIssue.EntityID != issue.EntityID {
			t.Errorf("Expected EntityID %s, got %s", issue.EntityID, foundIssue.EntityID)
		}
		if foundIssue.Field != issue.Field {
			t.Errorf("Expected Field %s, got %s", issue.Field, foundIssue.Field)
		}
		if foundIssue.Value != issue.Value {
			t.Errorf("Expected Value %s, got %s", issue.Value, foundIssue.Value)
		}
		if foundIssue.Suggestion != issue.Suggestion {
			t.Errorf("Expected Suggestion %s, got %s", issue.Suggestion, foundIssue.Suggestion)
		}
		if foundIssue.Location != issue.Location {
			t.Errorf("Expected Location %s, got %s", issue.Location, foundIssue.Location)
		}
		if len(foundIssue.Context) != 2 {
			t.Errorf("Expected 2 context items, got %d", len(foundIssue.Context))
		}
		if foundIssue.Context["context_key_1"] != "context_value_1" {
			t.Error("Context key/value mismatch")
		}
	})
}

func TestValidator_ProcessingStats_Comprehensive(t *testing.T) {
	validator := NewValidator()

	t.Run("Update and retrieve processing stats", func(t *testing.T) {
		validator.Reset()

		// Update stats for different entity types
		validator.UpdateProcessingStats("Agency", 10, 9, 1)
		validator.UpdateProcessingStats("Route", 25, 23, 2)
		validator.UpdateProcessingStats("Stop", 100, 95, 5)

		report := validator.GetReport()
		stats := report.ProcessingStats

		// Check individual entity stats
		if stats.EntitiesProcessed["Agency"] != 10 {
			t.Errorf("Expected 10 Agency entities processed, got %d", stats.EntitiesProcessed["Agency"])
		}
		if stats.EntitiesConverted["Route"] != 23 {
			t.Errorf("Expected 23 Route entities converted, got %d", stats.EntitiesConverted["Route"])
		}
		if stats.EntitiesSkipped["Stop"] != 5 {
			t.Errorf("Expected 5 Stop entities skipped, got %d", stats.EntitiesSkipped["Stop"])
		}

		// Check conversion rate calculation
		expectedRate := float64(9+23+95) / float64(10+25+100) * 100
		if stats.ConversionRate != expectedRate {
			t.Errorf("Expected conversion rate %.2f%%, got %.2f%%", expectedRate, stats.ConversionRate)
		}
	})

	t.Run("Zero processing stats", func(t *testing.T) {
		validator.Reset()

		report := validator.GetReport()
		stats := report.ProcessingStats

		if stats.ConversionRate != 0 {
			t.Errorf("Expected 0%% conversion rate for empty stats, got %.2f%%", stats.ConversionRate)
		}
	})
}

func TestValidator_ConfigurationOptions_Comprehensive(t *testing.T) {
	t.Run("Strict mode configuration", func(t *testing.T) {
		validator := NewValidator()

		config := ValidationConfig{
			EnableStrictMode:           true,
			MaxIssuesPerType:           50,
			EnableContextualValidation: true,
			SeverityThreshold:          SeverityInfo,
			EnableGTFSValidation:       true,
			EnableNeTExValidation:      true,
			ValidateReferences:         true,
			ValidateGeometry:           true,
			ValidateTiming:             true,
			ValidateAccessibility:      true,
		}
		validator.SetConfig(config)

		// Test that config is applied
		if validator.config.EnableStrictMode != true {
			t.Error("EnableStrictMode should be true")
		}
		if validator.config.MaxIssuesPerType != 50 {
			t.Error("MaxIssuesPerType should be 50")
		}
	})

	t.Run("Geometry validation toggle", func(t *testing.T) {
		validator := NewValidator()

		// Test with geometry validation disabled
		config := validator.config
		config.ValidateGeometry = false
		validator.SetConfig(config)

		stop := &model.Stop{
			StopID:   "test_stop",
			StopName: "Test Stop",
			StopLat:  -100.0, // Invalid but should be ignored
			StopLon:  200.0,  // Invalid but should be ignored
		}

		validator.ValidateGTFSStop(stop)
		report := validator.GetReport()

		// Should not have geometry issues since validation is disabled
		for _, issue := range report.Issues {
			if strings.Contains(issue.Code, "LATITUDE") || strings.Contains(issue.Code, "LONGITUDE") {
				t.Errorf("Found geometry validation issue when disabled: %s", issue.Code)
			}
		}
	})

	t.Run("Timing validation toggle", func(t *testing.T) {
		validator := NewValidator()

		// Test with timing validation disabled
		config := validator.config
		config.ValidateTiming = false
		validator.SetConfig(config)

		stopTime := &model.StopTime{
			TripID:        "test_trip",
			StopID:        "test_stop",
			StopSequence:  1,
			ArrivalTime:   "invalid_time", // Invalid but should be ignored
			DepartureTime: "25:70:90",     // Invalid but should be ignored
		}

		validator.ValidateGTFSStopTime(stopTime)
		report := validator.GetReport()

		// Should not have timing issues since validation is disabled
		for _, issue := range report.Issues {
			if strings.Contains(issue.Code, "TIME") {
				t.Errorf("Found timing validation issue when disabled: %s", issue.Code)
			}
		}
	})

	t.Run("Accessibility validation toggle", func(t *testing.T) {
		validator := NewValidator()

		// Test with accessibility validation disabled
		config := validator.config
		config.ValidateAccessibility = false
		validator.SetConfig(config)

		stop := &model.Stop{
			StopID:             "test_stop",
			StopName:           "Test Stop",
			StopLat:            59.911491,
			StopLon:            10.757933,
			WheelchairBoarding: "invalid_value", // Invalid but should be ignored
		}

		validator.ValidateGTFSStop(stop)
		report := validator.GetReport()

		// Should not have accessibility issues since validation is disabled
		for _, issue := range report.Issues {
			if strings.Contains(issue.Code, "WHEELCHAIR") {
				t.Errorf("Found accessibility validation issue when disabled: %s", issue.Code)
			}
		}
	})
}

func TestValidator_ValidationPatterns_Comprehensive(t *testing.T) {
	validator := NewValidator()

	t.Run("GTFS pattern validation", func(t *testing.T) {
		testCases := []struct {
			name         string
			pattern      string
			validInput   string
			invalidInput string
		}{
			{
				name:         "GTFS Time",
				pattern:      "time",
				validInput:   "14:30:45",
				invalidInput: "25:70:90",
			},
			{
				name:         "GTFS Date",
				pattern:      "date",
				validInput:   "20240101",
				invalidInput: "2024-01-01",
			},
			{
				name:         "GTFS Color",
				pattern:      "color",
				validInput:   "FF0000",
				invalidInput: "red",
			},
			{
				name:         "GTFS Email",
				pattern:      "email",
				validInput:   "test@example.com",
				invalidInput: "invalid-email",
			},
			{
				name:         "GTFS URL",
				pattern:      "url",
				validInput:   "https://example.com",
				invalidInput: "not-a-url",
			},
			{
				name:         "GTFS Phone",
				pattern:      "phone",
				validInput:   "+1-234-567-8900",
				invalidInput: "abc",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				var pattern *regexp.Regexp

				switch tc.pattern {
				case "time":
					pattern = validator.patterns.GTFSTime
				case "date":
					pattern = validator.patterns.GTFSDate
				case "color":
					pattern = validator.patterns.GTFSColor
				case "email":
					pattern = validator.patterns.GTFSEmail
				case "url":
					pattern = validator.patterns.GTFSURL
				case "phone":
					pattern = validator.patterns.GTFSPhone
				}

				if pattern == nil {
					t.Fatalf("Pattern for %s is nil", tc.pattern)
				}

				if !pattern.MatchString(tc.validInput) {
					t.Errorf("Valid input '%s' should match %s pattern", tc.validInput, tc.name)
				}

				if pattern.MatchString(tc.invalidInput) {
					t.Errorf("Invalid input '%s' should not match %s pattern", tc.invalidInput, tc.name)
				}
			})
		}
	})
}

func TestValidator_AdvancedGTFSValidation(t *testing.T) {
	t.Run("Route validation edge cases", func(t *testing.T) {
		validator := NewValidator()

		// Test route with extended route types (100+)
		route := &model.GtfsRoute{
			RouteID:        "test_route",
			RouteShortName: "Test",
			RouteType:      101, // Extended type - should be invalid per current validation
		}

		validator.ValidateGTFSRoute(route)
		report := validator.GetReport()

		// Should have invalid route type error
		found := false
		for _, issue := range report.Issues {
			if issue.Code == "ROUTE_INVALID_TYPE" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected ROUTE_INVALID_TYPE error for extended route type")
		}
	})

	t.Run("Stop time sequence validation", func(t *testing.T) {
		validator := NewValidator()

		// Test with zero stop sequence (should be invalid)
		stopTime := &model.StopTime{
			TripID:        "test_trip",
			StopID:        "test_stop",
			StopSequence:  0, // Invalid - should be positive
			ArrivalTime:   "08:00:00",
			DepartureTime: "08:00:00",
		}

		validator.ValidateGTFSStopTime(stopTime)
		report := validator.GetReport()

		found := false
		for _, issue := range report.Issues {
			if issue.Code == "STOPTIME_INVALID_SEQUENCE" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected STOPTIME_INVALID_SEQUENCE error for zero sequence")
		}
	})

	t.Run("Time parsing edge cases", func(t *testing.T) {
		validator := NewValidator()

		testCases := []struct {
			timeStr  string
			expected int
		}{
			{"00:00:00", 0},
			{"24:00:00", 24 * 60},    // Valid in GTFS
			{"25:30:00", 25*60 + 30}, // Valid in GTFS for next day
			{"invalid", -1},
			{"", -1},
			{"12:30", -1},            // Missing seconds
			{"12:30:60", 12*60 + 30}, // Invalid seconds but should parse hours/minutes
		}

		for _, tc := range testCases {
			result := validator.parseTimeToMinutes(tc.timeStr)
			if result != tc.expected {
				t.Errorf("parseTimeToMinutes(%q) = %d, expected %d", tc.timeStr, result, tc.expected)
			}
		}
	})
}

// Tests for ValidationService

func TestValidationService_BasicOperations(t *testing.T) {
	service := NewValidationService()

	if service == nil {
		t.Fatal("NewValidationService() returned nil")
	}

	// Test service configuration
	config := ServiceConfig{
		EnableRealTimeValidation:    true,
		EnablePostProcessValidation: true,
		EnableProgressReporting:     false,
		ValidateOnSave:              true,
		ProgressUpdateInterval:      100,
	}

	service.SetConfig(config)

	// Verify configuration was set
	if !service.config.EnableRealTimeValidation {
		t.Error("EnableRealTimeValidation should be true")
	}
	if service.config.ProgressUpdateInterval != 100 {
		t.Error("ProgressUpdateInterval should be 100")
	}
}

func TestValidationService_ConversionManagement(t *testing.T) {
	service := NewValidationService()

	t.Run("Start and finish conversion", func(t *testing.T) {
		// Start conversion
		ctx := service.StartConversion()
		if ctx == nil {
			t.Fatal("StartConversion() returned nil context")
		}

		if ctx.ConversionStats.NetexEntitiesLoaded == nil {
			t.Error("ConversionStats.NetexEntitiesLoaded should be initialized")
		}

		// Simulate some processing
		time.Sleep(10 * time.Millisecond)

		// Record some processing activities
		service.RecordProcessingTime(ctx, "test_stage", 5*time.Millisecond)
		service.RecordMemoryUsage(ctx, "test_stage")

		// Finish conversion
		report := service.FinishConversion(ctx)

		if report.ProcessingStats.ProcessingDuration == 0 {
			t.Error("ProcessingDuration should be set")
		}

		if report.ProcessingStats.MemoryUsageMB == 0 {
			t.Error("MemoryUsageMB should be set")
		}
	})

	t.Run("Record conversion errors", func(t *testing.T) {
		ctx := service.StartConversion()

		testEntity := &model.Agency{AgencyID: "test"}
		testError := fmt.Errorf("test conversion error")

		service.RecordConversionError(ctx, "test_stage", testEntity, testError)

		if len(ctx.ConversionStats.ConversionErrors["test_stage"]) == 0 {
			t.Error("Conversion error should be recorded")
		}

		// Check that validation issue was also created
		report := service.GetCurrentReport()
		foundError := false
		for _, issue := range report.Issues {
			if issue.Code == "CONVERSION_ERROR" {
				foundError = true
				break
			}
		}
		if !foundError {
			t.Error("Expected CONVERSION_ERROR validation issue")
		}
	})

	t.Run("Performance monitoring", func(t *testing.T) {
		ctx := service.StartConversion()

		// Record slow processing time (should trigger warning)
		service.RecordProcessingTime(ctx, "slow_stage", 35*time.Second)

		report := service.GetCurrentReport()
		foundWarning := false
		for _, issue := range report.Issues {
			if issue.Code == "PERFORMANCE_SLOW_STAGE" {
				foundWarning = true
				break
			}
		}
		if !foundWarning {
			t.Error("Expected PERFORMANCE_SLOW_STAGE warning for slow processing")
		}
	})
}

func TestValidationService_EntityValidation(t *testing.T) {
	service := NewValidationService()
	ctx := service.StartConversion()

	t.Run("GTFS entity validation", func(t *testing.T) {
		// Test with valid agency
		agency := &model.Agency{
			AgencyID:       "test_agency",
			AgencyName:     "Test Agency",
			AgencyURL:      "https://example.com",
			AgencyTimezone: "Europe/Oslo",
		}

		service.ValidateGTFSEntity(ctx, agency)

		if ctx.ConversionStats.GtfsEntitiesGenerated["Agency"] != 1 {
			t.Error("Agency count should be incremented")
		}

		// Test with invalid agency
		invalidAgency := &model.Agency{
			AgencyID: "test_invalid",
			// Missing required fields
		}

		service.ValidateGTFSEntity(ctx, invalidAgency)

		report := service.GetCurrentReport()
		if len(report.Issues) == 0 {
			t.Error("Expected validation issues for invalid agency")
		}
	})

	t.Run("Real-time validation toggle", func(t *testing.T) {
		// Disable real-time validation
		config := service.config
		config.EnableRealTimeValidation = false
		service.SetConfig(config)

		initialIssueCount := len(service.GetCurrentReport().Issues)

		// This should not add validation issues
		invalidAgency := &model.Agency{AgencyID: "invalid"}
		service.ValidateGTFSEntity(ctx, invalidAgency)

		finalIssueCount := len(service.GetCurrentReport().Issues)
		if finalIssueCount != initialIssueCount {
			t.Error("Validation issues should not be added when real-time validation is disabled")
		}
	})
}

func TestValidationService_ReportGeneration(t *testing.T) {
	service := NewValidationService()

	// Add some validation issues
	service.validator.AddIssue(ValidationIssue{
		Severity:   SeverityError,
		Code:       "TEST_ERROR",
		Message:    "Test error message",
		EntityType: "TestEntity",
		EntityID:   "test_1",
	})

	report := service.GetCurrentReport()
	if len(report.Issues) == 0 {
		t.Error("Expected validation issues in report")
	}

	// Test report generation in different formats
	formats := []ReportFormat{FormatJSON, FormatText, FormatMarkdown, FormatCSV}

	for _, format := range formats {
		t.Run(fmt.Sprintf("Generate %s report", []string{"JSON", "HTML", "Text", "CSV", "Markdown"}[format]), func(t *testing.T) {
			reportStr, err := service.GenerateReport(report, format)
			if err != nil {
				t.Errorf("Failed to generate report: %v", err)
			}
			if reportStr == "" {
				t.Error("Generated report should not be empty")
			}
		})
	}
}

func TestValidationService_ProgressReporting(t *testing.T) {
	service := NewValidationService()

	// Ensure real-time validation is enabled (should be by default)
	if !service.config.EnableRealTimeValidation {
		t.Fatal("RealTimeValidation should be enabled by default")
	}

	ctx := service.StartConversion()

	// Test simple entity counting without complex progress logic
	// Just verify that the service is properly incrementing counters
	authority := &model.Authority{
		ID:   "test_auth",
		Name: "Test Authority",
	}

	service.ValidateNeTExEntity(ctx, authority)

	// Debug: check if the counter was incremented
	if ctx.ConversionStats.NetexEntitiesLoaded["Authority"] != 1 {
		t.Errorf("Expected 1 Authority entity, got %d", ctx.ConversionStats.NetexEntitiesLoaded["Authority"])

		// Debug information
		t.Logf("EnableRealTimeValidation: %v", service.config.EnableRealTimeValidation)
		t.Logf("ConversionStats: %+v", ctx.ConversionStats.NetexEntitiesLoaded)
	}

	// Test GTFS entity as well
	agency := &model.Agency{
		AgencyID:       "test_agency",
		AgencyName:     "Test Agency",
		AgencyURL:      "https://example.com",
		AgencyTimezone: "Europe/Oslo",
	}

	service.ValidateGTFSEntity(ctx, agency)

	if ctx.ConversionStats.GtfsEntitiesGenerated["Agency"] != 1 {
		t.Errorf("Expected 1 Agency entity generated, got %d", ctx.ConversionStats.GtfsEntitiesGenerated["Agency"])

		// Debug information
		t.Logf("GTFS Stats: %+v", ctx.ConversionStats.GtfsEntitiesGenerated)
	}
}
