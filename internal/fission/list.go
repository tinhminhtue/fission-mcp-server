package fission

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// ListFunctions lists all Fission functions across all namespaces
func (s *Service) ListFunctions(ctx context.Context) ([]FunctionInfo, error) {
	functions, err := s.client.Resource(FunctionGVR).Namespace("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("error listing Fission functions: %w", err)
	}

	var functionInfos []FunctionInfo
	for _, fn := range functions.Items {
		info := FunctionInfo{
			Name:      fn.GetName(),
			Namespace: fn.GetNamespace(),
		}

		// Extract function details
		spec, found, _ := unstructured.NestedMap(fn.Object, "spec")
		status, statusFound, _ := unstructured.NestedMap(fn.Object, "status")

		if found {
			if runtime, ok := spec["runtime"].(map[string]interface{}); ok {
				if image, ok := runtime["image"].(string); ok {
					info.Runtime = image
				}
			}
			if env, ok := spec["environment"].(map[string]interface{}); ok {
				if envName, ok := env["name"].(string); ok {
					info.Env = envName
				}
			}
			if pkg, ok := spec["package"].(map[string]interface{}); ok {
				if pkgRef, ok := pkg["packageref"].(map[string]interface{}); ok {
					if pkgName, ok := pkgRef["name"].(string); ok {
						info.Package = pkgName
					}
				}
			}
		}

		if statusFound {
			if statusStr, ok := status["status"].(string); ok {
				info.Status = statusStr
			}
		}

		functionInfos = append(functionInfos, info)
	}

	return functionInfos, nil
}

// PrintFunctions prints the list of functions to stdout
func (s *Service) PrintFunctions(ctx context.Context) error {
	functions, err := s.ListFunctions(ctx)
	if err != nil {
		return err
	}

	fmt.Printf("Found %d Fission function(s):\n\n", len(functions))
	for i, fn := range functions {
		fmt.Printf("%d. %s", i+1, fn.Name)
		if fn.Namespace != "" {
			fmt.Printf(" (namespace: %s)", fn.Namespace)
		}
		fmt.Println()

		if fn.Runtime != "" {
			fmt.Printf("   Runtime Image: %s\n", fn.Runtime)
		}
		if fn.Env != "" {
			fmt.Printf("   Environment: %s\n", fn.Env)
		}
		if fn.Status != "" {
			fmt.Printf("   Status: %s\n", fn.Status)
		}

		fmt.Println()
	}

	if len(functions) == 0 {
		fmt.Println("No Fission functions found.")
		fmt.Println("You can create a function using: fission function create --name <name> --code <code> --env <env>")
	}

	return nil
}
