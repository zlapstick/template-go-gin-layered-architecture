package config

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/lib/pq"
)

func DBLoad(dsn string) *sql.DB {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("gagal terhubung ke database: %v", err)
	}

	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(20)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		log.Fatalf("gagal terhubung ke database: %v", err)
	}

	return db
}
