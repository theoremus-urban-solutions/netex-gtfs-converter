package exporter

import (
	"strings"
	"testing"

	"github.com/theoremus-urban-solutions/netex-gtfs-converter/repository"
)

func TestNewDefaultGtfsExporter(t *testing.T) {
	stopAreaRepo := repository.NewDefaultStopAreaRepository()
	exporter := NewDefaultGtfsExporter("TEST", stopAreaRepo)

	if exporter == nil {
		t.Fatal("NewDefaultGtfsExporter() returned nil")
	}

	if exporter.codespace != "TEST" {
		t.Errorf("Expected codespace 'TEST', got '%s'", exporter.codespace)
	}

	if exporter.netexRepository == nil {
		t.Error("NeTEx repository should not be nil")
	}

	if exporter.gtfsRepository == nil {
		t.Error("GTFS repository should not be nil")
	}

	if exporter.stopAreaRepository == nil {
		t.Error("Stop area repository should not be nil")
	}

	// Check that producers are initialized
	if exporter.agencyProducer == nil {
		t.Error("Agency producer should not be nil")
	}

	if exporter.routeProducer == nil {
		t.Error("Route producer should not be nil")
	}

	if exporter.stopProducer == nil {
		t.Error("Stop producer should not be nil")
	}
}

func TestDefaultGtfsExporter_ConvertStopsToGtfs(t *testing.T) {
	stopAreaRepo := repository.NewDefaultStopAreaRepository()
	exporter := NewDefaultGtfsExporter("TEST", stopAreaRepo)

	// This should work even with empty data (creates default agency and empty stops file)
	result, err := exporter.ConvertStopsToGtfs()
	if err != nil {
		t.Fatalf("ConvertStopsToGtfs() failed: %v", err)
	}

	if result == nil {
		t.Fatal("ConvertStopsToGtfs() returned nil result")
	}

	// Read some data to verify it's a valid ZIP
	buffer := make([]byte, 4)
	n, err := result.Read(buffer)
	if err != nil && err.Error() != eofError {
		t.Fatalf("Failed to read result: %v", err)
	}

	if n < 4 {
		t.Fatal("Result too small to be a valid ZIP")
	}

	// Check ZIP magic bytes
	if buffer[0] != 'P' || buffer[1] != 'K' {
		t.Error("Result does not appear to be a ZIP file")
	}
}

func TestDefaultGtfsExporter_ConvertTimetablesToGtfs_MissingCodespace(t *testing.T) {
	stopAreaRepo := repository.NewDefaultStopAreaRepository()
	exporter := NewDefaultGtfsExporter("", stopAreaRepo) // Empty codespace

	netexData := strings.NewReader("<xml></xml>")
	_, err := exporter.ConvertTimetablesToGtfs(netexData)

	if err == nil {
		t.Fatal("Expected error for missing codespace")
	}

	if err != ErrMissingCodespace {
		t.Errorf("Expected ErrMissingCodespace, got %v", err)
	}
}

func TestValidationError(t *testing.T) {
	err := ValidationError{
		Field:   "TestField",
		Value:   "TestValue",
		Message: "Test message",
	}

	expected := "validation error in field TestField (value: TestValue): Test message"
	if err.Error() != expected {
		t.Errorf("Expected error message '%s', got '%s'", expected, err.Error())
	}

	// Test without value
	err2 := ValidationError{
		Field:   "TestField",
		Message: "Test message",
	}

	expected2 := "validation error in field TestField: Test message"
	if err2.Error() != expected2 {
		t.Errorf("Expected error message '%s', got '%s'", expected2, err2.Error())
	}
}

func TestConversionError(t *testing.T) {
	innerErr := ValidationError{Field: "test", Message: "inner error"}

	err := ConversionError{
		Stage:    "TestStage",
		EntityID: "TestEntity",
		Err:      innerErr,
	}

	expected := "conversion error in stage TestStage for entity TestEntity: validation error in field test: inner error"
	if err.Error() != expected {
		t.Errorf("Expected error message '%s', got '%s'", expected, err.Error())
	}

	// Test unwrapping
	if err.Unwrap() != innerErr {
		t.Error("Unwrap() should return the inner error")
	}

	// Test without entity ID
	err2 := ConversionError{
		Stage: "TestStage",
		Err:   innerErr,
	}

	expected2 := "conversion error in stage TestStage: validation error in field test: inner error"
	if err2.Error() != expected2 {
		t.Errorf("Expected error message '%s', got '%s'", expected2, err2.Error())
	}
}
