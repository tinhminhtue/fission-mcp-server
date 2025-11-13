package fission

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// GetFunction retrieves a Fission function by name and namespace
func (s *Service) GetFunction(ctx context.Context, name, namespace string) (*FunctionInfo, error) {
	if namespace == "" {
		namespace = DefaultNamespace
	}

	// Get the function
	fn, err := s.client.Resource(FunctionGVR).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("function '%s' not found in namespace '%s': %w", name, namespace, err)
	}

	// Extract function details
	info := &FunctionInfo{
		Name:      fn.GetName(),
		Namespace: fn.GetNamespace(),
	}

	// Extract spec details
	spec, found, _ := unstructured.NestedMap(fn.Object, "spec")
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

	// Extract status details
	status, statusFound, _ := unstructured.NestedMap(fn.Object, "status")
	if statusFound {
		if statusStr, ok := status["status"].(string); ok {
			info.Status = statusStr
		}
	}

	return info, nil
}

