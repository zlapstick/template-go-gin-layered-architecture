package service

import (
	"context"
	"errors"
	"go-gin-layered-architecture/internal/dto"
	"go-gin-layered-architecture/internal/helper"
	"go-gin-layered-architecture/internal/model"
	"go-gin-layered-architecture/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo                *repository.UserRepository
	revokedRefreshTokenRepo *repository.RevokedRefreshTokenRepository
	jwtSecret               string
}

func NewAuthService(userRepo *repository.UserRepository, revokedRefreshTokenRepo *repository.RevokedRefreshTokenRepository, jwtSecret string) *AuthService {
	return &AuthService{userRepo: userRepo, revokedRefreshTokenRepo: revokedRefreshTokenRepo, jwtSecret: jwtSecret}
}

func (s *AuthService) Login(ctx context.Context, email string, password string) (string, string, error) {
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, model.ErrUserNotFound) {
			return "", "", model.ErrInvalidCredential
		}
		return "", "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", "", model.ErrInvalidCredential
	}

	accessToken, err := helper.GenerateAccessToken(&user, s.jwtSecret)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := helper.GenerateRefreshToken(user.ID, s.jwtSecret)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (s *AuthService) Register(ctx context.Context, req *dto.RegisterRequest) (string, string, error) {
	existing, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil && !errors.Is(err, model.ErrUserNotFound) {
		return "", "", err
	}

	if existing.ID != "" {
		return "", "", model.ErrEmailAlreadyExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", "", err
	}

	user := &model.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hashedPassword),
		Role:     "user",
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return "", "", err
	}

	accessToken, err := helper.GenerateAccessToken(user, s.jwtSecret)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := helper.GenerateRefreshToken(user.ID, s.jwtSecret)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (s *AuthService) Me(ctx context.Context, userID string) (model.UserSummary, error) {
	user, err := s.userRepo.FindById(ctx, userID)
	if err != nil {
		return model.UserSummary{}, err
	}

	return model.UserSummary{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	}, nil
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (string, error) {
	claims, err := helper.ValidateRefreshToken(refreshToken, s.jwtSecret)
	if err != nil {
		return "", err
	}

	revoked, err := s.revokedRefreshTokenRepo.Exists(ctx, helper.HashRefreshToken(refreshToken))
	if err != nil {
		return "", err
	}
	if revoked {
		return "", model.ErrTokenInvalid
	}

	user, err := s.userRepo.FindById(ctx, claims.UserID)
	if err != nil {
		return "", err
	}

	newAccessToken, err := helper.GenerateAccessToken(&user, s.jwtSecret)
	if err != nil {
		return "", err
	}

	return newAccessToken, nil
}

func (s *AuthService) Logout(ctx context.Context, userID string, refreshToken string) error {
	claims, err := helper.ValidateRefreshToken(refreshToken, s.jwtSecret)
	if err != nil {
		return err
	}

	if claims.UserID != userID {
		return model.ErrTokenInvalid
	}

	return s.revokedRefreshTokenRepo.Create(ctx, helper.HashRefreshToken(refreshToken))
}

func (s *AuthService) UpdateProfile(ctx context.Context, userID string, name string, email string) (model.UserSummary, error) {
	if email != "" {
		owner, err := s.userRepo.FindByEmail(ctx, email)
		if err != nil && !errors.Is(err, model.ErrUserNotFound) {
			return model.UserSummary{}, err
		}
		if owner.ID != "" && owner.ID != userID {
			return model.UserSummary{}, model.ErrEmailAlreadyExists
		}
	}

	user := model.User{Name: name, Email: email}

	if err := s.userRepo.Update(ctx, userID, &user); err != nil {
		return model.UserSummary{}, err
	}

	return model.UserSummary{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	}, nil
}

func (s *AuthService) ChangePassword(ctx context.Context, userID string, currentPassword string, newPassword string) error {
	user, err := s.userRepo.FindById(ctx, userID)
	if err != nil {
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(currentPassword)); err != nil {
		return model.ErrInvalidPassword
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return s.userRepo.Update(ctx, userID, &model.User{Password: string(hashedPassword)})
}
