# Contributing to Fission MCP Server

Thank you for your interest in contributing to Fission MCP Server! This document provides guidelines and instructions for contributing.

## Code of Conduct

- Be respectful and inclusive
- Welcome newcomers and help them learn
- Focus on constructive feedback
- Be open to different perspectives

## How to Contribute

### Reporting Bugs

If you find a bug, please open an issue with:
- A clear title and description
- Steps to reproduce the issue
- Expected vs actual behavior
- Environment details (OS, Go version, Kubernetes version, Fission version)
- Relevant logs or error messages

### Suggesting Features

Feature suggestions are welcome! Please open an issue with:
- A clear description of the feature
- Use cases and examples
- Potential implementation approach (if you have ideas)

### Pull Requests

1. **Fork the repository** and create a branch from `main`
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes**
   - Follow Go best practices and conventions
   - Add comments for complex logic
   - Ensure code is formatted with `gofmt`
   - Test your changes locally

3. **Commit your changes**
   - Write clear, descriptive commit messages
   - Use present tense ("Add feature" not "Added feature")
   - Reference issue numbers if applicable

4. **Push and create a Pull Request**
   - Push your branch to your fork
   - Open a PR with a clear description
   - Link to any related issues

## Development Setup

1. Clone your fork:
   ```bash
   git clone https://github.com/your-username/fission-mcp-server.git
   cd fission-mcp-server
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Build and test:
   ```bash
   make build
   make test
   ```

## Code Style

- Follow standard Go formatting (`gofmt`)
- Use `golint` or `golangci-lint` for linting
- Write clear, self-documenting code
- Add comments for exported functions and types
- Keep functions focused and small

## Testing

- Add tests for new features
- Ensure all tests pass before submitting
- Test against different Kubernetes versions if possible

## Questions?

Feel free to open an issue for any questions or clarifications. We're here to help!

Thank you for contributing! 🎉

