# Contributing to Secret Santa

We welcome contributions to Secret Santa! This guide will help you get started with contributing to the project.

## Getting Started

### Prerequisites

- Go 1.21+
- Docker
- Kubernetes cluster (kind, minikube, or cloud provider)
- kubectl configured
- Helm 3.x

### Development Setup

1. **Fork and Clone**

```bash
git clone https://github.com/LogicIQ/secret-santa.git
cd secret-santa
```

2. **Install Dependencies**

```bash
go mod download
```

3. **Run Tests**

```bash
make test
```

4. **Build and Run Locally**

```bash
# Build the manager binary
make build

# Install CRDs
make install

# Run the controller locally
make run
```

## Development Workflow

### Making Changes

1. **Create a Feature Branch**

```bash
git checkout -b feature/my-new-feature
```

2. **Make Your Changes**

Follow the existing code patterns and conventions:
- Use meaningful variable and function names
- Add comments for complex logic
- Follow Go best practices
- Update tests for new functionality

3. **Run Tests and Linting**

```bash
# Run unit tests
make test

# Run linting
make lint

# Run end-to-end tests
make test-e2e
```

4. **Update Documentation**

- Update relevant documentation in the `docs/` directory
- Add examples for new features
- Update the API reference if needed

### Code Structure

```
secret-santa/
├── api/v1alpha1/          # API definitions
├── cmd/                   # Main application entry point
├── config/                # Kubernetes manifests
├── docs/                  # Documentation
├── examples/              # Example manifests
├── internal/              # Internal packages
│   ├── controller/        # Controller logic
│   └── config/           # Configuration
├── pkg/                   # Public packages
│   ├── generators/        # Secret generators
│   ├── media/            # Media providers
│   ├── template/         # Template processing
│   └── validation/       # Input validation
└── e2e/                  # End-to-end tests
```

## Adding New Features

### Adding a New Generator

1. **Create the Generator**

```go
// pkg/generators/my_generator.go
package generators

import (
    "context"
    "fmt"
)

type MyGeneratorConfig struct {
    Length int    `json:"length"`
    Format string `json:"format"`
}

type MyGenerator struct {
    config MyGeneratorConfig
}

func NewMyGenerator(config map[string]interface{}) (Generator, error) {
    var cfg MyGeneratorConfig
    if err := mapstructure.Decode(config, &cfg); err != nil {
        return nil, fmt.Errorf("invalid config: %w", err)
    }
    
    // Validate configuration
    if cfg.Length <= 0 {
        return nil, fmt.Errorf("length must be positive")
    }
    
    return &MyGenerator{config: cfg}, nil
}

func (g *MyGenerator) Generate(ctx context.Context) (map[string]interface{}, error) {
    // Implementation here
    result := map[string]interface{}{
        "value": "generated-value",
    }
    return result, nil
}

func (g *MyGenerator) Type() string {
    return "my_generator"
}
```

2. **Register the Generator**

```go
// pkg/generators/registry.go
func init() {
    Register("my_generator", NewMyGenerator)
}
```

3. **Add Tests**

```go
// pkg/generators/my_generator_test.go
package generators

import (
    "context"
    "testing"
)

func TestMyGenerator(t *testing.T) {
    config := map[string]interface{}{
        "length": 10,
        "format": "hex",
    }
    
    gen, err := NewMyGenerator(config)
    if err != nil {
        t.Fatalf("failed to create generator: %v", err)
    }
    
    result, err := gen.Generate(context.Background())
    if err != nil {
        t.Fatalf("failed to generate: %v", err)
    }
    
    value, ok := result["value"].(string)
    if !ok {
        t.Fatal("expected string value")
    }
    
    if len(value) != 10 {
        t.Errorf("expected length 10, got %d", len(value))
    }
}
```

4. **Update Documentation**

Add documentation in `docs/guides/generators.md` and create an example in `docs/examples/`.

### Adding a New Media Provider

1. **Implement the Provider Interface**

```go
// pkg/media/my_provider.go
package media

import (
    "context"
    "fmt"
)

type MyProviderConfig struct {
    Endpoint string `json:"endpoint"`
    Token    string `json:"token"`
}

type MyProvider struct {
    config MyProviderConfig
}

func NewMyProvider(config map[string]interface{}) (Provider, error) {
    var cfg MyProviderConfig
    if err := mapstructure.Decode(config, &cfg); err != nil {
        return nil, fmt.Errorf("invalid config: %w", err)
    }
    
    return &MyProvider{config: cfg}, nil
}

func (p *MyProvider) Store(ctx context.Context, name string, data []byte, metadata map[string]string) error {
    // Implementation here
    return nil
}

func (p *MyProvider) Type() string {
    return "my_provider"
}
```

2. **Register the Provider**

```go
// pkg/media/registry.go
func init() {
    Register("my_provider", NewMyProvider)
}
```

3. **Add Tests and Documentation**

Follow the same pattern as generators.

## Testing

### Unit Tests

```bash
# Run all unit tests
make test

# Run tests with coverage
make test-coverage

# Run specific package tests
go test ./pkg/generators/...
```

### Integration Tests

```bash
# Run integration tests (requires Kubernetes cluster)
make test-integration
```

### End-to-End Tests

```bash
# Run e2e tests (requires Kubernetes cluster)
make test-e2e
```

### Manual Testing

1. **Deploy to Test Cluster**

```bash
# Build and load image
make docker-build
kind load docker-image secret-santa:latest

# Deploy to cluster
make deploy
```

2. **Test with Examples**

```bash
kubectl apply -f examples/basic-password.yaml
kubectl get secretsanta
kubectl get secrets
```

## Documentation

### Writing Documentation

- Use clear, concise language
- Include practical examples
- Follow the existing structure
- Test all code examples

### Building Documentation Locally

The documentation is built using Docusaurus as part of the logiciq.ca website:

```bash
cd ../logiciq.ca/docusaurus
npm install
npm start
```

Navigate to `http://localhost:3000/secret-santa/docs` to view the documentation.

## Submitting Changes

### Pull Request Process

1. **Ensure Tests Pass**

```bash
make test
make lint
make test-e2e
```

2. **Update Documentation**

- Update relevant docs
- Add examples for new features
- Update API reference if needed

3. **Create Pull Request**

- Use a descriptive title
- Include a detailed description
- Reference any related issues
- Add screenshots for UI changes

4. **Code Review**

- Address reviewer feedback
- Keep the PR focused and small
- Rebase if needed to maintain clean history

### Commit Message Format

Use conventional commit format:

```
type(scope): description

[optional body]

[optional footer]
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `test`: Test changes
- `refactor`: Code refactoring
- `chore`: Maintenance tasks

Examples:
```
feat(generators): add UUID generator
fix(media): handle AWS authentication errors
docs(examples): add TLS certificate example
```

## Release Process

### Versioning

We follow [Semantic Versioning](https://semver.org/):
- `MAJOR.MINOR.PATCH`
- Breaking changes increment MAJOR
- New features increment MINOR
- Bug fixes increment PATCH

### Creating a Release

1. **Update Version**

```bash
# Update version in relevant files
make version VERSION=v0.3.0
```

2. **Create Release PR**

- Update CHANGELOG.md
- Update documentation
- Create PR with version changes

3. **Tag Release**

```bash
git tag v0.3.0
git push origin v0.3.0
```

4. **GitHub Actions**

The CI/CD pipeline will automatically:
- Build and test
- Create container images
- Update Helm charts
- Create GitHub release

## Getting Help

### Communication Channels

- **GitHub Issues**: Bug reports and feature requests
- **GitHub Discussions**: Questions and general discussion
- **LogicIQ Website**: https://logiciq.ca/secret-santa

### Code of Conduct

Please follow our [Code of Conduct](CODE_OF_CONDUCT.md) in all interactions.

## Recognition

Contributors are recognized in:
- GitHub contributors list
- Release notes
- Project documentation

Thank you for contributing to Secret Santa!