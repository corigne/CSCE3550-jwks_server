package main

import (
	"fmt"
	"net/http"
	"time"
)

var (
	startTime time.Time
	mux       *http.ServeMux = http.NewServeMux()
)

const HTTP_LISTEN_PORT = 8080

func main() {
	genKeys()
	registerHandlers()
	startTime = time.Now()
	fmt.Printf("Starting JWKS server at: %v...\n", startTime.Local().Format(time.RFC3339))
	http.ListenAndServe(fmt.Sprintf(":%d", HTTP_LISTEN_PORT), mux)
}
