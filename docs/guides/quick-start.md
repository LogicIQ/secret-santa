# Quick Start

Get up and running with Secret Santa in 5 minutes.

## Prerequisites

- Kubernetes cluster
- Helm 3.x
- kubectl configured

## 1. Install Secret Santa

```bash
# Add repository
helm repo add logiciq https://charts.logiciq.ca
helm repo update

# Install
helm install secret-santa logiciq/secret-santa

# Verify
kubectl get pods -l app.kubernetes.io/name=secret-santa
```

## 2. Create Your First Secret

```bash
cat << EOF | kubectl apply -f -
apiVersion: secrets.secret-santa.io/v1alpha1
kind: SecretSanta
metadata:
  name: my-first-secret
spec:
  template: |
    password: {{ .pass.password }}
  generators:
    - name: pass
      type: random_password
      config:
        length: 32
EOF
```

## 3. Verify Secret Creation

```bash
# Check SecretSanta resource
kubectl get secretsanta my-first-secret

# Check generated Kubernetes secret
kubectl get secret my-first-secret

# View the password (base64 decoded)
kubectl get secret my-first-secret -o jsonpath='{.data.password}' | base64 -d
```

## 4. Try Advanced Examples

### TLS Certificate

```bash
cat << EOF | kubectl apply -f -
apiVersion: secrets.secret-santa.io/v1alpha1
kind: SecretSanta
metadata:
  name: tls-cert
spec:
  secretType: kubernetes.io/tls
  template: |
    tls.crt: {{ .cert.certificate }}
    tls.key: {{ .key.private_key_pem }}
  generators:
    - name: key
      type: tls_private_key
    - name: cert
      type: tls_self_signed_cert
      config:
        key_pem: "{{ .key.private_key_pem }}"
        subject:
          common_name: "example.com"
EOF
```

### Database Credentials

```bash
cat << EOF | kubectl apply -f -
apiVersion: secrets.secret-santa.io/v1alpha1
kind: SecretSanta
metadata:
  name: db-creds
spec:
  template: |
    username: admin
    password: {{ .pass.password }}
    host: postgres.example.com
    port: "5432"
  generators:
    - name: pass
      type: random_password
      config:
        length: 24
        includeSymbols: false
EOF
```

## 5. Cloud Provider Setup (Optional)

### AWS Secrets Manager

```bash
# Configure AWS (requires IAM permissions)
helm upgrade secret-santa logiciq/secret-santa \
  --set aws.enabled=true \
  --set aws.region=us-west-2

# Create AWS secret
cat << EOF | kubectl apply -f -
apiVersion: secrets.secret-santa.io/v1alpha1
kind: SecretSanta
metadata:
  name: aws-secret
spec:
  template: |
    {"password": "{{ .pass.password }}"}
  generators:
    - name: pass
      type: random_password
      config:
        length: 32
  media:
    type: aws-secrets-manager
    config:
      secret_name: "myapp/password"
EOF
```

## Next Steps

- Read [Core Concepts](../introduction/concepts.md) to understand the architecture
- Explore [Examples](../examples/) for real-world use cases
- Check [API Reference](../api/spec.md) for complete configuration options
- Set up [Cloud Provider Authentication](installation.md#aws-configuration) for production use

## Troubleshooting

### Secret Not Created

```bash
# Check SecretSanta status
kubectl get secretsanta my-first-secret -o yaml

# Check operator logs
kubectl logs -l app.kubernetes.io/name=secret-santa
```

### Permission Issues

```bash
# Check RBAC
kubectl get clusterrole secret-santa-manager-role
kubectl get clusterrolebinding secret-santa-manager-rolebinding
```

### Dry-Run Testing

```bash
cat << EOF | kubectl apply -f -
apiVersion: secrets.secret-santa.io/v1alpha1
kind: SecretSanta
metadata:
  name: test-config
spec:
  dryRun: true
  template: |
    password: {{ .pass.password }}
  generators:
    - name: pass
      type: random_password
      config:
        length: 32
EOF

# Check dry-run results
kubectl get secretsanta test-config -o jsonpath='{.status.dryRunResult.maskedOutput}'
```