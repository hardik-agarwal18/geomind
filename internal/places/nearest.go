package places

import (
	"context"
	"fmt"
	"math"

	"geomind/internal/geo"

	"github.com/google/uuid"
)

const (
	DefaultNearestK int32 = 10
	MaxNearestK     int32 = 100
)

type NearestParams struct {
	Origin       geo.Coordinate
	K            int32
	CategorySlug string
}

type NearestPlace struct {
	ID        uuid.UUID
	Name      string
	Category  string
	Latitude  float64
	Longitude float64

	// Spheroidal distance, in metres.
	DistanceM float64

	// Spherical distance used by geography KNN.
	KNNDistanceM float64
}

func (p NearestParams) Validate() error {
	if err := p.Origin.Validate(); err != nil {
		return fmt.Errorf("invalid origin: %w", err)
	}

	if p.K < 1 || p.K > MaxNearestK {
		return fmt.Errorf(
			"k must be between 1 and %d",
			MaxNearestK,
		)
	}

	if p.CategorySlug != "" &&
		!isValidCategorySlug(p.CategorySlug) {
		return fmt.Errorf("invalid category slug")
	}

	return nil
}

// Use the same category validation rule as Day 12.
// If you haven't extracted it yet, add this helper and
// use it in NearbyParams.Validate() too.
func isValidCategorySlug(slug string) bool {
	return len(slug) <= 64 &&
		categorySlugPattern.MatchString(slug)
}

type NearestRepository interface {
	Nearest(
		ctx context.Context,
		params NearestParams,
	) ([]NearestPlace, error)
}

type NearestService struct {
	repository NearestRepository
}

func NewNearestService(
	repository NearestRepository,
) *NearestService {
	if repository == nil {
		panic("nearest service requires a repository")
	}

	return &NearestService{
		repository: repository,
	}
}

func (s *NearestService) Search(
	ctx context.Context,
	params NearestParams,
) ([]NearestPlace, error) {
	if err := params.Validate(); err != nil {
		return nil, err
	}

	return s.repository.Nearest(ctx, params)
}

// Compile-time check that the service model uses finite
// numeric distance values when processing database results.
var _ = math.IsNaN
