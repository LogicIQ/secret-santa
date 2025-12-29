# Examples Overview

This section contains practical examples for common Secret Santa use cases.

## Quick Reference

| Use Case | Example | Description |
|----------|---------|-------------|
| Basic Password | [basic-password.md](basic-password.md) | Simple password generation |
| TLS Certificate | [tls-self-signed.md](tls-self-signed.md) | Self-signed certificates |
| AWS Integration | [aws-secrets-manager.md](aws-secrets-manager.md) | Store in AWS Secrets Manager |

## Getting Started Examples

### [Basic Password Generation](basic-password.md)
Generate simple passwords for applications:
```yaml
spec:
  template: |
    password: {{ .pass.password }}
  generators:
    - name: pass
      type: random_password
      config:
        length: 32
```

### [TLS Certificates](tls-self-signed.md)
Create self-signed certificates for development:
```yaml
spec:
  secretType: kubernetes.io/tls
  template: |
    tls.crt: {{ .cert.certificate }}
    tls.key: {{ .key.private_key_pem }}
```

## Cloud Provider Examples

### [AWS Secrets Manager](aws-secrets-manager.md)
Store secrets in AWS with encryption:
```yaml
spec:
  media:
    type: aws-secrets-manager
    config:
      region: us-west-2
      kms_key_id: alias/secrets-key
```

## Common Patterns

### Database Credentials
```yaml
spec:
  template: |
    username: admin
    password: {{ .pass.password }}
    host: db.example.com
  generators:
    - name: pass
      type: random_password
      config:
        length: 24
        includeSymbols: false
```

### API Keys
```yaml
spec:
  template: |
    api_key: {{ .key.value }}
    client_id: {{ .id.uuid }}
  generators:
    - name: key
      type: random_string
      config:
        length: 64
        charset: base64
    - name: id
      type: random_uuid
```

### JWT Signing Keys
```yaml
spec:
  template: |
    private_key: {{ .key.private_key_pem }}
    public_key: {{ .key.public_key_pem }}
  generators:
    - name: key
      type: crypto_rsa_key
      config:
        keySize: 2048
```