package storage

import (
	"context"
	"curlflow/internal/domain"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// Service handles file system operations for persistent storage of requests.
type Service struct{}

type FileSummary struct {
	FileName string          `json:"fileName"`
	Meta     domain.MetaData `json:"meta"`
}

// NewService creates a new instance of the storage Service.
func NewService() *Service {
	return &Service{}
}

// SelectWorkingDirectory opens a directory selection dialog using Wails runtime.
func (s *Service) SelectWorkingDirectory(ctx context.Context) (string, error) {
	options := runtime.OpenDialogOptions{
		Title: "Select Working Directory",
	}
	dir, err := runtime.OpenDirectoryDialog(ctx, options)
	if err != nil {
		return "", fmt.Errorf("failed to open directory dialog: %w", err)
	}
	return dir, nil
}

// ListRequestFiles lists all .json files in the specified directory.
func (s *Service) ListRequestFiles(dirPath string) ([]string, error) {
	var files []string
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	for _, entry := range entries {
		name := entry.Name()
		lowerName := strings.ToLower(name)
		// Filter for .json files (case-insensitive) and ignore directories
		// Also skip environment configuration file (backward compatibility) and settings.json
		if !entry.IsDir() && strings.HasSuffix(lowerName, ".json") && lowerName != "environments.json" && lowerName != "settings.json" {
			files = append(files, name)
		}
	}
	return files, nil
}

// ListFileSummaries lists metadata for all .json files in the specified directory.
func (s *Service) ListFileSummaries(dirPath string) ([]FileSummary, error) {
	var summaries []FileSummary
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	for _, entry := range entries {
		name := entry.Name()
		lowerName := strings.ToLower(name)

		if entry.IsDir() || !strings.HasSuffix(lowerName, ".json") ||
			lowerName == "environments.json" || lowerName == "settings.json" {
			continue
		}

		filePath := filepath.Join(dirPath, name)
		file, err := os.Open(filePath)
		if err != nil {
			continue
		}

		// Use a local struct to only decode the metadata part
		var partial struct {
			Meta domain.MetaData `json:"_meta"`
		}

		decoder := json.NewDecoder(file)
		err = decoder.Decode(&partial)
		file.Close()

		if err == nil {
			summaries = append(summaries, FileSummary{
				FileName: name,
				Meta:     partial.Meta,
			})
		}
	}

	return summaries, nil
}

// SaveRequest saves the domain.RequestFile as a formatted JSON file.
// It automatically appends the .json suffix if missing.
func (s *Service) SaveRequest(dirPath string, filename string, reqFile domain.RequestFile) (string, error) {
	if !strings.HasSuffix(strings.ToLower(filename), ".json") {
		filename += ".json"
	}

	fullPath := filepath.Join(dirPath, filename)

	// Marshal with indentation for readability
	data, err := json.MarshalIndent(reqFile, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal content: %w", err)
	}

	err = os.WriteFile(fullPath, data, 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return fullPath, nil
}

// LoadRequest reads a JSON file and deserializes it into a domain.RequestFile.
func (s *Service) LoadRequest(filePath string) (domain.RequestFile, error) {
	var reqFile domain.RequestFile

	data, err := os.ReadFile(filePath)
	if err != nil {
		return reqFile, fmt.Errorf("failed to read file: %w", err)
	}

	if err := json.Unmarshal(data, &reqFile); err != nil {
		return reqFile, fmt.Errorf("failed to unmarshal request: %w", err)
	}

	return reqFile, nil
}

// SaveFile writes a string content to a file. It ensures the parent directory exists.
func (s *Service) SaveFile(path string, content string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	return os.WriteFile(path, []byte(content), 0644)
}

// LoadFile reads a file and returns its content as a string.
func (s *Service) LoadFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}
	return string(data), nil
}

// DeleteFile removes a file from the file system.
func (s *Service) DeleteFile(path string) error {
	return os.Remove(path)
}

type ProjectConfig struct {
	SwaggerURL string `json:"swagger_url"`
	// 未来可扩展其他字段
}

// LoadProjectConfig reads the project configuration from dir/.curlflow/project.json.
func (s *Service) LoadProjectConfig(dir string) (ProjectConfig, error) {
	config := ProjectConfig{}
	configPath := filepath.Join(dir, ".curlflow", "project.json")

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return config, nil // Return empty config if file doesn't exist
		}
		return config, fmt.Errorf("failed to read project config: %w", err)
	}

	if err := json.Unmarshal(data, &config); err != nil {
		return config, fmt.Errorf("failed to unmarshal project config: %w", err)
	}

	return config, nil
}

// SaveProjectConfig saves the project configuration to dir/.curlflow/project.json.
func (s *Service) SaveProjectConfig(dir string, config ProjectConfig) error {
	projectDir := filepath.Join(dir, ".curlflow")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		return fmt.Errorf("failed to create project directory: %w", err)
	}

	configPath := filepath.Join(projectDir, "project.json")
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal project config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write project config file: %w", err)
	}

	return nil
}
