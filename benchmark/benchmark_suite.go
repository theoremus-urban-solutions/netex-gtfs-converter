package benchmark

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/theoremus-urban-solutions/netex-gtfs-converter/model"
	"github.com/theoremus-urban-solutions/netex-gtfs-converter/producer"
	"github.com/theoremus-urban-solutions/netex-gtfs-converter/repository"
)

// RunPerformanceProfile runs a comprehensive performance profile
func (suite *BenchmarkSuite) RunPerformanceProfile() {
	suite.profileStopTimeProduction()
	suite.profileFrequencyProduction()
	suite.profilePathwaysGeneration()
	suite.profileMemoryUsage()
	suite.profileConcurrentLoad()
}

// profileStopTimeProduction profiles stop time production
func (suite *BenchmarkSuite) profileStopTimeProduction() {
	start := time.Now()
	var memStats runtime.MemStats

	runtime.GC()
	runtime.ReadMemStats(&memStats)
	initialMem := memStats.Alloc

	// Setup
	netexRepo := repository.NewDefaultNetexRepository()
	gtfsRepo := repository.NewDefaultGtfsRepository()
	producer := producer.NewEnhancedStopTimeProducer(netexRepo, gtfsRepo)

	// Test with various sizes
	sizes := []int{10, 50, 100, 500}
	totalProcessed := 0

	for _, size := range sizes {
		serviceJourney := createLargeServiceJourney(size)
		trip := &model.Trip{TripID: fmt.Sprintf("profile_trip_%d", size)}

		_, err := producer.ProduceStopTimesForTrip(serviceJourney, trip, nil, "Profile Test")
		if err == nil {
			totalProcessed += size
		}
	}

	duration := time.Since(start)
	runtime.ReadMemStats(&memStats)

	result := BenchmarkResults{
		TestName:        "StopTimeProduction",
		Duration:        duration,
		MemoryAllocated: memStats.Alloc - initialMem,
		ItemsProcessed:  totalProcessed,
		ItemsPerSecond:  float64(totalProcessed) / duration.Seconds(),
		PeakMemoryUsage: memStats.Alloc,
		GoroutineCount:  runtime.NumGoroutine(),
	}

	suite.results = append(suite.results, result)
}

// profileFrequencyProduction profiles frequency production
func (suite *BenchmarkSuite) profileFrequencyProduction() {
	start := time.Now()
	var memStats runtime.MemStats

	runtime.GC()
	runtime.ReadMemStats(&memStats)
	initialMem := memStats.Alloc

	netexRepo := repository.NewDefaultNetexRepository()
	gtfsRepo := repository.NewDefaultGtfsRepository()
	producer := producer.NewDefaultFrequencyProducer(netexRepo, gtfsRepo)

	totalProcessed := 0
	for i := 0; i < 100; i++ {
		group := createLargeHeadwayJourneyGroup(10)
		_, err := producer.ProduceFromHeadwayJourneyGroup(group)
		if err == nil {
			totalProcessed += 10
		}
	}

	duration := time.Since(start)
	runtime.ReadMemStats(&memStats)

	result := BenchmarkResults{
		TestName:        "FrequencyProduction",
		Duration:        duration,
		MemoryAllocated: memStats.Alloc - initialMem,
		ItemsProcessed:  totalProcessed,
		ItemsPerSecond:  float64(totalProcessed) / duration.Seconds(),
		PeakMemoryUsage: memStats.Alloc,
		GoroutineCount:  runtime.NumGoroutine(),
	}

	suite.results = append(suite.results, result)
}

// profilePathwaysGeneration profiles pathways generation
func (suite *BenchmarkSuite) profilePathwaysGeneration() {
	start := time.Now()
	var memStats runtime.MemStats

	runtime.GC()
	runtime.ReadMemStats(&memStats)
	initialMem := memStats.Alloc

	netexRepo := repository.NewDefaultNetexRepository()
	gtfsRepo := repository.NewDefaultGtfsRepository()
	producer := producer.NewDefaultPathwaysProducer(netexRepo, gtfsRepo)

	totalProcessed := 0
	sizes := []int{5, 10, 15, 20}

	for _, size := range sizes {
		stopPlace := createLargeStopPlace(size)
		_, err := producer.ProducePathwaysFromStopPlace(stopPlace)
		if err == nil {
			totalProcessed += size * size // N^2 pathways for N quays
		}
	}

	duration := time.Since(start)
	runtime.ReadMemStats(&memStats)

	result := BenchmarkResults{
		TestName:        "PathwaysGeneration",
		Duration:        duration,
		MemoryAllocated: memStats.Alloc - initialMem,
		ItemsProcessed:  totalProcessed,
		ItemsPerSecond:  float64(totalProcessed) / duration.Seconds(),
		PeakMemoryUsage: memStats.Alloc,
		GoroutineCount:  runtime.NumGoroutine(),
	}

	suite.results = append(suite.results, result)
}

// profileMemoryUsage profiles memory usage patterns
func (suite *BenchmarkSuite) profileMemoryUsage() {
	var memStats runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&memStats)
	initialMem := memStats.Alloc

	start := time.Now()

	// Create large dataset and measure memory growth
	netexRepo := repository.NewDefaultNetexRepository()
	entities := createLargeNetexDataset(5000)

	for i, entity := range entities {
		if err := netexRepo.SaveEntity(entity); err != nil {
			// ignore errors during profiling
			_ = err // Suppress unused variable warning
		}

		// Sample memory usage at intervals
		if i%1000 == 0 {
			runtime.ReadMemStats(&memStats)
		}
	}

	duration := time.Since(start)
	runtime.GC()
	runtime.ReadMemStats(&memStats)

	result := BenchmarkResults{
		TestName:        "MemoryUsage",
		Duration:        duration,
		MemoryAllocated: memStats.Alloc - initialMem,
		ItemsProcessed:  len(entities),
		ItemsPerSecond:  float64(len(entities)) / duration.Seconds(),
		PeakMemoryUsage: memStats.Alloc,
		GoroutineCount:  runtime.NumGoroutine(),
	}

	suite.results = append(suite.results, result)
}

// profileConcurrentLoad profiles concurrent processing load
func (suite *BenchmarkSuite) profileConcurrentLoad() {
	start := time.Now()
	var memStats runtime.MemStats

	runtime.GC()
	runtime.ReadMemStats(&memStats)
	initialMem := memStats.Alloc
	initialGoroutines := runtime.NumGoroutine()

	netexRepo := repository.NewDefaultNetexRepository()
	gtfsRepo := repository.NewDefaultGtfsRepository()

	// Simulate concurrent processing
	concurrentTasks := 10
	tasksPerWorker := 20
	totalProcessed := 0

	done := make(chan bool, concurrentTasks)

	for i := 0; i < concurrentTasks; i++ {
		go func(workerID int) {
			producer := producer.NewEnhancedStopTimeProducer(netexRepo, gtfsRepo)

			for j := 0; j < tasksPerWorker; j++ {
				serviceJourney := createLargeServiceJourney(25)
				trip := &model.Trip{TripID: fmt.Sprintf("concurrent_%d_%d", workerID, j)}
				_, err := producer.ProduceStopTimesForTrip(serviceJourney, trip, nil, "Concurrent Load Test")
				if err == nil {
					totalProcessed += 25
				}
			}
			done <- true
		}(i)
	}

	// Wait for all workers to complete
	for i := 0; i < concurrentTasks; i++ {
		<-done
	}

	duration := time.Since(start)
	runtime.ReadMemStats(&memStats)

	result := BenchmarkResults{
		TestName:        "ConcurrentLoad",
		Duration:        duration,
		MemoryAllocated: memStats.Alloc - initialMem,
		ItemsProcessed:  totalProcessed,
		ItemsPerSecond:  float64(totalProcessed) / duration.Seconds(),
		PeakMemoryUsage: memStats.Alloc,
		GoroutineCount:  runtime.NumGoroutine() - initialGoroutines,
	}

	suite.results = append(suite.results, result)
}

// GetResults returns all benchmark results
func (suite *BenchmarkSuite) GetResults() []BenchmarkResults {
	return suite.results
}

// PrintResults prints formatted benchmark results
func (suite *BenchmarkSuite) PrintResults() {
	fmt.Println("=== Performance Benchmark Results ===")
	fmt.Printf("%-20s %-10s %-12s %-12s %-15s %-12s %-12s\n",
		"Test Name", "Duration", "Memory (KB)", "Items", "Items/sec", "Peak Mem (KB)", "Goroutines")
	fmt.Println(strings.Repeat("-", 120))

	for _, result := range suite.results {
		fmt.Printf("%-20s %-10s %-12d %-12d %-15.2f %-12d %-12d\n",
			result.TestName,
			result.Duration.Round(time.Millisecond),
			result.MemoryAllocated/1024,
			result.ItemsProcessed,
			result.ItemsPerSecond,
			result.PeakMemoryUsage/1024,
			result.GoroutineCount)
	}
}
