package helper

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"go-gin-layered-architecture/internal/model"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	AccessTokenTTL  = 1 * time.Hour
	RefreshTokenTTL = 7 * 24 * time.Hour
	Issuer          = "go-gin-layered-architecture"
)

var (
	ErrTokenGeneration = errors.New("gagal membuat token")
)

type JWTClaims struct {
	UserID string `json:"user_id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func GenerateAccessToken(u *model.User, jwtSecret string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, JWTClaims{
		UserID: u.ID,
		Name:   u.Name,
		Email:  u.Email,
		Role:   u.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(AccessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    Issuer,
		},
	})
	tokenStr, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", ErrTokenGeneration
	}
	return tokenStr, nil
}

type RefreshTokenClaims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

func GenerateRefreshToken(userID string, jwtSecret string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, RefreshTokenClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(RefreshTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    Issuer,
		},
	})
	tokenStr, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", ErrTokenGeneration
	}
	return tokenStr, nil
}

func ValidateRefreshToken(refreshToken string, jwtSecret string) (*RefreshTokenClaims, error) {
	var claims RefreshTokenClaims
	token, err := jwt.ParseWithClaims(refreshToken, &claims, func(t *jwt.Token) (any, error) {
		return []byte(jwtSecret), nil
	}, jwt.WithValidMethods([]string{"HS256"}), jwt.WithIssuer(Issuer))
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, model.ErrTokenExpired
		}
		return nil, model.ErrTokenInvalid
	}
	if !token.Valid || claims.UserID == "" {
		return nil, model.ErrTokenInvalid
	}
	return &claims, nil
}

func HashRefreshToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return fmt.Sprintf("%x", hash)
}
