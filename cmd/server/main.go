package main

import (
	"context"
	"fmt"
	"os"

	"github.com/tinhminhtue/fission-mcp-server/internal/client"
	"github.com/tinhminhtue/fission-mcp-server/internal/fission"
)

const (
	functionName = "hello-world2"
	functionEnv  = "python"
	pythonCode   = `def main():
    return "Hello, World from hello-world2!"
`
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

	ctx := context.Background()

	// List Fission functions
	err = fissionService.PrintFunctions(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing functions: %v\n", err)
		fmt.Fprintf(os.Stderr, "Make sure Fission is installed and the Function CRD exists.\n")
		os.Exit(1)
	}

	// Create hello-world2 function
	fmt.Println("\n==================================================")
	fmt.Println("Creating hello-world2 function...")
	fmt.Println("==================================================")

	err = fissionService.CreateFunction(ctx, functionName, functionEnv, pythonCode)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating function: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully created function '%s'!\n", functionName)
}

