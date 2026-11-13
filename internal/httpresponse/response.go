package httpresponse

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// WriteJSON serializes a value and writes a JSON HTTP response.
func WriteJSON(w http.ResponseWriter, status int, payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal JSON response: %w", err)
	}

	body = append(body, '\n')

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	if _, err := w.Write(body); err != nil {
		return fmt.Errorf("write JSON response: %w", err)
	}

	return nil
}

// WriteError writes a consistent JSON error response.
func WriteError(
	w http.ResponseWriter,
	status int,
	code string,
	message string,
) error {
	payload := ErrorResponse{
		Error: ErrorDetail{
			Code:    code,
			Message: message,
		},
	}

	return WriteJSON(w, status, payload)
}

// WriteMethodNotAllowed rejects unsupported HTTP methods.
func WriteMethodNotAllowed(w http.ResponseWriter) error {
	w.Header().Set("Allow", http.MethodGet)

	return WriteError(
		w,
		http.StatusMethodNotAllowed,
		"method_not_allowed",
		"Only GET is supported",
	)
}
