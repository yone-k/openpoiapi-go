package openpoiapi

import (
	"net/url"
	"strconv"
)

// Point is a WGS84 coordinate. Lng (longitude) comes first, matching the
// "lng,lat" order used by the API.
type Point struct {
	// Lng is the longitude in degrees.
	Lng float64
	// Lat is the latitude in degrees.
	Lat float64
}

// BBox is a rectangular area given by its south-west and north-east corners,
// matching the "minLng,minLat,maxLng,maxLat" order used by the API.
type BBox struct {
	// MinLng is the western longitude in degrees.
	MinLng float64
	// MinLat is the southern latitude in degrees.
	MinLat float64
	// MaxLng is the eastern longitude in degrees.
	MaxLng float64
	// MaxLat is the northern latitude in degrees.
	MaxLat float64
}

func validateLocation(center *Point, bbox *BBox, radius int) error {
	if center != nil {
		if !validLng(center.Lng) {
			return &ValidationError{Field: "center", Reason: "longitude must be between -180 and 180"}
		}
		if !validLat(center.Lat) {
			return &ValidationError{Field: "center", Reason: "latitude must be between -90 and 90"}
		}
	}
	if bbox != nil {
		if !validLng(bbox.MinLng) || !validLng(bbox.MaxLng) {
			return &ValidationError{Field: "bbox", Reason: "longitudes must be between -180 and 180"}
		}
		if !validLat(bbox.MinLat) || !validLat(bbox.MaxLat) {
			return &ValidationError{Field: "bbox", Reason: "latitudes must be between -90 and 90"}
		}
		if bbox.MinLng > bbox.MaxLng {
			return &ValidationError{Field: "bbox", Reason: "MinLng must not be greater than MaxLng"}
		}
		if bbox.MinLat > bbox.MaxLat {
			return &ValidationError{Field: "bbox", Reason: "MinLat must not be greater than MaxLat"}
		}
	}
	if radius < 0 {
		return &ValidationError{Field: "radius", Reason: "must not be negative"}
	}
	return nil
}

// NaN fails every comparison and ±Inf is outside the range, so the range
// check rejects non-finite values without separate math.IsNaN/IsInf calls.
func validLng(v float64) bool {
	return v >= -180 && v <= 180
}

func validLat(v float64) bool {
	return v >= -90 && v <= 90
}

func setPoint(query url.Values, key string, p *Point) {
	if p == nil {
		return
	}
	query.Set(key, formatCoord(p.Lng)+","+formatCoord(p.Lat))
}

func setBBox(query url.Values, key string, b *BBox) {
	if b == nil {
		return
	}
	query.Set(key, formatCoord(b.MinLng)+","+formatCoord(b.MinLat)+","+formatCoord(b.MaxLng)+","+formatCoord(b.MaxLat))
}

// The 'g' format would switch to exponent notation for small or large
// values, which the API's coordinate pattern rejects.
func formatCoord(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64)
}
