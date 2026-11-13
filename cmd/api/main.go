package main

import (
	"log"
	"net/http"

	"geomind/internal/server"
)

func main() {
	addr := ":8080"

	handler := server.New(log.Default())

	log.Printf("GeoMind API listening on %s", addr)

	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatal(err)
	}
}
