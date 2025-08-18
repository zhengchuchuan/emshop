package response

// ImportResultItem 导入结果项
type ImportResultItem struct {
	Row     int   `json:"row"`
	GoodsId int32 `json:"goodsId"`
	Name    string `json:"name"`
}

// ImportResponse 导入响应
type ImportResponse struct {
	Results      []ImportResultItem `json:"results"`
	Errors       []string           `json:"errors"`
	Total        int                `json:"total"`
	SuccessCount int                `json:"successCount"`
	ErrorCount   int                `json:"errorCount"`
}

// ValidateResponse 验证响应
type ValidateResponse struct {
	Valid      bool     `json:"valid"`
	TotalRows  int      `json:"totalRows"`
	ValidRows  int      `json:"validRows"`
	ErrorRows  int      `json:"errorRows"`
	Errors     []string `json:"errors"`
	Filename   string   `json:"filename"`
	FileSize   int64    `json:"fileSize"`
}