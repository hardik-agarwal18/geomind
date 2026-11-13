package health

import (
	"log"
	"net/http"

	"geomind/internal/httpresponse"
)

type StatusResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
}

func Handler(w http.ResponseWriter, r *http.Request) {
	if !requireGET(w, r) {
		return
	}

	response := StatusResponse{
		Status:  "ok",
		Service: "geomind-api",
	}

	if err := httpresponse.WriteJSON(
		w,
		http.StatusOK,
		response,
	); err != nil {
		log.Printf("health response failed: %v", err)
	}
}

func ReadyHandler(w http.ResponseWriter, r *http.Request) {
	if !requireGET(w, r) {
		return
	}

	response := StatusResponse{
		Status:  "ready",
		Service: "geomind-api",
	}

	if err := httpresponse.WriteJSON(
		w,
		http.StatusOK,
		response,
	); err != nil {
		log.Printf("readiness response failed: %v", err)
	}
}

func requireGET(w http.ResponseWriter, r *http.Request) bool {
	if r.Method == http.MethodGet || r.Method == http.MethodHead {
		return true
	}

	if err := httpresponse.WriteMethodNotAllowed(w); err != nil {
		log.Printf("failed to write method error response: %v", err)
	}

	return false
}
