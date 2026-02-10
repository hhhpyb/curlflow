package parser

import (
	"curlflow/internal/domain"
	"encoding/base64"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/mattn/go-shellwords"
)

// headerBlacklist contains lower-cased headers that should be ignored.
var headerBlacklist = map[string]struct{}{
	"accept-encoding": {},
	"content-length":  {},
	"connection":      {},
	"host":            {},
}

// ParseCurl parses a curl command string into a HttpRequest struct.
// It uses shellwords for tokenization and handles browser-specific quirks.
func ParseCurl(curlCmd string) (*domain.HttpRequest, error) {
	if strings.TrimSpace(curlCmd) == "" {
		return nil, errors.New("empty curl command")
	}

	// Sanitize input: handle line continuations and newlines
	curlCmd = strings.ReplaceAll(curlCmd, "\\\n", " ")
	curlCmd = strings.ReplaceAll(curlCmd, "\n", " ")

	args, err := shellwords.Parse(curlCmd)
	if err != nil {
		return nil, err
	}

	if len(args) == 0 {
		return nil, errors.New("no arguments found")
	}

	// Remove "curl" from the beginning if present
	if args[0] == "curl" {
		args = args[1:]
	}

	req := &domain.HttpRequest{
		Method:  "GET",
		Headers: make(map[string]string),
	}

	var hasData bool
	var explicitMethod bool

	for i := 0; i < len(args); i++ {
		arg := args[i]

		if strings.HasPrefix(arg, "-") {
			// Handle Flags
			switch arg {
			case "-X", "--request":
				if i+1 < len(args) {
					req.Method = args[i+1]
					explicitMethod = true
					i++
				}
			case "-H", "--header":
				if i+1 < len(args) {
					parseHeader(req.Headers, args[i+1])
					i++
				}
			case "-d", "--data", "--data-raw", "--data-binary":
				if i+1 < len(args) {
					req.Body += args[i+1]
					hasData = true
					i++
				}
			case "--url":
				if i+1 < len(args) {
					req.URL = args[i+1]
					i++
				}
			default:
				// Unknown or unsupported flag (e.g., --compressed), ignore.
				continue
			}
		} else {
			// Positional argument (URL)
			if req.URL == "" {
				req.URL = arg
			}
		}
	}

	// Method Inference Logic
	if !explicitMethod {
		if hasData {
			req.Method = "POST"
		}
	}

	// Smart Parse Logic: Extract Auth from Headers
	// Iterate through keys to check for "Authorization" (case-insensitive check needed, but headers map keys are raw)
	// Since we don't normalize keys in parseHeader except for blacklist check, we need to find the key.
	var authKey string
	for k := range req.Headers {
		if strings.ToLower(k) == "authorization" {
			authKey = k
			break
		}
	}

	if authKey != "" {
		val := req.Headers[authKey]
		// 1. Bearer Token
		if strings.HasPrefix(strings.ToLower(val), "bearer ") {
			token := strings.TrimSpace(val[7:])
			req.Auth.Type = domain.AuthTypeBearer
			req.Auth.Data = map[string]string{"token": token}
			delete(req.Headers, authKey)
		} else if strings.HasPrefix(strings.ToLower(val), "basic ") {
			// 2. Basic Auth
			encoded := strings.TrimSpace(val[6:])
			decoded, err := base64.StdEncoding.DecodeString(encoded)
			if err == nil {
				payload := string(decoded)
				parts := strings.SplitN(payload, ":", 2)
				if len(parts) == 2 {
					req.Auth.Type = domain.AuthTypeBasic
					req.Auth.Data = map[string]string{
						"username": parts[0],
						"password": parts[1],
					}
					delete(req.Headers, authKey)
				}
			}
		}
		// If not matched, leave it as manual header
	}
	// Initializing Auth struct for safety if not set
	if req.Auth.Type == "" {
		req.Auth.Type = domain.AuthTypeNoAuth
		req.Auth.Data = make(map[string]string)
	}

	return req, nil
}

func parseHeader(headers map[string]string, headerStr string) {
	parts := strings.SplitN(headerStr, ":", 2)
	if len(parts) != 2 {
		return
	}

	key := strings.TrimSpace(parts[0])
	val := strings.TrimSpace(parts[1])

	// Check blacklist (case-insensitive)
	if _, ok := headerBlacklist[strings.ToLower(key)]; ok {
		return
	}

	headers[key] = val
}

// BuildCurl reconstructs a curl command string from a HttpRequest struct.
func BuildCurl(req domain.HttpRequest) string {
	var sb strings.Builder
	sb.WriteString("curl")

	// Method
	if req.Method != "" && req.Method != "GET" {
		sb.WriteString(fmt.Sprintf(" -X %s", req.Method))
	}

	// URL
	urlFinal := req.URL

	// Handle API Key in Query (modify URL)
	if req.Auth.Type == domain.AuthTypeApiKey && req.Auth.Data["addTo"] == "query" {
		key := req.Auth.Data["key"]
		val := req.Auth.Data["value"]
		if key != "" && val != "" {
			separator := "?"
			if strings.Contains(urlFinal, "?") {
				separator = "&"
			}
			urlFinal = fmt.Sprintf("%s%s%s=%s", urlFinal, separator, key, val)
		}
	}

	if urlFinal != "" {
		sb.WriteString(fmt.Sprintf(" '%s'", escapeSingleQuotes(urlFinal)))
	}

	// Headers
	// 1. Copy existing headers
	finalHeaders := make(map[string]string)
	for k, v := range req.Headers {
		finalHeaders[k] = v
	}

	// 2. Enrich with Auth Headers (Override if exists)
	switch req.Auth.Type {
	case domain.AuthTypeBearer:
		if token, ok := req.Auth.Data["token"]; ok && token != "" {
			finalHeaders["Authorization"] = "Bearer " + token
		}
	case domain.AuthTypeBasic:
		username := req.Auth.Data["username"]
		password := req.Auth.Data["password"]
		auth := username + ":" + password
		encoded := base64.StdEncoding.EncodeToString([]byte(auth))
		finalHeaders["Authorization"] = "Basic " + encoded
	case domain.AuthTypeApiKey:
		if req.Auth.Data["addTo"] != "query" { // Default to header
			key := req.Auth.Data["key"]
			val := req.Auth.Data["value"]
			if key != "" && val != "" {
				finalHeaders[key] = val
			}
		}
	}

	// Sort headers for deterministic output
	var keys []string
	for k := range finalHeaders {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		v := finalHeaders[k]
		sb.WriteString(fmt.Sprintf(" -H '%s: %s'", escapeSingleQuotes(k), escapeSingleQuotes(v)))
	}

	// Body
	if req.Body != "" {
		sb.WriteString(fmt.Sprintf(" --data-raw '%s'", escapeSingleQuotes(req.Body)))
	}

	return sb.String()
}

func escapeSingleQuotes(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}
