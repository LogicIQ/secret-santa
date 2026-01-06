//go:build e2e

package validation

import (
	"context"
	"strings"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

var (
	secretSantaGVR = schema.GroupVersionResource{
		Group:    "secrets.secret-santa.io",
		Version:  "v1alpha1",
		Resource: "secretsanta",
	}
)

func TestDryRunValidation(t *testing.T) {
	cfg, err := config.GetConfig()
	if err != nil {
		t.Fatalf("Failed to get config: %v", err)
	}

	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	dynClient, err := dynamic.NewForConfig(cfg)
	if err != nil {
		t.Fatalf("Failed to create dynamic client: %v", err)
	}

	namespace := "default"
	name := "dry-run-test"

	secretSanta := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "secrets.secret-santa.io/v1alpha1",
			"kind":       "SecretSanta",
			"metadata": map[string]interface{}{
				"name":      name,
				"namespace": namespace,
			},
			"spec": map[string]interface{}{
				"dryRun": true,
				"template": `{
  "database": {
    "password": "{{ .dbpass.value }}",
    "username": "admin"
  },
  "api": {
    "key": "{{ .apikey.value }}",
    "secret": "{{ .apisecret.value }}"
  }
}`,
				"generators": []interface{}{
					map[string]interface{}{
						"name": "dbpass",
						"type": "random_password",
						"config": map[string]interface{}{
							"length": float64(32),
						},
					},
					map[string]interface{}{
						"name": "apikey",
						"type": "random_string",
						"config": map[string]interface{}{
							"length": float64(64),
						},
					},
					map[string]interface{}{
						"name": "apisecret",
						"type": "random_bytes",
						"config": map[string]interface{}{
							"length":   float64(32),
							"encoding": "base64",
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
	defer dynClient.Resource(secretSantaGVR).Namespace(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})

	var finalStatus map[string]interface{}
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
			condMap := condition.(map[string]interface{})
			if condMap["type"] == "DryRunComplete" && condMap["status"] == "True" {
				finalStatus = status
				return true, nil
			}
		}
		return false, nil
	})
	if err != nil {
		t.Fatalf("Dry-run did not complete successfully: %v", err)
	}

	dryRunResult, found, err := unstructured.NestedMap(finalStatus, "dryRunResult")
	if err != nil {
		t.Fatalf("Failed to get dryRunResult: %v", err)
	}
	if !found {
		t.Fatal("DryRunResult not found in status")
	}

	maskedOutput, found, err := unstructured.NestedString(dryRunResult, "maskedOutput")
	if err != nil {
		t.Fatalf("Failed to get maskedOutput: %v", err)
	}
	if !found {
		t.Fatal("MaskedOutput not found in dryRunResult")
	}

	generatorsUsed, found, err := unstructured.NestedStringSlice(dryRunResult, "generatorsUsed")
	if err != nil {
		t.Fatalf("Failed to get generatorsUsed: %v", err)
	}
	if !found {
		t.Fatal("GeneratorsUsed not found in dryRunResult")
	}

	if !strings.Contains(maskedOutput, "MASKED") {
		t.Error("Masked output should contain MASKED values")
	}

	if !strings.Contains(maskedOutput, "database") {
		t.Error("Masked output should preserve structure (database key)")
	}

	if !strings.Contains(maskedOutput, "api") {
		t.Error("Masked output should preserve structure (api key)")
	}

	expectedGenerators := []string{"dbpass (random_password)", "apikey (random_string)", "apisecret (random_bytes)"}
	for _, expected := range expectedGenerators {
		found := false
		for _, actual := range generatorsUsed {
			if actual == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected generator %s not found in generatorsUsed", expected)
		}
	}

	_, err = client.CoreV1().Secrets(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err == nil {
		t.Error("Secret should not be created in dry-run mode")
	}

	t.Logf("Dry-run test passed! Masked output: %s", maskedOutput)
}

func TestDryRunValidationError(t *testing.T) {
	cfg, err := config.GetConfig()
	if err != nil {
		t.Fatalf("Failed to get config: %v", err)
	}

	dynClient, err := dynamic.NewForConfig(cfg)
	if err != nil {
		t.Fatalf("Failed to create dynamic client: %v", err)
	}

	namespace := "default"
	name := "dry-run-error-test"

	secretSanta := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "secrets.secret-santa.io/v1alpha1",
			"kind":       "SecretSanta",
			"metadata": map[string]interface{}{
				"name":      name,
				"namespace": namespace,
			},
			"spec": map[string]interface{}{
				"dryRun": true,
				"template": `{
  "password": "{{ .valid.value }}"
}`,
				"generators": []interface{}{
					map[string]interface{}{
						"name": "valid",
						"type": "random_password",
						"config": map[string]interface{}{
							"length": float64(16),
						},
					},
					map[string]interface{}{
						"name": "invalid",
						"type": "unsupported_type",
					},
				},
			},
		},
	}

	_, err = dynClient.Resource(secretSantaGVR).Namespace(namespace).Create(context.TODO(), secretSanta, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create SecretSanta: %v", err)
	}
	defer dynClient.Resource(secretSantaGVR).Namespace(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})

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
			condMap := condition.(map[string]interface{})
			if condMap["type"] == "DryRunFailed" && condMap["status"] == "False" {
				message := condMap["message"].(string)
				if strings.Contains(message, "unsupported generator type") {
					return true, nil
				}
			}
		}
		return false, nil
	})
	if err != nil {
		t.Fatalf("Expected dry-run to fail with validation error: %v", err)
	}

	t.Log("Dry-run validation error test passed!")
}