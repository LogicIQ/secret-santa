package k8s

import (
	"context"
	"encoding/base64"
	"encoding/hex"
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

func TestCryptoGenerators(t *testing.T) {
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
	name := "test-crypto-keys"

	// Create SecretSanta CR for crypto generators
	secretSanta := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "secrets.secret-santa.io/v1alpha1",
			"kind":       "SecretSanta",
			"metadata": map[string]interface{}{
				"name":      name,
				"namespace": namespace,
			},
			"spec": map[string]interface{}{
				"template": `aes_key: {{ .AESKey.key_base64 }}
hmac_key: {{ .HMAC.key_base64 }}
hmac_signature: {{ .HMAC.signature_hex }}`,
				"generators": []interface{}{
					map[string]interface{}{
						"name": "AESKey",
						"type": "crypto_aes_key",
						"config": map[string]interface{}{
							"key_size": float64(256),
						},
					},
					map[string]interface{}{
						"name": "HMAC",
						"type": "crypto_hmac",
						"config": map[string]interface{}{
							"algorithm": "sha256",
							"key_size":  float64(32),
							"message":   "test-payload",
						},
					},
				},
				"secretType": "Opaque",
			},
		},
	}

	_, err = dynClient.Resource(secretSantaGVR).Namespace(namespace).Create(context.TODO(), secretSanta, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create SecretSanta: %v", err)
	}

	// Wait for secret to be created
	err = wait.PollImmediate(2*time.Second, 60*time.Second, func() (bool, error) {
		_, err := client.CoreV1().Secrets(namespace).Get(context.TODO(), name, metav1.GetOptions{})
		return err == nil, nil
	})
	if err != nil {
		t.Fatalf("Secret was not created: %v", err)
	}

	// Verify secret content
	secret, err := client.CoreV1().Secrets(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Failed to get secret: %v", err)
	}

	if secret.Type != corev1.SecretTypeOpaque {
		t.Errorf("Expected secret type Opaque, got %s", secret.Type)
	}

	secretData := string(secret.Data["data"])
	if len(secretData) == 0 {
		t.Error("Secret data is empty")
	}

	// Basic validation that the template was processed
	if !contains(secretData, "aes_key:") {
		t.Error("Missing aes_key in secret data")
	}
	if !contains(secretData, "hmac_key:") {
		t.Error("Missing hmac_key in secret data")
	}
	if !contains(secretData, "hmac_signature:") {
		t.Error("Missing hmac_signature in secret data")
	}

	// Cleanup
	dynClient.Resource(secretSantaGVR).Namespace(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsAt(s, substr)))
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}