package places

import (
	"context"
	"fmt"
	"math"
	"regexp"

	"geomind/internal/geo"

	"github.com/google/uuid"
)

const (
	DefaultNearbyRadiusM = 1000.0
	MaxNearbyRadiusM     = 50000.0

	DefaultNearbyLimit int32 = 20
	MaxNearbyLimit     int32 = 100
)

var categorySlugPattern = regexp.MustCompile(
	`^[a-z0-9]+(-[a-z0-9]+)*$`,
)

type NearbyParams struct {
	Origin       geo.Coordinate
	RadiusM      float64
	CategorySlug string
	Limit        int32
}

type NearbyPlace struct {
	ID        uuid.UUID
	Name      string
	Category  string
	Latitude  float64
	Longitude float64
	DistanceM float64
}

func (p NearbyParams) Validate() error {
	if err := p.Origin.Validate(); err != nil {
		return fmt.Errorf("invalid origin: %w", err)
	}

	if math.IsNaN(p.RadiusM) ||
		math.IsInf(p.RadiusM, 0) ||
		p.RadiusM <= 0 ||
		p.RadiusM > MaxNearbyRadiusM {
		return fmt.Errorf(
			"radius must be greater than zero and at most %.0f metres",
			MaxNearbyRadiusM,
		)
	}

	if p.Limit < 1 || p.Limit > MaxNearbyLimit {
		return fmt.Errorf(
			"limit must be between 1 and %d",
			MaxNearbyLimit,
		)
	}

	if p.CategorySlug != "" &&
		(len(p.CategorySlug) > 64 ||
			!categorySlugPattern.MatchString(p.CategorySlug)) {
		return fmt.Errorf("invalid category slug")
	}

	return nil
}

// NearbyRepository describes the persistence capability
// required by the places service.
type NearbyRepository interface {
	Nearby(
		ctx context.Context,
		params NearbyParams,
	) ([]NearbyPlace, error)
}

type Service struct {
	repository NearbyRepository
}

func NewService(repository NearbyRepository) *Service {
	if repository == nil {
		panic("places service requires a repository")
	}

	return &Service{
		repository: repository,
	}
}

func (s *Service) SearchNearby(
	ctx context.Context,
	params NearbyParams,
) ([]NearbyPlace, error) {

	if err := params.Validate(); err != nil {
		return nil, err
	}

	return s.repository.Nearby(ctx, params)
}
