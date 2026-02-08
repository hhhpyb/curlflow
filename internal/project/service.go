package project

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

// ProjectRegistry stores the list of recent projects and the last opened one.
type ProjectRegistry struct {
	RecentProjects []ProjectEntry `json:"recent_projects"`
	LastOpened     string         `json:"last_opened"`
}

type ProjectEntry struct {
	Path     string    `json:"path"`
	LastUsed time.Time `json:"last_used"`
}

type Service struct {
	configPath string
	mu         sync.RWMutex
	registry   ProjectRegistry
}

// NewService creates a new instance of the project Service.
func NewService() *Service {
	configDir, err := os.UserConfigDir()
	if err != nil {
		fmt.Printf("Error getting user config dir: %v\n", err)
		return &Service{}
	}

	appDir := filepath.Join(configDir, "curlflow")
	if err := os.MkdirAll(appDir, 0755); err != nil {
		fmt.Printf("Error creating app config dir: %v\n", err)
		return &Service{}
	}

	s := &Service{
		configPath: filepath.Join(appDir, "projects.json"),
		registry: ProjectRegistry{
			RecentProjects: []ProjectEntry{},
		},
	}
	s.load()
	return s
}

func (s *Service) load() {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.configPath)
	if err != nil {
		// If file doesn't exist, just ignore
		return
	}

	if err := json.Unmarshal(data, &s.registry); err != nil {
		fmt.Printf("Error parsing projects.json: %v\n", err)
	}
}

func (s *Service) save() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := json.MarshalIndent(s.registry, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.configPath, data, 0644)
}

// AddProject adds a project path to the registry or updates its timestamp.
// It also sets it as the LastOpened project.
func (s *Service) AddProject(path string) {
	if path == "" {
		return
	}
	s.mu.Lock()

	// Remove existing entry if present to update timestamp and move to top (conceptually)
	// We will resort by time anyway.
	found := false
	for i, p := range s.registry.RecentProjects {
		if p.Path == path {
			s.registry.RecentProjects[i].LastUsed = time.Now()
			found = true
			break
		}
	}

	if !found {
		s.registry.RecentProjects = append(s.registry.RecentProjects, ProjectEntry{
			Path:     path,
			LastUsed: time.Now(),
		})
	}

	s.registry.LastOpened = path
	s.mu.Unlock()

	s.save()
}

// RemoveProject removes a project from the registry.
func (s *Service) RemoveProject(path string) {
	s.mu.Lock()
	newProjects := []ProjectEntry{}
	for _, p := range s.registry.RecentProjects {
		if p.Path != path {
			newProjects = append(newProjects, p)
		}
	}
	s.registry.RecentProjects = newProjects

	if s.registry.LastOpened == path {
		s.registry.LastOpened = ""
	}
	s.mu.Unlock()

	s.save()
}

// GetRecentProjects returns the list of projects sorted by LastUsed (descending).
func (s *Service) GetRecentProjects() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Clone to avoid data race during sort
	projects := make([]ProjectEntry, len(s.registry.RecentProjects))
	copy(projects, s.registry.RecentProjects)

	sort.Slice(projects, func(i, j int) bool {
		return projects[i].LastUsed.After(projects[j].LastUsed)
	})

	paths := make([]string, len(projects))
	for i, p := range projects {
		paths[i] = p.Path
	}
	return paths
}

// GetLastOpened returns the last opened project path.
func (s *Service) GetLastOpened() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.registry.LastOpened
}
