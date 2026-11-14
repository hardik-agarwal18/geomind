
-- +goose Up

-- USERS

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    email TEXT NOT NULL
        CHECK (email <> '' AND email = btrim(email)),

    display_name TEXT NOT NULL
        CHECK (length(btrim(display_name)) > 0),

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Case-insensitive uniqueness for email addresses.
CREATE UNIQUE INDEX users_email_unique_idx
ON users (LOWER(email));


-- CATEGORIES

CREATE TABLE categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    slug TEXT NOT NULL,

    name TEXT NOT NULL
        CHECK (length(btrim(name)) > 0),

    description TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT categories_slug_unique UNIQUE (slug),

    CONSTRAINT categories_slug_format_check
        CHECK (slug ~ '^[a-z0-9]+(-[a-z0-9]+)*$')
);


-- PLACES

CREATE TABLE places (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    category_id UUID NOT NULL,

    name TEXT NOT NULL
        CHECK (length(btrim(name)) > 0),

    address_line TEXT,
    locality TEXT,
    region TEXT,
    postal_code TEXT,

    country_code TEXT
        CHECK (
            country_code IS NULL
            OR country_code ~ '^[A-Z]{2}$'
        ),

    -- Flexible non-spatial metadata.
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb
        CHECK (jsonb_typeof(metadata) = 'object'),

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT places_category_fk
        FOREIGN KEY (category_id)
        REFERENCES categories(id)
        ON DELETE RESTRICT
);

CREATE INDEX places_category_id_idx
ON places (category_id);


-- +goose Down

DROP TABLE IF EXISTS places;

DROP TABLE IF EXISTS categories;

DROP TABLE IF EXISTS users;
