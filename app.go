package main

import (
	"context"
	"curlflow/internal/domain"
	"curlflow/internal/history"
	"curlflow/internal/parser"
	"curlflow/internal/project"
	"curlflow/internal/runner"
	"curlflow/internal/storage"
	"curlflow/internal/syncer"
	"curlflow/internal/websocket"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/google/uuid"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx       context.Context
	runner    *runner.Service
	storage   *storage.Service
	syncer    *syncer.Service
	history   *history.Service
	project   *project.Service
	wsManager *websocket.Manager
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
		runner:    runner.NewService(),
		storage:   storageService,
		syncer:    syncer.NewService(storageService),
		history:   history.NewService(),
		project:   project.NewService(),
		wsManager: websocket.NewManager(),
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.wsManager.SetContext(ctx)
	// Load settings on startup and apply to runner
	cfg := a.GetSettings()
	a.runner.UpdateConfig(runner.RunnerConfig{
		ProxyURL: cfg.ProxyURL,
		Insecure: cfg.Insecure,
		Timeout:  cfg.Timeout,
	})

	// Auto-load last project
	lastProject := a.project.GetLastOpened()
	if lastProject != "" {
		// Emit event to frontend to set workDir
		// We use a slight delay or just emit. Frontend listens for this?
		// Actually, frontend calls init() which checks localStorage.
		// We should probably rely on the frontend asking "GetLastOpened" OR
		// just let the frontend know via event.
		//
		// Better approach: Since frontend currently uses localStorage ('curlflow_workDir'),
		// we can keep that for now, OR we can override it if we want the backend to be the source of truth.
		// Let's defer to the user's request: "App starts, prioritize last_opened_path".

		// Note: We cannot easily "push" to frontend before frontend is ready.
		// Wails 'domready' event is when frontend is ready.
		// Instead, we will expose a method GetLastOpenedProject() and let frontend call it in init().
	}
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
// Note: workDir is now required to save history
func (a *App) SendRequest(reqFile domain.RequestFile, workDir string, filename string) domain.HttpResponse {
	req := reqFile.Data

	// Resolve Auth if Inherit
	if req.Auth.Type == domain.AuthTypeInherit {
		var parentAuths []domain.Auth

		// 1. Project Auth (Global Parent)
		// Use LoadProjectConfig from storage service directly
		projConfig, err := a.storage.LoadProjectConfig(workDir)
		if err == nil && projConfig.Auth.Type != "" {
			parentAuths = append(parentAuths, projConfig.Auth)
		}

		// 2. Main File Auth (Folder/Group Parent)
		// Only if we have an ID and we are not the main file itself
		if reqFile.Meta.ID != "" {
			summaries, err := a.storage.ListFileSummaries(workDir)
			if err == nil {
				// Filter by ID
				var group []storage.FileSummary
				for _, s := range summaries {
					if s.Meta.ID == reqFile.Meta.ID {
						group = append(group, s)
					}
				}

				// Find Main File (shortest name)
				if len(group) > 0 {
					sort.Slice(group, func(i, j int) bool {
						return len(group[i].FileName) < len(group[j].FileName)
					})
					mainFileSummary := group[0]

					// If we are NOT the main file, we inherit from it
					// If filename is empty (new unsaved), we assume we are not the main file unless... wait.
					// If new unsaved request has generated ID, it won't match.
					// If we are editing a case (saved), filename matches.
					if filename != mainFileSummary.FileName {
						// Load the main file to get its full Auth config
						mainReqFile, err := a.storage.LoadRequest(filepath.Join(workDir, mainFileSummary.FileName))
						if err == nil {
							// Prepend to parentAuths so it takes precedence over Project
							parentAuths = append([]domain.Auth{mainReqFile.Data.Auth}, parentAuths...)
						}
					}
				}
			}
		}

		// Resolve
		a.runner.ResolveAuth(&req, parentAuths)

		// Resolve Proxy if configured in project
		projConfig, err = a.storage.LoadProjectConfig(workDir)
		if err == nil && projConfig.ProxyURL != "" {
			// Temporary override for this request if we had a way,
			// but SendRequest uses the service's internal config.
			// For now, let's stick to the UI implementation as requested.
			// We can refactor the runner to accept options later if needed.
		}
	}

	// Async save to history
	if workDir != "" {
		go func() {
			err := a.history.Add(workDir, req)
			if err != nil {
				fmt.Printf("Failed to save history: %v\n", err)
			} else {
				// Emit event to frontend to refresh history list
				runtime.EventsEmit(a.ctx, "history_updated")
			}
		}()
	}

	return a.runner.SendRequest(req)
}

// GetHistoryList returns the history for the current working directory
func (a *App) GetHistoryList(workDir string) []history.HistoryEntry {
	list, err := a.history.List(workDir)
	if err != nil {
		fmt.Printf("GetHistoryList error: %v\n", err)
		return []history.HistoryEntry{}
	}
	return list
}

// ClearHistory clears the history for the current working directory
func (a *App) ClearHistory(workDir string) string {
	err := a.history.Clear(workDir)
	if err != nil {
		return fmt.Sprintf("Error clearing history: %v", err)
	}
	return "success"
}

// SelectWorkDir opens a directory selection dialog
func (a *App) SelectWorkDir() string {
	dir, err := a.storage.SelectWorkingDirectory(a.ctx)
	if err != nil {
		fmt.Printf("SelectWorkDir error: %v\n", err)
		return ""
	}
	if dir != "" {
		a.project.AddProject(dir)
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

// PurgeDeletedFiles physically removes all files marked as "deleted" in the directory.
func (a *App) PurgeDeletedFiles(dir string) string {
	count, err := a.storage.PurgeDeletedFiles(dir)
	if err != nil {
		return fmt.Sprintf("Error purging files: %v", err)
	}
	return fmt.Sprintf("Successfully purged %d files", count)
}

// GetProjectConfig loads the project configuration for the specified directory.
func (a *App) GetProjectConfig(dir string) domain.ProjectConfig {
	config, err := a.storage.LoadProjectConfig(dir)
	if err != nil {
		fmt.Printf("GetProjectConfig error: %v\n", err)
		return domain.ProjectConfig{}
	}
	return config
}

// SaveProjectConfig saves the project configuration.
func (a *App) SaveProjectConfig(dir string, config domain.ProjectConfig) string {
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

// GetRecentProjects returns the list of recently opened projects
func (a *App) GetRecentProjects() []string {
	return a.project.GetRecentProjects()
}

// RemoveProject removes a project from the recent list
func (a *App) RemoveProject(path string) {
	a.project.RemoveProject(path)
}

// OpenProject explicitly adds a path to the recent list (used when switching projects from UI)
func (a *App) OpenProject(path string) {
	a.project.AddProject(path)
}

// GetLastOpenedProject returns the last opened project path
func (a *App) GetLastOpenedProject() string {
	return a.project.GetLastOpened()
}

func (a *App) beforeClose(ctx context.Context) (prevent bool) {
	a.wsManager.CloseAll()
	return false
}

// WebSocket Methods

// WsConnect establishes a new WebSocket connection
func (a *App) WsConnect(sessionID string, url string, headers map[string]string) string {
	err := a.wsManager.Connect(sessionID, url, headers)
	if err != nil {
		return err.Error()
	}
	return "success"
}

// WsDisconnect closes a WebSocket connection
func (a *App) WsDisconnect(sessionID string) string {
	a.wsManager.Disconnect(sessionID)
	return "success"
}

// WsSend sends a message to a WebSocket connection
func (a *App) WsSend(sessionID string, message string) string {
	err := a.wsManager.SendMessage(sessionID, message)
	if err != nil {
		return err.Error()
	}
	return "success"
}
