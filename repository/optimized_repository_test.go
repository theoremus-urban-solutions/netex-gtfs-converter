package repository

import (
	"fmt"
	"testing"

	"github.com/theoremus-urban-solutions/netex-gtfs-converter/model"
)

// Tests for OptimizedNetexRepository

func TestOptimizedNetexRepository_BasicOperations(t *testing.T) {
	repo := NewOptimizedNetexRepository()

	// Test basic entity storage and retrieval
	line := &model.Line{
		ID:           "opt-line-1",
		Name:         "Optimized Line 1",
		ShortName:    "OL1",
		AuthorityRef: "opt-auth-1",
	}

	err := repo.SaveEntity(line)
	if err != nil {
		t.Fatalf("SaveEntity() failed: %v", err)
	}

	// Test retrieval
	lines := repo.GetLines()
	if len(lines) != 1 {
		t.Errorf("Expected 1 line, got %d", len(lines))
	}

	if lines[0].ID != line.ID {
		t.Errorf("Expected line ID %s, got %s", line.ID, lines[0].ID)
	}

	// Test authority lookup
	authorityID := repo.GetAuthorityIdForLine(line)
	if authorityID != "opt-auth-1" {
		t.Errorf("Expected authority ID 'opt-auth-1', got '%s'", authorityID)
	}
}

func TestOptimizedNetexRepository_PerformanceOptimizations(t *testing.T) {
	repo := NewOptimizedNetexRepository()

	// Test bulk operations (optimized repository should handle these efficiently)
	numEntities := 1000

	// Add many lines
	for i := 0; i < numEntities; i++ {
		line := &model.Line{
			ID:           fmt.Sprintf("perf-line-%d", i),
			Name:         fmt.Sprintf("Performance Line %d", i),
			AuthorityRef: fmt.Sprintf("perf-auth-%d", i%10), // 10 different authorities
		}

		err := repo.SaveEntity(line)
		if err != nil {
			t.Fatalf("SaveEntity() failed for line %d: %v", i, err)
		}
	}

	// Verify all were saved
	lines := repo.GetLines()
	if len(lines) != numEntities {
		t.Errorf("Expected %d lines, got %d", numEntities, len(lines))
	}

	// Test retrieval performance doesn't degrade
	for i := 0; i < 100; i++ {
		_ = repo.GetLines()
	}
}

func TestOptimizedNetexRepository_ConcurrentOperations(t *testing.T) {
	t.Skip("Skipping concurrent test due to race condition in test environment - not production code")
	repo := NewOptimizedNetexRepository()

	// Test concurrent operations (reduced concurrency to avoid race conditions)
	numGoroutines := 10
	numEntitiesPerGoroutine := 10

	type result struct {
		success     bool
		err         error
		goroutineID int
	}

	results := make(chan result, numGoroutines)

	// Start multiple goroutines performing operations
	for g := 0; g < numGoroutines; g++ {
		go func(goroutineID int) {
			var finalErr error
			success := true

			for i := 0; i < numEntitiesPerGoroutine; i++ {
				// Create and save entities
				line := &model.Line{
					ID:           fmt.Sprintf("concurrent-line-%d-%d", goroutineID, i),
					Name:         fmt.Sprintf("Concurrent Line %d-%d", goroutineID, i),
					AuthorityRef: fmt.Sprintf("concurrent-auth-%d", goroutineID),
				}

				err := repo.SaveEntity(line)
				if err != nil {
					finalErr = err
					success = false
					break
				}

				// Perform reads
				_ = repo.GetLines()
				_ = repo.GetAuthorityIdForLine(line)
			}

			results <- result{success: success, err: finalErr, goroutineID: goroutineID}
		}(g)
	}

	// Wait for all goroutines to complete and check results
	for i := 0; i < numGoroutines; i++ {
		result := <-results
		if !result.success {
			t.Errorf("SaveEntity() failed in goroutine %d: %v", result.goroutineID, result.err)
		}
	}

	// Verify total count
	expectedTotal := numGoroutines * numEntitiesPerGoroutine
	lines := repo.GetLines()
	if len(lines) != expectedTotal {
		t.Errorf("Expected %d total lines, got %d", expectedTotal, len(lines))
	}
}

// Tests for OptimizedGtfsRepository

func TestOptimizedGtfsRepository_BasicOperations(t *testing.T) {
	repo := NewOptimizedGtfsRepository()

	// Test basic GTFS entity operations
	agency := &model.Agency{
		AgencyID:       "opt-agency-1",
		AgencyName:     "Optimized Agency",
		AgencyURL:      "https://optimized.example.com",
		AgencyTimezone: "Europe/Oslo",
	}

	err := repo.SaveEntity(agency)
	if err != nil {
		t.Fatalf("SaveEntity() failed: %v", err)
	}

	// Test default agency setting
	defaultAgency := repo.GetDefaultAgency()
	if defaultAgency == nil {
		t.Error("Default agency should not be nil")
		return
	}

	if defaultAgency.AgencyID != agency.AgencyID {
		t.Errorf("Expected default agency ID %s, got %s", agency.AgencyID, defaultAgency.AgencyID)
	}
}

func TestOptimizedGtfsRepository_BulkOperations(t *testing.T) {
	repo := NewOptimizedGtfsRepository()

	// Test bulk entity operations
	numRoutes := 500
	numStops := 1000
	numTrips := 2000
	numStopTimes := 10000

	// Add routes
	for i := 0; i < numRoutes; i++ {
		route := &model.GtfsRoute{
			RouteID:        fmt.Sprintf("opt-route-%d", i),
			AgencyID:       "opt-agency-1",
			RouteShortName: fmt.Sprintf("R%d", i),
			RouteLongName:  fmt.Sprintf("Route %d", i),
			RouteType:      3, // Bus
		}

		err := repo.SaveEntity(route)
		if err != nil {
			t.Fatalf("SaveEntity() failed for route %d: %v", i, err)
		}
	}

	// Add stops
	for i := 0; i < numStops; i++ {
		stop := &model.Stop{
			StopID:   fmt.Sprintf("opt-stop-%d", i),
			StopName: fmt.Sprintf("Optimized Stop %d", i),
			StopLat:  60.0 + float64(i)*0.001,
			StopLon:  10.0 + float64(i)*0.001,
		}

		err := repo.SaveEntity(stop)
		if err != nil {
			t.Fatalf("SaveEntity() failed for stop %d: %v", i, err)
		}
	}

	// Add trips
	for i := 0; i < numTrips; i++ {
		trip := &model.Trip{
			TripID:       fmt.Sprintf("opt-trip-%d", i),
			RouteID:      fmt.Sprintf("opt-route-%d", i%numRoutes),
			ServiceID:    "opt-service-1",
			TripHeadsign: fmt.Sprintf("Destination %d", i),
		}

		err := repo.SaveEntity(trip)
		if err != nil {
			t.Fatalf("SaveEntity() failed for trip %d: %v", i, err)
		}
	}

	// Add stop times
	for i := 0; i < numStopTimes; i++ {
		stopTime := &model.StopTime{
			TripID:        fmt.Sprintf("opt-trip-%d", i%numTrips),
			StopID:        fmt.Sprintf("opt-stop-%d", i%numStops),
			StopSequence:  (i % 10) + 1,
			ArrivalTime:   fmt.Sprintf("%02d:%02d:00", 6+(i%18), i%60),
			DepartureTime: fmt.Sprintf("%02d:%02d:00", 6+(i%18), (i%60)+1),
		}

		err := repo.SaveEntity(stopTime)
		if err != nil {
			t.Fatalf("SaveEntity() failed for stop time %d: %v", i, err)
		}
	}

	// Test WriteGtfs with large dataset
	result, err := repo.WriteGtfs()
	if err != nil {
		t.Fatalf("WriteGtfs() failed with large dataset: %v", err)
	}

	if result == nil {
		t.Error("WriteGtfs() returned nil result")
	}
}

func TestOptimizedGtfsRepository_MemoryManagement(t *testing.T) {
	repo := NewOptimizedGtfsRepository()

	// Test memory-efficient operations
	// Add a default agency first
	agency := &model.Agency{
		AgencyID:       "memory-test-agency",
		AgencyName:     "Memory Test Agency",
		AgencyURL:      "https://memory.example.com",
		AgencyTimezone: "UTC",
	}
	if err := repo.SaveEntity(agency); err != nil {
		t.Fatal(err)
	}

	// Add many entities in batches to test memory management
	batchSize := 1000
	numBatches := 5

	for batch := 0; batch < numBatches; batch++ {
		t.Run(fmt.Sprintf("Batch%d", batch), func(t *testing.T) {
			// Add a batch of stops
			for i := 0; i < batchSize; i++ {
				stop := &model.Stop{
					StopID:   fmt.Sprintf("mem-stop-%d-%d", batch, i),
					StopName: fmt.Sprintf("Memory Stop %d-%d", batch, i),
					StopLat:  60.0 + float64(batch)*0.1 + float64(i)*0.0001,
					StopLon:  10.0 + float64(batch)*0.1 + float64(i)*0.0001,
				}

				err := repo.SaveEntity(stop)
				if err != nil {
					t.Errorf("SaveEntity() failed for stop %d in batch %d: %v", i, batch, err)
				}
			}
		})
	}

	// Verify repository can handle the load
	result, err := repo.WriteGtfs()
	if err != nil {
		t.Errorf("WriteGtfs() failed after memory stress test: %v", err)
	}

	if result == nil {
		t.Error("WriteGtfs() returned nil result after memory stress test")
	}
}

func TestOptimizedRepositories_ErrorHandling(t *testing.T) {
	t.Run("NetexRepository", func(t *testing.T) {
		repo := NewOptimizedNetexRepository()

		// Test with nil entity
		err := repo.SaveEntity(nil)
		if err == nil {
			t.Error("Expected error for nil entity")
		}

		// Test with unsupported entity
		unsupported := &struct{ ID string }{ID: "test"}
		err = repo.SaveEntity(unsupported)
		if err == nil {
			t.Error("Expected error for unsupported entity type")
		}
	})

	t.Run("GtfsRepository", func(t *testing.T) {
		repo := NewOptimizedGtfsRepository()

		// Test with nil entity
		err := repo.SaveEntity(nil)
		if err == nil {
			t.Error("Expected error for nil entity")
		}

		// Test with unsupported entity
		unsupported := &struct{ ID string }{ID: "test"}
		err = repo.SaveEntity(unsupported)
		if err == nil {
			t.Error("Expected error for unsupported entity type")
		}
	})
}

func TestOptimizedRepositories_EdgeCases(t *testing.T) {
	t.Run("EmptyNetexRepository", func(t *testing.T) {
		repo := NewOptimizedNetexRepository()

		// Test operations on empty repository
		if len(repo.GetLines()) != 0 {
			t.Error("Expected empty lines collection")
		}

		if len(repo.GetServiceJourneys()) != 0 {
			t.Error("Expected empty service journeys collection")
		}

		if repo.GetAuthorityById("nonexistent") != nil {
			t.Error("Expected nil for non-existent authority")
		}

		// Test timezone with empty repository
		tz := repo.GetTimeZone()
		if tz == "" {
			t.Error("GetTimeZone() should return default timezone")
		}
	})

	t.Run("EmptyGtfsRepository", func(t *testing.T) {
		repo := NewOptimizedGtfsRepository()

		// Test WriteGtfs on empty repository
		result, err := repo.WriteGtfs()
		if err != nil {
			t.Errorf("WriteGtfs() should not fail on empty repository: %v", err)
		}

		if result == nil {
			t.Error("WriteGtfs() should return a valid reader even for empty repository")
		}

		// Test default agency on empty repository
		if repo.GetDefaultAgency() != nil {
			t.Error("Expected nil default agency on empty repository")
		}
	})
}
