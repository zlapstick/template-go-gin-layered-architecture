package main

import (
	"fmt"
	"os"
	"strings"

	"go-gin-layered-architecture/internal/database"
)

const (
	dir    = "internal/database/migrations"
	table  = "migrations"
	column = "migration"
)

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		usage()
		return
	}

	switch args[0] {
	case "create":
		if len(args) < 2 {
			fmt.Println("Usage: go run cmd/migration/main.go create <name>")
			return
		}
		create(args[1])
	case "migrate":
		migrate()
	case "down":
		down()
	case "drop":
		db := database.Open(table, column)
		defer db.Close()
		database.DropAll(db)
	default:
		usage()
	}
}

func usage() {
	fmt.Println("Usage: go run cmd/migration/main.go <migrate|down|create|drop>")
}

func create(name string) {
	migration := database.NextName(dir, name)

	for _, kind := range []string{"up", "down"} {
		file := fmt.Sprintf("%s.%s.sql", migration, kind)
		database.WriteFile(dir, file, fmt.Sprintf("-- %s migration\n", strings.ToUpper(kind)))
	}
}

func migrate() {
	db := database.Open(table, column)
	defer db.Close()

	applied := database.Applied(db, table, column)

	for _, file := range database.Files(dir, ".up.sql") {
		name := strings.TrimSuffix(file, ".up.sql")
		if applied[name] {
			fmt.Printf("Skipped: %s (already executed)\n", name)
			continue
		}

		database.ExecFile(db, dir, file)
		database.Record(db, table, column, name)
		fmt.Printf("Migrated: %s\n", name)
	}
}

func down() {
	db := database.Open(table, column)
	defer db.Close()

	last, ok := database.Last(db, table, column)
	if !ok {
		fmt.Println("No migration to rollback")
		return
	}

	database.ExecFile(db, dir, last+".down.sql")
	database.Forget(db, table, column, last)
	fmt.Printf("Rolled back: %s\n", last)
}
