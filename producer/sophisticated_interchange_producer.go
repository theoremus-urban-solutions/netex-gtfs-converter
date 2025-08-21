package producer

import (
	"fmt"
	"math"
	"strings"

	"github.com/theoremus-urban-solutions/netex-gtfs-converter/model"
)

// SophisticatedInterchangeProducer handles complex European transit interchanges
type SophisticatedInterchangeProducer struct {
	netexRepository  NetexRepository
	gtfsRepository   GtfsRepository
	pathwaysProducer PathwaysProducer

	// Configuration for European transit patterns
	defaultWalkingTime     int     // seconds for walking between platforms
	maximumWalkingDistance float64 // meters
	minimumConnectionTime  int     // seconds minimum for any connection
	maximumConnectionTime  int     // seconds maximum before connection expires
	wheelchairExtraTime    int     // additional seconds for wheelchair users
	crossPlatformTime      int     // seconds for cross-platform transfers
	levelChangeTime        int     // additional seconds per level change
}

// NewSophisticatedInterchangeProducer creates a new sophisticated interchange producer
func NewSophisticatedInterchangeProducer(netexRepo NetexRepository, gtfsRepo GtfsRepository, pathwaysProducer PathwaysProducer) *SophisticatedInterchangeProducer {
	return &SophisticatedInterchangeProducer{
		netexRepository:        netexRepo,
		gtfsRepository:         gtfsRepo,
		pathwaysProducer:       pathwaysProducer,
		defaultWalkingTime:     180,  // 3 minutes default walk
		maximumWalkingDistance: 500,  // 500 meters max walking
		minimumConnectionTime:  60,   // 1 minute minimum
		maximumConnectionTime:  1800, // 30 minutes maximum
		wheelchairExtraTime:    60,   // 1 minute extra for accessibility
		crossPlatformTime:      30,   // 30 seconds cross-platform
		levelChangeTime:        45,   // 45 seconds per level change
	}
}

// ProduceComplexInterchanges generates transfers for complex interchange scenarios
func (p *SophisticatedInterchangeProducer) ProduceComplexInterchanges(stopPlace *model.StopPlace) ([]*model.Transfer, error) {
	if stopPlace == nil || stopPlace.Quays == nil {
		return nil, nil
	}

	var transfers []*model.Transfer

	// Generate intra-station transfers
	intraStationTransfers, err := p.generateIntraStationTransfers(stopPlace)
	if err != nil {
		return nil, fmt.Errorf("failed to generate intra-station transfers: %w", err)
	}
	transfers = append(transfers, intraStationTransfers...)

	// Generate mode-specific transfers
	modeTransfers, err := p.generateModeSpecificTransfers(stopPlace)
	if err != nil {
		return nil, fmt.Errorf("failed to generate mode-specific transfers: %w", err)
	}
	transfers = append(transfers, modeTransfers...)

	// Generate timed transfers based on service patterns
	timedTransfers, err := p.generateTimedTransfers(stopPlace)
	if err != nil {
		return nil, fmt.Errorf("failed to generate timed transfers: %w", err)
	}
	transfers = append(transfers, timedTransfers...)

	// Generate accessibility-specific transfers
	accessibilityTransfers, err := p.generateAccessibilityTransfers(stopPlace)
	if err != nil {
		return nil, fmt.Errorf("failed to generate accessibility transfers: %w", err)
	}
	transfers = append(transfers, accessibilityTransfers...)

	// Optimize transfer times based on pathways
	optimizedTransfers := p.optimizeTransferTimes(transfers, stopPlace)

	return optimizedTransfers, nil
}

// generateIntraStationTransfers creates transfers within the same station
func (p *SophisticatedInterchangeProducer) generateIntraStationTransfers(stopPlace *model.StopPlace) ([]*model.Transfer, error) {
	var transfers []*model.Transfer
	quays := stopPlace.Quays.Quay

	// Create transfers between all quay combinations
	for i := 0; i < len(quays); i++ {
		for j := i + 1; j < len(quays); j++ {
			transfer, err := p.createQuayToQuayTransfer(&quays[i], &quays[j], stopPlace)
			if err != nil {
				continue // Log error but continue processing
			}

			if transfer != nil {
				transfers = append(transfers, transfer)

				// Create reverse transfer with potentially different time
				reverseTransfer, err := p.createQuayToQuayTransfer(&quays[j], &quays[i], stopPlace)
				if err == nil && reverseTransfer != nil {
					transfers = append(transfers, reverseTransfer)
				}
			}
		}
	}

	return transfers, nil
}

// generateModeSpecificTransfers creates transfers considering transport modes
func (p *SophisticatedInterchangeProducer) generateModeSpecificTransfers(stopPlace *model.StopPlace) ([]*model.Transfer, error) {
	var transfers []*model.Transfer

	// Group quays by transport mode
	modeGroups := p.groupQuaysByMode(stopPlace.Quays.Quay)

	// Create mode-specific transfer patterns
	for fromMode, fromQuays := range modeGroups {
		for toMode, toQuays := range modeGroups {
			if fromMode == toMode {
				continue // Skip same-mode transfers (handled elsewhere)
			}

			// Create transfers between different modes
			for _, fromQuay := range fromQuays {
				for _, toQuay := range toQuays {
					transfer := p.createModeTransfer(fromQuay, toQuay, fromMode, toMode, stopPlace)
					if transfer != nil {
						transfers = append(transfers, transfer)
					}
				}
			}
		}
	}

	return transfers, nil
}

// generateTimedTransfers creates transfers based on service timing coordination
func (p *SophisticatedInterchangeProducer) generateTimedTransfers(stopPlace *model.StopPlace) ([]*model.Transfer, error) {
	var transfers []*model.Transfer

	// Get service journeys that serve this stop place
	serviceJourneys := p.getServiceJourneysForStopPlace(stopPlace)

	// Analyze timing patterns for coordinated transfers
	for i, journey1 := range serviceJourneys {
		for j, journey2 := range serviceJourneys {
			if i >= j {
				continue
			}

			// Check if journeys have coordinated timing
			if p.hasCoordinatedTiming(journey1, journey2, stopPlace) {
				transfer := p.createTimedTransfer(journey1, journey2, stopPlace)
				if transfer != nil {
					transfers = append(transfers, transfer)
				}
			}
		}
	}

	return transfers, nil
}

// generateAccessibilityTransfers creates accessibility-specific transfer options
func (p *SophisticatedInterchangeProducer) generateAccessibilityTransfers(stopPlace *model.StopPlace) ([]*model.Transfer, error) {
	var transfers []*model.Transfer

	if stopPlace.Quays == nil {
		return transfers, nil
	}

	// Find accessible quays
	accessibleQuays := make([]*model.Quay, 0)
	regularQuays := make([]*model.Quay, 0)

	for i := range stopPlace.Quays.Quay {
		quay := &stopPlace.Quays.Quay[i]
		if p.isWheelchairAccessible(quay) {
			accessibleQuays = append(accessibleQuays, quay)
		} else {
			regularQuays = append(regularQuays, quay)
		}
	}

	// Create transfers that prioritize accessible routes
	for _, fromQuay := range accessibleQuays {
		for _, toQuay := range accessibleQuays {
			if fromQuay.ID == toQuay.ID {
				continue
			}

			transfer := p.createAccessibilityTransfer(fromQuay, toQuay, stopPlace)
			if transfer != nil {
				transfers = append(transfers, transfer)
			}
		}
	}

	// Create fallback transfers from accessible to non-accessible (with warnings)
	for _, fromQuay := range accessibleQuays {
		for _, toQuay := range regularQuays {
			transfer := p.createAccessibilityFallbackTransfer(fromQuay, toQuay, stopPlace)
			if transfer != nil {
				transfers = append(transfers, transfer)
			}
		}
	}

	return transfers, nil
}

// Helper methods

func (p *SophisticatedInterchangeProducer) createQuayToQuayTransfer(fromQuay, toQuay *model.Quay, stopPlace *model.StopPlace) (*model.Transfer, error) {
	transfer := &model.Transfer{
		FromStopID: fromQuay.ID,
		ToStopID:   toQuay.ID,
	}

	// Calculate base transfer time
	baseTime := p.calculateBaseTransferTime(fromQuay, toQuay, stopPlace)

	// Apply modifiers based on quay characteristics
	modifiedTime := p.applyTransferTimeModifiers(baseTime, fromQuay, toQuay, stopPlace)

	transfer.MinTransferTime = modifiedTime

	// Determine transfer type
	transferType := p.determineTransferType(fromQuay, toQuay, stopPlace)
	transfer.TransferType = transferType

	return transfer, nil
}

func (p *SophisticatedInterchangeProducer) calculateBaseTransferTime(fromQuay, toQuay *model.Quay, stopPlace *model.StopPlace) int {
	// Calculate walking time based on distance
	distance := p.calculateQuayDistance(fromQuay, toQuay)

	if distance > p.maximumWalkingDistance {
		// Too far to walk - might need shuttle or alternative
		return p.maximumConnectionTime
	}

	// Base walking time: 1.2 m/s average walking speed
	walkingTime := int(distance / 1.2)

	// Add platform-specific time
	platformTime := p.calculatePlatformTime(fromQuay, toQuay)

	// Ensure minimum time
	totalTime := walkingTime + platformTime
	if totalTime < p.minimumConnectionTime {
		totalTime = p.minimumConnectionTime
	}

	return totalTime
}

func (p *SophisticatedInterchangeProducer) applyTransferTimeModifiers(baseTime int, fromQuay, toQuay *model.Quay, stopPlace *model.StopPlace) int {
	modifiedTime := baseTime

	// Level change modifier
	levelDiff := p.calculateLevelDifference(fromQuay, toQuay)
	if levelDiff > 0 {
		modifiedTime += int(levelDiff) * p.levelChangeTime
	}

	// Cross-platform modifier
	if p.isCrossPlatformTransfer(fromQuay, toQuay) {
		modifiedTime += p.crossPlatformTime
	}

	// Mode change modifier (e.g., bus to train)
	if p.isDifferentMode(fromQuay, toQuay) {
		modifiedTime += 60 // Extra minute for mode change
	}

	// Accessibility modifier
	if p.requiresAccessibilityAccommodation(fromQuay, toQuay) {
		modifiedTime += p.wheelchairExtraTime
	}

	// Station complexity modifier
	complexityFactor := p.calculateStationComplexity(stopPlace)
	modifiedTime = int(float64(modifiedTime) * complexityFactor)

	return modifiedTime
}

func (p *SophisticatedInterchangeProducer) determineTransferType(fromQuay, toQuay *model.Quay, stopPlace *model.StopPlace) int {
	// GTFS transfer types:
	// 0 = Recommended transfer point
	// 1 = Timed transfer point
	// 2 = Minimum transfer time required
	// 3 = Transfer not possible

	distance := p.calculateQuayDistance(fromQuay, toQuay)

	if distance > p.maximumWalkingDistance {
		return 3 // Not possible
	}

	if p.isCrossPlatformTransfer(fromQuay, toQuay) {
		return 0 // Recommended (easy cross-platform)
	}

	if p.hasCoordinatedServices(fromQuay, toQuay) {
		return 1 // Timed transfer
	}

	return 2 // Minimum time required
}

func (p *SophisticatedInterchangeProducer) groupQuaysByMode(quays []model.Quay) map[string][]*model.Quay {
	modeGroups := make(map[string][]*model.Quay)

	for i := range quays {
		quay := &quays[i]
		mode := p.extractTransportMode(quay)
		modeGroups[mode] = append(modeGroups[mode], quay)
	}

	return modeGroups
}

func (p *SophisticatedInterchangeProducer) createModeTransfer(fromQuay, toQuay *model.Quay, fromMode, toMode string, stopPlace *model.StopPlace) *model.Transfer {
	// Create mode-specific transfer with appropriate timing
	baseTime := p.calculateBaseTransferTime(fromQuay, toQuay, stopPlace)

	// Add mode-specific transfer time
	modeTransferTime := p.getModeTransferTime(fromMode, toMode)
	totalTime := baseTime + modeTransferTime

	transfer := &model.Transfer{
		FromStopID:      fromQuay.ID,
		ToStopID:        toQuay.ID,
		TransferType:    2, // Minimum time required for mode changes
		MinTransferTime: totalTime,
	}

	return transfer
}

func (p *SophisticatedInterchangeProducer) createTimedTransfer(journey1, journey2 *model.ServiceJourney, stopPlace *model.StopPlace) *model.Transfer {
	// Create timed transfer between coordinated services
	fromQuay := p.getQuayForJourney(journey1, stopPlace)
	toQuay := p.getQuayForJourney(journey2, stopPlace)

	if fromQuay == nil || toQuay == nil {
		return nil
	}

	// Calculate coordinated transfer time
	coordinationTime := p.calculateCoordinatedTransferTime(journey1, journey2, stopPlace)

	transfer := &model.Transfer{
		FromStopID:      fromQuay.ID,
		ToStopID:        toQuay.ID,
		TransferType:    1, // Timed transfer
		MinTransferTime: coordinationTime,
	}

	return transfer
}

func (p *SophisticatedInterchangeProducer) createAccessibilityTransfer(fromQuay, toQuay *model.Quay, stopPlace *model.StopPlace) *model.Transfer {
	baseTime := p.calculateBaseTransferTime(fromQuay, toQuay, stopPlace)

	// Add accessibility time
	accessibilityTime := baseTime + p.wheelchairExtraTime

	transfer := &model.Transfer{
		FromStopID:      fromQuay.ID,
		ToStopID:        toQuay.ID,
		TransferType:    0, // Recommended for accessibility
		MinTransferTime: accessibilityTime,
	}

	return transfer
}

func (p *SophisticatedInterchangeProducer) createAccessibilityFallbackTransfer(fromQuay, toQuay *model.Quay, stopPlace *model.StopPlace) *model.Transfer {
	baseTime := p.calculateBaseTransferTime(fromQuay, toQuay, stopPlace)

	// Check if pathway exists
	if p.hasAccessiblePathway(fromQuay, toQuay) {
		return &model.Transfer{
			FromStopID:      fromQuay.ID,
			ToStopID:        toQuay.ID,
			TransferType:    2, // Minimum time (with accessibility caveat)
			MinTransferTime: baseTime + p.wheelchairExtraTime,
		}
	}

	// No accessible pathway - mark as not possible
	return &model.Transfer{
		FromStopID:   fromQuay.ID,
		ToStopID:     toQuay.ID,
		TransferType: 3, // Not possible
	}
}

func (p *SophisticatedInterchangeProducer) optimizeTransferTimes(transfers []*model.Transfer, stopPlace *model.StopPlace) []*model.Transfer {
	// Use pathways information to optimize transfer times
	if p.pathwaysProducer == nil {
		return transfers
	}

	pathways, err := p.pathwaysProducer.ProducePathwaysFromStopPlace(stopPlace)
	if err != nil || len(pathways) == 0 {
		return transfers
	}

	// Create pathway lookup map
	pathwayMap := make(map[string]*model.Pathway)
	for _, pathway := range pathways {
		key := fmt.Sprintf("%s_%s", pathway.FromStopID, pathway.ToStopID)
		pathwayMap[key] = pathway
	}

	// Optimize each transfer
	for _, transfer := range transfers {
		key := fmt.Sprintf("%s_%s", transfer.FromStopID, transfer.ToStopID)
		if pathway, exists := pathwayMap[key]; exists {
			// Use pathway traversal time if available and reasonable
			if pathway.TraversalTime > 0 && pathway.TraversalTime < transfer.MinTransferTime {
				transfer.MinTransferTime = pathway.TraversalTime

				// Adjust transfer type based on pathway mode
				if pathway.PathwayMode == 5 { // Elevator
					transfer.TransferType = 0 // Recommended for accessibility
				}
			}
		}
	}

	return transfers
}

// Utility methods

func (p *SophisticatedInterchangeProducer) calculateQuayDistance(fromQuay, toQuay *model.Quay) float64 {
	if fromQuay.Centroid == nil || toQuay.Centroid == nil ||
		fromQuay.Centroid.Location == nil || toQuay.Centroid.Location == nil {
		return 100.0 // Default 100m if no coordinates
	}

	// Haversine formula for distance
	lat1 := fromQuay.Centroid.Location.Latitude * math.Pi / 180
	lat2 := toQuay.Centroid.Location.Latitude * math.Pi / 180
	deltaLat := (toQuay.Centroid.Location.Latitude - fromQuay.Centroid.Location.Latitude) * math.Pi / 180
	deltaLon := (toQuay.Centroid.Location.Longitude - fromQuay.Centroid.Location.Longitude) * math.Pi / 180

	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1)*math.Cos(lat2)*
			math.Sin(deltaLon/2)*math.Sin(deltaLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return 6371000 * c // Earth radius in meters
}

func (p *SophisticatedInterchangeProducer) calculatePlatformTime(fromQuay, toQuay *model.Quay) int {
	// Time needed for platform-specific operations
	// - Disembarking: 30 seconds
	// - Walking on platform: 15 seconds
	// - Boarding: 45 seconds
	return 90 // Base platform time
}

func (p *SophisticatedInterchangeProducer) calculateLevelDifference(fromQuay, toQuay *model.Quay) float64 {
	// Extract level information and calculate difference
	fromLevel := p.extractLevelFromQuay(fromQuay)
	toLevel := p.extractLevelFromQuay(toQuay)

	return math.Abs(fromLevel - toLevel)
}

func (p *SophisticatedInterchangeProducer) extractLevelFromQuay(quay *model.Quay) float64 {
	// Extract level from quay name or properties
	name := strings.ToLower(quay.Name)

	if strings.Contains(name, "upper") {
		return 1.0
	}
	if strings.Contains(name, "lower") {
		return -1.0
	}
	return 0.0 // Ground level
}

func (p *SophisticatedInterchangeProducer) isCrossPlatformTransfer(fromQuay, toQuay *model.Quay) bool {
	// Check if quays are on opposite sides of the same platform
	fromName := strings.ToLower(fromQuay.Name)
	toName := strings.ToLower(toQuay.Name)

	// Simple heuristic: same platform number, different direction
	return strings.Contains(fromName, "platform") && strings.Contains(toName, "platform") &&
		(strings.Contains(fromName, "a") && strings.Contains(toName, "b") ||
			strings.Contains(fromName, "1") && strings.Contains(toName, "2"))
}

func (p *SophisticatedInterchangeProducer) isDifferentMode(fromQuay, toQuay *model.Quay) bool {
	fromMode := p.extractTransportMode(fromQuay)
	toMode := p.extractTransportMode(toQuay)
	return fromMode != toMode
}

func (p *SophisticatedInterchangeProducer) extractTransportMode(quay *model.Quay) string {
	// Extract transport mode from quay name or properties
	name := strings.ToLower(quay.Name)

	if strings.Contains(name, "bus") {
		return "bus"
	}
	if strings.Contains(name, "train") || strings.Contains(name, "railway") {
		return "train"
	}
	if strings.Contains(name, "metro") || strings.Contains(name, "subway") {
		return "metro"
	}
	if strings.Contains(name, "tram") {
		return "tram"
	}

	return "unknown"
}

func (p *SophisticatedInterchangeProducer) isWheelchairAccessible(quay *model.Quay) bool {
	if quay.AccessibilityAssessment == nil {
		return false
	}

	if quay.AccessibilityAssessment.Limitations == nil {
		return true // No limitations means accessible
	}

	if quay.AccessibilityAssessment.Limitations.AccessibilityLimitation == nil {
		return true
	}

	return strings.ToLower(quay.AccessibilityAssessment.Limitations.AccessibilityLimitation.WheelchairAccess) == "true"
}

func (p *SophisticatedInterchangeProducer) requiresAccessibilityAccommodation(fromQuay, toQuay *model.Quay) bool {
	return p.isWheelchairAccessible(fromQuay) || p.isWheelchairAccessible(toQuay)
}

func (p *SophisticatedInterchangeProducer) calculateStationComplexity(stopPlace *model.StopPlace) float64 {
	// Calculate complexity factor based on station characteristics
	complexity := 1.0

	if stopPlace.Quays != nil {
		quayCount := len(stopPlace.Quays.Quay)

		// More quays = more complex
		switch {
		case quayCount > 10:
			complexity += 0.3
		case quayCount > 5:
			complexity += 0.2
		case quayCount > 2:
			complexity += 0.1
		}
	}

	// Check for multi-level station
	levels := p.countLevels(stopPlace)
	if levels > 1 {
		complexity += float64(levels-1) * 0.15
	}

	// Check for multiple modes
	modes := p.countTransportModes(stopPlace)
	if modes > 1 {
		complexity += float64(modes-1) * 0.1
	}

	return complexity
}

func (p *SophisticatedInterchangeProducer) countLevels(stopPlace *model.StopPlace) int {
	if stopPlace.Quays == nil {
		return 1
	}

	levelSet := make(map[float64]bool)
	for _, quay := range stopPlace.Quays.Quay {
		level := p.extractLevelFromQuay(&quay)
		levelSet[level] = true
	}

	return len(levelSet)
}

func (p *SophisticatedInterchangeProducer) countTransportModes(stopPlace *model.StopPlace) int {
	if stopPlace.Quays == nil {
		return 1
	}

	modeSet := make(map[string]bool)
	for _, quay := range stopPlace.Quays.Quay {
		mode := p.extractTransportMode(&quay)
		modeSet[mode] = true
	}

	return len(modeSet)
}

func (p *SophisticatedInterchangeProducer) getModeTransferTime(fromMode, toMode string) int {
	// European transit mode-specific transfer times
	transferTimes := map[string]map[string]int{
		"train": {
			"bus":   300, // 5 minutes train to bus
			"metro": 180, // 3 minutes train to metro
			"tram":  240, // 4 minutes train to tram
		},
		"bus": {
			"train": 240, // 4 minutes bus to train
			"metro": 120, // 2 minutes bus to metro
			"tram":  90,  // 1.5 minutes bus to tram
		},
		"metro": {
			"train": 180, // 3 minutes metro to train
			"bus":   120, // 2 minutes metro to bus
			"tram":  90,  // 1.5 minutes metro to tram
		},
		"tram": {
			"train": 240, // 4 minutes tram to train
			"bus":   90,  // 1.5 minutes tram to bus
			"metro": 90,  // 1.5 minutes tram to metro
		},
	}

	if fromTimes, exists := transferTimes[fromMode]; exists {
		if transferTime, exists := fromTimes[toMode]; exists {
			return transferTime
		}
	}

	return 120 // Default 2 minutes for unknown mode combinations
}

// Placeholder methods - would be implemented with real service analysis

func (p *SophisticatedInterchangeProducer) getServiceJourneysForStopPlace(stopPlace *model.StopPlace) []*model.ServiceJourney {
	// This would query the repository for all service journeys serving this stop place
	return []*model.ServiceJourney{} // Placeholder
}

func (p *SophisticatedInterchangeProducer) hasCoordinatedTiming(journey1, journey2 *model.ServiceJourney, stopPlace *model.StopPlace) bool {
	// Analyze if two services have coordinated timing at this interchange
	return false // Placeholder
}

func (p *SophisticatedInterchangeProducer) hasCoordinatedServices(fromQuay, toQuay *model.Quay) bool {
	// Check if services at these quays are coordinated
	return false // Placeholder
}

func (p *SophisticatedInterchangeProducer) calculateCoordinatedTransferTime(journey1, journey2 *model.ServiceJourney, stopPlace *model.StopPlace) int {
	// Calculate transfer time for coordinated services
	return p.minimumConnectionTime // Placeholder
}

func (p *SophisticatedInterchangeProducer) getQuayForJourney(journey *model.ServiceJourney, stopPlace *model.StopPlace) *model.Quay {
	// Find which quay this journey uses at this stop place
	return nil // Placeholder
}

func (p *SophisticatedInterchangeProducer) hasAccessiblePathway(fromQuay, toQuay *model.Quay) bool {
	// Check if there's an accessible pathway between quays
	if p.pathwaysProducer == nil {
		return false
	}

	pathway, err := p.pathwaysProducer.ProduceAccessibilityPathways(fromQuay, toQuay)
	if err != nil || pathway == nil {
		return false
	}

	// Check if pathway is accessible (not stairs, or has elevator alternative)
	return pathway.PathwayMode != 2 || pathway.MaxSlope <= 0.08 // Not stairs or reasonable slope
}
