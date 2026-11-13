package middleware

import (
	"log"
	"net/http"
	"time"
)

func AccessLog(logger *log.Logger) Middleware {
	if logger == nil {
		logger = log.Default()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			recorder := &statusWriter{
				ResponseWriter: w,
			}

			defer func() {
				duration := time.Since(start)

				logger.Printf(
					"http_request request_id=%s method=%q path=%q status=%d bytes=%d duration=%s",
					RequestIDFromContext(r.Context()),
					r.Method,
					r.URL.Path,
					recorder.StatusCode(),
					recorder.bytes,
					duration,
				)
			}()

			next.ServeHTTP(recorder, r)
		})
	}
}
