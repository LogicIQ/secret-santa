# Installation

This guide covers installing Secret Santa using Helm with various authentication methods for AWS and GCP.

## Prerequisites

- Kubernetes cluster (v1.19+)
- Helm 3.x
- kubectl configured to access your cluster

## Basic Installation

### Add Helm Repository

```bash
helm repo add logiciq https://charts.logiciq.ca
helm repo update
```

### Install with Default Configuration

```bash
helm install secret-santa logiciq/secret-santa
```

This installs Secret Santa with:
- Kubernetes secrets only (no cloud providers)
- Single replica controller
- Default resource limits
- All namespaces watched

### Verify Installation

```bash
# Check operator pod
kubectl get pods -l app.kubernetes.io/name=secret-santa

# Verify CRD installation
kubectl get crd secretsantas.secrets.secret-santa.io

# Check operator logs
kubectl logs -l app.kubernetes.io/name=secret-santa
```

## AWS Configuration

### Method 1: IAM Roles for Service Accounts (IRSA) - Recommended

**Prerequisites:**
- EKS cluster with OIDC provider
- IAM role with Secret Santa permissions

**Create IAM Role:**

```bash
# Create trust policy
cat > trust-policy.json << EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Federated": "arn:aws:iam::ACCOUNT-ID:oidc-provider/oidc.eks.REGION.amazonaws.com/id/OIDC-ID"
      },
      "Action": "sts:AssumeRoleWithWebIdentity",
      "Condition": {
        "StringEquals": {
          "oidc.eks.REGION.amazonaws.com/id/OIDC-ID:sub": "system:serviceaccount:default:secret-santa",
          "oidc.eks.REGION.amazonaws.com/id/OIDC-ID:aud": "sts.amazonaws.com"
        }
      }
    }
  ]
}
EOF

# Create IAM role
aws iam create-role \
  --role-name secret-santa-role \
  --assume-role-policy-document file://trust-policy.json

# Attach permissions policy
aws iam attach-role-policy \
  --role-name secret-santa-role \
  --policy-arn arn:aws:iam::ACCOUNT-ID:policy/SecretSantaPolicy
```

**Install with IRSA:**

```bash
helm install secret-santa logiciq/secret-santa \
  --set aws.enabled=true \
  --set aws.region=us-west-2 \
  --set serviceAccount.annotations."eks\.amazonaws\.com/role-arn"=arn:aws:iam::ACCOUNT-ID:role/secret-santa-role
```

### Method 2: Access Keys

**Create Kubernetes Secret:**

```bash
kubectl create secret generic aws-credentials \
  --from-literal=access-key-id=AKIA... \
  --from-literal=secret-access-key=...
```

**Install with Access Keys:**

```bash
helm install secret-santa logiciq/secret-santa \
  --set aws.enabled=true \
  --set aws.region=us-west-2 \
  --set aws.credentials.useServiceAccount=false \
  --set aws.credentials.existingSecret=aws-credentials
```

### Method 3: Instance Profile (EC2/EKS Nodes)

```bash
helm install secret-santa logiciq/secret-santa \
  --set aws.enabled=true \
  --set aws.region=us-west-2 \
  --set aws.credentials.useServiceAccount=false
```

### Required AWS IAM Policy

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
      "Resource": "arn:aws:secretsmanager:*:*:secret:*"
    },
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
        "kms:GenerateDataKey",
        "kms:DescribeKey"
      ],
      "Resource": "arn:aws:kms:*:*:key/*"
    }
  ]
}
```

## GCP Configuration

### Method 1: Workload Identity - Recommended

**Prerequisites:**
- GKE cluster with Workload Identity enabled
- GCP service account with Secret Manager permissions

**Setup Workload Identity:**

```bash
# Create GCP service account
gcloud iam service-accounts create secret-santa \
  --display-name="Secret Santa Operator"

# Grant Secret Manager permissions
gcloud projects add-iam-policy-binding PROJECT-ID \
  --member="serviceAccount:secret-santa@PROJECT-ID.iam.gserviceaccount.com" \
  --role="roles/secretmanager.admin"

# Enable Workload Identity binding
gcloud iam service-accounts add-iam-policy-binding \
  --role roles/iam.workloadIdentityUser \
  --member "serviceAccount:PROJECT-ID.svc.id.goog[default/secret-santa]" \
  secret-santa@PROJECT-ID.iam.gserviceaccount.com
```

**Install with Workload Identity:**

```bash
helm install secret-santa logiciq/secret-santa \
  --set gcp.enabled=true \
  --set gcp.projectId=PROJECT-ID \
  --set gcp.credentials.useWorkloadIdentity=true \
  --set serviceAccount.annotations."iam\.gke\.io/gcp-service-account"=secret-santa@PROJECT-ID.iam.gserviceaccount.com
```

### Method 2: Service Account Key

**Create Service Account Key:**

```bash
# Create service account
gcloud iam service-accounts create secret-santa

# Grant permissions
gcloud projects add-iam-policy-binding PROJECT-ID \
  --member="serviceAccount:secret-santa@PROJECT-ID.iam.gserviceaccount.com" \
  --role="roles/secretmanager.admin"

# Create and download key
gcloud iam service-accounts keys create key.json \
  --iam-account=secret-santa@PROJECT-ID.iam.gserviceaccount.com
```

**Create Kubernetes Secret:**

```bash
kubectl create secret generic gcp-credentials \
  --from-file=key.json=key.json
```

**Install with Service Account Key:**

```bash
helm install secret-santa logiciq/secret-santa \
  --set gcp.enabled=true \
  --set gcp.projectId=PROJECT-ID \
  --set gcp.credentials.useWorkloadIdentity=false \
  --set gcp.credentials.existingSecret=gcp-credentials
```

### Method 3: Application Default Credentials

For GCE instances with appropriate scopes:

```bash
helm install secret-santa logiciq/secret-santa \
  --set gcp.enabled=true \
  --set gcp.projectId=PROJECT-ID \
  --set gcp.credentials.useWorkloadIdentity=false
```

## Multi-Cloud Configuration

Enable both AWS and GCP:

```bash
helm install secret-santa logiciq/secret-santa \
  --set aws.enabled=true \
  --set aws.region=us-west-2 \
  --set serviceAccount.annotations."eks\.amazonaws\.com/role-arn"=arn:aws:iam::ACCOUNT-ID:role/secret-santa-role \
  --set gcp.enabled=true \
  --set gcp.projectId=PROJECT-ID \
  --set serviceAccount.annotations."iam\.gke\.io/gcp-service-account"=secret-santa@PROJECT-ID.iam.gserviceaccount.com
```

## Configuration Options

### Complete Helm Values

```yaml
# values.yaml
controller:
  replicas: 1
  image:
    repository: logiciq/secret-santa
    tag: "v0.2.0"
    pullPolicy: IfNotPresent
  
  args:
    maxConcurrentReconciles: 5
    watchNamespaces: []  # Empty = all namespaces
    logLevel: "info"
    dryRun: false
    enableMetadata: true
  
  resources:
    limits:
      cpu: 500m
      memory: 128Mi
    requests:
      cpu: 10m
      memory: 64Mi

serviceAccount:
  create: true
  name: ""
  annotations: {}

aws:
  enabled: false
  region: ""
  credentials:
    useServiceAccount: true
    existingSecret: ""
    existingSecretKeys:
      accessKeyId: "access-key-id"
      secretAccessKey: "secret-access-key"

gcp:
  enabled: false
  projectId: ""
  credentials:
    useWorkloadIdentity: true
    existingSecret: ""
    existingSecretKey: "key.json"

rbac:
  create: true

podSecurityContext:
  runAsNonRoot: true
  runAsUser: 65532

securityContext:
  allowPrivilegeEscalation: false
  capabilities:
    drop:
    - ALL
  readOnlyRootFilesystem: true

nodeSelector: {}
tolerations: []
affinity: {}
```

### Install with Custom Values

```bash
helm install secret-santa logiciq/secret-santa -f values.yaml
```

## Namespace-Scoped Installation

Limit operator to specific namespaces:

```bash
helm install secret-santa logiciq/secret-santa \
  --set controller.args.watchNamespaces="{production,staging}"
```

## High Availability

Deploy multiple replicas with leader election:

```bash
helm install secret-santa logiciq/secret-santa \
  --set controller.replicas=3 \
  --set controller.args.maxConcurrentReconciles=10
```

## Upgrading

```bash
# Update repository
helm repo update

# Upgrade installation
helm upgrade secret-santa logiciq/secret-santa

# Upgrade with new values
helm upgrade secret-santa logiciq/secret-santa -f new-values.yaml
```

## Uninstalling

```bash
# Remove Helm release
helm uninstall secret-santa

# Remove CRDs (optional)
kubectl delete crd secretsantas.secrets.secret-santa.io
```

## Troubleshooting

### Pod Not Starting

Check pod status and logs:

```bash
kubectl get pods -l app.kubernetes.io/name=secret-santa
kubectl describe pod -l app.kubernetes.io/name=secret-santa
kubectl logs -l app.kubernetes.io/name=secret-santa
```

### AWS Authentication Issues

Verify IAM role and permissions:

```bash
# Check service account annotations
kubectl get sa secret-santa -o yaml

# Test AWS access from pod
kubectl exec -it deployment/secret-santa -- aws sts get-caller-identity
```

### GCP Authentication Issues

Verify Workload Identity setup:

```bash
# Check service account annotations
kubectl get sa secret-santa -o yaml

# Test GCP access from pod
kubectl exec -it deployment/secret-santa -- gcloud auth list
```

### RBAC Issues

Check cluster role and bindings:

```bash
kubectl get clusterrole secret-santa-manager-role
kubectl get clusterrolebinding secret-santa-manager-rolebinding
```