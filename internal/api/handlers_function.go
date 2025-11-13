package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/fission/fission/pkg/fission-cli/cmd/function"
	flagkey "github.com/fission/fission/pkg/fission-cli/flag/key"
	"github.com/tinhminhtue/fission-mcp-server/internal/cli"
)

// FunctionHandler handles function-related HTTP requests using CLI executor
type FunctionHandler struct {
	executor *cli.Executor
}

// NewFunctionHandler creates a new function handler
func NewFunctionHandler(executor *cli.Executor) *FunctionHandler {
	return &FunctionHandler{
		executor: executor,
	}
}

// CreateFunctionRequest represents request for creating a function
type CreateFunctionRequest struct {
	Name                    string            `json:"name"`
	Environment             string            `json:"environment,omitempty"`
	PackageName             string            `json:"package_name,omitempty"`
	EntryPoint              string            `json:"entrypoint,omitempty"`
	ExecutorType            string            `json:"executor_type,omitempty"`
	Code                    string            `json:"code,omitempty"`
	SourceArchive           string            `json:"source_archive,omitempty"`
	DeployArchive           string            `json:"deploy_archive,omitempty"`
	SourceChecksum          string            `json:"source_checksum,omitempty"`
	DeployChecksum          string            `json:"deploy_checksum,omitempty"`
	Insecure                bool              `json:"insecure,omitempty"`
	BuildCommand            string            `json:"build_command,omitempty"`
	Secrets                 []string          `json:"secrets,omitempty"`
	ConfigMaps              []string          `json:"configmaps,omitempty"`
	SpecializationTimeout   int               `json:"specialization_timeout,omitempty"`
	ExecutionTimeout        int               `json:"execution_timeout,omitempty"`
	IdleTimeout             int               `json:"idle_timeout,omitempty"`
	Concurrency             int               `json:"concurrency,omitempty"`
	RequestsPerPod           int               `json:"requests_per_pod,omitempty"`
	OnceOnly                bool              `json:"once_only,omitempty"`
	RetainPods              int               `json:"retain_pods,omitempty"`
	Labels                  map[string]string `json:"labels,omitempty"`
	Annotations             map[string]string `json:"annotations,omitempty"`
	MinCPU                  string            `json:"min_cpu,omitempty"`
	MaxCPU                  string            `json:"max_cpu,omitempty"`
	MinMemory               string            `json:"min_memory,omitempty"`
	MaxMemory               string            `json:"max_memory,omitempty"`
	MinScale                int               `json:"min_scale,omitempty"`
	MaxScale                int               `json:"max_scale,omitempty"`
	TargetCPU               int               `json:"target_cpu,omitempty"`
	Namespace               string            `json:"namespace,omitempty"`
	URL                     string            `json:"url,omitempty"`
	Method                  string            `json:"method,omitempty"`
	Prefix                  string            `json:"prefix,omitempty"`
	SpecSave                bool              `json:"spec_save,omitempty"`
	SpecDry                 bool              `json:"spec_dry,omitempty"`
}

// CreateFunction handles POST /api/v1/functions
func (h *FunctionHandler) CreateFunction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req CreateFunctionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	if req.Name == "" {
		h.writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	// Create HTTP input from request
	var stdoutBuf, stderrBuf bytes.Buffer
	httpInput := cli.NewHTTPInputFromRequestWithWriters(r, &stdoutBuf, &stderrBuf).(*cli.HTTPInput)

	// Set function name
	httpInput.SetValue(flagkey.FnName, req.Name)

	// Set optional fields
	if req.Environment != "" {
		httpInput.SetValue(flagkey.FnEnvironmentName, req.Environment)
	}
	if req.PackageName != "" {
		httpInput.SetValue(flagkey.FnPackageName, req.PackageName)
	}
	if req.EntryPoint != "" {
		httpInput.SetValue(flagkey.FnEntrypoint, req.EntryPoint)
	}
	if req.ExecutorType != "" {
		httpInput.SetValue(flagkey.FnExecutorType, req.ExecutorType)
	}
	if req.Code != "" {
		httpInput.SetValue(flagkey.PkgCode, req.Code)
	}
	if req.SourceArchive != "" {
		httpInput.SetValue(flagkey.PkgSrcArchive, req.SourceArchive)
	}
	if req.DeployArchive != "" {
		httpInput.SetValue(flagkey.PkgDeployArchive, req.DeployArchive)
	}
	if req.SourceChecksum != "" {
		httpInput.SetValue(flagkey.PkgSrcChecksum, req.SourceChecksum)
	}
	if req.DeployChecksum != "" {
		httpInput.SetValue(flagkey.PkgDeployChecksum, req.DeployChecksum)
	}
	if req.Insecure {
		httpInput.SetValue(flagkey.PkgInsecure, true)
	}
	if req.BuildCommand != "" {
		httpInput.SetValue(flagkey.FnBuildCmd, req.BuildCommand)
	}
	if len(req.Secrets) > 0 {
		httpInput.SetValue(flagkey.FnSecret, req.Secrets)
	}
	if len(req.ConfigMaps) > 0 {
		httpInput.SetValue(flagkey.FnCfgMap, req.ConfigMaps)
	}
	if req.SpecializationTimeout > 0 {
		httpInput.SetValue(flagkey.FnSpecializationTimeout, req.SpecializationTimeout)
	}
	if req.ExecutionTimeout > 0 {
		httpInput.SetValue(flagkey.FnExecutionTimeout, req.ExecutionTimeout)
	}
	if req.IdleTimeout > 0 {
		httpInput.SetValue(flagkey.FnIdleTimeout, req.IdleTimeout)
	}
	if req.Concurrency > 0 {
		httpInput.SetValue(flagkey.FnConcurrency, req.Concurrency)
	}
	if req.RequestsPerPod > 0 {
		httpInput.SetValue(flagkey.FnRequestsPerPod, req.RequestsPerPod)
	}
	if req.OnceOnly {
		httpInput.SetValue(flagkey.FnOnceOnly, true)
	}
	if req.RetainPods > 0 {
		httpInput.SetValue(flagkey.FnRetainPods, req.RetainPods)
	}
	if req.Namespace != "" {
		httpInput.SetValue(flagkey.NamespaceFunction, req.Namespace)
	}
	if req.URL != "" {
		httpInput.SetValue(flagkey.HtUrl, req.URL)
	}
	if req.Method != "" {
		httpInput.SetValue(flagkey.HtMethod, req.Method)
	}
	if req.Prefix != "" {
		httpInput.SetValue(flagkey.HtPrefix, req.Prefix)
	}
	if req.SpecSave {
		httpInput.SetValue(flagkey.SpecSave, true)
	}
	if req.SpecDry {
		httpInput.SetValue(flagkey.SpecDry, true)
	}
	if req.MinCPU != "" {
		httpInput.SetValue(flagkey.RuntimeMincpu, req.MinCPU)
	}
	if req.MaxCPU != "" {
		httpInput.SetValue(flagkey.RuntimeMaxcpu, req.MaxCPU)
	}
	if req.MinMemory != "" {
		httpInput.SetValue(flagkey.RuntimeMinmemory, req.MinMemory)
	}
	if req.MaxMemory != "" {
		httpInput.SetValue(flagkey.RuntimeMaxmemory, req.MaxMemory)
	}
	if req.MinScale > 0 {
		httpInput.SetValue(flagkey.ReplicasMinscale, req.MinScale)
	}
	if req.MaxScale > 0 {
		httpInput.SetValue(flagkey.ReplicasMaxscale, req.MaxScale)
	}
	if req.TargetCPU > 0 {
		httpInput.SetValue(flagkey.RuntimeTargetcpu, req.TargetCPU)
	}

	// Handle labels and annotations
	if len(req.Labels) > 0 {
		var labelPairs []string
		for k, v := range req.Labels {
			labelPairs = append(labelPairs, fmt.Sprintf("%s=%s", k, v))
		}
		httpInput.SetValue(flagkey.Labels, labelPairs)
	}
	if len(req.Annotations) > 0 {
		var annotationPairs []string
		for k, v := range req.Annotations {
			annotationPairs = append(annotationPairs, fmt.Sprintf("%s=%s", k, v))
		}
		httpInput.SetValue(flagkey.Annotation, annotationPairs)
	}

	// Execute create command

	err := h.executor.ExecuteCommand(function.Create, httpInput)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := map[string]interface{}{
		"message": "Function created successfully",
		"name":    req.Name,
	}
	h.writeJSON(w, http.StatusCreated, response)
}

// ListFunctions handles GET /api/v1/functions
func (h *FunctionHandler) ListFunctions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var stdoutBuf, stderrBuf bytes.Buffer
	httpInput := cli.NewHTTPInputFromRequestWithWriters(r, &stdoutBuf, &stderrBuf).(*cli.HTTPInput)

	// Set namespace if provided
	if namespace := r.URL.Query().Get("namespace"); namespace != "" {
		httpInput.SetValue(flagkey.NamespaceFunction, namespace)
	}
	if r.URL.Query().Get("all_namespaces") == "true" {
		httpInput.SetValue(flagkey.AllNamespaces, true)
	}

	err := h.executor.ExecuteCommand(function.List, httpInput)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Parse tabular output and convert to JSON
	// For now, return the raw output
	response := map[string]interface{}{
		"output": stdoutBuf.String(),
	}
	h.writeJSON(w, http.StatusOK, response)
}

// GetFunction handles GET /api/v1/functions/{name}
func (h *FunctionHandler) GetFunction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	name := extractFunctionName(r.URL.Path)
	if name == "" {
		h.writeError(w, http.StatusBadRequest, "function name is required")
		return
	}

	var stdoutBuf, stderrBuf bytes.Buffer
	httpInput := cli.NewHTTPInputFromRequestWithWriters(r, &stdoutBuf, &stderrBuf).(*cli.HTTPInput)
	httpInput.SetValue(flagkey.FnName, name)

	if namespace := r.URL.Query().Get("namespace"); namespace != "" {
		httpInput.SetValue(flagkey.NamespaceFunction, namespace)
	}

	err := h.executor.ExecuteCommand(function.Get, httpInput)
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

// GetFunctionMeta handles GET /api/v1/functions/{name}/meta
func (h *FunctionHandler) GetFunctionMeta(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	name := extractFunctionName(r.URL.Path)
	if name == "" {
		h.writeError(w, http.StatusBadRequest, "function name is required")
		return
	}

	var stdoutBuf, stderrBuf bytes.Buffer
	httpInput := cli.NewHTTPInputFromRequestWithWriters(r, &stdoutBuf, &stderrBuf).(*cli.HTTPInput)
	httpInput.SetValue(flagkey.FnName, name)

	if namespace := r.URL.Query().Get("namespace"); namespace != "" {
		httpInput.SetValue(flagkey.NamespaceFunction, namespace)
	}

	err := h.executor.ExecuteCommand(function.GetMeta, httpInput)
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

// UpdateFunction handles PUT /api/v1/functions/{name}
func (h *FunctionHandler) UpdateFunction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	name := extractFunctionName(r.URL.Path)
	if name == "" {
		h.writeError(w, http.StatusBadRequest, "function name is required")
		return
	}

	var req CreateFunctionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	var stdoutBuf, stderrBuf bytes.Buffer
	httpInput := cli.NewHTTPInputFromRequestWithWriters(r, &stdoutBuf, &stderrBuf).(*cli.HTTPInput)
	httpInput.SetValue(flagkey.FnName, name)

	// Set all update fields similar to create
	if req.Environment != "" {
		httpInput.SetValue(flagkey.FnEnvironmentName, req.Environment)
	}
	if req.PackageName != "" {
		httpInput.SetValue(flagkey.FnPackageName, req.PackageName)
	}
	if req.EntryPoint != "" {
		httpInput.SetValue(flagkey.FnEntrypoint, req.EntryPoint)
	}
	if req.ExecutorType != "" {
		httpInput.SetValue(flagkey.FnExecutorType, req.ExecutorType)
	}
	if req.Code != "" {
		httpInput.SetValue(flagkey.PkgCode, req.Code)
	}
	if req.SourceArchive != "" {
		httpInput.SetValue(flagkey.PkgSrcArchive, req.SourceArchive)
	}
	if req.DeployArchive != "" {
		httpInput.SetValue(flagkey.PkgDeployArchive, req.DeployArchive)
	}
	if req.BuildCommand != "" {
		httpInput.SetValue(flagkey.FnBuildCmd, req.BuildCommand)
	}
	if len(req.Secrets) > 0 {
		httpInput.SetValue(flagkey.FnSecret, req.Secrets)
	}
	if len(req.ConfigMaps) > 0 {
		httpInput.SetValue(flagkey.FnCfgMap, req.ConfigMaps)
	}
	if req.ExecutionTimeout > 0 {
		httpInput.SetValue(flagkey.FnExecutionTimeout, req.ExecutionTimeout)
	}
	if req.IdleTimeout > 0 {
		httpInput.SetValue(flagkey.FnIdleTimeout, req.IdleTimeout)
	}
	if req.Concurrency > 0 {
		httpInput.SetValue(flagkey.FnConcurrency, req.Concurrency)
	}
	if req.RequestsPerPod > 0 {
		httpInput.SetValue(flagkey.FnRequestsPerPod, req.RequestsPerPod)
	}
	if req.OnceOnly {
		httpInput.SetValue(flagkey.FnOnceOnly, true)
	}
	if req.RetainPods > 0 {
		httpInput.SetValue(flagkey.FnRetainPods, req.RetainPods)
	}
	if req.Namespace != "" {
		httpInput.SetValue(flagkey.NamespaceFunction, req.Namespace)
	}
	if req.SpecSave {
		httpInput.SetValue(flagkey.SpecSave, true)
	}

	err := h.executor.ExecuteCommand(function.Update, httpInput)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := map[string]interface{}{
		"message": "Function updated successfully",
		"name":    name,
	}
	h.writeJSON(w, http.StatusOK, response)
}

// DeleteFunction handles DELETE /api/v1/functions/{name}
func (h *FunctionHandler) DeleteFunction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	name := extractFunctionName(r.URL.Path)
	if name == "" {
		h.writeError(w, http.StatusBadRequest, "function name is required")
		return
	}

	var stdoutBuf, stderrBuf bytes.Buffer
	httpInput := cli.NewHTTPInputFromRequestWithWriters(r, &stdoutBuf, &stderrBuf).(*cli.HTTPInput)
	httpInput.SetValue(flagkey.FnName, name)

	if namespace := r.URL.Query().Get("namespace"); namespace != "" {
		httpInput.SetValue(flagkey.NamespaceFunction, namespace)
	}
	if r.URL.Query().Get("ignore_not_found") == "true" {
		httpInput.SetValue(flagkey.IgnoreNotFound, true)
	}

	err := h.executor.ExecuteCommand(function.Delete, httpInput)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.writeError(w, http.StatusNotFound, err.Error())
		} else {
			h.writeError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	response := map[string]interface{}{
		"message": "Function deleted successfully",
		"name":    name,
	}
	h.writeJSON(w, http.StatusOK, response)
}

// TestFunction handles POST /api/v1/functions/{name}/test
func (h *FunctionHandler) TestFunction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	name := extractFunctionName(r.URL.Path)
	if name == "" {
		h.writeError(w, http.StatusBadRequest, "function name is required")
		return
	}

	var req struct {
		Body    string            `json:"body"`
		Headers map[string]string `json:"headers"`
		Query   string            `json:"query"`
		Method  string            `json:"method"`
		Timeout int               `json:"timeout"`
		SubPath string            `json:"subpath"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		req.Body = ""
		req.Headers = make(map[string]string)
	}

	var stdoutBuf, stderrBuf bytes.Buffer
	httpInput := cli.NewHTTPInputFromRequestWithWriters(r, &stdoutBuf, &stderrBuf).(*cli.HTTPInput)
	httpInput.SetValue(flagkey.FnName, name)

	if req.Body != "" {
		httpInput.SetValue(flagkey.FnTestBody, req.Body)
	}
	if len(req.Headers) > 0 {
		var headerPairs []string
		for k, v := range req.Headers {
			headerPairs = append(headerPairs, fmt.Sprintf("%s:%s", k, v))
		}
		httpInput.SetValue(flagkey.FnTestHeader, headerPairs)
	}
	if req.Query != "" {
		httpInput.SetValue(flagkey.FnTestQuery, req.Query)
	}
	if req.Method != "" {
		httpInput.SetValue(flagkey.HtMethod, req.Method)
	}
	if req.Timeout > 0 {
		httpInput.SetValue(flagkey.FnTestTimeout, req.Timeout)
	}
	if req.SubPath != "" {
		httpInput.SetValue(flagkey.FnSubPath, req.SubPath)
	}
	if namespace := r.URL.Query().Get("namespace"); namespace != "" {
		httpInput.SetValue(flagkey.NamespaceFunction, namespace)
	}

	err := h.executor.ExecuteCommand(function.Test, httpInput)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := map[string]interface{}{
		"output": stdoutBuf.String(),
	}
	h.writeJSON(w, http.StatusOK, response)
}

// GetFunctionLogs handles GET /api/v1/functions/{name}/logs
func (h *FunctionHandler) GetFunctionLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	name := extractFunctionName(r.URL.Path)
	if name == "" {
		h.writeError(w, http.StatusBadRequest, "function name is required")
		return
	}

	var stdoutBuf, stderrBuf bytes.Buffer
	httpInput := cli.NewHTTPInputFromRequestWithWriters(r, &stdoutBuf, &stderrBuf).(*cli.HTTPInput)
	httpInput.SetValue(flagkey.FnName, name)

	if namespace := r.URL.Query().Get("namespace"); namespace != "" {
		httpInput.SetValue(flagkey.NamespaceFunction, namespace)
	}
	if pod := r.URL.Query().Get("pod"); pod != "" {
		httpInput.SetValue(flagkey.FnLogPod, pod)
	}
	if r.URL.Query().Get("follow") == "true" {
		httpInput.SetValue(flagkey.FnLogFollow, true)
	}
	if detail := r.URL.Query().Get("detail"); detail != "" {
		httpInput.SetValue(flagkey.FnLogDetail, detail)
	}
	if dbType := r.URL.Query().Get("dbtype"); dbType != "" {
		httpInput.SetValue(flagkey.FnLogDBType, dbType)
	}
	if reverse := r.URL.Query().Get("reverse"); reverse != "" {
		httpInput.SetValue(flagkey.FnLogReverseQuery, reverse)
	}
	if count := r.URL.Query().Get("recordcount"); count != "" {
		if countInt, err := parseInt(count); err == nil {
			httpInput.SetValue(flagkey.FnLogCount, countInt)
		}
	}
	if r.URL.Query().Get("all_pods") == "true" {
		httpInput.SetValue(flagkey.FnLogAllPods, true)
	}

	err := h.executor.ExecuteCommand(function.Log, httpInput)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := map[string]interface{}{
		"logs": stdoutBuf.String(),
	}
	h.writeJSON(w, http.StatusOK, response)
}

// ListFunctionPods handles GET /api/v1/functions/{name}/pods
func (h *FunctionHandler) ListFunctionPods(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	name := extractFunctionName(r.URL.Path)
	if name == "" {
		h.writeError(w, http.StatusBadRequest, "function name is required")
		return
	}

	var stdoutBuf, stderrBuf bytes.Buffer
	httpInput := cli.NewHTTPInputFromRequestWithWriters(r, &stdoutBuf, &stderrBuf).(*cli.HTTPInput)
	httpInput.SetValue(flagkey.FnName, name)

	if namespace := r.URL.Query().Get("namespace"); namespace != "" {
		httpInput.SetValue(flagkey.NamespaceFunction, namespace)
	}

	err := h.executor.ExecuteCommand(function.ListPods, httpInput)
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
func parseInt(s string) (int, error) {
	var i int
	_, err := fmt.Sscanf(s, "%d", &i)
	return i, err
}

// writeJSON writes a JSON response
func (h *FunctionHandler) writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// writeError writes an error response
func (h *FunctionHandler) writeError(w http.ResponseWriter, statusCode int, message string) {
	h.writeJSON(w, statusCode, map[string]interface{}{"error": message})
}

