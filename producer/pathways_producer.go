package producer

import (
	"fmt"
	"math"
	"strings"

	"github.com/theoremus-urban-solutions/netex-gtfs-converter/model"
)

// Interface is defined in producer.go to avoid duplication

// DefaultPathwaysProducer implements PathwaysProducer
type DefaultPathwaysProducer struct {
	netexRepository NetexRepository
	gtfsRepository  GtfsRepository

	// Configuration
	defaultTraversalTime int     // seconds
	defaultStairHeight   float64 // meters per stair
	maxSlopePercent      float64 // maximum slope percentage
}

// NewDefaultPathwaysProducer creates a new pathways producer
func NewDefaultPathwaysProducer(netexRepo NetexRepository, gtfsRepo GtfsRepository) *DefaultPathwaysProducer {
	return &DefaultPathwaysProducer{
		netexRepository:      netexRepo,
		gtfsRepository:       gtfsRepo,
		defaultTraversalTime: 60,   // 1 minute default
		defaultStairHeight:   0.15, // 15cm per stair
		maxSlopePercent:      8.0,  // 8% max slope for accessibility
	}
}

// ProducePathwaysFromStopPlace generates pathways within a stop place
func (p *DefaultPathwaysProducer) ProducePathwaysFromStopPlace(stopPlace *model.StopPlace) ([]*model.Pathway, error) {
	if stopPlace == nil {
		return nil, nil
	}

	var pathways []*model.Pathway

	// Get all quays in the stop place
	if stopPlace.Quays == nil || len(stopPlace.Quays.Quay) == 0 {
		return nil, nil
	}

	quays := stopPlace.Quays.Quay

	// Generate pathways between quays in the same stop place
	for i := 0; i < len(quays); i++ {
		for j := i + 1; j < len(quays); j++ {
			pathway, err := p.createPathwayBetweenQuays(&quays[i], &quays[j], stopPlace)
			if err != nil {
				continue // Log error but continue processing
			}
			if pathway != nil {
				pathways = append(pathways, pathway)
			}
		}
	}

	// Add entrance/exit pathways
	entrancePathways := p.createEntrancePathways(stopPlace, quays)
	pathways = append(pathways, entrancePathways...)

	// Add accessibility-specific pathways (elevators, ramps)
	accessibilityPathways := p.createAccessibilityPathways(stopPlace, quays)
	pathways = append(pathways, accessibilityPathways...)

	return pathways, nil
}

// ProduceLevelsFromStopPlace generates levels for a multi-level stop place
func (p *DefaultPathwaysProducer) ProduceLevelsFromStopPlace(stopPlace *model.StopPlace) ([]*model.Level, error) {
	if stopPlace == nil {
		return nil, nil
	}

	var levels []*model.Level
	levelMap := make(map[string]*model.Level)

	// Extract level information from quays
	if stopPlace.Quays != nil {
		for _, quay := range stopPlace.Quays.Quay {
			// Determine level from quay properties
			levelInfo := p.extractLevelInfo(&quay)
			if levelInfo == nil {
				continue
			}

			// Create or update level
			levelID := fmt.Sprintf("%s_level_%s", stopPlace.ID, levelInfo.ID)
			if existingLevel, exists := levelMap[levelID]; exists {
				// Update existing level if needed
				if levelInfo.Name != "" && existingLevel.LevelName == "" {
					existingLevel.LevelName = levelInfo.Name
				}
			} else {
				level := &model.Level{
					LevelID:    levelID,
					LevelIndex: levelInfo.Index,
					LevelName:  levelInfo.Name,
				}
				levelMap[levelID] = level
				levels = append(levels, level)
			}
		}
	}

	// Add default ground level if no levels found
	if len(levels) == 0 {
		groundLevel := &model.Level{
			LevelID:    fmt.Sprintf("%s_level_0", stopPlace.ID),
			LevelIndex: 0.0,
			LevelName:  "Ground Level",
		}
		levels = append(levels, groundLevel)
	}

	return levels, nil
}

// ProduceAccessibilityPathways creates pathways based on accessibility requirements
func (p *DefaultPathwaysProducer) ProduceAccessibilityPathways(from, to *model.Quay) (*model.Pathway, error) {
	if from == nil || to == nil {
		return nil, fmt.Errorf("both quays are required")
	}

	pathway := &model.Pathway{
		PathwayID:       fmt.Sprintf("pathway_%s_%s", from.ID, to.ID),
		FromStopID:      from.ID,
		ToStopID:        to.ID,
		IsBidirectional: 1, // Default bidirectional
	}

	// Determine pathway mode based on accessibility assessments
	mode := p.determinePathwayMode(from, to)
	pathway.PathwayMode = mode

	// Calculate distance if coordinates available
	if from.Centroid != nil && to.Centroid != nil {
		distance := p.calculateDistance(from.Centroid, to.Centroid)
		pathway.Length = distance

		// Estimate traversal time based on mode and distance
		traversalTime := p.estimateTraversalTime(distance, mode)
		pathway.TraversalTime = traversalTime
	}

	// Apply accessibility constraints
	if from.AccessibilityAssessment != nil || to.AccessibilityAssessment != nil {
		p.applyAccessibilityConstraints(pathway, from.AccessibilityAssessment, to.AccessibilityAssessment)
	}

	return pathway, nil
}

// Helper methods

func (p *DefaultPathwaysProducer) createPathwayBetweenQuays(quay1, quay2 *model.Quay, stopPlace *model.StopPlace) (*model.Pathway, error) {
	pathway := &model.Pathway{
		PathwayID:       fmt.Sprintf("pathway_%s_%s", quay1.ID, quay2.ID),
		FromStopID:      quay1.ID,
		ToStopID:        quay2.ID,
		PathwayMode:     1, // Walkway by default
		IsBidirectional: 1, // Bidirectional by default
	}

	// Calculate distance between quays
	if quay1.Centroid != nil && quay2.Centroid != nil {
		distance := p.calculateDistance(quay1.Centroid, quay2.Centroid)
		pathway.Length = distance

		// Estimate traversal time (assuming average walking speed of 1.2 m/s)
		pathway.TraversalTime = int(distance / 1.2)
	} else {
		// Use default traversal time
		pathway.TraversalTime = p.defaultTraversalTime
	}

	// Check for level differences to determine if stairs/elevators needed
	level1 := p.extractLevelInfo(quay1)
	level2 := p.extractLevelInfo(quay2)

	if level1 != nil && level2 != nil && level1.Index != level2.Index {
		// Different levels - might be stairs or elevator
		levelDiff := math.Abs(level1.Index - level2.Index)
		if levelDiff > 0 {
			// Assume stairs if level difference exists
			pathway.PathwayMode = 2 // Stairs
			pathway.StairCount = int(levelDiff / p.defaultStairHeight)

			// Adjust traversal time for stairs (slower than walking)
			pathway.TraversalTime = int(float64(pathway.TraversalTime) * 1.5)
		}
	}

	return pathway, nil
}

func (p *DefaultPathwaysProducer) createEntrancePathways(stopPlace *model.StopPlace, quays []model.Quay) []*model.Pathway {
	var pathways []*model.Pathway

	// Create main entrance node
	entranceID := fmt.Sprintf("%s_entrance", stopPlace.ID)

	// Create pathways from entrance to each quay
	for _, quay := range quays {
		pathway := &model.Pathway{
			PathwayID:       fmt.Sprintf("pathway_entrance_%s", quay.ID),
			FromStopID:      entranceID,
			ToStopID:        quay.ID,
			PathwayMode:     1, // Walkway
			IsBidirectional: 1, // Bidirectional
		}

		// Estimate distance and time from entrance
		if quay.Centroid != nil && stopPlace.Centroid != nil {
			distance := p.calculateDistance(stopPlace.Centroid, quay.Centroid)
			pathway.Length = distance
			pathway.TraversalTime = int(distance / 1.2) // Walking speed 1.2 m/s
		} else {
			pathway.TraversalTime = p.defaultTraversalTime
		}

		pathways = append(pathways, pathway)
	}

	return pathways
}

func (p *DefaultPathwaysProducer) createAccessibilityPathways(stopPlace *model.StopPlace, quays []model.Quay) []*model.Pathway {
	var pathways []*model.Pathway

	// Check for accessibility requirements
	hasAccessibilityNeeds := false
	for _, quay := range quays {
		if quay.AccessibilityAssessment != nil {
			hasAccessibilityNeeds = true
			break
		}
	}

	if !hasAccessibilityNeeds {
		return pathways
	}

	// Create elevator pathways for multi-level stations
	levels := make(map[float64][]model.Quay)
	for _, quay := range quays {
		levelInfo := p.extractLevelInfo(&quay)
		if levelInfo != nil {
			levels[levelInfo.Index] = append(levels[levelInfo.Index], quay)
		}
	}

	if len(levels) > 1 {
		// Create elevator shaft connecting all levels
		elevatorID := fmt.Sprintf("%s_elevator", stopPlace.ID)

		for level, quaysAtLevel := range levels {
			for _, quay := range quaysAtLevel {
				// Check if quay needs accessible access
				if p.needsAccessibleAccess(&quay) {
					pathway := &model.Pathway{
						PathwayID:       fmt.Sprintf("pathway_elevator_%s", quay.ID),
						FromStopID:      elevatorID,
						ToStopID:        quay.ID,
						PathwayMode:     5, // Elevator
						IsBidirectional: 1,
						TraversalTime:   30 + int(math.Abs(level)*10), // Base time + level time
					}
					pathways = append(pathways, pathway)
				}
			}
		}
	}

	// Create ramp pathways where needed
	for _, quay := range quays {
		if p.needsRampAccess(&quay) {
			rampPathway := &model.Pathway{
				PathwayID:       fmt.Sprintf("pathway_ramp_%s", quay.ID),
				FromStopID:      fmt.Sprintf("%s_entrance", stopPlace.ID),
				ToStopID:        quay.ID,
				PathwayMode:     1, // Walkway (accessible)
				IsBidirectional: 1,
				MaxSlope:        p.maxSlopePercent / 100.0, // Convert percentage to decimal
			}

			// Ramps are longer but more accessible
			if quay.Centroid != nil && stopPlace.Centroid != nil {
				distance := p.calculateDistance(stopPlace.Centroid, quay.Centroid)
				// Ramps are typically longer due to slope requirements
				rampPathway.Length = distance * 1.5
				rampPathway.TraversalTime = int(rampPathway.Length / 0.8) // Slower speed on ramp
			}

			pathways = append(pathways, rampPathway)
		}
	}

	return pathways
}

func (p *DefaultPathwaysProducer) determinePathwayMode(from, to *model.Quay) int {
	// Pathway modes in GTFS:
	// 1 = walkway
	// 2 = stairs
	// 3 = moving sidewalk/travelator
	// 4 = escalator
	// 5 = elevator
	// 6 = fare gate
	// 7 = exit gate

	// Check for level differences
	level1 := p.extractLevelInfo(from)
	level2 := p.extractLevelInfo(to)

	if level1 != nil && level2 != nil && level1.Index != level2.Index {
		// Different levels - check accessibility requirements
		if p.needsAccessibleAccess(from) || p.needsAccessibleAccess(to) {
			return 5 // Elevator for accessibility
		}
		return 2 // Stairs by default for level changes
	}

	return 1 // Walkway for same level
}

func (p *DefaultPathwaysProducer) calculateDistance(from, to *model.Centroid) float64 {
	if from == nil || to == nil || from.Location == nil || to.Location == nil {
		return 0
	}

	// Haversine formula for distance between two points
	lat1 := from.Location.Latitude * math.Pi / 180
	lat2 := to.Location.Latitude * math.Pi / 180
	deltaLat := (to.Location.Latitude - from.Location.Latitude) * math.Pi / 180
	deltaLon := (to.Location.Longitude - from.Location.Longitude) * math.Pi / 180

	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1)*math.Cos(lat2)*
			math.Sin(deltaLon/2)*math.Sin(deltaLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	// Earth radius in meters
	earthRadius := 6371000.0
	distance := earthRadius * c

	return distance
}

func (p *DefaultPathwaysProducer) estimateTraversalTime(distance float64, pathwayMode int) int {
	// Estimate traversal time based on pathway mode and typical speeds
	var speed float64 // meters per second

	switch pathwayMode {
	case 1: // Walkway
		speed = 1.2
	case 2: // Stairs
		speed = 0.5 // Slower on stairs
	case 3: // Moving sidewalk
		speed = 2.0 // Faster with moving sidewalk
	case 4: // Escalator
		speed = 0.75 // Fixed escalator speed
	case 5: // Elevator
		// Elevator time includes waiting and travel
		return 45 // Average elevator time
	case 6, 7: // Fare/exit gate
		return 5 // Fixed gate traversal time
	default:
		speed = 1.0
	}

	if speed > 0 {
		return int(distance / speed)
	}

	return p.defaultTraversalTime
}

func (p *DefaultPathwaysProducer) applyAccessibilityConstraints(pathway *model.Pathway, fromAccess, toAccess *model.AccessibilityAssessment) {
	// Check accessibility limitations
	hasWheelchairLimitation := false

	if fromAccess != nil && fromAccess.Limitations != nil {
		if fromAccess.Limitations.AccessibilityLimitation != nil {
			// Check for wheelchair limitations
			if strings.Contains(strings.ToLower(fromAccess.Limitations.AccessibilityLimitation.WheelchairAccess), "false") {
				hasWheelchairLimitation = true
			}
		}
	}

	if toAccess != nil && toAccess.Limitations != nil {
		if toAccess.Limitations.AccessibilityLimitation != nil {
			if strings.Contains(strings.ToLower(toAccess.Limitations.AccessibilityLimitation.WheelchairAccess), "false") {
				hasWheelchairLimitation = true
			}
		}
	}

	// Adjust pathway properties based on limitations
	if hasWheelchairLimitation {
		// If not wheelchair accessible, ensure appropriate mode
		if pathway.PathwayMode == 2 { // Stairs
			// Keep as stairs but note it's not accessible
			// In practice, we might add a custom field or use MaxSlope to indicate
			pathway.MaxSlope = -1 // Indicator of non-accessible
		}
	}
}

// LevelInfo represents extracted level information
type LevelInfo struct {
	ID    string
	Index float64
	Name  string
}

func (p *DefaultPathwaysProducer) extractLevelInfo(quay *model.Quay) *LevelInfo {
	// Extract level information from quay properties
	// This is simplified - real implementation would parse NeTEx level references

	// Check quay name for level indicators
	name := strings.ToLower(quay.Name)

	// Check for explicit level indicators first
	if strings.Contains(name, "upper") || strings.Contains(name, "mezzanine") {
		return &LevelInfo{ID: "upper", Index: 1, Name: "Upper Level"}
	}

	if strings.Contains(name, "lower") || strings.Contains(name, "basement") {
		return &LevelInfo{ID: "lower", Index: -1, Name: "Lower Level"}
	}

	// Check for "ground" explicitly
	if strings.Contains(name, "ground") {
		return &LevelInfo{ID: "0", Index: 0, Name: "Ground Level"}
	}

	if strings.Contains(name, "platform") {
		// For platforms, check for level indicators within the name
		if strings.Contains(name, "upper") {
			return &LevelInfo{ID: "upper", Index: 1, Name: "Upper Platform"}
		}
		if strings.Contains(name, "lower") {
			return &LevelInfo{ID: "lower", Index: -1, Name: "Lower Platform"}
		}

		// Extract platform number as same level by default
		if strings.Contains(name, "1") {
			return &LevelInfo{ID: "1", Index: 0, Name: "Platform 1"}
		}
		if strings.Contains(name, "2") {
			return &LevelInfo{ID: "2", Index: 0, Name: "Platform 2"}
		}
	}

	// Default ground level
	return &LevelInfo{ID: "0", Index: 0, Name: "Ground Level"}
}

func (p *DefaultPathwaysProducer) needsAccessibleAccess(quay *model.Quay) bool {
	// Check if quay requires accessible access
	if quay.AccessibilityAssessment != nil {
		if quay.AccessibilityAssessment.Limitations != nil {
			// If there are limitations, accessible access is needed
			return true
		}
	}

	// Check for keywords in quay description
	name := strings.ToLower(quay.Name)
	return strings.Contains(name, "accessible") || strings.Contains(name, "wheelchair")
}

func (p *DefaultPathwaysProducer) needsRampAccess(quay *model.Quay) bool {
	// Check if quay needs ramp access
	if quay.AccessibilityAssessment != nil {
		// If accessibility assessment exists, might need ramp
		return true
	}

	// Check for platform height differences that might require ramp
	name := strings.ToLower(quay.Name)
	return strings.Contains(name, "step-free") || strings.Contains(name, "level")
}
