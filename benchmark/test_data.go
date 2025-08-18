package benchmark

import (
	"fmt"
	"strings"
	"time"

	"github.com/theoremus-urban-solutions/netex-gtfs-converter/model"
	"github.com/theoremus-urban-solutions/netex-gtfs-converter/producer"
)

// createLargeServiceJourney creates a service journey with many stops for testing
func createLargeServiceJourney(stopCount int) *model.ServiceJourney {
	journey := &model.ServiceJourney{
		ID:                fmt.Sprintf("large_journey_%d_stops", stopCount),
		JourneyPatternRef: model.ServiceJourneyPatternRef{Ref: fmt.Sprintf("pattern_%d", stopCount)},
		PassingTimes:      &model.PassingTimes{},
	}

	// Create passing times for each stop
	passingTimes := make([]model.TimetabledPassingTime, stopCount)
	baseTime := time.Date(2024, 1, 1, 8, 0, 0, 0, time.UTC)

	for i := 0; i < stopCount; i++ {
		// Each stop is 2 minutes apart
		arrivalTime := baseTime.Add(time.Duration(i*2) * time.Minute)
		departureTime := arrivalTime.Add(30 * time.Second)

		passingTimes[i] = model.TimetabledPassingTime{
			PointInJourneyPatternRef: fmt.Sprintf("stop_point_%d", i+1),
			ArrivalTime:              arrivalTime.Format("15:04:05"),
			DepartureTime:            departureTime.Format("15:04:05"),
			DayOffset:                0,
		}
	}

	journey.PassingTimes.TimetabledPassingTime = passingTimes
	return journey
}

// createLargeHeadwayJourneyGroup creates a headway journey group with many time bands
func createLargeHeadwayJourneyGroup(timeBandCount int) *model.HeadwayJourneyGroup {
	group := &model.HeadwayJourneyGroup{
		ID:                       fmt.Sprintf("headway_group_%d_bands", timeBandCount),
		Name:                     fmt.Sprintf("Large Headway Group with %d time bands", timeBandCount),
		ScheduledHeadwayInterval: "PT15M", // 15 minutes base interval
		FirstDepartureTime:       "06:00:00",
		LastDepartureTime:        "22:00:00",
		JourneyPatternRef:        fmt.Sprintf("headway_pattern_%d", timeBandCount),
	}

	return group
}

// createLargeStopPlace creates a stop place with many quays for testing
func createLargeStopPlace(quayCount int) *model.StopPlace {
	stopPlace := &model.StopPlace{
		ID:   fmt.Sprintf("large_stop_place_%d_quays", quayCount),
		Name: fmt.Sprintf("Large Test Station with %d quays", quayCount),
		Centroid: &model.Centroid{
			Location: &model.Location{
				Latitude:  59.911491,
				Longitude: 10.757933,
			},
		},
		Quays: &model.Quays{
			Quay: make([]model.Quay, quayCount),
		},
	}

	// Create quays with different characteristics
	modes := []string{"train", "bus", "metro", "tram"}
	levels := []string{"ground", "upper", "lower"}

	for i := 0; i < quayCount; i++ {
		mode := modes[i%len(modes)]
		level := levels[i%len(levels)]

		quay := model.Quay{
			ID:   fmt.Sprintf("quay_%d", i+1),
			Name: fmt.Sprintf("%s %s %s", strings.Title(mode), strings.Title(level), fmt.Sprintf("Platform %d", i+1)),
			Centroid: &model.Centroid{
				Location: &model.Location{
					// Spread quays around the base location
					Latitude:  59.911491 + float64(i%10)*0.0001,
					Longitude: 10.757933 + float64(i/10)*0.0001,
				},
			},
		}

		// Add accessibility info to some quays
		if i%3 == 0 {
			quay.AccessibilityAssessment = &model.AccessibilityAssessment{
				Limitations: &model.Limitations{
					AccessibilityLimitation: &model.AccessibilityLimitation{
						WheelchairAccess: "true",
						StepFreeAccess:   "true",
					},
				},
			}
		}

		stopPlace.Quays.Quay[i] = quay
	}

	return stopPlace
}

// createComplexInterchangeStation creates a complex multi-modal interchange
func createComplexInterchangeStation(quayCount int) *model.StopPlace {
	stopPlace := createLargeStopPlace(quayCount)
	stopPlace.Name = "Complex Interchange Hub"
	stopPlace.ID = "complex_interchange"

	// Ensure we have multiple transport modes
	modes := []string{"train", "bus", "metro", "tram", "ferry"}
	levels := []string{"ground", "upper", "lower", "mezzanine", "basement"}

	for i := range stopPlace.Quays.Quay {
		quay := &stopPlace.Quays.Quay[i]
		mode := modes[i%len(modes)]
		level := levels[i%len(levels)]

		quay.Name = fmt.Sprintf("%s Platform %d %s Level",
			strings.Title(mode), (i%5)+1, strings.Title(level))

		// Add more complex accessibility patterns
		if strings.Contains(mode, "train") || strings.Contains(mode, "metro") {
			quay.AccessibilityAssessment = &model.AccessibilityAssessment{
				Limitations: &model.Limitations{
					AccessibilityLimitation: &model.AccessibilityLimitation{
						WheelchairAccess:        "true",
						StepFreeAccess:          "true",
						EscalatorFreeAccess:     "false",
						LiftFreeAccess:          "false",
						AudibleSignalsAvailable: "true",
						VisualSignalsAvailable:  "true",
					},
				},
			}
		}
	}

	return stopPlace
}

// createJourneyPattern creates a journey pattern with specified number of stops
func createJourneyPattern(stopCount int) *model.JourneyPattern {
	pattern := &model.JourneyPattern{
		ID:   fmt.Sprintf("journey_pattern_%d_stops", stopCount),
		Name: fmt.Sprintf("Journey Pattern with %d stops", stopCount),
		PointsInSequence: &model.PointsInSequence{
			PointInJourneyPatternOrStopPointInJourneyPatternOrTimingPointInJourneyPattern: make([]interface{}, stopCount),
		},
	}

	for i := 0; i < stopCount; i++ {
		stopPoint := &model.StopPointInJourneyPattern{
			ID:                    fmt.Sprintf("stop_point_%d", i+1),
			Order:                 i + 1,
			ScheduledStopPointRef: fmt.Sprintf("scheduled_stop_%d", i+1),
		}

		pattern.PointsInSequence.PointInJourneyPatternOrStopPointInJourneyPatternOrTimingPointInJourneyPattern[i] = stopPoint
	}

	return pattern
}

// createLargeNetexDataset creates a large dataset of various NeTEx entities
func createLargeNetexDataset(entityCount int) []interface{} {
	entities := make([]interface{}, 0, entityCount*7) // Multiple entity types

	// Create various entity types
	for i := 0; i < entityCount; i++ {
		// Authority
		authority := &model.Authority{
			ID:        fmt.Sprintf("authority_%d", i),
			Name:      fmt.Sprintf("Test Authority %d", i),
			ShortName: fmt.Sprintf("TA%d", i),
			URL:       fmt.Sprintf("https://authority%d.example.com", i),
		}
		entities = append(entities, authority)

		// Line
		line := &model.Line{
			ID:           fmt.Sprintf("line_%d", i),
			Name:         fmt.Sprintf("Test Line %d", i),
			ShortName:    fmt.Sprintf("L%d", i),
			AuthorityRef: authority.ID,
		}
		entities = append(entities, line)

		// Route
		route := &model.Route{
			ID:      fmt.Sprintf("route_%d", i),
			Name:    fmt.Sprintf("Test Route %d", i),
			LineRef: model.RouteLineRef{Ref: line.ID},
		}
		entities = append(entities, route)

		// Stop Place
		stopPlace := &model.StopPlace{
			ID:   fmt.Sprintf("stop_place_%d", i),
			Name: fmt.Sprintf("Test Stop Place %d", i),
			Centroid: &model.Centroid{
				Location: &model.Location{
					Latitude:  59.9 + float64(i%100)*0.001,
					Longitude: 10.7 + float64(i/100)*0.001,
				},
			},
		}
		entities = append(entities, stopPlace)

		// Quay
		quay := &model.Quay{
			ID:   fmt.Sprintf("quay_%d", i),
			Name: fmt.Sprintf("Test Quay %d", i),
			Centroid: &model.Centroid{
				Location: &model.Location{
					Latitude:  59.9 + float64(i%100)*0.001,
					Longitude: 10.7 + float64(i/100)*0.001,
				},
			},
		}
		entities = append(entities, quay)

		// Service Journey
		serviceJourney := &model.ServiceJourney{
			ID:                fmt.Sprintf("service_journey_%d", i),
			JourneyPatternRef: model.ServiceJourneyPatternRef{Ref: fmt.Sprintf("journey_pattern_%d", i)},
		}
		entities = append(entities, serviceJourney)

		// Journey Pattern
		journeyPattern := createJourneyPattern(10) // 10 stops each
		journeyPattern.ID = fmt.Sprintf("journey_pattern_%d", i)
		entities = append(entities, journeyPattern)
	}

	return entities
}

// createMultipleServiceJourneys creates multiple service journeys for concurrent testing
func createMultipleServiceJourneys(count int) []*model.ServiceJourney {
	journeys := make([]*model.ServiceJourney, count)

	for i := 0; i < count; i++ {
		journeys[i] = createLargeServiceJourney(20) // 20 stops each
		journeys[i].ID = fmt.Sprintf("concurrent_journey_%d", i)
	}

	return journeys
}

// populateGtfsRepository populates a GTFS repository with test data
func populateGtfsRepository(gtfsRepo producer.GtfsRepository, entityCount int) {
	// Create agencies
	for i := 0; i < entityCount; i++ {
		agency := &model.Agency{
			AgencyID:       fmt.Sprintf("agency_%d", i),
			AgencyName:     fmt.Sprintf("Test Agency %d", i),
			AgencyURL:      fmt.Sprintf("https://agency%d.example.com", i),
			AgencyTimezone: "Europe/Oslo",
		}
		gtfsRepo.SaveEntity(agency)
	}

	// Create routes
	for i := 0; i < entityCount; i++ {
		route := &model.GtfsRoute{
			RouteID:        fmt.Sprintf("route_%d", i),
			AgencyID:       fmt.Sprintf("agency_%d", i%100), // Distribute across agencies
			RouteShortName: fmt.Sprintf("R%d", i),
			RouteLongName:  fmt.Sprintf("Test Route %d", i),
			RouteType:      3, // Bus
		}
		gtfsRepo.SaveEntity(route)
	}

	// Create stops
	for i := 0; i < entityCount; i++ {
		stop := &model.Stop{
			StopID:   fmt.Sprintf("stop_%d", i),
			StopName: fmt.Sprintf("Test Stop %d", i),
			StopLat:  59.9 + float64(i%100)*0.001,
			StopLon:  10.7 + float64(i/100)*0.001,
		}
		gtfsRepo.SaveEntity(stop)
	}

	// Create trips
	for i := 0; i < entityCount; i++ {
		trip := &model.Trip{
			RouteID:   fmt.Sprintf("route_%d", i%100),  // Distribute across routes
			ServiceID: fmt.Sprintf("service_%d", i%50), // Distribute across services
			TripID:    fmt.Sprintf("trip_%d", i),
		}
		gtfsRepo.SaveEntity(trip)
	}

	// Create stop times
	for i := 0; i < entityCount; i++ {
		stopTime := &model.StopTime{
			TripID:        fmt.Sprintf("trip_%d", i%100),
			ArrivalTime:   fmt.Sprintf("08:%02d:00", (i % 60)),
			DepartureTime: fmt.Sprintf("08:%02d:30", (i % 60)),
			StopID:        fmt.Sprintf("stop_%d", i%1000),
			StopSequence:  (i % 50) + 1,
		}
		gtfsRepo.SaveEntity(stopTime)
	}

	// Create frequencies
	for i := 0; i < entityCount/10; i++ { // Fewer frequencies
		frequency := &model.Frequency{
			TripID:      fmt.Sprintf("trip_%d", i),
			StartTime:   "06:00:00",
			EndTime:     "22:00:00",
			HeadwaySecs: 300 + (i%12)*60, // 5-17 minutes
			ExactTimes:  "0",
		}
		gtfsRepo.SaveEntity(frequency)
	}

	// Create pathways
	for i := 0; i < entityCount/5; i++ { // Fewer pathways
		pathway := &model.Pathway{
			PathwayID:       fmt.Sprintf("pathway_%d", i),
			FromStopID:      fmt.Sprintf("stop_%d", i),
			ToStopID:        fmt.Sprintf("stop_%d", i+1),
			PathwayMode:     1, // Walkway
			IsBidirectional: 1,
			Length:          float64(50 + (i%20)*10), // 50-250 meters
			TraversalTime:   60 + (i%5)*30,           // 1-3 minutes
		}
		gtfsRepo.SaveEntity(pathway)
	}

	// Create levels
	levelCount := entityCount / 20 // Much fewer levels
	for i := 0; i < levelCount; i++ {
		level := &model.Level{
			LevelID:    fmt.Sprintf("level_%d", i),
			LevelIndex: float64(i%5 - 2), // -2 to +2 levels
			LevelName:  fmt.Sprintf("Level %d", i),
		}
		gtfsRepo.SaveEntity(level)
	}
}
