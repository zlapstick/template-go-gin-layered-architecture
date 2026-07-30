package model

import (
	"errors"
	"time"
)

var (
	ErrEmailAlreadyExists = errors.New("email sudah digunakan")
	ErrUserNotFound       = errors.New("user tidak ditemukan")
)

type User struct {
	ID        string
	Name      string
	Email     string
	Password  string
	Role      string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

type UserSummary struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type UserListFilter struct {
	SearchGlobal  string
	Name          string
	Email         string
	Page          int
	PerPage       int
	SortBy        string
	SortDirection string
}
