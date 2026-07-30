package handler

import (
	"errors"
	"go-gin-layered-architecture/internal/dto"
	"go-gin-layered-architecture/internal/helper"
	"go-gin-layered-architecture/internal/model"
	"go-gin-layered-architecture/internal/service"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService *service.UserService
	logger      *slog.Logger
}

func NewUserHandler(userService *service.UserService, logger *slog.Logger) *UserHandler {
	return &UserHandler{userService: userService, logger: logger}
}

// ListUsers godoc
//
// @Summary List Users
// @Description Get paginated list of users with search, filter, and sort
// @Tags Users
// @Produce json
// @Param search_global query string false "Search across id, name, and email"
// @Param name query string false "Filter by name"
// @Param email query string false "Filter by email"
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page (max 100)" default(10)
// @Param sort_by query string false "Sort column" Enums(id, name, email)
// @Param sort_dir query string false "Sort direction" Enums(asc, desc)
// @Success 200 {object} helper.PaginatedResponse{data=[]model.UserSummary} "OK"
// @Failure 500 {object} helper.MessageResponse "Internal Server Error"
// @Router /api/users [get]
func (h *UserHandler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.Query("page"))
	perPage, _ := strconv.Atoi(c.Query("per_page"))

	result, used, total, err := h.userService.List(c.Request.Context(), model.UserListFilter{
		SearchGlobal:  c.Query("search_global"),
		Name:          c.Query("name"),
		Email:         c.Query("email"),
		Page:          page,
		PerPage:       perPage,
		SortBy:        c.Query("sort_by"),
		SortDirection: c.Query("sort_dir"),
	})
	if err != nil {
		h.logger.Error("error list users", "error", err.Error())
		c.JSON(http.StatusInternalServerError, helper.MessageResponse{
			Message: helper.ErrInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, helper.PaginatedResponse{
		Message: "Daftar user berhasil diambil",
		Data:    result,
		Meta:    helper.NewMeta(used.Page, used.PerPage, total, len(result)),
	})
}

// GetUserById godoc
//
// @Summary Get User by ID
// @Description Get user by ID
// @Tags Users
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} helper.DataResponse{data=model.UserSummary} "OK"
// @Failure 404 {object} helper.MessageResponse "Not Found"
// @Router /api/users/{id} [get]
func (h *UserHandler) GetById(c *gin.Context) {
	id := c.Param("id")
	result, err := h.userService.GetById(c.Request.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, model.ErrUserNotFound):
			c.JSON(http.StatusNotFound, helper.MessageResponse{
				Message: err.Error(),
			})
		default:
			h.logger.Error("error get user by id", "error", err.Error())
			c.JSON(http.StatusInternalServerError, helper.MessageResponse{
				Message: helper.ErrInternalServerError,
			})
		}
		return
	}

	c.JSON(http.StatusOK, helper.DataResponse{
		Message: "User berhasil diambil",
		Data:    result,
	})
}

// CreateUser godoc
//
// @Summary Create User
// @Description Create new user
// @Tags Users
// @Produce json
// @Param user body dto.CreateUserRequest true "User data"
// @Success 201 {object} helper.DataResponse{data=model.UserSummary} "OK"
// @Failure 400 {object} helper.ErrorResponse "Bad Request"
// @Failure 409 {object} helper.ErrorResponse "Conflict"
// @Router /api/users [post]
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, helper.ErrorResponse{
			Message: "Validasi gagal",
			Errors:  helper.ParseValidationError(err),
		})
		return
	}

	user := model.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	}

	if err := h.userService.Create(c.Request.Context(), &user); err != nil {
		switch {
		case errors.Is(err, model.ErrEmailAlreadyExists):
			c.JSON(http.StatusConflict, helper.ErrorResponse{
				Message: "Email sudah terdaftar",
				Errors:  map[string]string{"email": "Email sudah terdaftar"},
			})
		default:
			h.logger.Error("error create user", "error", err.Error())
			c.JSON(http.StatusInternalServerError, helper.MessageResponse{
				Message: helper.ErrInternalServerError,
			})
		}
		return
	}

	c.JSON(http.StatusCreated, helper.DataResponse{
		Message: "User berhasil dibuat",
		Data:    user,
	})
}

// UpdateUser godoc
//
// @Summary Update User
// @Description Update existing user by ID
// @Tags Users
// @Produce json
// @Param id path string true "User ID"
// @Param user body dto.UpdateUserRequest true "User data"
// @Success 200 {object} helper.DataResponse{data=model.UserSummary} "OK"
// @Failure 400 {object} helper.ErrorResponse "Bad Request"
// @Failure 404 {object} helper.MessageResponse "Not Found"
// @Failure 409 {object} helper.ErrorResponse "Conflict"
// @Router /api/users/{id} [put]
func (h *UserHandler) UpdateUser(c *gin.Context) {
	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, helper.ErrorResponse{
			Message: "Validasi gagal",
			Errors:  helper.ParseValidationError(err),
		})
		return
	}

	user := model.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	}

	err := h.userService.Update(c.Request.Context(), c.Param("id"), &user)
	if err != nil {
		switch {
		case errors.Is(err, model.ErrUserNotFound):
			c.JSON(http.StatusNotFound, helper.MessageResponse{
				Message: err.Error(),
			})
		case errors.Is(err, model.ErrEmailAlreadyExists):
			c.JSON(http.StatusConflict, helper.ErrorResponse{
				Message: "Validasi error",
				Errors:  map[string]string{"email": "email sudah terdaftar"},
			})
		default:
			h.logger.Error("error update user", "error", err.Error())
			c.JSON(http.StatusInternalServerError, helper.MessageResponse{
				Message: helper.ErrInternalServerError,
			})
		}
		return
	}

	c.JSON(http.StatusOK, helper.DataResponse{
		Message: "User berhasil diupdate",
		Data:    user,
	})
}

// DeleteUser godoc
//
// @Summary Delete User
// @Description Delete user by ID
// @Tags Users
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} helper.MessageResponse "OK"
// @Failure 404 {object} helper.MessageResponse "Not Found"
// @Router /api/users/{id} [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	if err := h.userService.Delete(c.Request.Context(), c.Param("id")); err != nil {
		switch {
		case errors.Is(err, model.ErrUserNotFound):
			c.JSON(http.StatusNotFound, helper.MessageResponse{
				Message: err.Error(),
			})
		default:
			h.logger.Error("error delete user", "error", err.Error())
			c.JSON(http.StatusInternalServerError, helper.MessageResponse{
				Message: helper.ErrInternalServerError,
			})
		}
		return
	}

	c.JSON(http.StatusOK, helper.MessageResponse{
		Message: "User berhasil dihapus",
	})
}
