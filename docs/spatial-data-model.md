# GeoMind Spatial Data Model

## Spatial reference system

WGS 84 / EPSG:4326.

Coordinates use longitude followed by latitude.

## Places

Column:
location geography(Point, 4326)

Represents one geographic location.

Distance measurements:
Metres.

## Service Areas

Column:
boundary geometry(MultiPolygon, 4326)

Represents one or more connected or disconnected
service coverage regions.

Uses:

- Point-in-polygon checks
- Service-area eligibility
- Boundary intersection
- GeoJSON interoperability

## Spatial indexes

service_areas.boundary:
GiST index.

places.location:
GiST index from the previous spatial migration.

## Polygon validity

Every service area must be:

- Non-empty
- Valid according to PostGIS
- Closed at every ring
- Free from self-intersections

## GeoJSON

Input:
ST_GeomFromGeoJSON

Output:
ST_AsGeoJSON

Coordinate order:
[longitude, latitude]

## Storage decisions

POINT:
Individual places.

LINESTRING:
Routes and paths.

POLYGON:
One connected geographic region.

MULTIPOLYGON:
One region with multiple polygon components.

GEOMETRYCOLLECTION:
Mixed spatial objects, generally for temporary
or specialised operations.
