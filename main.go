package main

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	functionName     = "hello-world2"
	functionEnv      = "python"
	defaultNamespace = "ai-code"
	pythonCode       = `def main():
    return "Hello, World from hello-world2!"
`
)

func main() {
	// Get the kubeconfig file path (defaults to ~/.kube/config)
	kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	if kubeconfigPath := os.Getenv("KUBECONFIG"); kubeconfigPath != "" {
		kubeconfig = kubeconfigPath
	}

	// Build the config from the kubeconfig file
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error building kubeconfig: %v\n", err)
		os.Exit(1)
	}

	// Create the dynamic client
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating dynamic client: %v\n", err)
		os.Exit(1)
	}

	// Define the Fission Function CRD resource
	functionGVR := schema.GroupVersionResource{
		Group:    "fission.io",
		Version:  "v1",
		Resource: "functions",
	}

	// List Fission functions across all namespaces
	functions, err := dynamicClient.Resource(functionGVR).Namespace("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing Fission functions: %v\n", err)
		fmt.Fprintf(os.Stderr, "Make sure Fission is installed and the Function CRD exists.\n")
		os.Exit(1)
	}

	// Print the functions
	fmt.Printf("Found %d Fission function(s):\n\n", len(functions.Items))
	for i, fn := range functions.Items {
		name := fn.GetName()
		namespace := fn.GetNamespace()

		// Extract function details
		spec, found, _ := unstructured.NestedMap(fn.Object, "spec")
		status, statusFound, _ := unstructured.NestedMap(fn.Object, "status")

		fmt.Printf("%d. %s", i+1, name)
		if namespace != "" {
			fmt.Printf(" (namespace: %s)", namespace)
		}
		fmt.Println()

		// Display runtime if available
		if found {
			if runtime, ok := spec["runtime"].(map[string]interface{}); ok {
				if image, ok := runtime["image"].(string); ok {
					fmt.Printf("   Runtime Image: %s\n", image)
				}
			}
			if env, ok := spec["environment"].(map[string]interface{}); ok {
				if envName, ok := env["name"].(string); ok {
					fmt.Printf("   Environment: %s\n", envName)
				}
			}
		}

		// Display status if available
		if statusFound {
			if statusStr, ok := status["status"].(string); ok {
				fmt.Printf("   Status: %s\n", statusStr)
			}
		}

		fmt.Println()
	}

	// If no functions found, provide helpful message
	if len(functions.Items) == 0 {
		fmt.Println("No Fission functions found.")
		fmt.Println("You can create a function using: fission function create --name <name> --code <code> --env <env>")
	}

	// Create hello-world2 function
	fmt.Println("\n==================================================")
	fmt.Println("Creating hello-world2 function...")
	fmt.Println("==================================================")

	err = createFissionFunction(dynamicClient, functionGVR, functionName, functionEnv, pythonCode)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating function: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully created function '%s'!\n", functionName)
}

// createFissionFunction creates a Fission function with the given name, environment, and code
func createFissionFunction(client dynamic.Interface, gvr schema.GroupVersionResource, name, env, code string) error {
	// First, we need to create a Package that contains the code
	// Then create a Function that references the package

	// Create package GVR
	packageGVR := schema.GroupVersionResource{
		Group:    "fission.io",
		Version:  "v1",
		Resource: "packages",
	}

	packageName := name + "-pkg"

	// Check if package already exists
	existingPkg, err := client.Resource(packageGVR).Namespace(defaultNamespace).Get(context.TODO(), packageName, metav1.GetOptions{})
	if err == nil && existingPkg != nil {
		fmt.Printf("Package '%s' already exists, reusing it...\n", packageName)
	} else {
		// Create a zip archive containing the Python code
		zipData, err := createZipArchive(code, "main.py")
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
					"namespace": defaultNamespace,
				},
				"spec": map[string]interface{}{
					"environment": map[string]interface{}{
						"name":      env,
						"namespace": defaultNamespace,
					},
					"source": map[string]interface{}{
						"literal": encodedCode,
					},
				},
			},
		}

		_, err = client.Resource(packageGVR).Namespace(defaultNamespace).Create(context.TODO(), packageObj, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("failed to create package: %w", err)
		}
		fmt.Printf("Created package '%s'\n", packageName)
	}

	// Check if function already exists
	existingFn, err := client.Resource(gvr).Namespace(defaultNamespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err == nil && existingFn != nil {
		fmt.Printf("Function '%s' already exists. Updating it...\n", name)

		// Update existing function
		existingFn.Object["spec"].(map[string]interface{})["package"] = map[string]interface{}{
			"packageref": map[string]interface{}{
				"name":      packageName,
				"namespace": defaultNamespace,
			},
		}
		existingFn.Object["spec"].(map[string]interface{})["environment"] = map[string]interface{}{
			"name":      env,
			"namespace": defaultNamespace,
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

		_, err = client.Resource(gvr).Namespace(defaultNamespace).Update(context.TODO(), existingFn, metav1.UpdateOptions{})
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
				"namespace": defaultNamespace,
			},
			"spec": map[string]interface{}{
				"package": map[string]interface{}{
					"packageref": map[string]interface{}{
						"name":      packageName,
						"namespace": defaultNamespace,
					},
				},
				"environment": map[string]interface{}{
					"name":      env,
					"namespace": defaultNamespace,
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

	_, err = client.Resource(gvr).Namespace(defaultNamespace).Create(context.TODO(), functionObj, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create function: %w", err)
	}

	fmt.Printf("Created function '%s'\n", name)
	return nil
}

// createZipArchive creates a zip archive containing the given code in a file
func createZipArchive(code, filename string) ([]byte, error) {
	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)

	// Create a file in the zip archive
	fileWriter, err := zipWriter.Create(filename)
	if err != nil {
		zipWriter.Close()
		return nil, err
	}

	// Write the code to the file
	_, err = fileWriter.Write([]byte(code))
	if err != nil {
		zipWriter.Close()
		return nil, err
	}

	// Close the zip writer
	err = zipWriter.Close()
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
