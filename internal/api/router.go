package api

import (
	"net/http"
	"strings"

	"github.com/tinhminhtue/fission-mcp-server/internal/cli"
)

// SetupRouter sets up the HTTP router and returns the handler
func SetupRouter(executor *cli.Executor) http.Handler {
	// Create handlers
	functionHandler := NewFunctionHandler(executor)
	environmentHandler := NewEnvironmentHandler(executor)
	packageHandler := NewPackageHandler(executor)
	httpTriggerHandler := NewHTTPTriggerHandler(executor)
	remainingHandlers := NewRemainingHandlers(executor)

	// Keep old handler for OpenAPI and Swagger UI
	oldHandler := NewHandlerForOpenAPI()

	mux := http.NewServeMux()

	// Function routes
	mux.HandleFunc("/api/v1/functions/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		if strings.HasSuffix(path, "/test") {
			if r.Method == http.MethodPost {
				functionHandler.TestFunction(w, r)
			} else {
				functionHandler.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
			}
			return
		}

		if strings.HasSuffix(path, "/logs") {
			if r.Method == http.MethodGet {
				functionHandler.GetFunctionLogs(w, r)
			} else {
				functionHandler.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
			}
			return
		}

		if strings.HasSuffix(path, "/meta") {
			if r.Method == http.MethodGet {
				functionHandler.GetFunctionMeta(w, r)
			} else {
				functionHandler.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
			}
			return
		}

		if strings.HasSuffix(path, "/pods") {
			if r.Method == http.MethodGet {
				functionHandler.ListFunctionPods(w, r)
			} else {
				functionHandler.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
			}
			return
		}

		// Handle GET, PUT, DELETE for individual functions
		switch r.Method {
		case http.MethodGet:
			functionHandler.GetFunction(w, r)
		case http.MethodPut:
			functionHandler.UpdateFunction(w, r)
		case http.MethodDelete:
			functionHandler.DeleteFunction(w, r)
		default:
			functionHandler.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})

	mux.HandleFunc("/api/v1/functions", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			functionHandler.ListFunctions(w, r)
		case http.MethodPost:
			functionHandler.CreateFunction(w, r)
		default:
			functionHandler.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})

	// Environment routes
	mux.HandleFunc("/api/v1/environments/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		if strings.HasSuffix(path, "/pods") {
			if r.Method == http.MethodGet {
				environmentHandler.ListEnvironmentPods(w, r)
			} else {
				environmentHandler.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
			}
			return
		}

		switch r.Method {
		case http.MethodGet:
			environmentHandler.GetEnvironment(w, r)
		case http.MethodPut:
			environmentHandler.UpdateEnvironment(w, r)
		case http.MethodDelete:
			environmentHandler.DeleteEnvironment(w, r)
		default:
			environmentHandler.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})

	mux.HandleFunc("/api/v1/environments", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			environmentHandler.ListEnvironments(w, r)
		case http.MethodPost:
			environmentHandler.CreateEnvironment(w, r)
		default:
			environmentHandler.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})

	// Package routes
	mux.HandleFunc("/api/v1/packages/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		if strings.HasSuffix(path, "/src") {
			if r.Method == http.MethodGet {
				packageHandler.GetPackageSource(w, r)
			} else {
				packageHandler.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
			}
			return
		}

		if strings.HasSuffix(path, "/deploy") {
			if r.Method == http.MethodGet {
				packageHandler.GetPackageDeploy(w, r)
			} else {
				packageHandler.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
			}
			return
		}

		if strings.HasSuffix(path, "/info") {
			if r.Method == http.MethodGet {
				packageHandler.GetPackageInfo(w, r)
			} else {
				packageHandler.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
			}
			return
		}

		if strings.HasSuffix(path, "/rebuild") {
			if r.Method == http.MethodPost {
				packageHandler.RebuildPackage(w, r)
			} else {
				packageHandler.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
			}
			return
		}

		switch r.Method {
		case http.MethodGet:
			packageHandler.GetPackageInfo(w, r)
		case http.MethodPut:
			packageHandler.UpdatePackage(w, r)
		case http.MethodDelete:
			packageHandler.DeletePackage(w, r)
		default:
			packageHandler.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})

	mux.HandleFunc("/api/v1/packages", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			packageHandler.ListPackages(w, r)
		case http.MethodPost:
			packageHandler.CreatePackage(w, r)
		default:
			packageHandler.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})

	// HTTP Trigger routes
	mux.HandleFunc("/api/v1/httptriggers/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			httpTriggerHandler.GetHTTPTrigger(w, r)
		case http.MethodPut:
			httpTriggerHandler.UpdateHTTPTrigger(w, r)
		case http.MethodDelete:
			httpTriggerHandler.DeleteHTTPTrigger(w, r)
		default:
			httpTriggerHandler.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})

	mux.HandleFunc("/api/v1/httptriggers", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			httpTriggerHandler.ListHTTPTriggers(w, r)
		case http.MethodPost:
			httpTriggerHandler.CreateHTTPTrigger(w, r)
		default:
			httpTriggerHandler.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})

	// Time Trigger routes
	mux.HandleFunc("/api/v1/timetriggers/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			remainingHandlers.DeleteTimeTrigger(w, r)
		} else {
			remainingHandlers.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})

	mux.HandleFunc("/api/v1/timetriggers", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			remainingHandlers.ListTimeTriggers(w, r)
		case http.MethodPost:
			remainingHandlers.CreateTimeTrigger(w, r)
		default:
			remainingHandlers.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})

	// MQ Trigger routes
	mux.HandleFunc("/api/v1/mqtriggers/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			remainingHandlers.DeleteMQTrigger(w, r)
		} else {
			remainingHandlers.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})

	mux.HandleFunc("/api/v1/mqtriggers", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			remainingHandlers.ListMQTriggers(w, r)
		case http.MethodPost:
			remainingHandlers.CreateMQTrigger(w, r)
		default:
			remainingHandlers.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})

	// KubeWatch routes
	mux.HandleFunc("/api/v1/kubewatches/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			remainingHandlers.DeleteKubeWatch(w, r)
		} else {
			remainingHandlers.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})

	mux.HandleFunc("/api/v1/kubewatches", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			remainingHandlers.ListKubeWatches(w, r)
		case http.MethodPost:
			remainingHandlers.CreateKubeWatch(w, r)
		default:
			remainingHandlers.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})

	// Version route
	mux.HandleFunc("/api/v1/version", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			remainingHandlers.GetVersion(w, r)
		} else {
			remainingHandlers.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})

	// Check route
	mux.HandleFunc("/api/v1/check", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			remainingHandlers.Check(w, r)
		} else {
			remainingHandlers.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})

	// Plugin route
	mux.HandleFunc("/api/v1/plugins", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			remainingHandlers.ListPlugins(w, r)
		} else {
			remainingHandlers.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// OpenAPI specification endpoints (YAML and JSON)
	mux.HandleFunc("/openapi.yaml", oldHandler.ServeOpenAPI)
	mux.HandleFunc("/openapi.json", oldHandler.ServeOpenAPI)
	mux.HandleFunc("/api/v1/openapi.yaml", oldHandler.ServeOpenAPI)
	mux.HandleFunc("/api/v1/openapi.json", oldHandler.ServeOpenAPI)

	// Swagger UI endpoint
	mux.HandleFunc("/swagger", oldHandler.ServeSwaggerUI)
	mux.HandleFunc("/swagger/", oldHandler.ServeSwaggerUI)
	mux.HandleFunc("/docs", oldHandler.ServeSwaggerUI)
	mux.HandleFunc("/docs/", oldHandler.ServeSwaggerUI)

	return mux
}
