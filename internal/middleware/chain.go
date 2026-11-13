package middleware

import "net/http"

type Middleware func(http.Handler) http.Handler

// Chain applies middleware from left to right during request processing.
// Responses unwind through middleware in reverse order.
func Chain(handler http.Handler, middlewares ...Middleware) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}

	return handler
}
