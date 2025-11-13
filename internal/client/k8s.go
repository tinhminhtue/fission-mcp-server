package client

import (
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
)

// NewDynamicClient creates a new Kubernetes dynamic client from kubeconfig
func NewDynamicClient() (dynamic.Interface, error) {
	// Get the kubeconfig file path (defaults to ~/.kube/config)
	kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	if kubeconfigPath := os.Getenv("KUBECONFIG"); kubeconfigPath != "" {
		kubeconfig = kubeconfigPath
	}

	// Build the config from the kubeconfig file
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("error building kubeconfig: %w", err)
	}

	// Create the dynamic client
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("error creating dynamic client: %w", err)
	}

	return dynamicClient, nil
}

