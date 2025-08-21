package benchmark

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/theoremus-urban-solutions/netex-gtfs-converter/memory"
	"github.com/theoremus-urban-solutions/netex-gtfs-converter/model"
	"github.com/theoremus-urban-solutions/netex-gtfs-converter/producer"
	"github.com/theoremus-urban-solutions/netex-gtfs-converter/repository"
)

// MemoryBenchmarkSuite provides memory optimization benchmarks
type MemoryBenchmarkSuite struct {
	results []BenchmarkResults
}

// NewMemoryBenchmarkSuite creates a new memory benchmark suite
func NewMemoryBenchmarkSuite() *MemoryBenchmarkSuite {
	return &MemoryBenchmarkSuite{
		results: make([]BenchmarkResults, 0),
	}
}

// RunMemoryOptimizationBenchmarks runs comprehensive memory optimization tests
func (suite *MemoryBenchmarkSuite) RunMemoryOptimizationBenchmarks() {
	fmt.Println("=== Memory Optimization Benchmark Suite ===")

	// Test 1: Compare default vs optimized repositories
	suite.benchmarkRepositoryMemoryUsage()

	// Test 2: Test batch processing efficiency
	suite.benchmarkBatchProcessing()

	// Test 3: Test object pooling efficiency
	suite.benchmarkObjectPooling()

	// Test 4: Test memory manager effectiveness
	suite.benchmarkMemoryManager()

	// Test 5: Large dataset processing
	suite.benchmarkLargeDatasetMemory()

	suite.PrintResults()
}

// benchmarkRepositoryMemoryUsage compares memory usage between default and optimized repositories
func (suite *MemoryBenchmarkSuite) benchmarkRepositoryMemoryUsage() {
	// Benchmark default repository
	suite.benchmarkDefaultRepository()

	// Benchmark optimized repository
	suite.benchmarkOptimizedRepository()
}

func (suite *MemoryBenchmarkSuite) benchmarkDefaultRepository() {
	start := time.Now()
	var memStats runtime.MemStats

	runtime.GC()
	runtime.ReadMemStats(&memStats)
	initialMem := memStats.Alloc

	// Create default repository and populate with test data
	repo := repository.NewDefaultNetexRepository()
	entities := createLargeNetexDataset(10000) // 10k entities

	for _, entity := range entities {
		if err := repo.SaveEntity(entity); err != nil {
			// ignore in benchmark population
			_ = err // Suppress unused variable warning
		}
	}

	duration := time.Since(start)
	runtime.ReadMemStats(&memStats)

	result := BenchmarkResults{
		TestName:        "DefaultRepository_10k",
		Duration:        duration,
		MemoryAllocated: memStats.Alloc - initialMem,
		ItemsProcessed:  len(entities),
		ItemsPerSecond:  float64(len(entities)) / duration.Seconds(),
		PeakMemoryUsage: memStats.Alloc,
		GoroutineCount:  runtime.NumGoroutine(),
	}

	suite.results = append(suite.results, result)
}

func (suite *MemoryBenchmarkSuite) benchmarkOptimizedRepository() {
	start := time.Now()
	var memStats runtime.MemStats

	runtime.GC()
	runtime.ReadMemStats(&memStats)
	initialMem := memStats.Alloc

	// Create optimized repository and populate with test data
	repo := repository.NewOptimizedNetexRepository()
	entities := createLargeNetexDataset(10000) // 10k entities

	// Use batch processing
	if optimizedRepo, ok := repo.(*repository.OptimizedNetexRepository); ok {
		if err := optimizedRepo.SaveEntitiesBatch(entities); err != nil {
			// ignore in benchmark population
			_ = err // Suppress unused variable warning
		}
	} else {
		for _, entity := range entities {
			if err := repo.SaveEntity(entity); err != nil {
				// ignore in benchmark population
				_ = err // Suppress unused variable warning
			}
		}
	}

	duration := time.Since(start)
	runtime.ReadMemStats(&memStats)

	result := BenchmarkResults{
		TestName:        "OptimizedRepository_10k",
		Duration:        duration,
		MemoryAllocated: memStats.Alloc - initialMem,
		ItemsProcessed:  len(entities),
		ItemsPerSecond:  float64(len(entities)) / duration.Seconds(),
		PeakMemoryUsage: memStats.Alloc,
		GoroutineCount:  runtime.NumGoroutine(),
	}

	suite.results = append(suite.results, result)
}

// benchmarkBatchProcessing tests batch processing efficiency
func (suite *MemoryBenchmarkSuite) benchmarkBatchProcessing() {
	batchSizes := []int{100, 500, 1000, 2000}

	for _, batchSize := range batchSizes {
		suite.benchmarkSpecificBatchSize(batchSize)
	}
}

func (suite *MemoryBenchmarkSuite) benchmarkSpecificBatchSize(batchSize int) {
	start := time.Now()
	var memStats runtime.MemStats

	runtime.GC()
	runtime.ReadMemStats(&memStats)
	initialMem := memStats.Alloc

	// Create memory manager with specific batch size
	memManager := memory.NewMemoryManager()
	memManager.SetBatchSize(batchSize)

	batchProcessor := memory.NewBatchProcessor(memManager)
	entities := createLargeNetexDataset(10000)

	processedCount := 0
	_ = batchProcessor.ProcessInBatches(entities, func(batch []interface{}) error {
		processedCount += len(batch)
		return nil
	})

	duration := time.Since(start)
	runtime.ReadMemStats(&memStats)

	result := BenchmarkResults{
		TestName:        fmt.Sprintf("BatchProcessing_%d", batchSize),
		Duration:        duration,
		MemoryAllocated: memStats.Alloc - initialMem,
		ItemsProcessed:  processedCount,
		ItemsPerSecond:  float64(processedCount) / duration.Seconds(),
		PeakMemoryUsage: memStats.Alloc,
		GoroutineCount:  runtime.NumGoroutine(),
	}

	suite.results = append(suite.results, result)
}

// benchmarkObjectPooling tests object pooling efficiency
func (suite *MemoryBenchmarkSuite) benchmarkObjectPooling() {
	// Benchmark without pooling
	suite.benchmarkWithoutPooling()

	// Benchmark with pooling
	suite.benchmarkWithPooling()
}

func (suite *MemoryBenchmarkSuite) benchmarkWithoutPooling() {
	start := time.Now()
	var memStats runtime.MemStats

	runtime.GC()
	runtime.ReadMemStats(&memStats)
	initialMem := memStats.Alloc

	// Create objects without pooling
	created := 0
	for i := 0; i < 50000; i++ {
		stopTime := &model.StopTime{
			TripID:       fmt.Sprintf("trip_%d", i),
			StopID:       fmt.Sprintf("stop_%d", i%1000),
			StopSequence: i%50 + 1,
		}
		// Simulate usage
		_ = stopTime.TripID
		created++
	}

	duration := time.Since(start)
	runtime.ReadMemStats(&memStats)

	result := BenchmarkResults{
		TestName:        "WithoutPooling_50k",
		Duration:        duration,
		MemoryAllocated: memStats.Alloc - initialMem,
		ItemsProcessed:  created,
		ItemsPerSecond:  float64(created) / duration.Seconds(),
		PeakMemoryUsage: memStats.Alloc,
		GoroutineCount:  runtime.NumGoroutine(),
	}

	suite.results = append(suite.results, result)
}

func (suite *MemoryBenchmarkSuite) benchmarkWithPooling() {
	start := time.Now()
	var memStats runtime.MemStats

	runtime.GC()
	runtime.ReadMemStats(&memStats)
	initialMem := memStats.Alloc

	// Create optimized GTFS repository with pooling
	repo := repository.NewOptimizedGtfsRepository()

	// Simulate creating and reusing objects through the pool
	created := 0
	for i := 0; i < 50000; i++ {
		if optimizedRepo, ok := repo.(*repository.OptimizedGtfsRepository); ok {
			stopTime := optimizedRepo.GetFromPool("StopTime").(*model.StopTime)

			// Use the object
			stopTime.TripID = fmt.Sprintf("trip_%d", i)
			stopTime.StopID = fmt.Sprintf("stop_%d", i%1000)
			stopTime.StopSequence = i%50 + 1

			// Return to pool
			optimizedRepo.ReturnToPool("StopTime", stopTime)
			created++
		}
	}

	duration := time.Since(start)
	runtime.ReadMemStats(&memStats)

	result := BenchmarkResults{
		TestName:        "WithPooling_50k",
		Duration:        duration,
		MemoryAllocated: memStats.Alloc - initialMem,
		ItemsProcessed:  created,
		ItemsPerSecond:  float64(created) / duration.Seconds(),
		PeakMemoryUsage: memStats.Alloc,
		GoroutineCount:  runtime.NumGoroutine(),
	}

	suite.results = append(suite.results, result)
}

// benchmarkMemoryManager tests memory manager effectiveness
func (suite *MemoryBenchmarkSuite) benchmarkMemoryManager() {
	start := time.Now()
	var memStats runtime.MemStats

	runtime.GC()
	runtime.ReadMemStats(&memStats)
	initialMem := memStats.Alloc

	memManager := memory.NewMemoryManager()
	memManager.SetMemoryLimit(100) // 100MB limit
	memManager.SetGCInterval(time.Millisecond * 100)

	// Simulate heavy memory usage with automatic cleanup
	allocatedObjects := 0
	for i := 0; i < 1000; i++ {
		// Create large objects
		entities := createLargeNetexDataset(1000)
		allocatedObjects += len(entities)

		// Check and force cleanup
		if memManager.CheckMemoryPressure() {
			memManager.ForceGC()
		}

		// Simulate processing
		for range entities {
			// Process...
		}
	}

	duration := time.Since(start)
	runtime.ReadMemStats(&memStats)

	result := BenchmarkResults{
		TestName:        "MemoryManager_1M",
		Duration:        duration,
		MemoryAllocated: memStats.Alloc - initialMem,
		ItemsProcessed:  allocatedObjects,
		ItemsPerSecond:  float64(allocatedObjects) / duration.Seconds(),
		PeakMemoryUsage: memStats.Alloc,
		GoroutineCount:  runtime.NumGoroutine(),
	}

	suite.results = append(suite.results, result)
}

// benchmarkLargeDatasetMemory tests memory efficiency with very large datasets
func (suite *MemoryBenchmarkSuite) benchmarkLargeDatasetMemory() {
	start := time.Now()
	var memStats runtime.MemStats

	runtime.GC()
	runtime.ReadMemStats(&memStats)
	initialMem := memStats.Alloc

	// Create optimized components
	netexRepo := repository.NewOptimizedNetexRepository()
	gtfsRepo := repository.NewOptimizedGtfsRepository()
	stopTimeProducer := producer.NewOptimizedStopTimeProducer(netexRepo, gtfsRepo)

	// Generate large service journey
	serviceJourney := createLargeServiceJourney(1000) // 1000 stops
	trip := &model.Trip{TripID: "large_trip_test"}

	// Process with optimized producer
	stopTimes, err := stopTimeProducer.ProduceStopTimesForTrip(serviceJourney, trip, nil, "Large Test")
	processedItems := 0
	if err == nil {
		processedItems = len(stopTimes)
	}

	duration := time.Since(start)
	runtime.ReadMemStats(&memStats)

	result := BenchmarkResults{
		TestName:        "LargeDataset_1000stops",
		Duration:        duration,
		MemoryAllocated: memStats.Alloc - initialMem,
		ItemsProcessed:  processedItems,
		ItemsPerSecond:  float64(processedItems) / duration.Seconds(),
		PeakMemoryUsage: memStats.Alloc,
		GoroutineCount:  runtime.NumGoroutine(),
	}

	suite.results = append(suite.results, result)
}

// GetResults returns all benchmark results
func (suite *MemoryBenchmarkSuite) GetResults() []BenchmarkResults {
	return suite.results
}

// PrintResults prints formatted benchmark results
func (suite *MemoryBenchmarkSuite) PrintResults() {
	fmt.Println("\n=== Memory Optimization Benchmark Results ===")
	fmt.Printf("%-25s %-10s %-12s %-12s %-15s %-12s %-12s\n",
		"Test Name", "Duration", "Memory (KB)", "Items", "Items/sec", "Peak Mem (KB)", "Goroutines")
	fmt.Println(strings.Repeat("-", 130))

	for _, result := range suite.results {
		fmt.Printf("%-25s %-10s %-12d %-12d %-15.2f %-12d %-12d\n",
			result.TestName,
			result.Duration.Round(time.Millisecond),
			result.MemoryAllocated/1024,
			result.ItemsProcessed,
			result.ItemsPerSecond,
			result.PeakMemoryUsage/1024,
			result.GoroutineCount)
	}

	suite.printMemoryAnalysis()
}

// printMemoryAnalysis provides analysis of memory optimization results
func (suite *MemoryBenchmarkSuite) printMemoryAnalysis() {
	fmt.Println("\n=== Memory Optimization Analysis ===")

	// Find optimization improvements
	defaultRepoResult := suite.findResult("DefaultRepository_10k")
	optimizedRepoResult := suite.findResult("OptimizedRepository_10k")

	if defaultRepoResult != nil && optimizedRepoResult != nil {
		memoryImprovement := float64(defaultRepoResult.MemoryAllocated-optimizedRepoResult.MemoryAllocated) / float64(defaultRepoResult.MemoryAllocated) * 100
		speedImprovement := (optimizedRepoResult.ItemsPerSecond - defaultRepoResult.ItemsPerSecond) / defaultRepoResult.ItemsPerSecond * 100

		fmt.Printf("Repository Optimization:\n")
		fmt.Printf("  Memory Reduction: %.2f%%\n", memoryImprovement)
		fmt.Printf("  Speed Improvement: %.2f%%\n", speedImprovement)
	}

	// Find pooling improvements
	withoutPoolResult := suite.findResult("WithoutPooling_50k")
	withPoolResult := suite.findResult("WithPooling_50k")

	if withoutPoolResult != nil && withPoolResult != nil {
		poolMemoryImprovement := float64(withoutPoolResult.MemoryAllocated-withPoolResult.MemoryAllocated) / float64(withoutPoolResult.MemoryAllocated) * 100
		poolSpeedImprovement := (withPoolResult.ItemsPerSecond - withoutPoolResult.ItemsPerSecond) / withoutPoolResult.ItemsPerSecond * 100

		fmt.Printf("\nObject Pooling Optimization:\n")
		fmt.Printf("  Memory Reduction: %.2f%%\n", poolMemoryImprovement)
		fmt.Printf("  Speed Improvement: %.2f%%\n", poolSpeedImprovement)
	}

	// Find best batch size
	bestBatchSize := suite.findBestBatchSize()
	if bestBatchSize > 0 {
		fmt.Printf("\nOptimal Batch Size: %d\n", bestBatchSize)
	}
}

func (suite *MemoryBenchmarkSuite) findResult(testName string) *BenchmarkResults {
	for _, result := range suite.results {
		if result.TestName == testName {
			return &result
		}
	}
	return nil
}

func (suite *MemoryBenchmarkSuite) findBestBatchSize() int {
	bestThroughput := 0.0
	bestSize := 0

	for _, result := range suite.results {
		if strings.HasPrefix(result.TestName, "BatchProcessing_") && result.ItemsPerSecond > bestThroughput {
			bestThroughput = result.ItemsPerSecond
			if n, err := fmt.Sscanf(result.TestName, "BatchProcessing_%d", &bestSize); n != 1 || err != nil {
				continue
			}
		}
	}

	return bestSize
}
