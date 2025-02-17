package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRegisterHandlers(t *testing.T) {
	// Reset mux before registering handlers
	mux = http.NewServeMux()

	// Call function to register handlers
	registerHandlers()

	// Define test cases for expected routes
	tests := []struct {
		method         string
		endpoint       string
		expectedStatus int
	}{
		{"GET", "/", http.StatusOK},
		{"GET", "/.well-known/jwks.json", http.StatusOK},
		{"POST", "/auth", http.StatusOK},
		{"GET", "/invalid", http.StatusNotFound}, // Ensure 404 for unknown routes
	}

	for _, test := range tests {
		req := httptest.NewRequest(test.method, test.endpoint, nil)
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		if rec.Code != test.expectedStatus {
			t.Errorf("Expected status %d for %s, but got %d", test.expectedStatus, test.endpoint, rec.Code)
		}
	}
}

func TestGenKeys(t *testing.T) {
	keys = nil // Reset keys
	genKeys()

	var hasExpired, hasUnexpired bool
	for _, key := range keys {
		if key.ExpiresAt.Before(time.Now()) {
			hasExpired = true
		} else {
			hasUnexpired = true
		}
	}

	if !hasExpired || !hasUnexpired {
		t.Errorf("ensureKeys did not generate both expired and unexpired keys")
	}
}

func TestMethodNotAllowedHandler(t *testing.T) {
	req, _ := http.NewRequest("PUT", "/jwks", nil)
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(jwksHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected 415/MethodNotAllowed, got %d", rr.Code)
	}
}

func TestIndex(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(index)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected 200/OK, got %d", rr.Code)
	}

	req, _ = http.NewRequest("GET", "/someshitidk", nil)
	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(index)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected 404/NotFound, got %d", rr.Code)
	}

	req, _ = http.NewRequest("POST", "/", nil)
	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(index)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected 415/MethodNotAllowed, got %d", rr.Code)
	}
}

func TestJWKSHandler(t *testing.T) {
	req, _ := http.NewRequest("GET", "/jwks", nil)
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(jwksHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rr.Code)
	}

	var response JWKS
	json.Unmarshal(rr.Body.Bytes(), &response)
	if len(response.Keys) == 0 {
		t.Errorf("Expected at least one key in JWKS response")
	}
}

func TestAuthHandler(t *testing.T) {
	req, _ := http.NewRequest("POST", "/auth?expired=true", nil)
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(authHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rr.Code)
	}

	req, _ = http.NewRequest("GET", "/auth", nil)
	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(authHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected 415/MethodNotAllowed, got %d", rr.Code)
	}

	var response map[string]string
	json.Unmarshal(rr.Body.Bytes(), &response)
	if _, exists := response["token"]; !exists {
		t.Errorf("Expected a token in response")
	}
}
