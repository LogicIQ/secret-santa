# secret-santa
Kubernetes operator for sensitive data generation with Go template support

## Features

### Core Functionality
- **Create-Once Policy**: Secrets are generated once and never modified
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

### Environment Variables
All flags support environment variables with `SECRET_SANTA_` prefix:
```bash
SECRET_SANTA_MAX_CONCURRENT_RECONCILES=5
SECRET_SANTA_WATCH_NAMESPACES=default,kube-system
SECRET_SANTA_DRY_RUN=true
SECRET_SANTA_LOG_FORMAT=console
SECRET_SANTA_LOG_LEVEL=debug
``` 
