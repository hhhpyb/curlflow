package domain

// AuthType defines the type of authorization
type AuthType string

const (
	AuthTypeNoAuth  AuthType = "noauth"
	AuthTypeInherit AuthType = "inherit"
	AuthTypeBearer  AuthType = "bearer"
	AuthTypeBasic   AuthType = "basic"
	AuthTypeApiKey  AuthType = "apikey"
)

// Auth holds authorization configuration
type Auth struct {
	Type AuthType          `json:"type"`
	Data map[string]string `json:"data"`
}

// HttpRequest represents the parsed components of a curl command or a request to be executed.
type HttpRequest struct {
	URL        string            `json:"url"`
	Method     string            `json:"method"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
	Compressed bool              `json:"compressed"`
	Auth       Auth              `json:"auth"`
}

type RequestFile struct {
	Meta MetaData    `json:"_meta"`
	Data HttpRequest `json:"data"`
}

type MetaData struct {
	ID              string            `json:"id"`     // UUID, 唯一标识
	Key             string            `json:"key"`    // 辅助标识 (如 Method + Path)
	Status          string            `json:"status"` // active, deleted, new
	Summary         string            `json:"summary"`
	Description     string            `json:"description"` // 接口的详细描述
	Tags            []string          `json:"tags"`
	SwaggerPath     string            `json:"swagger_path"`
	LastSyncedAt    int64             `json:"last_synced_at"`
	ParamDocs       map[string]string `json:"param_docs,omitempty"`    // 用于存储每个参数的说明
	ResponseDocs    map[string]string `json:"response_docs,omitempty"` // 存储返回参数说明
	ResponseExample string            `json:"response_example"`        // 存储返回数据示例
	Source          string            `json:"source"`                  // 来源: "swagger" (自动同步) 或 "user" (手动创建)
}

// HttpResponse represents the result of an executed HTTP request.
type HttpResponse struct {
	StatusCode int               `json:"statusCode"`
	Time       int64             `json:"time"`
	Body       string            `json:"body"`
	Headers    map[string]string `json:"headers"`
	Error      string            `json:"error"`
}

type ProjectConfig struct {
	Name        string `json:"name"`
	SwaggerURL  string `json:"swagger_url"`
	Auth        Auth   `json:"auth"`
	ProxyURL    string `json:"proxy_url"`
	Description string `json:"description"`
}
