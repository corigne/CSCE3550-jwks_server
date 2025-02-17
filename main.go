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
	genKeys()
	registerHandlers()

	startTime = time.Now()
	fmt.Printf("Starting JWKS server at: %v...\n", startTime.Local().Format(time.RFC3339))
	err := http.ListenAndServe(fmt.Sprintf(":%d", HTTP_LISTEN_PORT), mux)
	if err != nil {
		log.Fatalf("ERROR IN HTTP SERVER: %v", err)
	}
}
