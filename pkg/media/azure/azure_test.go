package azure

import (
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	secretsantav1alpha1 "github.com/logicIQ/secret-santa/api/v1alpha1"
)

func TestAzureKeyVaultMedia_GetType(t *testing.T) {
	media := &AzureKeyVaultMedia{}
	assert.Equal(t, "azure-key-vault", media.GetType())
}

func TestAzureKeyVaultMedia_SecretNameResolution(t *testing.T) {
	tests := []struct {
		name            string
		mediaSecretName string
		specSecretName  string
		metaName        string
		expectedName    string
	}{
		{
			name:            "media secret name takes priority",
			mediaSecretName: "media-secret",
			specSecretName:  "spec-secret",
			metaName:        "meta-name",
			expectedName:    "media-secret",
		},
		{
			name:            "spec secret name when no media name",
			mediaSecretName: "",
			specSecretName:  "spec-secret",
			metaName:        "meta-name",
			expectedName:    "spec-secret",
		},
		{
			name:            "meta name when no spec or media name",
			mediaSecretName: "",
			specSecretName:  "",
			metaName:        "meta-name",
			expectedName:    "meta-name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			media := &AzureKeyVaultMedia{
				SecretName: tt.mediaSecretName,
			}

			secretSanta := &secretsantav1alpha1.SecretSanta{
				ObjectMeta: metav1.ObjectMeta{
					Name:      tt.metaName,
					Namespace: "default",
				},
				Spec: secretsantav1alpha1.SecretSantaSpec{
					SecretName: tt.specSecretName,
				},
			}

			// Test name resolution using the actual method
			secretName := media.resolveSecretName(secretSanta)

			assert.Equal(t, tt.expectedName, secretName)
		})
	}
}

func TestAzureKeyVaultMedia_SecretNameSanitization(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "replace underscores with hyphens",
			input:    "my_secret_name",
			expected: "my-secret-name",
		},
		{
			name:     "replace dots with hyphens",
			input:    "my.secret.name",
			expected: "my-secret-name",
		},
		{
			name:     "mixed special characters",
			input:    "my_secret.name",
			expected: "my-secret-name",
		},
		{
			name:     "already valid name",
			input:    "my-secret-name",
			expected: "my-secret-name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the sanitization logic
			result := sanitizeAzureSecretName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAzureKeyVaultMedia_ConfigFields(t *testing.T) {
	media := &AzureKeyVaultMedia{
		VaultURL:   "https://my-vault.vault.azure.net",
		SecretName: "my-secret",
		TenantID:   "00000000-0000-0000-0000-000000000000",
	}

	assert.Equal(t, "https://my-vault.vault.azure.net", media.VaultURL)
	assert.Equal(t, "my-secret", media.SecretName)
	assert.Equal(t, "00000000-0000-0000-0000-000000000000", media.TenantID)
}

func TestAzureKeyVaultMedia_getGeneratorTypes(t *testing.T) {
	media := &AzureKeyVaultMedia{}
	generators := []secretsantav1alpha1.GeneratorConfig{
		{Type: "random_password"},
		{Type: "crypto_aes_key"},
	}
	result := media.getGeneratorTypes(generators)
	assert.Equal(t, "random_password,crypto_aes_key", result)
}

func TestAzureKeyVaultMedia_calculateTemplateChecksum(t *testing.T) {
	media := &AzureKeyVaultMedia{}
	template := "password: {{ .pass.password }}"
	checksum := media.calculateTemplateChecksum(template)
	assert.Len(t, checksum, 16)
	assert.Equal(t, checksum, media.calculateTemplateChecksum(template))
}

func TestAzureKeyVaultMedia_EmptyGenerators(t *testing.T) {
	media := &AzureKeyVaultMedia{}
	generators := []secretsantav1alpha1.GeneratorConfig{}
	result := media.getGeneratorTypes(generators)
	assert.Equal(t, "", result)
}

func TestAzureKeyVaultMedia_SingleGenerator(t *testing.T) {
	media := &AzureKeyVaultMedia{}
	generators := []secretsantav1alpha1.GeneratorConfig{
		{Type: "random_uuid"},
	}
	result := media.getGeneratorTypes(generators)
	assert.Equal(t, "random_uuid", result)
}
