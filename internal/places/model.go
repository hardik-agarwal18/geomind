package places

import (
	"time"

	"geomind/internal/geo"

	"github.com/google/uuid"
)

// Place represents a persisted GeoMind place.
type Place struct {
	ID         uuid.UUID
	CategoryID uuid.UUID

	Name string

	// May be nil for records without a known location.
	Location *geo.Coordinate

	LocationSource     *string
	LocationObservedAt *time.Time
	LocationAccuracyM  *float64

	CreatedAt time.Time
	UpdatedAt time.Time
}

// CreatePlaceParams contains the data required to create a place.
type CreatePlaceParams struct {
	CategoryID uuid.UUID
	Name       string

	Location geo.Coordinate

	LocationSource     string
	LocationObservedAt time.Time
	LocationAccuracyM  *float64
}
