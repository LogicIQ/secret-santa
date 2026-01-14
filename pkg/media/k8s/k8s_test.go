package k8s

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	secretsantav1alpha1 "github.com/logicIQ/secret-santa/api/v1alpha1"
)

func TestK8sSecretsMedia_Store(t *testing.T) {
	scheme := runtime.NewScheme()
	require.NoError(t, corev1.AddToScheme(scheme))
	require.NoError(t, secretsantav1alpha1.AddToScheme(scheme))

	tests := []struct {
		name            string
		secretSanta     *secretsantav1alpha1.SecretSanta
		mediaSecretName string
		data            string
		expectedName    string
		expectedData    map[string]string
	}{
		{
			name: "basic secret with default name",
			secretSanta: &secretsantav1alpha1.SecretSanta{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-secret",
					Namespace: "default",
				},
				Spec: secretsantav1alpha1.SecretSantaSpec{
					SecretType: "Opaque",
				},
			},
			data:         "password: secret123",
			expectedName: "test-secret",
			expectedData: map[string]string{"data": "password: secret123"},
		},
		{
			name: "secret with spec secretName",
			secretSanta: &secretsantav1alpha1.SecretSanta{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-secret",
					Namespace: "default",
				},
				Spec: secretsantav1alpha1.SecretSantaSpec{
					SecretName: "custom-secret",
					SecretType: "Opaque",
				},
			},
			data:         "password: secret123",
			expectedName: "custom-secret",
			expectedData: map[string]string{"data": "password: secret123"},
		},
		{
			name: "secret with media secretName override",
			secretSanta: &secretsantav1alpha1.SecretSanta{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-secret",
					Namespace: "default",
				},
				Spec: secretsantav1alpha1.SecretSantaSpec{
					SecretName: "custom-secret",
					SecretType: "Opaque",
				},
			},
			mediaSecretName: "media-override-secret",
			data:            "password: secret123",
			expectedName:    "media-override-secret",
			expectedData:    map[string]string{"data": "password: secret123"},
		},
		{
			name: "TLS secret with proper parsing",
			secretSanta: &secretsantav1alpha1.SecretSanta{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "tls-secret",
					Namespace: "default",
				},
				Spec: secretsantav1alpha1.SecretSantaSpec{
					SecretType: "kubernetes.io/tls",
				},
			},
			data:         "tls.crt: cert-data\ntls.key: key-data",
			expectedName: "tls-secret",
			expectedData: map[string]string{
				"tls.crt": "cert-data",
				"tls.key": "key-data",
			},
		},
		{
			name: "TLS secret with incomplete data returns error",
			secretSanta: &secretsantav1alpha1.SecretSanta{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "tls-secret",
					Namespace: "default",
				},
				Spec: secretsantav1alpha1.SecretSantaSpec{
					SecretType: "kubernetes.io/tls",
				},
			},
			data:         "incomplete tls data",
			expectedName: "tls-secret",
			expectedData: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := fake.NewClientBuilder().WithScheme(scheme).Build()
			media := &K8sSecretsMedia{
				Client:     client,
				SecretName: tt.mediaSecretName,
			}

			err := media.Store(context.Background(), tt.secretSanta, tt.data, true)

			// If expectedData is nil, we expect an error
			if tt.expectedData == nil {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			// Verify secret was created
			var secret corev1.Secret
			err = client.Get(context.Background(),
				types.NamespacedName{Name: tt.expectedName, Namespace: "default"}, &secret)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedName, secret.Name)
			assert.Equal(t, "default", secret.Namespace)
			assert.Equal(t, corev1.SecretType(tt.secretSanta.Spec.SecretType), secret.Type)
			assert.Equal(t, tt.expectedData, secret.StringData)
		})
	}
}

func TestK8sSecretsMedia_GetType(t *testing.T) {
	media := &K8sSecretsMedia{}
	assert.Equal(t, "k8s", media.GetType())
}

func TestK8sSecretsMedia_StoreWithLabelsAndAnnotations(t *testing.T) {
	scheme := runtime.NewScheme()
	require.NoError(t, corev1.AddToScheme(scheme))
	require.NoError(t, secretsantav1alpha1.AddToScheme(scheme))

	secretSanta := &secretsantav1alpha1.SecretSanta{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-secret",
			Namespace: "default",
		},
		Spec: secretsantav1alpha1.SecretSantaSpec{
			SecretType: "Opaque",
			Labels: map[string]string{
				"app":         "test-app",
				"environment": "test",
			},
			Annotations: map[string]string{
				"description": "Test secret",
				"version":     "1.0",
			},
		},
	}

	client := fake.NewClientBuilder().WithScheme(scheme).Build()
	media := &K8sSecretsMedia{Client: client}

	err := media.Store(context.Background(), secretSanta, "test-data", true)
	require.NoError(t, err)

	var secret corev1.Secret
	err = client.Get(context.Background(),
		types.NamespacedName{Name: "test-secret", Namespace: "default"}, &secret)
	require.NoError(t, err)

	assert.Equal(t, secretSanta.Spec.Labels, secret.Labels)

	// Check user annotations are preserved
	for k, v := range secretSanta.Spec.Annotations {
		assert.Equal(t, v, secret.Annotations[k])
	}

	// Check metadata annotations are added
	assert.Contains(t, secret.Annotations, "secrets.secret-santa.io/created-at")
	assert.Contains(t, secret.Annotations, "secrets.secret-santa.io/generator-types")
	assert.Contains(t, secret.Annotations, "secrets.secret-santa.io/template-checksum")
	assert.Contains(t, secret.Annotations, "secrets.secret-santa.io/source-cr")
}

func TestK8sSecretsMedia_MetadataEnhancement(t *testing.T) {
	scheme := runtime.NewScheme()
	require.NoError(t, corev1.AddToScheme(scheme))
	require.NoError(t, secretsantav1alpha1.AddToScheme(scheme))

	secretSanta := &secretsantav1alpha1.SecretSanta{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-secret",
			Namespace: "default",
		},
		Spec: secretsantav1alpha1.SecretSantaSpec{
			Template: "password: {{ .pass.password }}",
			Generators: []secretsantav1alpha1.GeneratorConfig{
				{Name: "pass", Type: "random_password"},
				{Name: "api", Type: "random_string"},
			},
			SecretType: "Opaque",
		},
	}

	client := fake.NewClientBuilder().WithScheme(scheme).Build()
	media := &K8sSecretsMedia{Client: client}

	err := media.Store(context.Background(), secretSanta, "test-data", true)
	require.NoError(t, err)

	var secret corev1.Secret
	err = client.Get(context.Background(),
		types.NamespacedName{Name: "test-secret", Namespace: "default"}, &secret)
	require.NoError(t, err)

	// Check metadata annotations
	assert.Contains(t, secret.Annotations, "secrets.secret-santa.io/created-at")
	assert.Equal(t, "random_password,random_string", secret.Annotations["secrets.secret-santa.io/generator-types"])
	assert.Contains(t, secret.Annotations, "secrets.secret-santa.io/template-checksum")
	assert.Equal(t, "default/test-secret", secret.Annotations["secrets.secret-santa.io/source-cr"])
}

func TestK8sSecretsMedia_getGeneratorTypes(t *testing.T) {
	media := &K8sSecretsMedia{}
	generators := []secretsantav1alpha1.GeneratorConfig{
		{Type: "random_password"},
		{Type: "random_string"},
		{Type: "tls_private_key"},
	}
	result := media.getGeneratorTypes(generators)
	assert.Equal(t, "random_password,random_string,tls_private_key", result)
}

func TestK8sSecretsMedia_calculateTemplateChecksum(t *testing.T) {
	media := &K8sSecretsMedia{}
	template := "password: {{ .pass.password }}"
	checksum := media.calculateTemplateChecksum(template)
	assert.Len(t, checksum, 16)
	// Same template should produce same checksum
	assert.Equal(t, checksum, media.calculateTemplateChecksum(template))
}

func TestK8sSecretsMedia_MetadataDisabled(t *testing.T) {
	scheme := runtime.NewScheme()
	require.NoError(t, corev1.AddToScheme(scheme))
	require.NoError(t, secretsantav1alpha1.AddToScheme(scheme))

	secretSanta := &secretsantav1alpha1.SecretSanta{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-secret",
			Namespace: "default",
		},
		Spec: secretsantav1alpha1.SecretSantaSpec{
			Template: "password: {{ .pass.password }}",
			Generators: []secretsantav1alpha1.GeneratorConfig{
				{Name: "pass", Type: "random_password"},
			},
			SecretType:  "Opaque",
			Annotations: map[string]string{"user": "test"},
		},
	}

	client := fake.NewClientBuilder().WithScheme(scheme).Build()
	media := &K8sSecretsMedia{Client: client}

	err := media.Store(context.Background(), secretSanta, "test-data", false)
	require.NoError(t, err)

	var secret corev1.Secret
	err = client.Get(context.Background(),
		types.NamespacedName{Name: "test-secret", Namespace: "default"}, &secret)
	require.NoError(t, err)

	// Check only user annotations exist, no metadata
	assert.Equal(t, "test", secret.Annotations["user"])
	assert.NotContains(t, secret.Annotations, "secrets.secret-santa.io/created-at")
	assert.NotContains(t, secret.Annotations, "secrets.secret-santa.io/generator-types")
	assert.NotContains(t, secret.Annotations, "secrets.secret-santa.io/template-checksum")
	assert.NotContains(t, secret.Annotations, "secrets.secret-santa.io/source-cr")
}
