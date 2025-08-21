package exporter

import (
	"archive/zip"
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/theoremus-urban-solutions/netex-gtfs-converter/model"
	"github.com/theoremus-urban-solutions/netex-gtfs-converter/producer"
)

// Mock stop area repository for testing
type mockStopAreaRepository struct{}

func (m *mockStopAreaRepository) GetQuayById(quayId string) *model.Quay               { return nil }
func (m *mockStopAreaRepository) GetStopPlaceByQuayId(quayId string) *model.StopPlace { return nil }
func (m *mockStopAreaRepository) GetAllQuays() []*model.Quay                          { return nil }
func (m *mockStopAreaRepository) LoadStopAreas(data []byte) error                     { return nil }

// TestGTFSCSVOutputValidation tests actual CSV file generation and content
func TestGTFSCSVOutputValidation(t *testing.T) {
	// Create a simple test dataset
	stopAreaRepo := &mockStopAreaRepository{}

	// Create exporter and export GTFS
	exporter := NewDefaultGtfsExporter("TEST", stopAreaRepo)

	// Add test entities to the exporter's repositories
	netexRepoFromExporter := exporter.GetNetexRepository()
	gtfsRepoFromExporter := exporter.GetGtfsRepository()

	setupTestData(netexRepoFromExporter, gtfsRepoFromExporter)

	output, err := gtfsRepoFromExporter.WriteGtfs()
	if err != nil {
		t.Fatalf("WriteGtfs() failed: %v", err)
	}

	// Read the output as a ZIP file
	outputBytes, err := io.ReadAll(output)
	if err != nil {
		t.Fatalf("Failed to read output: %v", err)
	}

	// Parse ZIP file
	zipReader, err := zip.NewReader(bytes.NewReader(outputBytes), int64(len(outputBytes)))
	if err != nil {
		t.Fatalf("Failed to parse ZIP: %v", err)
	}

	// Test each GTFS file
	testGTFSFileStructure(t, zipReader)
}

func testGTFSFileStructure(t *testing.T, zipReader *zip.Reader) {
	requiredFiles := []string{
		"agency.txt",
		"stops.txt",
		"routes.txt",
		"trips.txt",
		"stop_times.txt",
		"calendar.txt",
	}

	optionalFiles := []string{
		"calendar_dates.txt",
		"transfers.txt",
		"shapes.txt",
		"feed_info.txt",
	}

	fileMap := make(map[string]*zip.File)
	for _, file := range zipReader.File {
		fileMap[file.Name] = file
	}

	// Check required files exist
	for _, fileName := range requiredFiles {
		if file, exists := fileMap[fileName]; exists {
			t.Run("Test "+fileName, func(t *testing.T) {
				testCSVFile(t, file, fileName)
			})
		} else {
			t.Errorf("Required GTFS file missing: %s", fileName)
		}
	}

	// Test optional files if they exist
	for _, fileName := range optionalFiles {
		if file, exists := fileMap[fileName]; exists {
			t.Run("Test "+fileName, func(t *testing.T) {
				testCSVFile(t, file, fileName)
			})
		}
	}
}

func testCSVFile(t *testing.T, zipFile *zip.File, fileName string) {
	reader, err := zipFile.Open()
	if err != nil {
		t.Fatalf("Failed to open %s: %v", fileName, err)
	}
	defer func() {
		if err := reader.Close(); err != nil {
			t.Logf("Failed to close reader: %v", err)
		}
	}()

	content, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("Failed to read %s: %v", fileName, err)
	}

	csvReader := csv.NewReader(strings.NewReader(string(content)))
	records, err := csvReader.ReadAll()
	if err != nil {
		t.Fatalf("Failed to parse CSV %s: %v", fileName, err)
	}

	if len(records) == 0 {
		t.Errorf("%s is empty", fileName)
		return
	}

	headers := records[0]
	dataRows := records[1:]

	// Test file-specific structure and content
	switch fileName {
	case "agency.txt":
		testAgencyFile(t, headers, dataRows)
	case "stops.txt":
		testStopsFile(t, headers, dataRows)
	case "routes.txt":
		testRoutesFile(t, headers, dataRows)
	case "trips.txt":
		testTripsFile(t, headers, dataRows)
	case "stop_times.txt":
		testStopTimesFile(t, headers, dataRows)
	case "calendar.txt":
		testCalendarFile(t, headers, dataRows)
	case "calendar_dates.txt":
		testCalendarDatesFile(t, headers, dataRows)
	case "feed_info.txt":
		testFeedInfoFile(t, headers, dataRows)
	case "transfers.txt":
		testTransfersFile(t, headers, dataRows)
	}
}

func testAgencyFile(t *testing.T, headers []string, dataRows [][]string) {
	requiredFields := []string{"agency_id", "agency_name", "agency_url", "agency_timezone"}
	_ = []string{"agency_lang", "agency_phone", "agency_email"} // optionalFields - suppress unused warning

	testRequiredFields(t, "agency.txt", headers, requiredFields)

	if len(dataRows) == 0 {
		t.Error("agency.txt has no data rows")
		return
	}

	// Test first agency entry
	agency := dataRows[0]
	if len(agency) != len(headers) {
		t.Errorf("agency.txt: data row length %d doesn't match header length %d", len(agency), len(headers))
		return
	}

	// Create field map for easier access
	fieldMap := createFieldMap(headers, agency)

	// Test required fields are not empty
	for _, field := range requiredFields {
		if value, exists := fieldMap[field]; !exists || value == "" {
			t.Errorf("agency.txt: required field %s is missing or empty", field)
		}
	}

	// Test agency_id format (should not contain spaces)
	if agencyID, exists := fieldMap["agency_id"]; exists && strings.Contains(agencyID, " ") {
		t.Errorf("agency.txt: agency_id should not contain spaces: %s", agencyID)
	}

	// Test URL format
	if url, exists := fieldMap["agency_url"]; exists && !strings.HasPrefix(url, "http") {
		t.Errorf("agency.txt: agency_url should be a valid URL: %s", url)
	}
}

func testStopsFile(t *testing.T, headers []string, dataRows [][]string) {
	requiredFields := []string{"stop_id", "stop_name", "stop_lat", "stop_lon"}

	testRequiredFields(t, "stops.txt", headers, requiredFields)

	if len(dataRows) == 0 {
		t.Error("stops.txt has no data rows")
		return
	}

	// Test first stop entry
	stop := dataRows[0]
	fieldMap := createFieldMap(headers, stop)

	// Test required fields
	for _, field := range requiredFields {
		if value, exists := fieldMap[field]; !exists || value == "" {
			t.Errorf("stops.txt: required field %s is missing or empty", field)
		}
	}

	// Test coordinate validity (basic check)
	if lat, exists := fieldMap["stop_lat"]; exists {
		if !isValidCoordinate(lat, -90, 90) {
			t.Errorf("stops.txt: invalid latitude: %s", lat)
		}
	}

	if lon, exists := fieldMap["stop_lon"]; exists {
		if !isValidCoordinate(lon, -180, 180) {
			t.Errorf("stops.txt: invalid longitude: %s", lon)
		}
	}
}

func testRoutesFile(t *testing.T, headers []string, dataRows [][]string) {
	requiredFields := []string{"route_id", "route_type"}

	testRequiredFields(t, "routes.txt", headers, requiredFields)

	if len(dataRows) == 0 {
		t.Error("routes.txt has no data rows")
		return
	}

	route := dataRows[0]
	fieldMap := createFieldMap(headers, route)

	// Test route_type is valid GTFS route type
	if routeType, exists := fieldMap["route_type"]; exists {
		if !isValidGTFSRouteType(routeType) {
			t.Errorf("routes.txt: invalid route_type: %s", routeType)
		}
	}

	// Test that route has either short name or long name
	shortName, hasShort := fieldMap["route_short_name"]
	longName, hasLong := fieldMap["route_long_name"]

	if (!hasShort || shortName == "") && (!hasLong || longName == "") {
		t.Error("routes.txt: route must have either route_short_name or route_long_name")
	}
}

func testTripsFile(t *testing.T, headers []string, dataRows [][]string) {
	requiredFields := []string{"route_id", "service_id", "trip_id"}

	testRequiredFields(t, "trips.txt", headers, requiredFields)

	if len(dataRows) == 0 {
		t.Error("trips.txt has no data rows")
		return
	}

	trip := dataRows[0]
	fieldMap := createFieldMap(headers, trip)

	// Test required fields are not empty
	for _, field := range requiredFields {
		if value, exists := fieldMap[field]; !exists || value == "" {
			t.Errorf("trips.txt: required field %s is missing or empty", field)
		}
	}

	// Test direction_id if present
	if directionID, exists := fieldMap["direction_id"]; exists && directionID != "" {
		if directionID != "0" && directionID != "1" {
			t.Errorf("trips.txt: direction_id must be 0 or 1: %s", directionID)
		}
	}
}

func testStopTimesFile(t *testing.T, headers []string, dataRows [][]string) {
	requiredFields := []string{"trip_id", "arrival_time", "departure_time", "stop_id", "stop_sequence"}

	testRequiredFields(t, "stop_times.txt", headers, requiredFields)

	if len(dataRows) == 0 {
		t.Error("stop_times.txt has no data rows")
		return
	}

	stopTime := dataRows[0]
	fieldMap := createFieldMap(headers, stopTime)

	// Test time format
	if arrivalTime, exists := fieldMap["arrival_time"]; exists {
		if !isValidGTFSTime(arrivalTime) {
			t.Errorf("stop_times.txt: invalid arrival_time format: %s", arrivalTime)
		}
	}

	if departureTime, exists := fieldMap["departure_time"]; exists {
		if !isValidGTFSTime(departureTime) {
			t.Errorf("stop_times.txt: invalid departure_time format: %s", departureTime)
		}
	}

	// Test stop_sequence is numeric
	if stopSeq, exists := fieldMap["stop_sequence"]; exists {
		if !isNumeric(stopSeq) {
			t.Errorf("stop_times.txt: stop_sequence must be numeric: %s", stopSeq)
		}
	}
}

func testCalendarFile(t *testing.T, headers []string, dataRows [][]string) {
	requiredFields := []string{"service_id", "monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday", "start_date", "end_date"}

	testRequiredFields(t, "calendar.txt", headers, requiredFields)

	if len(dataRows) == 0 {
		// Calendar.txt can be empty if calendar_dates.txt is used instead
		return
	}

	calendar := dataRows[0]
	fieldMap := createFieldMap(headers, calendar)

	// Test date format
	if startDate, exists := fieldMap["start_date"]; exists {
		if !isValidGTFSDate(startDate) {
			t.Errorf("calendar.txt: invalid start_date format: %s", startDate)
		}
	}

	if endDate, exists := fieldMap["end_date"]; exists {
		if !isValidGTFSDate(endDate) {
			t.Errorf("calendar.txt: invalid end_date format: %s", endDate)
		}
	}

	// Test day fields are 0 or 1
	dayFields := []string{"monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday"}
	for _, day := range dayFields {
		if value, exists := fieldMap[day]; exists {
			if value != "0" && value != "1" && value != "true" && value != "false" {
				t.Errorf("calendar.txt: %s must be 0, 1, true, or false: %s", day, value)
			}
		}
	}
}

func testCalendarDatesFile(t *testing.T, headers []string, dataRows [][]string) {
	requiredFields := []string{"service_id", "date", "exception_type"}

	testRequiredFields(t, "calendar_dates.txt", headers, requiredFields)

	if len(dataRows) == 0 {
		return // Optional file, can be empty
	}

	calendarDate := dataRows[0]
	fieldMap := createFieldMap(headers, calendarDate)

	// Test date format
	if date, exists := fieldMap["date"]; exists {
		if !isValidGTFSDate(date) {
			t.Errorf("calendar_dates.txt: invalid date format: %s", date)
		}
	}

	// Test exception_type is 1 or 2
	if exceptionType, exists := fieldMap["exception_type"]; exists {
		if exceptionType != "1" && exceptionType != "2" {
			t.Errorf("calendar_dates.txt: exception_type must be 1 or 2: %s", exceptionType)
		}
	}
}

func testFeedInfoFile(t *testing.T, headers []string, dataRows [][]string) {
	requiredFields := []string{"feed_publisher_name", "feed_publisher_url", "feed_lang"}

	testRequiredFields(t, "feed_info.txt", headers, requiredFields)

	if len(dataRows) == 0 {
		t.Error("feed_info.txt has no data rows")
		return
	}
}

func testTransfersFile(t *testing.T, headers []string, dataRows [][]string) {
	requiredFields := []string{"from_stop_id", "to_stop_id", "transfer_type"}

	testRequiredFields(t, "transfers.txt", headers, requiredFields)

	if len(dataRows) == 0 {
		return // Optional file, can be empty
	}

	transfer := dataRows[0]
	fieldMap := createFieldMap(headers, transfer)

	// Test transfer_type is valid (0-3)
	if transferType, exists := fieldMap["transfer_type"]; exists {
		if transferType != "0" && transferType != "1" && transferType != "2" && transferType != "3" {
			t.Errorf("transfers.txt: transfer_type must be 0, 1, 2, or 3: %s", transferType)
		}
	}
}

// Helper functions

func setupTestData(netexRepo producer.NetexRepository, gtfsRepo producer.GtfsRepository) {
	// Add a test authority
	authority := &model.Authority{
		ID:   "test-authority",
		Name: "Test Authority",
		URL:  "https://example.com",
	}
	if err := netexRepo.SaveEntity(authority); err != nil {
		panic(fmt.Sprintf("Failed to save authority: %v", err))
	}

	// Add a test agency
	agency := &model.Agency{
		AgencyID:       "test-authority",
		AgencyName:     "Test Authority",
		AgencyURL:      "https://example.com",
		AgencyTimezone: "UTC",
	}
	if err := gtfsRepo.SaveEntity(agency); err != nil {
		panic(fmt.Sprintf("Failed to save agency: %v", err))
	}

	// Add a test line/route
	line := &model.Line{
		ID:            "test-line",
		Name:          "Test Line",
		PublicCode:    "1",
		TransportMode: "bus",
		AuthorityRef:  "test-authority",
	}
	if err := netexRepo.SaveEntity(line); err != nil {
		panic(fmt.Sprintf("Failed to save line: %v", err))
	}

	route := &model.GtfsRoute{
		RouteID:        "test-line",
		AgencyID:       "test-authority",
		RouteShortName: "1",
		RouteLongName:  "Test Line",
		RouteType:      3, // Bus
	}
	if err := gtfsRepo.SaveEntity(route); err != nil {
		panic(fmt.Sprintf("Failed to save route: %v", err))
	}

	// Add a test stop
	stop := &model.Stop{
		StopID:   "test-stop",
		StopName: "Test Stop",
		StopLat:  59.9139,
		StopLon:  10.7522,
	}
	if err := gtfsRepo.SaveEntity(stop); err != nil {
		panic(fmt.Sprintf("Failed to save stop: %v", err))
	}

	// Add a test trip
	trip := &model.Trip{
		RouteID:   "test-line",
		ServiceID: "test-service",
		TripID:    "test-trip",
	}
	if err := gtfsRepo.SaveEntity(trip); err != nil {
		panic(fmt.Sprintf("Failed to save trip: %v", err))
	}

	// Add a test calendar
	calendar := &model.Calendar{
		ServiceID: "test-service",
		Monday:    true,
		Tuesday:   true,
		Wednesday: true,
		Thursday:  true,
		Friday:    true,
		Saturday:  false,
		Sunday:    false,
		StartDate: "20240101",
		EndDate:   "20241231",
	}
	if err := gtfsRepo.SaveEntity(calendar); err != nil {
		panic(fmt.Sprintf("Failed to save calendar: %v", err))
	}

	// Add a test stop time
	stopTime := &model.StopTime{
		TripID:        "test-trip",
		ArrivalTime:   "08:00:00",
		DepartureTime: "08:00:00",
		StopID:        "test-stop",
		StopSequence:  1,
	}
	if err := gtfsRepo.SaveEntity(stopTime); err != nil {
		panic(fmt.Sprintf("Failed to save stop time: %v", err))
	}
}

func testRequiredFields(t *testing.T, fileName string, headers []string, requiredFields []string) {
	headerSet := make(map[string]bool)
	for _, header := range headers {
		headerSet[header] = true
	}

	for _, field := range requiredFields {
		if !headerSet[field] {
			t.Errorf("%s: missing required field %s", fileName, field)
		}
	}
}

func createFieldMap(headers []string, row []string) map[string]string {
	fieldMap := make(map[string]string)
	for i, header := range headers {
		if i < len(row) {
			fieldMap[header] = row[i]
		}
	}
	return fieldMap
}

func isValidCoordinate(value string, min, max float64) bool {
	var coord float64
	if _, err := fmt.Sscanf(value, "%f", &coord); err != nil {
		return false
	}
	return coord >= min && coord <= max
}

func isValidGTFSRouteType(routeType string) bool {
	validTypes := []string{"0", "1", "2", "3", "4", "5", "6", "7"} // Basic GTFS types
	for _, valid := range validTypes {
		if routeType == valid {
			return true
		}
	}
	// Also accept extended types (100+)
	var typeNum int
	if _, err := fmt.Sscanf(routeType, "%d", &typeNum); err == nil {
		return typeNum >= 0 && typeNum <= 1799 // Valid GTFS range
	}
	return false
}

func isValidGTFSTime(timeStr string) bool {
	// GTFS time format: HH:MM:SS (can exceed 24 hours)
	var hours, minutes, seconds int
	if _, err := fmt.Sscanf(timeStr, "%d:%d:%d", &hours, &minutes, &seconds); err != nil {
		return false
	}
	return hours >= 0 && minutes >= 0 && minutes < 60 && seconds >= 0 && seconds < 60
}

func isValidGTFSDate(dateStr string) bool {
	// GTFS date format: YYYYMMDD
	if len(dateStr) != 8 {
		return false
	}
	var year, month, day int
	if _, err := fmt.Sscanf(dateStr, "%4d%2d%2d", &year, &month, &day); err != nil {
		return false
	}
	return year >= 1900 && month >= 1 && month <= 12 && day >= 1 && day <= 31
}

func isNumeric(value string) bool {
	var num int
	_, err := fmt.Sscanf(value, "%d", &num)
	return err == nil
}
