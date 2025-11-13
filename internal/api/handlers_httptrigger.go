package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/fission/fission/pkg/fission-cli/cmd/httptrigger"
	flagkey "github.com/fission/fission/pkg/fission-cli/flag/key"
	"github.com/tinhminhtue/fission-mcp-server/internal/cli"
)

// HTTPTriggerHandler handles HTTP trigger-related requests
type HTTPTriggerHandler struct {
	executor *cli.Executor
}

// NewHTTPTriggerHandler creates a new HTTP trigger handler
func NewHTTPTriggerHandler(executor *cli.Executor) *HTTPTriggerHandler {
	return &HTTPTriggerHandler{
		executor: executor,
	}
}

// CreateHTTPTrigger handles POST /api/v1/httptriggers
func (h *HTTPTriggerHandler) CreateHTTPTrigger(w http.ResponseWriter, r *http.Request) {
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

	if fnName, ok := req["function"].(string); ok && fnName != "" {
		httpInput.SetValue(flagkey.HtFnName, fnName)
	}
	if name, ok := req["name"].(string); ok && name != "" {
		httpInput.SetValue(flagkey.HtName, name)
	}
	if url, ok := req["url"].(string); ok && url != "" {
		httpInput.SetValue(flagkey.HtUrl, url)
	}
	if method, ok := req["method"].(string); ok && method != "" {
		httpInput.SetValue(flagkey.HtMethod, method)
	}
	if prefix, ok := req["prefix"].(string); ok && prefix != "" {
		httpInput.SetValue(flagkey.HtPrefix, prefix)
	}

	err := h.executor.ExecuteCommand(httptrigger.Create, httpInput)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := map[string]interface{}{
		"message": "HTTP trigger created successfully",
		"output":  stdoutBuf.String(),
	}
	h.writeJSON(w, http.StatusCreated, response)
}

// ListHTTPTriggers handles GET /api/v1/httptriggers
func (h *HTTPTriggerHandler) ListHTTPTriggers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var stdoutBuf, stderrBuf bytes.Buffer
	httpInput := cli.NewHTTPInputFromRequestWithWriters(r, &stdoutBuf, &stderrBuf).(*cli.HTTPInput)

	if namespace := r.URL.Query().Get("namespace"); namespace != "" {
		httpInput.SetValue(flagkey.NamespaceTrigger, namespace)
	}
	if r.URL.Query().Get("all_namespaces") == "true" {
		httpInput.SetValue(flagkey.AllNamespaces, true)
	}
	if fnFilter := r.URL.Query().Get("function"); fnFilter != "" {
		httpInput.SetValue(flagkey.HtFilter, fnFilter)
	}

	err := h.executor.ExecuteCommand(httptrigger.List, httpInput)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := map[string]interface{}{
		"output": stdoutBuf.String(),
	}
	h.writeJSON(w, http.StatusOK, response)
}

// GetHTTPTrigger handles GET /api/v1/httptriggers/{name}
func (h *HTTPTriggerHandler) GetHTTPTrigger(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	name := extractResourceName(r.URL.Path, "/api/v1/httptriggers/")
	if name == "" {
		h.writeError(w, http.StatusBadRequest, "trigger name is required")
		return
	}

	var stdoutBuf, stderrBuf bytes.Buffer
	httpInput := cli.NewHTTPInputFromRequestWithWriters(r, &stdoutBuf, &stderrBuf).(*cli.HTTPInput)
	httpInput.SetValue(flagkey.HtName, name)

	err := h.executor.ExecuteCommand(httptrigger.Get, httpInput)
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

// UpdateHTTPTrigger handles PUT /api/v1/httptriggers/{name}
func (h *HTTPTriggerHandler) UpdateHTTPTrigger(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	name := extractResourceName(r.URL.Path, "/api/v1/httptriggers/")
	if name == "" {
		h.writeError(w, http.StatusBadRequest, "trigger name is required")
		return
	}

	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	var stdoutBuf, stderrBuf bytes.Buffer
	httpInput := cli.NewHTTPInputFromRequestWithWriters(r, &stdoutBuf, &stderrBuf).(*cli.HTTPInput)
	httpInput.SetValue(flagkey.HtName, name)

	if fnName, ok := req["function"].(string); ok && fnName != "" {
		httpInput.SetValue(flagkey.HtFnName, fnName)
	}
	if url, ok := req["url"].(string); ok && url != "" {
		httpInput.SetValue(flagkey.HtUrl, url)
	}
	if method, ok := req["method"].(string); ok && method != "" {
		httpInput.SetValue(flagkey.HtMethod, method)
	}

	err := h.executor.ExecuteCommand(httptrigger.Update, httpInput)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := map[string]interface{}{
		"message": "HTTP trigger updated successfully",
		"name":    name,
	}
	h.writeJSON(w, http.StatusOK, response)
}

// DeleteHTTPTrigger handles DELETE /api/v1/httptriggers/{name}
func (h *HTTPTriggerHandler) DeleteHTTPTrigger(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	name := extractResourceName(r.URL.Path, "/api/v1/httptriggers/")
	if name == "" {
		h.writeError(w, http.StatusBadRequest, "trigger name is required")
		return
	}

	var stdoutBuf, stderrBuf bytes.Buffer
	httpInput := cli.NewHTTPInputFromRequestWithWriters(r, &stdoutBuf, &stderrBuf).(*cli.HTTPInput)
	httpInput.SetValue(flagkey.HtName, name)

	if namespace := r.URL.Query().Get("namespace"); namespace != "" {
		httpInput.SetValue(flagkey.NamespaceTrigger, namespace)
	}
	if r.URL.Query().Get("ignore_not_found") == "true" {
		httpInput.SetValue(flagkey.IgnoreNotFound, true)
	}

	err := h.executor.ExecuteCommand(httptrigger.Delete, httpInput)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.writeError(w, http.StatusNotFound, err.Error())
		} else {
			h.writeError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	response := map[string]interface{}{
		"message": "HTTP trigger deleted successfully",
		"name":    name,
	}
	h.writeJSON(w, http.StatusOK, response)
}

func (h *HTTPTriggerHandler) writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func (h *HTTPTriggerHandler) writeError(w http.ResponseWriter, statusCode int, message string) {
	h.writeJSON(w, statusCode, map[string]interface{}{"error": message})
}

