package geometry

import (
	"math"
	"testing"

	"github.com/theoremus-urban-solutions/netex-gtfs-converter/model"
)

// mockNetexRepository implements NetexRepositoryInterface for testing
type mockNetexRepository struct {
	scheduledStopPoints map[string]*model.ScheduledStopPoint
	quays               map[string]*model.Quay
	stopPlaces          map[string]*model.StopPlace
}

func (m *mockNetexRepository) GetScheduledStopPointRefByPointInJourneyPatternRef(pjpRef string) string {
	return pjpRef // Simple mapping for tests
}

func (m *mockNetexRepository) GetScheduledStopPointById(id string) *model.ScheduledStopPoint {
	return m.scheduledStopPoints[id]
}

func (m *mockNetexRepository) GetQuayById(id string) *model.Quay {
	return m.quays[id]
}

func (m *mockNetexRepository) GetStopPlaceByQuayId(quayId string) *model.StopPlace {
	return m.stopPlaces[quayId]
}

func TestNewShapeGenerator(t *testing.T) {
	sg := NewShapeGenerator()
	if sg == nil {
		t.Fatal("NewShapeGenerator() returned nil")
	}

	// Check default values
	if sg.simplificationTolerance != 0.0001 {
		t.Errorf("Expected default simplificationTolerance 0.0001, got %f", sg.simplificationTolerance)
	}
	if sg.maxPointsPerShape != 1000 {
		t.Errorf("Expected default maxPointsPerShape 1000, got %d", sg.maxPointsPerShape)
	}
	if sg.interpolationDistance != 50.0 {
		t.Errorf("Expected default interpolationDistance 50.0, got %f", sg.interpolationDistance)
	}
}

func TestShapeGenerator_Configuration(t *testing.T) {
	sg := NewShapeGenerator()

	// Test SetSimplificationTolerance
	sg.SetSimplificationTolerance(0.001)
	if sg.simplificationTolerance != 0.001 {
		t.Errorf("Expected simplificationTolerance 0.001, got %f", sg.simplificationTolerance)
	}

	// Test SetMaxPointsPerShape
	sg.SetMaxPointsPerShape(500)
	if sg.maxPointsPerShape != 500 {
		t.Errorf("Expected maxPointsPerShape 500, got %d", sg.maxPointsPerShape)
	}

	// Test SetInterpolationDistance
	sg.SetInterpolationDistance(100.0)
	if sg.interpolationDistance != 100.0 {
		t.Errorf("Expected interpolationDistance 100.0, got %f", sg.interpolationDistance)
	}
}

func TestShapeGenerator_GenerateShapeNilInput(t *testing.T) {
	sg := NewShapeGenerator()
	repo := &mockNetexRepository{}

	shapes, err := sg.GenerateShape(nil, repo)
	if err != nil {
		t.Errorf("GenerateShape(nil) should not error, got: %v", err)
	}
	if shapes != nil {
		t.Error("GenerateShape(nil) should return nil shapes")
	}
}

func TestShapeGenerator_GenerateShapeEmptyPattern(t *testing.T) {
	sg := NewShapeGenerator()
	repo := &mockNetexRepository{}

	jp := &model.JourneyPattern{
		ID: "test-pattern",
	}

	shapes, err := sg.GenerateShape(jp, repo)
	if err != nil {
		t.Errorf("GenerateShape() with empty pattern should not error, got: %v", err)
	}
	if shapes != nil {
		t.Error("GenerateShape() with empty pattern should return nil shapes")
	}
}

func TestShapeGenerator_GenerateShapeValidPattern(t *testing.T) {
	sg := NewShapeGenerator()

	// Create mock repository with test data
	repo := &mockNetexRepository{
		scheduledStopPoints: map[string]*model.ScheduledStopPoint{
			"ssp1": {
				ID:      "ssp1",
				QuayRef: "quay1",
			},
			"ssp2": {
				ID:      "ssp2",
				QuayRef: "quay2",
			},
		},
		quays: map[string]*model.Quay{
			"quay1": {
				ID: "quay1",
				Centroid: &model.Centroid{
					Location: &model.Location{
						Latitude:  60.0,
						Longitude: 10.0,
					},
				},
			},
			"quay2": {
				ID: "quay2",
				Centroid: &model.Centroid{
					Location: &model.Location{
						Latitude:  60.1,
						Longitude: 10.1,
					},
				},
			},
		},
	}

	jp := &model.JourneyPattern{
		ID: "test-pattern",
		PointsInSequence: &model.PointsInSequence{
			PointInJourneyPatternOrStopPointInJourneyPatternOrTimingPointInJourneyPattern: []interface{}{
				&model.StopPointInJourneyPattern{
					ID:                    "spjp1",
					ScheduledStopPointRef: "ssp1",
				},
				&model.StopPointInJourneyPattern{
					ID:                    "spjp2",
					ScheduledStopPointRef: "ssp2",
				},
			},
		},
	}

	shapes, err := sg.GenerateShape(jp, repo)
	if err != nil {
		t.Fatalf("GenerateShape() failed: %v", err)
	}

	if len(shapes) < 2 {
		t.Errorf("Expected at least 2 shape points, got %d", len(shapes))
	}

	// Check shape ID format
	if len(shapes) > 0 && shapes[0].ShapeID != "shape_test-pattern" {
		t.Errorf("Expected ShapeID 'shape_test-pattern', got '%s'", shapes[0].ShapeID)
	}

	// Check sequence ordering
	for i := 1; i < len(shapes); i++ {
		if shapes[i].ShapePtSequence <= shapes[i-1].ShapePtSequence {
			t.Errorf("Shape points should be in sequence order")
		}
	}
}

func TestShapeGenerator_HaversineDistance(t *testing.T) {
	sg := NewShapeGenerator()

	// Test distance between known points
	oslo := Point{Lat: 59.9139, Lon: 10.7522}
	bergen := Point{Lat: 60.3913, Lon: 5.3221}

	distance := sg.haversineDistance(oslo, bergen)

	// Expected distance is approximately 308 km
	if distance < 300000 || distance > 320000 {
		t.Errorf("Expected distance between Oslo and Bergen ~308km, got %.0f meters", distance)
	}

	// Test same point
	sameDistance := sg.haversineDistance(oslo, oslo)
	if sameDistance != 0 {
		t.Errorf("Distance between same point should be 0, got %f", sameDistance)
	}

	// Test nearby points
	nearby := Point{Lat: 59.9140, Lon: 10.7523}
	nearDistance := sg.haversineDistance(oslo, nearby)
	if nearDistance > 200 { // Should be less than 200 meters
		t.Errorf("Distance between nearby points should be small, got %f", nearDistance)
	}
}

func TestShapeGenerator_InterpolatePoint(t *testing.T) {
	sg := NewShapeGenerator()

	start := Point{Lat: 60.0, Lon: 10.0}
	end := Point{Lat: 60.1, Lon: 10.1}

	// Test midpoint
	midpoint := sg.interpolatePoint(start, end, 0.5)
	expectedLat := 60.05
	expectedLon := 10.05

	if math.Abs(midpoint.Lat-expectedLat) > 0.001 {
		t.Errorf("Expected midpoint lat %f, got %f", expectedLat, midpoint.Lat)
	}
	if math.Abs(midpoint.Lon-expectedLon) > 0.001 {
		t.Errorf("Expected midpoint lon %f, got %f", expectedLon, midpoint.Lon)
	}

	// Test start point (ratio 0)
	startPoint := sg.interpolatePoint(start, end, 0.0)
	if startPoint.Lat != start.Lat || startPoint.Lon != start.Lon {
		t.Error("Interpolation with ratio 0 should return start point")
	}

	// Test end point (ratio 1)
	endPoint := sg.interpolatePoint(start, end, 1.0)
	if endPoint.Lat != end.Lat || endPoint.Lon != end.Lon {
		t.Error("Interpolation with ratio 1 should return end point")
	}
}

func TestShapeGenerator_PerpendicularDistance(t *testing.T) {
	sg := NewShapeGenerator()

	// Test perpendicular distance
	lineStart := Point{Lat: 60.0, Lon: 10.0}
	lineEnd := Point{Lat: 60.0, Lon: 10.1}
	point := Point{Lat: 60.01, Lon: 10.05} // Point above the line

	distance := sg.perpendicularDistance(point, lineStart, lineEnd)

	// Should be approximately 1111 meters (0.01 degree latitude)
	if distance < 1000 || distance > 1200 {
		t.Errorf("Expected perpendicular distance ~1111m, got %f", distance)
	}

	// Test point on line
	pointOnLine := Point{Lat: 60.0, Lon: 10.05}
	distanceOnLine := sg.perpendicularDistance(pointOnLine, lineStart, lineEnd)

	if distanceOnLine > 10 { // Should be very close to 0
		t.Errorf("Distance to point on line should be near 0, got %f", distanceOnLine)
	}
}

func TestShapeGenerator_DouglasPeucker(t *testing.T) {
	sg := NewShapeGenerator()

	// Test with simple line that should not be simplified
	points := []Point{
		{Lat: 60.0, Lon: 10.0},
		{Lat: 60.1, Lon: 10.1},
	}

	simplified := sg.douglasPeucker(points, 0.001)
	if len(simplified) != 2 {
		t.Errorf("Simple line should remain unchanged, got %d points", len(simplified))
	}

	// Test with zigzag that should be simplified
	zigzag := []Point{
		{Lat: 60.0, Lon: 10.0},
		{Lat: 60.05, Lon: 10.05},
		{Lat: 60.1, Lon: 10.1},
	}

	simplifiedZigzag := sg.douglasPeucker(zigzag, 0.1) // High tolerance
	if len(simplifiedZigzag) >= len(zigzag) {
		t.Error("Zigzag should be simplified with high tolerance")
	}
}

func TestShapeGenerator_SimplifyShape(t *testing.T) {
	sg := NewShapeGenerator()
	sg.SetSimplificationTolerance(0.01) // High tolerance for testing

	// Create shape with many points
	shapes := []*model.Shape{
		{ShapeID: "test", ShapePtLat: 60.0, ShapePtLon: 10.0, ShapePtSequence: 1, ShapeDistTraveled: 0},
		{ShapeID: "test", ShapePtLat: 60.001, ShapePtLon: 10.001, ShapePtSequence: 2, ShapeDistTraveled: 100},
		{ShapeID: "test", ShapePtLat: 60.002, ShapePtLon: 10.002, ShapePtSequence: 3, ShapeDistTraveled: 200},
		{ShapeID: "test", ShapePtLat: 60.1, ShapePtLon: 10.1, ShapePtSequence: 4, ShapeDistTraveled: 1000},
	}

	simplified := sg.simplifyShape(shapes)

	// Should be simplified
	if len(simplified) >= len(shapes) {
		t.Error("Shape should be simplified")
	}

	// Check that sequences are renumbered
	for i, shape := range simplified {
		if shape.ShapePtSequence != i+1 {
			t.Errorf("Shape sequence should be renumbered, expected %d, got %d", i+1, shape.ShapePtSequence)
		}
	}
}

func TestShapeGenerator_ExtractStopLocations(t *testing.T) {
	sg := NewShapeGenerator()

	repo := &mockNetexRepository{
		scheduledStopPoints: map[string]*model.ScheduledStopPoint{
			"ssp1": {
				ID:      "ssp1",
				QuayRef: "quay1",
			},
		},
		quays: map[string]*model.Quay{
			"quay1": {
				ID: "quay1",
				Centroid: &model.Centroid{
					Location: &model.Location{
						Latitude:  60.0,
						Longitude: 10.0,
					},
				},
			},
		},
	}

	jp := &model.JourneyPattern{
		ID: "test",
		PointsInSequence: &model.PointsInSequence{
			PointInJourneyPatternOrStopPointInJourneyPatternOrTimingPointInJourneyPattern: []interface{}{
				&model.StopPointInJourneyPattern{
					ID:                    "spjp1",
					ScheduledStopPointRef: "ssp1",
				},
			},
		},
	}

	locations, err := sg.extractStopLocations(jp, repo)
	if err != nil {
		t.Fatalf("extractStopLocations() failed: %v", err)
	}

	if len(locations) != 1 {
		t.Errorf("Expected 1 location, got %d", len(locations))
	}

	if len(locations) > 0 {
		location := locations[0]
		if location.ID != "spjp1" {
			t.Errorf("Expected location ID 'spjp1', got '%s'", location.ID)
		}
		if location.Sequence != 1 {
			t.Errorf("Expected sequence 1, got %d", location.Sequence)
		}
		if location.Location.Lat != 60.0 {
			t.Errorf("Expected lat 60.0, got %f", location.Location.Lat)
		}
	}
}

func TestShapeGenerator_ResolveStopLocationFromStopPlace(t *testing.T) {
	sg := NewShapeGenerator()

	repo := &mockNetexRepository{
		scheduledStopPoints: map[string]*model.ScheduledStopPoint{
			"ssp1": {
				ID:           "ssp1",
				StopPlaceRef: "sp1",
			},
		},
		stopPlaces: map[string]*model.StopPlace{
			"sp1": {
				ID: "sp1",
				Centroid: &model.Centroid{
					Location: &model.Location{
						Latitude:  59.0,
						Longitude: 9.0,
					},
				},
			},
		},
	}

	stopPoint := &model.StopPointInJourneyPattern{
		ScheduledStopPointRef: "ssp1",
	}

	location, err := sg.resolveStopLocation(stopPoint, repo)
	if err != nil {
		t.Fatalf("resolveStopLocation() failed: %v", err)
	}

	if location.Lat != 59.0 || location.Lon != 9.0 {
		t.Errorf("Expected location (59.0, 9.0), got (%f, %f)", location.Lat, location.Lon)
	}
}

func TestShapeGenerator_ResolveStopLocationErrors(t *testing.T) {
	sg := NewShapeGenerator()
	repo := &mockNetexRepository{
		scheduledStopPoints: map[string]*model.ScheduledStopPoint{},
		quays:               map[string]*model.Quay{},
		stopPlaces:          map[string]*model.StopPlace{},
	}

	tests := []struct {
		name      string
		stopPoint *model.StopPointInJourneyPattern
		expectErr bool
	}{
		{
			name: "no scheduled stop point reference",
			stopPoint: &model.StopPointInJourneyPattern{
				ScheduledStopPointRef: "",
			},
			expectErr: true,
		},
		{
			name: "scheduled stop point not found",
			stopPoint: &model.StopPointInJourneyPattern{
				ScheduledStopPointRef: "nonexistent",
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := sg.resolveStopLocation(tt.stopPoint, repo)
			if tt.expectErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestShapeGenerator_CalculateDistances(t *testing.T) {
	sg := NewShapeGenerator()

	locations := []StopLocation{
		{Location: Point{Lat: 60.0, Lon: 10.0}},
		{Location: Point{Lat: 60.0, Lon: 10.1}}, // ~8km east
		{Location: Point{Lat: 60.1, Lon: 10.1}}, // ~11km north
	}

	sg.calculateDistances(locations)

	// First location should have distance 0
	if locations[0].Distance != 0.0 {
		t.Errorf("First location should have distance 0, got %f", locations[0].Distance)
	}

	// Distances should be increasing
	if locations[1].Distance <= locations[0].Distance {
		t.Error("Distances should increase")
	}
	if locations[2].Distance <= locations[1].Distance {
		t.Error("Distances should increase")
	}

	// Check rough distance values (0.1 degree longitude at 60°N ≈ 5.6km)
	if locations[1].Distance < 5000 || locations[1].Distance > 7000 {
		t.Errorf("Expected distance ~6km, got %f", locations[1].Distance)
	}
}

func TestShapeGenerator_FindClosestOriginalPoint(t *testing.T) {
	sg := NewShapeGenerator()

	target := Point{Lat: 60.05, Lon: 10.05}
	original := []*model.Shape{
		{ShapePtLat: 60.0, ShapePtLon: 10.0},
		{ShapePtLat: 60.1, ShapePtLon: 10.1},
		{ShapePtLat: 60.04, ShapePtLon: 10.04}, // Closest
	}

	index := sg.findClosestOriginalPoint(target, original)
	if index != 2 {
		t.Errorf("Expected closest point index 2, got %d", index)
	}
}

func TestShapeGenerator_CalculateShapeDistanceForStopTime(t *testing.T) {
	sg := NewShapeGenerator()

	shapes := []*model.Shape{
		{ShapeDistTraveled: 0.0},
		{ShapeDistTraveled: 100.0},
	}

	distance := sg.CalculateShapeDistanceForStopTime("stop1", shapes)

	// Current implementation returns 0.0
	if distance != 0.0 {
		t.Errorf("Expected 0.0 (placeholder implementation), got %f", distance)
	}
}
