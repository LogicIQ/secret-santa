# HashiCorp Vault

Store generated secrets in HashiCorp Vault's KV v2 secrets engine.

## Prerequisites

- HashiCorp Vault with KV v2 enabled
- Network access from the operator pod to Vault
- A Vault token or Kubernetes auth role configured

## Token Auth

The simplest setup — provide a Vault token directly:

```yaml
apiVersion: secrets.secret-santa.io/v1alpha1
kind: SecretSanta
metadata:
  name: app-credentials
spec:
  template: |
    {
      "password": "{{ .pass.value }}",
      "api_key": "{{ .apikey.value }}"
    }
  generators:
    - name: pass
      type: random_password
      config:
        length: 32
    - name: apikey
      type: random_string
      config:
        length: 40
  media:
    type: hashicorp-vault
    config:
      address: https://vault.example.com
      mount_path: secret
      path: myapp/credentials
      token: <vault-token>
```

## Kubernetes Auth (Recommended)

Uses the operator's service account token — no static credentials needed:

```yaml
apiVersion: secrets.secret-santa.io/v1alpha1
kind: SecretSanta
metadata:
  name: database-credentials
spec:
  template: |
    {
      "username": "admin",
      "password": "{{ .pass.value }}"
    }
  generators:
    - name: pass
      type: random_password
      config:
        length: 32
  media:
    type: hashicorp-vault
    config:
      address: https://vault.example.com
      mount_path: secret
      path: myapp/database
      auth_method: kubernetes
      role: secret-santa
```

**Required Vault setup**:

```bash
vault auth enable kubernetes

vault write auth/kubernetes/config \
  kubernetes_host="https://kubernetes.default.svc"

vault policy write secret-santa - <<EOF
path "secret/data/*" {
  capabilities = ["create", "read", "update"]
}
EOF

vault write auth/kubernetes/role/secret-santa \
  bound_service_account_names=secret-santa \
  bound_service_account_namespaces=default \
  policies=secret-santa \
  ttl=1h
```

## Expected Output

The secret is stored at the configured path in Vault KV v2. Retrieve it with:

```bash
vault kv get secret/myapp/credentials
```

```
====== Data ======
Key        Value
---        -----
password   xK9mP2...
api_key    aB3nQ7...
```

## Troubleshooting

- **Permission denied** — verify the Vault policy grants `create`/`update` on the target path
- **Connection refused** — ensure the operator pod can reach the Vault address (use the in-cluster service URL when running in Kubernetes, e.g. `http://vault.vault.svc.cluster.local:8200`)
- **Secret not created** — check operator logs: `kubectl logs -n secret-santa-system deployment/secret-santa-controller`
