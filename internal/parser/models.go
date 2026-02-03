package parser

// HttpRequest 对应前端需要的请求对象
type HttpRequest struct {
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
}

// HttpResponse 对应前端收到的响应结果
type HttpResponse struct {
	StatusCode int               `json:"statusCode"` // 状态码 (200, 404)
	Time       int64             `json:"time"`       // 耗时 (毫秒)
	Body       string            `json:"body"`       // 响应内容
	Headers    map[string]string `json:"headers"`    // 响应头
	Error      string            `json:"error"`      // 错误信息 (如果有)
}
