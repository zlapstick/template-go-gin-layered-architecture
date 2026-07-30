package middleware

import (
	"errors"
	"go-gin-layered-architecture/internal/helper"
	"go-gin-layered-architecture/internal/model"
	"go-gin-layered-architecture/internal/repository"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const (
	authPayloadKey    = "auth_payload"
	thirdPartyNameKey = "third_party_name"

	thirdPartyPrefix = "third_party:"
)

type AuthMiddleware struct {
	jwtSecret      string
	thirdPartyRepo *repository.ThirdPartyKeyRepository
	logger         *slog.Logger
}

func NewAuthMiddleware(jwtSecret string, thirdPartyRepo *repository.ThirdPartyKeyRepository, logger *slog.Logger) *AuthMiddleware {
	return &AuthMiddleware{jwtSecret: jwtSecret, thirdPartyRepo: thirdPartyRepo, logger: logger}
}

func (m *AuthMiddleware) RequiredAuth(methods ...string) gin.HandlerFunc {
	jwtAllowed := len(methods) == 0
	var thirdPartyNames []string
	for _, method := range methods {
		if names, ok := strings.CutPrefix(method, thirdPartyPrefix); ok {
			for _, name := range strings.Split(names, ",") {
				if name != "" {
					thirdPartyNames = append(thirdPartyNames, name)
				}
			}
			continue
		}
		if method == "jwt" {
			jwtAllowed = true
		}
	}

	return func(c *gin.Context) {
		tokenStr := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
		if tokenStr == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, helper.MessageResponse{
				Message: "Token tidak ditemukan",
			})
			return
		}

		if jwtAllowed {
			var claims helper.JWTClaims
			token, err := jwt.ParseWithClaims(tokenStr, &claims, func(t *jwt.Token) (any, error) {
				return []byte(m.jwtSecret), nil
			}, jwt.WithValidMethods([]string{"HS256"}), jwt.WithIssuer(helper.Issuer))
			if err == nil && token.Valid && claims.UserID != "" && claims.Role != "" {
				c.Set(authPayloadKey, claims)
				c.Next()
				return
			}
		}

		if len(thirdPartyNames) > 0 && m.thirdPartyRepo != nil {
			name, err := m.thirdPartyRepo.FindNameByKey(c.Request.Context(), tokenStr, thirdPartyNames)
			switch {
			case err == nil:
				c.Set(thirdPartyNameKey, name)
				c.Next()
				m.logger.Info("akses third party",
					"name", name,
					"method", c.Request.Method,
					"path", c.Request.URL.Path,
					"status", c.Writer.Status(),
				)
				return
			case !errors.Is(err, model.ErrThirdPartyKeyNotFound):
				c.AbortWithStatusJSON(http.StatusInternalServerError, helper.MessageResponse{
					Message: helper.ErrInternalServerError,
				})
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusUnauthorized, helper.MessageResponse{
			Message: "Token tidak valid",
		})
	}
}

func GetAuthPayload(c *gin.Context) (helper.JWTClaims, error) {
	value, ok := c.Get(authPayloadKey)
	if !ok {
		return helper.JWTClaims{}, errors.New("auth payload tidak ditemukan di context")
	}
	claims, ok := value.(helper.JWTClaims)
	if !ok {
		return helper.JWTClaims{}, errors.New("auth payload bukan JWT claims")
	}
	return claims, nil
}

func GetThirdPartyName(c *gin.Context) (string, error) {
	value, ok := c.Get(thirdPartyNameKey)
	if !ok {
		return "", errors.New("third party name tidak ditemukan di context")
	}
	name, ok := value.(string)
	if !ok {
		return "", errors.New("third party name bukan string")
	}
	return name, nil
}
