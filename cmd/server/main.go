package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/tinhminhtue/fission-mcp-server/internal/api"
	"github.com/tinhminhtue/fission-mcp-server/internal/client"
	"github.com/tinhminhtue/fission-mcp-server/internal/fission"
)

const (
	defaultPort = "8080"
)

func main() {
	// Initialize Kubernetes client
	dynamicClient, err := client.NewDynamicClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing client: %v\n", err)
		os.Exit(1)
	}

	// Create Fission service
	fissionService := fission.NewService(dynamicClient)

	// Setup HTTP router
	router := api.SetupRouter(fissionService)

	// Get port from environment variable or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	// Start HTTP server
	addr := ":" + port
	log.Printf("Starting Fission MCP Server on %s", addr)
	log.Printf("API endpoints:")
	log.Printf("  GET  /api/v1/functions - List all Fission functions")
	log.Printf("  POST /api/v1/functions - Create a new Fission function")
	log.Printf("  GET  /openapi.yaml - OpenAPI specification (YAML)")
	log.Printf("  GET  /openapi.json - OpenAPI specification (JSON)")
	log.Printf("  GET  /api/v1/openapi.yaml - OpenAPI specification (YAML)")
	log.Printf("  GET  /api/v1/openapi.json - OpenAPI specification (JSON)")
	log.Printf("  GET  /swagger - Swagger UI (interactive API documentation)")
	log.Printf("  GET  /docs - Swagger UI (interactive API documentation)")
	log.Printf("  GET  /health - Health check")

	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
