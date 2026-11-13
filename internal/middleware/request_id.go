package middleware

import (
	"context"
	"crypto/rand"
	"fmt"
	"log/slog"
	"net/http"

	"geomind/internal/httpresponse"
)

const RequestIDHeader = "X-Request-ID"

type requestIDKey struct{}

// RequestIDFromContext retrieves the request ID from the context.
func RequestIDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey{}).(string)
	return id
}

func RequestID(logger *slog.Logger) Middleware {
	if logger == nil {
		logger = slog.Default()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id, err := newRequestID()
			if err != nil {
				logger.Error("request ID generation failed", "error", err)

				if writeErr := httpresponse.WriteError(
					w,
					http.StatusInternalServerError,
					"internal_error",
					"Internal server error",
				); writeErr != nil {
					logger.Error("failed to write error response", "error", writeErr)
				}

				return
			}

			ctx := context.WithValue(r.Context(), requestIDKey{}, id)

			w.Header().Set(RequestIDHeader, id)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func newRequestID() (string, error) {
	var id [16]byte

	if _, err := rand.Read(id[:]); err != nil {
		return "", err
	}

	// Set UUID version 4 and variant bits.
	id[6] = (id[6] & 0x0f) | 0x40
	id[8] = (id[8] & 0x3f) | 0x80

	return fmt.Sprintf(
		"%x-%x-%x-%x-%x",
		id[0:4],
		id[4:6],
		id[6:8],
		id[8:10],
		id[10:16],
	), nil
}
