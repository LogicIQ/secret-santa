# Generators

Generators are the core components that create cryptographic material, passwords, and other secret data in Secret Santa. Each generator produces specific types of output that can be referenced in templates.

## Generator Basics

All generators follow the same basic structure:

```yaml
generators:
  - name: myGenerator        # Reference name for templates
    type: generator_type     # Type of generator
    config:                  # Type-specific configuration
      key: value
```

The `name` field is used to reference the generator output in templates: `{{ .myGenerator.field }}`

## Random Generators

### Random Password

Generates cryptographically secure passwords with configurable complexity.

```yaml
- name: dbpass
  type: random_password
  config:
    length: 32              # Password length (default: 16)
    includeSymbols: true    # Include symbols (!@#$%^&*) (default: true)
    excludeSimilar: false   # Exclude similar chars (0O1lI) (default: false)
    minNumeric: 2           # Minimum numeric characters (optional)
    minSymbols: 2           # Minimum symbol characters (optional)
```

**Template Usage**: `{{ .dbpass.password }}`

**Example Output**: `K9#mP2$vX8@nQ5!wR7&zL3*uY6%tE4^`

### Random String

Generates random strings with customizable character sets.

```yaml
- name: apikey
  type: random_string
  config:
    length: 64                    # String length (required)
    charset: "alphanumeric"       # Character set (default: "alphanumeric")
```

**Character Set Options**:
- `alphanumeric`: A-Z, a-z, 0-9
- `alpha`: A-Z, a-z
- `numeric`: 0-9
- `hex`: 0-9, A-F
- `base64`: A-Z, a-z, 0-9, +, /

**Template Usage**: `{{ .apikey.value }}`

### Random UUID

Generates UUID version 4 identifiers.

```yaml
- name: sessionid
  type: random_uuid
```

**Template Usage**: `{{ .sessionid.uuid }}`

**Example Output**: `f47ac10b-58cc-4372-a567-0e02b2c3d479`

### Random Bytes

Generates random byte arrays with configurable encoding.

```yaml
- name: salt
  type: random_bytes
  config:
    length: 32        # Byte length (required)
    encoding: "hex"   # Output encoding: "hex", "base64" (default: "base64")
```

**Template Usage**: `{{ .salt.value }}`

## Cryptographic Generators

### AES Key

Generates AES encryption keys.

```yaml
- name: encryption
  type: crypto_aes_key
  config:
    keySize: 256      # Key size: 128, 192, 256 (default: 256)
    encoding: "base64" # Output encoding: "hex", "base64" (default: "base64")
```

**Template Usage**: `{{ .encryption.key }}`

### RSA Key Pair

Generates RSA public/private key pairs.

```yaml
- name: signing
  type: crypto_rsa_key
  config:
    keySize: 2048     # Key size: 1024, 2048, 4096 (default: 2048)
```

**Template Usage**:
- Private key: `{{ .signing.private_key_pem }}`
- Public key: `{{ .signing.public_key_pem }}`

### Ed25519 Key Pair

Generates Ed25519 public/private key pairs for modern cryptographic applications.

```yaml
- name: modern
  type: crypto_ed25519_key
```

**Template Usage**:
- Private key: `{{ .modern.private_key_pem }}`
- Public key: `{{ .modern.public_key_pem }}`

## TLS Generators

### TLS Private Key

Generates private keys for TLS certificates with SSH public key formats.

```yaml
- name: tlskey
  type: tls_private_key
  config:
    algorithm: "RSA"    # Algorithm: "RSA", "ECDSA", "Ed25519" (default: "RSA")
    keySize: 2048       # RSA key size (default: 2048)
    curve: "P256"       # ECDSA curve: "P224", "P256", "P384", "P521"
```

**Template Usage**:
- Private key PEM: `{{ .tlskey.private_key_pem }}`
- Public key PEM: `{{ .tlskey.public_key_pem }}`
- OpenSSH public key: `{{ .tlskey.public_key_openssh }}`
- MD5 fingerprint: `{{ .tlskey.public_key_fingerprint_md5 }}`
- SHA256 fingerprint: `{{ .tlskey.public_key_fingerprint_sha256 }}`

**Note**: All algorithms (RSA, ECDSA, Ed25519) now provide the same output fields including SSH public key formats and fingerprints.

### Self-Signed Certificate

Generates self-signed X.509 certificates.

```yaml
- name: cert
  type: tls_self_signed_cert
  config:
    key_pem: "{{ .tlskey.private_key_pem }}"  # Private key (required)
    subject:
      common_name: "example.com"              # CN (required)
      organization: "My Company"              # O (optional)
      organizational_unit: "IT Department"    # OU (optional)
      country: "US"                          # C (optional)
      province: "California"                 # ST (optional)
      locality: "San Francisco"              # L (optional)
    dns_names:                               # Subject Alternative Names
      - "example.com"
      - "*.example.com"
      - "api.example.com"
    ip_addresses:                            # IP SANs
      - "192.168.1.100"
      - "10.0.0.1"
    validity_days: 365                       # Certificate validity (default: 365)
    is_ca: false                            # CA certificate (default: false)
```

**Template Usage**: `{{ .cert.certificate }}`

### Certificate Signing Request

Generates certificate signing requests (CSR) for CA-signed certificates.

```yaml
- name: csr
  type: tls_cert_request
  config:
    key_pem: "{{ .tlskey.private_key_pem }}"  # Private key (required)
    subject:
      common_name: "app.example.com"          # CN (required)
      organization: "My Company"              # O (optional)
    dns_names:                               # Subject Alternative Names
      - "app.example.com"
      - "api.example.com"
    ip_addresses:                            # IP SANs
      - "192.168.1.100"
```

**Template Usage**: `{{ .csr.request }}`

## Generator Dependencies

Generators can reference outputs from other generators defined earlier in the list:

```yaml
generators:
  # 1. Generate private key first
  - name: key
    type: tls_private_key
    config:
      algorithm: "RSA"
      keySize: 2048
  
  # 2. Use the private key to generate certificate
  - name: cert
    type: tls_self_signed_cert
    config:
      key_pem: "{{ .key.private_key_pem }}"  # Reference previous generator
      subject:
        common_name: "example.com"
      validity_days: 365
```

## Best Practices

### Security
- Use appropriate key sizes for your security requirements
- Consider using Ed25519 for new applications
- Set reasonable password complexity requirements
- Use strong random sources for cryptographic material

### Performance
- Order generators by dependency requirements
- Avoid unnecessary cryptographic operations
- Consider caching implications for expensive operations

### Maintainability
- Use descriptive generator names
- Document complex generator configurations
- Group related generators logically

## Common Patterns

### Database Credentials
```yaml
generators:
  - name: dbuser
    type: random_string
    config:
      length: 16
      charset: "alphanumeric"
  - name: dbpass
    type: random_password
    config:
      length: 32
      includeSymbols: true
```

### TLS Certificate Bundle
```yaml
generators:
  - name: key
    type: tls_private_key
  - name: cert
    type: tls_self_signed_cert
    config:
      key_pem: "{{ .key.private_key_pem }}"
      subject:
        common_name: "{{ .Values.hostname }}"
```

### API Authentication
```yaml
generators:
  - name: clientid
    type: random_uuid
  - name: secret
    type: random_string
    config:
      length: 64
      charset: "base64"
```