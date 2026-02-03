package main

import (
	"context"
	"curlflow/internal/domain"
	"curlflow/internal/parser"
	"curlflow/internal/runner"
	"curlflow/internal/storage"
	"fmt"
	"path/filepath"
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

// ParseCurl parses a curl command string into a HttpRequest struct
func (a *App) ParseCurl(curl string) domain.HttpRequest {
	req, err := parser.ParseCurl(curl)
	if err != nil {
		fmt.Printf("ParseCurl error: %v\n", err)
		return domain.HttpRequest{}
	}
	return *req
}

// BuildCurl reconstructs a curl command string from a HttpRequest struct
func (a *App) BuildCurl(req domain.HttpRequest) string {
	return parser.BuildCurl(req)
}

// SendRequest executes the HTTP request
func (a *App) SendRequest(req domain.HttpRequest) domain.HttpResponse {
	return a.runner.SendRequest(req)
}

// SelectWorkDir opens a directory selection dialog
func (a *App) SelectWorkDir() string {
	dir, err := a.storage.SelectWorkingDirectory(a.ctx)
	if err != nil {
		fmt.Printf("SelectWorkDir error: %v\n", err)
		return ""
	}
	return dir
}

// GetFileList lists all request files in the specified directory
func (a *App) GetFileList(dir string) ([]string, error) {
	files, err := a.storage.ListRequestFiles(dir)
	if err != nil {
		return nil, err
	}
	return files, nil
}

// SaveRequest saves the request to a file
func (a *App) SaveRequest(dir string, filename string, req domain.HttpRequest) (string, error) {
	path, err := a.storage.SaveRequest(dir, filename, req)
	if err != nil {
		return "", err
	}
	return path, nil
}

// LoadRequest loads a request from a file
func (a *App) LoadRequest(dir string, filename string) (domain.HttpRequest, error) {
	// Construct full path since storage expects it
	fullPath := filepath.Join(dir, filename)

	req, err := a.storage.LoadRequest(fullPath)
	if err != nil {
		return domain.HttpRequest{}, err
	}
	return req, nil
}

// SaveConfig saves a configuration string to a file
func (a *App) SaveConfig(dir string, filename string, content string) (string, error) {
	fullPath := filepath.Join(dir, filename)
	err := a.storage.SaveFile(fullPath, content)
	if err != nil {
		return "", err
	}
	return fullPath, nil
}

// LoadConfig loads a configuration string from a file
func (a *App) LoadConfig(dir string, filename string) (string, error) {
	fullPath := filepath.Join(dir, filename)
	content, err := a.storage.LoadFile(fullPath)
	if err != nil {
		return "", err
	}
	return content, nil
}
