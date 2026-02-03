package parser

// HttpRequest represents the parsed components of a curl command or a request to be executed.
type HttpRequest struct {
	URL        string            `json:"url"`
	Method     string            `json:"method"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
	Compressed bool              `json:"compressed"`
}

// HttpResponse represents the result of an executed HTTP request.
type HttpResponse struct {
	StatusCode int               `json:"statusCode"`
	Time       int64             `json:"time"`
	Body       string            `json:"body"`
	Headers    map[string]string `json:"headers"`
	Error      string            `json:"error"`
}
