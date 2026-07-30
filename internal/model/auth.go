package model

import "errors"

var (
	ErrInvalidCredential = errors.New("email atau password salah")
	ErrInvalidPassword   = errors.New("password lama salah")
	ErrTokenExpired      = errors.New("refresh token kadaluarsa")
	ErrTokenInvalid      = errors.New("refresh token tidak valid")
)
