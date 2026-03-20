# Secret Santa Roadmap

## Current Status

### Supported Media Storage
- **Kubernetes Secrets** - Native K8s secret storage
- **AWS Secrets Manager** - AWS managed secrets service
- **AWS Parameter Store** - AWS Systems Manager parameters
- **GCP Secret Manager** - Google Cloud secret management
- **Azure Key Vault** - Azure managed secrets service
- **HashiCorp Vault** - KV v1/v2 with token and Kubernetes auth

### Core Features
- Template-based secret generation
- Multiple generator types (random, TLS, crypto)
- Create-once workflow
- Multi-destination storage
- Cloud authentication (IAM roles, workload identity, managed identity)
- Dry-run mode with masked output and template validation
- Automatic secret metadata for traceability

## Features to Implement

### Media Storage Providers

- **Oracle Cloud Infrastructure (OCI) Vault**
  - Instance principal authentication
  - User principal support
  - Compartment-based access

- **Alibaba Cloud KMS**
  - RAM role authentication
  - Resource access management
  - Multi-region support

- **IBM Cloud Secrets Manager**
  - IAM authentication
  - Secret groups support
  - Hybrid cloud integration

- **DigitalOcean Spaces** (Object storage for secrets)
- **Vultr Object Storage**
- **Linode Object Storage**

### High-Value Features

- **Secret Dependencies & Ordering**
  - Reference secrets from other SecretSanta CRs
  - Dependency resolution and creation ordering
  - Cross-namespace secret references

- **Backup & Recovery**
  - Cross-cluster migration support

- **Compliance & Auditing**
  - Detailed audit logs for secret operations
  - Compliance reporting (SOC2, PCI-DSS)
  - Secret lifecycle tracking

- **Multi-Destination Storage**
  - Store same secret in multiple backends simultaneously

### Operational Features

- **Monitoring & Alerting**
  - Prometheus metrics expansion
  - Alerting integration

- **Batch Operations**
  - Template libraries and reusable components
  - Namespace-wide operations

- **Security Enhancements**
  - Access control policies per CR
  - Integration with admission controllers

- **Secret Rotation Integration**
  - Webhook-based rotation triggers
  - External rotation orchestration

- **Multi-Region Replication**
  - Cross-region secret synchronization
  - Disaster recovery patterns

### Developer Experience

- **CLI Tool**
  - Local secret generation and testing
  - Template validation offline
  - Migration utilities

- **GitOps Integration**
  - ArgoCD/Flux integration patterns
  - Git-based secret management workflows

- **Enhanced Templating**
  - Include/import other templates
  - Conditional generation based on environment
  - Conditional logic
  - External data sources
  - Custom functions

## Out of Scope

**Will Not Support:**
- Dynamic Secret Generation (conflicts with create-once model)
- Vault Dynamic Secrets (use External Secrets Operator instead)
- Secret Rotation Management (external tools handle this better)
- Real-time Secret Updates (breaks immutability principle)

## Decision Framework

### Adding New Media Storage
**Criteria for inclusion:**
1. Market demand - Significant user requests
2. Cloud provider size - Major market share or regional dominance
3. Enterprise adoption - Used in production environments
4. Maintenance burden - Sustainable long-term support
5. API stability - Mature, stable APIs

### Priority Scoring
- **High**: AWS, GCP, Azure, HashiCorp Vault (shipped)
- **Medium-High**: OCI, Alibaba (enterprise/regional)
- **Medium**: IBM (specialized)
- **Low-Medium**: DigitalOcean, Vultr (developer platforms)

## Community Input

### Request Process
1. Open GitHub issue with use case
2. Provide market research/demand evidence
3. Community discussion and voting
4. Maintainer evaluation against criteria
5. Roadmap inclusion decision

### Contribution Welcome
- Media storage implementations
- Documentation improvements
- Testing and validation
- Performance optimizations