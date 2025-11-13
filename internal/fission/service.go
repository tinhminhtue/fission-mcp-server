package fission

import (
	"context"
	"encoding/base64"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
)

// Service provides operations for managing Fission functions
type Service struct {
	client dynamic.Interface
}

// NewService creates a new Fission service
func NewService(client dynamic.Interface) *Service {
	return &Service{
		client: client,
	}
}

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

// CreateFunction creates a Fission function with the given name, environment, and code
func (s *Service) CreateFunction(ctx context.Context, name, env, code string) error {
	packageName := name + "-pkg"

	// Check if package already exists
	existingPkg, err := s.client.Resource(PackageGVR).Namespace(DefaultNamespace).Get(ctx, packageName, metav1.GetOptions{})
	if err == nil && existingPkg != nil {
		fmt.Printf("Package '%s' already exists, reusing it...\n", packageName)
	} else {
		// Create a zip archive containing the Python code
		zipData, err := CreateZipArchive(code, "main.py")
		if err != nil {
			return fmt.Errorf("failed to create zip archive: %w", err)
		}

		// Base64 encode the zip archive
		encodedCode := base64.StdEncoding.EncodeToString(zipData)

		// Create package with base64 encoded zip archive
		packageObj := &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "fission.io/v1",
				"kind":       "Package",
				"metadata": map[string]interface{}{
					"name":      packageName,
					"namespace": DefaultNamespace,
				},
				"spec": map[string]interface{}{
					"environment": map[string]interface{}{
						"name":      env,
						"namespace": DefaultNamespace,
					},
					"source": map[string]interface{}{
						"literal": encodedCode,
					},
				},
			},
		}

		_, err = s.client.Resource(PackageGVR).Namespace(DefaultNamespace).Create(ctx, packageObj, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("failed to create package: %w", err)
		}
		fmt.Printf("Created package '%s'\n", packageName)
	}

	// Check if function already exists
	existingFn, err := s.client.Resource(FunctionGVR).Namespace(DefaultNamespace).Get(ctx, name, metav1.GetOptions{})
	if err == nil && existingFn != nil {
		fmt.Printf("Function '%s' already exists. Updating it...\n", name)

		// Update existing function
		existingFn.Object["spec"].(map[string]interface{})["package"] = map[string]interface{}{
			"packageref": map[string]interface{}{
				"name":      packageName,
				"namespace": DefaultNamespace,
			},
		}
		existingFn.Object["spec"].(map[string]interface{})["environment"] = map[string]interface{}{
			"name":      env,
			"namespace": DefaultNamespace,
		}
		// Ensure InvokeStrategy exists
		if _, ok := existingFn.Object["spec"].(map[string]interface{})["InvokeStrategy"]; !ok {
			existingFn.Object["spec"].(map[string]interface{})["InvokeStrategy"] = map[string]interface{}{
				"StrategyType": "execution",
				"ExecutionStrategy": map[string]interface{}{
					"ExecutorType": "poolmgr",
				},
			}
		} else {
			// Ensure StrategyType is set even if InvokeStrategy exists
			invokeStrategy := existingFn.Object["spec"].(map[string]interface{})["InvokeStrategy"].(map[string]interface{})
			invokeStrategy["StrategyType"] = "execution"
		}

		_, err = s.client.Resource(FunctionGVR).Namespace(DefaultNamespace).Update(ctx, existingFn, metav1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("failed to update function: %w", err)
		}
		fmt.Printf("Updated function '%s'\n", name)
		return nil
	}

	// Create new function
	functionObj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "fission.io/v1",
			"kind":       "Function",
			"metadata": map[string]interface{}{
				"name":      name,
				"namespace": DefaultNamespace,
			},
			"spec": map[string]interface{}{
				"package": map[string]interface{}{
					"packageref": map[string]interface{}{
						"name":      packageName,
						"namespace": DefaultNamespace,
					},
				},
				"environment": map[string]interface{}{
					"name":      env,
					"namespace": DefaultNamespace,
				},
				"InvokeStrategy": map[string]interface{}{
					"StrategyType": "execution",
					"ExecutionStrategy": map[string]interface{}{
						"ExecutorType": "poolmgr",
					},
				},
			},
		},
	}

	_, err = s.client.Resource(FunctionGVR).Namespace(DefaultNamespace).Create(ctx, functionObj, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create function: %w", err)
	}

	fmt.Printf("Created function '%s'\n", name)
	return nil
}

