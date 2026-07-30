package repository

import (
	"bufio"
	"encoding/json"
	"fmt"
	"go-gin-layered-architecture/internal/model"
	"os"
	"path/filepath"
	"sort"
)

type LogRepository struct{}

func NewLogRepository() *LogRepository {
	return &LogRepository{}
}

func (r *LogRepository) GetAllLogFiles() ([]model.LogFile, error) {
	logsDir := "logs"
	entries, err := os.ReadDir(logsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []model.LogFile{}, nil
		}
		return nil, fmt.Errorf("failed to read logs directory: %w", err)
	}

	var files []model.LogFile
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".log" {
			info, err := entry.Info()
			if err != nil {
				continue
			}

			name := entry.Name()
			name = name[:len(name)-4]

			files = append(files, model.LogFile{
				FileName: name,
				Size:     info.Size(),
			})
		}
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].FileName < files[j].FileName
	})

	return files, nil
}

func (r *LogRepository) GetLogsByFileName(fileName string) ([]model.LogEntry, error) {
	path := filepath.Join("logs", fileName+".log")

	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("log file not found: %w", err)
	}

	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("path is not a regular file")
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}
	defer file.Close()

	var entries []model.LogEntry
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		var entry model.LogEntry
		line := scanner.Bytes()

		if err := json.Unmarshal(line, &entry); err != nil {
			continue
		}

		entries = append(entries, entry)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading log file: %w", err)
	}

	return entries, nil
}
