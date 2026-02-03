package main

import (
	"context"

	"curlflow/internal/executor"
	"curlflow/internal/parser"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// ================== 1. 定义数据结构 (DTO) ==================

// HttpRequest 对应前端需要的请求对象
// 注意：字段首字母必须大写(Public)，否则 JSON 序列化时会忽略，前端拿不到数据！
// json tag 指定前端看到的字段名
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

// ================== 2. 核心业务逻辑 (Service) ==================

// ParseCurl 解析 Curl 字符串
// 这个方法会被导出给前端调用
func (a *App) ParseCurl(curlCommand string) (HttpRequest, error) {
	// 调用 internal/parser 的逻辑
	parsed, err := parser.ParseCurl(curlCommand)
	if err != nil {
		return HttpRequest{}, err
	}

	// DTO 转换 (internal model -> view model)
	return HttpRequest{
		Method:  parsed.Method,
		URL:     parsed.URL,
		Headers: parsed.Headers,
		Body:    parsed.Body,
	}, nil
}

// SendRequest 接收前端的请求对象，发送 HTTP 请求，返回结果
func (a *App) SendRequest(req HttpRequest) HttpResponse {
	// DTO 转换 (view model -> internal model)
	internalReq := parser.HttpRequest{
		Method:  req.Method,
		URL:     req.URL,
		Headers: req.Headers,
		Body:    req.Body,
	}

	// 调用 internal/executor
	resp := executor.SendRequest(internalReq)

	// DTO 转换 (internal model -> view model)
	return HttpResponse{
		StatusCode: resp.StatusCode,
		Time:       resp.Time,
		Body:       resp.Body,
		Headers:    resp.Headers,
		Error:      resp.Error,
	}
}

func (a *App) BuildCurl(req HttpRequest) string {
	// DTO 转换
	internalReq := parser.HttpRequest{
		Method:  req.Method,
		URL:     req.URL,
		Headers: req.Headers,
		Body:    req.Body,
	}

	return parser.BuildCurl(internalReq)
}
