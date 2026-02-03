package main

import (
	"bytes"
	"context"
	"fmt"

	"io"
	"net/http"
	"strings"
	"time"

	"github.com/mattn/go-shellwords"
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
// ParseCurl 解析 Curl 字符串
func (a *App) ParseCurl(curlCommand string) HttpRequest {
	// 初始化
	req := HttpRequest{
		Method:  "GET",
		Headers: make(map[string]string),
	}

	// ================= 1. 强力清洗 (预处理) =================
	// Chrome/Bash 特有的格式处理

	// 1.1 处理换行符：把 "反斜杠+换行" 和 "普通换行" 都替换为空格
	// 这样可以把多行命令变成单行，方便解析
	curlCommand = strings.ReplaceAll(curlCommand, "\\\n", " ")
	curlCommand = strings.ReplaceAll(curlCommand, "\n", " ")

	// 1.2 处理 Chrome 的 ANSI C Quoting ($'...')
	// Chrome 会把带 Unicode 的 JSON 包裹在 $'...' 里。
	// 大多数解析库不支持这个 $，我们简单粗暴地把它替换成普通单引号 '...'
	// 注意：这里加个空格 " $'" 是为了防止替换掉非参数内容的字符，虽不完美但够用
	curlCommand = strings.ReplaceAll(curlCommand, " $'", " '")
	// 针对可能出现在开头的 (虽然少见)
	if strings.HasPrefix(curlCommand, "$'") {
		curlCommand = "'" + curlCommand[2:]
	}

	// 1.3 去掉首尾空格
	curlCommand = strings.TrimSpace(curlCommand)

	// ================= 2. 解析 =================
	if !strings.HasPrefix(curlCommand, "curl") {
		req.URL = curlCommand
		return req
	}

	args, err := shellwords.Parse(curlCommand)
	if err != nil {
		// 如果解析失败，返回一个带错误信息的 Dummy 对象，或者打印日志
		fmt.Println("Curl Parse Error:", err)
		return req
	}

	for i := 0; i < len(args); i++ {
		arg := args[i]

		switch arg {
		case "curl":
			continue
		case "-X", "--request":
			if i+1 < len(args) {
				req.Method = strings.ToUpper(args[i+1])
				i++
			}
		case "-H", "--header":
			if i+1 < len(args) {
				headerStr := args[i+1]
				parts := strings.SplitN(headerStr, ":", 2)
				if len(parts) == 2 {
					key := strings.TrimSpace(parts[0])
					val := strings.TrimSpace(parts[1])
					req.Headers[key] = val
				}
				i++
			}
		// >>> 修复点：添加 --data-binary 和 --data-urlencode 支持 <<<
		case "-d", "--data", "--data-raw", "--data-binary", "--data-ascii":
			if i+1 < len(args) {
				req.Body = args[i+1]
				if req.Method == "GET" {
					req.Method = "POST"
				}
				i++
			}
		// 有些 curl 会写 --compressed，我们要忽略它，否则可能会误判为 URL
		case "--compressed":
			continue
		default:
			// 简单的 URL 识别逻辑：非 flag 且以 http 开头
			if strings.HasPrefix(arg, "http") {
				req.URL = arg
			}
		}
	}

	return req
}

// SendRequest 接收前端的请求对象，发送 HTTP 请求，返回结果
func (a *App) SendRequest(req HttpRequest) HttpResponse {
	start := time.Now()
	response := HttpResponse{
		Headers: make(map[string]string),
	}

	// 0. 基本校验
	if req.URL == "" {
		response.Error = "URL cannot be empty"
		return response
	}
	if !strings.HasPrefix(strings.ToLower(req.URL), "http") {
		response.Error = "Invalid URL: Must start with http:// or https://"
		return response
	}

	// 1. 构造 Request 对象
	// 如果 Body 不为空，转为 Reader
	var bodyReader io.Reader
	if req.Body != "" {
		bodyReader = bytes.NewBufferString(req.Body)
	}

	// 创建原生 http.Request
	clientReq, err := http.NewRequest(req.Method, req.URL, bodyReader)
	if err != nil {
		response.Error = fmt.Sprintf("Build request failed: %v", err)
		return response
	}

	// 2. 填充 Headers
	for k, v := range req.Headers {
		clientReq.Header.Set(k, v)
	}

	// 3. 发送请求 (默认 Client)
	client := &http.Client{
		// 设置一个超时，防止卡死
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(clientReq)
	if err != nil {
		response.Error = fmt.Sprintf("Network error: %v", err)
		return response
	}
	defer resp.Body.Close()

	// 4. 读取响应 Body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		response.Error = fmt.Sprintf("Read body failed: %v", err)
		return response
	}

	// 5. 组装结果
	response.StatusCode = resp.StatusCode
	response.Body = string(bodyBytes)
	response.Time = time.Since(start).Milliseconds()

	// 收集响应头 (为了简单，只取第一个值)
	for k, v := range resp.Header {
		if len(v) > 0 {
			response.Headers[k] = v[0]
		}
	}

	return response
}

func (a *App) BuildCurl(req HttpRequest) string {
	curl := fmt.Sprintf("curl -X %s '%s'", req.Method, req.URL)

	for key, value := range req.Headers {
		curl += fmt.Sprintf(" -H '%s: %s'", key, value)
	}

	if req.Body != "" {
		// 简单处理 Body 中的单引号转义，确保在 shell 中可用
		escapedBody := strings.ReplaceAll(req.Body, "'", "'\\''")
		curl += fmt.Sprintf(" -d '%s'", escapedBody)
	}

	return curl
}
