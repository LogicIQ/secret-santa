# Self-Signed TLS Certificate

This example shows how to generate a complete TLS certificate with private key for development and testing purposes.

## Basic TLS Certificate

```yaml
apiVersion: secrets.secret-santa.io/v1alpha1
kind: SecretSanta
metadata:
  name: tls-certificate
  namespace: default
spec:
  secretType: kubernetes.io/tls
  template: |
    tls.crt: {{ .cert.certificate }}
    tls.key: {{ .key.private_key_pem }}
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
          common_name: "example.com"
          organization: "My Company"
          country: "US"
        validity_days: 365
```

## What This Creates

This creates a Kubernetes TLS secret that can be used directly with ingress controllers:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: tls-certificate
  namespace: default
type: kubernetes.io/tls
data:
  tls.crt: <base64-encoded-certificate>
  tls.key: <base64-encoded-private-key>
```

## Using with Ingress

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: my-app
spec:
  tls:
    - hosts:
        - example.com
      secretName: tls-certificate  # Reference the generated secret
  rules:
    - host: example.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: my-app
                port:
                  number: 80
```

## Advanced Configuration

### Multiple Subject Alternative Names

```yaml
apiVersion: secrets.secret-santa.io/v1alpha1
kind: SecretSanta
metadata:
  name: multi-san-cert
spec:
  secretType: kubernetes.io/tls
  template: |
    tls.crt: {{ .cert.certificate }}
    tls.key: {{ .key.private_key_pem }}
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
          common_name: "example.com"
          organization: "My Company"
          organizational_unit: "IT Department"
          country: "US"
          province: "California"
          locality: "San Francisco"
        dns_names:
          - "example.com"
          - "*.example.com"
          - "api.example.com"
          - "www.example.com"
        ip_addresses:
          - "192.168.1.100"
          - "10.0.0.1"
        validity_days: 365
```

### ECDSA Certificate

```yaml
apiVersion: secrets.secret-santa.io/v1alpha1
kind: SecretSanta
metadata:
  name: ecdsa-cert
spec:
  secretType: kubernetes.io/tls
  template: |
    tls.crt: {{ .cert.certificate }}
    tls.key: {{ .key.private_key_pem }}
  generators:
    - name: key
      type: tls_private_key
      config:
        algorithm: "ECDSA"
        curve: "P256"
    - name: cert
      type: tls_self_signed_cert
      config:
        key_pem: "{{ .key.private_key_pem }}"
        subject:
          common_name: "example.com"
        validity_days: 365
```

### Ed25519 Certificate

```yaml
apiVersion: secrets.secret-santa.io/v1alpha1
kind: SecretSanta
metadata:
  name: ed25519-cert
spec:
  secretType: kubernetes.io/tls
  template: |
    tls.crt: {{ .cert.certificate }}
    tls.key: {{ .key.private_key_pem }}
  generators:
    - name: key
      type: tls_private_key
      config:
        algorithm: "Ed25519"
    - name: cert
      type: tls_self_signed_cert
      config:
        key_pem: "{{ .key.private_key_pem }}"
        subject:
          common_name: "example.com"
        validity_days: 365
```

## Certificate Authority (CA)

Create a self-signed CA certificate:

```yaml
apiVersion: secrets.secret-santa.io/v1alpha1
kind: SecretSanta
metadata:
  name: ca-certificate
spec:
  template: |
    ca.crt: {{ .cert.certificate }}
    ca.key: {{ .key.private_key_pem }}
  generators:
    - name: key
      type: tls_private_key
      config:
        algorithm: "RSA"
        keySize: 4096
    - name: cert
      type: tls_self_signed_cert
      config:
        key_pem: "{{ .key.private_key_pem }}"
        subject:
          common_name: "My CA"
          organization: "My Company"
          country: "US"
        validity_days: 3650  # 10 years
        is_ca: true
```

## Certificate Signing Request

Generate a CSR for CA signing:

```yaml
apiVersion: secrets.secret-santa.io/v1alpha1
kind: SecretSanta
metadata:
  name: cert-request
spec:
  template: |
    tls.key: {{ .key.private_key_pem }}
    tls.csr: {{ .csr.request }}
  generators:
    - name: key
      type: tls_private_key
      config:
        algorithm: "RSA"
        keySize: 2048
    - name: csr
      type: tls_cert_request
      config:
        key_pem: "{{ .key.private_key_pem }}"
        subject:
          common_name: "app.example.com"
          organization: "My Company"
          organizational_unit: "Development"
        dns_names:
          - "app.example.com"
          - "api.example.com"
```

## Verification

### Check Certificate Details

```bash
# Extract and view certificate
kubectl get secret tls-certificate -o jsonpath='{.data.tls\.crt}' | base64 -d | openssl x509 -text -noout

# Check certificate validity
kubectl get secret tls-certificate -o jsonpath='{.data.tls\.crt}' | base64 -d | openssl x509 -noout -dates

# Verify certificate and key match
kubectl get secret tls-certificate -o jsonpath='{.data.tls\.crt}' | base64 -d | openssl x509 -noout -modulus | openssl md5
kubectl get secret tls-certificate -o jsonpath='{.data.tls\.key}' | base64 -d | openssl rsa -noout -modulus | openssl md5
```

### Test with OpenSSL

```bash
# Create a test server
kubectl get secret tls-certificate -o jsonpath='{.data.tls\.crt}' | base64 -d > cert.pem
kubectl get secret tls-certificate -o jsonpath='{.data.tls\.key}' | base64 -d > key.pem

# Test the certificate
openssl s_server -cert cert.pem -key key.pem -port 8443 -www
```

## Troubleshooting

### Invalid Subject

Ensure the subject contains at least a common name:

```yaml
subject:
  common_name: "example.com"  # Required
```

### Key Algorithm Mismatch

Make sure the certificate generator uses the same key that was generated:

```yaml
generators:
  - name: key
    type: tls_private_key
    config:
      algorithm: "RSA"  # Must match certificate expectations
  - name: cert
    type: tls_self_signed_cert
    config:
      key_pem: "{{ .key.private_key_pem }}"  # Reference the correct key
```

### Certificate Not Trusted

Self-signed certificates will show as untrusted in browsers. For production use:

1. Use a proper CA-signed certificate
2. Add the CA certificate to trust stores
3. Use cert-manager for automatic certificate management

### Ingress Controller Issues

Some ingress controllers require specific certificate formats or additional configuration. Check your ingress controller documentation for TLS secret requirements.