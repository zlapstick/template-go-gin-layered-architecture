package handler

import (
	"go-gin-layered-architecture/internal/helper"
	"go-gin-layered-architecture/internal/service"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

type LogHandler struct {
	logService *service.LogService
	logger     *slog.Logger
}

func NewLogHandler(logService *service.LogService, logger *slog.Logger) *LogHandler {
	return &LogHandler{
		logService: logService,
		logger:     logger,
	}
}

// @Summary Get All Log Files
// @Description Retrieve list of all log files with their sizes
// @Tags Logs
// @Produce json
// @Success 200 {object} helper.DataResponse{data=[]model.LogFile} "OK"
// @Router /logs [get]
func (h *LogHandler) GetAllLogs(c *gin.Context) {
	files, err := h.logService.GetAllLogFiles()
	if err != nil {
		h.logger.Error("gagal mendapatkan daftar file log", "error", err)
		c.JSON(http.StatusInternalServerError, helper.ErrorResponse{
			Message: helper.ErrInternalServerError,
			Errors:  map[string]string{},
		})
		return
	}

	c.JSON(http.StatusOK, helper.DataResponse{
		Message: "Daftar file log berhasil diambil",
		Data:    files,
	})
}

// @Summary Get Log by File Name
// @Description Retrieve all log entries from a specific log file
// @Tags Logs
// @Param file_name path string true "Log file name (without .log extension)"
// @Produce json
// @Success 200 {object} helper.DataResponse{data=[]model.LogEntry}
// @Router /logs/{file_name} [get]
func (h *LogHandler) GetLogsByFileName(c *gin.Context) {
	fileName := c.Param("file_name")

	entries, err := h.logService.GetLogsByFileName(fileName)
	if err != nil {
		h.logger.Error("gagal mendapatkan log file", "file_name", fileName, "error", err)
		c.JSON(http.StatusNotFound, helper.ErrorResponse{
			Message: "Log file tidak ditemukan",
			Errors:  map[string]string{},
		})
		return
	}

	c.JSON(http.StatusOK, helper.DataResponse{
		Message: "Log berhasil diambil",
		Data:    entries,
	})
}
