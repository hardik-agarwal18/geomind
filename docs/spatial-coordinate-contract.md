# GeoMind Spatial Coordinate Contract

## 1. Coordinate system

GeoMind initially uses WGS 84 (EPSG:4326).

## 2. Coordinate order

All application coordinate structures use:

- Longitude: X
- Latitude: Y

PostGIS construction:

ST_MakePoint(longitude, latitude)

Never reverse the order.

## 3. Valid ranges

Longitude: -180 to +180 degrees.

Latitude: -90 to +90 degrees.

NaN and Infinity are rejected.

## 4. Spatial storage

places.location:

geography(Point, 4326)

Coordinates represent longitude and latitude.

Geography distance calculations return metres.

## 5. Measurement conventions

location_accuracy_m:
Horizontal accuracy in metres.

ST_Distance(geography, geography):
Distance in metres.

ST_DWithin(geography, geography, distance):
Distance threshold in metres.

Geometry with EPSG:4326:
Planar measurements use angular coordinate units,
not metres.

## 6. Coordinate provenance

Every populated location must include:

- location_source
- location_observed_at

Reported horizontal accuracy is stored separately.

## 7. Timestamp semantics

location_observed_at:
When the location was measured or reported.

created_at:
When the record was inserted into GeoMind.

## 8. Validation

Coordinates must be validated before database insertion.

Validation includes:

- NaN rejection.
- Infinity rejection.
- Longitude bounds.
- Latitude bounds.

Global range validation does not independently
detect all longitude/latitude swaps.
