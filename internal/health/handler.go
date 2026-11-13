package health

import (
	"log"
	"net/http"
)

func Handler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	_, err := w.Write([]byte("{\"status\":\"ok\",\"service\":\"geomind-api\"}\n"))
	if err != nil {
		log.Printf("failed to write health response: %v", err)
	}
}
