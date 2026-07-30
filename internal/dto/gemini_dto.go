package dto

type AskGeminiRequest struct {
	Prompt string `json:"prompt" binding:"required"`
}

type GenerateImageRequest struct {
	Prompt string `json:"prompt" binding:"required"`
	Model  string `json:"model"`
	// File akan di-handle di handler melalui multipart form-data
}

type GenerateImageResponse struct {
	ImageURL string `json:"image_url"`
}

type ChatWithImageRequest struct {
	Prompt string `json:"prompt" binding:"required"`
	Model  string `json:"model"`
}

type GeminiResponse struct {
	Response string `json:"response"`
}
