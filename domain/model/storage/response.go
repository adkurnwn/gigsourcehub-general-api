package storage_model

type UploadResponse struct {
	Key         string `json:"key"`
	ContentType string `json:"content_type"`
	URL         string `json:"url"`
	Filename    string `json:"filename,omitempty"`
	FileSize    int64  `json:"file_size,omitempty"`
}
