package runner

import (
	"bytes"
	"curlflow/internal/domain"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Service handles HTTP request execution.
type Service struct{}

// NewService creates a new instance of the runner Service.
func NewService() *Service {
	return &Service{}
}

// SendRequest executes the HTTP request described by domain.HttpRequest and returns domain.HttpResponse.
func (s *Service) SendRequest(req domain.HttpRequest) domain.HttpResponse {
	start := time.Now()
	response := domain.HttpResponse{
		Headers: make(map[string]string),
	}

	// 0. Basic Validation
	if req.URL == "" {
		response.Error = "URL cannot be empty"
		return response
	}
	if !strings.HasPrefix(strings.ToLower(req.URL), "http") {
		response.Error = "Invalid URL: Must start with http:// or https://"
		return response
	}

	// 1. Build Request Object
	var bodyReader io.Reader
	if req.Body != "" {
		bodyReader = bytes.NewBufferString(req.Body)
	}

	clientReq, err := http.NewRequest(req.Method, req.URL, bodyReader)
	if err != nil {
		response.Error = fmt.Sprintf("Build request failed: %v", err)
		return response
	}

	// 2. Fill Headers
	for k, v := range req.Headers {
		// Ignore Accept-Encoding to let Go's http.Transport handle gzip automatically.
		if strings.EqualFold(k, "Accept-Encoding") {
			continue
		}
		// Use direct assignment to preserve header case
		clientReq.Header[k] = []string{v}
	}

	// 3. Send Request
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(clientReq)
	if err != nil {
		response.Error = fmt.Sprintf("Network error: %v", err)
		return response
	}
	defer resp.Body.Close()

	// 4. Read Response Body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		response.Error = fmt.Sprintf("Read body failed: %v", err)
		return response
	}

	// 5. Assemble Result
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
