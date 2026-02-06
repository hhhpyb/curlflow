package history

import (
	"curlflow/internal/domain"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
)

// HistoryEntry represents a single executed request record.
type HistoryEntry struct {
	ID         string             `json:"id"`
	ExecutedAt int64              `json:"executed_at"` // Unix timestamp
	Request    domain.HttpRequest `json:"request"`
}

// Service manages the history of executed requests.
type Service struct {
	mu           sync.Mutex
	entries      []HistoryEntry
	maxEntries   int
	workDir      string
	historyCache map[string][]HistoryEntry // Cache per working directory
}

// NewService creates a new History Service.
func NewService() *Service {
	return &Service{
		entries:      make([]HistoryEntry, 0),
		maxEntries:   50,
		historyCache: make(map[string][]HistoryEntry),
	}
}

// getHistoryFilePath returns the path to the history file for the given working directory.
func (s *Service) getHistoryFilePath(workDir string) string {
	return filepath.Join(workDir, ".curlflow", "history.json")
}

// loadHistory loads the history from the file system for the specified directory.
func (s *Service) loadHistory(workDir string) error {
	path := s.getHistoryFilePath(workDir)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			s.entries = []HistoryEntry{}
			return nil
		}
		return err
	}

	var entries []HistoryEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return err
	}

	s.entries = entries
	s.historyCache[workDir] = entries
	return nil
}

// saveHistory saves the current entries to the file system.
func (s *Service) saveHistory(workDir string) error {
	path := s.getHistoryFilePath(workDir)

	// Ensure .curlflow directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(s.entries, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// Add appends a new request to the history.
func (s *Service) Add(workDir string, req domain.HttpRequest) error {
	if workDir == "" {
		return fmt.Errorf("working directory not set")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Ensure we have loaded the latest history for this dir
	// In a real app we might want to optimize this to not reload every time,
	// but for now, let's trust the in-memory cache if we switched dirs properly.
	// However, since Add is async, we should be careful.
	// Simple approach: Use memory state if matches workDir, else load.
	if s.workDir != workDir {
		if cached, ok := s.historyCache[workDir]; ok {
			s.entries = cached
		} else {
			_ = s.loadHistory(workDir)
		}
		s.workDir = workDir
	}

	entry := HistoryEntry{
		ID:         uuid.New().String(),
		ExecutedAt: time.Now().Unix(),
		Request:    req,
	}

	// Prepend
	s.entries = append([]HistoryEntry{entry}, s.entries...)

	// Trim
	if len(s.entries) > s.maxEntries {
		s.entries = s.entries[:s.maxEntries]
	}

	// Update cache
	s.historyCache[workDir] = s.entries

	// Save
	return s.saveHistory(workDir)
}

// List returns the history entries for the specified directory.
func (s *Service) List(workDir string) ([]HistoryEntry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if workDir == "" {
		return []HistoryEntry{}, nil
	}

	if s.workDir != workDir {
		if err := s.loadHistory(workDir); err != nil {
			return nil, err
		}
		s.workDir = workDir
	}

	return s.entries, nil
}

// Clear removes all history entries for the specified directory.
func (s *Service) Clear(workDir string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if workDir == "" {
		return nil
	}

	s.entries = []HistoryEntry{}
	s.historyCache[workDir] = s.entries
	s.workDir = workDir

	path := s.getHistoryFilePath(workDir)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}
