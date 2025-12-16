package controller

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log"

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
		{
			name:          "include label present",
			includeLabels: []string{"env"},
			secretSanta: &secretsantav1alpha1.SecretSanta{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "test",
					Labels: map[string]string{"env": "prod"},
				},
			},
			want: true,
		},
		{
			name:          "exclude label present",
			excludeLabels: []string{"skip"},
			secretSanta: &secretsantav1alpha1.SecretSanta{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "test",
					Labels: map[string]string{"skip": "true"},
				},
			},
			want: false,
		},
		{
			name:               "multiple include annotations - all present",
			includeAnnotations: []string{"app.kubernetes.io/name", "app.kubernetes.io/component"},
			secretSanta: &secretsantav1alpha1.SecretSanta{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
					Annotations: map[string]string{
						"app.kubernetes.io/name":      "myapp",
						"app.kubernetes.io/component": "backend",
					},
				},
			},
			want: true,
		},
		{
			name:               "multiple include annotations - one missing",
			includeAnnotations: []string{"app.kubernetes.io/name", "app.kubernetes.io/component"},
			secretSanta: &secretsantav1alpha1.SecretSanta{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "test",
					Annotations: map[string]string{"app.kubernetes.io/name": "myapp"},
				},
			},
			want: false,
		},
		{
			name:               "multiple exclude annotations - one present",
			excludeAnnotations: []string{"skip", "deprecated"},
			secretSanta: &secretsantav1alpha1.SecretSanta{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "test",
					Annotations: map[string]string{"skip": "true"},
				},
			},
			want: false,
		},
		{
			name:          "multiple include labels - all present",
			includeLabels: []string{"environment", "app"},
			secretSanta: &secretsantav1alpha1.SecretSanta{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
					Labels: map[string]string{
						"environment": "production",
						"app":         "web",
					},
				},
			},
			want: true,
		},
		{
			name:          "multiple include labels - one missing",
			includeLabels: []string{"environment", "app"},
			secretSanta: &secretsantav1alpha1.SecretSanta{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "test",
					Labels: map[string]string{"environment": "production"},
				},
			},
			want: false,
		},
		{
			name:          "multiple exclude labels - one present",
			excludeLabels: []string{"skip", "deprecated"},
			secretSanta: &secretsantav1alpha1.SecretSanta{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "test",
					Labels: map[string]string{"deprecated": "true"},
				},
			},
			want: false,
		},
		{
			name:               "combined filters - include and exclude annotations",
			includeAnnotations: []string{"process"},
			excludeAnnotations: []string{"skip"},
			secretSanta: &secretsantav1alpha1.SecretSanta{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
					Annotations: map[string]string{
						"process": "true",
						"skip":    "true",
					},
				},
			},
			want: false,
		},
		{
			name:               "combined filters - include present, exclude absent",
			includeAnnotations: []string{"process"},
			excludeAnnotations: []string{"skip"},
			secretSanta: &secretsantav1alpha1.SecretSanta{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "test",
					Annotations: map[string]string{"process": "true"},
				},
			},
			want: true,
		},
		{
			name:          "combined filters - labels and annotations",
			includeLabels: []string{"environment"},
			excludeLabels: []string{"skip"},
			includeAnnotations: []string{"process"},
			excludeAnnotations: []string{"ignore"},
			secretSanta: &secretsantav1alpha1.SecretSanta{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
					Labels: map[string]string{
						"environment": "production",
					},
					Annotations: map[string]string{
						"process": "true",
					},
				},
			},
			want: true,
		},
		{
			name:               "nil annotations map",
			includeAnnotations: []string{"process"},
			secretSanta: &secretsantav1alpha1.SecretSanta{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
			},
			want: false,
		},
		{
			name:          "nil labels map",
			includeLabels: []string{"environment"},
			secretSanta: &secretsantav1alpha1.SecretSanta{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
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
		name      string
		configs   []secretsantav1alpha1.GeneratorConfig
		wantErr   bool
		wantKeys  []string
	}{
		{
			name: "valid random password generator",
			configs: []secretsantav1alpha1.GeneratorConfig{
				{
					Name: "password",
					Type: "random_password",
					Config: &runtime.RawExtension{Raw: []byte(`{
						"length": 12}`),
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
					Config: &runtime.RawExtension{Raw: []byte(`{
						"length": 12}`),
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

func TestSecretSantaReconciler_Reconcile(t *testing.T) {
	scheme := runtime.NewScheme()
	require.NoError(t, secretsantav1alpha1.AddToScheme(scheme))
	require.NoError(t, corev1.AddToScheme(scheme))

	tests := []struct {
		name        string
		secretSanta *secretsantav1alpha1.SecretSanta
		existing    []client.Object
		dryRun      bool
		wantResult  ctrl.Result
		wantErr     bool
		wantSecret  bool
	}{
		{
			name: "resource not found",
			wantResult: ctrl.Result{},
		},
		{
			name: "successful reconcile",
			secretSanta: &secretsantav1alpha1.SecretSanta{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-secret",
					Namespace: "default",
					Finalizers: []string{SecretSantaFinalizer},
				},
				Spec: secretsantav1alpha1.SecretSantaSpec{
					SecretType: "Opaque",
					Template:   "password: {{.password.value}}",
					Generators: []secretsantav1alpha1.GeneratorConfig{
						{
							Name: "password",
							Type: "random_password",
							Config: &runtime.RawExtension{Raw: []byte(`{
								"length": 12}`),
							},
						},
					},
				},
			},
			wantResult: ctrl.Result{},
			wantSecret: true,
		},
		{
			name: "dry run mode",
			secretSanta: &secretsantav1alpha1.SecretSanta{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-secret",
					Namespace: "default",
					Finalizers: []string{SecretSantaFinalizer},
				},
				Spec: secretsantav1alpha1.SecretSantaSpec{
					SecretType: "Opaque",
					Template:   "password: {{.password.value}}",
					Generators: []secretsantav1alpha1.GeneratorConfig{
						{
							Name: "password",
							Type: "random_password",
							Config: &runtime.RawExtension{Raw: []byte(`{
								"length": 12}`),
							},
						},
					},
				},
			},
			dryRun:     true,
			wantResult: ctrl.Result{},
			wantSecret: false,
		},
		{
			name: "secret already exists",
			secretSanta: &secretsantav1alpha1.SecretSanta{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-secret",
					Namespace: "default",
					Finalizers: []string{SecretSantaFinalizer},
				},
				Spec: secretsantav1alpha1.SecretSantaSpec{
					SecretType: "Opaque",
					Template:   "password: test",
				},
			},
			existing: []client.Object{
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-secret",
						Namespace: "default",
					},
				},
			},
			wantResult: ctrl.Result{},
			wantSecret: true,
		},
		{
			name: "invalid template",
			secretSanta: &secretsantav1alpha1.SecretSanta{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-secret",
					Namespace: "default",
					Finalizers: []string{SecretSantaFinalizer},
				},
				Spec: secretsantav1alpha1.SecretSantaSpec{
					SecretType: "Opaque",
					Template:   "{{.invalid syntax",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			objs := tt.existing
			if tt.secretSanta != nil {
				objs = append(objs, tt.secretSanta)
			}

			client := fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(objs...).
				WithStatusSubresource(&secretsantav1alpha1.SecretSanta{}).
				Build()

			r := &SecretSantaReconciler{
				Client: client,
				Scheme: scheme,
				DryRun: tt.dryRun,
			}

			req := ctrl.Request{
				NamespacedName: types.NamespacedName{
					Name:      "test-secret",
					Namespace: "default",
				},
			}

			ctx := log.IntoContext(context.Background(), log.Log)
			result, err := r.Reconcile(ctx, req)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantResult, result)

			if tt.wantSecret && !tt.dryRun {
				var secret corev1.Secret
				err := client.Get(ctx, req.NamespacedName, &secret)
				assert.NoError(t, err)
			}
		})
	}
}

func TestSecretSantaReconciler_handleDeletion(t *testing.T) {
	scheme := runtime.NewScheme()
	require.NoError(t, secretsantav1alpha1.AddToScheme(scheme))

	now := metav1.Now()
	secretSanta := &secretsantav1alpha1.SecretSanta{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-secret",
			Namespace:         "default",
			DeletionTimestamp: &now,
			Finalizers:        []string{SecretSantaFinalizer},
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

	ctx := log.IntoContext(context.Background(), log.Log)
	result, err := r.handleDeletion(ctx, secretSanta)

	require.NoError(t, err)
	assert.Equal(t, ctrl.Result{}, result)
	assert.Empty(t, secretSanta.Finalizers)
}

func TestSecretSantaReconciler_createOrUpdateSecret(t *testing.T) {
	scheme := runtime.NewScheme()
	require.NoError(t, secretsantav1alpha1.AddToScheme(scheme))
	require.NoError(t, corev1.AddToScheme(scheme))

	secretSanta := &secretsantav1alpha1.SecretSanta{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-secret",
			Namespace: "default",
		},
		Spec: secretsantav1alpha1.SecretSantaSpec{
			SecretType: "Opaque",
			Labels:     map[string]string{"app": "test"},
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

	ctx := log.IntoContext(context.Background(), log.Log)
	result, err := r.createOrUpdateSecret(ctx, secretSanta, "test-data", "test-secret")

	require.NoError(t, err)
	assert.Equal(t, ctrl.Result{}, result)

	var secret corev1.Secret
	err = client.Get(ctx, types.NamespacedName{
		Name:      "test-secret",
		Namespace: "default",
	}, &secret)
	require.NoError(t, err)

	assert.Equal(t, "test-data", secret.StringData["data"])
	assert.Equal(t, corev1.SecretType("Opaque"), secret.Type)
	assert.Equal(t, map[string]string{"app": "test"}, secret.Labels)
	assert.Equal(t, map[string]string{"created-by": "secret-santa"}, secret.Annotations)
}

func TestSecretSantaReconciler_updateStatus(t *testing.T) {
	scheme := runtime.NewScheme()
	require.NoError(t, secretsantav1alpha1.AddToScheme(scheme))

	secretSanta := &secretsantav1alpha1.SecretSanta{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-secret",
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
	err := r.updateStatus(ctx, secretSanta, "Ready", "True", "Success")
	require.NoError(t, err)

	assert.NotNil(t, secretSanta.Status.LastGenerated)
	require.Len(t, secretSanta.Status.Conditions, 1)
	
	condition := secretSanta.Status.Conditions[0]
	assert.Equal(t, "Ready", condition.Type)
	assert.Equal(t, metav1.ConditionTrue, condition.Status)
	assert.Equal(t, "Success", condition.Message)
}

func TestSecretSantaReconciler_ReconcileWithFiltering(t *testing.T) {
	scheme := runtime.NewScheme()
	require.NoError(t, secretsantav1alpha1.AddToScheme(scheme))
	require.NoError(t, corev1.AddToScheme(scheme))

	tests := []struct {
		name               string
		includeAnnotations []string
		excludeAnnotations []string
		includeLabels      []string
		excludeLabels      []string
		secretSanta        *secretsantav1alpha1.SecretSanta
		wantSecret         bool
	}{
		{
			name: "no filters - should process",
			secretSanta: &secretsantav1alpha1.SecretSanta{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "test-secret",
					Namespace:  "default",
					Finalizers: []string{SecretSantaFinalizer},
				},
				Spec: secretsantav1alpha1.SecretSantaSpec{
					SecretType: "Opaque",
					Template:   "data: test",
				},
			},
			wantSecret: true,
		},
		{
			name:               "include annotation present - should process",
			includeAnnotations: []string{"secret-santa.io/managed"},
			secretSanta: &secretsantav1alpha1.SecretSanta{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "test-secret",
					Namespace:  "default",
					Finalizers: []string{SecretSantaFinalizer},
					Annotations: map[string]string{
						"secret-santa.io/managed": "true",
					},
				},
				Spec: secretsantav1alpha1.SecretSantaSpec{
					SecretType: "Opaque",
					Template:   "data: test",
				},
			},
			wantSecret: true,
		},
		{
			name:               "include annotation missing - should skip",
			includeAnnotations: []string{"secret-santa.io/managed"},
			secretSanta: &secretsantav1alpha1.SecretSanta{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "test-secret",
					Namespace:  "default",
					Finalizers: []string{SecretSantaFinalizer},
				},
				Spec: secretsantav1alpha1.SecretSantaSpec{
					SecretType: "Opaque",
					Template:   "data: test",
				},
			},
			wantSecret: false,
		},
		{
			name:               "exclude annotation present - should skip",
			excludeAnnotations: []string{"secret-santa.io/ignore"},
			secretSanta: &secretsantav1alpha1.SecretSanta{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "test-secret",
					Namespace:  "default",
					Finalizers: []string{SecretSantaFinalizer},
					Annotations: map[string]string{
						"secret-santa.io/ignore": "true",
					},
				},
				Spec: secretsantav1alpha1.SecretSantaSpec{
					SecretType: "Opaque",
					Template:   "data: test",
				},
			},
			wantSecret: false,
		},
		{
			name:          "include label present - should process",
			includeLabels: []string{"environment"},
			secretSanta: &secretsantav1alpha1.SecretSanta{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "test-secret",
					Namespace:  "default",
					Finalizers: []string{SecretSantaFinalizer},
					Labels: map[string]string{
						"environment": "production",
					},
				},
				Spec: secretsantav1alpha1.SecretSantaSpec{
					SecretType: "Opaque",
					Template:   "data: test",
				},
			},
			wantSecret: true,
		},
		{
			name:          "exclude label present - should skip",
			excludeLabels: []string{"skip"},
			secretSanta: &secretsantav1alpha1.SecretSanta{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "test-secret",
					Namespace:  "default",
					Finalizers: []string{SecretSantaFinalizer},
					Labels: map[string]string{
						"skip": "true",
					},
				},
				Spec: secretsantav1alpha1.SecretSantaSpec{
					SecretType: "Opaque",
					Template:   "data: test",
				},
			},
			wantSecret: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(tt.secretSanta).
				WithStatusSubresource(&secretsantav1alpha1.SecretSanta{}).
				Build()

			r := &SecretSantaReconciler{
				Client:             client,
				Scheme:             scheme,
				IncludeAnnotations: tt.includeAnnotations,
				ExcludeAnnotations: tt.excludeAnnotations,
				IncludeLabels:      tt.includeLabels,
				ExcludeLabels:      tt.excludeLabels,
			}

			req := ctrl.Request{
				NamespacedName: types.NamespacedName{
					Name:      "test-secret",
					Namespace: "default",
				},
			}

			ctx := log.IntoContext(context.Background(), log.Log)
			result, err := r.Reconcile(ctx, req)

			require.NoError(t, err)
			assert.Equal(t, ctrl.Result{}, result)

			// Check if secret was created based on expectation
			var secret corev1.Secret
			err = client.Get(ctx, req.NamespacedName, &secret)
			if tt.wantSecret {
				assert.NoError(t, err, "Expected secret to be created")
			} else {
				assert.Error(t, err, "Expected secret NOT to be created")
				assert.True(t, errors.IsNotFound(err), "Expected NotFound error")
			}
		})
	}
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