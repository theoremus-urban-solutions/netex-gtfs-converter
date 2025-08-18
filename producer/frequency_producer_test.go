package producer

import (
	"testing"

	"github.com/theoremus-urban-solutions/netex-gtfs-converter/model"
)

func TestDefaultFrequencyProducer_ConvertDurationToSeconds(t *testing.T) {
	netexRepo := &mockNetexRepository{}
	gtfsRepo := &mockGtfsRepository{}
	producer := NewDefaultFrequencyProducer(netexRepo, gtfsRepo)

	tests := []struct {
		name     string
		duration string
		expected int
		hasError bool
	}{
		{
			name:     "15 minutes",
			duration: "PT15M",
			expected: 900,
			hasError: false,
		},
		{
			name:     "1 hour",
			duration: "PT1H",
			expected: 3600,
			hasError: false,
		},
		{
			name:     "1 hour 30 minutes",
			duration: "PT1H30M",
			expected: 5400,
			hasError: false,
		},
		{
			name:     "30 seconds",
			duration: "PT30S",
			expected: 30,
			hasError: false,
		},
		{
			name:     "2 hours 15 minutes 30 seconds",
			duration: "PT2H15M30S",
			expected: 8130,
			hasError: false,
		},
		{
			name:     "empty duration",
			duration: "",
			expected: 0,
			hasError: true,
		},
		{
			name:     "invalid format",
			duration: "invalid",
			expected: 0,
			hasError: true,
		},
		{
			name:     "no PT prefix",
			duration: "15M",
			expected: 0,
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := producer.convertDurationToSeconds(tt.duration)

			if tt.hasError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("Expected %d seconds, got %d", tt.expected, result)
				}
			}
		})
	}
}

func TestDefaultFrequencyProducer_ConvertTimeToSeconds(t *testing.T) {
	netexRepo := &mockNetexRepository{}
	gtfsRepo := &mockGtfsRepository{}
	producer := NewDefaultFrequencyProducer(netexRepo, gtfsRepo)

	tests := []struct {
		name     string
		timeStr  string
		expected int
		hasError bool
	}{
		{
			name:     "midnight",
			timeStr:  "00:00:00",
			expected: 0,
			hasError: false,
		},
		{
			name:     "6 AM",
			timeStr:  "06:00:00",
			expected: 21600,
			hasError: false,
		},
		{
			name:     "9:30 AM",
			timeStr:  "09:30:00",
			expected: 34200,
			hasError: false,
		},
		{
			name:     "11:59:59 PM",
			timeStr:  "23:59:59",
			expected: 86399,
			hasError: false,
		},
		{
			name:     "empty time",
			timeStr:  "",
			expected: 0,
			hasError: true,
		},
		{
			name:     "invalid format",
			timeStr:  "invalid",
			expected: 0,
			hasError: true,
		},
		{
			name:     "missing seconds",
			timeStr:  "09:30",
			expected: 0,
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := producer.convertTimeToSeconds(tt.timeStr)

			if tt.hasError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("Expected %d seconds, got %d", tt.expected, result)
				}
			}
		})
	}
}

func TestDefaultFrequencyProducer_FormatTimeForGTFS(t *testing.T) {
	netexRepo := &mockNetexRepository{}
	gtfsRepo := &mockGtfsRepository{}
	producer := NewDefaultFrequencyProducer(netexRepo, gtfsRepo)

	tests := []struct {
		name     string
		seconds  int
		expected string
	}{
		{
			name:     "midnight",
			seconds:  0,
			expected: "00:00:00",
		},
		{
			name:     "6 AM",
			seconds:  21600,
			expected: "06:00:00",
		},
		{
			name:     "9:30:45 AM",
			seconds:  34245,
			expected: "09:30:45",
		},
		{
			name:     "11:59:59 PM",
			seconds:  86399,
			expected: "23:59:59",
		},
		{
			name:     "25:30:00 (next day)",
			seconds:  91800,
			expected: "25:30:00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := producer.formatTimeForGTFS(tt.seconds)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestDefaultFrequencyProducer_ProduceFromHeadwayJourneyGroup(t *testing.T) {
	// Mock repository with journey pattern
	mockRepo := &mockNetexRepositoryWithJourneyPatterns{
		journeyPatterns: map[string]*model.JourneyPattern{
			"jp1": {
				ID:   "jp1",
				Name: "Test Journey Pattern",
			},
		},
	}
	gtfsRepo := &mockGtfsRepository{}
	producer := NewDefaultFrequencyProducer(mockRepo, gtfsRepo)

	tests := []struct {
		name     string
		group    *model.HeadwayJourneyGroup
		expected int
		hasError bool
	}{
		{
			name:     "nil group",
			group:    nil,
			expected: 0,
			hasError: false,
		},
		{
			name: "valid headway group",
			group: &model.HeadwayJourneyGroup{
				ID:                       "hj1",
				Name:                     "Test Headway",
				ScheduledHeadwayInterval: "PT15M",
				FirstDepartureTime:       "06:00:00",
				LastDepartureTime:        "22:00:00",
				JourneyPatternRef:        "jp1",
			},
			expected: 1,
			hasError: false,
		},
		{
			name: "missing journey pattern",
			group: &model.HeadwayJourneyGroup{
				ID:                       "hj2",
				Name:                     "Test Headway 2",
				ScheduledHeadwayInterval: "PT10M",
				FirstDepartureTime:       "07:00:00",
				LastDepartureTime:        "21:00:00",
				JourneyPatternRef:        "nonexistent",
			},
			expected: 0,
			hasError: true,
		},
		{
			name: "invalid headway interval",
			group: &model.HeadwayJourneyGroup{
				ID:                       "hj3",
				Name:                     "Test Headway 3",
				ScheduledHeadwayInterval: "invalid",
				FirstDepartureTime:       "06:00:00",
				LastDepartureTime:        "22:00:00",
				JourneyPatternRef:        "jp1",
			},
			expected: 0,
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			frequencies, err := producer.ProduceFromHeadwayJourneyGroup(tt.group)

			if tt.hasError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if len(frequencies) != tt.expected {
					t.Errorf("Expected %d frequencies, got %d", tt.expected, len(frequencies))
				}

				if len(frequencies) > 0 {
					freq := frequencies[0]
					if freq.TripID == "" {
						t.Error("Trip ID should not be empty")
					}
					if freq.HeadwaySecs <= 0 {
						t.Error("Headway seconds should be positive")
					}
				}
			}
		})
	}
}

func TestDefaultFrequencyProducer_ValidateFrequencyTrip(t *testing.T) {
	netexRepo := &mockNetexRepository{}
	gtfsRepo := &mockGtfsRepository{}
	producer := NewDefaultFrequencyProducer(netexRepo, gtfsRepo)

	trip := &model.Trip{
		TripID:    "test_trip",
		RouteID:   "test_route",
		ServiceID: "test_service",
	}

	tests := []struct {
		name        string
		trip        *model.Trip
		frequencies []*model.Frequency
		hasError    bool
	}{
		{
			name:        "nil trip",
			trip:        nil,
			frequencies: []*model.Frequency{},
			hasError:    true,
		},
		{
			name:        "no frequencies",
			trip:        trip,
			frequencies: []*model.Frequency{},
			hasError:    true,
		},
		{
			name: "valid frequency",
			trip: trip,
			frequencies: []*model.Frequency{
				{
					TripID:      "test_trip",
					StartTime:   "06:00:00",
					EndTime:     "22:00:00",
					HeadwaySecs: 900,
					ExactTimes:  "0",
				},
			},
			hasError: false,
		},
		{
			name: "mismatched trip ID",
			trip: trip,
			frequencies: []*model.Frequency{
				{
					TripID:      "different_trip",
					StartTime:   "06:00:00",
					EndTime:     "22:00:00",
					HeadwaySecs: 900,
					ExactTimes:  "0",
				},
			},
			hasError: true,
		},
		{
			name: "invalid headway",
			trip: trip,
			frequencies: []*model.Frequency{
				{
					TripID:      "test_trip",
					StartTime:   "06:00:00",
					EndTime:     "22:00:00",
					HeadwaySecs: 0,
					ExactTimes:  "0",
				},
			},
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := producer.ValidateFrequencyTrip(tt.trip, tt.frequencies)

			if tt.hasError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

// Enhanced mock repository with journey pattern support
type mockNetexRepositoryWithJourneyPatterns struct {
	mockNetexRepository
	journeyPatterns map[string]*model.JourneyPattern
}

func (m *mockNetexRepositoryWithJourneyPatterns) GetJourneyPatternById(id string) *model.JourneyPattern {
	return m.journeyPatterns[id]
}

func TestFrequencyService_GetTotalServiceHours(t *testing.T) {
	service := &FrequencyService{
		Frequencies: []*model.Frequency{
			{
				StartTime: "06:00:00",
				EndTime:   "10:00:00",
			},
			{
				StartTime: "14:00:00",
				EndTime:   "18:00:00",
			},
		},
	}

	expected := 8.0 // 4 hours + 4 hours
	result := service.GetTotalServiceHours()

	if result != expected {
		t.Errorf("Expected %f hours, got %f", expected, result)
	}
}

func TestFrequencyService_GetAverageHeadway(t *testing.T) {
	service := &FrequencyService{
		Frequencies: []*model.Frequency{
			{HeadwaySecs: 600},  // 10 minutes
			{HeadwaySecs: 1200}, // 20 minutes
		},
	}

	expected := 15.0 // Average of 10 and 20 minutes
	result := service.GetAverageHeadway()

	if result != expected {
		t.Errorf("Expected %f minutes, got %f", expected, result)
	}
}
