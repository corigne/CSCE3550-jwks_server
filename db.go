package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const dbFile = "totally_not_my_privateKeys.db"

var db *sql.DB

func InitDatabase() error {
	var err error
	db, err = sql.Open("sqlite3", dbFile)
	if err != nil {
		return fmt.Errorf("Failed to open database: %v", err)
	}
	createTableSQL := `CREATE TABLE IF NOT EXISTS keys(
		kid INTEGER PRIMARY KEY AUTOINCREMENT,
		key BLOB NOT NULL,
		exp INTEGER NOT NULL
	);`
	_, err = db.Exec(createTableSQL)
	if err != nil {
		return fmt.Errorf("Failed to create table: %v", err)
	}
	log.Println("Initialized database connection.")
	return nil
}

func InsertKey(privateKey string, expiry int64) error {
	_, err := db.Exec("INSERT INTO keys (key, exp) VALUES (?, ?)", privateKey, expiry)
	if err != nil {
		return fmt.Errorf("Failed to insert key: %v", err)
	}
	return nil
}

func GetKey(expired bool) (*Key, error) {
	var keyPEM string
	var kid string
	var exp int64

	var query string
	if expired {
		query = "SELECT key, kid, exp FROM keys WHERE exp <= ? ORDER BY exp DESC LIMIT 1"
	} else {
		query = "SELECT key, kid, exp FROM keys WHERE exp > ? ORDER BY exp ASC LIMIT 1"
	}

	err := db.QueryRow(query, time.Now().Unix()).Scan(&keyPEM, &kid, &exp)
	if err != nil {
		return nil, err
	}

	key, err := decodePEMToPrivateKey(keyPEM)
	if err != nil {
		return nil, err
	}

	return &Key{
		PrivateKey: key,
		Kid:        kid,
		ExpiresAt:  time.Unix(exp, 0),
	}, nil
}
