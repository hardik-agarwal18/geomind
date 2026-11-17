package places

import (
	"context"
	"errors"
	"testing"
	"time"

	"geomind/internal/geo"

	"github.com/google/uuid"
)

func TestRepositoryCreateGetAndList(t *testing.T) {
	pool := newTestPool(t)
	repo := NewRepository(pool)

	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	categoryID := uuid.New()
	categorySlug := "integration-" + uuid.NewString()

	_, err := pool.Exec(
		ctx,
		`INSERT INTO categories (id, slug, name)
         VALUES ($1, $2, $3)`,
		categoryID,
		categorySlug,
		"Integration Test Category",
	)
	if err != nil {
		t.Fatalf("insert test category: %v", err)
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

	observedAt := time.Now().UTC().Truncate(time.Microsecond)
	accuracy := 7.5

	params := CreatePlaceParams{
		CategoryID: categoryID,
		Name:       "GeoMind Test Place",

		Location: geo.Coordinate{
			Longitude: 77.5946,
			Latitude:  12.9716,
		},

		LocationSource:     "integration-test",
		LocationObservedAt: observedAt,
		LocationAccuracyM:  &accuracy,
	}

	if err := params.Location.Validate(); err != nil {
		t.Fatalf("validate test coordinate: %v", err)
	}

	created, err := repo.Create(ctx, params)
	if err != nil {
		t.Fatalf("create place: %v", err)
	}

	if created.ID == uuid.Nil {
		t.Fatal("expected generated place ID")
	}

	if created.Location == nil {
		t.Fatal("expected location")
	}

	if created.Location.Longitude != 77.5946 {
		t.Errorf(
			"unexpected longitude: %v",
			created.Location.Longitude,
		)
	}

	if created.Location.Latitude != 12.9716 {
		t.Errorf(
			"unexpected latitude: %v",
			created.Location.Latitude,
		)
	}

	fetched, err := repo.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("get place: %v", err)
	}

	if fetched.ID != created.ID {
		t.Errorf("place ID mismatch")
	}

	if fetched.Name != params.Name {
		t.Errorf(
			"expected name %q, got %q",
			params.Name,
			fetched.Name,
		)
	}

	listed, err := repo.ListByCategory(
		ctx,
		categoryID,
		10,
		0,
	)
	if err != nil {
		t.Fatalf("list places: %v", err)
	}

	if len(listed) != 1 {
		t.Fatalf(
			"expected one place, got %d",
			len(listed),
		)
	}
}

func TestRepositoryGetByIDNotFound(t *testing.T) {
	pool := newTestPool(t)
	repo := NewRepository(pool)

	ctx, cancel := context.WithTimeout(
		context.Background(),
		2*time.Second,
	)
	defer cancel()

	_, err := repo.GetByID(ctx, uuid.New())

	if !errors.Is(err, ErrNotFound) {
		t.Fatalf(
			"expected ErrNotFound, got %v",
			err,
		)
	}
}

func TestDatabaseQueryTimeout(t *testing.T) {
	pool := newTestPool(t)

	ctx, cancel := context.WithTimeout(
		context.Background(),
		50*time.Millisecond,
	)
	defer cancel()

	_, err := pool.Exec(
		ctx,
		"SELECT pg_sleep(2)",
	)

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf(
			"expected context deadline exceeded, got %v",
			err,
		)
	}
}

func TestAcquireRespectsContextCancellation(t *testing.T) {
	pool := newTestPool(t)

	ctx := context.Background()

	conn1, err := pool.Acquire(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer conn1.Release()

	conn2, err := pool.Acquire(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer conn2.Release()

	// MaxConns is two. No connection is available.
	acquireCtx, cancel := context.WithTimeout(
		ctx,
		50*time.Millisecond,
	)
	defer cancel()

	_, err = pool.Acquire(acquireCtx)

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf(
			"expected acquisition timeout, got %v",
			err,
		)
	}
}

type failingRow struct {
	err error
}

func (r failingRow) Scan(dest ...any) error {
	return r.err
}

func TestScanPlaceReturnsError(t *testing.T) {
	expected := errors.New("simulated scan failure")

	_, err := scanPlace(failingRow{
		err: expected,
	})

	if !errors.Is(err, expected) {
		t.Fatalf(
			"expected scan error %v, got %v",
			expected,
			err,
		)
	}
}
