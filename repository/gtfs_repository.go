package repository

import (
	"archive/zip"
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"

	"github.com/theoremus-urban-solutions/netex-gtfs-converter/model"
	"github.com/theoremus-urban-solutions/netex-gtfs-converter/producer"
)

// DefaultGtfsRepository implements GtfsRepository
type DefaultGtfsRepository struct {
	// Entity storage
	agencies       map[string]*model.Agency
	routes         map[string]*model.GtfsRoute
	trips          map[string]*model.Trip
	stops          map[string]*model.Stop
	stopTimes      []*model.StopTime
	calendars      map[string]*model.Calendar
	calendarDates  []*model.CalendarDate
	transfers      []*model.Transfer
	shapes         []*model.Shape
	feedInfo       *model.FeedInfo
	frequencies    []*model.Frequency
	fareAttributes map[string]*model.FareAttribute
	fareRules      []*model.FareRule
	pathways       []*model.Pathway
	levels         []*model.Level

	// Default agency
	defaultAgency *model.Agency
}

// NewDefaultGtfsRepository creates a new DefaultGtfsRepository
func NewDefaultGtfsRepository() producer.GtfsRepository {
	return &DefaultGtfsRepository{
		agencies:       make(map[string]*model.Agency),
		routes:         make(map[string]*model.GtfsRoute),
		trips:          make(map[string]*model.Trip),
		stops:          make(map[string]*model.Stop),
		stopTimes:      make([]*model.StopTime, 0),
		calendars:      make(map[string]*model.Calendar),
		calendarDates:  make([]*model.CalendarDate, 0),
		transfers:      make([]*model.Transfer, 0),
		shapes:         make([]*model.Shape, 0),
		frequencies:    make([]*model.Frequency, 0),
		fareAttributes: make(map[string]*model.FareAttribute),
		fareRules:      make([]*model.FareRule, 0),
		pathways:       make([]*model.Pathway, 0),
		levels:         make([]*model.Level, 0),
	}
}

// SaveEntity saves an entity to the repository
func (r *DefaultGtfsRepository) SaveEntity(entity interface{}) error {
	switch e := entity.(type) {
	case *model.Agency:
		r.agencies[e.AgencyID] = e
		if r.defaultAgency == nil {
			r.defaultAgency = e
		}
	case *model.GtfsRoute:
		r.routes[e.RouteID] = e
	case *model.Trip:
		r.trips[e.TripID] = e
	case *model.Stop:
		r.stops[e.StopID] = e
	case *model.StopTime:
		r.stopTimes = append(r.stopTimes, e)
	case *model.Calendar:
		r.calendars[e.ServiceID] = e
	case *model.CalendarDate:
		r.calendarDates = append(r.calendarDates, e)
	case *model.Transfer:
		r.transfers = append(r.transfers, e)
	case *model.Shape:
		r.shapes = append(r.shapes, e)
	case *model.FeedInfo:
		r.feedInfo = e
	case *model.Frequency:
		r.frequencies = append(r.frequencies, e)
	case *model.FareAttribute:
		r.fareAttributes[e.FareID] = e
	case *model.FareRule:
		r.fareRules = append(r.fareRules, e)
	case *model.Pathway:
		r.pathways = append(r.pathways, e)
	case *model.Level:
		r.levels = append(r.levels, e)
	default:
		return fmt.Errorf("unknown GTFS entity type: %T", entity)
	}
	return nil
}

// GetAgencyById returns an agency by ID
func (r *DefaultGtfsRepository) GetAgencyById(id string) *model.Agency {
	return r.agencies[id]
}

// GetTripById returns a trip by ID
func (r *DefaultGtfsRepository) GetTripById(id string) *model.Trip {
	return r.trips[id]
}

// GetStopById returns a stop by ID
func (r *DefaultGtfsRepository) GetStopById(id string) *model.Stop {
	return r.stops[id]
}

// GetDefaultAgency returns the default agency
func (r *DefaultGtfsRepository) GetDefaultAgency() *model.Agency {
	return r.defaultAgency
}

// WriteGtfs generates a GTFS ZIP archive
func (r *DefaultGtfsRepository) WriteGtfs() (io.Reader, error) {
	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)

	// Write each GTFS file
	if err := r.writeAgencies(zipWriter); err != nil {
		return nil, fmt.Errorf("failed to write agencies: %w", err)
	}

	if err := r.writeStops(zipWriter); err != nil {
		return nil, fmt.Errorf("failed to write stops: %w", err)
	}

	if err := r.writeRoutes(zipWriter); err != nil {
		return nil, fmt.Errorf("failed to write routes: %w", err)
	}

	if err := r.writeTrips(zipWriter); err != nil {
		return nil, fmt.Errorf("failed to write trips: %w", err)
	}

	if err := r.writeStopTimes(zipWriter); err != nil {
		return nil, fmt.Errorf("failed to write stop times: %w", err)
	}

	if err := r.writeCalendar(zipWriter); err != nil {
		return nil, fmt.Errorf("failed to write calendar: %w", err)
	}

	if err := r.writeCalendarDates(zipWriter); err != nil {
		return nil, fmt.Errorf("failed to write calendar dates: %w", err)
	}

	if err := r.writeTransfers(zipWriter); err != nil {
		return nil, fmt.Errorf("failed to write transfers: %w", err)
	}

	if err := r.writeShapes(zipWriter); err != nil {
		return nil, fmt.Errorf("failed to write shapes: %w", err)
	}

	if err := r.writeFrequencies(zipWriter); err != nil {
		return nil, fmt.Errorf("failed to write frequencies: %w", err)
	}

	if err := r.writePathways(zipWriter); err != nil {
		return nil, fmt.Errorf("failed to write pathways: %w", err)
	}

	if err := r.writeLevels(zipWriter); err != nil {
		return nil, fmt.Errorf("failed to write levels: %w", err)
	}

	if err := r.writeFeedInfo(zipWriter); err != nil {
		return nil, fmt.Errorf("failed to write feed info: %w", err)
	}

	if err := zipWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to close ZIP writer: %w", err)
	}

	return bytes.NewReader(buf.Bytes()), nil
}

// Helper methods for writing CSV files

func (r *DefaultGtfsRepository) writeAgencies(zipWriter *zip.Writer) error {
	if len(r.agencies) == 0 {
		return nil
	}

	agencies := make([]*model.Agency, 0, len(r.agencies))
	for _, agency := range r.agencies {
		agencies = append(agencies, agency)
	}

	return r.writeCSV(zipWriter, "agency.txt", agencies)
}

func (r *DefaultGtfsRepository) writeStops(zipWriter *zip.Writer) error {
	if len(r.stops) == 0 {
		return nil
	}

	stops := make([]*model.Stop, 0, len(r.stops))
	for _, stop := range r.stops {
		stops = append(stops, stop)
	}

	return r.writeCSV(zipWriter, "stops.txt", stops)
}

func (r *DefaultGtfsRepository) writeRoutes(zipWriter *zip.Writer) error {
	if len(r.routes) == 0 {
		return nil
	}

	routes := make([]*model.GtfsRoute, 0, len(r.routes))
	for _, route := range r.routes {
		routes = append(routes, route)
	}

	return r.writeCSV(zipWriter, "routes.txt", routes)
}

func (r *DefaultGtfsRepository) writeTrips(zipWriter *zip.Writer) error {
	if len(r.trips) == 0 {
		return nil
	}

	trips := make([]*model.Trip, 0, len(r.trips))
	for _, trip := range r.trips {
		trips = append(trips, trip)
	}

	return r.writeCSV(zipWriter, "trips.txt", trips)
}

func (r *DefaultGtfsRepository) writeStopTimes(zipWriter *zip.Writer) error {
	if len(r.stopTimes) == 0 {
		return nil
	}

	return r.writeCSV(zipWriter, "stop_times.txt", r.stopTimes)
}

func (r *DefaultGtfsRepository) writeCalendar(zipWriter *zip.Writer) error {
	if len(r.calendars) == 0 {
		return nil
	}

	calendars := make([]*model.Calendar, 0, len(r.calendars))
	for _, calendar := range r.calendars {
		calendars = append(calendars, calendar)
	}

	return r.writeCSV(zipWriter, "calendar.txt", calendars)
}

func (r *DefaultGtfsRepository) writeCalendarDates(zipWriter *zip.Writer) error {
	if len(r.calendarDates) == 0 {
		return nil
	}

	return r.writeCSV(zipWriter, "calendar_dates.txt", r.calendarDates)
}

func (r *DefaultGtfsRepository) writeTransfers(zipWriter *zip.Writer) error {
	if len(r.transfers) == 0 {
		return nil
	}

	return r.writeCSV(zipWriter, "transfers.txt", r.transfers)
}

func (r *DefaultGtfsRepository) writeShapes(zipWriter *zip.Writer) error {
	if len(r.shapes) == 0 {
		return nil
	}

	return r.writeCSV(zipWriter, "shapes.txt", r.shapes)
}

func (r *DefaultGtfsRepository) writeFeedInfo(zipWriter *zip.Writer) error {
	if r.feedInfo == nil {
		return nil
	}

	return r.writeCSV(zipWriter, "feed_info.txt", []*model.FeedInfo{r.feedInfo})
}

func (r *DefaultGtfsRepository) writeFrequencies(zipWriter *zip.Writer) error {
	if len(r.frequencies) == 0 {
		return nil
	}

	return r.writeCSV(zipWriter, "frequencies.txt", r.frequencies)
}

func (r *DefaultGtfsRepository) writePathways(zipWriter *zip.Writer) error {
	if len(r.pathways) == 0 {
		return nil
	}

	return r.writeCSV(zipWriter, "pathways.txt", r.pathways)
}

func (r *DefaultGtfsRepository) writeLevels(zipWriter *zip.Writer) error {
	if len(r.levels) == 0 {
		return nil
	}

	return r.writeCSV(zipWriter, "levels.txt", r.levels)
}

// writeCSV writes entities to a CSV file in the ZIP archive
func (r *DefaultGtfsRepository) writeCSV(zipWriter *zip.Writer, filename string, entities interface{}) error {
	writer, err := zipWriter.Create(filename)
	if err != nil {
		return err
	}

	csvWriter := csv.NewWriter(writer)
	defer csvWriter.Flush()

	// Use reflection to get entity data
	entitiesValue := reflect.ValueOf(entities)
	if entitiesValue.Kind() != reflect.Slice {
		return fmt.Errorf("entities must be a slice")
	}

	if entitiesValue.Len() == 0 {
		return nil
	}

	// Get the first entity to determine structure
	firstEntity := entitiesValue.Index(0).Elem()
	entityType := firstEntity.Type()

	// Write header
	header := make([]string, entityType.NumField())
	for i := 0; i < entityType.NumField(); i++ {
		field := entityType.Field(i)
		header[i] = r.getCSVFieldName(field.Name)
	}
	if err := csvWriter.Write(header); err != nil {
		return err
	}

	// Write data rows
	for i := 0; i < entitiesValue.Len(); i++ {
		entity := entitiesValue.Index(i).Elem()
		row := make([]string, entity.NumField())

		for j := 0; j < entity.NumField(); j++ {
			field := entity.Field(j)
			row[j] = r.getCSVFieldValue(field)
		}

		if err := csvWriter.Write(row); err != nil {
			return err
		}
	}

	return nil
}

// getCSVFieldName converts Go field name to GTFS CSV field name
func (r *DefaultGtfsRepository) getCSVFieldName(fieldName string) string {
	// Convert CamelCase to snake_case, handling consecutive capitals properly
	var result strings.Builder
	runes := []rune(fieldName)

	for i, char := range runes {
		// Check if we need to add underscore
		if i > 0 && char >= 'A' && char <= 'Z' {
			// Don't add underscore if previous char was also uppercase (e.g., "ID")
			// unless the next char is lowercase (e.g., "URLTest" -> "url_test")
			prevIsUpper := runes[i-1] >= 'A' && runes[i-1] <= 'Z'
			nextIsLower := i+1 < len(runes) && runes[i+1] >= 'a' && runes[i+1] <= 'z'

			if !prevIsUpper || nextIsLower {
				result.WriteByte('_')
			}
		}
		result.WriteRune(char)
	}
	return strings.ToLower(result.String())
}

// getCSVFieldValue converts Go field value to CSV string
func (r *DefaultGtfsRepository) getCSVFieldValue(field reflect.Value) string {
	switch field.Kind() {
	case reflect.String:
		return field.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(field.Int(), 10)
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(field.Float(), 'f', -1, 64)
	case reflect.Bool:
		if field.Bool() {
			return "1"
		}
		return "0"
	default:
		return field.String()
	}
}
