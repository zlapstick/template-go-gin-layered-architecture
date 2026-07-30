package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"go-gin-layered-architecture/internal/model"

	"github.com/lib/pq"
)

type ThirdPartyKeyRepository struct {
	db *sql.DB
}

func NewThirdPartyKeyRepository(db *sql.DB) *ThirdPartyKeyRepository {
	return &ThirdPartyKeyRepository{db: db}
}

func (r *ThirdPartyKeyRepository) FindNameByKey(ctx context.Context, key string, names []string) (string, error) {
	var name string
	query := "SELECT name FROM third_party_keys WHERE key = $1 AND name = ANY($2) AND deleted_at IS NULL"
	err := r.db.QueryRowContext(ctx, query, key, pq.Array(names)).Scan(&name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", model.ErrThirdPartyKeyNotFound
		}
		return "", fmt.Errorf("query third party key: %w", err)
	}
	return name, nil
}
