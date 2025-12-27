# Example: Disabling Metadata

This example shows how to deploy Secret Santa with metadata disabled.

## Helm Installation

```bash
# Install with metadata disabled
helm install secret-santa logiciq/secret-santa \\\n  --set controller.args.enableMetadata=false

# Or via values file
cat > values.yaml << EOF
controller:
  args:
    enableMetadata: false
EOF

helm install secret-santa logiciq/secret-santa -f values.yaml
```

## Result

When metadata is disabled, generated secrets will only contain:
- User-defined labels and annotations from the SecretSanta spec
- The actual secret data

No system metadata annotations/tags will be added:
- `secrets.secret-santa.io/created-at`
- `secrets.secret-santa.io/generator-types`  
- `secrets.secret-santa.io/template-checksum`
- `secrets.secret-santa.io/source-cr`

## Use Cases

Disable metadata when:
- You have strict secret size requirements
- Your organization has its own metadata system
- You want minimal secrets without additional annotations
- Compliance requires clean secrets without system metadata