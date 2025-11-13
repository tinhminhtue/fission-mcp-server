package fission

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// TestFunction invokes a Fission function and returns the response
func (s *Service) TestFunction(ctx context.Context, name, namespace string, body string, headers map[string]string) (*TestFunctionResponse, error) {
	if namespace == "" {
		namespace = DefaultNamespace
	}

	// Get the function to verify it exists
	_, err := s.client.Resource(FunctionGVR).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("function '%s' not found in namespace '%s': %w", name, namespace, err)
	}

	// Find HTTP trigger for this function
	triggerURL, err := s.findHTTPTriggerURL(ctx, name, namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to find HTTP trigger for function '%s': %w", name, err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, triggerURL, bytes.NewBufferString(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set default headers
	req.Header.Set("Content-Type", "application/json")
	if body == "" {
		req.Header.Set("Content-Type", "text/plain")
	}

	// Add custom headers
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	// Make HTTP request
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to invoke function: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Extract response headers
	respHeaders := make(map[string][]string)
	for k, v := range resp.Header {
		respHeaders[k] = v
	}

	return &TestFunctionResponse{
		StatusCode: resp.StatusCode,
		Body:       string(respBody),
		Headers:    respHeaders,
	}, nil
}

// findHTTPTriggerURL finds the HTTP trigger URL for a given function
func (s *Service) findHTTPTriggerURL(ctx context.Context, functionName, namespace string) (string, error) {
	// List all HTTP triggers
	triggers, err := s.client.Resource(HTTPTriggerGVR).Namespace("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to list HTTP triggers: %w", err)
	}

	// Find trigger that references this function
	for _, trigger := range triggers.Items {
		spec, found, _ := unstructured.NestedMap(trigger.Object, "spec")
		if !found {
			continue
		}

		// Check function reference
		if fnRef, ok := spec["functionref"].(map[string]interface{}); ok {
			if fnName, ok := fnRef["name"].(string); ok && fnName == functionName {
				// Extract host and path
				host := ""
				if h, ok := spec["host"].(string); ok {
					host = h
				}
				path := "/"
				if p, ok := spec["relativeurl"].(string); ok {
					path = p
				}

				// Construct URL (assuming Fission router is accessible)
				// In production, this would use the Fission router service
				if host != "" {
					return fmt.Sprintf("http://%s%s", host, path), nil
				}
				// Fallback: try to use Fission router service
				// This is a simplified approach - in production, you'd query the router service
				return fmt.Sprintf("http://router.fission/%s%s", functionName, path), nil
			}
		}
	}

	// If no trigger found, try direct invocation via Fission router
	// This assumes the function can be invoked via /fission-function/{name}
	return fmt.Sprintf("http://router.fission/fission-function/%s", functionName), nil
}

