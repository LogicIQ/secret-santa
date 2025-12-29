# Media Providers

Media providers define where Secret Santa stores generated secrets. Each provider has specific configuration options and authentication requirements.

## Kubernetes Secrets (Default)

The default media provider stores secrets as native Kubernetes secrets in the same namespace as the SecretSanta resource.

### Basic Configuration

```yaml
# Default - can be omitted entirely
media:
  type: k8s
```

### Advanced Configuration

```yaml
media:
  type: k8s
  config:
    secret_name: "my-custom-secret"    # Custom secret name (default: SecretSanta name)
    namespace: "target-namespace"      # Target namespace (default: same as SecretSanta)
```

### Secret Types

Control the Kubernetes secret type using `spec.secretType`:

```yaml
apiVersion: secrets.secret-santa.io/v1alpha1
kind: SecretSanta
metadata:
  name: tls-example
spec:
  secretType: kubernetes.io/tls  # TLS secret type
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
  media:
    type: k8s
```

**Supported Secret Types**:
- `Opaque` (default)
- `kubernetes.io/tls`
- `kubernetes.io/dockerconfigjson`
- `kubernetes.io/basic-auth`
- `kubernetes.io/ssh-auth`
- `kubernetes.io/service-account-token`

### Metadata and Annotations

Secret Santa automatically adds metadata annotations to Kubernetes secrets:

```yaml
metadata:
  annotations:
    secrets.secret-santa.io/created-at: "2024-01-15T10:30:00Z"
    secrets.secret-santa.io/generator-types: "random_password,random_string"
    secrets.secret-santa.io/template-checksum: "sha256:abc123..."
    secrets.secret-santa.io/source-cr: "default/my-secret"
```

## AWS Secrets Manager

Store secrets in AWS Secrets Manager with automatic encryption and versioning.

### Configuration

```yaml
media:
  type: aws-secrets-manager
  config:
    region: "us-west-2"                    # AWS region (optional, uses default)
    secret_name: "my-app/database"         # Secret name (optional, uses SecretSanta name)
    kms_key_id: "alias/secrets-key"        # KMS key for encryption (optional)
    description: "Database credentials"     # Secret description (optional)
```

### Authentication

#### IAM Roles for Service Accounts (IRSA) - Recommended

```bash
helm upgrade secret-santa logiciq/secret-santa \
  --set aws.enabled=true \
  --set aws.region=us-west-2 \
  --set serviceAccount.annotations."eks\.amazonaws\.com/role-arn"=arn:aws:iam::123456789012:role/secret-santa
```

**Required IAM Policy**:
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "secretsmanager:CreateSecret",
        "secretsmanager:GetSecretValue",
        "secretsmanager:UpdateSecret",
        "secretsmanager:TagResource"
      ],
      "Resource": "arn:aws:secretsmanager:*:*:secret:*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "kms:Decrypt",
        "kms:GenerateDataKey"
      ],
      "Resource": "arn:aws:kms:*:*:key/*"
    }
  ]
}
```

#### Access Keys

```bash
kubectl create secret generic aws-credentials \
  --from-literal=access-key-id=AKIA... \
  --from-literal=secret-access-key=...

helm upgrade secret-santa logiciq/secret-santa \
  --set aws.enabled=true \
  --set aws.credentials.useServiceAccount=false \
  --set aws.credentials.existingSecret=aws-credentials
```

### Example

```yaml
apiVersion: secrets.secret-santa.io/v1alpha1
kind: SecretSanta
metadata:
  name: database-credentials
spec:
  template: |
    {
      "username": "admin",
      "password": "{{ .pass.password }}",
      "host": "db.example.com",
      "port": 5432
    }
  generators:
    - name: pass
      type: random_password
      config:
        length: 32
  media:
    type: aws-secrets-manager
    config:
      region: us-west-2
      secret_name: "myapp/database"
      kms_key_id: "alias/myapp-secrets"
```

## AWS Parameter Store

Store secrets as encrypted parameters in AWS Systems Manager Parameter Store.

### Configuration

```yaml
media:
  type: aws-parameter-store
  config:
    region: "us-east-1"                    # AWS region (optional)
    parameter_name: "/myapp/database-url"  # Parameter name (optional)
    kms_key_id: "alias/parameter-key"      # KMS key for encryption (optional)
    description: "Database connection URL" # Parameter description (optional)
    tier: "Standard"                       # Parameter tier: Standard, Advanced (default: Standard)
```

### Authentication

Same as AWS Secrets Manager - use IRSA or access keys.

**Required IAM Policy**:
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ssm:PutParameter",
        "ssm:GetParameter",
        "ssm:AddTagsToResource"
      ],
      "Resource": "arn:aws:ssm:*:*:parameter/*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "kms:Decrypt",
        "kms:GenerateDataKey"
      ],
      "Resource": "arn:aws:kms:*:*:key/*"
    }
  ]
}
```

### Example

```yaml
apiVersion: secrets.secret-santa.io/v1alpha1
kind: SecretSanta
metadata:
  name: connection-string
spec:
  template: |
    postgresql://user:{{ .pass.password }}@db.example.com:5432/myapp
  generators:
    - name: pass
      type: random_password
      config:
        length: 24
  media:
    type: aws-parameter-store
    config:
      region: us-east-1
      parameter_name: "/myapp/database-url"
      kms_key_id: "alias/myapp-parameters"
```

## GCP Secret Manager

Store secrets in Google Cloud Secret Manager with automatic versioning.

### Configuration

```yaml
media:
  type: gcp-secret-manager
  config:
    project_id: "my-gcp-project"           # GCP project ID (required)
    secret_name: "database-credentials"    # Secret name (optional)
    labels:                                # Secret labels (optional)
      environment: "production"
      application: "myapp"
```

### Authentication

#### Workload Identity - Recommended

```bash
helm upgrade secret-santa logiciq/secret-santa \
  --set gcp.enabled=true \
  --set gcp.projectId=my-project-id \
  --set gcp.credentials.useWorkloadIdentity=true \
  --set serviceAccount.annotations."iam\.gke\.io/gcp-service-account"=secret-santa@my-project.iam.gserviceaccount.com
```

**Required IAM Roles**:
- `roles/secretmanager.admin` or custom role with:
  - `secretmanager.secrets.create`
  - `secretmanager.secrets.get`
  - `secretmanager.versions.add`
  - `secretmanager.versions.access`

#### Service Account Key

```bash
kubectl create secret generic gcp-credentials \
  --from-file=key.json=/path/to/service-account.json

helm upgrade secret-santa logiciq/secret-santa \
  --set gcp.enabled=true \
  --set gcp.projectId=my-project-id \
  --set gcp.credentials.useWorkloadIdentity=false \
  --set gcp.credentials.existingSecret=gcp-credentials
```

### Example

```yaml
apiVersion: secrets.secret-santa.io/v1alpha1
kind: SecretSanta
metadata:
  name: api-credentials
spec:
  template: |
    {
      "client_id": "{{ .clientid.uuid }}",
      "client_secret": "{{ .secret.value }}",
      "api_key": "{{ .apikey.value }}"
    }
  generators:
    - name: clientid
      type: random_uuid
    - name: secret
      type: random_string
      config:
        length: 64
        charset: "base64"
    - name: apikey
      type: random_string
      config:
        length: 32
        charset: "alphanumeric"
  media:
    type: gcp-secret-manager
    config:
      project_id: "my-gcp-project"
      secret_name: "myapp-api-credentials"
      labels:
        environment: "production"
        team: "backend"
```

## Multi-Destination Storage

Store the same secret in multiple destinations by creating multiple SecretSanta resources with the same generators:

```yaml
# Store in Kubernetes
apiVersion: secrets.secret-santa.io/v1alpha1
kind: SecretSanta
metadata:
  name: app-secret-k8s
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
---
# Store in AWS Secrets Manager
apiVersion: secrets.secret-santa.io/v1alpha1
kind: SecretSanta
metadata:
  name: app-secret-aws
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
```

**Note**: Each SecretSanta resource generates independently. For true multi-destination storage, consider using external tools or custom controllers.

## Best Practices

### Security
- Use IAM roles instead of access keys when possible
- Apply least-privilege permissions
- Enable encryption at rest for all cloud providers
- Regularly rotate cloud provider credentials

### Naming
- Use consistent naming conventions across providers
- Include environment and application identifiers
- Avoid special characters that might cause issues

### Monitoring
- Enable CloudTrail/Cloud Audit logs for secret access
- Monitor secret creation and access patterns
- Set up alerts for unauthorized access attempts

### Cost Optimization
- Use appropriate storage tiers (Standard vs Advanced for Parameter Store)
- Clean up unused secrets regularly
- Consider secret versioning costs for high-frequency updates