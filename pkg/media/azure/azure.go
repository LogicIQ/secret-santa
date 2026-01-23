package azure

import (
	"context"
	"crypto/sha256"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/keyvault/azsecrets"

	secretsantav1alpha1 "github.com/logicIQ/secret-santa/api/v1alpha1"
)

var (
	azureSecretNameInvalidCharsRegex = regexp.MustCompile(`[^0-9a-zA-Z-]`)
	azureSecretNameValidRegex        = regexp.MustCompile(`^[0-9a-zA-Z-]+$`)
)

// AzureKeyVaultMedia stores secrets in Azure Key Vault
type AzureKeyVaultMedia struct {
	VaultURL   string
	SecretName string
	TenantID   string
	client     *azsecrets.Client
	clientOnce sync.Once
	clientErr  error
}

func (m *AzureKeyVaultMedia) getClient(ctx context.Context) (*azsecrets.Client, error) {
	m.clientOnce.Do(func() {
		cred, err := azidentity.NewDefaultAzureCredential(nil)
		if err != nil {
			m.clientErr = fmt.Errorf("failed to create Azure credential: %w", err)
			return
		}

		client, err := azsecrets.NewClient(m.VaultURL, cred, nil)
		if err != nil {
			m.clientErr = fmt.Errorf("failed to create Azure Key Vault client: %w", err)
			return
		}

		m.client = client
	})
	return m.client, m.clientErr
}

func (m *AzureKeyVaultMedia) Store(ctx context.Context, secretSanta *secretsantav1alpha1.SecretSanta, data string, enableMetadata bool) error {
	client, err := m.getClient(ctx)
	if err != nil {
		return err
	}

	secretName := m.SecretName
	if secretName == "" {
		if secretSanta.Spec.SecretName != "" {
			secretName = secretSanta.Spec.SecretName
		} else {
			secretName = secretSanta.Name
		}
	}

	// Azure Key Vault secret names must match ^[0-9a-zA-Z-]+$
	secretName = sanitizeAzureSecretName(secretName)
	// Validate final secret name format
	if !isValidAzureSecretName(secretName) {
		return fmt.Errorf("invalid Azure Key Vault secret name after sanitization: %s", secretName)
	}

	// Build tags from labels, annotations, and metadata
	tags := make(map[string]*string)
	for k, v := range secretSanta.Spec.Labels {
		val := v
		tags[k] = &val
	}
	for k, v := range secretSanta.Spec.Annotations {
		val := v
		tags[k] = &val
	}

	// Add metadata tags only if enabled
	if enableMetadata {
		createdAt := time.Now().UTC().Format(time.RFC3339)
		generatorTypes := m.getGeneratorTypes(secretSanta.Spec.Generators)
		templateChecksum := m.calculateTemplateChecksum(secretSanta.Spec.Template)
		sourceCR := fmt.Sprintf("%s/%s", secretSanta.Namespace, secretSanta.Name)

		tags["secrets-secret-santa-io-created-at"] = &createdAt
		tags["secrets-secret-santa-io-generator-types"] = &generatorTypes
		tags["secrets-secret-santa-io-template-checksum"] = &templateChecksum
		tags["secrets-secret-santa-io-source-cr"] = &sourceCR
	}

	params := azsecrets.SetSecretParameters{
		Value: &data,
		Tags:  tags,
	}

	_, err = client.SetSecret(ctx, secretName, params, nil)
	if err != nil {
		return fmt.Errorf("failed to set secret %s in Azure Key Vault: %w", secretName, err)
	}
	return nil
}

func (m *AzureKeyVaultMedia) GetType() string {
	return "azure-key-vault"
}

// getGeneratorTypes extracts generator types from the configuration
func (m *AzureKeyVaultMedia) getGeneratorTypes(generators []secretsantav1alpha1.GeneratorConfig) string {
	types := make([]string, len(generators))
	for i, gen := range generators {
		types[i] = gen.Type
	}
	return strings.Join(types, ",")
}

// calculateTemplateChecksum creates a SHA256 checksum of the template
func (m *AzureKeyVaultMedia) calculateTemplateChecksum(template string) string {
	hash := sha256.Sum256([]byte(template))
	return fmt.Sprintf("%x", hash)[:16] // Use first 16 chars for brevity
}

// sanitizeAzureSecretName replaces all invalid characters with hyphens
func sanitizeAzureSecretName(name string) string {
	return azureSecretNameInvalidCharsRegex.ReplaceAllString(name, "-")
}

// isValidAzureSecretName validates Azure Key Vault secret name format
func isValidAzureSecretName(name string) bool {
	return len(name) > 0 && azureSecretNameValidRegex.MatchString(name)
}
