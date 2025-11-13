package fission

import "k8s.io/apimachinery/pkg/runtime/schema"

const (
	// DefaultNamespace is the default namespace for Fission resources
	DefaultNamespace = "ai-code"
)

var (
	// FunctionGVR is the GroupVersionResource for Fission Functions
	FunctionGVR = schema.GroupVersionResource{
		Group:    "fission.io",
		Version:  "v1",
		Resource: "functions",
	}

	// PackageGVR is the GroupVersionResource for Fission Packages
	PackageGVR = schema.GroupVersionResource{
		Group:    "fission.io",
		Version:  "v1",
		Resource: "packages",
	}

	// HTTPTriggerGVR is the GroupVersionResource for Fission HTTP Triggers
	HTTPTriggerGVR = schema.GroupVersionResource{
		Group:    "fission.io",
		Version:  "v1",
		Resource: "httptriggers",
	}
)

// FunctionInfo represents information about a Fission function
type FunctionInfo struct {
	Name      string
	Namespace string
	Runtime   string
	Env       string
	Status    string
	Package   string
}

// TestFunctionResponse represents the response from testing a function
type TestFunctionResponse struct {
	StatusCode int
	Body       string
	Headers    map[string][]string
}

// LogOptions represents options for retrieving function logs
type LogOptions struct {
	Tail   int    // Number of lines to tail (0 for all)
	Follow bool   // Follow logs (streaming)
	Since  string // Time range (e.g., "5m", "1h")
}
