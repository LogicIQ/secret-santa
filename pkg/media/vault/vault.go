package vault

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	vault "github.com/hashicorp/vault/api"

	secretsantav1alpha1 "github.com/logicIQ/secret-santa/api/v1alpha1"
)

// HashiCorpVaultMedia stores secrets in HashiCorp Vault
type HashiCorpVaultMedia struct {
	Address    string
	Path       string
	MountPath  string
	Token      string
	Role       string
	AuthMethod string
}

func (m *HashiCorpVaultMedia) Store(ctx context.Context, secretSanta *secretsantav1alpha1.SecretSanta, data string, enableMetadata bool) error {
	config := vault.DefaultConfig()
	if m.Address != "" {
		config.Address = m.Address
	}

	client, err := vault.NewClient(config)
	if err != nil {
		return fmt.Errorf("failed to create Vault client: %w", err)
	}

	if m.Token != "" {
		client.SetToken(m.Token)
	}

	if err := m.authenticate(client); err != nil {
		return fmt.Errorf("failed to authenticate with Vault: %w", err)
	}

	secretPath := m.resolveSecretPath(secretSanta)
	mountPath := m.MountPath
	if mountPath == "" {
		mountPath = "secret"
	}

	// Parse data as JSON if possible, otherwise wrap as {"data": value}
	secretData := make(map[string]interface{})
	if err := json.Unmarshal([]byte(data), &secretData); err != nil {
		secretData["data"] = data
	}

	if enableMetadata {
		secretData["_metadata"] = map[string]interface{}{
			"created-at":        time.Now().UTC().Format(time.RFC3339),
			"generator-types":   m.getGeneratorTypes(secretSanta.Spec.Generators),
			"template-checksum": m.calculateTemplateChecksum(secretSanta.Spec.Template),
			"source-cr":         fmt.Sprintf("%s/%s", secretSanta.Namespace, secretSanta.Name),
		}
	}

	// Check if secret already exists (create-once policy)
	existing, err := client.KVv2(mountPath).Get(ctx, secretPath)
	if err == nil && existing != nil && existing.Data != nil {
		return nil
	}

	_, err = client.KVv2(mountPath).Put(ctx, secretPath, secretData)
	if err != nil {
		return fmt.Errorf("failed to write secret to Vault path %s: %w", secretPath, err)
	}

	return nil
}

func (m *HashiCorpVaultMedia) GetType() string {
	return "hashicorp-vault"
}

func (m *HashiCorpVaultMedia) authenticate(client *vault.Client) error {
	if m.Token != "" || m.AuthMethod == "" {
		return nil
	}

	switch m.AuthMethod {
	case "kubernetes":
		return m.authenticateKubernetes(client)
	default:
		return fmt.Errorf("unsupported auth method: %s", m.AuthMethod)
	}
}

func (m *HashiCorpVaultMedia) authenticateKubernetes(client *vault.Client) error {
	role := m.Role
	if role == "" {
		role = "secret-santa"
	}

	// Read the service account token from the pod
	jwt, err := readServiceAccountToken()
	if err != nil {
		return fmt.Errorf("failed to read service account token: %w", err)
	}

	params := map[string]interface{}{
		"jwt":  jwt,
		"role": role,
	}

	resp, err := client.Logical().Write("auth/kubernetes/login", params)
	if err != nil {
		return fmt.Errorf("kubernetes auth failed: %w", err)
	}
	if resp == nil || resp.Auth == nil {
		return fmt.Errorf("kubernetes auth returned empty response")
	}

	client.SetToken(resp.Auth.ClientToken)
	return nil
}

func (m *HashiCorpVaultMedia) resolveSecretPath(secretSanta *secretsantav1alpha1.SecretSanta) string {
	if m.Path != "" {
		return m.Path
	}
	name := secretSanta.Spec.SecretName
	if name == "" {
		name = secretSanta.Name
	}
	return fmt.Sprintf("%s/%s", secretSanta.Namespace, name)
}

func (m *HashiCorpVaultMedia) getGeneratorTypes(generators []secretsantav1alpha1.GeneratorConfig) string {
	types := make([]string, len(generators))
	for i, gen := range generators {
		types[i] = gen.Type
	}
	return strings.Join(types, ",")
}

func (m *HashiCorpVaultMedia) calculateTemplateChecksum(template string) string {
	hash := sha256.Sum256([]byte(template))
	return fmt.Sprintf("%x", hash)[:16]
}

func readServiceAccountToken() (string, error) {
	// Standard Kubernetes service account token path
	const tokenPath = "/var/run/secrets/kubernetes.io/serviceaccount/token"
	data, err := os.ReadFile(tokenPath)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}
