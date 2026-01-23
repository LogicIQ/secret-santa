//go:build e2e

package integration

import (
	"context"
	"strings"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
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

	pollInterval = 2 * time.Second
	pollTimeout  = 60 * time.Second
)

func TestBasicPasswordGeneration(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

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
	name := "basic-password-test"

	secretSanta := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "secrets.secret-santa.io/v1alpha1",
			"kind":       "SecretSanta",
			"metadata": map[string]interface{}{
				"name":      name,
				"namespace": namespace,
			},
			"spec": map[string]interface{}{
				"template": `password: {{ .Password.value }}`,
				"generators": []interface{}{
					map[string]interface{}{
						"name": "Password",
						"type": "random_password",
						"config": map[string]interface{}{
							"length": float64(16),
						},
					},
				},
				"secretType": "Opaque",
			},
		},
	}

	_, err = dynClient.Resource(secretSantaGVR).Namespace(namespace).Create(ctx, secretSanta, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create SecretSanta: %v", err)
	}
	t.Cleanup(func() {
		delCtx, delCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer delCancel()
		if delErr := dynClient.Resource(secretSantaGVR).Namespace(namespace).Delete(delCtx, name, metav1.DeleteOptions{}); delErr != nil {
			t.Logf("Failed to delete SecretSanta: %v", delErr)
		}
	})

	err = wait.PollImmediate(pollInterval, pollTimeout, func() (bool, error) {
		_, err := client.CoreV1().Secrets(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			t.Logf("Waiting for secret to be created: %v", err)
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		t.Fatalf("Secret was not created: %v", err)
	}

	secret, err := client.CoreV1().Secrets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Failed to get secret: %v", err)
	}

	if secret.Type != corev1.SecretTypeOpaque {
		t.Errorf("Expected secret type Opaque, got %s", secret.Type)
	}

	data, exists := secret.Data["data"]
	if !exists {
		t.Fatalf("Secret data key 'data' not found")
	}
	dataStr := string(data)
	if !strings.Contains(dataStr, "password:") {
		t.Errorf("Password not found in secret data")
	}

	t.Logf("Test passed! Secret created successfully with expected fields")
}