package dto

type UploadFileResponse struct {
	FileURL string `json:"file_url"`
}

type UploadBase64FileRequest struct {
	Folder   string `json:"folder"`
	FileName string `json:"file_name" binding:"required"`
	// File menerima base64 murni maupun data URI (contoh: data:image/png;base64,iVBORw0...)
	File string `json:"file" binding:"required"`
	// ContentType opsional, diambil dari data URI atau dideteksi dari isi file bila kosong
	ContentType string `json:"content_type"`
}

type DeleteFileRequest struct {
	Folder string `json:"folder"`
	// FileName menerima nama file, path lengkap, maupun URL CDN penuh hasil upload.
	// Bila berupa URL penuh, Folder diabaikan karena URL sudah memuat path lengkap.
	FileName string `json:"file_name" binding:"required"`
}
