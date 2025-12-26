# Secret Santa Roadmap

## Current Status

### Supported Media Storage
- **Kubernetes Secrets** - Native K8s secret storage
- **AWS Secrets Manager** - AWS managed secrets service
- **AWS Parameter Store** - AWS Systems Manager parameters
- **GCP Secret Manager** - Google Cloud secret management

### Core Features
- Template-based secret generation
- Multiple generator types (random, TLS, crypto)
- Create-once workflow
- Multi-destination storage
- Cloud authentication (IAM roles, workload identity)

## Phase 1: Core Cloud Completion

**Priority: High**

- **Azure Key Vault** - Complete big 3 cloud providers
  - Managed identity authentication
  - Service principal support
  - Custom vault URL configuration
  - Secret versioning support

## Phase 2: Enterprise & Regional Expansion

**Priority: Medium-High**

- **Oracle Cloud Infrastructure (OCI) Vault**
  - Instance principal authentication
  - User principal support
  - Compartment-based access
- **Alibaba Cloud KMS**
  - RAM role authentication
  - Resource access management
  - Multi-region support

## Phase 3: Specialized Providers

**Priority: Medium**

- **IBM Cloud Secrets Manager**
  - IAM authentication
  - Secret groups support
  - Hybrid cloud integration
- **HashiCorp Vault** (Static secrets only)
  - Token authentication
  - AppRole authentication
  - KV v1/v2 support
  - Note: Dynamic secrets remain out of scope

## Phase 4: Developer & Edge Platforms

**Priority: Low-Medium**

- **DigitalOcean Spaces** (Object storage for secrets)
- **Vultr Object Storage**
- **Linode Object Storage**

## Phase 5: Advanced Features

**Priority: Based on User Feedback**

- **Secret Rotation Integration**
  - Webhook-based rotation triggers
  - External rotation orchestration
- **Multi-Region Replication**
  - Cross-region secret synchronization
  - Disaster recovery patterns
- **Advanced Templating**
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
- **High**: AWS, GCP, Azure (core clouds)
- **Medium-High**: OCI, Alibaba (enterprise/regional)
- **Medium**: IBM, Vault static (specialized)
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