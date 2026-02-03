package main

import (
	"context"
	"path/filepath"

	"curlflow/internal/domain"
	"curlflow/internal/parser"
	"curlflow/internal/runner"
	"curlflow/internal/storage"
)

// App struct
type App struct {
	ctx     context.Context
	runner  *runner.Service
	storage *storage.Service
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		runner:  runner.NewService(),
		storage: storage.NewService(),
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// ================== Parser Logic (Stateless) ==================

// ParseCurl parses a curl command string into a HttpRequest object.
func (a *App) ParseCurl(curlCommand string) (domain.HttpRequest, error) {
	reqPtr, err := parser.ParseCurl(curlCommand)
	if err != nil {
		return domain.HttpRequest{}, err
	}
	// Dereference the pointer to return the struct value
	return *reqPtr, nil
}

// BuildCurl builds a curl command string from a HttpRequest object.
func (a *App) BuildCurl(req domain.HttpRequest) string {
	return parser.BuildCurl(req)
}

// ================== Runner Logic ==================

// SendRequest executes the HTTP request.
func (a *App) SendRequest(req domain.HttpRequest) domain.HttpResponse {
	return a.runner.SendRequest(req)
}

// ================== Storage Logic ==================

// SelectWorkDir opens a dialog to select the working directory.
func (a *App) SelectWorkDir() string {
	dir, err := a.storage.SelectWorkingDirectory(a.ctx)
	if err != nil {
		return ""
	}
	return dir
}

// GetFileList lists all request files in the given directory.
func (a *App) GetFileList(dir string) []string {
	files, err := a.storage.ListRequestFiles(dir)
	if err != nil {
		return []string{}
	}
	return files
}

// SaveRequest saves the request object to a file.
func (a *App) SaveRequest(dir string, name string, req domain.HttpRequest) string {
	path, err := a.storage.SaveRequest(dir, name, req)
	if err != nil {
		return ""
	}
	return path
}

// LoadRequest loads a request object from a file.
func (a *App) LoadRequest(dir string, name string) domain.HttpRequest {
	fullPath := filepath.Join(dir, name)
	req, err := a.storage.LoadRequest(fullPath)
	if err != nil {
		return domain.HttpRequest{}
	}
	return req
}
