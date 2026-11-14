package geo

import (
	"fmt"
	"math"
)

// Coordinate represents a WGS 84 geographic coordinate.
//
// Always use longitude as X and latitude as Y.
type Coordinate struct {
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
}

// Validate checks whether the coordinate contains
// finite values within WGS 84 geographic bounds.
func (c Coordinate) Validate() error {
	if math.IsNaN(c.Longitude) ||
		math.IsInf(c.Longitude, 0) {
		return fmt.Errorf("longitude must be finite")
	}

	if math.IsNaN(c.Latitude) ||
		math.IsInf(c.Latitude, 0) {
		return fmt.Errorf("latitude must be finite")
	}

	if c.Longitude < -180 || c.Longitude > 180 {
		return fmt.Errorf(
			"longitude must be between -180 and 180",
		)
	}

	if c.Latitude < -90 || c.Latitude > 90 {
		return fmt.Errorf(
			"latitude must be between -90 and 90",
		)
	}

	return nil
}
