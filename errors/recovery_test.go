package errors

import (
	"fmt"
	"testing"
	"time"
)

func TestConversionResult_AddError(t *testing.T) {
	result := NewConversionResult()
	
	err := fmt.Errorf("test error")
	result.AddError("test_stage", "test_entity", "test_id", err, true)
	
	if len(result.Errors) != 1 {
		t.Fatalf("Expected 1 error, got %d", len(result.Errors))
	}
	
	convErr := result.Errors[0]
	if convErr.Stage != "test_stage" {
		t.Errorf("Expected stage 'test_stage', got '%s'", convErr.Stage)
	}
	
	if convErr.EntityType != "test_entity" {
		t.Errorf("Expected entity type 'test_entity', got '%s'", convErr.EntityType)
	}
	
	if convErr.EntityID != "test_id" {
		t.Errorf("Expected entity ID 'test_id', got '%s'", convErr.EntityID)
	}
	
	if convErr.Severity != SeverityError {
		t.Errorf("Expected severity Error, got %v", convErr.Severity)
	}
	
	if !convErr.Recoverable {
		t.Error("Expected error to be recoverable")
	}
	
	if !result.Success {
		t.Error("Expected success to remain true for recoverable error")
	}
}

func TestConversionResult_AddNonRecoverableError(t *testing.T) {
	result := NewConversionResult()
	
	err := fmt.Errorf("fatal error")
	result.AddError("test_stage", "test_entity", "test_id", err, false)
	
	if result.Success {
		t.Error("Expected success to be false for non-recoverable error")
	}
	
	if result.HasFatalErrors() != true {
		t.Error("Expected to have fatal errors")
	}
}

func TestConversionResult_AddWarning(t *testing.T) {
	result := NewConversionResult()
	
	result.AddWarning("test_stage", "test_entity", "test_id", "test warning")
	
	if len(result.Warnings) != 1 {
		t.Fatalf("Expected 1 warning, got %d", len(result.Warnings))
	}
	
	warning := result.Warnings[0]
	if warning.Severity != SeverityWarning {
		t.Errorf("Expected severity Warning, got %v", warning.Severity)
	}
	
	if !warning.Recoverable {
		t.Error("Expected warning to be recoverable")
	}
	
	if !result.Success {
		t.Error("Expected success to remain true for warning")
	}
}

func TestConversionResult_IncrementCounters(t *testing.T) {
	result := NewConversionResult()
	
	result.IncrementProcessed("test_entity")
	result.IncrementProcessed("test_entity")
	result.IncrementSkipped("test_entity")
	
	if result.ProcessedCount["test_entity"] != 2 {
		t.Errorf("Expected processed count 2, got %d", result.ProcessedCount["test_entity"])
	}
	
	if result.SkippedCount["test_entity"] != 1 {
		t.Errorf("Expected skipped count 1, got %d", result.SkippedCount["test_entity"])
	}
}

func TestConversionResult_Finalize(t *testing.T) {
	result := NewConversionResult()
	startTime := result.StartTime
	
	// Wait a small amount to ensure duration > 0
	time.Sleep(1 * time.Millisecond)
	
	result.Finalize()
	
	if result.EndTime.Before(startTime) {
		t.Error("End time should be after start time")
	}
	
	if result.Duration <= 0 {
		t.Error("Duration should be greater than 0")
	}
}

func TestConversionResult_GetSummary(t *testing.T) {
	result := NewConversionResult()
	
	result.AddError("test", "entity1", "id1", fmt.Errorf("error1"), true)
	result.AddWarning("test", "entity2", "id2", "warning1")
	result.IncrementProcessed("entity1")
	result.IncrementSkipped("entity2")
	result.Finalize()
	
	summary := result.GetSummary()
	
	if summary == "" {
		t.Error("Summary should not be empty")
	}
	
	// Check that summary contains expected information
	expectedStrings := []string{
		"Conversion Summary:",
		"Success:",
		"Duration:",
		"Errors: 1",
		"Warnings: 1",
		"Processed entities:",
		"entity1: 1",
		"Skipped entities:",
		"entity2: 1",
	}
	
	for _, expected := range expectedStrings {
		if !contains(summary, expected) {
			t.Errorf("Summary should contain '%s'", expected)
		}
	}
}

func TestDefaultValueStrategy(t *testing.T) {
	strategy := NewDefaultValueStrategy()
	
	// Test recoverable error with known field
	err := &ConversionError{
		FieldName:   "agency_name",
		Recoverable: true,
	}
	
	if !strategy.CanRecover(err) {
		t.Error("Strategy should be able to recover agency_name")
	}
	
	recovered, recErr := strategy.Recover(err, nil)
	if recErr != nil {
		t.Fatalf("Recovery failed: %v", recErr)
	}
	
	if recovered != "Unknown Agency" {
		t.Errorf("Expected 'Unknown Agency', got '%v'", recovered)
	}
	
	// Test non-recoverable error
	err.Recoverable = false
	if strategy.CanRecover(err) {
		t.Error("Strategy should not recover non-recoverable error")
	}
	
	// Test unknown field
	err.FieldName = "unknown_field"
	err.Recoverable = true
	if strategy.CanRecover(err) {
		t.Error("Strategy should not recover unknown field")
	}
}

func TestSkipEntityStrategy(t *testing.T) {
	strategy := NewSkipEntityStrategy()
	
	// Test recoverable error with skippable entity
	err := &ConversionError{
		EntityType:  "ServiceJourneyInterchange",
		Recoverable: true,
	}
	
	if !strategy.CanRecover(err) {
		t.Error("Strategy should be able to recover ServiceJourneyInterchange")
	}
	
	recovered, recErr := strategy.Recover(err, nil)
	if recErr != nil {
		t.Fatalf("Recovery failed: %v", recErr)
	}
	
	if recovered != nil {
		t.Errorf("Expected nil (skip), got %v", recovered)
	}
	
	// Test non-skippable entity
	err.EntityType = "Agency"
	if strategy.CanRecover(err) {
		t.Error("Strategy should not recover non-skippable entity")
	}
}

func TestRecoveryManager(t *testing.T) {
	result := NewConversionResult()
	manager := NewRecoveryManager(result)
	
	// Test successful recovery with default value strategy
	testData := struct{ Name string }{Name: ""}
	_, recovered := manager.TryRecover("test", "agency", "test_id", 
		fmt.Errorf("missing name"), testData)
	
	// This should not recover because the error doesn't match the field pattern
	// But it should be added to the result as a non-recoverable error
	if recovered {
		t.Error("Should not have recovered generic error")
	}
	
	if len(result.Errors) != 1 {
		t.Fatalf("Expected 1 error in result, got %d", len(result.Errors))
	}
	
	// Test SafeFieldAccess
	value := manager.SafeFieldAccess("test_entity", "test_id", "agency_name", 
		func() (interface{}, error) {
			return nil, fmt.Errorf("field not found")
		})
	
	// Should return recovered default value
	if value != "Unknown Agency" {
		t.Errorf("Expected default agency name, got %v", value)
	}
	
	// Test ValidateAndRecover
	validData, valid := manager.ValidateAndRecover("test_entity", "test_id", "valid_data",
		func(data interface{}) error {
			return nil // validation passes
		})
	
	if !valid {
		t.Error("Valid data should pass validation")
	}
	
	if validData != "valid_data" {
		t.Errorf("Expected original data, got %v", validData)
	}
}

func TestSeverityString(t *testing.T) {
	tests := []struct {
		severity Severity
		expected string
	}{
		{SeverityInfo, "INFO"},
		{SeverityWarning, "WARNING"},
		{SeverityError, "ERROR"},
		{SeverityFatal, "FATAL"},
		{Severity(999), "UNKNOWN"},
	}
	
	for _, test := range tests {
		if test.severity.String() != test.expected {
			t.Errorf("Expected %s, got %s", test.expected, test.severity.String())
		}
	}
}

func TestConversionError_Error(t *testing.T) {
	err := &ConversionError{
		Stage:      "test_stage",
		EntityType: "test_entity",
		EntityID:   "test_id",
		Err:        fmt.Errorf("test error"),
		Severity:   SeverityError,
	}
	
	expected := "[ERROR] test_stage test_entity/test_id: test error"
	if err.Error() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, err.Error())
	}
}

func TestConversionResult_GetErrorsByEntityType(t *testing.T) {
	result := NewConversionResult()
	
	result.AddError("stage1", "entity1", "id1", fmt.Errorf("error1"), true)
	result.AddError("stage2", "entity1", "id2", fmt.Errorf("error2"), true)
	result.AddError("stage3", "entity2", "id3", fmt.Errorf("error3"), true)
	
	errorsByType := result.GetErrorsByEntityType()
	
	if len(errorsByType) != 2 {
		t.Fatalf("Expected 2 entity types, got %d", len(errorsByType))
	}
	
	if len(errorsByType["entity1"]) != 2 {
		t.Errorf("Expected 2 errors for entity1, got %d", len(errorsByType["entity1"]))
	}
	
	if len(errorsByType["entity2"]) != 1 {
		t.Errorf("Expected 1 error for entity2, got %d", len(errorsByType["entity2"]))
	}
}

func TestConversionResult_GetErrorsBySeverity(t *testing.T) {
	result := NewConversionResult()
	
	result.AddError("stage1", "entity1", "id1", fmt.Errorf("error1"), true)
	result.AddError("stage2", "entity1", "id2", fmt.Errorf("error2"), false)
	result.AddWarning("stage3", "entity2", "id3", "warning1")
	
	errorsBySeverity := result.GetErrorsBySeverity()
	
	if len(errorsBySeverity[SeverityError]) != 2 {
		t.Errorf("Expected 2 errors, got %d", len(errorsBySeverity[SeverityError]))
	}
	
	if len(errorsBySeverity[SeverityWarning]) != 1 {
		t.Errorf("Expected 1 warning, got %d", len(errorsBySeverity[SeverityWarning]))
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr) != -1
}

func findSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if s[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}