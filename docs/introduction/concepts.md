# Core Concepts

Secret Santa is built around a few key concepts that work together to provide secure, automated secret generation and distribution.

## Architecture Overview

Secret Santa follows a declarative approach where you define what secrets you want, and the operator ensures they exist and are properly distributed.

![Architecture](../images/secret-santa.webp)

## Key Components

### SecretSanta Resource

The `SecretSanta` custom resource is the primary interface for defining secret generation requirements:

```yaml
apiVersion: secrets.secret-santa.io/v1alpha1
kind: SecretSanta
metadata:
  name: my-secret
spec:
  template: |
    password: {{ .pass.password }}
  generators:
    - name: pass
      type: random_password
      config:
        length: 32
  media:
    type: k8s
```

### Generators

Generators create the raw cryptographic material and secret data:

- **Random Generators**: Passwords, strings, UUIDs, bytes
- **Cryptographic Generators**: AES keys, RSA keys, Ed25519 keys
- **TLS Generators**: Private keys, certificates, CSRs

### Template Engine

Go templates process generator outputs into the final secret format:

```go
// Template has access to all generator outputs
password: {{ .pass.password }}
api_key: {{ .key.value }}
certificate: {{ .cert.certificate }}
```

### Media Providers

Media providers handle secret storage across different destinations:

- **Kubernetes**: Native secret storage (default)
- **AWS**: Secrets Manager and Parameter Store
- **GCP**: Secret Manager

## Design Principles

### Create-Once Semantics

Secrets are generated once and never modified. This ensures:
- Consistent secret values across restarts
- No accidental overwrites
- Predictable behavior for dependent applications

### Declarative Configuration

Define the desired state, and Secret Santa maintains it:
- Automatic reconciliation
- Self-healing on failures
- Consistent secret availability

### Multi-Destination Support

Store the same secret across multiple systems:
- Kubernetes secrets for pod consumption
- Cloud provider secrets for external access
- Consistent values across all destinations

### Security by Design

- Minimal required permissions
- Encrypted storage where supported
- Masked values in logs and status
- Automatic metadata for traceability

## Workflow

1. **Resource Creation**: User creates SecretSanta resource
2. **Generator Execution**: Operator runs configured generators
3. **Template Processing**: Template engine combines generator outputs
4. **Media Storage**: Processed secret stored in configured destinations
5. **Status Updates**: Resource status reflects success/failure
6. **Continuous Reconciliation**: Operator ensures secrets remain available

## Use Cases

### Application Secrets

Generate database passwords, API keys, and other application credentials:

```yaml
spec:
  template: |
    {
      "db_password": "{{ .dbpass.password }}",
      "api_key": "{{ .apikey.value }}"
    }
```

### TLS Certificates

Create self-signed certificates for development and testing:

```yaml
spec:
  template: |
    tls.crt: {{ .cert.certificate }}
    tls.key: {{ .key.private_key_pem }}
```

### Cloud Integration

Store secrets in cloud provider secret management services:

```yaml
spec:
  media:
    type: aws-secrets-manager
    config:
      region: us-west-2
```

## Security Model

### Operator Permissions

The Secret Santa operator requires:
- Kubernetes: Create/update secrets in target namespaces
- AWS: Secrets Manager and Parameter Store access
- GCP: Secret Manager access

### Secret Lifecycle

- **Generation**: Cryptographically secure random generation
- **Storage**: Encrypted at rest in cloud providers
- **Access**: Standard Kubernetes RBAC for K8s secrets
- **Rotation**: Manual deletion and recreation (future: automatic rotation)

### Audit and Compliance

- Automatic metadata tagging
- CloudTrail/Audit log integration
- Masked sensitive values in operator logs
- Traceability from secret to source resource