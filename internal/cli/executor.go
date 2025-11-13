package cli

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/fission/fission/pkg/fission-cli/cmd"
	wCli "github.com/fission/fission/pkg/fission-cli/cliwrapper/cli"
)

// Executor wraps fission-cli command execution
type Executor struct {
	client *cmd.Client
}

// NewExecutor creates a new CLI executor
func NewExecutor(kubeContext, namespace string) (*Executor, error) {
	clientOptions := cmd.ClientOptions{
		KubeContext: kubeContext,
		Namespace:   namespace,
	}

	client, err := cmd.NewClient(clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to create fission client: %w", err)
	}

	// Set the client for command execution
	cmd.SetClientset(*client)

	return &Executor{
		client: client,
	}, nil
}

// ExecuteCommand executes a CLI command action with the given input
func (e *Executor) ExecuteCommand(action cmd.CommandAction, input wCli.Input) error {
	return action(input)
}

// ExecuteCommandWithOutput executes a CLI command and captures output
func (e *Executor) ExecuteCommandWithOutput(action cmd.CommandAction, input wCli.Input) (string, string, error) {
	var stdoutBuf, stderrBuf bytes.Buffer

	// Create new input with custom writers
	httpInput, ok := input.(*HTTPInput)
	if ok {
		httpInput.stdout = &stdoutBuf
		httpInput.stderr = &stderrBuf
	} else {
		// If input is not HTTPInput, we need to wrap it
		// For now, create a new HTTPInput-like wrapper
		// This shouldn't happen in practice since we always use HTTPInput
		return "", "", fmt.Errorf("unexpected input type")
	}

	err := action(input)
	stdout := stdoutBuf.String()
	stderr := stderrBuf.String()

	return stdout, stderr, err
}

// GetClient returns the underlying fission client
func (e *Executor) GetClient() *cmd.Client {
	return e.client
}

// SetClient sets the fission client (useful for testing or reinitialization)
func (e *Executor) SetClient(client *cmd.Client) {
	e.client = client
	cmd.SetClientset(*client)
}

// NewHTTPInputFromRequest creates an HTTPInput from an HTTP request
func NewHTTPInputFromRequest(r *http.Request) wCli.Input {
	return NewHTTPInput(r, os.Stdout, os.Stderr)
}

// NewHTTPInputFromRequestWithWriters creates an HTTPInput with custom writers
func NewHTTPInputFromRequestWithWriters(r *http.Request, stdout, stderr io.Writer) wCli.Input {
	return NewHTTPInput(r, stdout, stderr)
}

