package api

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/fission/fission/pkg/fission-cli/cmd/check"
	"github.com/fission/fission/pkg/fission-cli/cmd/kubewatch"
	"github.com/fission/fission/pkg/fission-cli/cmd/mqtrigger"
	"github.com/fission/fission/pkg/fission-cli/cmd/plugin"
	"github.com/fission/fission/pkg/fission-cli/cmd/timetrigger"
	"github.com/fission/fission/pkg/fission-cli/cmd/version"
	flagkey "github.com/fission/fission/pkg/fission-cli/flag/key"
	"github.com/tinhminhtue/fission-mcp-server/internal/cli"
)

// RemainingHandlers handles all remaining command groups
type RemainingHandlers struct {
	executor *cli.Executor
}

// NewRemainingHandlers creates handlers for remaining command groups
func NewRemainingHandlers(executor *cli.Executor) *RemainingHandlers {
	return &RemainingHandlers{
		executor: executor,
	}
}

// Time Trigger Handlers

// CreateTimeTrigger handles POST /api/v1/timetriggers
func (h *RemainingHandlers) CreateTimeTrigger(w http.ResponseWriter, r *http.Request) {
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
		httpInput.SetValue(flagkey.TtName, name)
	}
	if fnName, ok := req["function"].(string); ok && fnName != "" {
		httpInput.SetValue(flagkey.TtFnName, fnName)
	}
	if cron, ok := req["cron"].(string); ok && cron != "" {
		httpInput.SetValue(flagkey.TtCron, cron)
	}

	err := h.executor.ExecuteCommand(timetrigger.Create, httpInput)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := map[string]interface{}{
		"message": "Time trigger created successfully",
		"output":  stdoutBuf.String(),
	}
	h.writeJSON(w, http.StatusCreated, response)
}

// ListTimeTriggers handles GET /api/v1/timetriggers
func (h *RemainingHandlers) ListTimeTriggers(w http.ResponseWriter, r *http.Request) {
	var stdoutBuf, stderrBuf bytes.Buffer
	httpInput := cli.NewHTTPInputFromRequestWithWriters(r, &stdoutBuf, &stderrBuf).(*cli.HTTPInput)

	if namespace := r.URL.Query().Get("namespace"); namespace != "" {
		httpInput.SetValue(flagkey.NamespaceTrigger, namespace)
	}

	err := h.executor.ExecuteCommand(timetrigger.List, httpInput)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := map[string]interface{}{
		"output": stdoutBuf.String(),
	}
	h.writeJSON(w, http.StatusOK, response)
}

// DeleteTimeTrigger handles DELETE /api/v1/timetriggers/{name}
func (h *RemainingHandlers) DeleteTimeTrigger(w http.ResponseWriter, r *http.Request) {
	name := extractResourceName(r.URL.Path, "/api/v1/timetriggers/")
	if name == "" {
		h.writeError(w, http.StatusBadRequest, "trigger name is required")
		return
	}

	var stdoutBuf, stderrBuf bytes.Buffer
	httpInput := cli.NewHTTPInputFromRequestWithWriters(r, &stdoutBuf, &stderrBuf).(*cli.HTTPInput)
	httpInput.SetValue(flagkey.TtName, name)

	err := h.executor.ExecuteCommand(timetrigger.Delete, httpInput)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := map[string]interface{}{
		"message": "Time trigger deleted successfully",
		"name":    name,
	}
	h.writeJSON(w, http.StatusOK, response)
}

// MQ Trigger Handlers

// CreateMQTrigger handles POST /api/v1/mqtriggers
func (h *RemainingHandlers) CreateMQTrigger(w http.ResponseWriter, r *http.Request) {
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
		httpInput.SetValue(flagkey.MqtName, name)
	}
	if fnName, ok := req["function"].(string); ok && fnName != "" {
		httpInput.SetValue(flagkey.MqtFnName, fnName)
	}
	if mqType, ok := req["mqtype"].(string); ok && mqType != "" {
		httpInput.SetValue(flagkey.MqtMQType, mqType)
	}
	if topic, ok := req["topic"].(string); ok && topic != "" {
		httpInput.SetValue(flagkey.MqtTopic, topic)
	}

	err := h.executor.ExecuteCommand(mqtrigger.Create, httpInput)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := map[string]interface{}{
		"message": "MQ trigger created successfully",
		"output":  stdoutBuf.String(),
	}
	h.writeJSON(w, http.StatusCreated, response)
}

// ListMQTriggers handles GET /api/v1/mqtriggers
func (h *RemainingHandlers) ListMQTriggers(w http.ResponseWriter, r *http.Request) {
	var stdoutBuf, stderrBuf bytes.Buffer
	httpInput := cli.NewHTTPInputFromRequestWithWriters(r, &stdoutBuf, &stderrBuf).(*cli.HTTPInput)

	if namespace := r.URL.Query().Get("namespace"); namespace != "" {
		httpInput.SetValue(flagkey.NamespaceTrigger, namespace)
	}

	err := h.executor.ExecuteCommand(mqtrigger.List, httpInput)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := map[string]interface{}{
		"output": stdoutBuf.String(),
	}
	h.writeJSON(w, http.StatusOK, response)
}

// DeleteMQTrigger handles DELETE /api/v1/mqtriggers/{name}
func (h *RemainingHandlers) DeleteMQTrigger(w http.ResponseWriter, r *http.Request) {
	name := extractResourceName(r.URL.Path, "/api/v1/mqtriggers/")
	if name == "" {
		h.writeError(w, http.StatusBadRequest, "trigger name is required")
		return
	}

	var stdoutBuf, stderrBuf bytes.Buffer
	httpInput := cli.NewHTTPInputFromRequestWithWriters(r, &stdoutBuf, &stderrBuf).(*cli.HTTPInput)
	httpInput.SetValue(flagkey.MqtName, name)

	err := h.executor.ExecuteCommand(mqtrigger.Delete, httpInput)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := map[string]interface{}{
		"message": "MQ trigger deleted successfully",
		"name":    name,
	}
	h.writeJSON(w, http.StatusOK, response)
}

// KubeWatch Handlers

// CreateKubeWatch handles POST /api/v1/kubewatches
func (h *RemainingHandlers) CreateKubeWatch(w http.ResponseWriter, r *http.Request) {
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
		httpInput.SetValue(flagkey.KwName, name)
	}
	if fnName, ok := req["function"].(string); ok && fnName != "" {
		httpInput.SetValue(flagkey.KwFnName, fnName)
	}
	if objType, ok := req["type"].(string); ok && objType != "" {
		httpInput.SetValue(flagkey.KwObjType, objType)
	}

	err := h.executor.ExecuteCommand(kubewatch.Create, httpInput)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := map[string]interface{}{
		"message": "KubeWatch created successfully",
		"output":  stdoutBuf.String(),
	}
	h.writeJSON(w, http.StatusCreated, response)
}

// ListKubeWatches handles GET /api/v1/kubewatches
func (h *RemainingHandlers) ListKubeWatches(w http.ResponseWriter, r *http.Request) {
	var stdoutBuf, stderrBuf bytes.Buffer
	httpInput := cli.NewHTTPInputFromRequestWithWriters(r, &stdoutBuf, &stderrBuf).(*cli.HTTPInput)

	err := h.executor.ExecuteCommand(kubewatch.List, httpInput)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := map[string]interface{}{
		"output": stdoutBuf.String(),
	}
	h.writeJSON(w, http.StatusOK, response)
}

// DeleteKubeWatch handles DELETE /api/v1/kubewatches/{name}
func (h *RemainingHandlers) DeleteKubeWatch(w http.ResponseWriter, r *http.Request) {
	name := extractResourceName(r.URL.Path, "/api/v1/kubewatches/")
	if name == "" {
		h.writeError(w, http.StatusBadRequest, "kubewatch name is required")
		return
	}

	var stdoutBuf, stderrBuf bytes.Buffer
	httpInput := cli.NewHTTPInputFromRequestWithWriters(r, &stdoutBuf, &stderrBuf).(*cli.HTTPInput)
	httpInput.SetValue(flagkey.KwName, name)

	err := h.executor.ExecuteCommand(kubewatch.Delete, httpInput)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := map[string]interface{}{
		"message": "KubeWatch deleted successfully",
		"name":    name,
	}
	h.writeJSON(w, http.StatusOK, response)
}

// Version Handler

// GetVersion handles GET /api/v1/version
func (h *RemainingHandlers) GetVersion(w http.ResponseWriter, r *http.Request) {
	var stdoutBuf, stderrBuf bytes.Buffer
	httpInput := cli.NewHTTPInputFromRequestWithWriters(r, &stdoutBuf, &stderrBuf).(*cli.HTTPInput)

	err := h.executor.ExecuteCommand(version.Version, httpInput)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := map[string]interface{}{
		"output": stdoutBuf.String(),
	}
	h.writeJSON(w, http.StatusOK, response)
}

// Check Handler

// Check handles GET /api/v1/check
func (h *RemainingHandlers) Check(w http.ResponseWriter, r *http.Request) {
	var stdoutBuf, stderrBuf bytes.Buffer
	httpInput := cli.NewHTTPInputFromRequestWithWriters(r, &stdoutBuf, &stderrBuf).(*cli.HTTPInput)

	err := h.executor.ExecuteCommand(check.Check, httpInput)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := map[string]interface{}{
		"output": stdoutBuf.String(),
	}
	h.writeJSON(w, http.StatusOK, response)
}

// Plugin Handler

// ListPlugins handles GET /api/v1/plugins
func (h *RemainingHandlers) ListPlugins(w http.ResponseWriter, r *http.Request) {
	var stdoutBuf, stderrBuf bytes.Buffer
	httpInput := cli.NewHTTPInputFromRequestWithWriters(r, &stdoutBuf, &stderrBuf).(*cli.HTTPInput)

	err := h.executor.ExecuteCommand(plugin.List, httpInput)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := map[string]interface{}{
		"output": stdoutBuf.String(),
	}
	h.writeJSON(w, http.StatusOK, response)
}

// Helper methods
func (h *RemainingHandlers) writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func (h *RemainingHandlers) writeError(w http.ResponseWriter, statusCode int, message string) {
	h.writeJSON(w, statusCode, map[string]interface{}{"error": message})
}
