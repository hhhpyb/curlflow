package parser

import (
	"fmt"
	"strings"
)

// ParseCurl 解析 Curl 字符串
func ParseCurl(curlCommand string) HttpRequest {
	// 初始化
	req := HttpRequest{
		Method:  "GET",
		Headers: make(map[string]string),
	}

	// ================= 1. 强力清洗 (预处理) =================
	// Chrome/Bash 特有的格式处理

	// 1.1 处理换行符：把 "反斜杠+换行" 和 "普通换行" 都替换为空格
	curlCommand = strings.ReplaceAll(curlCommand, "\\\n", " ")
	curlCommand = strings.ReplaceAll(curlCommand, "\n", " ")

	// 1.2 处理 Chrome 的 ANSI C Quoting ($'...')
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

	args, err := Tokenize(curlCommand)
	if err != nil {
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
		case "-d", "--data", "--data-raw", "--data-binary", "--data-ascii":
			if i+1 < len(args) {
				req.Body = args[i+1]
				if req.Method == "GET" {
					req.Method = "POST"
				}
				i++
			}
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

// BuildCurl 将 HttpRequest 对象转换回 Curl 字符串
func BuildCurl(req HttpRequest) string {
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
