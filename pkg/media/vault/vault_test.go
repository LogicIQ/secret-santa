package vault

import (
	"testing"

	vaultapi "github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	secretsantav1alpha1 "github.com/logicIQ/secret-santa/api/v1alpha1"
)

func TestHashiCorpVaultMedia_GetType(t *testing.T) {
	media := &HashiCorpVaultMedia{}
	assert.Equal(t, "hashicorp-vault", media.GetType())
}

func TestHashiCorpVaultMedia_SecretPathResolution(t *testing.T) {
	tests := []struct {
		name         string
		mediaPath    string
		specName     string
		metaName     string
		namespace    string
		expectedPath string
	}{
		{
			name:         "media path takes priority",
			mediaPath:    "myapp/credentials",
			specName:     "spec-secret",
			metaName:     "meta-name",
			namespace:    "default",
			expectedPath: "myapp/credentials",
		},
		{
			name:         "spec secret name when no media path",
			mediaPath:    "",
			specName:     "spec-secret",
			metaName:     "meta-name",
			namespace:    "default",
			expectedPath: "default/spec-secret",
		},
		{
			name:         "meta name when no spec or media path",
			mediaPath:    "",
			specName:     "",
			metaName:     "meta-name",
			namespace:    "default",
			expectedPath: "default/meta-name",
		},
		{
			name:         "custom namespace",
			mediaPath:    "",
			specName:     "",
			metaName:     "my-secret",
			namespace:    "production",
			expectedPath: "production/my-secret",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			media := &HashiCorpVaultMedia{
				Path: tt.mediaPath,
			}

			secretSanta := &secretsantav1alpha1.SecretSanta{
				ObjectMeta: metav1.ObjectMeta{
					Name:      tt.metaName,
					Namespace: tt.namespace,
				},
				Spec: secretsantav1alpha1.SecretSantaSpec{
					SecretName: tt.specName,
				},
			}

			path := media.resolveSecretPath(secretSanta)
			assert.Equal(t, tt.expectedPath, path)
		})
	}
}

func TestHashiCorpVaultMedia_ConfigFields(t *testing.T) {
	media := &HashiCorpVaultMedia{
		Address:    "https://vault.example.com",
		Path:       "myapp/credentials",
		MountPath:  "secret",
		Token:      "test-token",
		Role:       "secret-santa",
		AuthMethod: "kubernetes",
	}

	assert.Equal(t, "https://vault.example.com", media.Address)
	assert.Equal(t, "myapp/credentials", media.Path)
	assert.Equal(t, "secret", media.MountPath)
	assert.Equal(t, "test-token", media.Token)
	assert.Equal(t, "secret-santa", media.Role)
	assert.Equal(t, "kubernetes", media.AuthMethod)
}

func TestHashiCorpVaultMedia_getGeneratorTypes(t *testing.T) {
	media := &HashiCorpVaultMedia{}

	tests := []struct {
		name       string
		generators []secretsantav1alpha1.GeneratorConfig
		expected   string
	}{
		{
			name:       "multiple generators",
			generators: []secretsantav1alpha1.GeneratorConfig{{Type: "random_password"}, {Type: "crypto_aes_key"}},
			expected:   "random_password,crypto_aes_key",
		},
		{
			name:       "single generator",
			generators: []secretsantav1alpha1.GeneratorConfig{{Type: "random_uuid"}},
			expected:   "random_uuid",
		},
		{
			name:       "empty generators",
			generators: []secretsantav1alpha1.GeneratorConfig{},
			expected:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := media.getGeneratorTypes(tt.generators)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHashiCorpVaultMedia_calculateTemplateChecksum(t *testing.T) {
	media := &HashiCorpVaultMedia{}

	t.Run("returns 16 char hex string", func(t *testing.T) {
		checksum := media.calculateTemplateChecksum("password: {{ .pass.password }}")
		assert.Len(t, checksum, 16)
	})

	t.Run("deterministic", func(t *testing.T) {
		template := "api_key: {{ .key.value }}"
		assert.Equal(t, media.calculateTemplateChecksum(template), media.calculateTemplateChecksum(template))
	})

	t.Run("different templates produce different checksums", func(t *testing.T) {
		c1 := media.calculateTemplateChecksum("template-a")
		c2 := media.calculateTemplateChecksum("template-b")
		assert.NotEqual(t, c1, c2)
	})
}

func TestHashiCorpVaultMedia_authenticate(t *testing.T) {
	t.Run("skips auth when token is set", func(t *testing.T) {
		media := &HashiCorpVaultMedia{Token: "my-token", AuthMethod: "kubernetes"}
		// Should return nil because token takes precedence
		err := media.authenticate(nil)
		assert.NoError(t, err)
	})

	t.Run("skips auth when no auth method", func(t *testing.T) {
		media := &HashiCorpVaultMedia{}
		err := media.authenticate(nil)
		assert.NoError(t, err)
	})

	t.Run("rejects unsupported auth method", func(t *testing.T) {
		media := &HashiCorpVaultMedia{AuthMethod: "ldap"}
		config := vaultapi.DefaultConfig()
		config.Address = "http://127.0.0.1:1"
		client, _ := vaultapi.NewClient(config)
		err := media.authenticate(client)
		assert.ErrorContains(t, err, "unsupported auth method: ldap")
	})
}
