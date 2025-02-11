package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	_ "crypto/rand"
	_ "crypto/rsa"
	_ "crypto/x509"

	_ "github.com/golang-jwt/jwt/v5"
)

func registerHandlers() {
	indexHandler := http.HandlerFunc(index)

	mux.Handle("/", mid.Then(indexHandler))
}

func index(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	resp := make(map[string]string)
	resp["message"] = "welcome to the njj0057 JWKS Server - STATUS: OK"
	resp["start_time"] = startTime.Local().Format(time.RFC3339)
	resp["uptime"] = time.Now().Sub(startTime).String()

	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Fatalf("Error in JSON Marshal. Err: %s", err)
	}

	w.Write(jsonResp)
}
