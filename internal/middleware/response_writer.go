package middleware

import "net/http"

type statusWriter struct {
	http.ResponseWriter

	statusCode int
	wroteFinal bool
	bytes      int64
}

func (w *statusWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

func (w *statusWriter) WriteHeader(code int) {
	if w.wroteFinal {
		return
	}

	// Informational responses (except 101) don't commit the final response.
	if code >= 200 || code == http.StatusSwitchingProtocols {
		w.statusCode = code
		w.wroteFinal = true
	}

	w.ResponseWriter.WriteHeader(code)
}

func (w *statusWriter) Write(body []byte) (int, error) {
	if !w.wroteFinal {
		w.WriteHeader(http.StatusOK)
	}

	n, err := w.ResponseWriter.Write(body)
	w.bytes += int64(n)

	return n, err
}

func (w *statusWriter) StatusCode() int {
	if !w.wroteFinal {
		return http.StatusOK
	}

	return w.statusCode
}

func (w *statusWriter) WroteFinalResponse() bool {
	return w.wroteFinal
}
