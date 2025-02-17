package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"math/big"

	"github.com/golang-jwt/jwt/v5"
)

type Key struct {
	PrivateKey *rsa.PrivateKey
	Kid        string
	ExpiresAt  time.Time
}

type JWKS struct {
	Keys []map[string]interface{} `json:"keys"`
}

var (
	keys     []Key
	keysLock sync.RWMutex
)

func registerHandlers() {
	mux.Handle("/", mid.Then(http.HandlerFunc(index)))
	mux.Handle("/.well-known/jwks.json", mid.Then(http.HandlerFunc(jwksHandler)))
	mux.Handle("/auth", mid.Then(http.HandlerFunc(authHandler)))
}

func methodNotAllowedHandler(w http.ResponseWriter, _ *http.Request) {
	http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
}

func index(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		methodNotAllowedHandler(w, req)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	resp := make(map[string]string)
	resp["message"] = "welcome to the njj0057 JWKS Server - STATUS: OK"
	resp["start_time"] = startTime.Local().Format(time.RFC3339)
	resp["uptime"] = time.Since(startTime).String()

	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Fatalf("Error in JSON Marshal. Err: %s", err)
	}

	_, err = w.Write(jsonResp)
	if err != nil {
		log.Fatal("Unable to write to http response writer.")
	}
}

func genKeys() {
	keysLock.Lock()
	defer keysLock.Unlock()

	var hasExpired, hasUnexpired bool
	for _, key := range keys {
		if key.ExpiresAt.Before(time.Now()) {
			hasExpired = true
		} else {
			hasUnexpired = true
		}
	}

	if !hasExpired {
		privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			log.Fatalf("Failed to generate expired RSA key: %v", err)
		}
		keys = append(keys, Key{
			PrivateKey: privateKey,
			Kid:        "expiredkey",
			ExpiresAt:  time.Now().Add(-5 * time.Minute),
		})
	}

	if !hasUnexpired {
		privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			log.Fatalf("Failed to generate unexpired RSA key: %v", err)
		}
		keys = append(keys, Key{
			PrivateKey: privateKey,
			Kid:        "validkey",
			ExpiresAt:  time.Now().Add(10 * time.Minute),
		})
	}
}

func encodeBase64(b *big.Int) string {
	return base64.RawURLEncoding.EncodeToString(b.Bytes())
}

func jwksHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		methodNotAllowedHandler(w, req)
		return
	}

	keysLock.RLock()
	defer keysLock.RUnlock()

	jwks := JWKS{}
	for _, key := range keys {
		if key.ExpiresAt.After(time.Now()) {
			pubKey := key.PrivateKey.Public().(*rsa.PublicKey)
			jwks.Keys = append(jwks.Keys, map[string]interface{}{
				"kty": "RSA",
				"kid": key.Kid,
				"exp": key.ExpiresAt.Unix(),
				"alg": "RS256",
				"n":   encodeBase64(pubKey.N),
				"e":   "AQAB",
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(jwks)
	if err != nil {
		log.Fatal("Unable to encode jwks json.")
	}
}

func authHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		methodNotAllowedHandler(w, req)
		return
	}
	genKeys()
	expired := req.URL.Query().Get("expired") == "true"

	keysLock.RLock()
	var signingKey *Key
	for _, key := range keys {
		if expired {
			if key.ExpiresAt.Before(time.Now()) {
				signingKey = &key
				break
			}
		} else {
			if key.ExpiresAt.After(time.Now()) {
				signingKey = &key
				break
			}
		}
	}
	keysLock.RUnlock()

	if signingKey == nil {
		http.Error(w, "No valid signing key available", http.StatusInternalServerError)
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"sub": "example_user",
		"iat": time.Now().Unix(),
		"exp": signingKey.ExpiresAt.Unix(),
	})
	token.Header["kid"] = signingKey.Kid

	signedToken, err := token.SignedString(signingKey.PrivateKey)
	if err != nil {
		http.Error(w, "Failed to sign token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(map[string]string{"token": signedToken})
	if err != nil {
		log.Fatal("Unable to encode token json.")
	}
}
