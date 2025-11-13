# Fission MCP Server

This project is on development and not yet released.

A Go-based server for managing Fission functions on Kubernetes using the Model Context Protocol (MCP). This project provides a simple and efficient way to create, list, and manage Fission serverless functions programmatically.

## Features

- 🚀 List all Fission functions across namespaces
- 📦 Create and manage Fission functions and packages
- 🔧 Update existing functions
- 🐍 Support for Python runtime environments
- ☸️ Native Kubernetes integration using client-go

## Prerequisites

- Go 1.21 or higher
- Kubernetes cluster with Fission installed
- Valid kubeconfig file (`~/.kube/config` or set `KUBECONFIG` environment variable)
- Fission CRDs installed in your cluster

## Installation

### Clone the repository

```bash
git clone https://github.com/tinhminhtue/fission-mcp-server.git
cd fission-mcp-server
```

### Build

```bash
go build -o fission-mcp-server main.go
```

Or use the Makefile:

```bash
make build
```

### Run

```bash
./fission-mcp-server
```

Or directly with Go:

```bash
go run main.go
```

## Usage

The server will:
1. List all existing Fission functions in your cluster
2. Create a new function named `hello-world2` with Python runtime

### Configuration

You can customize the behavior by modifying the constants in `main.go`:

```go
const (
    functionName     = "hello-world2"  // Function name
    functionEnv      = "python"        // Runtime environment
    defaultNamespace = "ai-code"       // Kubernetes namespace
    pythonCode       = `...`           // Function code
)
```

### Environment Variables

- `KUBECONFIG`: Path to kubeconfig file (defaults to `~/.kube/config`)

## Project Structure

```
fission-mcp-server/
├── main.go           # Main application code
├── go.mod            # Go module dependencies
├── go.sum            # Go module checksums
├── README.md         # This file
├── LICENSE           # License file
├── CONTRIBUTING.md   # Contribution guidelines
├── .gitignore        # Git ignore rules
└── Makefile          # Build automation
```

## Development

### Running Tests

```bash
make test
```

### Building for Different Platforms

```bash
make build-linux    # Build for Linux
make build-darwin   # Build for macOS
make build-windows  # Build for Windows
```

## Contributing

Contributions are welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) for details on our code of conduct and the process for submitting pull requests.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [Fission](https://fission.io/) - Serverless framework for Kubernetes
- [Kubernetes client-go](https://github.com/kubernetes/client-go) - Go client for Kubernetes

## Author

Created by [tinhminhtue](https://github.com/tinhminhtue)

