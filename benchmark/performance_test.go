package benchmark

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/theoremus-urban-solutions/netex-gtfs-converter/model"
	"github.com/theoremus-urban-solutions/netex-gtfs-converter/producer"
	"github.com/theoremus-urban-solutions/netex-gtfs-converter/repository"
)

// BenchmarkStopTimeProduction tests stop time production performance
func BenchmarkStopTimeProduction(b *testing.B) {
	// Setup
	netexRepo := repository.NewDefaultNetexRepository()
	gtfsRepo := repository.NewDefaultGtfsRepository()
	producer := producer.NewEnhancedStopTimeProducer(netexRepo, gtfsRepo)

	// Create test data
	serviceJourney := createLargeServiceJourney(100) // 100 stops
	trip := &model.Trip{TripID: "benchmark_trip"}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := producer.ProduceStopTimesForTrip(serviceJourney, trip, nil, "Benchmark Route")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkFrequencyProduction tests frequency-based service production
func BenchmarkFrequencyProduction(b *testing.B) {
	netexRepo := repository.NewDefaultNetexRepository()
	gtfsRepo := repository.NewDefaultGtfsRepository()
	producer := producer.NewDefaultFrequencyProducer(netexRepo, gtfsRepo)

	// Create test headway journey group
	group := createLargeHeadwayJourneyGroup(50) // 50 time bands

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := producer.ProduceFromHeadwayJourneyGroup(group)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkPathwaysGeneration tests pathways generation performance
func BenchmarkPathwaysGeneration(b *testing.B) {
	netexRepo := repository.NewDefaultNetexRepository()
	gtfsRepo := repository.NewDefaultGtfsRepository()
	producer := producer.NewDefaultPathwaysProducer(netexRepo, gtfsRepo)

	// Create large multi-modal station
	stopPlace := createLargeStopPlace(25) // 25 quays

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := producer.ProducePathwaysFromStopPlace(stopPlace)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkInterchangeGeneration tests interchange generation performance
func BenchmarkInterchangeGeneration(b *testing.B) {
	netexRepo := repository.NewDefaultNetexRepository()
	gtfsRepo := repository.NewDefaultGtfsRepository()
	pathwaysProducer := producer.NewDefaultPathwaysProducer(netexRepo, gtfsRepo)
	producer := producer.NewSophisticatedInterchangeProducer(netexRepo, gtfsRepo, pathwaysProducer)

	// Create complex interchange station
	stopPlace := createComplexInterchangeStation(30) // 30 quays, multiple modes

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := producer.ProduceComplexInterchanges(stopPlace)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkGTFSGeneration tests complete GTFS generation
func BenchmarkGTFSGeneration(b *testing.B) {
	gtfsRepo := repository.NewDefaultGtfsRepository()

	// Populate with test data
	populateGtfsRepository(gtfsRepo, 1000) // 1000 entities of each type

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		reader, err := gtfsRepo.WriteGtfs()
		if err != nil {
			b.Fatal(err)
		}

		// Consume the reader to measure actual generation time
		buf := make([]byte, 1024)
		for {
			_, err := reader.Read(buf)
			if err != nil {
				break
			}
		}
	}
}

// BenchmarkMemoryUsage tests memory usage patterns
func BenchmarkMemoryUsage(b *testing.B) {
	var memStats runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&memStats)

	initialAlloc := memStats.Alloc

	netexRepo := repository.NewDefaultNetexRepository()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Create and process large dataset
		entities := createLargeNetexDataset(1000)

		for _, entity := range entities {
			netexRepo.SaveEntity(entity)
		}

		// Force garbage collection periodically
		if i%100 == 0 {
			runtime.GC()
		}
	}

	runtime.GC()
	runtime.ReadMemStats(&memStats)

	b.ReportMetric(float64(memStats.Alloc-initialAlloc)/float64(b.N), "bytes/op")
	b.ReportMetric(float64(memStats.Mallocs)/float64(b.N), "allocs/op")
}

// BenchmarkConcurrentProcessing tests concurrent processing performance
func BenchmarkConcurrentProcessing(b *testing.B) {
	netexRepo := repository.NewDefaultNetexRepository()
	gtfsRepo := repository.NewDefaultGtfsRepository()

	// Pre-populate with test data
	serviceJourneys := createMultipleServiceJourneys(100)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Process multiple service journeys concurrently
		results := make(chan bool, len(serviceJourneys))

		for _, journey := range serviceJourneys {
			go func(sj *model.ServiceJourney) {
				producer := producer.NewEnhancedStopTimeProducer(netexRepo, gtfsRepo)
				trip := &model.Trip{TripID: sj.ID + "_trip"}
				_, err := producer.ProduceStopTimesForTrip(sj, trip, nil, "Concurrent Test")
				results <- (err == nil)
			}(journey)
		}

		// Wait for all to complete
		for range serviceJourneys {
			<-results
		}
	}
}

// BenchmarkLargeDatasetProcessing tests processing of large datasets
func BenchmarkLargeDatasetProcessing(b *testing.B) {
	sizes := []int{100, 1000, 5000, 10000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("Dataset_%d", size), func(b *testing.B) {
			netexRepo := repository.NewDefaultNetexRepository()

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				entities := createLargeNetexDataset(size)

				for _, entity := range entities {
					netexRepo.SaveEntity(entity)
				}
			}

			b.ReportMetric(float64(size), "entities")
		})
	}
}

// BenchmarkShapeGeneration tests shape generation performance
func BenchmarkShapeGeneration(b *testing.B) {
	netexRepo := repository.NewDefaultNetexRepository()
	gtfsRepo := repository.NewDefaultGtfsRepository()
	producer := producer.NewDefaultShapeProducer(netexRepo, gtfsRepo)

	// Create journey patterns with varying complexity
	patterns := []*model.JourneyPattern{
		createJourneyPattern(10),  // 10 stops
		createJourneyPattern(25),  // 25 stops
		createJourneyPattern(50),  // 50 stops
		createJourneyPattern(100), // 100 stops
	}

	for i, pattern := range patterns {
		b.Run(fmt.Sprintf("Stops_%d", []int{10, 25, 50, 100}[i]), func(b *testing.B) {
			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				_, err := producer.Produce(pattern)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
