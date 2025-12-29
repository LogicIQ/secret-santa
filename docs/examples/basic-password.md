# Basic Password Generation

This example demonstrates the simplest use case: generating a random password and storing it as a Kubernetes secret.

## Example

```yaml
apiVersion: secrets.secret-santa.io/v1alpha1
kind: SecretSanta
metadata:
  name: basic-password
  namespace: default
spec:
  template: |
    password: {{ .pass.password }}
  generators:
    - name: pass
      type: random_password
      config:
        length: 32
        includeSymbols: true
```

## What This Creates

This will create a Kubernetes secret named `basic-password` with the following structure:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: basic-password
  namespace: default
  annotations:
    secrets.secret-santa.io/created-at: "2024-01-15T10:30:00Z"
    secrets.secret-santa.io/generator-types: "random_password"
    secrets.secret-santa.io/template-checksum: "sha256:abc123..."
    secrets.secret-santa.io/source-cr: "default/basic-password"
type: Opaque
data:
  password: <base64-encoded-password>
```

## Verification

Check that the secret was created:

```bash
kubectl get secret basic-password
```

View the generated password (decode from base64):

```bash
kubectl get secret basic-password -o jsonpath='{.data.password}' | base64 -d
```

## Customization Options

### Password Length

```yaml
generators:
  - name: pass
    type: random_password
    config:
      length: 16  # Shorter password
```

### Exclude Symbols

```yaml
generators:
  - name: pass
    type: random_password
    config:
      length: 32
      includeSymbols: false  # Only alphanumeric characters
```

### Exclude Similar Characters

```yaml
generators:
  - name: pass
    type: random_password
    config:
      length: 32
      excludeSimilar: true  # Exclude 0, O, 1, l, I
```

### Minimum Character Requirements

```yaml
generators:
  - name: pass
    type: random_password
    config:
      length: 32
      minNumeric: 4   # At least 4 numbers
      minSymbols: 2   # At least 2 symbols
```

## Custom Secret Name

Store the password in a secret with a different name:

```yaml
apiVersion: secrets.secret-santa.io/v1alpha1
kind: SecretSanta
metadata:
  name: password-generator
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
    config:
      secret_name: "my-app-password"
```

## Multiple Passwords

Generate multiple passwords in a single secret:

```yaml
apiVersion: secrets.secret-santa.io/v1alpha1
kind: SecretSanta
metadata:
  name: multi-password
spec:
  template: |
    admin_password: {{ .admin.password }}
    user_password: {{ .user.password }}
    api_key: {{ .api.password }}
  generators:
    - name: admin
      type: random_password
      config:
        length: 32
        includeSymbols: true
    - name: user
      type: random_password
      config:
        length: 24
        includeSymbols: false
    - name: api
      type: random_password
      config:
        length: 64
        includeSymbols: true
```

## Troubleshooting

### Secret Not Created

Check the SecretSanta resource status:

```bash
kubectl get secretsanta basic-password -o yaml
```

Look for conditions in the status section:

```yaml
status:
  conditions:
    - type: Ready
      status: "False"
      reason: GeneratorError
      message: "Failed to generate password: invalid length"
```

### Invalid Configuration

Use dry-run mode to validate configuration:

```yaml
apiVersion: secrets.secret-santa.io/v1alpha1
kind: SecretSanta
metadata:
  name: test-config
spec:
  dryRun: true  # Add this line
  template: |
    password: {{ .pass.password }}
  generators:
    - name: pass
      type: random_password
      config:
        length: 32
```

Check the dry-run results:

```bash
kubectl get secretsanta test-config -o jsonpath='{.status.dryRunResult.maskedOutput}'
```

### Permission Issues

Ensure the Secret Santa operator has proper RBAC permissions:

```bash
kubectl get clusterrole secret-santa-manager-role -o yaml
```

The operator needs permissions to create and update secrets in the target namespace.