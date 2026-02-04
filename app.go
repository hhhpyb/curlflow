package main

import (
	"context"
	"curlflow/internal/domain"
	"curlflow/internal/parser"
	"curlflow/internal/runner"
	"curlflow/internal/storage"
	"curlflow/internal/syncer"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

// App struct
type App struct {
	ctx     context.Context
	runner  *runner.Service
	storage *storage.Service
	syncer  *syncer.Service
}

// AppConfig holds global application settings
type AppConfig struct {
	ProxyURL string `json:"proxyUrl"`
	Insecure bool   `json:"insecure"`
	Timeout  int    `json:"timeout"`
}

// NewApp creates a new App application struct
func NewApp() *App {
	storageService := storage.NewService()
	return &App{
		runner:  runner.NewService(),
		storage: storageService,
		syncer:  syncer.NewService(storageService),
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	// Load settings on startup and apply to runner
	cfg := a.GetSettings()
	a.runner.UpdateConfig(runner.RunnerConfig{
		ProxyURL: cfg.ProxyURL,
		Insecure: cfg.Insecure,
		Timeout:  cfg.Timeout,
	})
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

// GetFileSummaries returns metadata summaries for all request files in the directory.
func (a *App) GetFileSummaries(dir string) []storage.FileSummary {
	summaries, err := a.storage.ListFileSummaries(dir)
	if err != nil {
		fmt.Printf("GetFileSummaries error: %v\n", err)
		return []storage.FileSummary{}
	}
	return summaries
}

// SaveRequest saves the request to a file
func (a *App) SaveRequest(dir string, filename string, req domain.HttpRequest) (string, error) {
	fullPath := filepath.Join(dir, filename)

	// Attempt to load existing file to preserve metadata
	existingFile, err := a.storage.LoadRequest(fullPath)
	var reqFile domain.RequestFile

	if err == nil {
		// File exists, preserve Meta, update Data
		reqFile = existingFile
		reqFile.Data = req
	} else {
		// New file or error reading (assume new)
		reqFile = domain.RequestFile{
			Meta: domain.MetaData{
				ID:     uuid.New().String(),
				Status: "active",
			},
			Data: req,
		}
	}

	path, err := a.storage.SaveRequest(dir, filename, reqFile)
	if err != nil {
		return "", err
	}
	return path, nil
}

// SaveFullRequest saves a complete RequestFile including its metadata.
func (a *App) SaveFullRequest(dir string, filename string, reqFile domain.RequestFile) (string, error) {
	return a.storage.SaveRequest(dir, filename, reqFile)
}

// LoadRequest loads a request from a file
func (a *App) LoadRequest(dir string, filename string) (domain.RequestFile, error) {
	// Construct full path since storage expects it
	fullPath := filepath.Join(dir, filename)

	req, err := a.storage.LoadRequest(fullPath)
	if err != nil {
		return domain.RequestFile{}, err
	}
	return req, nil
}

// SyncSwagger synchronizes the request files in workDir with the remote Swagger URL
func (a *App) SyncSwagger(workDir string, url string) string {
	result, err := a.syncer.SyncSwagger(context.Background(), workDir, url)
	if err != nil {
		return fmt.Sprintf("Error syncing swagger: %v", err)
	}
	return result
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

// LoadConfig loads a configuration string from a file.
// It specifically checks for .curlflow/ subdirectory if the file is not found in the root.
func (a *App) LoadConfig(dir string, filename string) (string, error) {
	// Try the requested path first
	fullPath := filepath.Join(dir, filename)
	content, err := a.storage.LoadFile(fullPath)
	if err == nil {
		return content, nil
	}

	// If not found and it's a "standard" config file name, try .curlflow/filename
	if os.IsNotExist(err) && (filename == "environments.json") {
		altPath := filepath.Join(dir, ".curlflow", filename)
		content, err = a.storage.LoadFile(altPath)
		if err == nil {
			return content, nil
		}
	}

	return "", err
}

// DeleteFile deletes a request file or config file.
func (a *App) DeleteFile(dir string, filename string) error {
	fullPath := filepath.Join(dir, filename)
	return a.storage.DeleteFile(fullPath)
}

// GetProjectConfig loads the project configuration for the specified directory.
func (a *App) GetProjectConfig(dir string) storage.ProjectConfig {
	config, err := a.storage.LoadProjectConfig(dir)
	if err != nil {
		fmt.Printf("GetProjectConfig error: %v\n", err)
		return storage.ProjectConfig{}
	}
	return config
}

// SaveProjectConfig saves the project configuration (currently just Swagger URL).
func (a *App) SaveProjectConfig(dir string, url string) string {
	config := storage.ProjectConfig{
		SwaggerURL: url,
	}
	err := a.storage.SaveProjectConfig(dir, config)
	if err != nil {
		return fmt.Sprintf("Error saving project config: %v", err)
	}
	return "success"
}

// GetEnvConfig loads the environment configuration for the specified directory.
func (a *App) GetEnvConfig(dir string) storage.EnvConfig {
	config, err := a.storage.LoadEnvConfig(dir)
	if err != nil {
		fmt.Printf("GetEnvConfig error: %v\n", err)
		// Return defaults on error
		return storage.EnvConfig{
			ActiveEnvName: "dev",
			Environments:  make(map[string]storage.EnvVarContainer),
		}
	}
	return config
}

func (a *App) getConfigPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	appDir := filepath.Join(configDir, "curlflow")
	if err := os.MkdirAll(appDir, 0755); err != nil {
		return "", err
	}
	return filepath.Join(appDir, "settings.json"), nil
}

// GetSettings loads the global application configuration.
func (a *App) GetSettings() AppConfig {
	// Default config
	config := AppConfig{
		Timeout: 30,
	}

	path, err := a.getConfigPath()
	if err != nil {
		fmt.Printf("Error getting config path: %v\n", err)
		return config
	}

	data, err := os.ReadFile(path)
	if err != nil {
		// If file doesn't exist, return defaults
		return config
	}

	if err := json.Unmarshal(data, &config); err != nil {
		fmt.Printf("Error parsing settings: %v\n", err)
		return config
	}

	return config
}

// SaveSettings saves the global application configuration and updates the runner.
func (a *App) SaveSettings(cfg AppConfig) string {
	fmt.Printf("DEBUG: SaveSettings called with: %+v\n", cfg)

	path, err := a.getConfigPath()
	if err != nil {
		return fmt.Sprintf("Error getting config path: %v", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Sprintf("Error marshalling config: %v", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Sprintf("Error writing config file: %v", err)
	}

	// Immediate effect: Update runner
	a.runner.UpdateConfig(runner.RunnerConfig{
		ProxyURL: cfg.ProxyURL,
		Insecure: cfg.Insecure,
		Timeout:  cfg.Timeout,
	})

	return "success"
}
