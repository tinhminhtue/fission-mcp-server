package fission

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
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

	// Get the function invocation URL
	// Like the Fission CLI, we can invoke functions directly through the router
	// without requiring an HTTP trigger. First try to find an HTTP trigger,
	// but if none exists, use the direct router endpoint.
	triggerURL, hasTrigger, err := s.findHTTPTriggerURL(ctx, name, namespace)
	if err != nil {
		// If we can't list triggers, fall back to direct invocation
		triggerURL = s.getDirectRouterURL(name, namespace)
	} else if !hasTrigger {
		// No HTTP trigger found, use direct router invocation (like CLI does)
		triggerURL = s.getDirectRouterURL(name, namespace)
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
// Returns: (triggerURL, hasTrigger, error)
func (s *Service) findHTTPTriggerURL(ctx context.Context, functionName, namespace string) (string, bool, error) {
	// List all HTTP triggers
	triggers, err := s.client.Resource(HTTPTriggerGVR).Namespace("").List(ctx, metav1.ListOptions{})
	if err != nil {
		// If we can't list triggers, return error
		return "", false, fmt.Errorf("failed to list HTTP triggers: %w", err)
	}

	// Find trigger that references this function
	for _, trigger := range triggers.Items {
		spec, found, _ := unstructured.NestedMap(trigger.Object, "spec")
		if !found {
			continue
		}

		// Check function reference - handle both functionref.name and functionref.functionname
		var fnName string
		var fnNamespace string
		if fnRef, ok := spec["functionref"].(map[string]interface{}); ok {
			if name, ok := fnRef["name"].(string); ok {
				fnName = name
			} else if name, ok := fnRef["functionname"].(string); ok {
				fnName = name
			}
			if ns, ok := fnRef["namespace"].(string); ok {
				fnNamespace = ns
			}

			// Check if this trigger references our function
			if fnName == functionName && (fnNamespace == "" || fnNamespace == namespace) {
				// Extract host and path
				host := ""
				if h, ok := spec["host"].(string); ok {
					host = h
				}
				path := "/"
				if p, ok := spec["relativeurl"].(string); ok {
					path = p
				}

				// Construct URL
				// If host is specified, use it; otherwise use Fission router service
				if host != "" {
					// Use the host directly (assumes it's a full URL or hostname)
					if strings.HasPrefix(host, "http://") || strings.HasPrefix(host, "https://") {
						return fmt.Sprintf("%s%s", host, path), true, nil
					}
					return fmt.Sprintf("http://%s%s", host, path), true, nil
				}

				// Use Fission router service (default in-cluster service)
				// Format: {routerURL}{path} where path is the relativeurl from the trigger
				routerURL := s.routerURL
				if strings.HasSuffix(routerURL, "/") {
					routerURL = strings.TrimSuffix(routerURL, "/")
				}
				// Ensure path starts with /
				if !strings.HasPrefix(path, "/") {
					path = "/" + path
				}
				return fmt.Sprintf("%s%s", routerURL, path), true, nil
			}
		}
	}

	// No trigger found
	return "", false, nil
}

// getDirectRouterURL returns the direct router URL for invoking a function
// This matches how the Fission CLI tests functions without HTTP triggers
func (s *Service) getDirectRouterURL(functionName, namespace string) string {
	// Fission router direct invocation format: {routerURL}/{namespace}/{function}
	// This is the same format used by the CLI's "fission function test" command
	routerURL := s.routerURL
	// Remove trailing slash if present
	if strings.HasSuffix(routerURL, "/") {
		routerURL = strings.TrimSuffix(routerURL, "/")
	}
	return fmt.Sprintf("%s/%s/%s", routerURL, namespace, functionName)
}
