package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/fission/fission/pkg/fission-cli/cmd/environment"
	flagkey "github.com/fission/fission/pkg/fission-cli/flag/key"
	"github.com/tinhminhtue/fission-mcp-server/internal/cli"
)

// EnvironmentHandler handles environment-related HTTP requests
type EnvironmentHandler struct {
	executor *cli.Executor
}

// NewEnvironmentHandler creates a new environment handler
func NewEnvironmentHandler(executor *cli.Executor) *EnvironmentHandler {
	return &EnvironmentHandler{
		executor: executor,
	}
}

// CreateEnvironment handles POST /api/v1/environments
func (h *EnvironmentHandler) CreateEnvironment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	var stdoutBuf, stderrBuf bytes.Buffer
	httpInput := cli.NewHTTPInputFromRequestWithWriters(r, &stdoutBuf, &stderrBuf).(*cli.HTTPInput)

	// Map request fields to CLI flags
	if name, ok := req["name"].(string); ok && name != "" {
		httpInput.SetValue(flagkey.EnvName, name)
	}
	if image, ok := req["image"].(string); ok && image != "" {
		httpInput.SetValue(flagkey.EnvImage, image)
	}
	if builderImage, ok := req["builder_image"].(string); ok && builderImage != "" {
		httpInput.SetValue(flagkey.EnvBuilderImage, builderImage)
	}
	if buildCmd, ok := req["build_command"].(string); ok && buildCmd != "" {
		httpInput.SetValue(flagkey.EnvBuildcommand, buildCmd)
	}
	if poolsize, ok := req["poolsize"].(float64); ok && poolsize > 0 {
		httpInput.SetValue(flagkey.EnvPoolsize, int(poolsize))
	}
	if namespace, ok := req["namespace"].(string); ok && namespace != "" {
		httpInput.SetValue(flagkey.NamespaceEnvironment, namespace)
	}

	err := h.executor.ExecuteCommand(environment.Create, httpInput)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := map[string]interface{}{
		"message": "Environment created successfully",
		"output":  stdoutBuf.String(),
	}
	h.writeJSON(w, http.StatusCreated, response)
}

// ListEnvironments handles GET /api/v1/environments
func (h *EnvironmentHandler) ListEnvironments(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var stdoutBuf, stderrBuf bytes.Buffer
	httpInput := cli.NewHTTPInputFromRequestWithWriters(r, &stdoutBuf, &stderrBuf).(*cli.HTTPInput)

	if namespace := r.URL.Query().Get("namespace"); namespace != "" {
		httpInput.SetValue(flagkey.NamespaceEnvironment, namespace)
	}
	if r.URL.Query().Get("all_namespaces") == "true" {
		httpInput.SetValue(flagkey.AllNamespaces, true)
	}

	err := h.executor.ExecuteCommand(environment.List, httpInput)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := map[string]interface{}{
		"output": stdoutBuf.String(),
	}
	h.writeJSON(w, http.StatusOK, response)
}

// GetEnvironment handles GET /api/v1/environments/{name}
func (h *EnvironmentHandler) GetEnvironment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	name := extractResourceName(r.URL.Path, "/api/v1/environments/")
	if name == "" {
		h.writeError(w, http.StatusBadRequest, "environment name is required")
		return
	}

	var stdoutBuf, stderrBuf bytes.Buffer
	httpInput := cli.NewHTTPInputFromRequestWithWriters(r, &stdoutBuf, &stderrBuf).(*cli.HTTPInput)
	httpInput.SetValue(flagkey.EnvName, name)

	if namespace := r.URL.Query().Get("namespace"); namespace != "" {
		httpInput.SetValue(flagkey.NamespaceEnvironment, namespace)
	}

	err := h.executor.ExecuteCommand(environment.Get, httpInput)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.writeError(w, http.StatusNotFound, err.Error())
		} else {
			h.writeError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	response := map[string]interface{}{
		"output": stdoutBuf.String(),
	}
	h.writeJSON(w, http.StatusOK, response)
}

// UpdateEnvironment handles PUT /api/v1/environments/{name}
func (h *EnvironmentHandler) UpdateEnvironment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	name := extractResourceName(r.URL.Path, "/api/v1/environments/")
	if name == "" {
		h.writeError(w, http.StatusBadRequest, "environment name is required")
		return
	}

	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	var stdoutBuf, stderrBuf bytes.Buffer
	httpInput := cli.NewHTTPInputFromRequestWithWriters(r, &stdoutBuf, &stderrBuf).(*cli.HTTPInput)
	httpInput.SetValue(flagkey.EnvName, name)

	// Map update fields
	if image, ok := req["image"].(string); ok && image != "" {
		httpInput.SetValue(flagkey.EnvImage, image)
	}
	if builderImage, ok := req["builder_image"].(string); ok && builderImage != "" {
		httpInput.SetValue(flagkey.EnvBuilderImage, builderImage)
	}
	if buildCmd, ok := req["build_command"].(string); ok && buildCmd != "" {
		httpInput.SetValue(flagkey.EnvBuildcommand, buildCmd)
	}
	if poolsize, ok := req["poolsize"].(float64); ok && poolsize > 0 {
		httpInput.SetValue(flagkey.EnvPoolsize, int(poolsize))
	}
	if namespace, ok := req["namespace"].(string); ok && namespace != "" {
		httpInput.SetValue(flagkey.NamespaceEnvironment, namespace)
	}

	err := h.executor.ExecuteCommand(environment.Update, httpInput)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := map[string]interface{}{
		"message": "Environment updated successfully",
		"name":    name,
	}
	h.writeJSON(w, http.StatusOK, response)
}

// DeleteEnvironment handles DELETE /api/v1/environments/{name}
func (h *EnvironmentHandler) DeleteEnvironment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	name := extractResourceName(r.URL.Path, "/api/v1/environments/")
	if name == "" {
		h.writeError(w, http.StatusBadRequest, "environment name is required")
		return
	}

	var stdoutBuf, stderrBuf bytes.Buffer
	httpInput := cli.NewHTTPInputFromRequestWithWriters(r, &stdoutBuf, &stderrBuf).(*cli.HTTPInput)
	httpInput.SetValue(flagkey.EnvName, name)

	if namespace := r.URL.Query().Get("namespace"); namespace != "" {
		httpInput.SetValue(flagkey.NamespaceEnvironment, namespace)
	}
	if r.URL.Query().Get("ignore_not_found") == "true" {
		httpInput.SetValue(flagkey.IgnoreNotFound, true)
	}
	if r.URL.Query().Get("force") == "true" {
		httpInput.SetValue(flagkey.EnvForce, true)
	}

	err := h.executor.ExecuteCommand(environment.Delete, httpInput)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.writeError(w, http.StatusNotFound, err.Error())
		} else {
			h.writeError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	response := map[string]interface{}{
		"message": "Environment deleted successfully",
		"name":    name,
	}
	h.writeJSON(w, http.StatusOK, response)
}

// ListEnvironmentPods handles GET /api/v1/environments/{name}/pods
func (h *EnvironmentHandler) ListEnvironmentPods(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	name := extractResourceName(r.URL.Path, "/api/v1/environments/")
	if name == "" {
		h.writeError(w, http.StatusBadRequest, "environment name is required")
		return
	}

	var stdoutBuf, stderrBuf bytes.Buffer
	httpInput := cli.NewHTTPInputFromRequestWithWriters(r, &stdoutBuf, &stderrBuf).(*cli.HTTPInput)
	httpInput.SetValue(flagkey.EnvName, name)

	if namespace := r.URL.Query().Get("namespace"); namespace != "" {
		httpInput.SetValue(flagkey.NamespaceEnvironment, namespace)
	}

	err := h.executor.ExecuteCommand(environment.ListPods, httpInput)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := map[string]interface{}{
		"output": stdoutBuf.String(),
	}
	h.writeJSON(w, http.StatusOK, response)
}

// Helper functions
func (h *EnvironmentHandler) writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func (h *EnvironmentHandler) writeError(w http.ResponseWriter, statusCode int, message string) {
	h.writeJSON(w, statusCode, map[string]interface{}{"error": message})
}

func extractResourceName(path, prefix string) string {
	if !strings.HasPrefix(path, prefix) {
		return ""
	}
	remaining := strings.TrimPrefix(path, prefix)
	parts := strings.Split(remaining, "/")
	return parts[0]
}

