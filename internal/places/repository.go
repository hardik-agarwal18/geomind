package places

import (
	"context"
	"errors"
	"fmt"

	"geomind/internal/geo"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("place not found")

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	if pool == nil {
		panic("places repository requires a non-nil pool")
	}

	return &Repository{
		pool: pool,
	}
}

// rowScanner lets both pgx.Row and test doubles use scanPlace.
type rowScanner interface {
	Scan(dest ...any) error
}

func scanPlace(row rowScanner) (Place, error) {
	var place Place

	var longitude pgtype.Float8
	var latitude pgtype.Float8
	var accuracy pgtype.Float8

	var source pgtype.Text
	var observedAt pgtype.Timestamptz

	err := row.Scan(
		&place.ID,
		&place.CategoryID,
		&place.Name,
		&longitude,
		&latitude,
		&source,
		&observedAt,
		&accuracy,
		&place.CreatedAt,
		&place.UpdatedAt,
	)

	if err != nil {
		return Place{}, fmt.Errorf(
			"scan place: %w",
			err,
		)
	}

	// A geographic point must have both coordinates.
	if longitude.Valid != latitude.Valid {
		return Place{}, fmt.Errorf(
			"invalid database location: incomplete coordinates",
		)
	}

	if longitude.Valid {
		place.Location = &geo.Coordinate{
			Longitude: longitude.Float64,
			Latitude:  latitude.Float64,
		}
	}

	if source.Valid {
		place.LocationSource = &source.String
	}

	if observedAt.Valid {
		place.LocationObservedAt = &observedAt.Time
	}

	if accuracy.Valid {
		value := accuracy.Float64
		place.LocationAccuracyM = &value
	}

	return place, nil
}

const placeProjection = `
    id,
    category_id,
    name,
    ST_X(location::geometry),
    ST_Y(location::geometry),
    location_source,
    location_observed_at,
    location_accuracy_m,
    created_at,
    updated_at
`

// Create inserts a place and returns the persisted record.
func (r *Repository) Create(
	ctx context.Context,
	params CreatePlaceParams,
) (Place, error) {

	query := `
        INSERT INTO places (
            category_id,
            name,
            location,
            location_source,
            location_observed_at,
            location_accuracy_m
        )
        VALUES (
            $1,
            $2,
            ST_SetSRID(
                ST_MakePoint($3, $4),
                4326
            )::geography,
            $5,
            $6,
            $7
        )
        RETURNING ` + placeProjection

	place, err := scanPlace(
		r.pool.QueryRow(
			ctx,
			query,
			params.CategoryID,
			params.Name,
			params.Location.Longitude,
			params.Location.Latitude,
			params.LocationSource,
			params.LocationObservedAt,
			params.LocationAccuracyM,
		),
	)

	if err != nil {
		return Place{}, fmt.Errorf(
			"create place: %w",
			err,
		)
	}

	return place, nil
}

// GetByID returns one place by its identifier.
func (r *Repository) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (Place, error) {

	query := `
        SELECT ` + placeProjection + `
        FROM places
        WHERE id = $1
    `

	place, err := scanPlace(
		r.pool.QueryRow(ctx, query, id),
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return Place{}, fmt.Errorf(
			"%w: %s",
			ErrNotFound,
			id,
		)
	}

	if err != nil {
		return Place{}, fmt.Errorf(
			"get place: %w",
			err,
		)
	}

	return place, nil
}

// ListByCategory returns a deterministic page of places.
func (r *Repository) ListByCategory(
	ctx context.Context,
	categoryID uuid.UUID,
	limit int32,
	offset int32,
) ([]Place, error) {

	query := `
        SELECT ` + placeProjection + `
        FROM places

        WHERE category_id = $1

        ORDER BY created_at DESC, id ASC

        LIMIT $2
        OFFSET $3
    `

	rows, err := r.pool.Query(
		ctx,
		query,
		categoryID,
		limit,
		offset,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"list places: %w",
			err,
		)
	}

	defer rows.Close()

	result := make([]Place, 0)

	for rows.Next() {
		place, err := scanPlace(rows)
		if err != nil {
			return nil, fmt.Errorf(
				"scan listed place: %w",
				err,
			)
		}

		result = append(result, place)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"iterate places: %w",
			err,
		)
	}

	return result, nil
}
