package exporter

import (
	"errors"
	"fmt"
)

// Error definitions
var (
	ErrMissingCodespace = errors.New("missing required codespace for timetable data export")
	ErrInvalidNetexData = errors.New("invalid NeTEx data format")
	ErrConversionFailed = errors.New("conversion from NeTEx to GTFS failed")
	ErrNoDataFound      = errors.New("no data found in NeTEx dataset")
	ErrInvalidProfile   = errors.New("invalid profile configuration")
	ErrMissingStopData  = errors.New("no stop data found in dataset")
	ErrInvalidXML       = errors.New("invalid XML format in NeTEx data")
	ErrEmptyDataset     = errors.New("dataset contains no usable data")
	ErrInvalidZIP       = errors.New("invalid or corrupted ZIP archive")
)

// ValidationError represents validation errors with context
type ValidationError struct {
	Field   string
	Value   string
	Message string
}

func (e ValidationError) Error() string {
	if e.Value != "" {
		return fmt.Sprintf("validation error in field %s (value: %s): %s", e.Field, e.Value, e.Message)
	}
	return fmt.Sprintf("validation error in field %s: %s", e.Field, e.Message)
}

// ConversionError represents errors that occur during conversion
type ConversionError struct {
	Stage   string
	EntityID string
	Err     error
}

func (e ConversionError) Error() string {
	if e.EntityID != "" {
		return fmt.Sprintf("conversion error in stage %s for entity %s: %v", e.Stage, e.EntityID, e.Err)
	}
	return fmt.Sprintf("conversion error in stage %s: %v", e.Stage, e.Err)
}

func (e ConversionError) Unwrap() error {
	return e.Err
}
