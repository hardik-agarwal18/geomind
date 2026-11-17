package httpapi

import (
	"context"
	"errors"
	"log/slog"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"geomind/internal/geo"
	"geomind/internal/httpresponse"
	"geomind/internal/places"
)

const nearbyQueryTimeout = 2 * time.Second

var categorySlugPattern = regexp.MustCompile(
	`^[a-z0-9]+(-[a-z0-9]+)*$`,
)

type NearbyPlacesHandler struct {
	service *places.Service
	logger  *slog.Logger
}

func NewNearbyPlacesHandler(
	service *places.Service,
	logger *slog.Logger,
) *NearbyPlacesHandler {
	if logger == nil {
		logger = slog.Default()
	}

	return &NearbyPlacesHandler{
		service: service,
		logger:  logger,
	}
}

type nearbyPlaceResponse struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Category  string  `json:"category"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	DistanceM float64 `json:"distance_m"`
}

type nearbyMetaResponse struct {
	Count     int     `json:"count"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	RadiusM   float64 `json:"radius_m"`
	Limit     int32   `json:"limit"`
	Category  *string `json:"category"`
}

type nearbySearchResponse struct {
	Data []nearbyPlaceResponse `json:"data"`
	Meta nearbyMetaResponse    `json:"meta"`
}

func newNearbyResponse(
	params places.NearbyParams,
	results []places.NearbyPlace,
) nearbySearchResponse {
	data := make(
		[]nearbyPlaceResponse,
		0,
		len(results),
	)

	for _, place := range results {
		// Round for presentation only. SQL sorting uses
		// the full-precision distance.
		distance := math.Round(
			place.DistanceM*100,
		) / 100

		data = append(data, nearbyPlaceResponse{
			ID:        place.ID.String(),
			Name:      place.Name,
			Category:  place.Category,
			Latitude:  place.Latitude,
			Longitude: place.Longitude,
			DistanceM: distance,
		})
	}

	var category *string
	if params.CategorySlug != "" {
		value := params.CategorySlug
		category = &value
	}

	return nearbySearchResponse{
		Data: data,
		Meta: nearbyMetaResponse{
			Count:     len(data),
			Latitude:  params.Origin.Latitude,
			Longitude: params.Origin.Longitude,
			RadiusM:   params.RadiusM,
			Limit:     params.Limit,
			Category:  category,
		},
	}
}

type requestError struct {
	code    string
	message string
}

func (e *requestError) Error() string {
	return e.message
}

func badRequest(code, message string) error {
	return &requestError{
		code:    code,
		message: message,
	}
}

func (h *NearbyPlacesHandler) ServeHTTP(
	w http.ResponseWriter,
	r *http.Request,
) {
	params, err := parseNearbyQuery(r.URL.RawQuery)
	if err != nil {
		var requestErr *requestError

		if errors.As(err, &requestErr) {
			h.writeError(
				w,
				http.StatusBadRequest,
				requestErr.code,
				requestErr.message,
			)
			return
		}

		h.writeError(
			w,
			http.StatusBadRequest,
			"invalid_query",
			"Invalid query string",
		)
		return
	}

	queryCtx, cancel := context.WithTimeout(
		r.Context(),
		nearbyQueryTimeout,
	)
	defer cancel()

	results, err := h.service.SearchNearby(
		queryCtx,
		params,
	)
	if err != nil {
		// Do not attempt to write a response after
		// the client has cancelled its request.
		if r.Context().Err() != nil {
			return
		}

		if errors.Is(err, context.DeadlineExceeded) {
			h.writeError(
				w,
				http.StatusGatewayTimeout,
				"search_timeout",
				"Nearby search exceeded its time limit",
			)
			return
		}

		h.logger.Error(
			"nearby_search_failed",
			"error", err,
		)

		h.writeError(
			w,
			http.StatusInternalServerError,
			"internal_error",
			"An unexpected error occurred",
		)
		return
	}

	response := newNearbyResponse(params, results)

	if err := httpresponse.WriteJSON(
		w,
		http.StatusOK,
		response,
	); err != nil {
		h.logger.Error(
			"nearby_response_write_failed",
			"error", err,
		)
	}
}

func (h *NearbyPlacesHandler) writeError(
	w http.ResponseWriter,
	status int,
	code string,
	message string,
) {
	if err := httpresponse.WriteError(
		w,
		status,
		code,
		message,
	); err != nil {
		h.logger.Error(
			"nearby_error_response_failed",
			"error", err,
		)
	}
}

// singleValue ensures a parameter occurs at most once.
func singleValue(
	values url.Values,
	name string,
	required bool,
) (string, error) {
	items, exists := values[name]

	if !exists {
		if required {
			return "", badRequest(
				"missing_parameter",
				name+" is required",
			)
		}
		return "", nil
	}

	if len(items) != 1 {
		return "", badRequest(
			"duplicate_parameter",
			name+" must appear only once",
		)
	}

	value := strings.TrimSpace(items[0])

	if value == "" {
		return "", badRequest(
			"invalid_parameter",
			name+" must not be empty",
		)
	}

	return value, nil
}

func parseNearbyQuery(
	rawQuery string,
) (places.NearbyParams, error) {
	values, err := url.ParseQuery(rawQuery)
	if err != nil {
		return places.NearbyParams{}, err
	}

	allowed := map[string]bool{
		"latitude":  true,
		"longitude": true,
		"radius_m":  true,
		"category":  true,
		"limit":     true,
	}

	for key := range values {
		if !allowed[key] {
			return places.NearbyParams{}, badRequest(
				"unknown_parameter",
				"Unsupported query parameter: "+key,
			)
		}
	}

	latitudeText, err := singleValue(
		values,
		"latitude",
		true,
	)
	if err != nil {
		return places.NearbyParams{}, err
	}

	longitudeText, err := singleValue(
		values,
		"longitude",
		true,
	)
	if err != nil {
		return places.NearbyParams{}, err
	}

	latitude, err := strconv.ParseFloat(
		latitudeText,
		64,
	)
	if err != nil ||
		math.IsNaN(latitude) ||
		math.IsInf(latitude, 0) ||
		latitude < -90 ||
		latitude > 90 {
		return places.NearbyParams{}, badRequest(
			"invalid_coordinate",
			"latitude must be a finite number between -90 and 90",
		)
	}

	longitude, err := strconv.ParseFloat(
		longitudeText,
		64,
	)
	if err != nil ||
		math.IsNaN(longitude) ||
		math.IsInf(longitude, 0) ||
		longitude < -180 ||
		longitude > 180 {
		return places.NearbyParams{}, badRequest(
			"invalid_coordinate",
			"longitude must be a finite number between -180 and 180",
		)
	}

	params := places.NearbyParams{
		Origin: geo.Coordinate{
			Longitude: longitude,
			Latitude:  latitude,
		},
		RadiusM: places.DefaultNearbyRadiusM,
		Limit:   places.DefaultNearbyLimit,
	}

	radiusText, err := singleValue(
		values,
		"radius_m",
		false,
	)
	if err != nil {
		return places.NearbyParams{}, err
	}

	if radiusText != "" {
		radius, err := strconv.ParseFloat(
			radiusText,
			64,
		)
		if err != nil ||
			math.IsNaN(radius) ||
			math.IsInf(radius, 0) ||
			radius <= 0 {
			return places.NearbyParams{}, badRequest(
				"invalid_radius",
				"radius_m must be a finite number greater than zero",
			)
		}

		if radius > places.MaxNearbyRadiusM {
			return places.NearbyParams{}, badRequest(
				"radius_too_large",
				"radius_m must not exceed 50000 metres",
			)
		}

		params.RadiusM = radius
	}

	category, err := singleValue(
		values,
		"category",
		false,
	)
	if err != nil {
		return places.NearbyParams{}, err
	}

	if category != "" {
		if len(category) > 64 ||
			!categorySlugPattern.MatchString(category) {
			return places.NearbyParams{}, badRequest(
				"invalid_category",
				"category must be a valid lowercase slug",
			)
		}

		params.CategorySlug = category
	}

	limitText, err := singleValue(
		values,
		"limit",
		false,
	)
	if err != nil {
		return places.NearbyParams{}, err
	}

	if limitText != "" {
		limit, err := strconv.ParseInt(
			limitText,
			10,
			32,
		)
		if err != nil ||
			limit < 1 ||
			limit > int64(places.MaxNearbyLimit) {
			return places.NearbyParams{}, badRequest(
				"invalid_limit",
				"limit must be between 1 and 100",
			)
		}

		params.Limit = int32(limit)
	}

	if err := params.Validate(); err != nil {
		return places.NearbyParams{}, badRequest(
			"invalid_search_parameters",
			err.Error(),
		)
	}

	return params, nil
}
