# E2E Tests

This directory contains end-to-end tests for the Secret Santa Kubernetes operator.

## Structure

```
e2e/
└── k8s/
    ├── crypto/          # Crypto generator tests
    ├── random/          # Random generator tests  
    ├── tls/             # TLS generator tests
    ├── time/            # Time generator tests
    ├── comprehensive_test.go  # Cross-generator tests
    ├── e2e_test.go      # Basic e2e tests
    └── Taskfile.yml     # Task runner configuration
```

## Running Tests

The e2e tests are organized to mirror the `pkg/generators` structure and use a simplified task runner setup.

### Prerequisites

- Docker
- kubectl
- Go 1.25+

### Commands

There are only 2 main commands you need:

#### Run E2E Tests
```bash
cd e2e/k8s
task e2e
```

This command will:
- Install required tools (kind, controller-gen) if needed
- Create or reuse existing Kind cluster
- Build and load the operator image
- Deploy CRDs, RBAC, and operator
- Run all e2e tests

#### Cleanup
```bash
cd e2e/k8s  
task cleanup
```

This removes the Kind cluster and all resources.

### Test Categories

- **crypto/**: Tests all cryptographic generators (AES, HMAC, RSA, Ed25519, ChaCha20, XChaCha20, ECDSA, ECDH)
- **random/**: Tests random data generators (password, string, UUID, integer, bytes, ID)
- **tls/**: Tests TLS generators (private key, self-signed cert, cert request, locally signed cert)
- **time/**: Tests time-based generators (static timestamp)
- **comprehensive_test.go**: Tests multiple generators together and template functions

### Cluster Reuse

The setup is designed to reuse existing clusters and resources efficiently:
- If a Kind cluster already exists, it will be reused
- If resources are already deployed, they will be updated or skipped
- Failed deployments won't stop the test execution

This makes iterative testing much faster during development.