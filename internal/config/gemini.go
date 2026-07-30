package config

import (
	"log"
	"os"
)

type GeminiConfig struct {
	GeminiAPIKey            string
	GeminiDefaultChatModel  string
	GeminiDefaultImageModel string
}

func GeminiLoad() *GeminiConfig {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Fatal("gagal memuat konfigurasi: GEMINI_API_KEY is not set")
	}

	chatModel := os.Getenv("GEMINI_DEFAULT_CHAT_MODEL")
	if chatModel == "" {
		chatModel = "gemini-3.1-flash-lite"
	}

	imageModel := os.Getenv("GEMINI_DEFAULT_IMAGE_MODEL")
	if imageModel == "" {
		imageModel = "gemini-3.1-flash-lite-image"
	}

	return &GeminiConfig{
		GeminiAPIKey:            apiKey,
		GeminiDefaultChatModel:  chatModel,
		GeminiDefaultImageModel: imageModel,
	}
}
