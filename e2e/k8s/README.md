# E2E Testing with Kind

This directory contains end-to-end tests for the Secret Santa operator using Kind (Kubernetes in Docker).

## Prerequisites

- Docker
- Go 1.21+
- Task (taskfile.dev)

## Running E2E Tests

### Quick Start

```bash
# Run full e2e test suite
cd e2e/k8s
task e2e
```

### Step by Step

```bash
cd e2e/k8s

# Setup Kind cluster and deploy operator
task setup

# Run tests
task test

# Cleanup
task cleanup
```

### Individual Tasks

```bash
# Install dependencies
task install-deps

# Create Kind cluster
task create-cluster

# Build and load operator image
task load-image

# Deploy CRDs, RBAC, and operator
task deploy-crd
task deploy-rbac
task deploy-manager

# Wait for operator to be ready
task wait-ready

# Run tests only
task test

# Delete cluster
task cleanup
```

## Test Structure

- `e2e_test.go` - Comprehensive e2e tests with template functions
- `random_test.go` - Random generator tests
- `crypto_test.go` - Crypto generator tests  
- `tls_test.go` - TLS generator tests

## Test Features

- **Template Functions**: Tests verify template functions work correctly
- **Multiple Generators**: Tests combinations of different generator types
- **Secret Validation**: Verifies Kubernetes secrets are created with correct content
- **Cleanup**: Automatic cleanup of test resources

## Configuration

- `kind-config.yaml` - Kind cluster configuration
- `Taskfile.yml` - Task definitions for e2e testing

## Troubleshooting

```bash
# Check cluster status
kubectl cluster-info

# Check operator logs
kubectl logs -n secret-santa-system deployment/secret-santa-controller

# List all resources
kubectl get all -n secret-santa-system

# Check CRDs
kubectl get crd secretsantas.secrets.secret-santa.io
```