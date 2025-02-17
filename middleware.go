package main

import (
	"github.com/justinas/alice"
	"log"
	"mime"
	"net/http"
	"net/http/httputil"
)

var (
	mid alice.Chain = alice.New(enforceJSONHandler, responseLogger)
)

// NOTE: middleware was really more of an "I want to do this"
// addition to the project since it would otherwise have been something I
// know how to do.. don't worry about this too much. There's coverage here!

// guarantees content-type, not really necessary but I wanted to do it
func enforceJSONHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contentType := r.Header.Get("Content-Type")

		if contentType != "" {
			mt, _, err := mime.ParseMediaType(contentType)
			if err != nil {
				http.Error(w, "Malformed Content-Type header", http.StatusBadRequest)
				return
			}

			if mt != "application/json" {
				http.Error(w, "Content-Type header must be application/json", http.StatusUnsupportedMediaType)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

// useful for diagnostics when testing/developing, also not really necessary
func responseLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		reqDump, err := httputil.DumpRequest(req, true)
		if err != nil {
			log.Fatalf("ERROR: %v", err)
		}
		log.Printf("REQ: %s", string(reqDump))
		next.ServeHTTP(w, req)
	})
}
