package validation

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/theoremus-urban-solutions/netex-gtfs-converter/model"
	"github.com/theoremus-urban-solutions/netex-gtfs-converter/producer"
)

// ValidationService provides comprehensive validation services for the conversion process
type ValidationService struct {
	validator *Validator
	reporter  *Reporter
	config    ServiceConfig
}

// ServiceConfig controls the validation service behavior
type ServiceConfig struct {
	EnableRealTimeValidation    bool `json:"enable_realtime_validation"`
	EnablePostProcessValidation bool `json:"enable_postprocess_validation"`
	EnableProgressReporting     bool `json:"enable_progress_reporting"`
	ValidateOnSave              bool `json:"validate_on_save"`
	ProgressUpdateInterval      int  `json:"progress_update_interval"` // Number of entities between progress updates
}

// ValidationContext provides context for validation operations
type ValidationContext struct {
	NetexRepository producer.NetexRepository
	GtfsRepository  producer.GtfsRepository
	ConversionStats ConversionStats
	StartTime       time.Time
}

// ConversionStats tracks detailed conversion statistics
type ConversionStats struct {
	NetexEntitiesLoaded     map[string]int           `json:"netex_entities_loaded"`
	GtfsEntitiesGenerated   map[string]int           `json:"gtfs_entities_generated"`
	ConversionErrors        map[string][]string      `json:"conversion_errors"`
	ProcessingTimes         map[string]time.Duration `json:"processing_times"`
	MemoryUsageByStage      map[string]float64       `json:"memory_usage_by_stage"`
	ValidationIssuesByStage map[string]int           `json:"validation_issues_by_stage"`
}

// NewValidationService creates a new validation service
func NewValidationService() *ValidationService {
	config := ServiceConfig{
		EnableRealTimeValidation:    true,
		EnablePostProcessValidation: true,
		EnableProgressReporting:     true,
		ValidateOnSave:              true,
		ProgressUpdateInterval:      1000, // Report progress every 1000 entities
	}

	return &ValidationService{
		validator: NewValidator(),
		reporter:  NewReporter(),
		config:    config,
	}
}

// SetConfig updates the service configuration
func (vs *ValidationService) SetConfig(config ServiceConfig) {
	vs.config = config
}

// SetValidatorConfig updates the validator configuration
func (vs *ValidationService) SetValidatorConfig(config ValidationConfig) {
	vs.validator.SetConfig(config)
}

// SetReporterConfig updates the reporter configuration
func (vs *ValidationService) SetReporterConfig(config ReporterConfig) {
	vs.reporter.SetConfig(config)
}

// StartConversion initializes validation for a new conversion process
func (vs *ValidationService) StartConversion() *ValidationContext {
	vs.validator.Reset()

	return &ValidationContext{
		ConversionStats: ConversionStats{
			NetexEntitiesLoaded:     make(map[string]int),
			GtfsEntitiesGenerated:   make(map[string]int),
			ConversionErrors:        make(map[string][]string),
			ProcessingTimes:         make(map[string]time.Duration),
			MemoryUsageByStage:      make(map[string]float64),
			ValidationIssuesByStage: make(map[string]int),
		},
		StartTime: time.Now(),
	}
}

// ValidateNeTExEntity validates a NeTEx entity during loading
func (vs *ValidationService) ValidateNeTExEntity(ctx *ValidationContext, entity interface{}) {
	if !vs.config.EnableRealTimeValidation {
		return
	}

	switch e := entity.(type) {
	case *model.Authority:
		vs.validateNeTExAuthority(e)
		ctx.ConversionStats.NetexEntitiesLoaded["Authority"]++
	case *model.Line:
		vs.validateNeTExLine(e)
		ctx.ConversionStats.NetexEntitiesLoaded["Line"]++
	case *model.Route:
		vs.validateNeTExRoute(e)
		ctx.ConversionStats.NetexEntitiesLoaded["Route"]++
	case *model.StopPlace:
		vs.validateNeTExStopPlace(e)
		ctx.ConversionStats.NetexEntitiesLoaded["StopPlace"]++
	case *model.Quay:
		vs.validateNeTExQuay(e)
		ctx.ConversionStats.NetexEntitiesLoaded["Quay"]++
	case *model.ServiceJourney:
		vs.validateNeTExServiceJourney(e)
		ctx.ConversionStats.NetexEntitiesLoaded["ServiceJourney"]++
	case *model.JourneyPattern:
		vs.validateNeTExJourneyPattern(e)
		ctx.ConversionStats.NetexEntitiesLoaded["JourneyPattern"]++
	case *model.HeadwayJourneyGroup:
		vs.validateNeTExHeadwayJourneyGroup(e)
		ctx.ConversionStats.NetexEntitiesLoaded["HeadwayJourneyGroup"]++
	default:
		// Unknown entity type
		vs.validator.AddIssue(ValidationIssue{
			Severity:   SeverityWarning,
			Code:       "NETEX_UNKNOWN_ENTITY_TYPE",
			Message:    fmt.Sprintf("Unknown NeTEx entity type: %T", entity),
			EntityType: "Unknown",
		})
	}

	// Update progress if enabled
	if vs.config.EnableProgressReporting {
		totalEntities := vs.getTotalLoadedEntities(ctx.ConversionStats.NetexEntitiesLoaded)
		if totalEntities%vs.config.ProgressUpdateInterval == 0 {
			vs.reportProgress(ctx, fmt.Sprintf("Loaded %d NeTEx entities", totalEntities))
		}
	}
}

// ValidateGTFSEntity validates a GTFS entity during conversion
func (vs *ValidationService) ValidateGTFSEntity(ctx *ValidationContext, entity interface{}) {
	if !vs.config.EnableRealTimeValidation {
		return
	}

	switch e := entity.(type) {
	case *model.Agency:
		vs.validator.ValidateGTFSAgency(e)
		ctx.ConversionStats.GtfsEntitiesGenerated["Agency"]++
	case *model.GtfsRoute:
		vs.validator.ValidateGTFSRoute(e)
		ctx.ConversionStats.GtfsEntitiesGenerated["Route"]++
	case *model.Stop:
		vs.validator.ValidateGTFSStop(e)
		ctx.ConversionStats.GtfsEntitiesGenerated["Stop"]++
	case *model.Trip:
		vs.validateGTFSTrip(e)
		ctx.ConversionStats.GtfsEntitiesGenerated["Trip"]++
	case *model.StopTime:
		vs.validator.ValidateGTFSStopTime(e)
		ctx.ConversionStats.GtfsEntitiesGenerated["StopTime"]++
	case *model.Calendar:
		vs.validateGTFSCalendar(e)
		ctx.ConversionStats.GtfsEntitiesGenerated["Calendar"]++
	case *model.Shape:
		vs.validateGTFSShape(e)
		ctx.ConversionStats.GtfsEntitiesGenerated["Shape"]++
	case *model.Frequency:
		vs.validateGTFSFrequency(e)
		ctx.ConversionStats.GtfsEntitiesGenerated["Frequency"]++
	case *model.Transfer:
		vs.validateGTFSTransfer(e)
		ctx.ConversionStats.GtfsEntitiesGenerated["Transfer"]++
	case *model.Pathway:
		vs.validateGTFSPathway(e)
		ctx.ConversionStats.GtfsEntitiesGenerated["Pathway"]++
	default:
		vs.validator.AddIssue(ValidationIssue{
			Severity:   SeverityWarning,
			Code:       "GTFS_UNKNOWN_ENTITY_TYPE",
			Message:    fmt.Sprintf("Unknown GTFS entity type: %T", entity),
			EntityType: "Unknown",
		})
	}

	// Update progress if enabled
	if vs.config.EnableProgressReporting {
		totalEntities := vs.getTotalLoadedEntities(ctx.ConversionStats.GtfsEntitiesGenerated)
		if totalEntities%vs.config.ProgressUpdateInterval == 0 {
			vs.reportProgress(ctx, fmt.Sprintf("Generated %d GTFS entities", totalEntities))
		}
	}
}

// RecordConversionError records an error during conversion
func (vs *ValidationService) RecordConversionError(ctx *ValidationContext, stage string, entity interface{}, err error) {
	entityType := fmt.Sprintf("%T", entity)

	if _, exists := ctx.ConversionStats.ConversionErrors[stage]; !exists {
		ctx.ConversionStats.ConversionErrors[stage] = make([]string, 0)
	}

	errorMsg := fmt.Sprintf("%s: %s", entityType, err.Error())
	ctx.ConversionStats.ConversionErrors[stage] = append(ctx.ConversionStats.ConversionErrors[stage], errorMsg)

	// Add as validation issue
	vs.validator.AddIssue(ValidationIssue{
		Severity:   SeverityError,
		Code:       "CONVERSION_ERROR",
		Message:    fmt.Sprintf("Conversion error in stage %s: %s", stage, err.Error()),
		EntityType: entityType,
		Location:   stage,
	})
}

// RecordProcessingTime records processing time for a conversion stage
func (vs *ValidationService) RecordProcessingTime(ctx *ValidationContext, stage string, duration time.Duration) {
	ctx.ConversionStats.ProcessingTimes[stage] = duration

	// Check for performance issues
	if duration > 30*time.Second {
		vs.validator.AddIssue(ValidationIssue{
			Severity:   SeverityWarning,
			Code:       "PERFORMANCE_SLOW_STAGE",
			Message:    fmt.Sprintf("Stage %s took %.2f seconds to complete", stage, duration.Seconds()),
			EntityType: "Performance",
			Location:   stage,
			Suggestion: "Consider optimizing this processing stage",
		})
	}
}

// RecordMemoryUsage records memory usage at a specific stage
func (vs *ValidationService) RecordMemoryUsage(ctx *ValidationContext, stage string) {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	memoryMB := float64(memStats.Alloc) / (1024 * 1024)
	ctx.ConversionStats.MemoryUsageByStage[stage] = memoryMB

	// Check for memory issues
	if memoryMB > 500 { // More than 500MB
		vs.validator.AddIssue(ValidationIssue{
			Severity:   SeverityWarning,
			Code:       "PERFORMANCE_HIGH_MEMORY",
			Message:    fmt.Sprintf("High memory usage at stage %s: %.2f MB", stage, memoryMB),
			EntityType: "Performance",
			Location:   stage,
			Suggestion: "Monitor memory usage to prevent out-of-memory issues",
		})
	}
}

// FinishConversion completes the validation process and generates the final report
func (vs *ValidationService) FinishConversion(ctx *ValidationContext) ValidationReport {
	duration := time.Since(ctx.StartTime)

	// Record final statistics
	vs.validator.UpdateProcessingStats("Total",
		vs.getTotalLoadedEntities(ctx.ConversionStats.NetexEntitiesLoaded),
		vs.getTotalLoadedEntities(ctx.ConversionStats.GtfsEntitiesGenerated),
		0, // TODO: Calculate skipped entities
	)

	// Update processing duration and memory usage
	report := vs.validator.GetReport()
	report.ProcessingStats.ProcessingDuration = duration

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	report.ProcessingStats.MemoryUsageMB = float64(memStats.Alloc) / (1024 * 1024)

	// Add conversion statistics to the report
	vs.addConversionStatsToReport(&report, ctx.ConversionStats)

	return report
}

// GetCurrentReport returns the current validation report
func (vs *ValidationService) GetCurrentReport() ValidationReport {
	return vs.validator.GetReport()
}

// GenerateReport generates a formatted validation report
func (vs *ValidationService) GenerateReport(report ValidationReport, format ReportFormat) (string, error) {
	var output strings.Builder
	err := vs.reporter.GenerateReport(report, format, &output)
	return output.String(), err
}

// NeTEx entity validation methods

func (vs *ValidationService) validateNeTExAuthority(authority *model.Authority) {
	if authority.ID == "" {
		vs.validator.AddIssue(ValidationIssue{
			Severity:   SeverityError,
			Code:       "NETEX_AUTHORITY_MISSING_ID",
			Message:    "NeTEx Authority missing required ID",
			EntityType: "Authority",
		})
	}

	if authority.Name == "" {
		vs.validator.AddIssue(ValidationIssue{
			Severity:   SeverityError,
			Code:       "NETEX_AUTHORITY_MISSING_NAME",
			Message:    "NeTEx Authority missing required name",
			EntityType: "Authority",
			EntityID:   authority.ID,
		})
	}
}

func (vs *ValidationService) validateNeTExLine(line *model.Line) {
	if line.ID == "" {
		vs.validator.AddIssue(ValidationIssue{
			Severity:   SeverityError,
			Code:       "NETEX_LINE_MISSING_ID",
			Message:    "NeTEx Line missing required ID",
			EntityType: "Line",
		})
	}

	if line.Name == "" && line.ShortName == "" {
		vs.validator.AddIssue(ValidationIssue{
			Severity:   SeverityWarning,
			Code:       "NETEX_LINE_MISSING_NAME",
			Message:    "NeTEx Line missing both name and short name",
			EntityType: "Line",
			EntityID:   line.ID,
			Suggestion: "Provide at least one name field for better GTFS conversion",
		})
	}

	if line.AuthorityRef == "" {
		vs.validator.AddIssue(ValidationIssue{
			Severity:   SeverityWarning,
			Code:       "NETEX_LINE_MISSING_AUTHORITY_REF",
			Message:    "NeTEx Line missing authority reference",
			EntityType: "Line",
			EntityID:   line.ID,
			Suggestion: "Authority reference helps with GTFS agency mapping",
		})
	}
}

func (vs *ValidationService) validateNeTExRoute(route *model.Route) {
	if route.ID == "" {
		vs.validator.AddIssue(ValidationIssue{
			Severity:   SeverityError,
			Code:       "NETEX_ROUTE_MISSING_ID",
			Message:    "NeTEx Route missing required ID",
			EntityType: "Route",
		})
	}

	if route.LineRef.Ref == "" {
		vs.validator.AddIssue(ValidationIssue{
			Severity:   SeverityError,
			Code:       "NETEX_ROUTE_MISSING_LINE_REF",
			Message:    "NeTEx Route missing line reference",
			EntityType: "Route",
			EntityID:   route.ID,
		})
	}
}

func (vs *ValidationService) validateNeTExStopPlace(stopPlace *model.StopPlace) {
	if stopPlace.ID == "" {
		vs.validator.AddIssue(ValidationIssue{
			Severity:   SeverityError,
			Code:       "NETEX_STOPPLACE_MISSING_ID",
			Message:    "NeTEx StopPlace missing required ID",
			EntityType: "StopPlace",
		})
	}

	if stopPlace.Name == "" {
		vs.validator.AddIssue(ValidationIssue{
			Severity:   SeverityWarning,
			Code:       "NETEX_STOPPLACE_MISSING_NAME",
			Message:    "NeTEx StopPlace missing name",
			EntityType: "StopPlace",
			EntityID:   stopPlace.ID,
		})
	}

	// Validate coordinates if present
	if stopPlace.Centroid != nil && stopPlace.Centroid.Location != nil {
		loc := stopPlace.Centroid.Location
		if loc.Latitude < -90 || loc.Latitude > 90 {
			vs.validator.AddIssue(ValidationIssue{
				Severity:   SeverityError,
				Code:       "NETEX_STOPPLACE_INVALID_LATITUDE",
				Message:    "StopPlace latitude out of valid range",
				EntityType: "StopPlace",
				EntityID:   stopPlace.ID,
				Value:      fmt.Sprintf("%.6f", loc.Latitude),
			})
		}

		if loc.Longitude < -180 || loc.Longitude > 180 {
			vs.validator.AddIssue(ValidationIssue{
				Severity:   SeverityError,
				Code:       "NETEX_STOPPLACE_INVALID_LONGITUDE",
				Message:    "StopPlace longitude out of valid range",
				EntityType: "StopPlace",
				EntityID:   stopPlace.ID,
				Value:      fmt.Sprintf("%.6f", loc.Longitude),
			})
		}
	}
}

func (vs *ValidationService) validateNeTExQuay(quay *model.Quay) {
	if quay.ID == "" {
		vs.validator.AddIssue(ValidationIssue{
			Severity:   SeverityError,
			Code:       "NETEX_QUAY_MISSING_ID",
			Message:    "NeTEx Quay missing required ID",
			EntityType: "Quay",
		})
	}

	if quay.Name == "" {
		vs.validator.AddIssue(ValidationIssue{
			Severity:   SeverityWarning,
			Code:       "NETEX_QUAY_MISSING_NAME",
			Message:    "NeTEx Quay missing name",
			EntityType: "Quay",
			EntityID:   quay.ID,
		})
	}
}

func (vs *ValidationService) validateNeTExServiceJourney(journey *model.ServiceJourney) {
	if journey.ID == "" {
		vs.validator.AddIssue(ValidationIssue{
			Severity:   SeverityError,
			Code:       "NETEX_SERVICEJOURNEY_MISSING_ID",
			Message:    "NeTEx ServiceJourney missing required ID",
			EntityType: "ServiceJourney",
		})
	}

	if journey.JourneyPatternRef.Ref == "" {
		vs.validator.AddIssue(ValidationIssue{
			Severity:   SeverityError,
			Code:       "NETEX_SERVICEJOURNEY_MISSING_PATTERN_REF",
			Message:    "NeTEx ServiceJourney missing journey pattern reference",
			EntityType: "ServiceJourney",
			EntityID:   journey.ID,
		})
	}

	if journey.PassingTimes == nil || len(journey.PassingTimes.TimetabledPassingTime) == 0 {
		vs.validator.AddIssue(ValidationIssue{
			Severity:   SeverityError,
			Code:       "NETEX_SERVICEJOURNEY_MISSING_PASSING_TIMES",
			Message:    "NeTEx ServiceJourney missing passing times",
			EntityType: "ServiceJourney",
			EntityID:   journey.ID,
		})
	}
}

func (vs *ValidationService) validateNeTExJourneyPattern(pattern *model.JourneyPattern) {
	if pattern.ID == "" {
		vs.validator.AddIssue(ValidationIssue{
			Severity:   SeverityError,
			Code:       "NETEX_JOURNEYPATTERN_MISSING_ID",
			Message:    "NeTEx JourneyPattern missing required ID",
			EntityType: "JourneyPattern",
		})
	}

	if pattern.PointsInSequence == nil || len(pattern.PointsInSequence.PointInJourneyPatternOrStopPointInJourneyPatternOrTimingPointInJourneyPattern) == 0 {
		vs.validator.AddIssue(ValidationIssue{
			Severity:   SeverityError,
			Code:       "NETEX_JOURNEYPATTERN_MISSING_POINTS",
			Message:    "NeTEx JourneyPattern missing points in sequence",
			EntityType: "JourneyPattern",
			EntityID:   pattern.ID,
		})
	}
}

func (vs *ValidationService) validateNeTExHeadwayJourneyGroup(group *model.HeadwayJourneyGroup) {
	if group.ID == "" {
		vs.validator.AddIssue(ValidationIssue{
			Severity:   SeverityError,
			Code:       "NETEX_HEADWAYGROUP_MISSING_ID",
			Message:    "NeTEx HeadwayJourneyGroup missing required ID",
			EntityType: "HeadwayJourneyGroup",
		})
	}

	if group.ScheduledHeadwayInterval == "" {
		vs.validator.AddIssue(ValidationIssue{
			Severity:   SeverityError,
			Code:       "NETEX_HEADWAYGROUP_MISSING_INTERVAL",
			Message:    "NeTEx HeadwayJourneyGroup missing scheduled headway interval",
			EntityType: "HeadwayJourneyGroup",
			EntityID:   group.ID,
		})
	}
}

// Additional GTFS entity validation methods

func (vs *ValidationService) validateGTFSTrip(trip *model.Trip) {
	if trip.TripID == "" {
		vs.validator.AddIssue(ValidationIssue{
			Severity:   SeverityError,
			Code:       "TRIP_MISSING_ID",
			Message:    "Trip ID is required but missing",
			EntityType: "Trip",
		})
	}

	if trip.RouteID == "" {
		vs.validator.AddIssue(ValidationIssue{
			Severity:   SeverityError,
			Code:       "TRIP_MISSING_ROUTE_ID",
			Message:    "Trip route ID is required but missing",
			EntityType: "Trip",
			EntityID:   trip.TripID,
		})
	}

	if trip.ServiceID == "" {
		vs.validator.AddIssue(ValidationIssue{
			Severity:   SeverityError,
			Code:       "TRIP_MISSING_SERVICE_ID",
			Message:    "Trip service ID is required but missing",
			EntityType: "Trip",
			EntityID:   trip.TripID,
		})
	}
}

func (vs *ValidationService) validateGTFSCalendar(calendar *model.Calendar) {
	if calendar.ServiceID == "" {
		vs.validator.AddIssue(ValidationIssue{
			Severity:   SeverityError,
			Code:       "CALENDAR_MISSING_SERVICE_ID",
			Message:    "Calendar service ID is required but missing",
			EntityType: "Calendar",
		})
	}
}

func (vs *ValidationService) validateGTFSShape(shape *model.Shape) {
	if shape.ShapeID == "" {
		vs.validator.AddIssue(ValidationIssue{
			Severity:   SeverityError,
			Code:       "SHAPE_MISSING_ID",
			Message:    "Shape ID is required but missing",
			EntityType: "Shape",
		})
	}
}

func (vs *ValidationService) validateGTFSFrequency(frequency *model.Frequency) {
	if frequency.TripID == "" {
		vs.validator.AddIssue(ValidationIssue{
			Severity:   SeverityError,
			Code:       "FREQUENCY_MISSING_TRIP_ID",
			Message:    "Frequency trip ID is required but missing",
			EntityType: "Frequency",
		})
	}

	if frequency.HeadwaySecs <= 0 {
		vs.validator.AddIssue(ValidationIssue{
			Severity:   SeverityError,
			Code:       "FREQUENCY_INVALID_HEADWAY",
			Message:    "Frequency headway must be positive",
			EntityType: "Frequency",
			EntityID:   frequency.TripID,
			Value:      fmt.Sprintf("%d", frequency.HeadwaySecs),
		})
	}
}

func (vs *ValidationService) validateGTFSTransfer(transfer *model.Transfer) {
	if transfer.FromStopID == "" {
		vs.validator.AddIssue(ValidationIssue{
			Severity:   SeverityError,
			Code:       "TRANSFER_MISSING_FROM_STOP",
			Message:    "Transfer from_stop_id is required but missing",
			EntityType: "Transfer",
		})
	}

	if transfer.ToStopID == "" {
		vs.validator.AddIssue(ValidationIssue{
			Severity:   SeverityError,
			Code:       "TRANSFER_MISSING_TO_STOP",
			Message:    "Transfer to_stop_id is required but missing",
			EntityType: "Transfer",
		})
	}
}

func (vs *ValidationService) validateGTFSPathway(pathway *model.Pathway) {
	if pathway.PathwayID == "" {
		vs.validator.AddIssue(ValidationIssue{
			Severity:   SeverityError,
			Code:       "PATHWAY_MISSING_ID",
			Message:    "Pathway ID is required but missing",
			EntityType: "Pathway",
		})
	}

	if pathway.FromStopID == "" || pathway.ToStopID == "" {
		vs.validator.AddIssue(ValidationIssue{
			Severity:   SeverityError,
			Code:       "PATHWAY_MISSING_STOP_IDS",
			Message:    "Pathway requires both from_stop_id and to_stop_id",
			EntityType: "Pathway",
			EntityID:   pathway.PathwayID,
		})
	}
}

// Utility methods

func (vs *ValidationService) getTotalLoadedEntities(entityCounts map[string]int) int {
	total := 0
	for _, count := range entityCounts {
		total += count
	}
	return total
}

func (vs *ValidationService) reportProgress(ctx *ValidationContext, message string) {
	// Progress reporting implementation would go here
	// This could be a callback, log entry, or event emission
	fmt.Printf("Validation Progress: %s (%.2fs elapsed)\n", message, time.Since(ctx.StartTime).Seconds())
}

func (vs *ValidationService) addConversionStatsToReport(report *ValidationReport, stats ConversionStats) {
	// Add conversion statistics as context to the report
	// This could include detailed timing information, memory usage patterns, etc.

	// Add processing time issues
	for stage, duration := range stats.ProcessingTimes {
		if duration > time.Minute {
			vs.validator.AddIssue(ValidationIssue{
				Severity:   SeverityInfo,
				Code:       "PROCESSING_TIME_INFO",
				Message:    fmt.Sprintf("Stage %s completed in %.2f seconds", stage, duration.Seconds()),
				EntityType: "Performance",
				Location:   stage,
			})
		}
	}

	// Add memory usage information
	maxMemoryStage := ""
	maxMemoryUsage := 0.0
	for stage, usage := range stats.MemoryUsageByStage {
		if usage > maxMemoryUsage {
			maxMemoryUsage = usage
			maxMemoryStage = stage
		}
	}

	if maxMemoryUsage > 0 {
		vs.validator.AddIssue(ValidationIssue{
			Severity:   SeverityInfo,
			Code:       "MEMORY_USAGE_INFO",
			Message:    fmt.Sprintf("Peak memory usage: %.2f MB at stage %s", maxMemoryUsage, maxMemoryStage),
			EntityType: "Performance",
			Location:   maxMemoryStage,
		})
	}
}
