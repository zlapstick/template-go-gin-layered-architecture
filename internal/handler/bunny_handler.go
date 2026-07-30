package handler

import (
	"errors"
	"go-gin-layered-architecture/internal/dto"
	"go-gin-layered-architecture/internal/helper"
	"go-gin-layered-architecture/internal/model"
	"go-gin-layered-architecture/internal/service"
	"io"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

type BunnyHandler struct {
	bunnyService *service.BunnyService
	logger       *slog.Logger
}

func NewBunnyHandler(bunnyService *service.BunnyService, logger *slog.Logger) *BunnyHandler {
	return &BunnyHandler{bunnyService: bunnyService, logger: logger}
}

// @Summary Upload file ke Bunny Storage
// @Description Mengunggah file ke Bunny Storage. Folder bersifat dinamis dan boleh bertingkat, misalnya "users/avatars"
// @Tags Bunny
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "File yang akan diunggah"
// @Param folder formData string false "Folder tujuan, boleh bertingkat (contoh: users/avatars)"
// @Param file_name formData string false "Nama file tujuan, default memakai nama file asli"
// @Success 201 {object} helper.DataResponse
// @Failure 400 {object} helper.ErrorResponse
// @Failure 500 {object} helper.MessageResponse
// @Router /files [post]
func (h *BunnyHandler) UploadFile(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, helper.ErrorResponse{
			Message: "Validasi gagal",
			Errors: map[string]string{
				"file": "file wajib diisi",
			},
		})
		return
	}

	src, err := file.Open()
	if err != nil {
		h.logger.Error("error opening file", "error", err.Error())
		c.JSON(http.StatusInternalServerError, helper.MessageResponse{
			Message: helper.ErrInternalServerError,
		})
		return
	}
	defer src.Close()

	fileBytes, err := io.ReadAll(src)
	if err != nil {
		h.logger.Error("error reading file", "error", err.Error())
		c.JSON(http.StatusInternalServerError, helper.MessageResponse{
			Message: helper.ErrInternalServerError,
		})
		return
	}

	folder := c.PostForm("folder")

	fileName := c.PostForm("file_name")
	if fileName == "" {
		fileName = file.Filename
	}

	contentType := file.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	fileURL, err := h.bunnyService.UploadFile(c.Request.Context(), folder, fileName, fileBytes, contentType)
	if err != nil {
		h.logger.Error("error uploading file", "error", err.Error(), "folder", folder, "file_name", fileName)
		c.JSON(http.StatusInternalServerError, helper.MessageResponse{
			Message: helper.ErrInternalServerError,
		})
		return
	}

	c.JSON(http.StatusCreated, helper.DataResponse{
		Message: "File berhasil diunggah",
		Data: dto.UploadFileResponse{
			FileURL: fileURL,
		},
	})
}

// @Summary Hapus file dari Bunny Storage
// @Description Menghapus file dari Bunny Storage. Folder bersifat dinamis, atau kirim path lengkap maupun URL CDN penuh pada file_name
// @Tags Bunny
// @Accept json
// @Produce json
// @Param request body dto.DeleteFileRequest true "Data file yang akan dihapus"
// @Success 200 {object} helper.MessageResponse
// @Failure 400 {object} helper.ErrorResponse
// @Failure 404 {object} helper.MessageResponse
// @Failure 500 {object} helper.MessageResponse
// @Router /files [delete]
func (h *BunnyHandler) DeleteFile(c *gin.Context) {
	var req dto.DeleteFileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, helper.ErrorResponse{
			Message: "Validasi gagal",
			Errors:  helper.ParseValidationError(err),
		})
		return
	}

	if err := h.bunnyService.DeleteFile(c.Request.Context(), req.Folder, req.FileName); err != nil {
		switch {
		case errors.Is(err, model.ErrFileNotFound):
			c.JSON(http.StatusNotFound, helper.MessageResponse{
				Message: model.ErrFileNotFound.Error(),
			})
		default:
			h.logger.Error("error deleting file", "error", err.Error(), "folder", req.Folder, "file_name", req.FileName)
			c.JSON(http.StatusInternalServerError, helper.MessageResponse{
				Message: helper.ErrInternalServerError,
			})
		}
		return
	}

	c.JSON(http.StatusOK, helper.MessageResponse{
		Message: "File berhasil dihapus",
	})
}
