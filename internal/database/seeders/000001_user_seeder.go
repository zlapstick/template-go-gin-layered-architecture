package seeders

import (
	"database/sql"

	"golang.org/x/crypto/bcrypt"
)

func init() {
	Register("000001_user_seeder", SeedUsers, RollbackUsers)
}

func SeedUsers(db *sql.DB) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("12345678"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	query := "INSERT INTO users (name, email, password, role) VALUES ($1, $2, $3, $4)"

	users := []struct {
		Name  string
		Email string
		Role  string
	}{
		{"fardan", "fardanm.azizi11@gmail.com", "superadmin"},
		{"arsil", "superadmin@gmail.com", "superadmin"},
		{"admin", "admin@gmail.com", "admin"},
	}

	for _, u := range users {
		if _, err := db.Exec(query, u.Name, u.Email, hashedPassword, u.Role); err != nil {
			return err
		}
	}
	return nil
}

func RollbackUsers(db *sql.DB) error {
	_, err := db.Exec("DELETE FROM users WHERE email IN ($1, $2, $3)", "fardanm.azizi11@gmail.com", "superadmin@gmail.com", "admin@gmail.com")
	return err
}
