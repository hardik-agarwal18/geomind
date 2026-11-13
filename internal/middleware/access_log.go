package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

func AccessLog(logger *slog.Logger) Middleware {
	if logger == nil {
		logger = slog.Default()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			recorder := &statusWriter{
				ResponseWriter: w,
			}

			defer func() {
				logger.Info(
					"http_request",
					slog.String("request_id", RequestIDFromContext(r.Context())),
					slog.String("method", r.Method),
					slog.String("path", r.URL.Path),
					slog.Int("status", recorder.StatusCode()),
					slog.Int64("bytes", recorder.bytes),
					slog.Duration("duration", time.Since(start)),
				)
			}()

			next.ServeHTTP(recorder, r)
		})
	}
}
