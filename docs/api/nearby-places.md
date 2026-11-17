# Nearby Places API

## Endpoint

GET /api/v1/places/nearby

## Query parameters

| Parameter | Required | Default | Constraint                    |
| --------- | -------: | ------: | ----------------------------- |
| latitude  |      Yes |       — | -90 to 90                     |
| longitude |      Yes |       — | -180 to 180                   |
| radius_m  |       No |    1000 | (0, 50000] metres             |
| category  |       No |     All | Existing category slug format |
| limit     |       No |      20 | 1 to 100                      |

## Success response

HTTP 200

{
"data": [
{
"id": "UUID",
"name": "Place name",
"category": "cafe",
"latitude": 12.9716,
"longitude": 77.5946,
"distance_m": 0
}
],
"meta": {
"count": 1,
"latitude": 12.9716,
"longitude": 77.5946,
"radius_m": 1000,
"limit": 20,
"category": null
}
}

## Empty results

An empty search returns HTTP 200 with data as an empty array.

## Error response

{
"error": {
"code": "invalid_coordinate",
"message": "latitude must be a finite number between -90 and 90"
}
}

## Examples

Search within the default radius:

curl -G http://localhost:8080/api/v1/places/nearby \
 --data-urlencode "latitude=12.9716" \
 --data-urlencode "longitude=77.5946"

Search within 2 kilometres:

curl -G http://localhost:8080/api/v1/places/nearby \
 --data-urlencode "latitude=12.9716" \
 --data-urlencode "longitude=77.5946" \
 --data-urlencode "radius_m=2000"

Search cafes only:

curl -G http://localhost:8080/api/v1/places/nearby \
 --data-urlencode "latitude=12.9716" \
 --data-urlencode "longitude=77.5946" \
 --data-urlencode "category=cafe"
