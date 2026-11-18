package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"geomind/internal/geo"
	"geomind/internal/httpresponse"
	"geomind/internal/places"
)

const nearestQueryTimeout = 2 * time.Second

type NearestPlacesHandler struct {
	service      *places.NearestService
	logger       *slog.Logger
	queryTimeout time.Duration
}

func NewNearestPlacesHandler(
	service *places.NearestService,
	logger *slog.Logger,
) *NearestPlacesHandler {
	if logger == nil {
		logger = slog.Default()
	}

	return &NearestPlacesHandler{
		service:      service,
		logger:       logger,
		queryTimeout: nearestQueryTimeout,
	}
}

type nearestPlaceResponse struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Category     string  `json:"category"`
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
	DistanceM    float64 `json:"distance_m"`
	KNNDistanceM float64 `json:"knn_distance_m"`
}

type nearestMetaResponse struct {
	Count     int     `json:"count"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	K         int32   `json:"k"`
	Category  *string `json:"category"`
}

type nearestSearchResponse struct {
	Data []nearestPlaceResponse `json:"data"`
	Meta nearestMetaResponse    `json:"meta"`
}

func (h *NearestPlacesHandler) ServeHTTP(
	w http.ResponseWriter,
	r *http.Request,
) {
	params, err := parseNearestQuery(r.URL.RawQuery)
	if err != nil {
		var requestErr *requestError

		if errors.As(err, &requestErr) {
			h.writeError(
				w,
				http.StatusBadRequest,
				requestErr.code,
				requestErr.message,
			)
		} else {
			h.writeError(
				w,
				http.StatusBadRequest,
				"invalid_query",
				"Invalid query string",
			)
		}
		return
	}

	queryCtx, cancel := context.WithTimeout(
		r.Context(),
		h.queryTimeout,
	)
	defer cancel()

	results, err := h.service.Search(queryCtx, params)
	if err != nil {
		if r.Context().Err() != nil {
			return
		}

		if errors.Is(err, context.DeadlineExceeded) {
			h.writeError(
				w,
				http.StatusGatewayTimeout,
				"search_timeout",
				"Nearest search exceeded its time limit",
			)
			return
		}

		h.logger.Error(
			"nearest_search_failed",
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

	response := buildNearestResponse(params, results)

	if err := httpresponse.WriteJSON(
		w,
		http.StatusOK,
		response,
	); err != nil {
		h.logger.Error(
			"nearest_response_write_failed",
			"error", err,
		)
	}
}

func parseNearestQuery(
	rawQuery string,
) (places.NearestParams, error) {
	values, err := url.ParseQuery(rawQuery)
	if err != nil {
		return places.NearestParams{}, err
	}

	allowed := map[string]bool{
		"latitude":  true,
		"longitude": true,
		"k":         true,
		"category":  true,
	}

	for key := range values {
		if !allowed[key] {
			return places.NearestParams{}, badRequest(
				"unknown_parameter",
				"Unsupported query parameter: "+key,
			)
		}
	}

	latitudeText, err := singleValue(values, "latitude", true)
	if err != nil {
		return places.NearestParams{}, err
	}

	longitudeText, err := singleValue(values, "longitude", true)
	if err != nil {
		return places.NearestParams{}, err
	}

	latitude, err := strconv.ParseFloat(latitudeText, 64)
	if err != nil {
		return places.NearestParams{}, badRequest(
			"invalid_coordinate",
			"Invalid latitude",
		)
	}

	longitude, err := strconv.ParseFloat(longitudeText, 64)
	if err != nil {
		return places.NearestParams{}, badRequest(
			"invalid_coordinate",
			"Invalid longitude",
		)
	}

	params := places.NearestParams{
		Origin: geo.Coordinate{
			Longitude: longitude,
			Latitude:  latitude,
		},
		K: places.DefaultNearestK,
	}

	kText, err := singleValue(values, "k", false)
	if err != nil {
		return places.NearestParams{}, err
	}

	if kText != "" {
		k, err := strconv.ParseInt(kText, 10, 32)
		if err != nil {
			return places.NearestParams{}, badRequest(
				"invalid_k",
				"k must be an integer between 1 and 100",
			)
		}

		params.K = int32(k)
	}

	category, err := singleValue(values, "category", false)
	if err != nil {
		return places.NearestParams{}, err
	}

	params.CategorySlug = category

	if err := params.Validate(); err != nil {
		return places.NearestParams{}, badRequest(
			"invalid_search_parameters",
			err.Error(),
		)
	}

	return params, nil
}

func buildNearestResponse(
	params places.NearestParams,
	results []places.NearestPlace,
) nearestSearchResponse {
	data := make([]nearestPlaceResponse, 0, len(results))

	for _, place := range results {
		data = append(data, nearestPlaceResponse{
			ID:           place.ID.String(),
			Name:         place.Name,
			Category:     place.Category,
			Latitude:     place.Latitude,
			Longitude:    place.Longitude,
			DistanceM:    place.DistanceM,
			KNNDistanceM: place.KNNDistanceM,
		})
	}

	var category *string
	if params.CategorySlug != "" {
		value := params.CategorySlug
		category = &value
	}

	return nearestSearchResponse{
		Data: data,
		Meta: nearestMetaResponse{
			Count:     len(data),
			Latitude:  params.Origin.Latitude,
			Longitude: params.Origin.Longitude,
			K:         params.K,
			Category:  category,
		},
	}
}

func (h *NearestPlacesHandler) writeError(
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
			"nearest_error_response_failed",
			"error", err,
		)
	}
}

// Ensure the response model remains JSON serializable.
var _ = json.Valid
