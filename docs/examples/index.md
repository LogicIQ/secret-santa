# Examples

This section contains practical examples of using Secret Santa for common use cases.

## Quick Reference

| Use Case | Example | Description |
|----------|---------|-------------|
| Basic Password | [basic-password.md](basic-password.md) | Simple password generation |
| TLS Certificate | [tls-self-signed.md](tls-self-signed.md) | Self-signed certificates |
| AWS Integration | [aws-secrets-manager.md](aws-secrets-manager.md) | Store in AWS Secrets Manager |

## Basic Examples

- [Simple Password Generation](basic-password.md) - Generate a basic password secret
- [Database Credentials](database-credentials.md) - Complete database credential setup
- [API Keys](api-keys.md) - Generate API keys and tokens

## TLS and Certificates

- [Self-Signed TLS Certificate](tls-self-signed.md) - Generate TLS certificates for development
- [Certificate Bundle](tls-bundle.md) - Complete certificate chain with private key
- [Multiple SANs](tls-multiple-sans.md) - Certificates with multiple Subject Alternative Names

## Cloud Provider Integration

- [AWS Secrets Manager](aws-secrets-manager.md) - Store secrets in AWS Secrets Manager
- [AWS Parameter Store](aws-parameter-store.md) - Store connection strings in Parameter Store
- [GCP Secret Manager](gcp-secret-manager.md) - Store secrets in Google Cloud Secret Manager

## Advanced Use Cases

- [Template Functions](template-functions.md) - Advanced template usage and functions
- [Dry-Run Validation](dry-run.md) - Validate configurations without creating secrets
- [Multi-Environment Setup](multi-environment.md) - Manage secrets across environments

## Integration Examples

- [Docker Registry Credentials](docker-registry.md) - Generate Docker registry authentication
- [SSH Keys](ssh-keys.md) - Generate SSH key pairs for automation
- [JWT Signing Keys](jwt-keys.md) - Generate keys for JWT token signing

Each example includes:
- Complete YAML manifests
- Explanation of configuration options
- Expected output format
- Common troubleshooting tips