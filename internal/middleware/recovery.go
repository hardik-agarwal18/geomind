package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"

	"geomind/internal/httpresponse"
)

func Recovery(logger *slog.Logger) Middleware {
	if logger == nil {
		logger = slog.Default()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			recorder := &statusWriter{
				ResponseWriter: w,
			}

			defer func() {
				recovered := recover()
				if recovered == nil {
					return
				}

				logger.Error(
					"panic_recovered",
					slog.String("request_id", RequestIDFromContext(r.Context())),
					slog.String("panic_type", fmt.Sprintf("%T", recovered)),
					slog.String("stack", string(debug.Stack())),
				)

				if recorder.WroteFinalResponse() {
					return
				}

				if err := httpresponse.WriteError(
					recorder,
					http.StatusInternalServerError,
					"internal_error",
					"Internal server error",
				); err != nil {
					logger.Error(
						"failed to write panic recovery response",
						"error", err,
					)
				}
			}()

			next.ServeHTTP(recorder, r)
		})
	}
}
