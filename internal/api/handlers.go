package api

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/tinhminhtue/fission-mcp-server/internal/fission"
)

var (
	openAPISpec     []byte
	openAPISpecOnce sync.Once
)

// loadOpenAPISpec loads the OpenAPI specification file
func loadOpenAPISpec() ([]byte, error) {
	var data []byte
	var err error

	// Try multiple possible paths
	possiblePaths := []string{
		"docs/openapi.yaml",                    // From project root
		"../docs/openapi.yaml",                 // From internal/api/
		filepath.Join("..", "docs", "openapi.yaml"), // From internal/api/ (cross-platform)
	}

	// Also try relative to executable location
	if execPath, execErr := os.Executable(); execErr == nil {
		execDir := filepath.Dir(execPath)
		possiblePaths = append(possiblePaths,
			filepath.Join(execDir, "docs", "openapi.yaml"),
			filepath.Join(execDir, "..", "docs", "openapi.yaml"),
		)
	}

	for _, path := range possiblePaths {
		data, err = os.ReadFile(path)
		if err == nil {
			return data, nil
		}
	}

	return nil, err
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

// ListFunctionsResponse represents the response for listing functions
type ListFunctionsResponse struct {
	Functions []fission.FunctionInfo `json:"functions"`
	Count     int                    `json:"count"`
}

// CreateFunctionRequest represents the request body for creating a function
type CreateFunctionRequest struct {
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

// ServeOpenAPI handles GET /openapi.yaml and GET /api/v1/openapi.yaml
func (h *Handler) ServeOpenAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Load the spec file once (thread-safe)
	openAPISpecOnce.Do(func() {
		var err error
		openAPISpec, err = loadOpenAPISpec()
		if err != nil {
			// If loading fails, set to empty to avoid repeated attempts
			openAPISpec = []byte{}
		}
	})

	if len(openAPISpec) == 0 {
		h.writeError(w, http.StatusInternalServerError, "Failed to load OpenAPI specification")
		return
	}

	w.Header().Set("Content-Type", "application/yaml")
	w.WriteHeader(http.StatusOK)
	w.Write(openAPISpec)
}

