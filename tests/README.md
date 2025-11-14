# Function API Smoke Tests

This directory contains integration smoke tests for the Fission MCP Server Function API endpoints. These tests verify that all Function CRUD operations work correctly against a real Fission server running on Kubernetes.

## Prerequisites

1. **Fission MCP Server Running**: The fission-mcp-server must be running and accessible
2. **Fission Installed**: A working Fission installation on your Kubernetes cluster
3. **Valid kubeconfig**: Proper Kubernetes configuration to access the cluster

## Running the Tests

### 1. Start the Fission MCP Server

```bash
# From the project root directory
./fission-mcp-server

# Or build and run
go build -o fission-mcp-server ./cmd/server
./fission-mcp-server
```

The server will start on `http://localhost:8080` by default.

### 2. Run the Smoke Tests

In a separate terminal, run:

```bash
# Basic test run
go test -v ./tests/

# Run with specific test
go test -v ./tests/ -run TestFunctionSmokeTest

# Run error handling tests
go test -v ./tests/ -run TestErrorHandling
```

### 3. Configuration Options

You can customize the test behavior using environment variables:

```bash
# Custom server URL
FISSION_SERVER_URL=http://localhost:9090 go test -v ./tests/

# Custom namespace
TEST_NAMESPACE=my-test-ns go test -v ./tests/

# Custom timeout
TEST_TIMEOUT=60s go test -v ./tests/

# Combine multiple options
FISSION_SERVER_URL=http://localhost:9090 TEST_NAMESPACE=test-ns TEST_TIMEOUT=45s go test -v ./tests/
```

## Test Coverage

The smoke tests cover the following Function API operations:

### CRUD Operations
- **Create Function**: `POST /api/v1/functions`
- **List Functions**: `GET /api/v1/functions`
- **Get Function**: `GET /api/v1/functions/{name}`
- **Update Function**: `PUT /api/v1/functions/{name}`
- **Delete Function**: `DELETE /api/v1/functions/{name}`

### Function Execution
- **Test Function**: `POST /api/v1/functions/{name}/test`

### Error Handling
- **Validation Errors**: Missing required fields
- **Not Found Errors**: Accessing non-existent resources

## Test Flow

The main smoke test (`TestFunctionSmokeTest`) executes a complete workflow:

1. **Create** a new Python function
2. **List** functions and verify the new function appears
3. **Get** the function details and verify the code
4. **Update** the function with new code
5. **Test** the function execution
6. **Delete** the function and verify cleanup

## Test Data

The tests use sample Python functions:

### Original Function
```python
def handler(context):
    """
    Simple Python handler that returns a greeting
    """
    name = context.request.headers.get("name", "World")
    return {"message": f"Hello, {name}!"}
```

### Updated Function
```python
def handler(context):
    """
    Updated Python handler that returns a different greeting
    """
    name = context.request.headers.get("name", "Universe")
    return {"message": f"Hi from updated function, {name}!"}
```

## Expected Output

When all tests pass, you should see output similar to:

```
Running Function API smoke tests against: http://localhost:8080
Using namespace: default
Test timeout: 30s
=== RUN   TestFunctionSmokeTest
=== RUN   TestFunctionSmokeTest/CreateFunction
    function_smoke_test.go:xxx: Function created successfully: smoke-test-1234567890
=== RUN   TestFunctionSmokeTest/ListFunctions
    function_smoke_test.go:xxx: Functions listed successfully
=== RUN   TestFunctionSmokeTest/GetFunction
    function_smoke_test.go:xxx: Function retrieved successfully: smoke-test-1234567890
=== RUN   TestFunctionSmokeTest/UpdateFunction
    function_smoke_test.go:xxx: Function updated successfully: smoke-test-1234567890
=== RUN   TestFunctionSmokeTest/TestFunction
    function_smoke_test.go:xxx: Function tested successfully: smoke-test-1234567890
=== RUN   TestFunctionSmokeTest/DeleteFunction
    function_smoke_test.go:xxx: Function deleted successfully: smoke-test-1234567890
=== RUN   TestErrorHandling
=== RUN   TestErrorHandling/CreateFunctionMissingName
=== RUN   TestErrorHandling/GetNonExistentFunction
--- PASS: TestFunctionSmokeTest (15.23s)
--- PASS: TestErrorHandling (2.45s)
PASS
```

## Troubleshooting

### Server Health Check Failed
```
Server health check failed: ...
Please ensure fission-mcp-server is running at http://localhost:8080
```

**Solution**: Make sure the fission-mcp-server is running and accessible at the specified URL.

### Function Creation Fails
```
Expected status code 201, got 500
```

**Solution**: Check that:
- Fission is properly installed on your cluster
- The target namespace exists
- You have sufficient permissions

### Test Timeouts
```
Expected status code 200, got context deadline exceeded
```

**Solution**: Increase the timeout using the `TEST_TIMEOUT` environment variable.

## Integration with CI/CD

These tests can be integrated into CI/CD pipelines:

```yaml
# Example GitHub Actions
- name: Run Function API Smoke Tests
  run: |
    # Start fission-mcp-server in background
    ./fission-mcp-server &
    SERVER_PID=$!
    
    # Wait for server to be ready
    sleep 10
    
    # Run smoke tests
    FISSION_SERVER_URL=http://localhost:8080 go test -v ./tests/
    
    # Cleanup
    kill $SERVER_PID
  env:
    KUBECONFIG: ${{ secrets.KUBECONFIG }}
```

## Notes

- Tests create unique function names using timestamps to avoid conflicts
- All test resources are cleaned up after test completion
- Tests are designed to be run against a real Fission installation
- The tests verify HTTP status codes and response structures according to the OpenAPI specification