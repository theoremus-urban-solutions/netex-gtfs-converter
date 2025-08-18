package producer

import (
	"testing"
	"time"

	"github.com/theoremus-urban-solutions/netex-gtfs-converter/model"
)

func TestAdvancedStopTimeProducer_TimeInterpolation(t *testing.T) {
	mockRepo := &mockNetexRepositoryWithJourneyPatterns{
		journeyPatterns: map[string]*model.JourneyPattern{
			"jp1": createTestJourneyPattern(),
		},
	}
	mockGtfsRepo := &mockGtfsRepository{}

	producer := NewAdvancedStopTimeProducer(mockRepo, mockGtfsRepo)

	// Create test service journey with partial times
	serviceJourney := &model.ServiceJourney{
		ID:                "sj1",
		JourneyPatternRef: "jp1",
		PassingTimes: &model.PassingTimes{
			TimetabledPassingTime: []model.TimetabledPassingTime{
				{
					PointInJourneyPatternRef: "stop1",
					DepartureTime:            timePtr("08:00:00"),
				},
				// stop2 has no time - should be interpolated
				{
					PointInJourneyPatternRef: "stop3",
					ArrivalTime:              timePtr("08:10:00"),
				},
			},
		},
	}

	trip := &model.Trip{TripID: "trip1"}

	input := TripStopTimeInput{
		ServiceJourney:  serviceJourney,
		Trip:            trip,
		CurrentHeadSign: "Test Route",
	}

	stopTimes, err := producer.ProduceAdvanced(input)
	if err != nil {
		t.Fatalf("ProduceAdvanced failed: %v", err)
	}

	if len(stopTimes) != 3 {
		t.Fatalf("Expected 3 stop times, got %d", len(stopTimes))
	}

	// Check that stop2 time was interpolated
	stop2Time := stopTimes[1]
	if stop2Time.ArrivalTime == "" {
		t.Error("Stop 2 arrival time should be interpolated")
	}

	// Verify time progression
	stop1Departure := parseTimeString(stopTimes[0].DepartureTime)
	stop2Arrival := parseTimeString(stopTimes[1].ArrivalTime)
	stop3Arrival := parseTimeString(stopTimes[2].ArrivalTime)

	if stop2Arrival <= stop1Departure {
		t.Error("Stop 2 arrival should be after stop 1 departure")
	}

	if stop3Arrival <= stop2Arrival {
		t.Error("Stop 3 arrival should be after stop 2 arrival")
	}
}

func TestAdvancedStopTimeProducer_ShapeDistanceCalculation(t *testing.T) {
	mockRepo := &mockNetexRepositoryWithJourneyPatterns{
		journeyPatterns: map[string]*model.JourneyPattern{
			"jp1": createTestJourneyPattern(),
		},
	}
	mockGtfsRepo := &mockGtfsRepository{}

	producer := NewAdvancedStopTimeProducer(mockRepo, mockGtfsRepo)

	// Create test shape (simplified for testing)
	shape := &model.Shape{
		ShapeID:           "shape1",
		ShapePtLat:        60.0,
		ShapePtLon:        10.0,
		ShapePtSequence:   1,
		ShapeDistTraveled: 0.0,
	}

	serviceJourney := &model.ServiceJourney{
		ID:                "sj1",
		JourneyPatternRef: "jp1",
		PassingTimes: &model.PassingTimes{
			TimetabledPassingTime: []model.TimetabledPassingTime{
				{
					PointInJourneyPatternRef: "stop1",
					DepartureTime:            timePtr("08:00:00"),
				},
				{
					PointInJourneyPatternRef: "stop2",
					ArrivalTime:              timePtr("08:05:00"),
					DepartureTime:            timePtr("08:05:30"),
				},
				{
					PointInJourneyPatternRef: "stop3",
					ArrivalTime:              timePtr("08:10:00"),
				},
			},
		},
	}

	trip := &model.Trip{TripID: "trip1"}

	input := TripStopTimeInput{
		ServiceJourney:  serviceJourney,
		Trip:            trip,
		Shape:           shape,
		CurrentHeadSign: "Test Route",
	}

	stopTimes, err := producer.ProduceAdvanced(input)
	if err != nil {
		t.Fatalf("ProduceAdvanced failed: %v", err)
	}

	// Check that shape distances are calculated
	if stopTimes[0].ShapeDistTraveled != 0.0 {
		t.Errorf("First stop should have distance 0, got %f", stopTimes[0].ShapeDistTraveled)
	}

	if stopTimes[1].ShapeDistTraveled <= stopTimes[0].ShapeDistTraveled {
		t.Error("Shape distance should increase along the route")
	}

	if stopTimes[2].ShapeDistTraveled <= stopTimes[1].ShapeDistTraveled {
		t.Error("Shape distance should increase along the route")
	}
}

func TestAdvancedStopTimeProducer_ExtrapolationForward(t *testing.T) {
	mockRepo := &mockNetexRepositoryWithJourneyPatterns{
		journeyPatterns: map[string]*model.JourneyPattern{
			"jp1": createTestJourneyPattern(),
		},
	}
	mockGtfsRepo := &mockGtfsRepository{}

	producer := NewAdvancedStopTimeProducer(mockRepo, mockGtfsRepo)

	// Only first stop has time - others should be extrapolated
	serviceJourney := &model.ServiceJourney{
		ID:                "sj1",
		JourneyPatternRef: "jp1",
		PassingTimes: &model.PassingTimes{
			TimetabledPassingTime: []model.TimetabledPassingTime{
				{
					PointInJourneyPatternRef: "stop1",
					DepartureTime:            timePtr("08:00:00"),
				},
			},
		},
	}

	trip := &model.Trip{TripID: "trip1"}

	input := TripStopTimeInput{
		ServiceJourney:  serviceJourney,
		Trip:            trip,
		CurrentHeadSign: "Test Route",
	}

	stopTimes, err := producer.ProduceAdvanced(input)
	if err != nil {
		t.Fatalf("ProduceAdvanced failed: %v", err)
	}

	if len(stopTimes) != 3 {
		t.Fatalf("Expected 3 stop times, got %d", len(stopTimes))
	}

	// Check that all stops have times
	for i, st := range stopTimes {
		if st.ArrivalTime == "" {
			t.Errorf("Stop %d missing arrival time", i+1)
		}
		if st.DepartureTime == "" {
			t.Errorf("Stop %d missing departure time", i+1)
		}
	}

	// Check time progression
	for i := 1; i < len(stopTimes); i++ {
		prevTime := parseTimeString(stopTimes[i-1].DepartureTime)
		currTime := parseTimeString(stopTimes[i].ArrivalTime)

		if currTime <= prevTime {
			t.Errorf("Time should progress: stop %d departure %s should be before stop %d arrival %s",
				i, stopTimes[i-1].DepartureTime, i+1, stopTimes[i].ArrivalTime)
		}
	}
}

func TestAdvancedStopTimeProducer_ExtrapolationBackward(t *testing.T) {
	mockRepo := &mockNetexRepositoryWithJourneyPatterns{
		journeyPatterns: map[string]*model.JourneyPattern{
			"jp1": createTestJourneyPattern(),
		},
	}
	mockGtfsRepo := &mockGtfsRepository{}

	producer := NewAdvancedStopTimeProducer(mockRepo, mockGtfsRepo)

	// Only last stop has time - others should be extrapolated backward
	serviceJourney := &model.ServiceJourney{
		ID:                "sj1",
		JourneyPatternRef: "jp1",
		PassingTimes: &model.PassingTimes{
			TimetabledPassingTime: []model.TimetabledPassingTime{
				{
					PointInJourneyPatternRef: "stop3",
					ArrivalTime:              timePtr("08:10:00"),
				},
			},
		},
	}

	trip := &model.Trip{TripID: "trip1"}

	input := TripStopTimeInput{
		ServiceJourney:  serviceJourney,
		Trip:            trip,
		CurrentHeadSign: "Test Route",
	}

	stopTimes, err := producer.ProduceAdvanced(input)
	if err != nil {
		t.Fatalf("ProduceAdvanced failed: %v", err)
	}

	// Check that all stops have times
	for i, st := range stopTimes {
		if st.ArrivalTime == "" {
			t.Errorf("Stop %d missing arrival time", i+1)
		}
		if st.DepartureTime == "" {
			t.Errorf("Stop %d missing departure time", i+1)
		}
	}

	// Check time progression
	for i := 1; i < len(stopTimes); i++ {
		prevTime := parseTimeString(stopTimes[i-1].DepartureTime)
		currTime := parseTimeString(stopTimes[i].ArrivalTime)

		if currTime <= prevTime {
			t.Errorf("Time should progress: stop %d departure %s should be before stop %d arrival %s",
				i, stopTimes[i-1].DepartureTime, i+1, stopTimes[i].ArrivalTime)
		}
	}
}

func TestAdvancedStopTimeProducer_ValidateStopTimeSequence(t *testing.T) {
	producer := NewAdvancedStopTimeProducer(nil, nil)

	tests := []struct {
		name      string
		stopTimes []*model.StopTime
		expectErr bool
	}{
		{
			name: "valid sequence",
			stopTimes: []*model.StopTime{
				{StopSequence: 1, ArrivalTime: "08:00:00", DepartureTime: "08:00:30"},
				{StopSequence: 2, ArrivalTime: "08:05:00", DepartureTime: "08:05:30"},
				{StopSequence: 3, ArrivalTime: "08:10:00", DepartureTime: "08:10:30"},
			},
			expectErr: false,
		},
		{
			name: "invalid sequence - time goes backward",
			stopTimes: []*model.StopTime{
				{StopSequence: 1, ArrivalTime: "08:00:00", DepartureTime: "08:00:30"},
				{StopSequence: 2, ArrivalTime: "07:55:00", DepartureTime: "07:55:30"}, // Earlier than previous
				{StopSequence: 3, ArrivalTime: "08:10:00", DepartureTime: "08:10:30"},
			},
			expectErr: true,
		},
		{
			name: "out of order sequence gets sorted",
			stopTimes: []*model.StopTime{
				{StopSequence: 3, ArrivalTime: "08:10:00", DepartureTime: "08:10:30"},
				{StopSequence: 1, ArrivalTime: "08:00:00", DepartureTime: "08:00:30"},
				{StopSequence: 2, ArrivalTime: "08:05:00", DepartureTime: "08:05:30"},
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := producer.ValidateStopTimeSequence(tt.stopTimes)
			if tt.expectErr && err == nil {
				t.Error("Expected validation error but got none")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Unexpected validation error: %v", err)
			}

			// Check that sequence is sorted
			for i := 1; i < len(tt.stopTimes); i++ {
				if tt.stopTimes[i].StopSequence < tt.stopTimes[i-1].StopSequence {
					t.Error("Stop times should be sorted by sequence after validation")
				}
			}
		})
	}
}

func TestAdvancedStopTimeProducer_ConfigurableParameters(t *testing.T) {
	producer := NewAdvancedStopTimeProducer(nil, nil)

	// Test default values
	if producer.defaultTravelSpeed != 25.0 {
		t.Errorf("Expected default speed 25.0, got %f", producer.defaultTravelSpeed)
	}

	if producer.minStopDuration != 30 {
		t.Errorf("Expected min stop duration 30, got %d", producer.minStopDuration)
	}

	// Test configuration
	producer.SetDefaultTravelSpeed(40.0)
	producer.SetMinStopDuration(60)

	if producer.defaultTravelSpeed != 40.0 {
		t.Errorf("Expected speed 40.0, got %f", producer.defaultTravelSpeed)
	}

	if producer.minStopDuration != 60 {
		t.Errorf("Expected min stop duration 60, got %d", producer.minStopDuration)
	}
}

// Helper functions

func createTestJourneyPattern() *model.JourneyPattern {
	return &model.JourneyPattern{
		ID:   "jp1",
		Name: "Test Journey Pattern",
		PointsInSequence: &model.PointsInSequence{
			PointInJourneyPatternOrStopPointInJourneyPatternOrTimingPointInJourneyPattern: []interface{}{
				&model.StopPointInJourneyPattern{
					ID:                    "stop1",
					ScheduledStopPointRef: "ssp1",
					Order:                 1,
				},
				&model.StopPointInJourneyPattern{
					ID:                    "stop2",
					ScheduledStopPointRef: "ssp2",
					Order:                 2,
				},
				&model.StopPointInJourneyPattern{
					ID:                    "stop3",
					ScheduledStopPointRef: "ssp3",
					Order:                 3,
				},
			},
		},
	}
}

func timePtr(timeStr string) *time.Time {
	t, _ := time.Parse("15:04:05", timeStr)
	return &t
}

func parseTimeString(timeStr string) int {
	if timeStr == "" {
		return 0
	}

	t, err := time.Parse("15:04:05", timeStr)
	if err != nil {
		return 0
	}

	return t.Hour()*3600 + t.Minute()*60 + t.Second()
}
