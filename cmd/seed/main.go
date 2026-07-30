package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"go-gin-layered-architecture/internal/database"
	"go-gin-layered-architecture/internal/database/seeders"
)

const (
	dir    = "internal/database/seeders"
	table  = "seeders"
	column = "seeder"
)

const template = `package seeders

import "database/sql"

func init() {
	Register("%s", Seed%s, Rollback%s)
}

func Seed%s(db *sql.DB) error {
	// TODO: implement seeder
	return nil
}

func Rollback%s(db *sql.DB) error {
	// TODO: implement rollback
	return nil
}
`

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		usage()
		return
	}

	switch args[0] {
	case "create":
		if len(args) < 2 {
			fmt.Println("Usage: go run cmd/seed/main.go create <name>")
			return
		}
		create(args[1])
	case "seed":
		seed()
	case "down":
		down()
	default:
		usage()
	}
}

func usage() {
	fmt.Println("Usage: go run cmd/seed/main.go <seed|down|create>")
}

func create(name string) {
	seeder := database.NextName(dir, name)

	// 000001_create_users -> 000001CreateUsers
	parts := strings.Split(seeder, "_")
	for i, p := range parts {
		if p != "" {
			parts[i] = strings.ToUpper(p[:1]) + p[1:]
		}
	}
	fn := strings.Join(parts, "")

	content := fmt.Sprintf(template, seeder, fn, fn, fn, fn)
	database.WriteFile(dir, seeder+".go", content)
}

func seed() {
	db := database.Open(table, column)
	defer db.Close()

	applied := database.Applied(db, table, column)

	for _, s := range seeders.All {
		if applied[s.Name] {
			fmt.Printf("Skipped: %s (already executed)\n", s.Name)
			continue
		}

		if err := s.Up(db); err != nil {
			log.Fatalf("gagal menjalankan %s: %v", s.Name, err)
		}
		database.Record(db, table, column, s.Name)
		fmt.Printf("Seeded: %s\n", s.Name)
	}
}

func down() {
	db := database.Open(table, column)
	defer db.Close()

	last, ok := database.Last(db, table, column)
	if !ok {
		fmt.Println("No seeder to rollback")
		return
	}

	for _, s := range seeders.All {
		if s.Name != last {
			continue
		}

		if err := s.Down(db); err != nil {
			log.Fatalf("gagal rollback %s: %v", last, err)
		}
		database.Forget(db, table, column, last)
		fmt.Printf("Rolled back: %s\n", last)
		return
	}

	log.Fatalf("seeder %s tidak ditemukan", last)
}
