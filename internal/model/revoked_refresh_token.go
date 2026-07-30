package model

import "time"

type RevokedRefreshToken struct {
	TokenHash string
	RevokedAt time.Time
}
