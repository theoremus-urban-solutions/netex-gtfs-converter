package producer

import (
	"strconv"
	"strings"

	"github.com/theoremus-urban-solutions/netex-gtfs-converter/model"
)

// EuropeanAccessibilityProducer handles European accessibility features
type EuropeanAccessibilityProducer struct {
	netexRepo NetexRepository
	gtfsRepo  GtfsRepository
}

// NewEuropeanAccessibilityProducer creates a producer for accessibility information
func NewEuropeanAccessibilityProducer(netexRepo NetexRepository, gtfsRepo GtfsRepository) *EuropeanAccessibilityProducer {
	return &EuropeanAccessibilityProducer{
		netexRepo: netexRepo,
		gtfsRepo:  gtfsRepo,
	}
}

// ProduceAccessibilityInfo converts NeTEx accessibility assessment to GTFS wheelchair accessibility
func (p *EuropeanAccessibilityProducer) ProduceAccessibilityInfo(assessment *model.AccessibilityAssessment, entityType string) string {
	if assessment == nil {
		return "0" // Unknown
	}

	// Check limitations for wheelchair access
	if assessment.Limitations != nil && assessment.Limitations.AccessibilityLimitation != nil {
		limitation := assessment.Limitations.AccessibilityLimitation

		// Check wheelchair access
		switch strings.ToLower(limitation.WheelchairAccess) {
		case "true":
			return "1" // Accessible
		case "false":
			return "2" // Not accessible
		}

		// Check step-free access as alternative indicator
		switch strings.ToLower(limitation.StepFreeAccess) {
		case "true":
			return "1" // Accessible
		case "false":
			return "2" // Not accessible
		}
	}

	return "0" // Unknown
}

// EuropeanVehicleProducer handles European vehicle type information
type EuropeanVehicleProducer struct {
	netexRepo NetexRepository
	gtfsRepo  GtfsRepository
}

// NewEuropeanVehicleProducer creates a producer for vehicle information
func NewEuropeanVehicleProducer(netexRepo NetexRepository, gtfsRepo GtfsRepository) *EuropeanVehicleProducer {
	return &EuropeanVehicleProducer{
		netexRepo: netexRepo,
		gtfsRepo:  gtfsRepo,
	}
}

// ProduceVehicleAccessibility determines wheelchair accessibility from vehicle type
func (p *EuropeanVehicleProducer) ProduceVehicleAccessibility(vehicleType *model.VehicleType) string {
	if vehicleType == nil {
		return "0" // Unknown
	}

	// Check explicit wheelchair accessibility
	switch strings.ToLower(vehicleType.WheelchairAccessible) {
	case "true":
		return "1" // Accessible
	case "false":
		return "2" // Not accessible
	case "partial":
		// Check if there are wheelchair places
		if vehicleType.VehicleTypeCapacity != nil {
			wheelchairPlaces, _ := strconv.Atoi(vehicleType.VehicleTypeCapacity.WheelchairPlaces)
			if wheelchairPlaces > 0 {
				return "1" // Accessible if there are designated spaces
			}
		}
		return "2" // Treat partial as not accessible
	}

	// Check low floor and lift/ramp indicators
	hasLowFloor := strings.ToLower(vehicleType.LowFloor) == "true" ||
		strings.ToLower(vehicleType.LowFloor) == "partial"
	hasLiftOrRamp := strings.ToLower(vehicleType.HasLiftOrRamp) == "true"

	if hasLowFloor || hasLiftOrRamp {
		// Check if there are wheelchair spaces
		if vehicleType.VehicleTypeCapacity != nil {
			wheelchairPlaces, _ := strconv.Atoi(vehicleType.VehicleTypeCapacity.WheelchairPlaces)
			if wheelchairPlaces > 0 {
				return "1" // Accessible
			}
		}
		// Has accessibility features but no confirmed spaces
		return "1" // Assume accessible if vehicle has accessibility features
	}

	return "0" // Unknown
}

// ProduceBikesAllowed determines if bikes are allowed from vehicle type
func (p *EuropeanVehicleProducer) ProduceBikesAllowed(vehicleType *model.VehicleType) string {
	if vehicleType == nil {
		return "0" // No information
	}

	// Check explicit cycles allowed
	switch strings.ToLower(vehicleType.CyclesAllowed) {
	case "true":
		return "1" // Bicycles allowed
	case "false":
		return "2" // Bicycles not allowed
	}

	// Check if there are bicycle places
	if vehicleType.VehicleTypeCapacity != nil {
		bicyclePlaces, _ := strconv.Atoi(vehicleType.VehicleTypeCapacity.BicyclePlaces)
		if bicyclePlaces > 0 {
			return "1" // Bicycles allowed
		}
	}

	// Check cycle storage equipment
	if vehicleType.CycleStorageEquipment != nil {
		spaces, _ := strconv.Atoi(vehicleType.CycleStorageEquipment.NumberOfSpaces)
		if spaces > 0 {
			return "1" // Bicycles allowed
		}
	}

	return "0" // No information
}

// EuropeanServiceAlterationProducer handles service alterations and disruptions
type EuropeanServiceAlterationProducer struct {
	netexRepo NetexRepository
	gtfsRepo  GtfsRepository
}

// NewEuropeanServiceAlterationProducer creates a producer for service alterations
func NewEuropeanServiceAlterationProducer(netexRepo NetexRepository, gtfsRepo GtfsRepository) *EuropeanServiceAlterationProducer {
	return &EuropeanServiceAlterationProducer{
		netexRepo: netexRepo,
		gtfsRepo:  gtfsRepo,
	}
}

// ProduceCalendarExceptions converts service alterations to GTFS calendar exceptions
func (p *EuropeanServiceAlterationProducer) ProduceCalendarExceptions(alteration *model.ServiceAlteration) ([]*model.CalendarDate, error) {
	if alteration == nil {
		return nil, nil
	}

	var exceptions []*model.CalendarDate

	// Extract dates from the alteration validity period
	if alteration.ValidFrom == "" || alteration.ValidTo == "" {
		return exceptions, nil
	}

	// Handle different alteration types
	switch strings.ToLower(alteration.AlterationType) {
	case "cancellation":
		// Create service removal exceptions
		for _, serviceJourneyRef := range alteration.AffectedServiceJourneys {
			// For simplicity, use ValidFrom date (in practice would expand date range)
			exception := &model.CalendarDate{
				ServiceID:     serviceJourneyRef,
				Date:          alteration.ValidFrom,
				ExceptionType: 2, // Service removed for this date
			}
			exceptions = append(exceptions, exception)
		}

	case "extrajourney":
		// Create service addition exceptions
		for _, serviceJourneyRef := range alteration.AffectedServiceJourneys {
			exception := &model.CalendarDate{
				ServiceID:     serviceJourneyRef,
				Date:          alteration.ValidFrom,
				ExceptionType: 1, // Service added for this date
			}
			exceptions = append(exceptions, exception)
		}

	default:
		// Unknown alteration type - create generic exceptions
		for _, serviceJourneyRef := range alteration.AffectedServiceJourneys {
			exception := &model.CalendarDate{
				ServiceID:     serviceJourneyRef,
				Date:          alteration.ValidFrom,
				ExceptionType: 2, // Remove service as default safe option
			}
			exceptions = append(exceptions, exception)
		}
	}

	return exceptions, nil
}

// EuropeanNoticeProducer handles notices and passenger information
type EuropeanNoticeProducer struct {
	netexRepo NetexRepository
	gtfsRepo  GtfsRepository
}

// NewEuropeanNoticeProducer creates a producer for notices
func NewEuropeanNoticeProducer(netexRepo NetexRepository, gtfsRepo GtfsRepository) *EuropeanNoticeProducer {
	return &EuropeanNoticeProducer{
		netexRepo: netexRepo,
		gtfsRepo:  gtfsRepo,
	}
}

// ProduceTranslations converts European multilingual notices to GTFS translations
func (p *EuropeanNoticeProducer) ProduceTranslations(notices []*model.Notice) ([]*model.Translation, error) {
	var translations []*model.Translation

	for _, notice := range notices {
		if notice == nil {
			continue
		}

		// Process delivery variants as different language versions
		for _, variant := range notice.DeliveryVariants {
			if variant.NoticeText != "" {
				// Try to extract language from media type or use heuristics
				language := p.extractLanguageFromMediaType(variant.DeliveryVariantMediaType)
				if language == "" {
					language = "en" // Default to English
				}

				// Create translation entry
				translation := &model.Translation{
					TableName:   "notices",
					FieldName:   "notice_text",
					Language:    language,
					Translation: variant.NoticeText,
					RecordID:    notice.ID,
					RecordSubID: "",
					FieldValue:  notice.Text, // Original text as reference
				}

				translations = append(translations, translation)
			}
		}
	}

	return translations, nil
}

// extractLanguageFromMediaType attempts to extract language from media type
func (p *EuropeanNoticeProducer) extractLanguageFromMediaType(mediaType string) string {
	// Simple heuristic - in practice, this would be more sophisticated
	mediaType = strings.ToLower(mediaType)

	if strings.Contains(mediaType, "en") || strings.Contains(mediaType, "english") {
		return "en"
	}
	if strings.Contains(mediaType, "de") || strings.Contains(mediaType, "german") || strings.Contains(mediaType, "deutsch") {
		return "de"
	}
	if strings.Contains(mediaType, "french") || strings.Contains(mediaType, "français") {
		return "fr"
	}
	if strings.Contains(mediaType, "fr") && !strings.Contains(mediaType, "german") {
		return "fr"
	}
	if strings.Contains(mediaType, "es") || strings.Contains(mediaType, "spanish") || strings.Contains(mediaType, "español") {
		return "es"
	}
	if strings.Contains(mediaType, "it") || strings.Contains(mediaType, "italian") || strings.Contains(mediaType, "italiano") {
		return "it"
	}
	if strings.Contains(mediaType, "dutch") || strings.Contains(mediaType, "nederlands") {
		return "nl"
	}
	if strings.Contains(mediaType, "nl") && !strings.Contains(mediaType, "english") {
		return "nl"
	}

	return ""
}

// EuropeanFlexibleServiceProducer handles demand-responsive transport
type EuropeanFlexibleServiceProducer struct {
	netexRepo NetexRepository
	gtfsRepo  GtfsRepository
}

// NewEuropeanFlexibleServiceProducer creates a producer for flexible services
func NewEuropeanFlexibleServiceProducer(netexRepo NetexRepository, gtfsRepo GtfsRepository) *EuropeanFlexibleServiceProducer {
	return &EuropeanFlexibleServiceProducer{
		netexRepo: netexRepo,
		gtfsRepo:  gtfsRepo,
	}
}

// ProduceBookingRules converts flexible service booking arrangements to GTFS booking rules
func (p *EuropeanFlexibleServiceProducer) ProduceBookingRules(flexibleService *model.FlexibleService) (*model.BookingRule, error) {
	if flexibleService == nil || flexibleService.BookingArrangements == nil {
		return nil, nil
	}

	arrangements := flexibleService.BookingArrangements

	// Create GTFS booking rule
	bookingRule := &model.BookingRule{
		BookingRuleID:          flexibleService.ID + "_booking",
		BookingType:            p.mapBookingType(flexibleService.FlexibleServiceType),
		PriorNoticeDurationMin: p.convertDurationToMinutes(arrangements.MinimumBookingPeriod),
		PriorNoticeDurationMax: p.convertDurationToMinutes(arrangements.LatestBookingTime),
		PriorNoticeLastDay:     0,  // Default
		PriorNoticeLastTime:    "", // Not specified
		PriorNoticeStartDay:    0,  // Default
		PriorNoticeStartTime:   "", // Not specified
		PriorNoticeServiceID:   "", // Not applicable
		Message:                arrangements.BookingNote,
		PickupMessage:          "",
		DropOffMessage:         "",
		PhoneNumber:            "",
		InfoURL:                "",
	}

	// Add contact information
	if arrangements.BookingContact != nil {
		contact := arrangements.BookingContact
		if contact.Phone != "" {
			bookingRule.PhoneNumber = contact.Phone
		}
		if contact.Url != "" {
			bookingRule.InfoURL = contact.Url
		}
		if contact.FurtherDetails != "" && bookingRule.Message == "" {
			bookingRule.Message = contact.FurtherDetails
		}
	}

	return bookingRule, nil
}

// mapBookingType maps NeTEx flexible service type to GTFS booking type
func (p *EuropeanFlexibleServiceProducer) mapBookingType(flexibleServiceType string) int {
	switch strings.ToLower(flexibleServiceType) {
	case "dynamicpassingtimes":
		return 2 // Phone agency, at least X min before
	case "fixedheadwayfrequency":
		return 1 // Real time
	case "hailandride":
		return 0 // Real time
	default:
		return 2 // Default to advance booking
	}
}

// convertDurationToMinutes converts ISO duration to minutes
func (p *EuropeanFlexibleServiceProducer) convertDurationToMinutes(isoDuration string) int {
	if isoDuration == "" {
		return 0
	}

	// Simple conversion - in practice would use proper ISO 8601 duration parsing
	// Example: "PT30M" = 30 minutes, "PT2H" = 120 minutes
	duration := strings.ToUpper(isoDuration)

	if strings.Contains(duration, "PT") && strings.Contains(duration, "M") {
		// Extract minutes
		start := strings.Index(duration, "PT") + 2
		end := strings.Index(duration, "M")
		if end > start {
			if minutes, err := strconv.Atoi(duration[start:end]); err == nil {
				return minutes
			}
		}
	}

	if strings.Contains(duration, "PT") && strings.Contains(duration, "H") {
		// Extract hours and convert to minutes
		start := strings.Index(duration, "PT") + 2
		end := strings.Index(duration, "H")
		if end > start {
			if hours, err := strconv.Atoi(duration[start:end]); err == nil {
				return hours * 60
			}
		}
	}

	return 0
}

// ProduceLocationGroups creates GTFS location groups from flexible areas
func (p *EuropeanFlexibleServiceProducer) ProduceLocationGroups(flexibleService *model.FlexibleService) ([]*model.LocationGroup, error) {
	if flexibleService == nil || flexibleService.FlexibleArea == nil {
		return nil, nil
	}

	area := flexibleService.FlexibleArea

	var locationGroups []*model.LocationGroup

	// Create main area group
	mainGroup := &model.LocationGroup{
		LocationGroupID:   area.ID,
		LocationGroupName: area.Name,
	}

	locationGroups = append(locationGroups, mainGroup)

	return locationGroups, nil
}

// ProduceStopAreas creates GTFS stop areas from flexible quays
func (p *EuropeanFlexibleServiceProducer) ProduceStopAreas(flexibleService *model.FlexibleService) ([]*model.StopArea, error) {
	if flexibleService == nil || flexibleService.FlexibleArea == nil {
		return nil, nil
	}

	area := flexibleService.FlexibleArea
	var stopAreas []*model.StopArea

	for _, flexibleQuay := range area.FlexibleQuays {
		stopArea := &model.StopArea{
			AreaID:   flexibleQuay.ID,
			AreaName: flexibleQuay.Name,
		}

		// Set coordinates if available
		if flexibleQuay.Centroid != nil && flexibleQuay.Centroid.Location != nil {
			stopArea.AreaLat = flexibleQuay.Centroid.Location.Latitude
			stopArea.AreaLon = flexibleQuay.Centroid.Location.Longitude
		}

		stopAreas = append(stopAreas, stopArea)
	}

	return stopAreas, nil
}
