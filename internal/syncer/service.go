package syncer

import (
	"context"
	"curlflow/internal/domain"
	"curlflow/internal/storage"
	"encoding/json"
	"fmt"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/google/uuid"
)

type Service struct {
	storage *storage.Service
}

func NewService(storage *storage.Service) *Service {
	return &Service{storage: storage}
}

// Stats holds the synchronization results
type Stats struct {
	Added   int
	Updated int
	Deleted int
	Total   int
}

// SyncSwagger synchronizes local request files with a remote Swagger definition.
func (s *Service) SyncSwagger(ctx context.Context, dir string, swaggerURL string) (map[string]int, error) {
	stats := map[string]int{
		"added":   0,
		"updated": 0,
		"deleted": 0,
		"total":   0,
	}

	// 1. Build Local Index
	localMap := make(map[string]*domain.RequestFile)
	pathMap := make(map[string]string) // Key -> Filename

	files, err := s.storage.ListRequestFiles(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to list request files: %w", err)
	}

	for _, filename := range files {
		fullPath := filepath.Join(dir, filename)
		reqFile, err := s.storage.LoadRequest(fullPath)
		if err != nil {
			// Skip malformed files, log error?
			fmt.Printf("Warning: failed to load %s: %v\n", filename, err)
			continue
		}

		// Ensure Key exists
		key := reqFile.Meta.Key
		if key == "" {
			// Generate temporary key for legacy files
			key = generateKey(reqFile.Data.Method, getPathFromURL(reqFile.Data.URL))
			// Note: We don't save back yet, we wait for match or update
		}

		localMap[key] = &reqFile
		pathMap[key] = filename
	}

	// 2. Parse Remote Swagger
	u, err := url.Parse(swaggerURL)
	if err != nil {
		return nil, fmt.Errorf("invalid swagger URL: %w", err)
	}

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	doc, err := loader.LoadFromURI(u)
	if err != nil {
		return nil, fmt.Errorf("failed to load swagger from %s: %w", swaggerURL, err)
	}

	// Track processed keys to identify deletions later
	processedKeys := make(map[string]bool)
	now := time.Now().Unix()

	// 3. Diff & Merge Loop
	if doc.Paths != nil {
		for path, pathItem := range doc.Paths.Map() {
			for method, operation := range pathItem.Operations() {
				method = strings.ToUpper(method)
				currentKey := generateKey(method, path)
				processedKeys[currentKey] = true
				stats["total"]++

				// Prepare Swagger Data
				summary := operation.Summary
				if summary == "" {
					summary = operation.Description
				}
				// Truncate summary if too long
				if len(summary) > 100 {
					summary = summary[:97] + "..."
				}

				// Extract tags
				tags := operation.Tags

				// Scenario A: Update Existing
				if existingFile, exists := localMap[currentKey]; exists {
					// Meta Update
					existingFile.Meta.Status = "active"
					existingFile.Meta.Summary = summary
					existingFile.Meta.LastSyncedAt = now
					existingFile.Meta.SwaggerPath = path
					existingFile.Meta.Tags = tags
					// If key was missing in file, ensure it's set
					existingFile.Meta.Key = currentKey

					// Data Protection Merge

					// Base URL logic
					baseURL := getBaseURL(doc, existingFile.Data.URL)
					fullURL := baseURL + path
					existingFile.Data.URL = fullURL
					existingFile.Data.Method = method

					// Merge Headers
					if existingFile.Data.Headers == nil {
						existingFile.Data.Headers = make(map[string]string)
					}

					for _, paramRef := range operation.Parameters {
						if paramRef.Value != nil && strings.ToLower(paramRef.Value.In) == "header" {
							headerName := paramRef.Value.Name
							// Only add if missing
							if _, ok := existingFile.Data.Headers[headerName]; !ok {
								existingFile.Data.Headers[headerName] = "" // Empty value
							}
						}
					}

					// Body
					if existingFile.Data.Body == "" && operation.RequestBody != nil && operation.RequestBody.Value != nil {
						// Try to generate example body
						example := getExampleBody(operation.RequestBody.Value)
						if example != "" {
							existingFile.Data.Body = example
						}
					}

					// Save
					filename := pathMap[currentKey]
					_, err := s.storage.SaveRequest(dir, filename, *existingFile)
					if err != nil {
						fmt.Printf("Error saving updated file %s: %v\n", filename, err)
					} else {
						stats["updated"]++
					}

				} else {
					// Scenario B: New File
					newReq := domain.RequestFile{
						Meta: domain.MetaData{
							ID:           uuid.New().String(),
							Key:          currentKey,
							Status:       "new",
							Summary:      summary,
							Tags:         tags,
							SwaggerPath:  path,
							LastSyncedAt: now,
						},
						Data: domain.HttpRequest{
							Method:  method,
							Headers: make(map[string]string),
						},
					}

					// Base URL
					baseURL := getBaseURL(doc, "")
					newReq.Data.URL = baseURL + path

					// Headers
					for _, paramRef := range operation.Parameters {
						if paramRef.Value != nil && strings.ToLower(paramRef.Value.In) == "header" {
							newReq.Data.Headers[paramRef.Value.Name] = ""
						}
					}

					// Body
					if operation.RequestBody != nil && operation.RequestBody.Value != nil {
						example := getExampleBody(operation.RequestBody.Value)
						if example != "" {
							newReq.Data.Body = example
						}
					}

					// Generate Filename
					filename := sanitizeFilename(method + "_" + path)

					_, err := s.storage.SaveRequest(dir, filename, newReq)
					if err != nil {
						fmt.Printf("Error creating new file %s: %v\n", filename, err)
					} else {
						stats["added"]++
					}
				}
			}
		}
	}

	// 4. Soft Delete
	for key, reqFile := range localMap {
		if !processedKeys[key] {
			if reqFile.Meta.Status != "deleted" {
				reqFile.Meta.Status = "deleted"
				reqFile.Meta.LastSyncedAt = now

				filename := pathMap[key]
				_, err := s.storage.SaveRequest(dir, filename, *reqFile)
				if err != nil {
					fmt.Printf("Error soft-deleting file %s: %v\n", filename, err)
				} else {
					stats["deleted"]++
				}
			}
		}
	}

	return stats, nil
}

// Helpers

func generateKey(method, path string) string {
	return strings.ToUpper(method) + "_" + path
}

func getPathFromURL(u string) string {
	parsed, err := url.Parse(u)
	if err == nil {
		return parsed.Path
	}
	// Fallback: try to split manually if parse fails (unlikely for valid urls)
	return u
}

func getBaseURL(doc *openapi3.T, fallbackURL string) string {
	if len(doc.Servers) > 0 {
		return doc.Servers[0].URL
	}
	if fallbackURL != "" {
		parsed, err := url.Parse(fallbackURL)
		if err == nil {
			return parsed.Scheme + "://" + parsed.Host
		}
	}
	return "http://localhost" // Default fallback
}

func getExampleBody(body *openapi3.RequestBody) string {
	content := body.Content.Get("application/json")
	if content != nil {
		if content.Example != nil {
			if s, ok := content.Example.(string); ok {
				return s
			}
			b, err := json.MarshalIndent(content.Example, "", "  ")
			if err == nil {
				return string(b)
			}
		}
	}
	return ""
}

func sanitizeFilename(name string) string {
	// Replace / with _
	name = strings.ReplaceAll(name, "/", "_")
	// Replace { } with -
	name = strings.ReplaceAll(name, "{", "")
	name = strings.ReplaceAll(name, "}", "")

	// Remove other illegal chars
	reg := regexp.MustCompile(`[^a-zA-Z0-9_\-]`)
	name = reg.ReplaceAllString(name, "")

	return name
}
