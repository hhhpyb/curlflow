package syncer

import (
	"context"
	"curlflow/internal/domain"
	"curlflow/internal/storage"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi2conv"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/google/uuid"
)

type Service struct {
	storage *storage.Service
}

func NewService(storage *storage.Service) *Service {
	return &Service{storage: storage}
}

type localEntry struct {
	File     *domain.RequestFile
	Filename string
}

// SyncSwagger synchronizes local request files with a remote Swagger definition.
func (s *Service) SyncSwagger(ctx context.Context, dir string, swaggerURL string) (string, error) {
	stats := map[string]int{"new": 0, "updated": 0, "deleted": 0}

	// 1. Build Local Index
	localMap := make(map[string]localEntry)
	existingFiles, err := s.storage.ListRequestFiles(dir)
	if err == nil {
		for _, filename := range existingFiles {
			fullPath := filepath.Join(dir, filename)
			file, err := s.storage.LoadRequest(fullPath)
			if err == nil {
				key := file.Meta.Key
				if key == "" {
					key = fmt.Sprintf("%s_%s", strings.ToUpper(file.Data.Method), file.Data.URL)
				}
				localMap[key] = localEntry{File: &file, Filename: filename}
			}
		}
	}

	// 2. Download and Parse (Smart v2/v3)
	doc, err := s.loadSwaggerSmart(ctx, swaggerURL)
	if err != nil {
		return "", fmt.Errorf("load swagger failed: %v", err)
	}

	processedKeys := make(map[string]bool)
	now := time.Now().Unix()

	// 3. Diff & Merge Loop
	if doc.Paths != nil {
		for path, pathItem := range doc.Paths.Map() {
			for method, op := range pathItem.Operations() {
				method = strings.ToUpper(method)
				currentKey := fmt.Sprintf("%s_%s", method, path)
				processedKeys[currentKey] = true

				finalURL := path
				if !strings.HasPrefix(path, "http") {
					finalURL = fmt.Sprintf("{{base_url}}%s", path)
				}

				summary := op.Summary
				if summary == "" {
					summary = op.OperationID
				}
				if len(summary) > 100 {
					summary = summary[:97] + "..."
				}

				if entry, exists := localMap[currentKey]; exists {
					stats["updated"]++
					existingFile := entry.File
					existingFile.Meta.Status = "active"
					existingFile.Meta.Summary = summary
					existingFile.Meta.LastSyncedAt = now
					existingFile.Meta.SwaggerPath = path
					existingFile.Meta.Tags = op.Tags
					existingFile.Meta.Key = currentKey
					existingFile.Data.Method = method
					existingFile.Data.URL = finalURL

					if existingFile.Data.Headers == nil {
						existingFile.Data.Headers = make(map[string]string)
					}
					s.storage.SaveRequest(dir, entry.Filename, *existingFile)
				} else {
					stats["new"]++
					bodyContent := "{}"
					if op.RequestBody != nil {
						bodyContent = "{\n  \"_note\": \"Check Swagger definition for body\"\n}"
					}

					newFile := domain.RequestFile{
						Meta: domain.MetaData{
							ID:           uuid.New().String(),
							Key:          currentKey,
							Status:       "active",
							Summary:      summary,
							Tags:         op.Tags,
							SwaggerPath:  path,
							LastSyncedAt: now,
						},
						Data: domain.HttpRequest{
							Method:  method,
							URL:     finalURL,
							Headers: map[string]string{"Content-Type": "application/json"},
							Body:    bodyContent,
						},
					}
					filename := s.sanitizeFilename(method + "_" + path)
					s.storage.SaveRequest(dir, filename, newFile)
				}
			}
		}
	}

	// 4. Soft Delete
	for key, entry := range localMap {
		if !processedKeys[key] && entry.File.Meta.Status != "deleted" {
			stats["deleted"]++
			entry.File.Meta.Status = "deleted"
			entry.File.Meta.LastSyncedAt = now
			s.storage.SaveRequest(dir, entry.Filename, *entry.File)
		}
	}

	return fmt.Sprintf("Sync Complete: %d New, %d Updated, %d Marked Deleted", stats["new"], stats["updated"], stats["deleted"]), nil
}

// loadSwaggerSmart handles both Swagger 2.0 and OpenAPI 3.0
func (s *Service) loadSwaggerSmart(ctx context.Context, uri string) (*openapi3.T, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequestWithContext(ctx, "GET", uri, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	rawBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Sanitize common bad data from Springfox
	jsonStr := string(rawBytes)
	jsonStr = strings.ReplaceAll(jsonStr, "«", "_")
	jsonStr = strings.ReplaceAll(jsonStr, "»", "_")
	jsonStr = strings.ReplaceAll(jsonStr, "对象", "_Object")
	cleanBytes := []byte(jsonStr)

	// Try Swagger 2.0 first
	var docV2 openapi2.T
	if err := json.Unmarshal(cleanBytes, &docV2); err == nil && docV2.Swagger == "2.0" {
		docV3, errConv := openapi2conv.ToV3(&docV2)
		if errConv == nil {
			return docV3, nil
		}
	}

	// Fallback to OpenAPI 3.0
	loaderV3 := openapi3.NewLoader()
	loaderV3.IsExternalRefsAllowed = true
	docV3, errV3 := loaderV3.LoadFromData(cleanBytes)
	if errV3 != nil {
		return nil, fmt.Errorf("failed to parse (v2/v3). V3 error: %v", errV3)
	}

	return docV3, nil
}

func (s *Service) sanitizeFilename(name string) string {
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, "{", "")
	name = strings.ReplaceAll(name, "}", "")
	reg := regexp.MustCompile(`[^a-zA-Z0-9_\-]`)
	return reg.ReplaceAllString(name, "")
}
