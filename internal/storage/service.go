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
		// Also skip environment configuration file
		if !entry.IsDir() && strings.HasSuffix(lowerName, ".json") && lowerName != "environments.json" {
			files = append(files, name)
		}
	}
	return files, nil
}

// SaveRequest saves the domain.HttpRequest as a formatted JSON file.
// It automatically appends the .json suffix if missing.
func (s *Service) SaveRequest(dirPath string, filename string, req domain.HttpRequest) (string, error) {
	if !strings.HasSuffix(strings.ToLower(filename), ".json") {
		filename += ".json"
	}

	fullPath := filepath.Join(dirPath, filename)

	// Marshal with indentation for readability
	data, err := json.MarshalIndent(req, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal content: %w", err)
	}

	err = os.WriteFile(fullPath, data, 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return fullPath, nil
}

// LoadRequest reads a JSON file and deserializes it into a domain.HttpRequest.
func (s *Service) LoadRequest(filePath string) (domain.HttpRequest, error) {
	var req domain.HttpRequest

	data, err := os.ReadFile(filePath)
	if err != nil {
		return req, fmt.Errorf("failed to read file: %w", err)
	}

	if err := json.Unmarshal(data, &req); err != nil {
		return req, fmt.Errorf("failed to unmarshal request: %w", err)
	}

	return req, nil
}

// SaveFile writes a string content to a file.
func (s *Service) SaveFile(path string, content string) error {
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
