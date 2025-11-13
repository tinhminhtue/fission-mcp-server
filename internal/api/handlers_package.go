package api

import (
	"bytes"
	"encoding/json"
	"net/http"

	_package "github.com/fission/fission/pkg/fission-cli/cmd/package"
	flagkey "github.com/fission/fission/pkg/fission-cli/flag/key"
	"github.com/tinhminhtue/fission-mcp-server/internal/cli"
)

// PackageHandler handles package-related HTTP requests
type PackageHandler struct {
	executor *cli.Executor
}

// NewPackageHandler creates a new package handler
func NewPackageHandler(executor *cli.Executor) *PackageHandler {
	return &PackageHandler{
		executor: executor,
	}
}

// CreatePackage handles POST /api/v1/packages
func (h *PackageHandler) CreatePackage(w http.ResponseWriter, r *http.Request) {
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

	if name, ok := req["name"].(string); ok && name != "" {
		httpInput.SetValue(flagkey.PkgName, name)
	}
	if env, ok := req["environment"].(string); ok && env != "" {
		httpInput.SetValue(flagkey.PkgEnvironment, env)
	}
	if code, ok := req["code"].(string); ok && code != "" {
		httpInput.SetValue(flagkey.PkgCode, code)
	}
	if srcArchive, ok := req["source_archive"].(string); ok && srcArchive != "" {
		httpInput.SetValue(flagkey.PkgSrcArchive, srcArchive)
	}
	if deployArchive, ok := req["deploy_archive"].(string); ok && deployArchive != "" {
		httpInput.SetValue(flagkey.PkgDeployArchive, deployArchive)
	}
	if namespace, ok := req["namespace"].(string); ok && namespace != "" {
		httpInput.SetValue(flagkey.NamespacePackage, namespace)
	}

	err := h.executor.ExecuteCommand(_package.Create, httpInput)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := map[string]interface{}{
		"message": "Package created successfully",
		"output":  stdoutBuf.String(),
	}
	h.writeJSON(w, http.StatusCreated, response)
}

// ListPackages handles GET /api/v1/packages
func (h *PackageHandler) ListPackages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var stdoutBuf, stderrBuf bytes.Buffer
	httpInput := cli.NewHTTPInputFromRequestWithWriters(r, &stdoutBuf, &stderrBuf).(*cli.HTTPInput)

	if namespace := r.URL.Query().Get("namespace"); namespace != "" {
		httpInput.SetValue(flagkey.NamespacePackage, namespace)
	}
	if r.URL.Query().Get("all_namespaces") == "true" {
		httpInput.SetValue(flagkey.AllNamespaces, true)
	}
	if r.URL.Query().Get("orphan") == "true" {
		httpInput.SetValue(flagkey.PkgOrphan, true)
	}
	if status := r.URL.Query().Get("status"); status != "" {
		httpInput.SetValue(flagkey.PkgStatus, status)
	}

	err := h.executor.ExecuteCommand(_package.List, httpInput)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := map[string]interface{}{
		"output": stdoutBuf.String(),
	}
	h.writeJSON(w, http.StatusOK, response)
}

// GetPackageSource handles GET /api/v1/packages/{name}/src
func (h *PackageHandler) GetPackageSource(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	name := extractResourceName(r.URL.Path, "/api/v1/packages/")
	if name == "" {
		h.writeError(w, http.StatusBadRequest, "package name is required")
		return
	}

	var stdoutBuf, stderrBuf bytes.Buffer
	httpInput := cli.NewHTTPInputFromRequestWithWriters(r, &stdoutBuf, &stderrBuf).(*cli.HTTPInput)
	httpInput.SetValue(flagkey.PkgName, name)

	if namespace := r.URL.Query().Get("namespace"); namespace != "" {
		httpInput.SetValue(flagkey.NamespacePackage, namespace)
	}

	err := h.executor.ExecuteCommand(_package.GetSrc, httpInput)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := map[string]interface{}{
		"output": stdoutBuf.String(),
	}
	h.writeJSON(w, http.StatusOK, response)
}

// GetPackageDeploy handles GET /api/v1/packages/{name}/deploy
func (h *PackageHandler) GetPackageDeploy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	name := extractResourceName(r.URL.Path, "/api/v1/packages/")
	if name == "" {
		h.writeError(w, http.StatusBadRequest, "package name is required")
		return
	}

	var stdoutBuf, stderrBuf bytes.Buffer
	httpInput := cli.NewHTTPInputFromRequestWithWriters(r, &stdoutBuf, &stderrBuf).(*cli.HTTPInput)
	httpInput.SetValue(flagkey.PkgName, name)

	if namespace := r.URL.Query().Get("namespace"); namespace != "" {
		httpInput.SetValue(flagkey.NamespacePackage, namespace)
	}

	err := h.executor.ExecuteCommand(_package.GetDeploy, httpInput)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := map[string]interface{}{
		"output": stdoutBuf.String(),
	}
	h.writeJSON(w, http.StatusOK, response)
}

// GetPackageInfo handles GET /api/v1/packages/{name}/info
func (h *PackageHandler) GetPackageInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	name := extractResourceName(r.URL.Path, "/api/v1/packages/")
	if name == "" {
		h.writeError(w, http.StatusBadRequest, "package name is required")
		return
	}

	var stdoutBuf, stderrBuf bytes.Buffer
	httpInput := cli.NewHTTPInputFromRequestWithWriters(r, &stdoutBuf, &stderrBuf).(*cli.HTTPInput)
	httpInput.SetValue(flagkey.PkgName, name)

	if namespace := r.URL.Query().Get("namespace"); namespace != "" {
		httpInput.SetValue(flagkey.NamespacePackage, namespace)
	}

	err := h.executor.ExecuteCommand(_package.Info, httpInput)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := map[string]interface{}{
		"output": stdoutBuf.String(),
	}
	h.writeJSON(w, http.StatusOK, response)
}

// UpdatePackage handles PUT /api/v1/packages/{name}
func (h *PackageHandler) UpdatePackage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	name := extractResourceName(r.URL.Path, "/api/v1/packages/")
	if name == "" {
		h.writeError(w, http.StatusBadRequest, "package name is required")
		return
	}

	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	var stdoutBuf, stderrBuf bytes.Buffer
	httpInput := cli.NewHTTPInputFromRequestWithWriters(r, &stdoutBuf, &stderrBuf).(*cli.HTTPInput)
	httpInput.SetValue(flagkey.PkgName, name)

	if env, ok := req["environment"].(string); ok && env != "" {
		httpInput.SetValue(flagkey.PkgEnvironment, env)
	}
	if code, ok := req["code"].(string); ok && code != "" {
		httpInput.SetValue(flagkey.PkgCode, code)
	}
	if namespace, ok := req["namespace"].(string); ok && namespace != "" {
		httpInput.SetValue(flagkey.NamespacePackage, namespace)
	}

	err := h.executor.ExecuteCommand(_package.Update, httpInput)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := map[string]interface{}{
		"message": "Package updated successfully",
		"name":    name,
	}
	h.writeJSON(w, http.StatusOK, response)
}

// DeletePackage handles DELETE /api/v1/packages/{name}
func (h *PackageHandler) DeletePackage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	name := extractResourceName(r.URL.Path, "/api/v1/packages/")
	if name == "" {
		h.writeError(w, http.StatusBadRequest, "package name is required")
		return
	}

	var stdoutBuf, stderrBuf bytes.Buffer
	httpInput := cli.NewHTTPInputFromRequestWithWriters(r, &stdoutBuf, &stderrBuf).(*cli.HTTPInput)
	httpInput.SetValue(flagkey.PkgName, name)

	if namespace := r.URL.Query().Get("namespace"); namespace != "" {
		httpInput.SetValue(flagkey.NamespacePackage, namespace)
	}
	if r.URL.Query().Get("ignore_not_found") == "true" {
		httpInput.SetValue(flagkey.IgnoreNotFound, true)
	}
	if r.URL.Query().Get("force") == "true" {
		httpInput.SetValue(flagkey.PkgForce, true)
	}
	if r.URL.Query().Get("orphan") == "true" {
		httpInput.SetValue(flagkey.PkgOrphan, true)
	}

	err := h.executor.ExecuteCommand(_package.Delete, httpInput)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := map[string]interface{}{
		"message": "Package deleted successfully",
		"name":    name,
	}
	h.writeJSON(w, http.StatusOK, response)
}

// RebuildPackage handles POST /api/v1/packages/{name}/rebuild
func (h *PackageHandler) RebuildPackage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	name := extractResourceName(r.URL.Path, "/api/v1/packages/")
	if name == "" {
		h.writeError(w, http.StatusBadRequest, "package name is required")
		return
	}

	var stdoutBuf, stderrBuf bytes.Buffer
	httpInput := cli.NewHTTPInputFromRequestWithWriters(r, &stdoutBuf, &stderrBuf).(*cli.HTTPInput)
	httpInput.SetValue(flagkey.PkgName, name)

	if namespace := r.URL.Query().Get("namespace"); namespace != "" {
		httpInput.SetValue(flagkey.NamespacePackage, namespace)
	}

	err := h.executor.ExecuteCommand(_package.Rebuild, httpInput)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := map[string]interface{}{
		"message": "Package rebuild initiated",
		"name":    name,
		"output":  stdoutBuf.String(),
	}
	h.writeJSON(w, http.StatusOK, response)
}

func (h *PackageHandler) writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func (h *PackageHandler) writeError(w http.ResponseWriter, statusCode int, message string) {
	h.writeJSON(w, statusCode, map[string]interface{}{"error": message})
}

