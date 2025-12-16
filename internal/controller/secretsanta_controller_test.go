package controller

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	secretsantav1alpha1 "github.com/logicIQ/secret-santa/api/v1alpha1"
)

func TestSecretSantaReconciler_executeTemplate(t *testing.T) {
	tests := []struct {
		name     string
		template string
		data     map[string]interface{}
		want     string
		wantErr  bool
	}{
		{
			name:     "simple template",
			template: "Hello {{.name}}",
			data:     map[string]interface{}{"name": "World"},
			want:     "Hello World",
		},
		{
			name:     "template with generator data",
			template: "Password: {{.gen1.value}}",
			data: map[string]interface{}{
				"gen1": map[string]string{"value": "secret123"},
			},
			want: "Password: secret123",
		},
		{
			name:     "template with functions",
			template: "Hash: {{sha256 \"test\"}}",
			data:     map[string]interface{}{},
			want:     "Hash: 9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08",
		},
		{
			name:     "invalid template syntax",
			template: "{{.invalid syntax",
			data:     map[string]interface{}{},
			wantErr:  true,
		},
		{
			name:     "template execution error",
			template: "{{.nonexistent.field}}",
			data:     map[string]interface{}{},
			want:     "<no value>",
		},
	}

	r := &SecretSantaReconciler{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := r.executeTemplate(tt.template, tt.data)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSecretSantaReconciler_validateTemplate(t *testing.T) {
	tests := []struct {
		name     string
		template string
		wantErr  bool
	}{
		{
			name:     "valid template",
			template: "Hello {{.name}}",
		},
		{
			name:     "empty template",
			template: "",
			wantErr:  true,
		},
		{
			name:     "invalid syntax",
			template: "{{.invalid syntax",
			wantErr:  true,
		},
		{
			name:     "template with functions",
			template: "{{sha256 \"test\"}}",
		},
	}

	r := &SecretSantaReconciler{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := r.validateTemplate(tt.template)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSecretSantaReconciler_shouldProcess(t *testing.T) {
	tests := []struct {
		name               string
		includeAnnotations []string
		excludeAnnotations []string
		includeLabels      []string
		excludeLabels      []string
		secretSanta        *secretsantav1alpha1.SecretSanta
		want               bool
	}{
		{
			name: "no filters - should process",
			secretSanta: &secretsantav1alpha1.SecretSanta{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
			},
			want: true,
		},
		{
			name:               "include annotation present",
			includeAnnotations: []string{"process"},
			secretSanta: &secretsantav1alpha1.SecretSanta{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "test",
					Annotations: map[string]string{"process": "true"},
				},
			},
			want: true,
		},
		{
			name:               "include annotation missing",
			includeAnnotations: []string{"process"},
			secretSanta: &secretsantav1alpha1.SecretSanta{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
			},
			want: false,
		},
		{
			name:               "exclude annotation present",
			excludeAnnotations: []string{"skip"},
			secretSanta: &secretsantav1alpha1.SecretSanta{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "test",
					Annotations: map[string]string{"skip": "true"},
				},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &SecretSantaReconciler{
				IncludeAnnotations: tt.includeAnnotations,
				ExcludeAnnotations: tt.excludeAnnotations,
				IncludeLabels:      tt.includeLabels,
				ExcludeLabels:      tt.excludeLabels,
			}
			got := r.shouldProcess(tt.secretSanta)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSecretSantaReconciler_validateGeneratorConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  secretsantav1alpha1.GeneratorConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: secretsantav1alpha1.GeneratorConfig{
				Name: "test",
				Type: "random_password",
			},
		},
		{
			name: "empty name",
			config: secretsantav1alpha1.GeneratorConfig{
				Type: "random_password",
			},
			wantErr: true,
		},
		{
			name: "empty type",
			config: secretsantav1alpha1.GeneratorConfig{
				Name: "test",
			},
			wantErr: true,
		},
	}

	r := &SecretSantaReconciler{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := r.validateGeneratorConfig(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSecretSantaReconciler_generateTemplateData(t *testing.T) {
	tests := []struct {
		name     string
		configs  []secretsantav1alpha1.GeneratorConfig
		wantErr  bool
		wantKeys []string
	}{
		{
			name: "valid random password generator",
			configs: []secretsantav1alpha1.GeneratorConfig{
				{
					Name: "password",
					Type: "random_password",
					Config: &runtime.RawExtension{
						Raw: []byte(`{"length": 12}`),
					},
				},
			},
			wantKeys: []string{"password"},
		},
		{
			name: "multiple generators",
			configs: []secretsantav1alpha1.GeneratorConfig{
				{
					Name: "password",
					Type: "random_password",
					Config: &runtime.RawExtension{
						Raw: []byte(`{"length": 12}`),
					},
				},
				{
					Name: "uuid",
					Type: "random_uuid",
				},
			},
			wantKeys: []string{"password", "uuid"},
		},
		{
			name: "invalid generator config",
			configs: []secretsantav1alpha1.GeneratorConfig{
				{
					Type: "random_password",
				},
			},
			wantErr: true,
		},
		{
			name: "unknown generator type",
			configs: []secretsantav1alpha1.GeneratorConfig{
				{
					Name: "test",
					Type: "unknown_type",
				},
			},
			wantKeys: []string{},
		},
	}

	r := &SecretSantaReconciler{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := r.generateTemplateData(tt.configs)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)

			for _, key := range tt.wantKeys {
				assert.Contains(t, data, key)
			}
			assert.Len(t, data, len(tt.wantKeys))
		})
	}
}

func TestSecretSantaReconciler_createOrUpdateSecret(t *testing.T) {
	scheme := runtime.NewScheme()
	require.NoError(t, secretsantav1alpha1.AddToScheme(scheme))
	require.NoError(t, corev1.AddToScheme(scheme))

	tests := []struct {
		name       string
		secretType string
		data       string
		wantData   map[string]string
	}{
		{
			name:       "opaque secret",
			secretType: "Opaque",
			data:       "test-data",
			wantData:   map[string]string{"data": "test-data"},
		},
		{
			name:       "tls secret with proper format",
			secretType: "kubernetes.io/tls",
			data:       "tls.crt: cert-data\ntls.key: key-data",
			wantData:   map[string]string{"tls.crt": "cert-data", "tls.key": "key-data"},
		},
		{
			name:       "tls secret fallback",
			secretType: "kubernetes.io/tls",
			data:       "invalid-format",
			wantData:   map[string]string{"data": "invalid-format"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			secretSanta := &secretsantav1alpha1.SecretSanta{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-secret",
					Namespace: "default",
				},
				Spec: secretsantav1alpha1.SecretSantaSpec{
					SecretType:  tt.secretType,
					Labels:      map[string]string{"app": "test"},
					Annotations: map[string]string{"created-by": "secret-santa"},
				},
			}

			client := fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(secretSanta).
				WithStatusSubresource(&secretsantav1alpha1.SecretSanta{}).
				Build()

			r := &SecretSantaReconciler{
				Client: client,
				Scheme: scheme,
			}

			ctx := context.Background()
			result, err := r.createOrUpdateSecret(ctx, secretSanta, tt.data, "test-secret")

			require.NoError(t, err)
			assert.Equal(t, ctrl.Result{}, result)

			var secret corev1.Secret
			err = client.Get(ctx, types.NamespacedName{Name: "test-secret", Namespace: "default"}, &secret)
			require.NoError(t, err)

			for key, expectedValue := range tt.wantData {
				assert.Equal(t, expectedValue, secret.StringData[key])
			}
			assert.Equal(t, corev1.SecretType(tt.secretType), secret.Type)
			assert.Equal(t, map[string]string{"app": "test"}, secret.Labels)
			assert.Equal(t, map[string]string{"created-by": "secret-santa"}, secret.Annotations)
		})
	}
}

func TestSecretSantaReconciler_updateStatus(t *testing.T) {
	scheme := runtime.NewScheme()
	require.NoError(t, secretsantav1alpha1.AddToScheme(scheme))

	secretSanta := &secretsantav1alpha1.SecretSanta{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
	}

	client := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(secretSanta).
		WithStatusSubresource(&secretsantav1alpha1.SecretSanta{}).
		Build()

	r := &SecretSantaReconciler{
		Client: client,
		Scheme: scheme,
	}

	ctx := context.Background()
	err := r.updateStatus(ctx, secretSanta, "Ready", "True", "Test message")
	require.NoError(t, err)

	var updated secretsantav1alpha1.SecretSanta
	err = client.Get(ctx, types.NamespacedName{Name: "test", Namespace: "default"}, &updated)
	require.NoError(t, err)

	require.Len(t, updated.Status.Conditions, 1)
	assert.Equal(t, "Ready", updated.Status.Conditions[0].Type)
	assert.Equal(t, metav1.ConditionTrue, updated.Status.Conditions[0].Status)
	assert.Equal(t, "Test message", updated.Status.Conditions[0].Message)
	assert.NotNil(t, updated.Status.LastGenerated)
}

func TestGetMapKeys(t *testing.T) {
	tests := []struct {
		name string
		m    map[string]string
		want int
	}{
		{
			name: "empty map",
			m:    map[string]string{},
			want: 0,
		},
		{
			name: "single key",
			m:    map[string]string{"key1": "value1"},
			want: 1,
		},
		{
			name: "multiple keys",
			m:    map[string]string{"key1": "value1", "key2": "value2"},
			want: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getMapKeys(tt.m)
			assert.Len(t, got, tt.want)
		})
	}
}

func TestSecretSantaReconciler_handleDeletion(t *testing.T) {
	scheme := runtime.NewScheme()
	require.NoError(t, secretsantav1alpha1.AddToScheme(scheme))

	secretSanta := &secretsantav1alpha1.SecretSanta{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test",
			Namespace:  "default",
			Finalizers: []string{SecretSantaFinalizer},
		},
	}

	client := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(secretSanta).
		Build()

	r := &SecretSantaReconciler{
		Client: client,
		Scheme: scheme,
	}

	ctx := context.Background()
	result, err := r.handleDeletion(ctx, secretSanta)

	require.NoError(t, err)
	assert.Equal(t, ctrl.Result{}, result)
	assert.Empty(t, secretSanta.Finalizers)
}

func TestSecretSantaReconciler_shouldProcessFiltering(t *testing.T) {
	tests := []struct {
		name               string
		includeAnnotations []string
		excludeAnnotations []string
		includeLabels      []string
		excludeLabels      []string
		annotations        map[string]string
		labels             map[string]string
		want               bool
	}{
		{
			name: "no filters",
			want: true,
		},
		{
			name:               "include annotation present",
			includeAnnotations: []string{"process"},
			annotations:        map[string]string{"process": "true"},
			want:               true,
		},
		{
			name:               "include annotation missing",
			includeAnnotations: []string{"process"},
			want:               false,
		},
		{
			name:               "exclude annotation present",
			excludeAnnotations: []string{"skip"},
			annotations:        map[string]string{"skip": "true"},
			want:               false,
		},
		{
			name:          "include label present",
			includeLabels: []string{"env"},
			labels:        map[string]string{"env": "prod"},
			want:          true,
		},
		{
			name:          "exclude label present",
			excludeLabels: []string{"skip"},
			labels:        map[string]string{"skip": "true"},
			want:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &SecretSantaReconciler{
				IncludeAnnotations: tt.includeAnnotations,
				ExcludeAnnotations: tt.excludeAnnotations,
				IncludeLabels:      tt.includeLabels,
				ExcludeLabels:      tt.excludeLabels,
			}

			secretSanta := &secretsantav1alpha1.SecretSanta{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: tt.annotations,
					Labels:      tt.labels,
				},
			}

			got := r.shouldProcess(secretSanta)
			assert.Equal(t, tt.want, got)
		})
	}
}
