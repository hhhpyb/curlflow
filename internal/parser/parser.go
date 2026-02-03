package parser

import (
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
func ParseCurl(curlCmd string) (*HttpRequest, error) {
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

	req := &HttpRequest{
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
			case "--compressed":
				req.Compressed = true
			case "--url":
				if i+1 < len(args) {
					req.URL = args[i+1]
					i++
				}
			default:
				// Unknown flag, ignore.
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
func BuildCurl(req HttpRequest) string {
	var sb strings.Builder
	sb.WriteString("curl")

	// Method
	if req.Method != "" && req.Method != "GET" {
		sb.WriteString(fmt.Sprintf(" -X %s", req.Method))
	}

	// URL
	if req.URL != "" {
		sb.WriteString(fmt.Sprintf(" '%s'", escapeSingleQuotes(req.URL)))
	}

	// Headers
	// Sort headers for deterministic output
	var keys []string
	for k := range req.Headers {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		v := req.Headers[k]
		sb.WriteString(fmt.Sprintf(" -H '%s: %s'", escapeSingleQuotes(k), escapeSingleQuotes(v)))
	}

	// Body
	if req.Body != "" {
		sb.WriteString(fmt.Sprintf(" --data-raw '%s'", escapeSingleQuotes(req.Body)))
	}

	// Compressed
	if req.Compressed {
		sb.WriteString(" --compressed")
	}

	return sb.String()
}

func escapeSingleQuotes(s string) string {
	return strings.ReplaceAll(s, "'", `''`)
}
