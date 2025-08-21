package validation

import (
	"fmt"
	"testing"

	"github.com/theoremus-urban-solutions/netex-gtfs-converter/model"
)

func TestValidator_ValidateGTFSAgency(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name           string
		agency         *model.Agency
		expectedIssues int
		expectedCodes  []string
	}{
		{
			name:           "nil agency",
			agency:         nil,
			expectedIssues: 1,
			expectedCodes:  []string{"AGENCY_NULL"},
		},
		{
			name: "valid agency",
			agency: &model.Agency{
				AgencyID:       "test_agency",
				AgencyName:     "Test Agency",
				AgencyURL:      "https://example.com",
				AgencyTimezone: "Europe/Oslo",
			},
			expectedIssues: 0,
			expectedCodes:  []string{},
		},
		{
			name: "missing required fields",
			agency: &model.Agency{
				AgencyID: "test_agency",
				// Missing name, URL, timezone
			},
			expectedIssues: 3,
			expectedCodes: []string{
				"AGENCY_MISSING_NAME",
				"AGENCY_MISSING_URL",
				"AGENCY_MISSING_TIMEZONE",
			},
		},
		{
			name: "invalid URL format",
			agency: &model.Agency{
				AgencyID:       "test_agency",
				AgencyName:     "Test Agency",
				AgencyURL:      "invalid-url",
				AgencyTimezone: "Europe/Oslo",
			},
			expectedIssues: 1,
			expectedCodes:  []string{"AGENCY_INVALID_URL"},
		},
		{
			name: "invalid timezone",
			agency: &model.Agency{
				AgencyID:       "test_agency",
				AgencyName:     "Test Agency",
				AgencyURL:      "https://example.com",
				AgencyTimezone: "Invalid/Timezone",
			},
			expectedIssues: 1,
			expectedCodes:  []string{"AGENCY_INVALID_TIMEZONE"},
		},
		{
			name: "invalid email format",
			agency: &model.Agency{
				AgencyID:       "test_agency",
				AgencyName:     "Test Agency",
				AgencyURL:      "https://example.com",
				AgencyTimezone: "Europe/Oslo",
				AgencyEmail:    "invalid-email",
			},
			expectedIssues: 1,
			expectedCodes:  []string{"AGENCY_INVALID_EMAIL"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator.Reset()
			validator.ValidateGTFSAgency(tt.agency)

			report := validator.GetReport()
			if len(report.Issues) != tt.expectedIssues {
				t.Errorf("Expected %d issues, got %d", tt.expectedIssues, len(report.Issues))
			}

			// Check that expected issue codes are present
			foundCodes := make(map[string]bool)
			for _, issue := range report.Issues {
				foundCodes[issue.Code] = true
			}

			for _, expectedCode := range tt.expectedCodes {
				if !foundCodes[expectedCode] {
					t.Errorf("Expected issue code %s not found", expectedCode)
				}
			}
		})
	}
}

func TestValidator_ValidateGTFSRoute(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name           string
		route          *model.GtfsRoute
		expectedIssues int
		expectedCodes  []string
	}{
		{
			name:           "nil route",
			route:          nil,
			expectedIssues: 1,
			expectedCodes:  []string{"ROUTE_NULL"},
		},
		{
			name: "valid route",
			route: &model.GtfsRoute{
				RouteID:        "test_route",
				RouteShortName: "Test",
				RouteType:      3, // Bus
			},
			expectedIssues: 0,
			expectedCodes:  []string{},
		},
		{
			name: "missing names",
			route: &model.GtfsRoute{
				RouteID:   "test_route",
				RouteType: 3,
				// Missing both short and long name
			},
			expectedIssues: 1,
			expectedCodes:  []string{"ROUTE_MISSING_NAME"},
		},
		{
			name: "invalid route type",
			route: &model.GtfsRoute{
				RouteID:        "test_route",
				RouteShortName: "Test",
				RouteType:      999, // Invalid type
			},
			expectedIssues: 1,
			expectedCodes:  []string{"ROUTE_INVALID_TYPE"},
		},
		{
			name: "invalid color format",
			route: &model.GtfsRoute{
				RouteID:        "test_route",
				RouteShortName: "Test",
				RouteType:      3,
				RouteColor:     "invalid",
			},
			expectedIssues: 1,
			expectedCodes:  []string{"ROUTE_INVALID_COLOR"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator.Reset()
			validator.ValidateGTFSRoute(tt.route)

			report := validator.GetReport()
			if len(report.Issues) != tt.expectedIssues {
				t.Errorf("Expected %d issues, got %d", tt.expectedIssues, len(report.Issues))
			}

			foundCodes := make(map[string]bool)
			for _, issue := range report.Issues {
				foundCodes[issue.Code] = true
			}

			for _, expectedCode := range tt.expectedCodes {
				if !foundCodes[expectedCode] {
					t.Errorf("Expected issue code %s not found", expectedCode)
				}
			}
		})
	}
}

func TestValidator_ValidateGTFSStop(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name           string
		stop           *model.Stop
		expectedIssues int
		expectedCodes  []string
	}{
		{
			name:           "nil stop",
			stop:           nil,
			expectedIssues: 1,
			expectedCodes:  []string{"STOP_NULL"},
		},
		{
			name: "valid stop",
			stop: &model.Stop{
				StopID:   "test_stop",
				StopName: "Test Stop",
				StopLat:  59.911491,
				StopLon:  10.757933,
			},
			expectedIssues: 0,
			expectedCodes:  []string{},
		},
		{
			name: "missing required fields",
			stop: &model.Stop{
				StopLat: 59.911491,
				StopLon: 10.757933,
				// Missing ID and name
			},
			expectedIssues: 2,
			expectedCodes: []string{
				"STOP_MISSING_ID",
				"STOP_MISSING_NAME",
			},
		},
		{
			name: "invalid coordinates",
			stop: &model.Stop{
				StopID:   "test_stop",
				StopName: "Test Stop",
				StopLat:  -100.0, // Invalid latitude
				StopLon:  200.0,  // Invalid longitude
			},
			expectedIssues: 2,
			expectedCodes: []string{
				"STOP_INVALID_LATITUDE",
				"STOP_INVALID_LONGITUDE",
			},
		},
		{
			name: "suspicious coordinates",
			stop: &model.Stop{
				StopID:   "test_stop",
				StopName: "Test Stop",
				StopLat:  0.0, // Suspicious coordinates
				StopLon:  0.0,
			},
			expectedIssues: 1,
			expectedCodes:  []string{"STOP_SUSPICIOUS_COORDINATES"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator.Reset()
			validator.ValidateGTFSStop(tt.stop)

			report := validator.GetReport()
			if len(report.Issues) != tt.expectedIssues {
				t.Errorf("Expected %d issues, got %d", tt.expectedIssues, len(report.Issues))
			}

			foundCodes := make(map[string]bool)
			for _, issue := range report.Issues {
				foundCodes[issue.Code] = true
			}

			for _, expectedCode := range tt.expectedCodes {
				if !foundCodes[expectedCode] {
					t.Errorf("Expected issue code %s not found", expectedCode)
				}
			}
		})
	}
}

func TestValidator_ValidateGTFSStopTime(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name           string
		stopTime       *model.StopTime
		expectedIssues int
		expectedCodes  []string
	}{
		{
			name:           "nil stop time",
			stopTime:       nil,
			expectedIssues: 1,
			expectedCodes:  []string{"STOPTIME_NULL"},
		},
		{
			name: "valid stop time",
			stopTime: &model.StopTime{
				TripID:        "test_trip",
				ArrivalTime:   "08:00:00",
				DepartureTime: "08:00:30",
				StopID:        "test_stop",
				StopSequence:  1,
			},
			expectedIssues: 0,
			expectedCodes:  []string{},
		},
		{
			name: "missing required fields",
			stopTime: &model.StopTime{
				ArrivalTime:   "08:00:00",
				DepartureTime: "08:00:30",
				// Missing trip ID, stop ID, sequence
			},
			expectedIssues: 3,
			expectedCodes: []string{
				"STOPTIME_MISSING_TRIP_ID",
				"STOPTIME_MISSING_STOP_ID",
				"STOPTIME_INVALID_SEQUENCE",
			},
		},
		{
			name: "invalid time formats",
			stopTime: &model.StopTime{
				TripID:        "test_trip",
				ArrivalTime:   "invalid_time",
				DepartureTime: "25:70:90",
				StopID:        "test_stop",
				StopSequence:  1,
			},
			expectedIssues: 2,
			expectedCodes: []string{
				"STOPTIME_INVALID_ARRIVAL_TIME",
				"STOPTIME_INVALID_DEPARTURE_TIME",
			},
		},
		{
			name: "departure before arrival",
			stopTime: &model.StopTime{
				TripID:        "test_trip",
				ArrivalTime:   "08:30:00",
				DepartureTime: "08:00:00", // Before arrival
				StopID:        "test_stop",
				StopSequence:  1,
			},
			expectedIssues: 1,
			expectedCodes:  []string{"STOPTIME_DEPARTURE_BEFORE_ARRIVAL"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator.Reset()
			validator.ValidateGTFSStopTime(tt.stopTime)

			report := validator.GetReport()
			if len(report.Issues) != tt.expectedIssues {
				t.Errorf("Expected %d issues, got %d", tt.expectedIssues, len(report.Issues))
				for _, issue := range report.Issues {
					t.Logf("Issue: %s - %s", issue.Code, issue.Message)
				}
			}

			foundCodes := make(map[string]bool)
			for _, issue := range report.Issues {
				foundCodes[issue.Code] = true
			}

			for _, expectedCode := range tt.expectedCodes {
				if !foundCodes[expectedCode] {
					t.Errorf("Expected issue code %s not found", expectedCode)
				}
			}
		})
	}
}

func TestValidator_ParseTimeToMinutes(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		timeStr  string
		expected int
	}{
		{"08:30:00", 8*60 + 30},
		{"00:00:00", 0},
		{"12:45:30", 12*60 + 45},
		{"24:00:00", 24 * 60},
		{"invalid", -1},
		{"8:30", -1}, // Missing seconds
	}

	for _, tt := range tests {
		t.Run(tt.timeStr, func(t *testing.T) {
			result := validator.parseTimeToMinutes(tt.timeStr)
			if result != tt.expected {
				t.Errorf("parseTimeToMinutes(%s) = %d, expected %d", tt.timeStr, result, tt.expected)
			}
		})
	}
}

func TestValidator_ValidationSummary(t *testing.T) {
	validator := NewValidator()

	// Add various issues
	validator.AddIssue(ValidationIssue{
		Severity:   SeverityCritical,
		Code:       "TEST_CRITICAL",
		Message:    "Critical issue",
		EntityType: "TestEntity",
	})

	validator.AddIssue(ValidationIssue{
		Severity:   SeverityError,
		Code:       "TEST_ERROR",
		Message:    "Error issue",
		EntityType: "TestEntity",
	})

	validator.AddIssue(ValidationIssue{
		Severity:   SeverityWarning,
		Code:       "TEST_WARNING",
		Message:    "Warning issue",
		EntityType: "AnotherEntity",
	})

	report := validator.GetReport()

	// Verify summary
	if report.Summary.TotalIssues != 3 {
		t.Errorf("Expected 3 total issues, got %d", report.Summary.TotalIssues)
	}

	if !report.Summary.HasCritical {
		t.Error("Expected HasCritical to be true")
	}

	if !report.Summary.HasErrors {
		t.Error("Expected HasErrors to be true")
	}

	if report.Summary.IsValid {
		t.Error("Expected IsValid to be false")
	}

	// Check severity breakdown
	if report.Summary.BySeverity[SeverityCritical] != 1 {
		t.Errorf("Expected 1 critical issue, got %d", report.Summary.BySeverity[SeverityCritical])
	}

	if report.Summary.BySeverity[SeverityError] != 1 {
		t.Errorf("Expected 1 error issue, got %d", report.Summary.BySeverity[SeverityError])
	}

	if report.Summary.BySeverity[SeverityWarning] != 1 {
		t.Errorf("Expected 1 warning issue, got %d", report.Summary.BySeverity[SeverityWarning])
	}

	// Check entity type breakdown
	if report.Summary.ByEntityType["TestEntity"] != 2 {
		t.Errorf("Expected 2 TestEntity issues, got %d", report.Summary.ByEntityType["TestEntity"])
	}

	if report.Summary.ByEntityType["AnotherEntity"] != 1 {
		t.Errorf("Expected 1 AnotherEntity issue, got %d", report.Summary.ByEntityType["AnotherEntity"])
	}
}

func TestValidator_MaxIssuesPerType(t *testing.T) {
	config := ValidationConfig{
		MaxIssuesPerType:  2,
		SeverityThreshold: SeverityInfo,
	}

	validator := NewValidator()
	validator.SetConfig(config)

	// Add more issues than the limit for the same type
	for i := 0; i < 5; i++ {
		validator.AddIssue(ValidationIssue{
			Severity:   SeverityWarning,
			Code:       "TEST_REPEATED",
			Message:    fmt.Sprintf("Repeated issue %d", i),
			EntityType: "TestEntity",
		})
	}

	report := validator.GetReport()

	// Should only have 2 issues due to the limit
	if len(report.Issues) != 2 {
		t.Errorf("Expected 2 issues due to MaxIssuesPerType limit, got %d", len(report.Issues))
	}
}

func TestValidator_SeverityThreshold(t *testing.T) {
	config := ValidationConfig{
		SeverityThreshold: SeverityError,
		MaxIssuesPerType:  100, // Ensure this doesn't limit our test
	}

	validator := NewValidator()
	validator.SetConfig(config)

	// Add issues of various severities
	validator.AddIssue(ValidationIssue{
		Severity: SeverityInfo,
		Code:     "TEST_INFO",
		Message:  "Info message",
	})

	validator.AddIssue(ValidationIssue{
		Severity: SeverityWarning,
		Code:     "TEST_WARNING",
		Message:  "Warning message",
	})

	validator.AddIssue(ValidationIssue{
		Severity: SeverityError,
		Code:     "TEST_ERROR",
		Message:  "Error message",
	})

	validator.AddIssue(ValidationIssue{
		Severity: SeverityCritical,
		Code:     "TEST_CRITICAL",
		Message:  "Critical message",
	})

	report := validator.GetReport()

	// Should only have error and critical issues (2 total)
	if len(report.Issues) != 2 {
		t.Errorf("Expected 2 issues above severity threshold, got %d", len(report.Issues))
	}

	for _, issue := range report.Issues {
		if issue.Severity < SeverityError {
			t.Errorf("Found issue below severity threshold: %s", issue.Severity.String())
		}
	}
}
