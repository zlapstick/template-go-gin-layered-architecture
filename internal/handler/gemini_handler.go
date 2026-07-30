package handler

import (
	"go-gin-layered-architecture/internal/dto"
	"go-gin-layered-architecture/internal/helper"
	"go-gin-layered-architecture/internal/service"
	"io"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ChatbotHandler struct {
	geminiService *service.GeminiService
	logger        *slog.Logger
}

func NewChatbotHandler(geminiService *service.GeminiService, logger *slog.Logger) *ChatbotHandler {
	return &ChatbotHandler{geminiService: geminiService, logger: logger}
}

// @Summary Chat dengan Gemini
// @Description Mengirim prompt ke Gemini dengan dukungan upload dokumen/gambar
// @Tags Chatbot
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Param prompt formData string false "Chat prompt (gunakan untuk multipart/form-data)"
// @Param model formData string false "Model name"
// @Param file formData file false "Dokumen atau gambar (opsional)"
// @Param body body dto.AskGeminiRequest false "Chat request (gunakan untuk application/json)"
// @Success 201 {object} helper.DataResponse
// @Failure 400 {object} helper.ErrorResponse
// @Failure 500 {object} helper.MessageResponse
// @Router /chatbot/ask [post]
func (h *ChatbotHandler) Ask(c *gin.Context) {
	var prompt string
	var model string
	var fileBytes []byte
	var mimeType string

	contentType := c.ContentType()

	if contentType == "application/json" || contentType == "" {
		var req dto.AskGeminiRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, helper.ErrorResponse{
				Message: "Validasi gagal",
				Errors:  helper.ParseValidationError(err),
			})
			return
		}
		prompt = req.Prompt
	} else {
		prompt = c.PostForm("prompt")
		model = c.PostForm("model")

		file, err := c.FormFile("file")
		if err == nil {
			src, err := file.Open()
			if err != nil {
				h.logger.Error("error opening file", "error", err.Error())
				c.JSON(http.StatusInternalServerError, helper.MessageResponse{
					Message: helper.ErrInternalServerError,
				})
				return
			}
			defer src.Close()

			fileBytes, err = io.ReadAll(src)
			if err != nil {
				h.logger.Error("error reading file", "error", err.Error())
				c.JSON(http.StatusInternalServerError, helper.MessageResponse{
					Message: helper.ErrInternalServerError,
				})
				return
			}
			mimeType = file.Header.Get("Content-Type")
		}
	}

	if prompt == "" {
		c.JSON(http.StatusBadRequest, helper.ErrorResponse{
			Message: "Validasi gagal",
			Errors: map[string]string{
				"prompt": "prompt is required",
			},
		})
		return
	}

	geminiResponse, err := h.geminiService.Ask(c.Request.Context(), prompt, fileBytes, mimeType, model)
	if err != nil {
		h.logger.Error("error ask gemini", "error", err.Error())
		c.JSON(http.StatusInternalServerError, helper.MessageResponse{
			Message: helper.ErrInternalServerError,
		})
		return
	}

	c.JSON(http.StatusCreated, helper.DataResponse{
		Message: "Respons berhasil dibuat",
		Data:    geminiResponse,
	})
}

// @Summary Generate Image dengan Gemini
// @Description Generate gambar berdasarkan prompt dengan dukungan upload dokumen/gambar opsional
// @Tags Chatbot
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Param prompt formData string false "Image generation prompt (gunakan untuk multipart/form-data)"
// @Param model formData string false "Model name"
// @Param file formData file false "Dokumen atau gambar untuk referensi (opsional)"
// @Param request body dto.GenerateImageRequest false "Image Generation Request (gunakan untuk application/json)"
// @Success 201 {object} helper.DataResponse
// @Failure 400 {object} helper.ErrorResponse
// @Failure 500 {object} helper.MessageResponse
// @Router /chatbot/generate-image [post]
func (h *ChatbotHandler) GenerateImage(c *gin.Context) {
	var prompt string
	var model string
	var fileBytes []byte
	var mimeType string

	contentType := c.ContentType()

	if contentType == "application/json" || contentType == "" {
		var req dto.GenerateImageRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, helper.ErrorResponse{
				Message: "Validasi gagal",
				Errors:  helper.ParseValidationError(err),
			})
			return
		}
		prompt = req.Prompt
		model = req.Model
	} else {
		prompt = c.PostForm("prompt")
		model = c.PostForm("model")

		file, err := c.FormFile("file")
		if err == nil {
			src, err := file.Open()
			if err != nil {
				h.logger.Error("error opening file", "error", err.Error())
				c.JSON(http.StatusInternalServerError, helper.MessageResponse{
					Message: helper.ErrInternalServerError,
				})
				return
			}
			defer src.Close()

			fileBytes, err = io.ReadAll(src)
			if err != nil {
				h.logger.Error("error reading file", "error", err.Error())
				c.JSON(http.StatusInternalServerError, helper.MessageResponse{
					Message: helper.ErrInternalServerError,
				})
				return
			}
			mimeType = file.Header.Get("Content-Type")
		}
	}

	if prompt == "" {
		c.JSON(http.StatusBadRequest, helper.ErrorResponse{
			Message: "Validasi gagal",
			Errors: map[string]string{
				"prompt": "prompt is required",
			},
		})
		return
	}

	imageURL, err := h.geminiService.GenerateImage(c.Request.Context(), prompt, fileBytes, mimeType, model)
	if err != nil {
		h.logger.Error("error generating image", "error", err.Error())
		c.JSON(http.StatusInternalServerError, helper.MessageResponse{
			Message: helper.ErrInternalServerError,
		})
		return
	}

	c.JSON(http.StatusCreated, helper.DataResponse{
		Message: "Gambar berhasil dibuat",
		Data: dto.GenerateImageResponse{
			ImageURL: imageURL,
		},
	})
}
