package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"go-gin-layered-architecture/internal/model"
	"strings"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) List(ctx context.Context, filter model.UserListFilter) ([]model.UserSummary, int, error) {
	where := []string{"deleted_at IS NULL"}
	args := []any{}
	n := 1

	if filter.SearchGlobal != "" {
		like := "%" + filter.SearchGlobal + "%"
		where = append(where, fmt.Sprintf("(id::text ILIKE $%d OR name ILIKE $%d OR email ILIKE $%d)", n, n+1, n+2))
		args = append(args, like, like, like)
		n += 3
	}
	if filter.Name != "" {
		where = append(where, fmt.Sprintf("name ILIKE $%d", n))
		args = append(args, "%"+filter.Name+"%")
		n++
	}
	if filter.Email != "" {
		where = append(where, fmt.Sprintf("email ILIKE $%d", n))
		args = append(args, "%"+filter.Email+"%")
		n++
	}

	whereClause := "WHERE " + strings.Join(where, " AND ")

	var total int
	countQuery := "SELECT COUNT(*) FROM users " + whereClause
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count user: %w", err)
	}

	if total == 0 {
		return []model.UserSummary{}, 0, nil
	}

	sortBy := "name"
	switch strings.ToLower(filter.SortBy) {
	case "id", "name", "email":
		sortBy = strings.ToLower(filter.SortBy)
	}
	sortDir := "DESC"
	if strings.EqualFold(filter.SortDirection, "asc") {
		sortDir = "ASC"
	}

	query := fmt.Sprintf(
		"SELECT id, name, email FROM users %s ORDER BY %s %s LIMIT $%d OFFSET $%d",
		whereClause, sortBy, sortDir, n, n+1,
	)
	dataArgs := make([]any, 0, len(args)+2)
	dataArgs = append(dataArgs, args...)
	dataArgs = append(dataArgs, filter.PerPage, (filter.Page-1)*filter.PerPage)

	rows, err := r.db.QueryContext(ctx, query, dataArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("query user: %w", err)
	}
	defer rows.Close()

	summaries := []model.UserSummary{}
	for rows.Next() {
		var u model.UserSummary
		if err := rows.Scan(&u.ID, &u.Name, &u.Email); err != nil {
			return nil, 0, fmt.Errorf("scan user: %w", err)
		}
		summaries = append(summaries, u)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterasi user: %w", err)
	}
	return summaries, total, nil
}

func (r *UserRepository) GetById(ctx context.Context, id string) (model.UserSummary, error) {
	var u model.UserSummary
	query := "SELECT id, name, email FROM users WHERE id = $1 AND deleted_at IS NULL"
	err := r.db.QueryRowContext(ctx, query, id).Scan(&u.ID, &u.Name, &u.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return u, model.ErrUserNotFound
		}
		return u, fmt.Errorf("query user: %w", err)
	}
	return u, nil
}

func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	query := "INSERT INTO users (name, email, password, role) VALUES ($1, $2, $3, $4) RETURNING id"

	if err := r.db.QueryRowContext(ctx, query, user.Name, user.Email, user.Password, user.Role).Scan(&user.ID); err != nil {
		return fmt.Errorf("insert user: %w", err)
	}
	return nil
}

func (r *UserRepository) Update(ctx context.Context, id string, user *model.User) error {
	sets := []string{"updated_at = NOW()"}
	args := []any{}
	n := 1

	if user.Email != "" {
		sets = append(sets, fmt.Sprintf("email = $%d", n))
		args = append(args, user.Email)
		n++
	}
	if user.Name != "" {
		sets = append(sets, fmt.Sprintf("name = $%d", n))
		args = append(args, user.Name)
		n++
	}
	if user.Password != "" {
		sets = append(sets, fmt.Sprintf("password = $%d", n))
		args = append(args, user.Password)
		n++
	}

	query := fmt.Sprintf(
		"UPDATE users SET %s WHERE id = $%d AND deleted_at IS NULL RETURNING id, name, email",
		strings.Join(sets, ", "), n,
	)
	args = append(args, id)

	err := r.db.QueryRowContext(ctx, query, args...).Scan(&user.ID, &user.Name, &user.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.ErrUserNotFound
		}
		return fmt.Errorf("update user: %w", err)
	}
	return nil
}

func (r *UserRepository) Delete(ctx context.Context, id string) error {
	query := "UPDATE users SET deleted_at = NOW(), updated_at = NOW(), email = email || '--trashed--' || id WHERE id = $1 AND deleted_at IS NULL"
	if _, err := r.db.ExecContext(ctx, query, id); err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	return nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (model.User, error) {
	var user model.User
	query := "SELECT id, name, email, password, role FROM users WHERE email = $1 AND deleted_at IS NULL"
	err := r.db.QueryRowContext(ctx, query, email).Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.Role)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return user, model.ErrUserNotFound
		}
		return user, fmt.Errorf("query user: %w", err)
	}
	return user, nil
}

func (r *UserRepository) FindById(ctx context.Context, id string) (model.User, error) {
	var user model.User
	query := "SELECT id, name, email, password, role FROM users WHERE id = $1 AND deleted_at IS NULL"
	err := r.db.QueryRowContext(ctx, query, id).Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.Role)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return user, model.ErrUserNotFound
		}
		return user, fmt.Errorf("query user: %w", err)
	}
	return user, nil
}
