package main

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestEnforceJSONHandler(t *testing.T) {
	tests := []struct {
		name           string
		contentType    string
		expectedStatus int
	}{
		{"Valid JSON Content-Type", "application/json", http.StatusOK},
		{"Missing Content-Type", "", http.StatusOK},
		{"Invalid Content-Type", "text/plain", http.StatusUnsupportedMediaType},
		{"Malformed Content-Type", "application/json; invalid", http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/", bytes.NewBuffer([]byte(`{"key": "value"}`)))
			req.Header.Set("Content-Type", tt.contentType)

			rr := httptest.NewRecorder()
			handler := enforceJSONHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}
		})
	}
}

func TestResponseLogger(t *testing.T) {
	var loggedOutput strings.Builder
	log.SetOutput(&loggedOutput) // Capture log output

	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	handler := responseLogger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	handler.ServeHTTP(rr, req)

	if !strings.Contains(loggedOutput.String(), "REQ:") {
		t.Error("Expected request to be logged but did not find log entry")
	}

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}
}
