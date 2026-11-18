package places

import (
	"context"
	"math"
	"testing"
	"time"

	"geomind/internal/geo"

	"github.com/google/uuid"
)

func TestNearestKAndCategory(t *testing.T) {
	pool := newTestPool(t)
	repo := NewRepository(pool)

	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	cafeID := uuid.New()
	parkID := uuid.New()

	cafeSlug := "nearest-cafe-" + uuid.NewString()
	parkSlug := "nearest-park-" + uuid.NewString()

	_, err := pool.Exec(
		ctx,
		`INSERT INTO categories (id, slug, name)
         VALUES
            ($1, $2, 'Nearest Test Cafe'),
            ($3, $4, 'Nearest Test Park')`,
		cafeID,
		cafeSlug,
		parkID,
		parkSlug,
	)
	if err != nil {
		t.Fatalf("insert categories: %v", err)
	}

	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(
			context.Background(),
			3*time.Second,
		)
		defer cleanupCancel()

		_, _ = pool.Exec(
			cleanupCtx,
			`DELETE FROM places
             WHERE category_id = ANY($1::uuid[])`,
			[]uuid.UUID{cafeID, parkID},
		)

		_, _ = pool.Exec(
			cleanupCtx,
			`DELETE FROM categories
             WHERE id = ANY($1::uuid[])`,
			[]uuid.UUID{cafeID, parkID},
		)
	})

	origin := geo.Coordinate{
		Longitude: 77.5946,
		Latitude:  12.9716,
	}

	insertPlace := func(
		categoryID uuid.UUID,
		name string,
		distanceM float64,
	) uuid.UUID {
		t.Helper()

		id := uuid.New()

		_, err := pool.Exec(
			ctx,
			`
            INSERT INTO places (
                id,
                category_id,
                name,
                location,
                location_source,
                location_observed_at
            )
            VALUES (
                $1,
                $2,
                $3,
                ST_Project(
                    ST_SetSRID(
                        ST_MakePoint($4, $5),
                        4326
                    )::geography,
                    $6::double precision,
                    radians(90)
                ),
                'day13-integration-test',
                NOW()
            )
            `,
			id,
			categoryID,
			name,
			origin.Longitude,
			origin.Latitude,
			distanceM,
		)
		if err != nil {
			t.Fatalf("insert place %q: %v", name, err)
		}

		return id
	}

	cafe100 := insertPlace(cafeID, "Cafe A", 100)
	cafe300 := insertPlace(cafeID, "Cafe B", 300)
	insertPlace(parkID, "Park A", 200)
	insertPlace(cafeID, "Cafe C", 1500)

	params := NearestParams{
		Origin:       origin,
		K:            1,
		CategorySlug: cafeSlug,
	}

	nearest, err := repo.Nearest(ctx, params)
	if err != nil {
		t.Fatalf("nearest K=1: %v", err)
	}

	if len(nearest) != 1 {
		t.Fatalf("expected one result, got %d", len(nearest))
	}

	if nearest[0].ID != cafe100 {
		t.Errorf(
			"expected closest cafe %s, got %s",
			cafe100,
			nearest[0].ID,
		)
	}

	// K = maximum. Only three cafes exist.
	params.K = MaxNearestK

	nearest, err = repo.Nearest(ctx, params)
	if err != nil {
		t.Fatalf("nearest maximum K: %v", err)
	}

	if len(nearest) != 3 {
		t.Fatalf(
			"expected three cafes, got %d",
			len(nearest),
		)
	}

	if nearest[0].ID != cafe100 ||
		nearest[1].ID != cafe300 {
		t.Fatalf("unexpected nearest ordering: %+v", nearest)
	}

	if math.Abs(nearest[0].DistanceM-100) > 1 {
		t.Errorf(
			"expected distance around 100m, got %.3f",
			nearest[0].DistanceM,
		)
	}

	if math.Abs(nearest[0].KNNDistanceM-100) > 1 {
		t.Errorf(
			"expected KNN distance around 100m, got %.3f",
			nearest[0].KNNDistanceM,
		)
	}

	// A valid but nonexistent category must return empty results.
	params.CategorySlug = "missing-" + uuid.NewString()

	empty, err := repo.Nearest(ctx, params)
	if err != nil {
		t.Fatalf("empty nearest search: %v", err)
	}

	if len(empty) != 0 {
		t.Fatalf(
			"expected empty results, got %d",
			len(empty),
		)
	}
}

func TestNearestEqualDistanceUsesUUIDTieBreaker(t *testing.T) {
	pool := newTestPool(t)
	repo := NewRepository(pool)

	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	categoryID := uuid.New()
	slug := "nearest-tie-" + uuid.NewString()

	_, err := pool.Exec(
		ctx,
		`INSERT INTO categories (id, slug, name)
         VALUES ($1, $2, 'Nearest Tie Test')`,
		categoryID,
		slug,
	)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(
			context.Background(),
			3*time.Second,
		)
		defer cleanupCancel()

		_, _ = pool.Exec(
			cleanupCtx,
			`DELETE FROM places WHERE category_id = $1`,
			categoryID,
		)

		_, _ = pool.Exec(
			cleanupCtx,
			`DELETE FROM categories WHERE id = $1`,
			categoryID,
		)
	})

	id1 := uuid.New()
	id2 := uuid.New()

	// Ensure id1 is lexically smaller than id2.
	if id1.String() > id2.String() {
		id1, id2 = id2, id1
	}

	origin := geo.Coordinate{
		Longitude: 0,
		Latitude:  0,
	}

	// Symmetric points on the equator have equal distance
	// from the origin under both sphere and spheroid metrics.
	_, err = pool.Exec(
		ctx,
		`
        INSERT INTO places (
            id,
            category_id,
            name,
            location,
            location_source,
            location_observed_at
        )
        VALUES
        (
            $1, $2, 'Tie A',
            ST_SetSRID(ST_MakePoint(-0.001, 0), 4326)::geography,
            'day13-test', NOW()
        ),
        (
            $3, $2, 'Tie B',
            ST_SetSRID(ST_MakePoint(0.001, 0), 4326)::geography,
            'day13-test', NOW()
        )
        `,
		id1,
		categoryID,
		id2,
	)
	if err != nil {
		t.Fatal(err)
	}

	results, err := repo.Nearest(ctx, NearestParams{
		Origin:       origin,
		K:            2,
		CategorySlug: slug,
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(results) != 2 {
		t.Fatalf("expected two results, got %d", len(results))
	}

	if results[0].ID != id1 || results[1].ID != id2 {
		t.Fatalf(
			"UUID tie-breaking failed: got %s, %s; want %s, %s",
			results[0].ID,
			results[1].ID,
			id1,
			id2,
		)
	}
}
