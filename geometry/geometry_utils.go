package geometry

import (
	"math"
)

// BoundingBox represents a geographic bounding box
type BoundingBox struct {
	Min Point
	Max Point
}

const (
	// Earth radius in meters
	EarthRadius = 6371000.0
)

// HaversineDistance calculates the distance between two points using the Haversine formula
func HaversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	dLat := toRadians(lat2 - lat1)
	dLon := toRadians(lon2 - lon1)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(toRadians(lat1))*math.Cos(toRadians(lat2))*
			math.Sin(dLon/2)*math.Sin(dLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return EarthRadius * c
}

// CalculateBearing calculates the bearing from point 1 to point 2
func CalculateBearing(lat1, lon1, lat2, lon2 float64) float64 {
	dLon := toRadians(lon2 - lon1)
	lat1Rad := toRadians(lat1)
	lat2Rad := toRadians(lat2)

	y := math.Sin(dLon) * math.Cos(lat2Rad)
	x := math.Cos(lat1Rad)*math.Sin(lat2Rad) -
		math.Sin(lat1Rad)*math.Cos(lat2Rad)*math.Cos(dLon)

	bearing := toDegrees(math.Atan2(y, x))

	// Normalize to 0-360
	return math.Mod(bearing+360, 360)
}

// DouglasPeucker simplifies a line using the Douglas-Peucker algorithm
func DouglasPeucker(points []Point, tolerance float64) []Point {
	if len(points) <= 2 {
		return points
	}

	// Find the point with maximum distance from the line
	maxDist := 0.0
	maxIndex := 0

	for i := 1; i < len(points)-1; i++ {
		dist := perpendicularDistance(points[i], points[0], points[len(points)-1])
		if dist > maxDist {
			maxDist = dist
			maxIndex = i
		}
	}

	// If max distance is greater than tolerance, recursively simplify
	if maxDist > tolerance {
		// Recursive simplification
		left := DouglasPeucker(points[:maxIndex+1], tolerance)
		right := DouglasPeucker(points[maxIndex:], tolerance)

		// Combine results (avoiding duplicate middle point)
		result := make([]Point, 0, len(left)+len(right)-1)
		result = append(result, left[:len(left)-1]...)
		result = append(result, right...)
		return result
	}

	// Otherwise, return just the endpoints
	return []Point{points[0], points[len(points)-1]}
}

// CalculateBoundingBox calculates the bounding box for a set of points
func CalculateBoundingBox(points []Point) BoundingBox {
	if len(points) == 0 {
		return BoundingBox{}
	}

	minLat, maxLat := points[0].Lat, points[0].Lat
	minLon, maxLon := points[0].Lon, points[0].Lon

	for _, p := range points[1:] {
		if p.Lat < minLat {
			minLat = p.Lat
		}
		if p.Lat > maxLat {
			maxLat = p.Lat
		}
		if p.Lon < minLon {
			minLon = p.Lon
		}
		if p.Lon > maxLon {
			maxLon = p.Lon
		}
	}

	return BoundingBox{
		Min: Point{Lat: minLat, Lon: minLon},
		Max: Point{Lat: maxLat, Lon: maxLon},
	}
}

// PointInPolygon checks if a point is inside a polygon using ray casting
func PointInPolygon(point Point, polygon []Point) bool {
	if len(polygon) < 3 {
		return false
	}

	inside := false
	p1 := polygon[0]

	for i := 1; i <= len(polygon); i++ {
		p2 := polygon[i%len(polygon)]

		if point.Lon > math.Min(p1.Lon, p2.Lon) {
			if point.Lon <= math.Max(p1.Lon, p2.Lon) {
				if point.Lat <= math.Max(p1.Lat, p2.Lat) {
					if p1.Lon != p2.Lon {
						xinters := (point.Lon-p1.Lon)*(p2.Lat-p1.Lat)/(p2.Lon-p1.Lon) + p1.Lat
						if p1.Lat == p2.Lat || point.Lat <= xinters {
							inside = !inside
						}
					}
				}
			}
		}
		p1 = p2
	}

	return inside
}

// LinesIntersect checks if two line segments intersect
func LinesIntersect(p1, p2, p3, p4 Point) bool {
	d := (p1.Lat-p2.Lat)*(p3.Lon-p4.Lon) - (p1.Lon-p2.Lon)*(p3.Lat-p4.Lat)

	// Lines are parallel
	if math.Abs(d) < 1e-10 {
		// Check if they are collinear and overlapping
		return isPointOnSegment(p1, p3, p4) || isPointOnSegment(p2, p3, p4) ||
			isPointOnSegment(p3, p1, p2) || isPointOnSegment(p4, p1, p2)
	}

	t := ((p1.Lat-p3.Lat)*(p3.Lon-p4.Lon) - (p1.Lon-p3.Lon)*(p3.Lat-p4.Lat)) / d
	u := -((p1.Lat-p2.Lat)*(p1.Lon-p3.Lon) - (p1.Lon-p2.Lon)*(p1.Lat-p3.Lat)) / d

	return t >= 0 && t <= 1 && u >= 0 && u <= 1
}

// ConvexHull calculates the convex hull of a set of points using Graham's scan
func ConvexHull(points []Point) []Point {
	if len(points) < 3 {
		return points
	}

	// Find the point with lowest y-coordinate (and leftmost if tie)
	start := 0
	for i := 1; i < len(points); i++ {
		if points[i].Lat < points[start].Lat ||
			(points[i].Lat == points[start].Lat && points[i].Lon < points[start].Lon) {
			start = i
		}
	}

	// Sort points by polar angle with respect to start point
	sorted := make([]Point, len(points))
	copy(sorted, points)
	sorted[0], sorted[start] = sorted[start], sorted[0]

	// Simple implementation - for production use a proper sorting algorithm
	hull := []Point{sorted[0]}

	for i := 1; i < len(sorted); i++ {
		// Remove points that make clockwise turn
		for len(hull) > 1 && !ccw(hull[len(hull)-2], hull[len(hull)-1], sorted[i]) {
			hull = hull[:len(hull)-1]
		}
		hull = append(hull, sorted[i])
	}

	// Handle collinear case
	if len(hull) == 2 && len(points) > 2 {
		// All points are collinear
		return []Point{hull[0], hull[len(hull)-1]}
	}

	return hull
}

// WGS84ToProjection converts WGS84 coordinates to a projection (simplified for Web Mercator)
func WGS84ToProjection(lat, lon float64, epsg int) (x, y float64) {
	if epsg == 3857 { // Web Mercator
		x = lon * 20037508.34 / 180
		y = math.Log(math.Tan((90+lat)*math.Pi/360)) / (math.Pi / 180)
		y = y * 20037508.34 / 180
		return x, y
	}
	// For other projections, return as-is (would need proj4 library for full support)
	return lon, lat
}

// ProjectionToWGS84 converts projected coordinates back to WGS84
func ProjectionToWGS84(x, y float64, epsg int) (lat, lon float64) {
	if epsg == 3857 { // Web Mercator
		lon = x * 180 / 20037508.34
		lat = math.Atan(math.Exp(y*math.Pi/20037508.34))*360/math.Pi - 90
		return lat, lon
	}
	// For other projections, return as-is
	return y, x
}

// Helper functions

func toRadians(degrees float64) float64 {
	return degrees * math.Pi / 180
}

func toDegrees(radians float64) float64 {
	return radians * 180 / math.Pi
}

func perpendicularDistance(point, lineStart, lineEnd Point) float64 {
	if lineStart.Lat == lineEnd.Lat && lineStart.Lon == lineEnd.Lon {
		return HaversineDistance(point.Lat, point.Lon, lineStart.Lat, lineStart.Lon)
	}

	// Calculate perpendicular distance using cross product
	A := point.Lon - lineStart.Lon
	B := point.Lat - lineStart.Lat
	C := lineEnd.Lon - lineStart.Lon
	D := lineEnd.Lat - lineStart.Lat

	dot := A*C + B*D
	lenSq := C*C + D*D
	param := dot / lenSq

	var xx, yy float64

	switch {
	case param < 0:
		xx = lineStart.Lon
		yy = lineStart.Lat
	case param > 1:
		xx = lineEnd.Lon
		yy = lineEnd.Lat
	default:
		xx = lineStart.Lon + param*C
		yy = lineStart.Lat + param*D
	}

	return HaversineDistance(point.Lat, point.Lon, yy, xx)
}

func isPointOnSegment(point, segStart, segEnd Point) bool {
	// Check if point is on the line segment
	minLat := math.Min(segStart.Lat, segEnd.Lat)
	maxLat := math.Max(segStart.Lat, segEnd.Lat)
	minLon := math.Min(segStart.Lon, segEnd.Lon)
	maxLon := math.Max(segStart.Lon, segEnd.Lon)

	if point.Lat < minLat || point.Lat > maxLat || point.Lon < minLon || point.Lon > maxLon {
		return false
	}

	// Check if point is collinear with segment
	area := (segEnd.Lon-segStart.Lon)*(point.Lat-segStart.Lat) -
		(point.Lon-segStart.Lon)*(segEnd.Lat-segStart.Lat)

	return math.Abs(area) < 1e-10
}

func ccw(a, b, c Point) bool {
	// Counter-clockwise test
	return (b.Lon-a.Lon)*(c.Lat-a.Lat)-(b.Lat-a.Lat)*(c.Lon-a.Lon) > 0
}
