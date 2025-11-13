package fission

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DeleteFunction deletes a Fission function by name and namespace
func (s *Service) DeleteFunction(ctx context.Context, name, namespace string) error {
	if namespace == "" {
		namespace = DefaultNamespace
	}

	// Check if function exists
	_, err := s.client.Resource(FunctionGVR).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("function '%s' not found in namespace '%s': %w", name, namespace, err)
	}

	// Delete the function
	err = s.client.Resource(FunctionGVR).Namespace(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete function '%s': %w", name, err)
	}

	fmt.Printf("Deleted function '%s' from namespace '%s'\n", name, namespace)
	return nil
}

