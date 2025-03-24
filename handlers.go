package main

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"

	"bytes"
	"errors"
	"log"
	"net/http"
	"sync"
	"time"

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
	if req.URL.Path != "/" {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("404 page not found\n"))
		return
	}
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

func encodePrivateKeyToPEM(privateKey *rsa.PrivateKey) string {
	// Convert RSA private key to PKCS1 ASN.1 DER format
	privDER := x509.MarshalPKCS1PrivateKey(privateKey)

	// Encode DER data to PEM format
	privBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privDER,
	}

	var pemBuffer bytes.Buffer
	if err := pem.Encode(&pemBuffer, privBlock); err != nil {
		panic("Failed to encode private key to PEM") // Shouldn't happen in normal execution
	}

	return pemBuffer.String()
}

func decodePEMToPrivateKey(pemStr string) (*rsa.PrivateKey, error) {
	// Decode PEM block
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		return nil, errors.New("failed to decode PEM block containing RSA private key")
	}

	// Parse the DER-encoded private key
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return privateKey, nil
}

func fetchJWKSFromDB() (*JWKS, error) {
	rows, err := db.Query("SELECT key, kid, exp FROM keys WHERE exp > ?", time.Now().Unix())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	jwks := JWKS{}

	for rows.Next() {
		var keyPEM string
		var kid string
		var exp int64

		if err := rows.Scan(&keyPEM, &kid, &exp); err != nil {
			return nil, err
		}

		// Decode PEM-encoded private key
		privateKey, err := decodePEMToPrivateKey(keyPEM)
		if err != nil {
			log.Printf("Skipping key %s due to decoding error: %v", kid, err)
			continue
		}

		// Extract the public key
		pubKey := privateKey.Public().(*rsa.PublicKey)

		// Add public key details to JWKS response
		jwks.Keys = append(jwks.Keys, map[string]interface{}{
			"kty": "RSA",
			"kid": kid,
			"exp": exp,
			"alg": "RS256",
			"n":   base64.RawURLEncoding.EncodeToString(pubKey.N.Bytes()),
			"e":   "AQAB",
		})
	}

	return &jwks, nil
}

// handles jwks route
func jwksHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		methodNotAllowedHandler(w, req)
		return
	}

	keysLock.RLock()
	defer keysLock.RUnlock()

	jwks, err := fetchJWKSFromDB()
	if err != nil {
		http.Error(w, "Failed to fetch JWKS", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(jwks)
	if err != nil {
		log.Fatal("Unable to encode jwks json.")
	}
}

// Handles auth route.
func authHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		methodNotAllowedHandler(w, req)
		return
	}

	// guarantees expired and unexpired key will exist when route hit
	// normally you wouldn't do it this way obviously, but this is easy
	genKeys()
	expired := req.URL.Query().Get("expired") == "true"
	signingKey, err := GetKey(expired)
	if err != nil {
		http.Error(w, "No valid signing key found when one should exist.",
			http.StatusInternalServerError)
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
