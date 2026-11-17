package places

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

const nearbySearchQuery = `
    WITH origin AS (
        SELECT
            ST_SetSRID(
                ST_MakePoint($1, $2),
                4326
            )::geography AS location
    )
    SELECT
        p.id,
        p.name,
        c.slug,

        ST_Y(p.location::geometry) AS latitude,
        ST_X(p.location::geometry) AS longitude,

        ST_Distance(
            p.location,
            origin.location
        ) AS distance_m

    FROM places AS p

    JOIN categories AS c
        ON c.id = p.category_id

    CROSS JOIN origin

    WHERE
        p.location IS NOT NULL

        AND ST_DWithin(
            p.location,
            origin.location,
            $3
        )

        AND (
            $4::text IS NULL
            OR c.slug = $4
        )

    ORDER BY
        distance_m ASC,
        p.id ASC

    LIMIT $5
`

func (r *Repository) Nearby(
	ctx context.Context,
	params NearbyParams,
) ([]NearbyPlace, error) {

	var category any

	if params.CategorySlug != "" {
		category = params.CategorySlug
	}

	rows, err := r.pool.Query(
		ctx,
		nearbySearchQuery,
		params.Origin.Longitude,
		params.Origin.Latitude,
		params.RadiusM,
		category,
		params.Limit,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"execute nearby places query: %w",
			err,
		)
	}

	defer rows.Close()

	result := make([]NearbyPlace, 0)

	for rows.Next() {
		var place NearbyPlace

		if err := rows.Scan(
			&place.ID,
			&place.Name,
			&place.Category,
			&place.Latitude,
			&place.Longitude,
			&place.DistanceM,
		); err != nil {
			return nil, fmt.Errorf(
				"scan nearby place: %w",
				err,
			)
		}

		result = append(result, place)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"iterate nearby places: %w",
			err,
		)
	}

	return result, nil
}

// Compile-time check that Repository implements NearbyRepository.
var _ NearbyRepository = (*Repository)(nil)

// Keep this import used when the repository evolves to map
// pgx-specific errors at this boundary.
var _ = pgx.ErrNoRows
