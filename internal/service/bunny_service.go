package service

import (
	"bytes"
	"context"
	"fmt"
	"go-gin-layered-architecture/internal/config"
	"go-gin-layered-architecture/internal/model"
	"net/http"
	"net/url"
	"path"
	"strings"
)

type BunnyService struct {
	bunnyConfig *config.BunnyConfig
}

func NewBunnyService(bunnyConfig *config.BunnyConfig) *BunnyService {
	return &BunnyService{bunnyConfig: bunnyConfig}
}

func (s *BunnyService) UploadFile(ctx context.Context, folder string, fileName string, fileBytes []byte, contentType string) (string, error) {
	parts := make([]string, 0)
	for _, segment := range []string{folder, fileName} {
		for _, part := range strings.Split(segment, "/") {
			part = strings.TrimSpace(part)
			if part == "" || part == "." || part == ".." {
				continue
			}
			parts = append(parts, part)
		}
	}

	storagePath := path.Join(parts...)
	if storagePath == "" {
		return "", fmt.Errorf("file path is empty")
	}

	escaped := make([]string, 0, len(parts))
	for _, part := range strings.Split(storagePath, "/") {
		escaped = append(escaped, url.PathEscape(part))
	}
	escapedPath := strings.Join(escaped, "/")

	host := "storage.bunnycdn.com"
	if location := strings.ToLower(strings.TrimSpace(s.bunnyConfig.BunnyStorageLocation)); location != "" && location != "de" {
		host = location + "." + host
	}

	uploadURL := fmt.Sprintf("https://%s/%s/%s", host, s.bunnyConfig.BunnyStorageZoneName, escapedPath)

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, uploadURL, bytes.NewReader(fileBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", contentType)
	req.Header.Set("AccessKey", s.bunnyConfig.BunnyStorageAccessKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("failed to upload file: %s", resp.Status)
	}

	return fmt.Sprintf("%s/%s", strings.TrimRight(s.bunnyConfig.BunnyStorageCDNUrl, "/"), escapedPath), nil
}

func (s *BunnyService) DeleteFile(ctx context.Context, folder string, fileName string) error {
	parts := make([]string, 0)
	fileName = strings.TrimSpace(fileName)
	if parsed, err := url.Parse(fileName); err == nil && parsed.Host != "" {
		fileName = parsed.Path
		folder = ""
	} else if decoded, err := url.PathUnescape(fileName); err == nil {
		fileName = decoded
	}

	for _, segment := range []string{folder, fileName} {
		for _, part := range strings.Split(segment, "/") {
			part = strings.TrimSpace(part)
			if part == "" || part == "." || part == ".." {
				continue
			}
			parts = append(parts, part)
		}
	}

	storagePath := path.Join(parts...)
	if storagePath == "" {
		return fmt.Errorf("file path is empty")
	}

	escaped := make([]string, 0, len(parts))
	for _, part := range strings.Split(storagePath, "/") {
		escaped = append(escaped, url.PathEscape(part))
	}
	escapedPath := strings.Join(escaped, "/")

	host := "storage.bunnycdn.com"
	if location := strings.ToLower(strings.TrimSpace(s.bunnyConfig.BunnyStorageLocation)); location != "" && location != "de" {
		host = location + "." + host
	}

	deleteURL := fmt.Sprintf("https://%s/%s/%s", host, s.bunnyConfig.BunnyStorageZoneName, escapedPath)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, deleteURL, nil)

	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("AccessKey", s.bunnyConfig.BunnyStorageAccessKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return model.ErrFileNotFound
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to delete file: %s", resp.Status)
	}

	return nil
}
