package producer

import (
	"testing"

	"github.com/theoremus-urban-solutions/netex-gtfs-converter/model"
)

func TestEnhancedStopTimeProducer_ProduceStopTimesForTrip(t *testing.T) {
	mockRepo := &mockNetexRepositoryWithJourneyPatterns{
		journeyPatterns: map[string]*model.JourneyPattern{
			"jp1": createTestJourneyPattern(),
		},
	}
	mockGtfsRepo := &mockGtfsRepository{}

	producer := NewEnhancedStopTimeProducer(mockRepo, mockGtfsRepo)

	serviceJourney := &model.ServiceJourney{
		ID:                "sj1",
		JourneyPatternRef: model.ServiceJourneyPatternRef{Ref: "jp1"},
		PassingTimes: &model.PassingTimes{
			TimetabledPassingTime: []model.TimetabledPassingTime{
				{
					PointInJourneyPatternRef: "stop1",
					DepartureTime:            "08:00:00",
				},
				{
					PointInJourneyPatternRef: "stop3",
					ArrivalTime:              "08:10:00",
				},
			},
		},
	}

	trip := &model.Trip{TripID: "trip1"}
	shape := &model.Shape{ShapeID: "shape1"}

	stopTimes, err := producer.ProduceStopTimesForTrip(serviceJourney, trip, shape, "Test Route")
	if err != nil {
		t.Fatalf("ProduceStopTimesForTrip failed: %v", err)
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
		if st.TripID != "trip1" {
			t.Errorf("Stop %d has wrong trip ID: %s", i+1, st.TripID)
		}
		if st.StopHeadsign != "Test Route" {
			t.Errorf("Stop %d missing headsign", i+1)
		}
	}
}

func TestEnhancedStopTimeProducer_EuropeanPickupDropoffRules(t *testing.T) {
	mockRepo := &mockNetexRepositoryWithJourneyPatterns{
		journeyPatterns: map[string]*model.JourneyPattern{
			"jp1": createTestJourneyPattern(),
		},
	}
	mockGtfsRepo := &mockGtfsRepository{}

	producer := NewEnhancedStopTimeProducer(mockRepo, mockGtfsRepo)

	serviceJourney := &model.ServiceJourney{
		ID:                "sj1",
		JourneyPatternRef: model.ServiceJourneyPatternRef{Ref: "jp1"},
		PassingTimes: &model.PassingTimes{
			TimetabledPassingTime: []model.TimetabledPassingTime{
				{
					PointInJourneyPatternRef: "stop1",
					DepartureTime:            "08:00:00",
				},
				{
					PointInJourneyPatternRef: "stop2",
					ArrivalTime:              "08:05:00",
					DepartureTime:            "08:05:30",
				},
				{
					PointInJourneyPatternRef: "stop3",
					ArrivalTime:              "08:10:00",
				},
			},
		},
	}

	trip := &model.Trip{TripID: "trip1"}

	stopTimes, err := producer.ProduceStopTimesForTrip(serviceJourney, trip, nil, "Test Route")
	if err != nil {
		t.Fatalf("ProduceStopTimesForTrip failed: %v", err)
	}

	// Check European pickup/dropoff rules
	// First stop should not allow drop-off
	if stopTimes[0].DropOffType != "1" {
		t.Error("First stop should not allow drop-off")
	}

	// Last stop should not allow pickup
	lastStopIndex := len(stopTimes) - 1
	if stopTimes[lastStopIndex].PickupType != "1" {
		t.Error("Last stop should not allow pickup")
	}

	// Middle stops should allow both
	if len(stopTimes) > 2 {
		middleStop := stopTimes[1]
		if middleStop.PickupType != "0" || middleStop.DropOffType != "0" {
			t.Error("Middle stops should allow both pickup and drop-off")
		}
	}
}

func TestEnhancedStopTimeProducer_InterpolateStopTimesWithConstraints(t *testing.T) {
	producer := NewEnhancedStopTimeProducer(nil, nil)

	stopTimes := []*model.StopTime{
		{
			TripID:        "trip1",
			StopSequence:  1,
			StopID:        "stop1",
			ArrivalTime:   "08:00:00",
			DepartureTime: "08:00:30",
		},
		{
			TripID:       "trip1",
			StopSequence: 2,
			StopID:       "stop2",
			// Missing times - should be interpolated
		},
		{
			TripID:        "trip1",
			StopSequence:  3,
			StopID:        "stop3",
			ArrivalTime:   "08:10:00",
			DepartureTime: "08:10:30",
		},
	}

	constraints := DefaultInterpolationConstraints()

	err := producer.InterpolateStopTimesWithConstraints(stopTimes, constraints)
	if err != nil {
		t.Fatalf("InterpolateStopTimesWithConstraints failed: %v", err)
	}

	// Check that missing times were filled
	middleStop := stopTimes[1]
	if middleStop.ArrivalTime == "" {
		t.Error("Middle stop arrival time should be interpolated")
	}
	if middleStop.DepartureTime == "" {
		t.Error("Middle stop departure time should be interpolated")
	}

	// Check time progression
	stop1Departure := parseTimeString(stopTimes[0].DepartureTime)
	stop2Arrival := parseTimeString(stopTimes[1].ArrivalTime)
	stop2Departure := parseTimeString(stopTimes[1].DepartureTime)
	stop3Arrival := parseTimeString(stopTimes[2].ArrivalTime)

	if stop2Arrival <= stop1Departure {
		t.Error("Stop 2 arrival should be after stop 1 departure")
	}
	if stop2Departure <= stop2Arrival {
		t.Error("Stop 2 departure should be after stop 2 arrival")
	}
	if stop3Arrival <= stop2Departure {
		t.Error("Stop 3 arrival should be after stop 2 departure")
	}
}

func TestDefaultInterpolationConstraints(t *testing.T) {
	constraints := DefaultInterpolationConstraints()

	// Check that constraints have reasonable values
	if constraints.MaxSpeed <= 0 || constraints.MaxSpeed > 200 {
		t.Errorf("Max speed should be reasonable: %f", constraints.MaxSpeed)
	}

	if constraints.MinSpeed <= 0 || constraints.MinSpeed >= constraints.MaxSpeed {
		t.Errorf("Min speed should be positive and less than max: %f", constraints.MinSpeed)
	}

	if constraints.MinStopTime <= 0 || constraints.MinStopTime > 300 {
		t.Errorf("Min stop time should be reasonable: %d", constraints.MinStopTime)
	}

	if constraints.MaxStopTime <= constraints.MinStopTime {
		t.Errorf("Max stop time should be greater than min: %d", constraints.MaxStopTime)
	}

	if constraints.RushHourFactor <= 0 || constraints.RushHourFactor > 1 {
		t.Errorf("Rush hour factor should be between 0 and 1: %f", constraints.RushHourFactor)
	}
}

func TestEnhancedStopTimeProducer_FallbackToDefault(t *testing.T) {
	mockRepo := &mockNetexRepository{}
	mockGtfsRepo := &mockGtfsRepository{}

	producer := NewEnhancedStopTimeProducer(mockRepo, mockGtfsRepo)

	// Test that the enhanced producer can still handle single stop time production
	input := StopTimeInput{
		TimetabledPassingTime: &model.TimetabledPassingTime{
			PointInJourneyPatternRef: "stop1",
			ArrivalTime:              "08:00:00",
			DepartureTime:            "08:00:30",
		},
		Trip:            &model.Trip{TripID: "trip1"},
		CurrentHeadSign: "Test Route",
	}

	stopTime, err := producer.Produce(input)
	if err != nil {
		t.Fatalf("Produce failed: %v", err)
	}

	if stopTime.TripID != "trip1" {
		t.Errorf("Expected trip ID 'trip1', got '%s'", stopTime.TripID)
	}

	if stopTime.ArrivalTime != "08:00:00" {
		t.Errorf("Expected arrival time '08:00:00', got '%s'", stopTime.ArrivalTime)
	}

	if stopTime.DepartureTime != "08:00:30" {
		t.Errorf("Expected departure time '08:00:30', got '%s'", stopTime.DepartureTime)
	}
}

func TestEnhancedStopTimeProducer_ErrorHandling(t *testing.T) {
	producer := NewEnhancedStopTimeProducer(nil, nil)

	// Test with nil service journey
	_, err := producer.ProduceStopTimesForTrip(nil, &model.Trip{}, nil, "")
	if err == nil {
		t.Error("Expected error with nil service journey")
	}

	// Test with nil trip
	_, err = producer.ProduceStopTimesForTrip(&model.ServiceJourney{}, nil, nil, "")
	if err == nil {
		t.Error("Expected error with nil trip")
	}

	// Test interpolation with no reference times
	stopTimes := []*model.StopTime{
		{StopSequence: 1, StopID: "stop1"},
		{StopSequence: 2, StopID: "stop2"},
	}

	constraints := DefaultInterpolationConstraints()
	err = producer.InterpolateStopTimesWithConstraints(stopTimes, constraints)
	if err == nil {
		t.Error("Expected error when no reference times available")
	}
}
