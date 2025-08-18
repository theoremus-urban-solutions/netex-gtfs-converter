package errors

import (
	"fmt"
	"strings"
	"time"
)

// ConversionError represents an error during conversion with context
type ConversionError struct {
	Stage      string    `json:"stage"`
	EntityType string    `json:"entity_type"`
	EntityID   string    `json:"entity_id"`
	FieldName  string    `json:"field_name,omitempty"`
	Err        error     `json:"error"`
	Timestamp  time.Time `json:"timestamp"`
	Severity   Severity  `json:"severity"`
	Recoverable bool     `json:"recoverable"`
}

// Severity levels for conversion errors
type Severity int

const (
	SeverityInfo Severity = iota
	SeverityWarning
	SeverityError
	SeverityFatal
)

func (s Severity) String() string {
	switch s {
	case SeverityInfo:
		return "INFO"
	case SeverityWarning:
		return "WARNING"
	case SeverityError:
		return "ERROR"
	case SeverityFatal:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

func (ce *ConversionError) Error() string {
	return fmt.Sprintf("[%s] %s %s/%s: %v", 
		ce.Severity.String(), ce.Stage, ce.EntityType, ce.EntityID, ce.Err)
}

// ConversionResult holds the result of a conversion with errors and warnings
type ConversionResult struct {
	Success        bool                `json:"success"`
	Errors         []*ConversionError  `json:"errors"`
	Warnings       []*ConversionError  `json:"warnings"`
	ProcessedCount map[string]int      `json:"processed_count"`
	SkippedCount   map[string]int      `json:"skipped_count"`
	StartTime      time.Time           `json:"start_time"`
	EndTime        time.Time           `json:"end_time"`
	Duration       time.Duration       `json:"duration"`
}

// NewConversionResult creates a new conversion result
func NewConversionResult() *ConversionResult {
	return &ConversionResult{
		Success:        true,
		Errors:         make([]*ConversionError, 0),
		Warnings:       make([]*ConversionError, 0),
		ProcessedCount: make(map[string]int),
		SkippedCount:   make(map[string]int),
		StartTime:      time.Now(),
	}
}

// AddError adds an error to the conversion result
func (cr *ConversionResult) AddError(stage, entityType, entityID string, err error, recoverable bool) {
	convErr := &ConversionError{
		Stage:       stage,
		EntityType:  entityType,
		EntityID:    entityID,
		Err:         err,
		Timestamp:   time.Now(),
		Severity:    SeverityError,
		Recoverable: recoverable,
	}
	
	cr.Errors = append(cr.Errors, convErr)
	
	if !recoverable {
		cr.Success = false
	}
}

// AddWarning adds a warning to the conversion result
func (cr *ConversionResult) AddWarning(stage, entityType, entityID string, message string) {
	convErr := &ConversionError{
		Stage:       stage,
		EntityType:  entityType,
		EntityID:    entityID,
		Err:         fmt.Errorf(message),
		Timestamp:   time.Now(),
		Severity:    SeverityWarning,
		Recoverable: true,
	}
	
	cr.Warnings = append(cr.Warnings, convErr)
}

// AddFieldError adds an error for a specific field
func (cr *ConversionResult) AddFieldError(stage, entityType, entityID, fieldName string, err error, recoverable bool) {
	convErr := &ConversionError{
		Stage:       stage,
		EntityType:  entityType,
		EntityID:    entityID,
		FieldName:   fieldName,
		Err:         err,
		Timestamp:   time.Now(),
		Severity:    SeverityError,
		Recoverable: recoverable,
	}
	
	cr.Errors = append(cr.Errors, convErr)
	
	if !recoverable {
		cr.Success = false
	}
}

// IncrementProcessed increments the processed count for an entity type
func (cr *ConversionResult) IncrementProcessed(entityType string) {
	cr.ProcessedCount[entityType]++
}

// IncrementSkipped increments the skipped count for an entity type
func (cr *ConversionResult) IncrementSkipped(entityType string) {
	cr.SkippedCount[entityType]++
}

// Finalize marks the conversion as complete
func (cr *ConversionResult) Finalize() {
	cr.EndTime = time.Now()
	cr.Duration = cr.EndTime.Sub(cr.StartTime)
}

// HasFatalErrors returns true if there are any fatal errors
func (cr *ConversionResult) HasFatalErrors() bool {
	for _, err := range cr.Errors {
		if !err.Recoverable {
			return true
		}
	}
	return false
}

// GetSummary returns a summary of the conversion result
func (cr *ConversionResult) GetSummary() string {
	var summary strings.Builder
	
	summary.WriteString(fmt.Sprintf("Conversion Summary:\n"))
	summary.WriteString(fmt.Sprintf("  Success: %t\n", cr.Success))
	summary.WriteString(fmt.Sprintf("  Duration: %v\n", cr.Duration))
	summary.WriteString(fmt.Sprintf("  Errors: %d\n", len(cr.Errors)))
	summary.WriteString(fmt.Sprintf("  Warnings: %d\n", len(cr.Warnings)))
	
	if len(cr.ProcessedCount) > 0 {
		summary.WriteString("  Processed entities:\n")
		for entityType, count := range cr.ProcessedCount {
			summary.WriteString(fmt.Sprintf("    %s: %d\n", entityType, count))
		}
	}
	
	if len(cr.SkippedCount) > 0 {
		summary.WriteString("  Skipped entities:\n")
		for entityType, count := range cr.SkippedCount {
			summary.WriteString(fmt.Sprintf("    %s: %d\n", entityType, count))
		}
	}
	
	return summary.String()
}

// GetErrorsByEntityType groups errors by entity type
func (cr *ConversionResult) GetErrorsByEntityType() map[string][]*ConversionError {
	errorsByType := make(map[string][]*ConversionError)
	
	for _, err := range cr.Errors {
		errorsByType[err.EntityType] = append(errorsByType[err.EntityType], err)
	}
	
	return errorsByType
}

// GetErrorsBySeverity groups errors by severity
func (cr *ConversionResult) GetErrorsBySeverity() map[Severity][]*ConversionError {
	errorsBySeverity := make(map[Severity][]*ConversionError)
	
	for _, err := range cr.Errors {
		errorsBySeverity[err.Severity] = append(errorsBySeverity[err.Severity], err)
	}
	
	for _, warning := range cr.Warnings {
		errorsBySeverity[warning.Severity] = append(errorsBySeverity[warning.Severity], warning)
	}
	
	return errorsBySeverity
}

// Recovery strategies for different types of errors
type RecoveryStrategy interface {
	CanRecover(err *ConversionError) bool
	Recover(err *ConversionError, data interface{}) (interface{}, error)
}

// DefaultValueStrategy provides default values for missing required fields
type DefaultValueStrategy struct {
	defaults map[string]interface{}
}

// NewDefaultValueStrategy creates a strategy with default values
func NewDefaultValueStrategy() *DefaultValueStrategy {
	return &DefaultValueStrategy{
		defaults: map[string]interface{}{
			"agency_name":     "Unknown Agency",
			"agency_timezone": "UTC",
			"route_type":      3, // Bus
			"trip_id":         "unknown",
			"stop_name":       "Unknown Stop",
		},
	}
}

func (dvs *DefaultValueStrategy) CanRecover(err *ConversionError) bool {
	_, exists := dvs.defaults[err.FieldName]
	return exists && err.Recoverable
}

func (dvs *DefaultValueStrategy) Recover(err *ConversionError, data interface{}) (interface{}, error) {
	if defaultValue, exists := dvs.defaults[err.FieldName]; exists {
		return defaultValue, nil
	}
	return nil, fmt.Errorf("no default value available for field: %s", err.FieldName)
}

// SkipEntityStrategy skips problematic entities
type SkipEntityStrategy struct {
	skippableEntities map[string]bool
}

// NewSkipEntityStrategy creates a strategy that skips certain entity types
func NewSkipEntityStrategy() *SkipEntityStrategy {
	return &SkipEntityStrategy{
		skippableEntities: map[string]bool{
			"ServiceJourneyInterchange": true,
			"DestinationDisplay":        true,
			"DayTypeAssignment":         true,
		},
	}
}

func (ses *SkipEntityStrategy) CanRecover(err *ConversionError) bool {
	return ses.skippableEntities[err.EntityType] && err.Recoverable
}

func (ses *SkipEntityStrategy) Recover(err *ConversionError, data interface{}) (interface{}, error) {
	// Return nil to indicate the entity should be skipped
	return nil, nil
}

// RecoveryManager manages multiple recovery strategies
type RecoveryManager struct {
	strategies []RecoveryStrategy
	result     *ConversionResult
}

// NewRecoveryManager creates a new recovery manager
func NewRecoveryManager(result *ConversionResult) *RecoveryManager {
	rm := &RecoveryManager{
		result:     result,
		strategies: make([]RecoveryStrategy, 0),
	}
	
	// Add default strategies
	rm.AddStrategy(NewDefaultValueStrategy())
	rm.AddStrategy(NewSkipEntityStrategy())
	
	return rm
}

// AddStrategy adds a recovery strategy
func (rm *RecoveryManager) AddStrategy(strategy RecoveryStrategy) {
	rm.strategies = append(rm.strategies, strategy)
}

// TryRecover attempts to recover from an error using available strategies
func (rm *RecoveryManager) TryRecover(stage, entityType, entityID string, err error, data interface{}) (interface{}, bool) {
	convErr := &ConversionError{
		Stage:       stage,
		EntityType:  entityType,
		EntityID:    entityID,
		Err:         err,
		Timestamp:   time.Now(),
		Severity:    SeverityError,
		Recoverable: true, // Assume recoverable initially
	}
	
	// Try each recovery strategy
	for _, strategy := range rm.strategies {
		if strategy.CanRecover(convErr) {
			recoveredData, recoveryErr := strategy.Recover(convErr, data)
			if recoveryErr == nil {
				rm.result.AddWarning(stage, entityType, entityID, 
					fmt.Sprintf("Recovered from error: %v", err))
				return recoveredData, true
			}
		}
	}
	
	// No recovery possible
	convErr.Recoverable = false
	rm.result.Errors = append(rm.result.Errors, convErr)
	return nil, false
}

// SafeFieldAccess provides safe access to struct fields with recovery
func (rm *RecoveryManager) SafeFieldAccess(entityType, entityID, fieldName string, getValue func() (interface{}, error)) interface{} {
	value, err := getValue()
	if err != nil {
		// Create error with field name for recovery strategies
		convErr := &ConversionError{
			Stage:       "field_access",
			EntityType:  entityType,
			EntityID:    entityID,
			FieldName:   fieldName,
			Err:         err,
			Timestamp:   time.Now(),
			Severity:    SeverityError,
			Recoverable: true,
		}
		
		// Try each recovery strategy
		for _, strategy := range rm.strategies {
			if strategy.CanRecover(convErr) {
				recoveredValue, recoveryErr := strategy.Recover(convErr, nil)
				if recoveryErr == nil {
					rm.result.AddWarning("field_access", entityType, entityID, 
						fmt.Sprintf("Recovered field %s: %v", fieldName, err))
					return recoveredValue
				}
			}
		}
		
		// No recovery possible
		convErr.Recoverable = false
		rm.result.Errors = append(rm.result.Errors, convErr)
		return nil
	}
	return value
}

// ValidateAndRecover validates an entity and attempts recovery if validation fails
func (rm *RecoveryManager) ValidateAndRecover(entityType, entityID string, data interface{}, validator func(interface{}) error) (interface{}, bool) {
	err := validator(data)
	if err != nil {
		recoveredData, recovered := rm.TryRecover("validation", entityType, entityID, err, data)
		if recovered {
			// Re-validate recovered data
			if validateErr := validator(recoveredData); validateErr == nil {
				return recoveredData, true
			}
		}
		return nil, false
	}
	return data, true
}