package database

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"go-gin-layered-architecture/internal/config"

	"github.com/joho/godotenv"
)

func Open(table, column string) *sql.DB {
	_ = godotenv.Load()

	db := config.DBLoad(config.AppLoad().DBDsn)

	ddl := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
  id SERIAL PRIMARY KEY,
  %s VARCHAR(255) NOT NULL UNIQUE,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);`, table, column)

	if _, err := db.Exec(ddl); err != nil {
		log.Fatalf("gagal membuat tabel %s: %v", table, err)
	}
	return db
}

func Applied(db *sql.DB, table, column string) map[string]bool {
	rows, err := db.Query(fmt.Sprintf("SELECT %s FROM %s", column, table))
	if err != nil {
		log.Fatalf("gagal membaca riwayat %s: %v", table, err)
	}
	defer rows.Close()

	done := map[string]bool{}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			log.Fatalf("gagal membaca riwayat %s: %v", table, err)
		}
		done[name] = true
	}
	if err := rows.Err(); err != nil {
		log.Fatalf("gagal membaca riwayat %s: %v", table, err)
	}
	return done
}

func Last(db *sql.DB, table, column string) (string, bool) {
	query := fmt.Sprintf("SELECT %s FROM %s ORDER BY id DESC LIMIT 1", column, table)

	var name string
	switch err := db.QueryRow(query).Scan(&name); {
	case err == nil:
		return name, true
	case errors.Is(err, sql.ErrNoRows):
		return "", false
	default:
		log.Fatalf("gagal membaca catatan terakhir %s: %v", table, err)
		return "", false
	}
}

func Record(db *sql.DB, table, column, name string) {
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES ($1)", table, column)
	if _, err := db.Exec(query, name); err != nil {
		log.Fatalf("gagal mencatat %s: %v", name, err)
	}
}

func Forget(db *sql.DB, table, column, name string) {
	query := fmt.Sprintf("DELETE FROM %s WHERE %s = $1", table, column)
	if _, err := db.Exec(query, name); err != nil {
		log.Fatalf("gagal menghapus catatan %s: %v", name, err)
	}
}

func ExecFile(db *sql.DB, dir, name string) {
	path := filepath.Join(dir, name)

	content, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("gagal membaca %s: %v", path, err)
	}
	if _, err := db.Exec(string(content)); err != nil {
		log.Fatalf("gagal menjalankan %s: %v", name, err)
	}
}

func DropAll(db *sql.DB) {
	rows, err := db.Query("SELECT tablename FROM pg_tables WHERE schemaname = 'public'")
	if err != nil {
		log.Fatalf("gagal membaca daftar tabel: %v", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			log.Fatalf("gagal scan tabel: %v", err)
		}
		tables = append(tables, table)
	}
	if err := rows.Err(); err != nil {
		log.Fatalf("gagal membaca daftar tabel: %v", err)
	}

	if len(tables) == 0 {
		fmt.Println("Database sudah kosong")
		return
	}

	for _, table := range tables {
		if _, err := db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", table)); err != nil {
			log.Fatalf("gagal drop tabel %s: %v", table, err)
		}
		fmt.Printf("Dropped: %s\n", table)
	}
	fmt.Println("Semua tabel berhasil dihapus")
}

func Files(dir, suffix string) []string {
	entries, _ := os.ReadDir(dir)

	var names []string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), suffix) {
			names = append(names, e.Name())
		}
	}
	return names
}

func NextName(dir, name string) string {
	entries, _ := os.ReadDir(dir)

	highest := 0
	for _, e := range entries {
		if n, _ := strconv.Atoi(strings.SplitN(e.Name(), "_", 2)[0]); n > highest {
			highest = n
		}
	}
	return fmt.Sprintf("%06d_%s", highest+1, name)
}

func WriteFile(dir, name, content string) {
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		log.Fatalf("gagal membuat %s: %v", path, err)
	}
	fmt.Printf("Created: %s\n", path)
}
