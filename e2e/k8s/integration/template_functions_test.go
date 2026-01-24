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
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

func TestTemplateFunctions(t *testing.T) {
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
	name := "template-functions-test"

	secretSanta := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "secrets.secret-santa.io/v1alpha1",
			"kind":       "SecretSanta",
			"metadata": map[string]interface{}{
				"name":      name,
				"namespace": namespace,
			},
			"spec": map[string]interface{}{
				"template": `password: {{ .Password.value }}
bcrypt_hash: {{ .Password.value | bcrypt }}
length: {{ len .Password.value }}
entropy: {{ entropy .Password.value .Password.charset | printf "%.2f" }}
sha256: {{ .Password.value | sha256 }}`,
				"generators": []interface{}{
					map[string]interface{}{
						"name": "Password",
						"type": "random_password",
						"config": map[string]interface{}{
							"length": float64(32),
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
	defer func() {
		delCtx, delCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer delCancel()
		dynClient.Resource(secretSantaGVR).Namespace(namespace).Delete(delCtx, name, metav1.DeleteOptions{})
	}()

	err = wait.PollUntilContextTimeout(ctx, 2*time.Second, 60*time.Second, true, func(ctx context.Context) (bool, error) {
		_, err := client.CoreV1().Secrets(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return false, nil // Continue polling on error
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
		t.Fatal("data key not found in secret")
	}
	dataStr := string(data)
	if len(dataStr) == 0 {
		t.Error("Secret data is empty")
	}

	if !strings.Contains(dataStr, "password:") {
		t.Error("Password not found in secret data")
	}
	if !strings.Contains(dataStr, "bcrypt_hash:") {
		t.Error("Bcrypt hash not found in secret data")
	}
	if !strings.Contains(dataStr, "length: 32") {
		t.Error("Length not computed correctly")
	}
	if !strings.Contains(dataStr, "entropy:") {
		t.Error("Entropy not computed")
	}
	if !strings.Contains(dataStr, "sha256:") {
		t.Error("SHA256 hash not computed")
	}
}