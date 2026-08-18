package httpclient

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type PageRequest struct {
	Page     int    `json:"page"`
	PageSize int    `json:"pageSize"`
	Keyword  string `json:"keyword,omitempty"`
}

type PageResponse struct {
	Total int64       `json:"total"`
	Page  int         `json:"page"`
	Size  int         `json:"size"`
	Data  interface{} `json:"data"`
}

type ErrorResponse struct {
	Error string `json:"error"`
	Code  string `json:"code,omitempty"`
}

type Options struct {
	Headers     map[string]string
	QueryParams map[string]string
	RetryCount  int
	NoRetry     bool
}

// Seeed API 相关结构体
type SeeedModifyRecordingRequest struct {
	Token   string `json:"Token"`
	SId     string `json:"SId"`
	Content string `json:"Content"`
}

type SeeedModifyRecordingResponse struct {
	Code  int         `json:"code"`
	Data  interface{} `json:"data"`
	Msg   string      `json:"msg"`
	Count int         `json:"count"`
}
