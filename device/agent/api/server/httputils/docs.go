package httputils

// HttpOK 正常返回
type HttpOK struct {
	Code   int    `json:"code" example:"200"`
	Result string `json:"result" example:"any result"`
}

// HttpError 异常返回
type HttpError struct {
	Code    int    `json:"code" example:"400"`
	Message string `json:"message" example:"status bad request"`
}
