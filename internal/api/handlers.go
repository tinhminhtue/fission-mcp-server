package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/tinhminhtue/fission-mcp-server/internal/fission"
	"gopkg.in/yaml.v2"
)

var (
	openAPISpec     []byte
	openAPISpecOnce sync.Once
)

// convertToJSONCompatible converts map[interface{}]interface{} to map[string]interface{}
// which is required for JSON encoding
func convertToJSONCompatible(data interface{}) interface{} {
	switch v := data.(type) {
	case map[interface{}]interface{}:
		result := make(map[string]interface{})
		for k, val := range v {
			result[fmt.Sprintf("%v", k)] = convertToJSONCompatible(val)
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, val := range v {
			result[i] = convertToJSONCompatible(val)
		}
		return result
	default:
		return data
	}
}

// loadOpenAPISpec loads the OpenAPI specification file from the docs directory
func loadOpenAPISpec() ([]byte, error) {
	// Get current working directory - this should be the project root when running
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current working directory: %w", err)
	}

	// Try paths relative to current working directory
	possiblePaths := []string{
		filepath.Join(cwd, "docs", "openapi.yaml"), // From project root
	}

	// Also try relative to executable location (for deployed binaries)
	if execPath, execErr := os.Executable(); execErr == nil {
		execDir := filepath.Dir(execPath)
		// Try common deployment patterns
		possiblePaths = append(possiblePaths,
			filepath.Join(execDir, "docs", "openapi.yaml"),             // Same directory as binary
			filepath.Join(execDir, "..", "docs", "openapi.yaml"),       // One level up
			filepath.Join(execDir, "..", "..", "docs", "openapi.yaml"), // Two levels up
		)
	}

	// Try each path
	for _, path := range possiblePaths {
		data, err := os.ReadFile(path)
		if err == nil {
			return data, nil
		}
	}

	return nil, fmt.Errorf("could not find openapi.yaml in docs directory. Tried paths: %v", possiblePaths)
}

// Handler provides HTTP handlers for the Fission API
type Handler struct {
	fissionService *fission.Service
}

// NewHandler creates a new API handler
func NewHandler(fissionService *fission.Service) *Handler {
	return &Handler{
		fissionService: fissionService,
	}
}

// NewHandlerForOpenAPI creates a minimal handler just for OpenAPI endpoints
func NewHandlerForOpenAPI() *Handler {
	return &Handler{
		fissionService: nil, // Not needed for OpenAPI endpoints
	}
}

// ListFunctionsResponse represents the response for listing functions
type ListFunctionsResponse struct {
	Functions []fission.FunctionInfo `json:"functions"`
	Count     int                    `json:"count"`
}

// LegacyCreateFunctionRequest represents the legacy request body for creating a function
// This is kept for backward compatibility but new code should use handlers_function.go
type LegacyCreateFunctionRequest struct {
	Name        string `json:"name"`
	Environment string `json:"environment"`
	Code        string `json:"code"`
}

// CreateFunctionResponse represents the response for creating a function
type CreateFunctionResponse struct {
	Message string `json:"message"`
	Name    string `json:"name"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// DeleteFunctionResponse represents the response for deleting a function
type DeleteFunctionResponse struct {
	Message string `json:"message"`
	Name    string `json:"name"`
}

// GetFunctionResponse represents the response for getting a function
type GetFunctionResponse struct {
	Function fission.FunctionInfo `json:"function"`
}

// TestFunctionRequest represents the request body for testing a function
type TestFunctionRequest struct {
	Body    string            `json:"body"`
	Headers map[string]string `json:"headers"`
}

// TestFunctionResponse represents the response for testing a function
type TestFunctionResponse struct {
	StatusCode int                 `json:"status_code"`
	Body       string              `json:"body"`
	Headers    map[string][]string `json:"headers"`
}

// GetFunctionLogsResponse represents the response for getting function logs
type GetFunctionLogsResponse struct {
	Logs string `json:"logs"`
}

// ListFunctions handles GET /api/v1/functions
func (h *Handler) ListFunctions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	functions, err := h.fissionService.ListFunctions(r.Context())
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := ListFunctionsResponse{
		Functions: functions,
		Count:     len(functions),
	}

	h.writeJSON(w, http.StatusOK, response)
}

// CreateFunction handles POST /api/v1/functions
func (h *Handler) CreateFunction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req CreateFunctionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	// Validate required fields
	if req.Name == "" {
		h.writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	if req.Environment == "" {
		h.writeError(w, http.StatusBadRequest, "environment is required")
		return
	}
	if req.Code == "" {
		h.writeError(w, http.StatusBadRequest, "code is required")
		return
	}

	err := h.fissionService.CreateFunction(r.Context(), req.Name, req.Environment, req.Code)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := CreateFunctionResponse{
		Message: "Function created successfully",
		Name:    req.Name,
	}

	h.writeJSON(w, http.StatusCreated, response)
}

// extractFunctionName extracts the function name from the URL path
// Expected format: /api/v1/functions/{name} or /api/v1/functions/{name}/...
func extractFunctionName(path string) string {
	// Remove /api/v1/functions prefix
	prefix := "/api/v1/functions/"
	if !strings.HasPrefix(path, prefix) {
		return ""
	}

	// Get the part after the prefix
	remaining := strings.TrimPrefix(path, prefix)

	// Extract function name (everything before the next /)
	parts := strings.Split(remaining, "/")
	return parts[0]
}

// GetFunction handles GET /api/v1/functions/{name}
func (h *Handler) GetFunction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	name := extractFunctionName(r.URL.Path)
	if name == "" {
		h.writeError(w, http.StatusBadRequest, "function name is required")
		return
	}

	// Get namespace from query parameter, default to empty (will use DefaultNamespace)
	namespace := r.URL.Query().Get("namespace")

	function, err := h.fissionService.GetFunction(r.Context(), name, namespace)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.writeError(w, http.StatusNotFound, err.Error())
		} else {
			h.writeError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	response := GetFunctionResponse{
		Function: *function,
	}

	h.writeJSON(w, http.StatusOK, response)
}

// DeleteFunction handles DELETE /api/v1/functions/{name}
func (h *Handler) DeleteFunction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	name := extractFunctionName(r.URL.Path)
	if name == "" {
		h.writeError(w, http.StatusBadRequest, "function name is required")
		return
	}

	// Get namespace from query parameter, default to empty (will use DefaultNamespace)
	namespace := r.URL.Query().Get("namespace")

	err := h.fissionService.DeleteFunction(r.Context(), name, namespace)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.writeError(w, http.StatusNotFound, err.Error())
		} else {
			h.writeError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	response := DeleteFunctionResponse{
		Message: "Function deleted successfully",
		Name:    name,
	}

	h.writeJSON(w, http.StatusOK, response)
}

// TestFunction handles POST /api/v1/functions/{name}/test
func (h *Handler) TestFunction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	name := extractFunctionName(r.URL.Path)
	if name == "" {
		h.writeError(w, http.StatusBadRequest, "function name is required")
		return
	}

	// Get namespace from query parameter, default to empty (will use DefaultNamespace)
	namespace := r.URL.Query().Get("namespace")

	var req TestFunctionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// If body is empty or invalid, use empty body
		req.Body = ""
		req.Headers = make(map[string]string)
	}

	if req.Headers == nil {
		req.Headers = make(map[string]string)
	}

	result, err := h.fissionService.TestFunction(r.Context(), name, namespace, req.Body, req.Headers)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.writeJSON(w, http.StatusOK, result)
}

// GetFunctionLogs handles GET /api/v1/functions/{name}/logs
func (h *Handler) GetFunctionLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	name := extractFunctionName(r.URL.Path)
	if name == "" {
		h.writeError(w, http.StatusBadRequest, "function name is required")
		return
	}

	// Get namespace from query parameter, default to empty (will use DefaultNamespace)
	namespace := r.URL.Query().Get("namespace")

	// Parse log options from query parameters
	options := fission.LogOptions{}
	if tail := r.URL.Query().Get("tail"); tail != "" {
		var tailInt int
		if _, err := fmt.Sscanf(tail, "%d", &tailInt); err == nil {
			options.Tail = tailInt
		}
	}
	if follow := r.URL.Query().Get("follow"); follow == "true" {
		options.Follow = true
	}
	if since := r.URL.Query().Get("since"); since != "" {
		options.Since = since
	}

	logs, err := h.fissionService.GetFunctionLogs(r.Context(), name, namespace, options)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := GetFunctionLogsResponse{
		Logs: logs,
	}

	h.writeJSON(w, http.StatusOK, response)
}

// writeJSON writes a JSON response
func (h *Handler) writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// writeError writes an error response
func (h *Handler) writeError(w http.ResponseWriter, statusCode int, message string) {
	h.writeJSON(w, statusCode, ErrorResponse{Error: message})
}

// ServeOpenAPI handles GET /openapi.yaml, GET /openapi.json, GET /api/v1/openapi.yaml, and GET /api/v1/openapi.json
func (h *Handler) ServeOpenAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Load the spec file once (thread-safe)
	var loadErr error
	openAPISpecOnce.Do(func() {
		openAPISpec, loadErr = loadOpenAPISpec()
		if loadErr != nil {
			// If loading fails, set to empty to avoid repeated attempts
			openAPISpec = []byte{}
		}
	})

	if loadErr != nil {
		h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to load OpenAPI specification: %v", loadErr))
		return
	}

	if len(openAPISpec) == 0 {
		h.writeError(w, http.StatusInternalServerError, "OpenAPI specification is empty")
		return
	}

	// Check if JSON format is requested
	path := r.URL.Path
	if strings.HasSuffix(path, ".json") {
		// Convert YAML to JSON
		var yamlData interface{}
		if err := yaml.Unmarshal(openAPISpec, &yamlData); err != nil {
			h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to parse YAML: %v", err))
			return
		}

		// Convert map[interface{}]interface{} to map[string]interface{} for JSON encoding
		jsonData := convertToJSONCompatible(yamlData)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(jsonData); err != nil {
			h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to encode JSON: %v", err))
			return
		}
		return
	}

	// Serve as YAML with text/yaml content type so browsers can display it
	w.Header().Set("Content-Type", "text/yaml; charset=utf-8")
	w.Header().Set("Content-Disposition", "inline")
	w.WriteHeader(http.StatusOK)
	w.Write(openAPISpec)
}

// ServeSwaggerUI serves the Swagger UI HTML page
func (h *Handler) ServeSwaggerUI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Determine the spec URL based on the request
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	host := r.Host
	specURL := fmt.Sprintf("%s://%s/openapi.json", scheme, host)

	// Swagger UI HTML with CDN
	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Fission MCP Server API - Swagger UI</title>
  <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@5.10.5/swagger-ui.css" />
  <style>
    html {
      box-sizing: border-box;
      overflow: -moz-scrollbars-vertical;
      overflow-y: scroll;
    }
    *, *:before, *:after {
      box-sizing: inherit;
    }
    body {
      margin:0;
      background: #fafafa;
    }
  </style>
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5.10.5/swagger-ui-bundle.js"></script>
  <script src="https://unpkg.com/swagger-ui-dist@5.10.5/swagger-ui-standalone-preset.js"></script>
  <script>
    window.onload = function() {
      const ui = SwaggerUIBundle({
        url: "%s",
        dom_id: '#swagger-ui',
        deepLinking: true,
        presets: [
          SwaggerUIBundle.presets.apis,
          SwaggerUIStandalonePreset
        ],
        plugins: [
          SwaggerUIBundle.plugins.DownloadUrl
        ],
        layout: "StandaloneLayout"
      });
    };
  </script>
</body>
</html>`, specURL)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}
