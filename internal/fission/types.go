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
)

// FunctionInfo represents information about a Fission function
type FunctionInfo struct {
	Name      string
	Namespace string
	Runtime   string
	Env       string
	Status    string
}

