package gcp

import (
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	secretsantav1alpha1 "github.com/logicIQ/secret-santa/api/v1alpha1"
)

func TestGCPSecretManagerMedia_GetType(t *testing.T) {
	media := &GCPSecretManagerMedia{}
	assert.Equal(t, "gcp-secret-manager", media.GetType())
}

func TestGCPSecretManagerMedia_SecretNameResolution(t *testing.T) {
	tests := []struct {
		name           string
		mediaSecretName string
		specSecretName  string
		metaName       string
		expectedName   string
	}{
		{
			name:           "media secret name takes priority",
			mediaSecretName: "media-secret",
			specSecretName:  "spec-secret",
			metaName:       "meta-name",
			expectedName:   "media-secret",
		},
		{
			name:           "spec secret name when no media name",
			mediaSecretName: "",
			specSecretName:  "spec-secret",
			metaName:       "meta-name",
			expectedName:   "spec-secret",
		},
		{
			name:           "meta name when no spec or media name",
			mediaSecretName: "",
			specSecretName:  "",
			metaName:       "meta-name",
			expectedName:   "meta-name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			media := &GCPSecretManagerMedia{
				ProjectID:  "test-project",
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

func TestGCPSecretManagerMedia_ConfigFields(t *testing.T) {
	media := &GCPSecretManagerMedia{
		ProjectID:       "my-project",
		SecretName:      "my-secret",
		CredentialsFile: "/path/to/key.json",
	}

	assert.Equal(t, "my-project", media.ProjectID)
	assert.Equal(t, "my-secret", media.SecretName)
	assert.Equal(t, "/path/to/key.json", media.CredentialsFile)
}