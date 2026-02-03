package file

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// Manager handles file system operations for the application.
type Manager struct{}

// NewManager creates a new instance of the file Manager.
func NewManager() *Manager {
	return &Manager{}
}

// SelectWorkingDirectory opens a directory selection dialog using Wails runtime.
func (m *Manager) SelectWorkingDirectory(ctx context.Context) (string, error) {
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
func (m *Manager) ListRequestFiles(dirPath string) ([]string, error) {
	var files []string
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	for _, entry := range entries {
		// Filter for .json files (case-insensitive) and ignore directories
		if !entry.IsDir() && strings.HasSuffix(strings.ToLower(entry.Name()), ".json") {
			files = append(files, entry.Name())
		}
	}
	return files, nil
}

// SaveRequestFile saves the content as a formatted JSON file in the specified directory.
// It automatically appends the .json suffix if missing.
func (m *Manager) SaveRequestFile(dirPath string, filename string, content interface{}) (string, error) {
	if !strings.HasSuffix(strings.ToLower(filename), ".json") {
		filename += ".json"
	}

	fullPath := filepath.Join(dirPath, filename)

	// Marshal with indentation for readability
	data, err := json.MarshalIndent(content, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal content: %w", err)
	}

	err = os.WriteFile(fullPath, data, 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return fullPath, nil
}

// LoadRequestFile reads the content of a file and returns the raw bytes.
func (m *Manager) LoadRequestFile(filePath string) ([]byte, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	return data, nil
}
