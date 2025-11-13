package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/tinhminhtue/fission-mcp-server/internal/api"
	"github.com/tinhminhtue/fission-mcp-server/internal/cli"
)

const (
	defaultPort = "8080"
)

func main() {
	// Get kube context and namespace from environment variables
	kubeContext := os.Getenv("KUBE_CONTEXT")
	namespace := os.Getenv("NAMESPACE")

	// Initialize CLI executor
	executor, err := cli.NewExecutor(kubeContext, namespace)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing CLI executor: %v\n", err)
		os.Exit(1)
	}

	// Setup HTTP router
	router := api.SetupRouter(executor)

	// Get port from environment variable or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	// Start HTTP server
	addr := ":" + port
	log.Printf("Starting Fission HTTP Server on %s", addr)
	log.Printf("API endpoints:")
	log.Printf("  Functions: GET/POST /api/v1/functions, GET/PUT/DELETE /api/v1/functions/{name}")
	log.Printf("  Environments: GET/POST /api/v1/environments, GET/PUT/DELETE /api/v1/environments/{name}")
	log.Printf("  Packages: GET/POST /api/v1/packages, GET/PUT/DELETE /api/v1/packages/{name}")
	log.Printf("  HTTP Triggers: GET/POST /api/v1/httptriggers, GET/PUT/DELETE /api/v1/httptriggers/{name}")
	log.Printf("  Time Triggers: GET/POST /api/v1/timetriggers, DELETE /api/v1/timetriggers/{name}")
	log.Printf("  MQ Triggers: GET/POST /api/v1/mqtriggers, DELETE /api/v1/mqtriggers/{name}")
	log.Printf("  KubeWatches: GET/POST /api/v1/kubewatches, DELETE /api/v1/kubewatches/{name}")
	log.Printf("  Version: GET /api/v1/version")
	log.Printf("  Check: GET /api/v1/check")
	log.Printf("  Plugins: GET /api/v1/plugins")
	log.Printf("  GET  /openapi.yaml - OpenAPI specification (YAML)")
	log.Printf("  GET  /openapi.json - OpenAPI specification (JSON)")
	log.Printf("  GET  /swagger - Swagger UI (interactive API documentation)")
	log.Printf("  GET  /health - Health check")

	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
