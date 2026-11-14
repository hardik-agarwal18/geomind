
-- +goose Up

-- PostGIS must be installed in the PostgreSQL image.
-- Provision it using an appropriately privileged database role
-- before running ordinary migrations when required.
CREATE EXTENSION IF NOT EXISTS postgis;

-- Service areas represent geographic boundaries.
CREATE TABLE service_areas (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    slug TEXT NOT NULL UNIQUE,

    name TEXT NOT NULL
        CHECK (length(btrim(name)) > 0),

    description TEXT,

    -- A service area can contain multiple disconnected regions.
    boundary geometry(MultiPolygon, 4326) NOT NULL,

    is_active BOOLEAN NOT NULL DEFAULT TRUE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT service_areas_boundary_not_empty
        CHECK (NOT ST_IsEmpty(boundary)),

    CONSTRAINT service_areas_boundary_valid
        CHECK (ST_IsValid(boundary))
);

-- Spatial index for polygon-based queries.
CREATE INDEX service_areas_boundary_gix
ON service_areas
USING GIST (boundary);


-- +goose Down

DROP TABLE IF EXISTS service_areas;

-- Intentionally retain the PostGIS extension.
-- Other application tables may depend on it.
