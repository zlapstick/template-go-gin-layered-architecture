package handler

import (
	"errors"
	"go-gin-layered-architecture/internal/dto"
	"go-gin-layered-architecture/internal/helper"
	"go-gin-layered-architecture/internal/middleware"
	"go-gin-layered-architecture/internal/model"
	"go-gin-layered-architecture/internal/service"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *service.AuthService
	logger      *slog.Logger
}

func NewAuthHandler(authService *service.AuthService, logger *slog.Logger) *AuthHandler {
	return &AuthHandler{authService: authService, logger: logger}
}

// Login godoc
//
// @Summary Login
// @Description Authenticate user with email and password to get access and refresh tokens
// @Tags Auth
// @Produce json
// @Param credentials body dto.LoginRequest true "User credentials"
// @Success 200 {object} helper.DataResponse{data=dto.LoginResponse} "OK"
// @Failure 400 {object} helper.ErrorResponse "Bad Request"
// @Failure 401 {object} helper.ErrorResponse "Unauthorized"
// @Router /api/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, helper.ErrorResponse{
			Message: "Validasi gagal",
			Errors:  helper.ParseValidationError(err),
		})
		return
	}

	accessToken, refreshToken, err := h.authService.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, model.ErrInvalidCredential):
			c.JSON(http.StatusUnauthorized, helper.ErrorResponse{
				Message: "Email atau password salah",
				Errors:  map[string]string{"email": "Email atau password salah", "password": "Email atau password salah"},
			})
		default:
			h.logger.Error("error login", "error", err.Error())
			c.JSON(http.StatusInternalServerError, helper.MessageResponse{
				Message: helper.ErrInternalServerError,
			})
		}
		return
	}

	c.JSON(http.StatusOK, helper.DataResponse{
		Message: "Login berhasil",
		Data: dto.LoginResponse{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		},
	})
}

// Register godoc
//
// @Summary Register
// @Description Register a new user account and get access and refresh tokens
// @Tags Auth
// @Produce json
// @Param account body dto.RegisterRequest true "New user account"
// @Success 201 {object} helper.DataResponse{data=dto.RegisterResponse} "Created"
// @Failure 400 {object} helper.ErrorResponse "Bad Request"
// @Failure 409 {object} helper.ErrorResponse "Conflict"
// @Router /api/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, helper.ErrorResponse{
			Message: "Validasi gagal",
			Errors:  helper.ParseValidationError(err),
		})
		return
	}

	accessToken, refreshToken, err := h.authService.Register(c.Request.Context(), &req)
	if err != nil {
		switch {
		case errors.Is(err, model.ErrEmailAlreadyExists):
			c.JSON(http.StatusConflict, helper.ErrorResponse{
				Message: "Email sudah terdaftar",
				Errors:  map[string]string{"email": "Email sudah terdaftar"},
			})
		default:
			h.logger.Error("error register", "error", err.Error())
			c.JSON(http.StatusInternalServerError, helper.MessageResponse{
				Message: helper.ErrInternalServerError,
			})
		}
		return
	}

	c.JSON(http.StatusCreated, helper.DataResponse{
		Message: "Register berhasil",
		Data: dto.RegisterResponse{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		},
	})
}

// Me godoc
//
// @Summary Me
// @Description Get current user profile information
// @Tags Auth
// @Produce json
// @Success 200 {object} helper.DataResponse{data=model.UserSummary} "OK"
// @Failure 401 {object} helper.MessageResponse "Unauthorized"
// @Router /api/auth/me [get]
// @Security BearerAuth
func (h *AuthHandler) Me(c *gin.Context) {
	payload, err := middleware.GetAuthPayload(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, helper.MessageResponse{
			Message: "Token tidak valid",
		})
		return
	}

	me, err := h.authService.Me(c.Request.Context(), payload.UserID)
	if err != nil {
		switch {
		case errors.Is(err, model.ErrUserNotFound):
			c.JSON(http.StatusUnauthorized, helper.MessageResponse{
				Message: "Token tidak valid",
			})
		default:
			h.logger.Error("error get me", "error", err.Error())
			c.JSON(http.StatusInternalServerError, helper.MessageResponse{
				Message: helper.ErrInternalServerError,
			})
		}
		return
	}

	c.JSON(http.StatusOK, helper.DataResponse{
		Message: "Data user berhasil diambil",
		Data:    me,
	})
}

// Refresh godoc
//
// @Summary Refresh
// @Description Exchange a valid refresh token for a new access token
// @Tags Auth
// @Produce json
// @Param credentials body dto.RefreshTokenRequest true "Refresh token"
// @Success 200 {object} helper.DataResponse{data=dto.RefreshTokenResponse} "OK"
// @Failure 400 {object} helper.ErrorResponse "Bad Request"
// @Failure 401 {object} helper.MessageResponse "Unauthorized"
// @Router /api/auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, helper.ErrorResponse{
			Message: "Validasi gagal",
			Errors:  helper.ParseValidationError(err),
		})
		return
	}

	newAccessToken, err := h.authService.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		switch {
		case errors.Is(err, model.ErrTokenExpired):
			c.JSON(http.StatusUnauthorized, helper.MessageResponse{
				Message: "Refresh token kadaluarsa",
			})
		case errors.Is(err, model.ErrTokenInvalid), errors.Is(err, model.ErrUserNotFound):
			c.JSON(http.StatusUnauthorized, helper.MessageResponse{
				Message: "Refresh token tidak valid",
			})
		default:
			h.logger.Error("error refresh", "error", err.Error())
			c.JSON(http.StatusInternalServerError, helper.MessageResponse{
				Message: helper.ErrInternalServerError,
			})
		}
		return
	}

	c.JSON(http.StatusOK, helper.DataResponse{
		Message: "Refresh token berhasil",
		Data: dto.RefreshTokenResponse{
			AccessToken: newAccessToken,
		},
	})
}

// Logout godoc
//
// @Summary Logout
// @Description Revoke the given refresh token so it can no longer be used
// @Tags Auth
// @Produce json
// @Param credentials body dto.RefreshTokenRequest true "Refresh token to revoke"
// @Success 200 {object} helper.MessageResponse "OK"
// @Failure 400 {object} helper.ErrorResponse "Bad Request"
// @Failure 401 {object} helper.MessageResponse "Unauthorized"
// @Router /api/auth/logout [post]
// @Security BearerAuth
func (h *AuthHandler) Logout(c *gin.Context) {
	payload, err := middleware.GetAuthPayload(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, helper.MessageResponse{
			Message: "Token tidak valid",
		})
		return
	}

	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, helper.ErrorResponse{
			Message: "Validasi gagal",
			Errors:  helper.ParseValidationError(err),
		})
		return
	}

	if err := h.authService.Logout(c.Request.Context(), payload.UserID, req.RefreshToken); err != nil {
		switch {
		case errors.Is(err, model.ErrTokenExpired):
			c.JSON(http.StatusUnauthorized, helper.MessageResponse{
				Message: "Refresh token kadaluarsa",
			})
		case errors.Is(err, model.ErrTokenInvalid):
			c.JSON(http.StatusUnauthorized, helper.MessageResponse{
				Message: "Refresh token tidak valid",
			})
		default:
			h.logger.Error("error logout", "error", err.Error())
			c.JSON(http.StatusInternalServerError, helper.MessageResponse{
				Message: helper.ErrInternalServerError,
			})
		}
		return
	}

	c.JSON(http.StatusOK, helper.MessageResponse{
		Message: "Logout berhasil",
	})
}

// UpdateProfile godoc
//
// @Summary Edit Profile
// @Description Update the name and/or email of the currently authenticated user (at least one required)
// @Tags Auth
// @Produce json
// @Param profile body dto.UpdateProfileRequest true "Profile data"
// @Success 200 {object} helper.DataResponse{data=model.UserSummary} "OK"
// @Failure 400 {object} helper.ErrorResponse "Bad Request"
// @Failure 401 {object} helper.MessageResponse "Unauthorized"
// @Failure 404 {object} helper.MessageResponse "Not Found"
// @Failure 409 {object} helper.ErrorResponse "Conflict"
// @Router /api/auth/profile [put]
// @Security BearerAuth
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	payload, err := middleware.GetAuthPayload(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, helper.MessageResponse{
			Message: "Token tidak valid",
		})
		return
	}

	var req dto.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, helper.ErrorResponse{
			Message: "Validasi gagal",
			Errors:  helper.ParseValidationError(err),
		})
		return
	}

	profile, err := h.authService.UpdateProfile(c.Request.Context(), payload.UserID, req.Name, req.Email)
	if err != nil {
		switch {
		case errors.Is(err, model.ErrUserNotFound):
			c.JSON(http.StatusNotFound, helper.MessageResponse{
				Message: err.Error(),
			})
		case errors.Is(err, model.ErrEmailAlreadyExists):
			c.JSON(http.StatusConflict, helper.ErrorResponse{
				Message: "Email sudah terdaftar",
				Errors:  map[string]string{"email": "Email sudah terdaftar"},
			})
		default:
			h.logger.Error("error update profile", "error", err.Error())
			c.JSON(http.StatusInternalServerError, helper.MessageResponse{
				Message: helper.ErrInternalServerError,
			})
		}
		return
	}

	c.JSON(http.StatusOK, helper.DataResponse{
		Message: "Profil berhasil diubah",
		Data:    profile,
	})
}

// ChangePassword godoc
//
// @Summary Edit Password
// @Description Change the password of the currently authenticated user after verifying the old one
// @Tags Auth
// @Produce json
// @Param password body dto.ChangePasswordRequest true "Password data"
// @Success 200 {object} helper.MessageResponse "OK"
// @Failure 400 {object} helper.ErrorResponse "Bad Request"
// @Failure 401 {object} helper.ErrorResponse "Unauthorized"
// @Failure 404 {object} helper.MessageResponse "Not Found"
// @Router /api/auth/password [put]
// @Security BearerAuth
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	payload, err := middleware.GetAuthPayload(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, helper.MessageResponse{
			Message: "Token tidak valid",
		})
		return
	}

	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, helper.ErrorResponse{
			Message: "Validasi gagal",
			Errors:  helper.ParseValidationError(err),
		})
		return
	}

	if err := h.authService.ChangePassword(c.Request.Context(), payload.UserID, req.CurrentPassword, req.NewPassword); err != nil {
		switch {
		case errors.Is(err, model.ErrInvalidPassword):
			c.JSON(http.StatusUnauthorized, helper.ErrorResponse{
				Message: "Password lama salah",
				Errors:  map[string]string{"current_password": "Password lama salah"},
			})
		case errors.Is(err, model.ErrUserNotFound):
			c.JSON(http.StatusNotFound, helper.MessageResponse{
				Message: err.Error(),
			})
		default:
			h.logger.Error("error change password", "error", err.Error())
			c.JSON(http.StatusInternalServerError, helper.MessageResponse{
				Message: helper.ErrInternalServerError,
			})
		}
		return
	}

	c.JSON(http.StatusOK, helper.MessageResponse{
		Message: "Password berhasil diubah",
	})
}
