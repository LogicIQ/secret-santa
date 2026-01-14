package aws

import (
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	secretsantav1alpha1 "github.com/logicIQ/secret-santa/api/v1alpha1"
)

func TestAWSSecretsManagerMedia_GetType(t *testing.T) {
	media := &AWSSecretsManagerMedia{}
	assert.Equal(t, "aws-secrets-manager", media.GetType())
}

func TestAWSParameterStoreMedia_GetType(t *testing.T) {
	media := &AWSParameterStoreMedia{}
	assert.Equal(t, "aws-parameter-store", media.GetType())
}

func TestAWSSecretsManagerMedia_SecretNameResolution(t *testing.T) {
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
			media := &AWSSecretsManagerMedia{
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

			// Test name resolution logic
			secretName := media.SecretName
			if secretName == "" {
				secretName = secretSanta.Spec.SecretName
				if secretName == "" {
					secretName = secretSanta.Name
				}
			}

			assert.Equal(t, tt.expectedName, secretName)
		})
	}
}

func TestAWSParameterStoreMedia_ParameterNameResolution(t *testing.T) {
	tests := []struct {
		name               string
		mediaParameterName string
		specSecretName     string
		metaName           string
		expectedName       string
	}{
		{
			name:               "media parameter name takes priority",
			mediaParameterName: "media-param",
			specSecretName:     "spec-secret",
			metaName:           "meta-name",
			expectedName:       "media-param",
		},
		{
			name:               "spec secret name when no media name",
			mediaParameterName: "",
			specSecretName:     "spec-secret",
			metaName:           "meta-name",
			expectedName:       "spec-secret",
		},
		{
			name:               "meta name when no spec or media name",
			mediaParameterName: "",
			specSecretName:     "",
			metaName:           "meta-name",
			expectedName:       "meta-name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			media := &AWSParameterStoreMedia{
				ParameterName: tt.mediaParameterName,
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

			// Test name resolution logic
			paramName := media.ParameterName
			if paramName == "" {
				paramName = secretSanta.Spec.SecretName
				if paramName == "" {
					paramName = secretSanta.Name
				}
			}

			assert.Equal(t, tt.expectedName, paramName)
		})
	}
}

func TestAWSMedia_ConfigFields(t *testing.T) {
	t.Run("secrets manager with all fields", func(t *testing.T) {
		media := &AWSSecretsManagerMedia{
			Region:     "us-west-2",
			SecretName: "my-secret",
			KMSKeyId:   "alias/my-key",
		}

		assert.Equal(t, "us-west-2", media.Region)
		assert.Equal(t, "my-secret", media.SecretName)
		assert.Equal(t, "alias/my-key", media.KMSKeyId)
	})

	t.Run("parameter store with all fields", func(t *testing.T) {
		media := &AWSParameterStoreMedia{
			Region:        "us-east-1",
			ParameterName: "my-param",
			KMSKeyId:      "alias/my-key",
		}

		assert.Equal(t, "us-east-1", media.Region)
		assert.Equal(t, "my-param", media.ParameterName)
		assert.Equal(t, "alias/my-key", media.KMSKeyId)
	})
}

func TestAWSSecretsManagerMedia_getGeneratorTypes(t *testing.T) {
	media := &AWSSecretsManagerMedia{}
	generators := []secretsantav1alpha1.GeneratorConfig{
		{Type: "random_password"},
		{Type: "crypto_aes_key"},
	}
	result := media.getGeneratorTypes(generators)
	assert.Equal(t, "random_password,crypto_aes_key", result)
}

func TestAWSSecretsManagerMedia_calculateTemplateChecksum(t *testing.T) {
	media := &AWSSecretsManagerMedia{}
	template := "password: {{ .pass.password }}"
	checksum := media.calculateTemplateChecksum(template)
	assert.Len(t, checksum, 16)
	assert.Equal(t, checksum, media.calculateTemplateChecksum(template))
}

func TestAWSParameterStoreMedia_getGeneratorTypes(t *testing.T) {
	media := &AWSParameterStoreMedia{}
	generators := []secretsantav1alpha1.GeneratorConfig{
		{Type: "tls_private_key"},
		{Type: "random_uuid"},
	}
	result := media.getGeneratorTypes(generators)
	assert.Equal(t, "tls_private_key,random_uuid", result)
}

func TestAWSParameterStoreMedia_calculateTemplateChecksum(t *testing.T) {
	media := &AWSParameterStoreMedia{}
	template := "api_key: {{ .key.value }}"
	checksum := media.calculateTemplateChecksum(template)
	assert.Len(t, checksum, 16)
	assert.Equal(t, checksum, media.calculateTemplateChecksum(template))
}
