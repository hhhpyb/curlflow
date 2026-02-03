package runner

import (
	"context"
	"crypto/tls"
	"curlflow/internal/domain"
	"curlflow/internal/parser"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// RunnerConfig holds configuration for the HTTP client
type RunnerConfig struct {
	ProxyURL string `json:"proxyUrl"`
	Insecure bool   `json:"insecure"`
	Timeout  int    `json:"timeout"`
}

type Service struct {
	config RunnerConfig
}

func NewService() *Service {
	// Default configuration
	return &Service{
		config: RunnerConfig{
			Timeout: 30, // Default 30s timeout
		},
	}
}

// UpdateConfig updates the runner configuration
func (s *Service) UpdateConfig(cfg RunnerConfig) {
	// Sanitize configuration
	cfg.ProxyURL = strings.TrimSpace(cfg.ProxyURL)
	s.config = cfg
}

func (s *Service) SendRequest(req domain.HttpRequest) domain.HttpResponse {
	start := time.Now()

	fmt.Printf("DEBUG: SendRequest Config - Proxy: '%s', Insecure: %v, Timeout: %d\n",
		s.config.ProxyURL, s.config.Insecure, s.config.Timeout)

	// 1. Prepare Transport
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: s.config.Insecure,
		},
		// Explicitly ensure Proxy is nil by default (direct connection)
		Proxy: nil,
	}

	// 2. Configure Proxy
	if s.config.ProxyURL != "" {
		proxyURL, err := url.Parse(s.config.ProxyURL)
		if err == nil {
			fmt.Printf("DEBUG: Using Proxy: %s\n", proxyURL.String())
			transport.Proxy = http.ProxyURL(proxyURL)
		} else {
			return domain.HttpResponse{
				Error: fmt.Sprintf("Invalid proxy URL: %v", err),
			}
		}
	} else {
		fmt.Println("DEBUG: No Proxy configured (Direct connection)")
	}

	// 3. Create Client
	timeout := time.Duration(s.config.Timeout) * time.Second
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}

	// 4. Create Request
	var bodyReader io.Reader
	if req.Body != "" {
		bodyReader = strings.NewReader(req.Body)
	}

	httpReq, err := http.NewRequest(req.Method, req.URL, bodyReader)
	if err != nil {
		return domain.HttpResponse{
			Error: fmt.Sprintf("Failed to create request: %v", err),
		}
	}

	// 5. Add Headers
	for k, v := range req.Headers {
		httpReq.Header.Set(k, v)
	}

	// 6. Execute
	resp, err := client.Do(httpReq)
	duration := time.Since(start).Milliseconds()

	if err != nil {
		return domain.HttpResponse{
			Error: fmt.Sprintf("Request failed: %v", err),
			Time:  duration,
		}
	}
	defer resp.Body.Close()

	// 7. Read Response
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return domain.HttpResponse{
			Error: fmt.Sprintf("Failed to read response body: %v", err),
			Time:  duration,
		}
	}

	// 8. Convert Headers
	headers := make(map[string]string)
	for k, v := range resp.Header {
		headers[k] = strings.Join(v, ", ")
	}

	return domain.HttpResponse{
		StatusCode: resp.StatusCode,
		Headers:    headers,
		Body:       string(bodyBytes),
		Time:       duration,
	}
}

// ParseCurl parses a cURL command into an HttpRequest object.
func (s *Service) ParseCurl(curlCommand string) (domain.HttpRequest, error) {
	req, err := parser.ParseCurl(curlCommand)
	if err != nil {
		return domain.HttpRequest{}, err
	}
	return *req, nil
}

// BuildCurl generates a cURL command from an HttpRequest object.
func (s *Service) BuildCurl(req domain.HttpRequest) (string, error) {
	return parser.BuildCurl(req), nil
}

func (s *Service) Execute(ctx context.Context, cmd string) (string, error) {
	return "", nil
}
