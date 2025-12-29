# AWS Secrets Manager Integration

This example demonstrates storing generated secrets in AWS Secrets Manager with proper authentication and encryption.

## Prerequisites

- AWS account with Secrets Manager access
- Secret Santa operator configured with AWS credentials
- Appropriate IAM permissions

## Basic Example

```yaml
apiVersion: secrets.secret-santa.io/v1alpha1
kind: SecretSanta
metadata:
  name: database-secret
  namespace: production
spec:
  template: |
    {
      "username": "admin",
      "password": "{{ .pass.password }}",
      "engine": "postgres",
      "host": "db.example.com",
      "port": 5432,
      "dbname": "myapp"
    }
  generators:
    - name: pass
      type: random_password
      config:
        length: 32
        includeSymbols: true
        excludeSimilar: true
  media:
    type: aws-secrets-manager
    config:
      region: "us-west-2"
      secret_name: "myapp/database/production"
      kms_key_id: "alias/myapp-secrets"
      description: "Database credentials for MyApp production"
```

## What This Creates

This creates a secret in AWS Secrets Manager with:

- **Name**: `myapp/database/production`
- **Region**: `us-west-2`
- **Encryption**: Using KMS key `alias/myapp-secrets`
- **Value**: JSON object with database connection details
- **Tags**: Automatic metadata tags for traceability

## Authentication Setup

### Option 1: IAM Roles for Service Accounts (IRSA) - Recommended

Create an IAM role with the following policy:

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
        "secretsmanager:TagResource",
        "secretsmanager:DescribeSecret"
      ],
      "Resource": [
        "arn:aws:secretsmanager:us-west-2:123456789012:secret:myapp/*"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "kms:Decrypt",
        "kms:GenerateDataKey",
        "kms:DescribeKey"
      ],
      "Resource": [
        "arn:aws:kms:us-west-2:123456789012:key/12345678-1234-1234-1234-123456789012"
      ]
    }
  ]
}
```

Configure the Helm installation:

```bash
helm upgrade secret-santa logiciq/secret-santa \
  --set aws.enabled=true \
  --set aws.region=us-west-2 \
  --set serviceAccount.annotations."eks\.amazonaws\.com/role-arn"=arn:aws:iam::123456789012:role/secret-santa-role
```

### Option 2: Access Keys

Create a Kubernetes secret with AWS credentials:

```bash
kubectl create secret generic aws-credentials \
  --from-literal=access-key-id=AKIA... \
  --from-literal=secret-access-key=...
```

Configure the Helm installation:

```bash
helm upgrade secret-santa logiciq/secret-santa \
  --set aws.enabled=true \
  --set aws.region=us-west-2 \
  --set aws.credentials.useServiceAccount=false \
  --set aws.credentials.existingSecret=aws-credentials
```

## Advanced Examples

### Multi-Region Deployment

Store the same secret in multiple regions:

```yaml
# Primary region
apiVersion: secrets.secret-santa.io/v1alpha1
kind: SecretSanta
metadata:
  name: api-key-us-west
spec:
  template: |
    {
      "api_key": "{{ .key.value }}",
      "created_at": "{{ now | date \"2006-01-02T15:04:05Z07:00\" }}"
    }
  generators:
    - name: key
      type: random_string
      config:
        length: 64
        charset: "base64"
  media:
    type: aws-secrets-manager
    config:
      region: "us-west-2"
      secret_name: "myapp/api-key"
---
# Secondary region
apiVersion: secrets.secret-santa.io/v1alpha1
kind: SecretSanta
metadata:
  name: api-key-us-east
spec:
  template: |
    {
      "api_key": "{{ .key.value }}",
      "created_at": "{{ now | date \"2006-01-02T15:04:05Z07:00\" }}"
    }
  generators:
    - name: key
      type: random_string
      config:
        length: 64
        charset: "base64"
  media:
    type: aws-secrets-manager
    config:
      region: "us-east-1"
      secret_name: "myapp/api-key"
```

### Complex Application Configuration

```yaml
apiVersion: secrets.secret-santa.io/v1alpha1
kind: SecretSanta
metadata:
  name: app-config
spec:
  template: |
    {
      "database": {
        "username": "app_user",
        "password": "{{ .dbpass.password }}",
        "host": "{{ .Values.database.host }}",
        "port": 5432
      },
      "redis": {
        "password": "{{ .redispass.password }}",
        "host": "{{ .Values.redis.host }}",
        "port": 6379
      },
      "jwt": {
        "secret": "{{ .jwtsecret.value }}",
        "algorithm": "HS256"
      },
      "encryption": {
        "key": "{{ .enckey.key }}"
      }
    }
  generators:
    - name: dbpass
      type: random_password
      config:
        length: 32
        includeSymbols: false
    - name: redispass
      type: random_password
      config:
        length: 24
        includeSymbols: false
    - name: jwtsecret
      type: random_string
      config:
        length: 64
        charset: "base64"
    - name: enckey
      type: crypto_aes_key
      config:
        keySize: 256
        encoding: "base64"
  media:
    type: aws-secrets-manager
    config:
      region: "us-west-2"
      secret_name: "myapp/config/production"
      kms_key_id: "alias/myapp-secrets"
```

### Certificate Storage

Store TLS certificates in Secrets Manager:

```yaml
apiVersion: secrets.secret-santa.io/v1alpha1
kind: SecretSanta
metadata:
  name: tls-cert-aws
spec:
  template: |
    {
      "certificate": "{{ .cert.certificate }}",
      "private_key": "{{ .key.private_key_pem }}",
      "common_name": "{{ .Values.hostname }}",
      "expires_at": "{{ .cert.not_after }}"
    }
  generators:
    - name: key
      type: tls_private_key
      config:
        algorithm: "RSA"
        keySize: 2048
    - name: cert
      type: tls_self_signed_cert
      config:
        key_pem: "{{ .key.private_key_pem }}"
        subject:
          common_name: "api.example.com"
        dns_names:
          - "api.example.com"
          - "*.api.example.com"
        validity_days: 365
  media:
    type: aws-secrets-manager
    config:
      region: "us-west-2"
      secret_name: "myapp/tls/api-certificate"
      kms_key_id: "alias/myapp-tls"
```

## Accessing Secrets from Applications

### AWS SDK

```python
import boto3
import json

def get_secret():
    client = boto3.client('secretsmanager', region_name='us-west-2')
    
    try:
        response = client.get_secret_value(SecretId='myapp/database/production')
        secret = json.loads(response['SecretString'])
        return secret
    except Exception as e:
        print(f"Error retrieving secret: {e}")
        return None

# Usage
db_config = get_secret()
if db_config:
    print(f"Connecting to {db_config['host']}:{db_config['port']}")
```

### AWS CLI

```bash
# Get the secret value
aws secretsmanager get-secret-value \
  --secret-id myapp/database/production \
  --region us-west-2 \
  --query SecretString \
  --output text | jq .

# List all secrets with a prefix
aws secretsmanager list-secrets \
  --filters Key=name,Values=myapp/ \
  --region us-west-2
```

### External Secrets Operator Integration

Use with External Secrets Operator to sync back to Kubernetes:

```yaml
apiVersion: external-secrets.io/v1beta1
kind: SecretStore
metadata:
  name: aws-secretsmanager
spec:
  provider:
    aws:
      service: SecretsManager
      region: us-west-2
      auth:
        serviceAccount:
          name: external-secrets-sa
---
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: database-secret
spec:
  refreshInterval: 1h
  secretStoreRef:
    name: aws-secretsmanager
    kind: SecretStore
  target:
    name: database-secret
    creationPolicy: Owner
  data:
    - secretKey: username
      remoteRef:
        key: myapp/database/production
        property: username
    - secretKey: password
      remoteRef:
        key: myapp/database/production
        property: password
```

## Monitoring and Alerting

### CloudWatch Metrics

Monitor secret access and operations:

```bash
# Get secret retrieval metrics
aws logs filter-log-events \
  --log-group-name /aws/secretsmanager \
  --filter-pattern "{ $.eventName = GetSecretValue }" \
  --start-time $(date -d '1 hour ago' +%s)000
```

### CloudTrail Events

Track secret management operations:

```json
{
  "eventVersion": "1.05",
  "userIdentity": {
    "type": "AssumedRole",
    "principalId": "AIDACKCEVSQ6C2EXAMPLE",
    "arn": "arn:aws:sts::123456789012:assumed-role/secret-santa-role/secret-santa-pod"
  },
  "eventTime": "2024-01-15T10:30:00Z",
  "eventSource": "secretsmanager.amazonaws.com",
  "eventName": "CreateSecret",
  "resources": [
    {
      "ARN": "arn:aws:secretsmanager:us-west-2:123456789012:secret:myapp/database/production",
      "accountId": "123456789012"
    }
  ]
}
```

## Troubleshooting

### Permission Denied

Check IAM permissions and ensure the role has access to:
- Secrets Manager operations
- KMS key operations (if using encryption)
- Correct resource ARNs

### Secret Already Exists

Secret Santa uses create-once semantics. If a secret already exists, it won't be modified. To update:

1. Delete the existing secret in AWS
2. Delete and recreate the SecretSanta resource

### KMS Key Access

Ensure the IAM role has permissions to use the KMS key:

```bash
aws kms describe-key --key-id alias/myapp-secrets
aws kms get-key-policy --key-id alias/myapp-secrets --policy-name default
```

### Region Mismatch

Verify the region configuration matches your AWS setup:

```bash
kubectl get secretsanta database-secret -o yaml | grep region
```

### Cost Optimization

- Use appropriate secret naming conventions
- Clean up unused secrets regularly
- Monitor secret retrieval patterns
- Consider secret rotation frequency