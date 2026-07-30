package repository

import (
	"context"
	"database/sql"
	"fmt"
)

type RevokedRefreshTokenRepository struct {
	db *sql.DB
}

func NewRevokedRefreshTokenRepository(db *sql.DB) *RevokedRefreshTokenRepository {
	return &RevokedRefreshTokenRepository{db: db}
}

func (r *RevokedRefreshTokenRepository) Create(ctx context.Context, tokenHash string) error {
	query := "INSERT INTO revoked_refresh_tokens (token_hash) VALUES ($1) ON CONFLICT (token_hash) DO NOTHING"
	if _, err := r.db.ExecContext(ctx, query, tokenHash); err != nil {
		return fmt.Errorf("insert revoked refresh token: %w", err)
	}
	return nil
}

func (r *RevokedRefreshTokenRepository) Exists(ctx context.Context, tokenHash string) (bool, error) {
	var exists bool
	query := "SELECT EXISTS (SELECT 1 FROM revoked_refresh_tokens WHERE token_hash = $1)"
	if err := r.db.QueryRowContext(ctx, query, tokenHash).Scan(&exists); err != nil {
		return false, fmt.Errorf("query revoked refresh token: %w", err)
	}
	return exists, nil
}
