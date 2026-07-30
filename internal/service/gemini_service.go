package service

import (
	"context"
	"encoding/base64"
	"fmt"
	"go-gin-layered-architecture/internal/config"

	"google.golang.org/genai"
)

type GeminiService struct {
	geminiConfig *config.GeminiConfig
}

func NewGeminiService(geminiConfig *config.GeminiConfig) *GeminiService {
	return &GeminiService{geminiConfig: geminiConfig}
}

func (s *GeminiService) Ask(ctx context.Context, prompt string, docBytes []byte, mimeType string, model string) (string, error) {
	if model == "" {
		model = s.geminiConfig.GeminiDefaultChatModel
	}

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: s.geminiConfig.GeminiAPIKey,
	})
	if err != nil {
		return "", fmt.Errorf("initializing gemini client: %w", err)
	}

	parts := []*genai.Part{genai.NewPartFromText(prompt)}

	if docBytes != nil {
		parts = append(parts, genai.NewPartFromBytes(docBytes, mimeType))
	}

	content := []*genai.Content{genai.NewContentFromParts(parts, genai.RoleUser)}

	res, err := client.Models.GenerateContent(ctx, model, content, nil)
	if err != nil {
		return "", fmt.Errorf("generating content: %w", err)
	}
	return res.Text(), nil
}

func (s *GeminiService) GenerateImage(ctx context.Context, prompt string, docBytes []byte, mimeType string, model string) (string, error) {
	if model == "" {
		model = s.geminiConfig.GeminiDefaultImageModel
	}

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: s.geminiConfig.GeminiAPIKey,
	})
	if err != nil {
		return "", fmt.Errorf("initializing gemini client: %w", err)
	}

	parts := []*genai.Part{genai.NewPartFromText(prompt)}

	if docBytes != nil {
		parts = append(parts, genai.NewPartFromBytes(docBytes, mimeType))
	}

	content := []*genai.Content{genai.NewContentFromParts(parts, genai.RoleUser)}

	res, err := client.Models.GenerateContent(ctx, model, content, nil)
	if err != nil {
		return "", fmt.Errorf("generating image: %w", err)
	}

	if len(res.Candidates) == 0 {
		return "", fmt.Errorf("no image generated")
	}

	if len(res.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no image data in response")
	}

	part := res.Candidates[0].Content.Parts[0]
	if part.InlineData == nil || part.InlineData.Data == nil {
		return "", fmt.Errorf("no inline data in response")
	}

	encodedData := base64.StdEncoding.EncodeToString(part.InlineData.Data)
	return fmt.Sprintf("data:%s;base64,%s", part.InlineData.MIMEType, encodedData), nil
}
