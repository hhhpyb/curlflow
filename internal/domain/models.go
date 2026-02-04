package domain

// HttpRequest represents the parsed components of a curl command or a request to be executed.
type HttpRequest struct {
	URL        string            `json:"url"`
	Method     string            `json:"method"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
	Compressed bool              `json:"compressed"`
}

type RequestFile struct {
	Meta MetaData    `json:"_meta"`
	Data HttpRequest `json:"data"`
}

type MetaData struct {
	ID           string            `json:"id"`     // UUID, 唯一标识
	Key          string            `json:"key"`    // 辅助标识 (如 Method + Path)
	Status       string            `json:"status"` // active, deleted, new
	Summary      string            `json:"summary"`
	Description  string            `json:"description"` // 接口的详细描述
	Tags         []string          `json:"tags"`
	SwaggerPath  string            `json:"swagger_path"`
	LastSyncedAt int64             `json:"last_synced_at"`
	ParamDocs    map[string]string `json:"param_docs"` // 用于存储每个参数的说明
}

// HttpResponse represents the result of an executed HTTP request.
type HttpResponse struct {
	StatusCode int               `json:"statusCode"`
	Time       int64             `json:"time"`
	Body       string            `json:"body"`
	Headers    map[string]string `json:"headers"`
	Error      string            `json:"error"`
}
