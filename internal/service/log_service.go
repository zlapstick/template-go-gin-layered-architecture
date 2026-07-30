package service

import (
	"go-gin-layered-architecture/internal/model"
	"go-gin-layered-architecture/internal/repository"
)

type LogService struct {
	logRepository *repository.LogRepository
}

func NewLogService(logRepository *repository.LogRepository) *LogService {
	return &LogService{
		logRepository: logRepository,
	}
}

func (s *LogService) GetAllLogFiles() ([]model.LogFile, error) {
	return s.logRepository.GetAllLogFiles()
}

func (s *LogService) GetLogsByFileName(fileName string) ([]model.LogEntry, error) {
	return s.logRepository.GetLogsByFileName(fileName)
}
