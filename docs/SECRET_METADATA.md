# Secret Metadata Enhancement

This feature automatically adds metadata to generated secrets across all storage backends to improve observability and traceability.

## Added Metadata

### Kubernetes Secrets (Annotations)
- `secrets.secret-santa.io/created-at`: ISO 8601 timestamp of secret creation
- `secrets.secret-santa.io/generator-types`: Comma-separated list of generator types used
- `secrets.secret-santa.io/template-checksum`: SHA256 checksum (first 16 chars) of the template
- `secrets.secret-santa.io/source-cr`: Reference to the source SecretSanta CR (`namespace/name`)

### AWS Secrets Manager & Parameter Store (Tags)
- `secrets.secret-santa.io/created-at`: ISO 8601 timestamp of secret creation
- `secrets.secret-santa.io/generator-types`: Comma-separated list of generator types used
- `secrets.secret-santa.io/template-checksum`: SHA256 checksum (first 16 chars) of the template
- `secrets.secret-santa.io/source-cr`: Reference to the source SecretSanta CR (`namespace/name`)

### GCP Secret Manager (Labels)
- `secrets_secret-santa_io_created-at`: ISO 8601 timestamp of secret creation
- `secrets_secret-santa_io_generator-types`: Comma-separated list of generator types used
- `secrets_secret-santa_io_template-checksum`: SHA256 checksum (first 16 chars) of the template
- `secrets_secret-santa_io_source-cr`: Reference to the source SecretSanta CR (`namespace_name`)

*Note: GCP labels use underscores instead of dots/slashes due to label naming restrictions*

## Example Output

### Kubernetes Secret
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: my-secret
  namespace: default
  annotations:
    secrets.secret-santa.io/created-at: "2024-01-15T10:30:00Z"
    secrets.secret-santa.io/generator-types: "random_password,random_string"
    secrets.secret-santa.io/template-checksum: "a1b2c3d4e5f6g7h8"
    secrets.secret-santa.io/source-cr: "default/my-secretsanta"
data:
  # ... secret data
```

### AWS Secrets Manager
Tags will include the same metadata keys with their respective values.

### GCP Secret Manager
Labels will include the metadata with underscores replacing dots and slashes.

## Configuration Options

### Global Controller Flag
Control metadata generation globally via controller flag:

```bash
# Enable metadata (default)
helm install secret-santa logiciq/secret-santa \
  --set controller.args.enableMetadata=true

# Disable metadata
helm install secret-santa logiciq/secret-santa \
  --set controller.args.enableMetadata=false
```

### Environment Variable
```bash
SECRET_SANTA_ENABLE_METADATA=false
```

### Command Line Flag
```bash
./secret-santa --enable-metadata=false
```

## Benefits

1. **Traceability**: Easily identify which SecretSanta CR generated a secret
2. **Change Detection**: Template checksum helps detect when templates have changed
3. **Audit Trail**: Creation timestamps provide audit information
4. **Generator Visibility**: Know which generators were used to create the secret
5. **Debugging**: Easier troubleshooting with metadata context

## Implementation Details

- Metadata is added automatically to all storage backends
- User-defined labels and annotations are preserved and merged with metadata
- Template checksums use SHA256 with first 16 characters for brevity
- Timestamps use RFC3339 format in UTC
- Generator types are comma-separated for multiple generators