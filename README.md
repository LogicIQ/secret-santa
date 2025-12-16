# secret-santa
Kubernetes operator for sensitive data generation with Go template support

> **⚠️ Development Version Warning**  
> This is a development version and should **NOT** be used in production environments. The API and functionality may change without notice.

## Features

### Core Functionality
- **Create-Once Policy**: Secrets are generated once and never modified or touched again
- **Independent Lifecycle**: Secrets have no ownership references and persist independently of SecretSanta resources
- **Go Template Engine**: Full Go template support with custom functions
- **Template Validation**: Pre-validates templates before execution with user-friendly errors
- **Multiple Generators**: TLS, crypto, random, time-based data generation

### Configuration & Management
- **Cobra/Viper CLI**: Command-line flags and environment variable support
- **Parallelism Control**: Configurable concurrent reconciles (`--max-concurrent-reconciles`)
- **Namespace Filtering**: Watch specific namespaces or all (`--watch-namespaces`)
- **Resource Filtering**: Include/exclude by annotations and labels
- **Status Tracking**: Comprehensive CR status updates with conditions

### Operational Features
- **Leader Election**: Multi-replica deployment support
- **Health Checks**: Built-in readiness and liveness probes
- **Metrics**: Prometheus-compatible metrics endpoint
- **Flexible Deployment**: Configurable via CLI flags or environment variables

## Supported Generators

### TLS
- `tls_private_key` - Generate private keys
- `tls_self_signed_cert` - Self-signed certificates
- `tls_cert_request` - Certificate signing requests
- `tls_locally_signed_cert` - Locally signed certificates

### Cryptographic
- `crypto_hmac` - HMAC generation
- `crypto_aes_key` - AES encryption keys
- `crypto_rsa_key` - RSA key pairs
- `crypto_ed25519_key` - Ed25519 keys
- `crypto_chacha20_key` - ChaCha20 keys
- `crypto_xchacha20_key` - XChaCha20 keys
- `crypto_ecdsa_key` - ECDSA keys
- `crypto_ecdh_key` - ECDH keys

### Random Data
- `random_password` - Secure passwords
- `random_string` - Random strings
- `random_uuid` - UUIDs
- `random_integer` - Random integers
- `random_bytes` - Random byte arrays
- `random_id` - Random identifiers

### Time-based
- `time_static` - Static timestamps

## Configuration

### CLI Flags
```bash
--metrics-bind-address string      Metrics endpoint address (default ":8080")
--health-probe-bind-address string Health probe address (default ":8081")
--leader-elect                     Enable leader election
--max-concurrent-reconciles int    Max concurrent reconciles (default 1)
--watch-namespaces strings         Namespaces to watch (empty = all)
--include-annotations strings      Required annotations
--exclude-annotations strings      Excluded annotations
--include-labels strings           Required labels
--exclude-labels strings           Excluded labels
--dry-run                          Validate templates without creating secrets
--log-format string                Log format: json or console (default "json")
--log-level string                 Log level: debug, info, warn, error (default "info")
```

### Namespace Filtering
Control which namespaces the operator watches:

```bash
# Watch all namespaces (default)
secret-santa

# Watch specific namespaces
secret-santa --watch-namespaces default,kube-system
secret-santa --watch-namespaces default --watch-namespaces kube-system

# Environment variable
export SECRET_SANTA_WATCH_NAMESPACES=default,kube-system
```

### Resource Filtering
Filter SecretSanta resources by annotations and labels:

#### Include Filters (AND logic)
Only process resources that have ALL specified annotations/labels:

```bash
# Process only resources with specific annotations
secret-santa --include-annotations secret-santa.io/managed=true
secret-santa --include-annotations app.kubernetes.io/name,app.kubernetes.io/component

# Process only resources with specific labels
secret-santa --include-labels environment=production
secret-santa --include-labels app=web,tier=backend

# Environment variables
export SECRET_SANTA_INCLUDE_ANNOTATIONS=secret-santa.io/managed=true
export SECRET_SANTA_INCLUDE_LABELS=environment=production,app=web
```

#### Exclude Filters (OR logic)
Skip resources that have ANY of the specified annotations/labels:

```bash
# Skip resources with specific annotations
secret-santa --exclude-annotations secret-santa.io/ignore=true
secret-santa --exclude-annotations skip.secret-santa.io/ignore,deprecated

# Skip resources with specific labels
secret-santa --exclude-labels environment=development
secret-santa --exclude-labels skip=true,deprecated=true

# Environment variables
export SECRET_SANTA_EXCLUDE_ANNOTATIONS=secret-santa.io/ignore=true
export SECRET_SANTA_EXCLUDE_LABELS=environment=development
```

#### Filter Examples

**Production-only processing:**
```yaml
apiVersion: secrets.secret-santa.io/v1alpha1
kind: SecretSanta
metadata:
  name: prod-db-secret
  labels:
    environment: production
    app: database
spec:
  # ... secret configuration
```

```bash
# Only process production resources
secret-santa --include-labels environment=production
```

**Skip development resources:**
```yaml
apiVersion: secrets.secret-santa.io/v1alpha1
kind: SecretSanta
metadata:
  name: dev-test-secret
  annotations:
    secret-santa.io/ignore: "true"
spec:
  # ... secret configuration
```

```bash
# Skip resources marked for ignoring
secret-santa --exclude-annotations secret-santa.io/ignore=true
```

### Environment Variables
All flags support environment variables with `SECRET_SANTA_` prefix:
```bash
# Basic configuration
SECRET_SANTA_MAX_CONCURRENT_RECONCILES=5
SECRET_SANTA_DRY_RUN=true
SECRET_SANTA_LOG_FORMAT=console
SECRET_SANTA_LOG_LEVEL=debug

# Namespace and resource filtering
SECRET_SANTA_WATCH_NAMESPACES=default,kube-system,production
SECRET_SANTA_INCLUDE_ANNOTATIONS=secret-santa.io/managed=true,app.kubernetes.io/name
SECRET_SANTA_EXCLUDE_ANNOTATIONS=secret-santa.io/ignore=true,deprecated
SECRET_SANTA_INCLUDE_LABELS=environment=production,tier=backend
SECRET_SANTA_EXCLUDE_LABELS=skip=true,environment=development
```

## Secret Lifecycle

### Create-Once Policy
- Secrets are generated **exactly once** when the SecretSanta resource is first processed
- If a secret already exists, it is **never modified** - the operator skips creation
- No ownership references are set on secrets, making them completely independent

### Independent Lifecycle
- **No Controller References**: Secrets do not reference their creating SecretSanta resource
- **No Ownership**: Deleting a SecretSanta resource does not delete the generated secret
- **No Reconciliation**: The operator never reconciles or watches secret changes
- **Persistent**: Secrets remain in the cluster until manually deleted

This design ensures maximum security and stability - once created, secrets are never touched by the operator again.