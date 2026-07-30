package service

import (
	"context"
	"errors"
	"go-gin-layered-architecture/internal/model"
	"go-gin-layered-architecture/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) List(ctx context.Context, filter model.UserListFilter) ([]model.UserSummary, model.UserListFilter, int, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PerPage < 1 {
		filter.PerPage = 10
	}
	if filter.PerPage > 100 {
		filter.PerPage = 100
	}

	data, total, err := s.repo.List(ctx, filter)
	if err != nil {
		return []model.UserSummary{}, filter, 0, err
	}

	return data, filter, total, nil
}

func (s *UserService) GetById(ctx context.Context, id string) (model.UserSummary, error) {
	user, err := s.repo.FindById(ctx, id)
	if err != nil {
		return model.UserSummary{}, err
	}
	return model.UserSummary{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	}, nil
}

func (s *UserService) Create(ctx context.Context, user *model.User) error {
	user.Role = "user"

	existing, err := s.repo.FindByEmail(ctx, user.Email)
	if err != nil && !errors.Is(err, model.ErrUserNotFound) {
		return err
	}
	if existing.ID != "" {
		return model.ErrEmailAlreadyExists
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hashed)

	return s.repo.Create(ctx, user)
}

func (s *UserService) Update(ctx context.Context, id string, user *model.User) error {
	if user.Email != "" {
		owner, err := s.repo.FindByEmail(ctx, user.Email)
		if err != nil && !errors.Is(err, model.ErrUserNotFound) {
			return err
		}
		if owner.ID != "" && owner.ID != id {
			return model.ErrEmailAlreadyExists
		}
	}

	if user.Password != "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		user.Password = string(hashed)
	}

	return s.repo.Update(ctx, id, user)
}

func (s *UserService) Delete(ctx context.Context, id string) error {
	if _, err := s.repo.FindById(ctx, id); err != nil {
		return err
	}
	return s.repo.Delete(ctx, id)
}
