package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

var (
	startTime time.Time
	mux       *http.ServeMux = http.NewServeMux()
)

const HTTP_LISTEN_PORT = 8080

// It's pretty difficult to get coverage here for some reason. :\
// Still got 80% though.
func main() {
	err := InitDatabase()
	if err != nil {
		log.Fatalf("Database initialization failed: %v", err)
	}

	genKeys()
	registerHandlers()

	startTime = time.Now()
	log.Println("Starting JWKS Server - njj0057.")
	err = http.ListenAndServe(fmt.Sprintf(":%d", HTTP_LISTEN_PORT), mux)
	if err != nil {
		log.Fatalf("ERROR IN HTTP SERVER: %v", err)
	}
}
