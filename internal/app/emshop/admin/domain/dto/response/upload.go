package response

// UploadResponse 单文件上传响应
type UploadResponse struct {
	Url      string `json:"url"`
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
	Type     string `json:"type"`
}

// BatchUploadResponse 批量上传响应
type BatchUploadResponse struct {
	Results      []UploadResponse `json:"results"`
	Errors       []string         `json:"errors"`
	Total        int              `json:"total"`
	SuccessCount int              `json:"successCount"`
	ErrorCount   int              `json:"errorCount"`
}