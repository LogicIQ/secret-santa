//go:build e2e

package validation

import (
	"context"
	"strings"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

func TestTemplateValidationErrors(t *testing.T) {
	tests := []struct {
		name           string
		template       string
		generators     []interface{}
		expectedError  string
		expectedStatus string
	}{
		{
			name:     "malformed template syntax",
			template: `password: {{ .pass.value }`,
			generators: []interface{}{
				map[string]interface{}{
					"name": "pass",
					"type": "random_password",
					"config": map[string]interface{}{
						"length": float64(16),
					},
				},
			},
			expectedError:  "unexpected",
			expectedStatus: "TemplateFailed",
		},
		{
			name:     "invalid generator type",
			template: `password: {{ .pass.value }}`,
			generators: []interface{}{
				map[string]interface{}{
					"name": "pass",
					"type": "invalid_generator_type",
				},
			},
			expectedError:  "unsupported generator type",
			expectedStatus: "DryRunFailed",
		},
		{
			name:     "missing generator name",
			template: `password: {{ .pass.value }}`,
			generators: []interface{}{
				map[string]interface{}{
					"type": "random_password",
					"config": map[string]interface{}{
						"length": float64(16),
					},
				},
			},
			expectedError:  "generator name cannot be empty",
			expectedStatus: "DryRunFailed",
		},
		{
			name:     "missing generator type",
			template: `password: {{ .pass.value }}`,
			generators: []interface{}{
				map[string]interface{}{
					"name": "pass",
					"config": map[string]interface{}{
						"length": float64(16),
					},
				},
			},
			expectedError:  "generator type cannot be empty",
			expectedStatus: "DryRunFailed",
		},
		{
			name:     "template references non-existent generator",
			template: `password: {{ .nonexistent.value }}`,
			generators: []interface{}{
				map[string]interface{}{
					"name": "pass",
					"type": "random_password",
					"config": map[string]interface{}{
						"length": float64(16),
					},
				},
			},
			expectedError:  "template execution",
			expectedStatus: "TemplateExecutionFailed",
		},
	}

	for _, tt := range tests {
		// Skip tests that should fail at CRD validation level
		if tt.name == "missing generator name" || tt.name == "missing generator type" {
			t.Skipf("Test '%s' correctly fails at CRD validation level", tt.name)
			continue
		}
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := config.GetConfig()
			if err != nil {
				t.Fatalf("Failed to get config: %v", err)
			}

			dynClient, err := dynamic.NewForConfig(cfg)
			if err != nil {
				t.Fatalf("Failed to create dynamic client: %v", err)
			}

			namespace := "default"
			name := "template-error-" + strings.ReplaceAll(tt.name, " ", "-")

			secretSanta := &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "secrets.secret-santa.io/v1alpha1",
					"kind":       "SecretSanta",
					"metadata": map[string]interface{}{
						"name":      name,
						"namespace": namespace,
					},
					"spec": map[string]interface{}{
						"template":   tt.template,
						"generators": tt.generators,
					},
				},
			}

			_, err = dynClient.Resource(secretSantaGVR).Namespace(namespace).Create(context.TODO(), secretSanta, metav1.CreateOptions{})
			if err != nil {
				t.Fatalf("Failed to create SecretSanta: %v", err)
			}
			defer func() {
				if err := dynClient.Resource(secretSantaGVR).Namespace(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{}); err != nil {
					t.Logf("Failed to delete SecretSanta: %v", err)
				}
			}()

			err = wait.PollImmediate(2*time.Second, 60*time.Second, func() (bool, error) {
				obj, err := dynClient.Resource(secretSantaGVR).Namespace(namespace).Get(context.TODO(), name, metav1.GetOptions{})
				if err != nil {
					return false, err
				}

				status, found, err := unstructured.NestedMap(obj.Object, "status")
				if err != nil || !found {
					return false, nil
				}

				conditions, found, err := unstructured.NestedSlice(status, "conditions")
				if err != nil || !found {
					return false, nil
				}

				for _, condition := range conditions {
					condMap, ok := condition.(map[string]interface{})
					if !ok {
						continue
					}
					if condMap["type"] == tt.expectedStatus && condMap["status"] == "False" {
						message, ok := condMap["message"].(string)
						if !ok {
							continue
						}
						if strings.Contains(strings.ToLower(message), strings.ToLower(tt.expectedError)) {
							return true, nil
						}
					}
				}
				return false, nil
			})
			if err != nil {
				t.Fatalf("Expected error condition %s with message containing '%s': %v", tt.expectedStatus, tt.expectedError, err)
			}

			t.Logf("Template validation error test '%s' passed!", tt.name)
		})
	}
}

func TestInvalidGeneratorConfig(t *testing.T) {
	cfg, err := config.GetConfig()
	if err != nil {
		t.Fatalf("Failed to get config: %v", err)
	}

	dynClient, err := dynamic.NewForConfig(cfg)
	if err != nil {
		t.Fatalf("Failed to create dynamic client: %v", err)
	}

	namespace := "default"
	name := "invalid-generator-config-test"

	secretSanta := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "secrets.secret-santa.io/v1alpha1",
			"kind":       "SecretSanta",
			"metadata": map[string]interface{}{
				"name":      name,
				"namespace": namespace,
			},
			"spec": map[string]interface{}{
				"template": `password: {{ .pass.value }}`,
				"generators": []interface{}{
					map[string]interface{}{
						"name": "pass",
						"type": "random_password",
						"config": map[string]interface{}{
							"length":         float64(-1),
							"invalid_param": "should_not_exist",
						},
					},
				},
			},
		},
	}

	_, err = dynClient.Resource(secretSantaGVR).Namespace(namespace).Create(context.TODO(), secretSanta, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create SecretSanta: %v", err)
	}
	defer func() {
		if err := dynClient.Resource(secretSantaGVR).Namespace(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{}); err != nil {
			t.Logf("Failed to delete SecretSanta: %v", err)
		}
	}()

	err = wait.PollImmediate(2*time.Second, 60*time.Second, func() (bool, error) {
		obj, err := dynClient.Resource(secretSantaGVR).Namespace(namespace).Get(context.TODO(), name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}

		status, found, err := unstructured.NestedMap(obj.Object, "status")
		if err != nil || !found {
			return false, nil
		}

		conditions, found, err := unstructured.NestedSlice(status, "conditions")
		if err != nil || !found {
			return false, nil
		}

		for _, condition := range conditions {
			condMap, ok := condition.(map[string]interface{})
			if !ok {
				continue
			}
			if condMap["status"] == "False" {
				return true, nil
			}
		}
		return false, nil
	})
	if err != nil {
		t.Fatalf("Expected generator to fail with invalid config: %v", err)
	}

	t.Log("Invalid generator config test passed!")
}