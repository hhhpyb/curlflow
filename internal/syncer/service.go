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
					// Fallback for older files without key
					key = fmt.Sprintf("%s_%s", strings.ToUpper(file.Data.Method), file.Meta.SwaggerPath)
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

				summary := op.Summary
				if summary == "" {
					summary = op.OperationID
				}
				if len(summary) > 100 {
					summary = summary[:97] + "..."
				}

				// Extract parameter docs
				paramDocs := make(map[string]string)
				for _, paramRef := range op.Parameters {
					p := paramRef.Value
					if p != nil {
						docKey := p.Name
						if p.In != "" {
							docKey = p.In + "." + p.Name
						}
						if p.Description != "" {
							paramDocs[docKey] = p.Description
						}
					}
				}

				// Generate Body and extract body property docs
				generatedBody := s.generateBody(op, paramDocs)

				if entry, exists := localMap[currentKey]; exists {
					stats["updated"]++
					existingFile := entry.File

					// Update Metadata
					existingFile.Meta.Status = "active"
					existingFile.Meta.Summary = summary
					existingFile.Meta.Description = op.Description
					existingFile.Meta.LastSyncedAt = now
					existingFile.Meta.SwaggerPath = path
					existingFile.Meta.Tags = op.Tags
					existingFile.Meta.Key = currentKey
					existingFile.Meta.ParamDocs = paramDocs
					existingFile.Meta.Source = "swagger" // 确保标记为 Swagger 来源

					// Update Data - Headers (Merge)
					if existingFile.Data.Headers == nil {
						existingFile.Data.Headers = make(map[string]string)
					}
					for _, paramRef := range op.Parameters {
						p := paramRef.Value
						if p != nil && p.In == "header" {
							if _, ok := existingFile.Data.Headers[p.Name]; !ok {
								existingFile.Data.Headers[p.Name] = ""
							}
						}
					}

					// Update Data - Body (only if empty or default)
					trimmedBody := strings.TrimSpace(existingFile.Data.Body)
					if trimmedBody == "" || trimmedBody == "{}" {
						existingFile.Data.Body = generatedBody
					}

					// Note: URL with Query is NOT updated to preserve user edits
					s.storage.SaveRequest(dir, entry.Filename, *existingFile)
				} else {
					stats["new"]++

					// Build URL with Query for NEW requests
					finalURL := s.buildURLWithQuery(path, op)

					newFile := domain.RequestFile{
						Meta: domain.MetaData{
							ID:           uuid.New().String(),
							Key:          currentKey,
							Status:       "active",
							Summary:      summary,
							Description:  op.Description,
							Tags:         op.Tags,
							SwaggerPath:  path,
							LastSyncedAt: now,
							ParamDocs:    paramDocs,
							Source:       "swagger",
						},
						Data: domain.HttpRequest{
							Method:  method,
							URL:     finalURL,
							Headers: map[string]string{"Content-Type": "application/json"},
							Body:    generatedBody,
						},
					}

					// Initial Headers
					for _, paramRef := range op.Parameters {
						p := paramRef.Value
						if p != nil && p.In == "header" {
							newFile.Data.Headers[p.Name] = ""
						}
					}

					filename := s.sanitizeFilename(method + "_" + path)
					s.storage.SaveRequest(dir, filename, newFile)
				}
			}
		}
	}

	// 4. Soft Delete
	for key, entry := range localMap {
		if !processedKeys[key] {
			// 【核心保护】如果是用户手动创建的文件，或者是旧的非Swagger文件，跳过删除
			if entry.File.Meta.Source == "user" {
				continue
			}
			// 兼容旧数据：如果 Source 为空，且 SwaggerPath 为空，也认为是用户手动创建的
			if entry.File.Meta.Source == "" && entry.File.Meta.SwaggerPath == "" {
				continue
			}

			if entry.File.Meta.Status != "deleted" {
				stats["deleted"]++
				entry.File.Meta.Status = "deleted"
				entry.File.Meta.LastSyncedAt = now
				s.storage.SaveRequest(dir, entry.Filename, *entry.File)
			}
		}
	}

	return fmt.Sprintf("同步完成: 新增 %d, 更新 %d, 已删除(标记) %d", stats["new"], stats["updated"], stats["deleted"]), nil
}

// buildURLWithQuery constructs URL with base_url placeholder and query parameters.
func (s *Service) buildURLWithQuery(path string, op *openapi3.Operation) string {
	finalURL := path
	if !strings.HasPrefix(path, "http") {
		finalURL = fmt.Sprintf("{{base_url}}%s", path)
	}

	var queries []string
	for _, paramRef := range op.Parameters {
		param := paramRef.Value
		if param != nil && param.In == "query" {
			val := ""
			if param.Schema != nil && param.Schema.Value != nil {
				v := param.Schema.Value
				if v.Default != nil {
					val = fmt.Sprintf("%v", v.Default)
				} else if v.Example != nil {
					val = fmt.Sprintf("%v", v.Example)
				}
			}
			queries = append(queries, fmt.Sprintf("%s=%s", param.Name, val))
		}
	}

	if len(queries) > 0 {
		connector := "?"
		if strings.Contains(finalURL, "?") {
			connector = "&"
		}
		finalURL += connector + strings.Join(queries, "&")
	}
	return finalURL
}

// generateBody parses request body schema and returns a JSON string.
func (s *Service) generateBody(op *openapi3.Operation, paramDocs map[string]string) string {
	if op.RequestBody == nil || op.RequestBody.Value == nil {
		return ""
	}
	content := op.RequestBody.Value.Content.Get("application/json")
	if content == nil || content.Schema == nil || content.Schema.Value == nil {
		return ""
	}

	schema := content.Schema.Value

	// Collect property descriptions into paramDocs (with depth limit)
	s.collectPropertyDocs("body", schema, paramDocs, 0)

	// Convert schema to JSON (with depth limit)
	data := s.schemaToJSON("", schema, 0)
	if data == nil {
		return "{}"
	}

	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "{}"
	}
	return string(bytes)
}

func (s *Service) collectPropertyDocs(prefix string, schema *openapi3.Schema, docs map[string]string, depth int) {
	if schema == nil || depth > 10 {
		return
	}
	for name, propRef := range schema.Properties {
		if propRef.Value != nil {
			fullKey := prefix + "." + name
			if propRef.Value.Description != "" {
				docs[fullKey] = propRef.Value.Description
			}
			if propRef.Value.Type.Is(openapi3.TypeObject) {
				s.collectPropertyDocs(fullKey, propRef.Value, docs, depth+1)
			}
		}
	}
}

func (s *Service) schemaToJSON(propName string, schema *openapi3.Schema, depth int) interface{} {
	if schema == nil || depth > 10 {
		if depth > 10 {
			return "<max depth reached>"
		}
		return nil
	}

	// 1. Schema Example & Default (Highest Priority)
	if schema.Example != nil {
		return schema.Example
	}
	if schema.Default != nil {
		return schema.Default
	}

	// In kin-openapi v0.123+, Type is *Types (a slice of strings)
	schemaType := ""
	if schema.Type != nil && len(*schema.Type) > 0 {
		schemaType = (*schema.Type)[0]
	}

	lowerName := strings.ToLower(propName)

	switch schemaType {
	case openapi3.TypeObject:
		obj := make(map[string]interface{})
		for name, propRef := range schema.Properties {
			if propRef.Value != nil {
				obj[name] = s.schemaToJSON(name, propRef.Value, depth+1)
			}
		}
		return obj
	case openapi3.TypeArray:
		if schema.Items != nil && schema.Items.Value != nil {
			// Pass empty name for array items as they don't have individual keys
			return []interface{}{s.schemaToJSON("", schema.Items.Value, depth+1)}
		}
		return []interface{}{}
	case openapi3.TypeString:
		// 2. Format
		if schema.Format == "date-time" {
			return time.Now().Format("2006-01-02 15:04:05")
		}
		if schema.Format == "date" {
			return time.Now().Format("2006-01-02")
		}
		if schema.Format == "email" {
			return "example@test.com"
		}
		if schema.Format == "uuid" {
			return "550e8400-e29b-41d4-a716-446655440000"
		}
		if schema.Format == "uri" || schema.Format == "url" {
			return "https://example.com"
		}

		// 3. Name Heuristics (String)
		if strings.Contains(lowerName, "time") || strings.Contains(lowerName, "date") || strings.Contains(lowerName, "at") {
			return "2024-01-01 12:00:00"
		}
		if strings.Contains(lowerName, "id") || strings.Contains(lowerName, "code") || strings.Contains(lowerName, "no") {
			return "1001"
		}
		if strings.Contains(lowerName, "name") || strings.Contains(lowerName, "user") {
			return "John Doe"
		}
		if strings.Contains(lowerName, "status") || strings.Contains(lowerName, "type") || strings.Contains(lowerName, "state") {
			return "ACTIVE"
		}
		if strings.Contains(lowerName, "desc") || strings.Contains(lowerName, "remark") || strings.Contains(lowerName, "content") {
			return "测试内容..."
		}
		if strings.Contains(lowerName, "phone") || strings.Contains(lowerName, "mobile") {
			return "13800138000"
		}
		if strings.Contains(lowerName, "ip") {
			return "127.0.0.1"
		}

		// 4. Fallback
		if len(schema.Enum) > 0 {
			return schema.Enum[0]
		}
		return "" // Empty string instead of "<string>"

	case openapi3.TypeInteger, openapi3.TypeNumber:
		// 3. Name Heuristics (Number)
		if strings.Contains(lowerName, "price") || strings.Contains(lowerName, "cost") ||
			strings.Contains(lowerName, "salary") || strings.Contains(lowerName, "amount") {
			return 1000.00
		}
		if strings.Contains(lowerName, "id") || strings.Contains(lowerName, "code") || strings.Contains(lowerName, "no") {
			return 1
		}
		if strings.Contains(lowerName, "status") || strings.Contains(lowerName, "type") || strings.Contains(lowerName, "state") {
			return 1
		}
		// 4. Fallback
		return 0

	case openapi3.TypeBoolean:
		// 3. Name Heuristics (Boolean)
		if strings.Contains(lowerName, "is") || strings.Contains(lowerName, "has") || strings.Contains(lowerName, "enable") {
			return true
		}
		// 4. Fallback
		return false
	default:
		return nil
	}
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

	// Sanitize common bad data (e.g. Springfox symbols)
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
		return nil, fmt.Errorf("解析失败 (v2/v3). V3 错误: %v", errV3)
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
