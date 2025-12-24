# Secret Santa

Kubernetes operator for generating secrets with templates and storing them in multiple destinations.

## Features

- **Multiple Storage**: Kubernetes secrets, AWS Secrets Manager, AWS Parameter Store, GCP Secret Manager
- **Template Engine**: Go templates with crypto, random, and TLS generators
- **Create-Once**: Secrets generated once and never modified
- **Cloud Integration**: AWS and GCP authentication support

## Installation

### Helm (Recommended)

```bash
helm repo add logiciq https://charts.logiciq.ca
helm install secret-santa logiciq/secret-santa
```

### AWS Setup (Optional)

For AWS Secrets Manager or Parameter Store:

```bash
# With service account annotations (EKS)
helm install secret-santa logiciq/secret-santa \
  --set aws.enabled=true \
  --set aws.region=us-west-2 \
  --set serviceAccount.annotations."eks\.amazonaws\.com/role-arn"=arn:aws:iam::123456789012:role/secret-santa

# With environment variables
helm install secret-santa logiciq/secret-santa \
  --set aws.enabled=true \
  --set aws.region=us-west-2 \
  --set aws.credentials.useServiceAccount=false \
  --set aws.credentials.accessKeyId=AKIA... \
  --set aws.credentials.secretAccessKey=...
```

### GCP Setup (Optional)

For GCP Secret Manager:

```bash
# With service account key file
helm install secret-santa logiciq/secret-santa \
  --set gcp.enabled=true \
  --set gcp.projectId=my-project-id \
  --set gcp.credentials.useWorkloadIdentity=false \
  --set gcp.credentials.keyFile=/etc/gcp/key.json \
  --set gcp.credentials.existingSecret=gcp-service-account-key

# With workload identity (GKE)
helm install secret-santa logiciq/secret-santa \
  --set gcp.enabled=true \
  --set gcp.projectId=my-project-id \
  --set gcp.credentials.useWorkloadIdentity=true \
  --set serviceAccount.annotations."iam\.gke\.io/gcp-service-account"=secret-santa@my-project.iam.gserviceaccount.com
```

## Quick Start

### Basic Password Generation (Kubernetes Secret)

```yaml
apiVersion: secrets.secret-santa.io/v1alpha1
kind: SecretSanta
metadata:
  name: app-password
spec:
  template: |
    password: {{ .pass.password }}
  generators:
    - name: pass
      type: random_password
      config:
        length: 32
  media:
    type: k8s  # Default - can be omitted
```

### TLS Certificate (Kubernetes Secret)

```yaml
apiVersion: secrets.secret-santa.io/v1alpha1
kind: SecretSanta
metadata:
  name: tls-cert
spec:
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
          common_name: example.com
  secretType: kubernetes.io/tls
  media:
    type: k8s  # Default - can be omitted
```

### AWS Secrets Manager

```yaml
apiVersion: secrets.secret-santa.io/v1alpha1
kind: SecretSanta
metadata:
  name: aws-secret
spec:
  template: |
    {
      "username": "admin",
      "password": "{{ .pass.password }}"
    }
  generators:
    - name: pass
      type: random_password
      config:
        length: 24
  media:
    type: aws-secrets-manager
    config:
      region: us-west-2
      kms_key_id: alias/secrets-key
```

### AWS Parameter Store

```yaml
apiVersion: secrets.secret-santa.io/v1alpha1
kind: SecretSanta
metadata:
  name: db-url
spec:
  template: |
    postgresql://user:{{ .pass.password }}@db.example.com/app
  generators:
    - name: pass
      type: random_password
      config:
        length: 32
  media:
    type: aws-parameter-store
    config:
      region: us-east-1
      parameter_name: /app/database-url
```

### GCP Secret Manager

```yaml
apiVersion: secrets.secret-santa.io/v1alpha1
kind: SecretSanta
metadata:
  name: gcp-secret
spec:
  template: |
    {
      "username": "admin",
      "password": "{{ .pass.password }}"
    }
  generators:
    - name: pass
      type: random_password
      config:
        length: 24
  media:
    type: gcp-secret-manager
    config:
      project_id: my-gcp-project
      secret_name: app-credentials
```

## Storage Destinations

### Kubernetes Secrets (Default)

```yaml
# Default media - can be omitted entirely
media:
  type: k8s

# Or with custom secret name
media:
  type: k8s
  config:
    secret_name: my-custom-secret-name
```

### AWS Secrets Manager

```yaml
media:
  type: aws-secrets-manager
  config:
    region: us-west-2                    # Optional
    secret_name: my-custom-secret        # Optional
    kms_key_id: alias/my-kms-key        # Optional
```

### AWS Parameter Store

```yaml
media:
  type: aws-parameter-store
  config:
    region: us-east-1                    # Optional
    parameter_name: /my/custom/param     # Optional
    kms_key_id: alias/my-kms-key        # Optional
```

### GCP Secret Manager

```yaml
media:
  type: gcp-secret-manager
  config:
    project_id: my-gcp-project           # Required
    secret_name: my-custom-secret        # Optional
    credentials_file: /path/to/key.json  # Optional - uses workload identity if empty
```

## Generators

### Random
- `random_password` - Secure passwords
- `random_string` - Random strings
- `random_uuid` - UUIDs
- `random_bytes` - Byte arrays

### TLS
- `tls_private_key` - Private keys
- `tls_self_signed_cert` - Self-signed certificates
- `tls_cert_request` - Certificate requests

### Crypto
- `crypto_aes_key` - AES keys
- `crypto_rsa_key` - RSA keys
- `crypto_ed25519_key` - Ed25519 keys

## Configuration

### Helm Values

```yaml
controller:
  replicas: 1
  args:
    maxConcurrentReconciles: 5
    watchNamespaces: ["default", "production"]
    logLevel: "info"

aws:
  enabled: true
  region: us-west-2
  credentials:
    useServiceAccount: true

gcp:
  enabled: true
  projectId: my-gcp-project
  credentials:
    useWorkloadIdentity: true
    # OR for service account key:
    # useWorkloadIdentity: false
    # existingSecret: gcp-service-account-key
    # existingSecretKey: key.json

serviceAccount:
  annotations:
    eks.amazonaws.com/role-arn: arn:aws:iam::123456789012:role/secret-santa
    iam.gke.io/gcp-service-account: secret-santa@my-project.iam.gserviceaccount.com
```

### Environment Variables

```bash
SECRET_SANTA_MAX_CONCURRENT_RECONCILES=5
SECRET_SANTA_WATCH_NAMESPACES=default,production
SECRET_SANTA_LOG_LEVEL=debug
AWS_REGION=us-west-2
GOOGLE_APPLICATION_CREDENTIALS=/path/to/service-account.json
GCP_PROJECT_ID=my-gcp-project
```