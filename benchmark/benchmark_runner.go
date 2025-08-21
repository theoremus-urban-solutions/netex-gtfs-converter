package benchmark

import (
	"fmt"
	"log"
	"strings"
	"time"
)

const (
	statusFail = "❌"
	statusPass = "✅"
)

// BenchmarkResults holds performance measurement results
type BenchmarkResults struct {
	TestName          string
	Duration          time.Duration
	MemoryAllocated   uint64
	MemoryAllocations uint64
	ItemsProcessed    int
	ItemsPerSecond    float64
	PeakMemoryUsage   uint64
	GoroutineCount    int
}

// BenchmarkSuite runs comprehensive performance tests
type BenchmarkSuite struct {
	results []BenchmarkResults
}

// NewBenchmarkSuite creates a new benchmark suite
func NewBenchmarkSuite() *BenchmarkSuite {
	return &BenchmarkSuite{
		results: make([]BenchmarkResults, 0),
	}
}

// BenchmarkRunner provides utilities for running performance benchmarks
type BenchmarkRunner struct {
	suite *BenchmarkSuite
}

// NewBenchmarkRunner creates a new benchmark runner
func NewBenchmarkRunner() *BenchmarkRunner {
	return &BenchmarkRunner{
		suite: NewBenchmarkSuite(),
	}
}

// RunFullBenchmarkSuite runs all performance benchmarks
func (runner *BenchmarkRunner) RunFullBenchmarkSuite() {
	fmt.Println("Starting comprehensive performance benchmark suite...")
	start := time.Now()

	runner.suite.RunPerformanceProfile()

	duration := time.Since(start)
	fmt.Printf("Benchmark suite completed in %v\n\n", duration)

	runner.suite.PrintResults()
	runner.printPerformanceAnalysis()
}

// RunMemoryOptimizationSuite runs memory optimization benchmarks
func (runner *BenchmarkRunner) RunMemoryOptimizationSuite() {
	fmt.Println("Starting memory optimization benchmark suite...")
	start := time.Now()

	memorySuite := NewMemoryBenchmarkSuite()
	memorySuite.RunMemoryOptimizationBenchmarks()

	duration := time.Since(start)
	fmt.Printf("Memory optimization benchmark completed in %v\n\n", duration)
}

// printPerformanceAnalysis provides analysis of benchmark results
func (runner *BenchmarkRunner) printPerformanceAnalysis() {
	results := runner.suite.GetResults()
	if len(results) == 0 {
		return
	}

	fmt.Println("\n=== Performance Analysis ===")

	// Find performance characteristics
	var totalMemory uint64
	var totalItems int
	var maxThroughput float64
	var slowestTest string
	var fastestTest string
	maxDuration := time.Duration(0)
	minDuration := 24 * time.Hour // Large initial value

	for _, result := range results {
		totalMemory += result.MemoryAllocated
		totalItems += result.ItemsProcessed

		if result.ItemsPerSecond > maxThroughput {
			maxThroughput = result.ItemsPerSecond
		}

		if result.Duration > maxDuration {
			maxDuration = result.Duration
			slowestTest = result.TestName
		}

		if result.Duration < minDuration {
			minDuration = result.Duration
			fastestTest = result.TestName
		}
	}

	fmt.Printf("Total Memory Allocated: %.2f MB\n", float64(totalMemory)/(1024*1024))
	fmt.Printf("Total Items Processed: %d\n", totalItems)
	fmt.Printf("Maximum Throughput: %.2f items/second (%s)\n", maxThroughput, getBestPerformingTest(results))
	fmt.Printf("Fastest Test: %s (Duration: %v)\n", fastestTest, minDuration)
	fmt.Printf("Slowest Test: %s (Duration: %v)\n", slowestTest, maxDuration)

	// Performance recommendations
	fmt.Println("\n=== Performance Recommendations ===")

	for _, result := range results {
		if result.MemoryAllocated > 50*1024*1024 { // > 50 MB
			fmt.Printf("⚠️  %s: High memory usage (%.2f MB) - consider optimization\n",
				result.TestName, float64(result.MemoryAllocated)/(1024*1024))
		}

		if result.ItemsPerSecond < 100 {
			fmt.Printf("⚠️  %s: Low throughput (%.2f items/sec) - consider optimization\n",
				result.TestName, result.ItemsPerSecond)
		}

		if result.GoroutineCount > 20 {
			fmt.Printf("⚠️  %s: High goroutine count (%d) - check for leaks\n",
				result.TestName, result.GoroutineCount)
		}

		if result.Duration > 5*time.Second {
			fmt.Printf("⚠️  %s: Long execution time (%v) - consider caching or optimization\n",
				result.TestName, result.Duration)
		}
	}

	// Success indicators
	fmt.Println("\n=== Performance Indicators ===")
	allGood := true

	for _, result := range results {
		status := "✅"
		issues := []string{}

		if result.MemoryAllocated > 100*1024*1024 { // > 100 MB
			status = statusFail
			issues = append(issues, "high memory")
			allGood = false
		}

		if result.ItemsPerSecond < 50 {
			status = statusFail
			issues = append(issues, "low throughput")
			allGood = false
		}

		if result.Duration > 10*time.Second {
			status = statusFail
			issues = append(issues, "slow execution")
			allGood = false
		}

		issueText := ""
		if len(issues) > 0 {
			issueText = fmt.Sprintf(" (%s)", strings.Join(issues, ", "))
		}

		fmt.Printf("%s %s%s\n", status, result.TestName, issueText)
	}

	if allGood {
		fmt.Println("\n🎉 All performance tests are within acceptable limits!")
	} else {
		fmt.Println("\n⚠️  Some performance issues detected - review recommendations above")
	}
}

// getBestPerformingTest finds the test with highest throughput
func getBestPerformingTest(results []BenchmarkResults) string {
	maxThroughput := 0.0
	bestTest := ""

	for _, result := range results {
		if result.ItemsPerSecond > maxThroughput {
			maxThroughput = result.ItemsPerSecond
			bestTest = result.TestName
		}
	}

	return bestTest
}

// RunQuickBenchmark runs a quick performance check
func (runner *BenchmarkRunner) RunQuickBenchmark() {
	fmt.Println("Running quick performance benchmark...")

	// Run just a few key tests
	runner.suite.profileStopTimeProduction()
	runner.suite.profileMemoryUsage()

	fmt.Println("Quick benchmark completed!")
	runner.suite.PrintResults()
}

// RunTargetedBenchmark runs benchmark for a specific component
func (runner *BenchmarkRunner) RunTargetedBenchmark(component string) {
	fmt.Printf("Running targeted benchmark for: %s\n", component)

	switch component {
	case "stoptimes":
		runner.suite.profileStopTimeProduction()
	case "frequencies":
		runner.suite.profileFrequencyProduction()
	case "pathways":
		runner.suite.profilePathwaysGeneration()
	case "memory":
		runner.suite.profileMemoryUsage()
	case "concurrent":
		runner.suite.profileConcurrentLoad()
	default:
		log.Printf("Unknown component: %s", component)
		return
	}

	fmt.Printf("Targeted benchmark for %s completed!\n", component)
	runner.suite.PrintResults()
}

// GetPerformanceMetrics returns key performance metrics
func (runner *BenchmarkRunner) GetPerformanceMetrics() map[string]interface{} {
	results := runner.suite.GetResults()

	metrics := make(map[string]interface{})

	if len(results) == 0 {
		return metrics
	}

	// Calculate aggregate metrics
	var totalDuration time.Duration
	var totalMemory uint64
	var totalItems int
	var totalThroughput float64

	for _, result := range results {
		totalDuration += result.Duration
		totalMemory += result.MemoryAllocated
		totalItems += result.ItemsProcessed
		totalThroughput += result.ItemsPerSecond
	}

	metrics["total_duration_seconds"] = totalDuration.Seconds()
	metrics["total_memory_mb"] = float64(totalMemory) / (1024 * 1024)
	metrics["total_items_processed"] = totalItems
	metrics["average_throughput"] = totalThroughput / float64(len(results))
	metrics["test_count"] = len(results)

	// Individual test metrics
	testMetrics := make(map[string]map[string]interface{})
	for _, result := range results {
		testMetrics[result.TestName] = map[string]interface{}{
			"duration_seconds":    result.Duration.Seconds(),
			"memory_allocated_mb": float64(result.MemoryAllocated) / (1024 * 1024),
			"items_processed":     result.ItemsProcessed,
			"items_per_second":    result.ItemsPerSecond,
			"peak_memory_mb":      float64(result.PeakMemoryUsage) / (1024 * 1024),
			"goroutine_count":     result.GoroutineCount,
		}
	}

	metrics["tests"] = testMetrics
	return metrics
}

// CompareWithBaseline compares current results with baseline performance
func (runner *BenchmarkRunner) CompareWithBaseline(baseline map[string]interface{}) {
	current := runner.GetPerformanceMetrics()

	fmt.Println("\n=== Performance Comparison with Baseline ===")

	if baseline["average_throughput"] != nil && current["average_throughput"] != nil {
		baselineThroughput := baseline["average_throughput"].(float64)
		currentThroughput := current["average_throughput"].(float64)

		improvement := ((currentThroughput - baselineThroughput) / baselineThroughput) * 100

		switch {
		case improvement > 5:
			fmt.Printf("🚀 Throughput improved by %.2f%% (%.2f → %.2f items/sec)\n",
				improvement, baselineThroughput, currentThroughput)
		case improvement < -5:
			fmt.Printf("⚠️  Throughput decreased by %.2f%% (%.2f → %.2f items/sec)\n",
				-improvement, baselineThroughput, currentThroughput)
		default:
			fmt.Printf("✅ Throughput stable: %.2f items/sec (±%.2f%%)\n",
				currentThroughput, improvement)
		}
	}

	if baseline["total_memory_mb"] != nil && current["total_memory_mb"] != nil {
		baselineMemory := baseline["total_memory_mb"].(float64)
		currentMemory := current["total_memory_mb"].(float64)

		memoryChange := ((currentMemory - baselineMemory) / baselineMemory) * 100

		switch {
		case memoryChange > 10:
			fmt.Printf("⚠️  Memory usage increased by %.2f%% (%.2f → %.2f MB)\n",
				memoryChange, baselineMemory, currentMemory)
		case memoryChange < -10:
			fmt.Printf("🎉 Memory usage improved by %.2f%% (%.2f → %.2f MB)\n",
				-memoryChange, baselineMemory, currentMemory)
		default:
			fmt.Printf("✅ Memory usage stable: %.2f MB (±%.2f%%)\n",
				currentMemory, memoryChange)
		}
	}
}
