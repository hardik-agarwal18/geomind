
-- +goose Up

ALTER TABLE places
ADD COLUMN location geography(Point, 4326),

ADD COLUMN location_source TEXT,

ADD COLUMN location_observed_at TIMESTAMPTZ,

ADD COLUMN location_accuracy_m DOUBLE PRECISION,

ADD CONSTRAINT places_location_accuracy_check
CHECK (
    location_accuracy_m IS NULL
    OR (
        location_accuracy_m >= 0
        AND location_accuracy_m < 'Infinity'::DOUBLE PRECISION
    )
),

ADD CONSTRAINT places_location_provenance_check
CHECK (
    (
        location IS NULL
        AND location_source IS NULL
        AND location_observed_at IS NULL
    )
    OR
    (
        location IS NOT NULL
        AND location_source IS NOT NULL
        AND length(btrim(location_source)) > 0
        AND location_observed_at IS NOT NULL
    )
);

-- Spatial index for proximity queries.
CREATE INDEX places_location_gix
ON places
USING GIST (location)
WHERE location IS NOT NULL;


-- +goose Down

DROP INDEX IF EXISTS places_location_gix;

ALTER TABLE places
DROP CONSTRAINT IF EXISTS places_location_provenance_check,
DROP CONSTRAINT IF EXISTS places_location_accuracy_check,
DROP COLUMN IF EXISTS location_accuracy_m,
DROP COLUMN IF EXISTS location_observed_at,
DROP COLUMN IF EXISTS location_source,
DROP COLUMN IF EXISTS location;
