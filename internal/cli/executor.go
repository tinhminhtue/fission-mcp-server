package cli

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"

	wCli "github.com/fission/fission/pkg/fission-cli/cliwrapper/cli"
	"github.com/fission/fission/pkg/fission-cli/cmd"
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

	// Redirect stdout and stderr to capture command output
	oldStdout := os.Stdout
	oldStderr := os.Stderr

	// Create pipes to capture output
	stdoutR, stdoutW, _ := os.Pipe()
	stderrR, stderrW, _ := os.Pipe()

	// Redirect stdout and stderr
	os.Stdout = stdoutW
	os.Stderr = stderrW

	// Create a channel to signal when goroutine is done
	done := make(chan error, 1)

	// Start goroutine to copy output to our buffers
	go func() {
		_, err := io.Copy(&stdoutBuf, stdoutR)
		if err != nil {
			done <- err
			return
		}
		_, err = io.Copy(&stderrBuf, stderrR)
		done <- err
	}()

	// Execute the command
	err := action(input)

	// Close writers to signal copying is complete
	stdoutW.Close()
	stderrW.Close()

	// Wait for copying to complete
	copyErr := <-done

	// Restore original stdout and stderr
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	// If there was a copy error, return it
	if copyErr != nil {
		return "", "", copyErr
	}

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
