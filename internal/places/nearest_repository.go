package places

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type NearestDB struct {
	pool *pgxpool.Pool
}

// If you prefer, add Nearest directly to your existing Repository
// instead of creating another database wrapper.
const nearestQuery = `
    SELECT
        p.id,
        p.name,
        c.slug,

        ST_Y(p.location::geometry) AS latitude,
        ST_X(p.location::geometry) AS longitude,

        ST_Distance(
            p.location,
            ST_SetSRID(
                ST_MakePoint($1, $2),
                4326
            )::geography
        ) AS distance_m,

        p.location <-> ST_SetSRID(
            ST_MakePoint($1, $2),
            4326
        )::geography AS knn_distance_m

    FROM places AS p

    JOIN categories AS c
        ON c.id = p.category_id

    WHERE
        p.location IS NOT NULL

        AND (
            $3::text IS NULL
            OR c.slug = $3
        )

    ORDER BY
        p.location <-> ST_SetSRID(
            ST_MakePoint($1, $2),
            4326
        )::geography ASC,

        p.id ASC

    LIMIT $4
`

func (r *Repository) Nearest(
	ctx context.Context,
	params NearestParams,
) ([]NearestPlace, error) {
	var category any

	if params.CategorySlug != "" {
		category = params.CategorySlug
	}

	rows, err := r.pool.Query(
		ctx,
		nearestQuery,
		params.Origin.Longitude,
		params.Origin.Latitude,
		category,
		params.K,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"execute nearest places query: %w",
			err,
		)
	}

	defer rows.Close()

	result := make([]NearestPlace, 0)

	for rows.Next() {
		var place NearestPlace

		if err := rows.Scan(
			&place.ID,
			&place.Name,
			&place.Category,
			&place.Latitude,
			&place.Longitude,
			&place.DistanceM,
			&place.KNNDistanceM,
		); err != nil {
			return nil, fmt.Errorf(
				"scan nearest place: %w",
				err,
			)
		}

		result = append(result, place)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"iterate nearest places: %w",
			err,
		)
	}

	return result, nil
}

var _ NearestRepository = (*Repository)(nil)
