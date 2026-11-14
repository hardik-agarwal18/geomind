
-- ==========================================
-- 1. Insert a sample category
-- ==========================================

WITH category AS (
    INSERT INTO categories (
        id,
        slug,
        name
    )
    VALUES (
        '11111111-1111-4111-8111-111111111111',
        'cafe',
        'Cafe'
    )
    ON CONFLICT (slug)
    DO UPDATE SET name = EXCLUDED.name
    RETURNING id
)

-- ==========================================
-- 2. Insert a representative place
-- ==========================================

INSERT INTO places (
    id,
    category_id,
    name,
    location,
    location_source,
    location_observed_at
)
SELECT
    '22222222-2222-4222-8222-222222222222',
    category.id,
    'GeoMind Bengaluru Demo Cafe',

    ST_SetSRID(
        ST_MakePoint(77.5946, 12.9716),
        4326
    )::geography,

    'seed:day8',
    TIMESTAMPTZ '2026-11-14 00:00:00+05:30'

FROM category

ON CONFLICT (id)
DO UPDATE SET
    category_id = EXCLUDED.category_id,
    name = EXCLUDED.name,
    location = EXCLUDED.location,
    location_source = EXCLUDED.location_source,
    location_observed_at = EXCLUDED.location_observed_at;


-- ==========================================
-- 3. Insert a service area
-- ==========================================

INSERT INTO service_areas (
    id,
    slug,
    name,
    description,
    boundary
)
VALUES (
    '33333333-3333-4333-8333-333333333333',

    'central-bengaluru-demo',

    'Central Bengaluru Demo Zone',

    'Synthetic polygon used for spatial development and testing.',

    ST_Multi(
        ST_GeomFromText(
            'POLYGON((
                77.58 12.96,
                77.61 12.96,
                77.61 12.985,
                77.58 12.985,
                77.58 12.96
            ))',
            4326
        )
    )
)

ON CONFLICT (slug)
DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    boundary = EXCLUDED.boundary;
