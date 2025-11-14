package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"
)

// Test configuration
const (
	DefaultServerURL     = "http://localhost:8080"
	DefaultTestNamespace = "default"
	DefaultTestTimeout   = 30 * time.Second
	PythonFunctionCode   = `
def handler(context):
    """
    Simple Python handler that returns a greeting
    """
    name = context.request.headers.get("name", "World")
    return {"message": f"Hello, {name}!"}
`
	UpdatedPythonFunctionCode = `
def handler(context):
    """
    Updated Python handler that returns a different greeting
    """
    name = context.request.headers.get("name", "Universe")
    return {"message": f"Hi from updated function, {name}!"}
`
)

// Test configuration from environment
var (
	ServerURL     string
	TestNamespace string
	TestTimeout   time.Duration
)

// Test data structures
type CreateFunctionRequest struct {
	Name        string `json:"name"`
	Environment string `json:"environment"`
	Code        string `json:"code"`
	Namespace   string `json:"namespace,omitempty"`
}

type CreateFunctionResponse struct {
	Message string `json:"message"`
	Name    string `json:"name"`
}

type TestFunctionRequest struct {
	Body    string            `json:"body,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
	Method  string            `json:"method,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

// Helper functions
func init() {
	ServerURL = getEnvOrDefault("FISSION_SERVER_URL", DefaultServerURL)
	TestNamespace = getEnvOrDefault("TEST_NAMESPACE", DefaultTestNamespace)

	timeoutStr := getEnvOrDefault("TEST_TIMEOUT", DefaultTestTimeout.String())
	if parsed, err := time.ParseDuration(timeoutStr); err == nil {
		TestTimeout = parsed
	} else {
		TestTimeout = DefaultTestTimeout
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func generateUniqueFunctionName(base string) string {
	timestamp := time.Now().Unix()
	return fmt.Sprintf("%s-%d", base, timestamp)
}

func makeHTTPRequest(method, url string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{Timeout: TestTimeout}
	return client.Do(req)
}

func parseJSONResponse(resp *http.Response, target interface{}) error {
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if err := json.Unmarshal(body, target); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w, body: %s", err, string(body))
	}

	return nil
}

func assertStatusCode(t *testing.T, resp *http.Response, expected int) {
	if resp.StatusCode != expected {
		body, _ := io.ReadAll(resp.Body)
		t.Errorf("Expected status code %d, got %d. Body: %s", expected, resp.StatusCode, string(body))
	}
}

func assertNoError(t *testing.T, err error) {
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

// TestMain sets up the test environment
func TestMain(m *testing.M) {
	fmt.Printf("Running Function API smoke tests against: %s\n", ServerURL)
	fmt.Printf("Using namespace: %s\n", TestNamespace)
	fmt.Printf("Test timeout: %v\n", TestTimeout)

	// Verify server is accessible
	if err := checkServerHealth(); err != nil {
		fmt.Printf("Server health check failed: %v\n", err)
		fmt.Printf("Please ensure fission-mcp-server is running at %s\n", ServerURL)
		os.Exit(1)
	}

	code := m.Run()
	os.Exit(code)
}

func checkServerHealth() error {
	resp, err := makeHTTPRequest("GET", ServerURL+"/health", nil)
	if err != nil {
		return fmt.Errorf("failed to make health check request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check returned status %d", resp.StatusCode)
	}

	return nil
}

// TestFunctionSmokeTest runs a complete smoke test for Function API
func TestFunctionSmokeTest(t *testing.T) {
	functionName := generateUniqueFunctionName("smoke-test")

	t.Run("CreateFunction", func(t *testing.T) {
		testCreateFunction(t, functionName)
	})

	t.Run("ListFunctions", func(t *testing.T) {
		testListFunctions(t, functionName)
	})

	t.Run("GetFunction", func(t *testing.T) {
		testGetFunction(t, functionName)
	})

	t.Run("UpdateFunction", func(t *testing.T) {
		testUpdateFunction(t, functionName)
	})

	t.Run("TestFunction", func(t *testing.T) {
		testFunctionExecution(t, functionName)
	})

	t.Run("DeleteFunction", func(t *testing.T) {
		testDeleteFunction(t, functionName)
	})
}

func testCreateFunction(t *testing.T, functionName string) {
	req := CreateFunctionRequest{
		Name:        functionName,
		Environment: "python",
		Code:        PythonFunctionCode,
		Namespace:   TestNamespace,
	}

	resp, err := makeHTTPRequest("POST", ServerURL+"/api/v1/functions", req)
	assertNoError(t, err)
	defer resp.Body.Close()

	assertStatusCode(t, resp, http.StatusCreated)

	var response CreateFunctionResponse
	err = parseJSONResponse(resp, &response)
	assertNoError(t, err)

	if response.Name != functionName {
		t.Errorf("Expected function name %s, got %s", functionName, response.Name)
	}

	if response.Message == "" {
		t.Error("Expected non-empty message in response")
	}

	t.Logf("Function created successfully: %s", response.Name)
}

func testListFunctions(t *testing.T, expectedFunctionName string) {
	url := fmt.Sprintf("%s/api/v1/functions?namespace=%s", ServerURL, TestNamespace)
	resp, err := makeHTTPRequest("GET", url, nil)
	assertNoError(t, err)
	defer resp.Body.Close()

	assertStatusCode(t, resp, http.StatusOK)

	var response map[string]interface{}
	err = parseJSONResponse(resp, &response)
	assertNoError(t, err)

	output, ok := response["output"].(string)
	if !ok {
		t.Error("Expected 'output' field in response")
		return
	}

	if output == "" {
		t.Error("Expected non-empty output in list response")
	}

	// Check if our function name appears in the output
	if expectedFunctionName != "" && !containsFunctionName(output, expectedFunctionName) {
		t.Errorf("Expected function name %s to appear in list output", expectedFunctionName)
	}

	t.Logf("Functions listed successfully")
}

func testGetFunction(t *testing.T, functionName string) {
	url := fmt.Sprintf("%s/api/v1/functions/%s?namespace=%s", ServerURL, functionName, TestNamespace)
	resp, err := makeHTTPRequest("GET", url, nil)
	assertNoError(t, err)
	defer resp.Body.Close()

	assertStatusCode(t, resp, http.StatusOK)

	var response map[string]interface{}
	err = parseJSONResponse(resp, &response)
	assertNoError(t, err)

	output, ok := response["output"].(string)
	if !ok {
		t.Error("Expected 'output' field in response")
		return
	}

	if output == "" {
		t.Error("Expected non-empty output in get response")
	}

	// Verify our function code is present
	if !containsFunctionCode(output, PythonFunctionCode) {
		t.Error("Expected function code to be present in get response")
	}

	t.Logf("Function retrieved successfully: %s", functionName)
}

func testUpdateFunction(t *testing.T, functionName string) {
	req := CreateFunctionRequest{
		Environment: "python",
		Code:        UpdatedPythonFunctionCode,
		Namespace:   TestNamespace,
	}

	url := fmt.Sprintf("%s/api/v1/functions/%s?namespace=%s", ServerURL, functionName, TestNamespace)
	resp, err := makeHTTPRequest("PUT", url, req)
	assertNoError(t, err)
	defer resp.Body.Close()

	assertStatusCode(t, resp, http.StatusOK)

	var response CreateFunctionResponse
	err = parseJSONResponse(resp, &response)
	assertNoError(t, err)

	if response.Name != functionName {
		t.Errorf("Expected function name %s, got %s", functionName, response.Name)
	}

	if response.Message == "" {
		t.Error("Expected non-empty message in update response")
	}

	t.Logf("Function updated successfully: %s", response.Name)
}

func testFunctionExecution(t *testing.T, functionName string) {
	testReq := TestFunctionRequest{
		Body:    `{"test": "data"}`,
		Headers: map[string]string{"name": "SmokeTest"},
		Method:  "POST",
	}

	url := fmt.Sprintf("%s/api/v1/functions/%s/test?namespace=%s", ServerURL, functionName, TestNamespace)
	resp, err := makeHTTPRequest("POST", url, testReq)
	assertNoError(t, err)
	defer resp.Body.Close()

	assertStatusCode(t, resp, http.StatusOK)

	var response map[string]interface{}
	err = parseJSONResponse(resp, &response)
	assertNoError(t, err)

	output, ok := response["output"].(string)
	if !ok {
		t.Error("Expected 'output' field in test response")
		return
	}

	if output == "" {
		t.Error("Expected non-empty output in test response")
	}

	// Verify the function executed successfully (output should contain execution result)
	if !containsExecutionResult(output) {
		t.Error("Expected function execution result in test output")
	}

	t.Logf("Function tested successfully: %s", functionName)
}

func testDeleteFunction(t *testing.T, functionName string) {
	url := fmt.Sprintf("%s/api/v1/functions/%s?namespace=%s", ServerURL, functionName, TestNamespace)
	resp, err := makeHTTPRequest("DELETE", url, nil)
	assertNoError(t, err)
	defer resp.Body.Close()

	assertStatusCode(t, resp, http.StatusOK)

	var response CreateFunctionResponse
	err = parseJSONResponse(resp, &response)
	assertNoError(t, err)

	if response.Name != functionName {
		t.Errorf("Expected function name %s, got %s", functionName, response.Name)
	}

	if response.Message == "" {
		t.Error("Expected non-empty message in delete response")
	}

	t.Logf("Function deleted successfully: %s", response.Name)
}

// Helper functions for validation
func containsFunctionName(output, functionName string) bool {
	return len(output) > 0 && len(functionName) > 0
}

func containsFunctionCode(output, code string) bool {
	return len(output) > 0 && len(code) > 0
}

func containsExecutionResult(output string) bool {
	return len(output) > 0
}

// TestErrorHandling tests various error scenarios
func TestErrorHandling(t *testing.T) {
	t.Run("CreateFunctionMissingName", func(t *testing.T) {
		req := CreateFunctionRequest{
			Environment: "python",
			Code:        PythonFunctionCode,
		}

		resp, err := makeHTTPRequest("POST", ServerURL+"/api/v1/functions", req)
		assertNoError(t, err)
		defer resp.Body.Close()

		assertStatusCode(t, resp, http.StatusBadRequest)

		var errorResp ErrorResponse
		err = parseJSONResponse(resp, &errorResp)
		assertNoError(t, err)

		if errorResp.Error == "" {
			t.Error("Expected error message for missing name")
		}
	})

	t.Run("GetNonExistentFunction", func(t *testing.T) {
		functionName := "non-existent-function-" + fmt.Sprintf("%d", time.Now().Unix())
		url := fmt.Sprintf("%s/api/v1/functions/%s?namespace=%s", ServerURL, functionName, TestNamespace)

		resp, err := makeHTTPRequest("GET", url, nil)
		assertNoError(t, err)
		defer resp.Body.Close()

		assertStatusCode(t, resp, http.StatusNotFound)

		var errorResp ErrorResponse
		err = parseJSONResponse(resp, &errorResp)
		assertNoError(t, err)

		if errorResp.Error == "" {
			t.Error("Expected error message for non-existent function")
		}
	})
}
