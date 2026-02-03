package executor

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"curlflow/internal/parser"
)

// SendRequest 接收请求对象，发送 HTTP 请求，返回结果
func SendRequest(req parser.HttpRequest) parser.HttpResponse {
	start := time.Now()
	response := parser.HttpResponse{
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
	var bodyReader io.Reader
	if req.Body != "" {
		bodyReader = bytes.NewBufferString(req.Body)
	}

	clientReq, err := http.NewRequest(req.Method, req.URL, bodyReader)
	if err != nil {
		response.Error = fmt.Sprintf("Build request failed: %v", err)
		return response
	}

	// 2. 填充 Headers
	for k, v := range req.Headers {
		clientReq.Header.Set(k, v)
	}

	// 3. 发送请求
	client := &http.Client{
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

	for k, v := range resp.Header {
		if len(v) > 0 {
			response.Headers[k] = v[0]
		}
	}

	return response
}
