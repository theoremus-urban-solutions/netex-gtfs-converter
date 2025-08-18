package producer

import (
	"testing"

	"github.com/theoremus-urban-solutions/netex-gtfs-converter/model"
)

func TestSophisticatedInterchangeProducer_ProduceComplexInterchanges(t *testing.T) {
	mockNetexRepo := &mockNetexRepository{}
	mockGtfsRepo := &mockGtfsRepository{}
	mockPathwaysProducer := &mockPathwaysProducer{}

	producer := NewSophisticatedInterchangeProducer(mockNetexRepo, mockGtfsRepo, mockPathwaysProducer)

	// Create complex multi-modal station
	stopPlace := &model.StopPlace{
		ID:   "complex_station",
		Name: "Central Interchange",
		Centroid: &model.Centroid{
			Location: &model.Location{
				Latitude:  59.911491,
				Longitude: 10.757933,
			},
		},
		Quays: &model.Quays{
			Quay: []model.Quay{
				{
					ID:   "train_platform_1",
					Name: "Train Platform 1 Ground Level",
					Centroid: &model.Centroid{
						Location: &model.Location{
							Latitude:  59.911491,
							Longitude: 10.757933,
						},
					},
				},
				{
					ID:   "metro_platform_a",
					Name: "Metro Platform A Lower Level",
					Centroid: &model.Centroid{
						Location: &model.Location{
							Latitude:  59.911391,
							Longitude: 10.757833,
						},
					},
				},
				{
					ID:   "bus_bay_1",
					Name: "Bus Bay 1 Ground Level",
					Centroid: &model.Centroid{
						Location: &model.Location{
							Latitude:  59.911591,
							Longitude: 10.758033,
						},
					},
					AccessibilityAssessment: &model.AccessibilityAssessment{
						Limitations: &model.Limitations{
							AccessibilityLimitation: &model.AccessibilityLimitation{
								WheelchairAccess: "true",
								StepFreeAccess:   "true",
							},
						},
					},
				},
				{
					ID:   "tram_stop_1",
					Name: "Tram Stop 1 Ground Level",
					Centroid: &model.Centroid{
						Location: &model.Location{
							Latitude:  59.911691,
							Longitude: 10.758133,
						},
					},
				},
			},
		},
	}

	transfers, err := producer.ProduceComplexInterchanges(stopPlace)
	if err != nil {
		t.Fatalf("ProduceComplexInterchanges failed: %v", err)
	}

	if len(transfers) == 0 {
		t.Error("Expected transfers to be generated")
	}

	// Verify that transfers exist between different modes
	hasModeTransfer := false
	hasAccessibilityTransfer := false

	for _, transfer := range transfers {
		if transfer.FromStopID == "" || transfer.ToStopID == "" {
			t.Error("Transfer should have valid stop IDs")
		}

		if transfer.MinTransferTime <= 0 {
			t.Error("Transfer should have positive minimum time")
		}

		if transfer.TransferType < 0 || transfer.TransferType > 3 {
			t.Errorf("Invalid transfer type: %d", transfer.TransferType)
		}

		// Check for mode transfers (e.g., train to bus)
		if (transfer.FromStopID == "train_platform_1" && transfer.ToStopID == "bus_bay_1") ||
			(transfer.FromStopID == "bus_bay_1" && transfer.ToStopID == "train_platform_1") {
			hasModeTransfer = true
		}

		// Check for accessibility transfers
		if transfer.FromStopID == "bus_bay_1" || transfer.ToStopID == "bus_bay_1" {
			hasAccessibilityTransfer = true
		}
	}

	if !hasModeTransfer {
		t.Error("Expected mode transfer between train and bus")
	}

	if !hasAccessibilityTransfer {
		t.Error("Expected accessibility transfer involving accessible bus bay")
	}
}

func TestSophisticatedInterchangeProducer_ModeSpecificTransfers(t *testing.T) {
	mockNetexRepo := &mockNetexRepository{}
	mockGtfsRepo := &mockGtfsRepository{}
	producer := NewSophisticatedInterchangeProducer(mockNetexRepo, mockGtfsRepo, nil)

	stopPlace := &model.StopPlace{
		ID:   "mode_test_station",
		Name: "Mode Test Station",
		Quays: &model.Quays{
			Quay: []model.Quay{
				{
					ID:   "train_1",
					Name: "Train Platform 1",
				},
				{
					ID:   "bus_1",
					Name: "Bus Bay 1",
				},
				{
					ID:   "metro_1",
					Name: "Metro Platform A",
				},
			},
		},
	}

	transfers, err := producer.generateModeSpecificTransfers(stopPlace)
	if err != nil {
		t.Fatalf("generateModeSpecificTransfers failed: %v", err)
	}

	// Should have transfers between all different mode combinations
	expectedTransfers := map[string]bool{
		"train_1_bus_1":   false,
		"train_1_metro_1": false,
		"bus_1_train_1":   false,
		"bus_1_metro_1":   false,
		"metro_1_train_1": false,
		"metro_1_bus_1":   false,
	}

	for _, transfer := range transfers {
		key := transfer.FromStopID + "_" + transfer.ToStopID
		if _, exists := expectedTransfers[key]; exists {
			expectedTransfers[key] = true
		}
	}

	for key, found := range expectedTransfers {
		if !found {
			t.Errorf("Expected transfer not found: %s", key)
		}
	}
}

func TestSophisticatedInterchangeProducer_AccessibilityTransfers(t *testing.T) {
	mockNetexRepo := &mockNetexRepository{}
	mockGtfsRepo := &mockGtfsRepository{}
	producer := NewSophisticatedInterchangeProducer(mockNetexRepo, mockGtfsRepo, nil)

	stopPlace := &model.StopPlace{
		ID:   "accessibility_test_station",
		Name: "Accessibility Test Station",
		Quays: &model.Quays{
			Quay: []model.Quay{
				{
					ID:   "accessible_1",
					Name: "Accessible Platform 1",
					AccessibilityAssessment: &model.AccessibilityAssessment{
						Limitations: &model.Limitations{
							AccessibilityLimitation: &model.AccessibilityLimitation{
								WheelchairAccess: "true",
								StepFreeAccess:   "true",
							},
						},
					},
				},
				{
					ID:   "accessible_2",
					Name: "Accessible Platform 2",
					AccessibilityAssessment: &model.AccessibilityAssessment{
						Limitations: &model.Limitations{
							AccessibilityLimitation: &model.AccessibilityLimitation{
								WheelchairAccess: "true",
								StepFreeAccess:   "true",
							},
						},
					},
				},
				{
					ID:   "regular_1",
					Name: "Regular Platform 1",
					AccessibilityAssessment: &model.AccessibilityAssessment{
						Limitations: &model.Limitations{
							AccessibilityLimitation: &model.AccessibilityLimitation{
								WheelchairAccess: "false",
								StepFreeAccess:   "false",
							},
						},
					},
				},
			},
		},
	}

	transfers, err := producer.generateAccessibilityTransfers(stopPlace)
	if err != nil {
		t.Fatalf("generateAccessibilityTransfers failed: %v", err)
	}

	// Check for accessible-to-accessible transfers
	hasAccessibleTransfer := false
	hasFallbackTransfer := false

	for _, transfer := range transfers {
		if (transfer.FromStopID == "accessible_1" && transfer.ToStopID == "accessible_2") ||
			(transfer.FromStopID == "accessible_2" && transfer.ToStopID == "accessible_1") {
			hasAccessibleTransfer = true
			if transfer.TransferType != 0 { // Should be recommended
				t.Error("Accessible transfers should be recommended type")
			}
		}

		if (transfer.FromStopID == "accessible_1" && transfer.ToStopID == "regular_1") ||
			(transfer.FromStopID == "accessible_2" && transfer.ToStopID == "regular_1") {
			hasFallbackTransfer = true
		}
	}

	if !hasAccessibleTransfer {
		t.Error("Expected transfer between accessible platforms")
	}

	if !hasFallbackTransfer {
		t.Error("Expected fallback transfer from accessible to regular platform")
	}
}

func TestSophisticatedInterchangeProducer_TransferTimeCalculation(t *testing.T) {
	mockNetexRepo := &mockNetexRepository{}
	mockGtfsRepo := &mockGtfsRepository{}
	producer := NewSophisticatedInterchangeProducer(mockNetexRepo, mockGtfsRepo, nil)

	// Test quays at different distances
	nearQuay := &model.Quay{
		ID:   "near_quay",
		Name: "Near Platform",
		Centroid: &model.Centroid{
			Location: &model.Location{
				Latitude:  59.911491,
				Longitude: 10.757933,
			},
		},
	}

	farQuay := &model.Quay{
		ID:   "far_quay",
		Name: "Far Platform",
		Centroid: &model.Centroid{
			Location: &model.Location{
				Latitude:  59.912491, // ~111m away
				Longitude: 10.757933,
			},
		},
	}

	veryFarQuay := &model.Quay{
		ID:   "very_far_quay",
		Name: "Very Far Platform",
		Centroid: &model.Centroid{
			Location: &model.Location{
				Latitude:  59.916491, // ~555m away
				Longitude: 10.757933,
			},
		},
	}

	stopPlace := &model.StopPlace{ID: "test_station"}

	// Test near transfer
	nearTime := producer.calculateBaseTransferTime(nearQuay, farQuay, stopPlace)
	if nearTime < producer.minimumConnectionTime {
		t.Error("Transfer time should respect minimum connection time")
	}

	// Test far transfer
	farTime := producer.calculateBaseTransferTime(nearQuay, veryFarQuay, stopPlace)
	if farTime <= nearTime {
		t.Error("Longer distance should result in longer transfer time")
	}

	// Test maximum distance handling
	if farTime > producer.maximumConnectionTime {
		t.Error("Very far transfer should be limited by maximum connection time")
	}
}

func TestSophisticatedInterchangeProducer_LevelChangeCalculation(t *testing.T) {
	mockNetexRepo := &mockNetexRepository{}
	mockGtfsRepo := &mockGtfsRepository{}
	producer := NewSophisticatedInterchangeProducer(mockNetexRepo, mockGtfsRepo, nil)

	groundQuay := &model.Quay{
		ID:   "ground_quay",
		Name: "Ground Level Platform",
	}

	upperQuay := &model.Quay{
		ID:   "upper_quay",
		Name: "Upper Level Platform",
	}

	lowerQuay := &model.Quay{
		ID:   "lower_quay",
		Name: "Lower Level Platform",
	}

	// Test same level
	sameLevelDiff := producer.calculateLevelDifference(groundQuay, groundQuay)
	if sameLevelDiff != 0 {
		t.Errorf("Same level difference should be 0, got %f", sameLevelDiff)
	}

	// Test different levels
	upLevelDiff := producer.calculateLevelDifference(groundQuay, upperQuay)
	if upLevelDiff != 1.0 {
		t.Errorf("Up level difference should be 1.0, got %f", upLevelDiff)
	}

	downLevelDiff := producer.calculateLevelDifference(groundQuay, lowerQuay)
	if downLevelDiff != 1.0 {
		t.Errorf("Down level difference should be 1.0, got %f", downLevelDiff)
	}

	// Test multi-level
	multiLevelDiff := producer.calculateLevelDifference(upperQuay, lowerQuay)
	if multiLevelDiff != 2.0 {
		t.Errorf("Multi-level difference should be 2.0, got %f", multiLevelDiff)
	}
}

func TestSophisticatedInterchangeProducer_TransferTypesDetermination(t *testing.T) {
	mockNetexRepo := &mockNetexRepository{}
	mockGtfsRepo := &mockGtfsRepository{}
	producer := NewSophisticatedInterchangeProducer(mockNetexRepo, mockGtfsRepo, nil)

	// Cross-platform transfer (should be recommended)
	platform1A := &model.Quay{
		ID:   "platform_1a",
		Name: "Platform 1A",
		Centroid: &model.Centroid{
			Location: &model.Location{Latitude: 59.911491, Longitude: 10.757933},
		},
	}

	platform1B := &model.Quay{
		ID:   "platform_1b",
		Name: "Platform 1B",
		Centroid: &model.Centroid{
			Location: &model.Location{Latitude: 59.911491, Longitude: 10.757933},
		},
	}

	stopPlace := &model.StopPlace{ID: "test_station"}

	transferType := producer.determineTransferType(platform1A, platform1B, stopPlace)
	if transferType != 0 {
		t.Errorf("Cross-platform transfer should be recommended (0), got %d", transferType)
	}

	// Very far transfer (should be not possible)
	farQuay := &model.Quay{
		ID:   "far_quay",
		Name: "Far Platform",
		Centroid: &model.Centroid{
			Location: &model.Location{Latitude: 59.920000, Longitude: 10.770000}, // Very far
		},
	}

	farTransferType := producer.determineTransferType(platform1A, farQuay, stopPlace)
	if farTransferType != 3 {
		t.Errorf("Very far transfer should not be possible (3), got %d", farTransferType)
	}
}

func TestSophisticatedInterchangeProducer_ModeTransferTimes(t *testing.T) {
	mockNetexRepo := &mockNetexRepository{}
	mockGtfsRepo := &mockGtfsRepository{}
	producer := NewSophisticatedInterchangeProducer(mockNetexRepo, mockGtfsRepo, nil)

	tests := []struct {
		fromMode string
		toMode   string
		expected int
	}{
		{"train", "bus", 300},    // 5 minutes
		{"bus", "metro", 120},    // 2 minutes
		{"metro", "tram", 90},    // 1.5 minutes
		{"unknown", "test", 120}, // Default 2 minutes
	}

	for _, tt := range tests {
		time := producer.getModeTransferTime(tt.fromMode, tt.toMode)
		if time != tt.expected {
			t.Errorf("Transfer time from %s to %s: expected %d, got %d",
				tt.fromMode, tt.toMode, tt.expected, time)
		}
	}
}

func TestSophisticatedInterchangeProducer_StationComplexity(t *testing.T) {
	mockNetexRepo := &mockNetexRepository{}
	mockGtfsRepo := &mockGtfsRepository{}
	producer := NewSophisticatedInterchangeProducer(mockNetexRepo, mockGtfsRepo, nil)

	// Simple station
	simpleStation := &model.StopPlace{
		ID:   "simple",
		Name: "Simple Station",
		Quays: &model.Quays{
			Quay: []model.Quay{
				{ID: "q1", Name: "Platform 1"},
				{ID: "q2", Name: "Platform 2"},
			},
		},
	}

	// Complex station
	complexStation := &model.StopPlace{
		ID:   "complex",
		Name: "Complex Station",
		Quays: &model.Quays{
			Quay: []model.Quay{
				{ID: "q1", Name: "Train Platform 1 Ground"},
				{ID: "q2", Name: "Train Platform 2 Ground"},
				{ID: "q3", Name: "Metro Platform A Lower"},
				{ID: "q4", Name: "Metro Platform B Lower"},
				{ID: "q5", Name: "Bus Bay 1 Ground"},
				{ID: "q6", Name: "Bus Bay 2 Ground"},
				{ID: "q7", Name: "Tram Stop 1 Upper"},
				{ID: "q8", Name: "Tram Stop 2 Upper"},
			},
		},
	}

	simpleComplexity := producer.calculateStationComplexity(simpleStation)
	complexComplexity := producer.calculateStationComplexity(complexStation)

	if complexComplexity <= simpleComplexity {
		t.Errorf("Complex station should have higher complexity factor: simple=%f, complex=%f",
			simpleComplexity, complexComplexity)
	}

	if simpleComplexity < 1.0 || complexComplexity < 1.0 {
		t.Error("Complexity factor should be at least 1.0")
	}
}

func TestSophisticatedInterchangeProducer_NilHandling(t *testing.T) {
	mockNetexRepo := &mockNetexRepository{}
	mockGtfsRepo := &mockGtfsRepository{}
	producer := NewSophisticatedInterchangeProducer(mockNetexRepo, mockGtfsRepo, nil)

	// Test with nil stop place
	transfers, err := producer.ProduceComplexInterchanges(nil)
	if err != nil {
		t.Errorf("Should handle nil stop place gracefully: %v", err)
	}
	if transfers != nil {
		t.Error("Expected nil transfers for nil stop place")
	}

	// Test with empty quays
	emptyStation := &model.StopPlace{
		ID:   "empty",
		Name: "Empty Station",
	}

	transfers, err = producer.ProduceComplexInterchanges(emptyStation)
	if err != nil {
		t.Errorf("Should handle empty station gracefully: %v", err)
	}
	if len(transfers) != 0 {
		t.Error("Expected no transfers for empty station")
	}
}

// Mock pathways producer for testing
type mockPathwaysProducer struct{}

func (m *mockPathwaysProducer) ProducePathwaysFromStopPlace(stopPlace *model.StopPlace) ([]*model.Pathway, error) {
	// Return some test pathways
	return []*model.Pathway{
		{
			PathwayID:       "test_pathway",
			FromStopID:      "train_platform_1",
			ToStopID:        "metro_platform_a",
			PathwayMode:     5, // Elevator
			TraversalTime:   45,
			IsBidirectional: 1,
		},
	}, nil
}

func (m *mockPathwaysProducer) ProduceLevelsFromStopPlace(stopPlace *model.StopPlace) ([]*model.Level, error) {
	return []*model.Level{}, nil
}

func (m *mockPathwaysProducer) ProduceAccessibilityPathways(from, to *model.Quay) (*model.Pathway, error) {
	return &model.Pathway{
		PathwayID:       "accessibility_pathway",
		FromStopID:      from.ID,
		ToStopID:        to.ID,
		PathwayMode:     1, // Walkway
		TraversalTime:   60,
		IsBidirectional: 1,
	}, nil
}
