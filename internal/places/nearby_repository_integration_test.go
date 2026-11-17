package places

import (
	"context"
	"geomind/internal/geo"
	"math"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNearbyFiltersByRadiusAndCategory(t *testing.T) {
	pool := newTestPool(t)
	repo := NewRepository(pool)

	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	cafeID := uuid.New()
	parkID := uuid.New()

	cafeSlug := "day12-cafe-" + uuid.NewString()
	parkSlug := "day12-park-" + uuid.NewString()

	_, err := pool.Exec(
		ctx,
		`INSERT INTO categories (id, slug, name)
         VALUES
            ($1, $2, 'Day 12 Test Cafe'),
            ($3, $4, 'Day 12 Test Park')`,
		cafeID,
		cafeSlug,
		parkID,
		parkSlug,
	)
	if err != nil {
		t.Fatalf("create test categories: %v", err)
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

	originLongitude := 10.0
	originLatitude := 10.0

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
                'day12-integration-test',
                NOW()
            )
            `,
			id,
			categoryID,
			name,
			originLongitude,
			originLatitude,
			distanceM,
		)
		if err != nil {
			t.Fatalf("insert test place %q: %v", name, err)
		}

		return id
	}

	insertPlace(cafeID, "Cafe 100m", 100)
	insertPlace(cafeID, "Cafe 300m", 300)
	insertPlace(parkID, "Park 200m", 200)
	insertPlace(cafeID, "Cafe 1500m", 1500)

	params := NearbyParams{
		Origin: geo.Coordinate{
			Longitude: originLongitude,
			Latitude:  originLatitude,
		},
		RadiusM:      500,
		CategorySlug: cafeSlug,
		Limit:        20,
	}

	results, err := repo.Nearby(ctx, params)
	if err != nil {
		t.Fatalf("nearby search: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf(
			"expected 2 nearby cafes, got %d",
			len(results),
		)
	}

	if results[0].Name != "Cafe 100m" {
		t.Errorf(
			"expected nearest place first, got %q",
			results[0].Name,
		)
	}

	if results[1].Name != "Cafe 300m" {
		t.Errorf(
			"expected second place Cafe 300m, got %q",
			results[1].Name,
		)
	}

	if math.Abs(results[0].DistanceM-100) > 1 {
		t.Errorf(
			"expected distance around 100m, got %.3f",
			results[0].DistanceM,
		)
	}

	if math.Abs(results[1].DistanceM-300) > 1 {
		t.Errorf(
			"expected distance around 300m, got %.3f",
			results[1].DistanceM,
		)
	}
}
