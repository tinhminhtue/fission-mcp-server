package api

import (
	"net/http"

	"github.com/tinhminhtue/fission-mcp-server/internal/fission"
)

// SetupRouter sets up the HTTP router and returns the handler
func SetupRouter(fissionService *fission.Service) http.Handler {
	handler := NewHandler(fissionService)

	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("/api/v1/functions", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handler.ListFunctions(w, r)
		case http.MethodPost:
			handler.CreateFunction(w, r)
		default:
			handler.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			handler.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// OpenAPI specification endpoint
	mux.HandleFunc("/openapi.yaml", handler.ServeOpenAPI)
	mux.HandleFunc("/api/v1/openapi.yaml", handler.ServeOpenAPI)

	return mux
}

