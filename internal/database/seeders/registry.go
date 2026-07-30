package seeders

import "database/sql"

type Seeder struct {
	Name string
	Up   func(db *sql.DB) error
	Down func(db *sql.DB) error
}

var All []Seeder

func Register(name string, up, down func(db *sql.DB) error) {
	All = append(All, Seeder{Name: name, Up: up, Down: down})
}
