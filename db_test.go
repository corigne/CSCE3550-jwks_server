package main

import (
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func setupTestDB(t *testing.T) {
	var err error
	db, err = sql.Open("sqlite3", ":memory:") // Use in-memory DB for testing
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	createTableSQL := `CREATE TABLE keys(
		kid INTEGER PRIMARY KEY AUTOINCREMENT,
		key BLOB NOT NULL,
		exp INTEGER NOT NULL
	);`
	_, err = db.Exec(createTableSQL)
	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}
}

func TestGenKeys(t *testing.T) {
	setupTestDB(t)
	genKeys()

	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM keys").Scan(&count)
	if err != nil {
		t.Errorf("Failed to query database: %v", err)
	}
	if count != 2 {
		t.Errorf("Expected 2 keys (expired and unexpired), got %d", count)
	}

	// Validate one expired and one unexpired key exists
	expiredKey, err := GetKey(true)
	if err != nil{
		t.Errorf("Expected an expired key, but got err: %v", err)
	}
	if !expiredKey.ExpiresAt.Before(time.Now()) {
		t.Errorf("Expected an expired key, but got a valid key")
	}

	unexpiredKey, err := GetKey(false)
	if err != nil || unexpiredKey == nil || unexpiredKey.ExpiresAt.Before(time.Now()) {
		t.Errorf("Expected an unexpired key, but got none or incorrect expiry")
	}
}
func TestInitDatabase(t *testing.T) {
	err := InitDatabase()
	if err != nil {
		t.Errorf("InitDatabase failed: %v", err)
	}
}

func TestInsertKey(t *testing.T) {
	setupTestDB(t)

	key := "test-key"
	expiry := time.Now().Unix() + 3600

	err := InsertKey(key, expiry)
	if err != nil {
		t.Errorf("InsertKey failed: %v", err)
	}

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM keys").Scan(&count)
	if err != nil {
		t.Errorf("Failed to query test database: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected 1 key in database, got %d", count)
	}
}

func TestGetKey(t *testing.T) {
	// Initialize the database
	setupTestDB(t)

	// Generate and insert both an expired and an unexpired key
	genKeys()

	// Retrieve an unexpired key
	key, err := GetKey(false)
	if err != nil {
		t.Errorf("GetKey (unexpired) failed: %v", err)
	} else if key.ExpiresAt.Before(time.Now()) {
		t.Errorf("Expected an unexpired key, but got an expired one")
	}

	// Retrieve an expired key
	key, err = GetKey(true)
	if err != nil {
		t.Errorf("GetKey (expired) failed: %v", err)
	} else if key.ExpiresAt.After(time.Now()) {
		t.Errorf("Expected an expired key, but got an unexpired one")
	}
}
