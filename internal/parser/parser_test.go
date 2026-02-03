package parser

import (
	"testing"
)

func TestParseCurl(t *testing.T) {
	tests := []struct {
		name       string
		curlCmd    string
		wantMethod string
		wantURL    string
		wantBody   string
		check      func(t *testing.T, req *HttpRequest)
	}{
		{
			name:       "Safari Style - Header Filtering",
			curlCmd:    `curl 'http://example.com' -H 'Accept-Encoding: gzip, deflate' -H 'User-Agent: Safari/1.0'`,
			wantMethod: "GET",
			wantURL:    "http://example.com",
			check: func(t *testing.T, req *HttpRequest) {
				if _, ok := req.Headers["Accept-Encoding"]; ok {
					t.Error("Accept-Encoding should be filtered out")
				}
				if req.Headers["User-Agent"] != "Safari/1.0" {
					t.Errorf("Expected User-Agent Safari/1.0, got %s", req.Headers["User-Agent"])
				}
			},
		},
		{
			name:       "Chrome Style - POST Inference & Complex JSON",
			curlCmd:    `curl 'http://api.example.com/v1/data' --data-raw '{"id": 123, "name": "test"}' --compressed`,
			wantMethod: "POST",
			wantURL:    "http://api.example.com/v1/data",
			wantBody:   `{"id": 123, "name": "test"}`,
			check: func(t *testing.T, req *HttpRequest) {
				if !req.Compressed {
					t.Error("Expected Compressed to be true")
				}
			},
		},
		{
			name:       "Explicit PUT with Data",
			curlCmd:    `curl -X PUT 'http://example.com/resource' -d 'data'`,
			wantMethod: "PUT",
			wantURL:    "http://example.com/resource",
			wantBody:   "data",
			check:      func(t *testing.T, req *HttpRequest) {},
		},
		{
			name:       "URL with --url flag",
			curlCmd:    `curl --url http://example.com`,
			wantMethod: "GET",
			wantURL:    "http://example.com",
			check:      func(t *testing.T, req *HttpRequest) {},
		},
		{
			name: "Complex Multi-line with Escapes",
			curlCmd: `curl 'http://example.com' \
  -H 'Connection: keep-alive' \
  -H 'Content-Type: application/json' \
  --data-binary '{"key": "val\nue"}'`,
			wantMethod: "POST",
			wantURL:    "http://example.com",
			wantBody:   `{"key": "val\nue"}`,
			check: func(t *testing.T, req *HttpRequest) {
				if _, ok := req.Headers["Connection"]; ok {
					t.Error("Connection header should be filtered")
				}
				if req.Headers["Content-Type"] != "application/json" {
					t.Errorf("Expected Content-Type application/json, got %s", req.Headers["Content-Type"])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseCurl(tt.curlCmd)
			if err != nil {
				t.Fatalf("ParseCurl() error = %v", err)
			}

			if got.Method != tt.wantMethod {
				t.Errorf("Method = %v, want %v", got.Method, tt.wantMethod)
			}
			if got.URL != tt.wantURL {
				t.Errorf("URL = %v, want %v", got.URL, tt.wantURL)
			}
			if got.Body != tt.wantBody {
				t.Errorf("Body = %v, want %v", got.Body, tt.wantBody)
			}
			if tt.check != nil {
				t.check(t, got)
			}
		})
	}
}
