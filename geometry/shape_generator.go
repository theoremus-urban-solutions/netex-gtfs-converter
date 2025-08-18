package geometry

import (
	"fmt"
	"math"

	"github.com/theoremus-urban-solutions/netex-gtfs-converter/model"
)

// Point represents a geographic coordinate
type Point struct {
	Lat float64
	Lon float64
}

// ShapeGenerator generates GTFS shapes from NeTEx journey patterns and stop locations
type ShapeGenerator struct {
	simplificationTolerance float64
	maxPointsPerShape       int
	interpolationDistance   float64 // meters between interpolated points
}

// NewShapeGenerator creates a new shape generator with default settings
func NewShapeGenerator() *ShapeGenerator {
	return &ShapeGenerator{
		simplificationTolerance: 0.0001, // ~10 meters
		maxPointsPerShape:       1000,   // Maximum points per shape
		interpolationDistance:   50.0,   // 50 meters between points
	}
}

// SetSimplificationTolerance sets the tolerance for shape simplification
func (sg *ShapeGenerator) SetSimplificationTolerance(tolerance float64) {
	sg.simplificationTolerance = tolerance
}

// SetMaxPointsPerShape sets the maximum number of points in a shape
func (sg *ShapeGenerator) SetMaxPointsPerShape(maxPoints int) {
	sg.maxPointsPerShape = maxPoints
}

// SetInterpolationDistance sets the distance between interpolated points
func (sg *ShapeGenerator) SetInterpolationDistance(distance float64) {
	sg.interpolationDistance = distance
}

// GenerateShape creates GTFS shape points from a journey pattern
func (sg *ShapeGenerator) GenerateShape(journeyPattern *model.JourneyPattern, netexRepo NetexRepositoryInterface) ([]*model.Shape, error) {
	if journeyPattern == nil {
		return nil, nil
	}

	// Extract stop locations from journey pattern
	stopLocations, err := sg.extractStopLocations(journeyPattern, netexRepo)
	if err != nil {
		return nil, fmt.Errorf("failed to extract stop locations: %w", err)
	}

	if len(stopLocations) < 2 {
		return nil, nil // Need at least 2 points for a shape
	}

	// Generate shape points
	shapePoints := sg.generateShapePoints(journeyPattern.ID, stopLocations)

	// Simplify if too many points
	if len(shapePoints) > sg.maxPointsPerShape {
		shapePoints = sg.simplifyShape(shapePoints)
	}

	return shapePoints, nil
}

// NetexRepositoryInterface defines the interface needed for shape generation
type NetexRepositoryInterface interface {
	GetScheduledStopPointRefByPointInJourneyPatternRef(pjpRef string) string
	GetScheduledStopPointById(id string) *model.ScheduledStopPoint
	GetQuayById(id string) *model.Quay
	GetStopPlaceByQuayId(quayId string) *model.StopPlace
}

// StopLocation represents a stop with its geographic location
type StopLocation struct {
	ID       string
	Sequence int
	Location Point
	Distance float64 // Distance from route start
}

// extractStopLocations extracts stop locations from journey pattern
func (sg *ShapeGenerator) extractStopLocations(journeyPattern *model.JourneyPattern, repo NetexRepositoryInterface) ([]StopLocation, error) {
	var locations []StopLocation

	if journeyPattern.PointsInSequence == nil {
		return locations, nil
	}

	// Process points in sequence
	for i, pointInterface := range journeyPattern.PointsInSequence.PointInJourneyPatternOrStopPointInJourneyPatternOrTimingPointInJourneyPattern {
		if stopPoint, ok := pointInterface.(*model.StopPointInJourneyPattern); ok {
			location, err := sg.resolveStopLocation(stopPoint, repo)
			if err != nil {
				continue // Skip stops we can't resolve
			}

			locations = append(locations, StopLocation{
				ID:       stopPoint.ID,
				Sequence: i + 1,
				Location: location,
			})
		}
	}

	// Calculate distances along route
	sg.calculateDistances(locations)

	return locations, nil
}

// resolveStopLocation resolves the geographic location of a stop
func (sg *ShapeGenerator) resolveStopLocation(stopPoint *model.StopPointInJourneyPattern, repo NetexRepositoryInterface) (Point, error) {
	// Get scheduled stop point reference
	sspRef := stopPoint.ScheduledStopPointRef
	if sspRef == "" {
		return Point{}, fmt.Errorf("no scheduled stop point reference")
	}

	// Look up scheduled stop point
	ssp := repo.GetScheduledStopPointById(sspRef)
	if ssp == nil {
		return Point{}, fmt.Errorf("scheduled stop point not found: %s", sspRef)
	}

	// Try to get location from quay first
	if ssp.QuayRef != "" {
		if quay := repo.GetQuayById(ssp.QuayRef); quay != nil {
			if quay.Centroid != nil && quay.Centroid.Location != nil {
				return Point{
					Lat: quay.Centroid.Location.Latitude,
					Lon: quay.Centroid.Location.Longitude,
				}, nil
			}
		}
	}

	// Fall back to stop place location
	if ssp.StopPlaceRef != "" {
		if stopPlace := repo.GetStopPlaceByQuayId(ssp.StopPlaceRef); stopPlace != nil {
			if stopPlace.Centroid != nil && stopPlace.Centroid.Location != nil {
				return Point{
					Lat: stopPlace.Centroid.Location.Latitude,
					Lon: stopPlace.Centroid.Location.Longitude,
				}, nil
			}
		}
	}

	return Point{}, fmt.Errorf("no location found for stop point")
}

// calculateDistances calculates cumulative distances along the route
func (sg *ShapeGenerator) calculateDistances(locations []StopLocation) {
	if len(locations) == 0 {
		return
	}

	locations[0].Distance = 0.0

	for i := 1; i < len(locations); i++ {
		distance := sg.haversineDistance(
			locations[i-1].Location,
			locations[i].Location,
		)
		locations[i].Distance = locations[i-1].Distance + distance
	}
}

// generateShapePoints creates GTFS shape points with interpolation
func (sg *ShapeGenerator) generateShapePoints(shapeID string, stopLocations []StopLocation) []*model.Shape {
	var shapePoints []*model.Shape
	sequence := 1

	for i := 0; i < len(stopLocations)-1; i++ {
		start := stopLocations[i]
		end := stopLocations[i+1]

		// Add start point
		shapePoints = append(shapePoints, &model.Shape{
			ShapeID:           "shape_" + shapeID,
			ShapePtLat:        start.Location.Lat,
			ShapePtLon:        start.Location.Lon,
			ShapePtSequence:   sequence,
			ShapeDistTraveled: start.Distance,
		})
		sequence++

		// Interpolate points between stops if distance is large
		segmentDistance := end.Distance - start.Distance
		if segmentDistance > sg.interpolationDistance {
			numInterpolations := int(segmentDistance / sg.interpolationDistance)

			for j := 1; j <= numInterpolations; j++ {
				ratio := float64(j) / float64(numInterpolations+1)
				interpolatedPoint := sg.interpolatePoint(start.Location, end.Location, ratio)
				interpolatedDistance := start.Distance + (segmentDistance * ratio)

				shapePoints = append(shapePoints, &model.Shape{
					ShapeID:           "shape_" + shapeID,
					ShapePtLat:        interpolatedPoint.Lat,
					ShapePtLon:        interpolatedPoint.Lon,
					ShapePtSequence:   sequence,
					ShapeDistTraveled: interpolatedDistance,
				})
				sequence++
			}
		}
	}

	// Add final point
	if len(stopLocations) > 0 {
		last := stopLocations[len(stopLocations)-1]
		shapePoints = append(shapePoints, &model.Shape{
			ShapeID:           "shape_" + shapeID,
			ShapePtLat:        last.Location.Lat,
			ShapePtLon:        last.Location.Lon,
			ShapePtSequence:   sequence,
			ShapeDistTraveled: last.Distance,
		})
	}

	return shapePoints
}

// interpolatePoint interpolates between two geographic points
func (sg *ShapeGenerator) interpolatePoint(start, end Point, ratio float64) Point {
	return Point{
		Lat: start.Lat + (end.Lat-start.Lat)*ratio,
		Lon: start.Lon + (end.Lon-start.Lon)*ratio,
	}
}

// simplifyShape reduces the number of points using Douglas-Peucker algorithm
func (sg *ShapeGenerator) simplifyShape(points []*model.Shape) []*model.Shape {
	if len(points) <= 2 {
		return points
	}

	// Convert to Point slice for algorithm
	coords := make([]Point, len(points))
	for i, p := range points {
		coords[i] = Point{Lat: p.ShapePtLat, Lon: p.ShapePtLon}
	}

	// Apply Douglas-Peucker simplification
	simplified := sg.douglasPeucker(coords, sg.simplificationTolerance)

	// Convert back to Shape points
	result := make([]*model.Shape, len(simplified))
	for i, coord := range simplified {
		// Find original point to preserve distance information
		originalIndex := sg.findClosestOriginalPoint(coord, points)
		original := points[originalIndex]

		result[i] = &model.Shape{
			ShapeID:           original.ShapeID,
			ShapePtLat:        coord.Lat,
			ShapePtLon:        coord.Lon,
			ShapePtSequence:   i + 1,
			ShapeDistTraveled: original.ShapeDistTraveled,
		}
	}

	return result
}

// douglasPeucker implements the Douglas-Peucker line simplification algorithm
func (sg *ShapeGenerator) douglasPeucker(points []Point, tolerance float64) []Point {
	if len(points) <= 2 {
		return points
	}

	// Find the point with maximum distance from the line segment
	maxDist := 0.0
	maxIndex := 0

	for i := 1; i < len(points)-1; i++ {
		dist := sg.perpendicularDistance(points[i], points[0], points[len(points)-1])
		if dist > maxDist {
			maxDist = dist
			maxIndex = i
		}
	}

	// If max distance is greater than tolerance, recursively simplify
	if maxDist > tolerance {
		// Recursively process both parts
		left := sg.douglasPeucker(points[:maxIndex+1], tolerance)
		right := sg.douglasPeucker(points[maxIndex:], tolerance)

		// Combine results (removing duplicate middle point)
		result := make([]Point, len(left)+len(right)-1)
		copy(result, left)
		copy(result[len(left):], right[1:])
		return result
	}

	// Return simplified line with just start and end points
	return []Point{points[0], points[len(points)-1]}
}

// perpendicularDistance calculates perpendicular distance from point to line segment
func (sg *ShapeGenerator) perpendicularDistance(point, lineStart, lineEnd Point) float64 {
	// Convert to meters using approximate conversion
	const latToMeters = 111319.9 // meters per degree latitude
	lonToMeters := latToMeters * math.Cos(point.Lat*math.Pi/180)

	// Convert points to metric coordinates
	px := point.Lon * lonToMeters
	py := point.Lat * latToMeters
	lsx := lineStart.Lon * lonToMeters
	lsy := lineStart.Lat * latToMeters
	lex := lineEnd.Lon * lonToMeters
	ley := lineEnd.Lat * latToMeters

	// Calculate perpendicular distance
	A := px - lsx
	B := py - lsy
	C := lex - lsx
	D := ley - lsy

	dot := A*C + B*D
	lenSq := C*C + D*D

	if lenSq == 0 {
		// Line start and end are the same point
		return math.Sqrt(A*A + B*B)
	}

	param := dot / lenSq

	var xx, yy float64

	if param < 0 {
		xx = lsx
		yy = lsy
	} else if param > 1 {
		xx = lex
		yy = ley
	} else {
		xx = lsx + param*C
		yy = lsy + param*D
	}

	dx := px - xx
	dy := py - yy
	return math.Sqrt(dx*dx + dy*dy)
}

// findClosestOriginalPoint finds the original point closest to the simplified point
func (sg *ShapeGenerator) findClosestOriginalPoint(target Point, original []*model.Shape) int {
	minDist := math.MaxFloat64
	minIndex := 0

	for i, point := range original {
		dist := sg.haversineDistance(target, Point{Lat: point.ShapePtLat, Lon: point.ShapePtLon})
		if dist < minDist {
			minDist = dist
			minIndex = i
		}
	}

	return minIndex
}

// haversineDistance calculates the great-circle distance between two points in meters
func (sg *ShapeGenerator) haversineDistance(point1, point2 Point) float64 {
	const earthRadius = 6371000 // Earth radius in meters

	lat1Rad := point1.Lat * math.Pi / 180
	lon1Rad := point1.Lon * math.Pi / 180
	lat2Rad := point2.Lat * math.Pi / 180
	lon2Rad := point2.Lon * math.Pi / 180

	dLat := lat2Rad - lat1Rad
	dLon := lon2Rad - lon1Rad

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(dLon/2)*math.Sin(dLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadius * c
}

// CalculateShapeDistanceForStopTime calculates shape distance for a stop time
func (sg *ShapeGenerator) CalculateShapeDistanceForStopTime(stopID string, shapes []*model.Shape) float64 {
	// This would require matching the stop location to the nearest shape point
	// For now, return 0 - in a full implementation, this would:
	// 1. Get the stop's geographic location
	// 2. Find the nearest point on the shape
	// 3. Return the shape_dist_traveled at that point
	return 0.0
}
