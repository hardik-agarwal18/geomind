package middleware

import (
	"log"
	"net/http"
	"runtime/debug"

	"geomind/internal/httpresponse"
)

func Recovery(logger *log.Logger) Middleware {
	if logger == nil {
		logger = log.Default()
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

				// Record panic type and stack, but not the panic value.
				logger.Printf(
					"panic_recovered request_id=%s panic_type=%T stack=%s",
					RequestIDFromContext(r.Context()),
					recovered,
					debug.Stack(),
				)

				// A committed response cannot be replaced with a 500.
				if recorder.WroteFinalResponse() {
					return
				}

				if err := httpresponse.WriteError(
					recorder,
					http.StatusInternalServerError,
					"internal_error",
					"Internal server error",
				); err != nil {
					logger.Printf(
						"failed to write panic recovery response: %v",
						err,
					)
				}
			}()

			next.ServeHTTP(recorder, r)
		})
	}
}
