//go:build e2e

package crypto

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
)

func setupClients(t *testing.T) (kubernetes.Interface, dynamic.Interface) {
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

	return client, dynClient
}

func TestCryptoGenerators(t *testing.T) {
	client, dynClient := setupClients(t)

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

	_, err := dynClient.Resource(secretSantaGVR).Namespace(namespace).Create(context.TODO(), secretSanta, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create SecretSanta: %v", err)
	}
	defer dynClient.Resource(secretSantaGVR).Namespace(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})

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
	if !strings.Contains(secretData, "aes_key:") {
		t.Error("Missing aes_key in secret data")
	}
	if !strings.Contains(secretData, "hmac_key:") {
		t.Error("Missing hmac_key in secret data")
	}
	if !strings.Contains(secretData, "hmac_signature:") {
		t.Error("Missing hmac_signature in secret data")
	}
}

func TestAllCryptoGenerators(t *testing.T) {
	client, dynClient := setupClients(t)

	namespace := "default"
	name := "e2e-all-crypto-test"

	secretSanta := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "secrets.secret-santa.io/v1alpha1",
			"kind":       "SecretSanta",
			"metadata": map[string]interface{}{
				"name":      name,
				"namespace": namespace,
			},
			"spec": map[string]interface{}{
				"template": `# HMAC
hmac_key: {{ .HMAC.value }}
hmac_algorithm: {{ .HMAC.algorithm }}

# AES Key
aes_key: {{ .AESKey.value }}
aes_key_size: {{ .AESKey.key_size }}

# RSA Key
rsa_private_key: {{ .RSAKey.private_key }}
rsa_public_key: {{ .RSAKey.public_key }}

# Ed25519 Key
ed25519_private_key: {{ .Ed25519Key.private_key }}
ed25519_public_key: {{ .Ed25519Key.public_key }}

# ChaCha20 Key
chacha20_key: {{ .ChaCha20Key.value }}

# XChaCha20 Key
xchacha20_key: {{ .XChaCha20Key.value }}

# ECDSA Key
ecdsa_private_key: {{ .ECDSAKey.private_key }}
ecdsa_public_key: {{ .ECDSAKey.public_key }}

# ECDH Key
ecdh_private_key: {{ .ECDHKey.private_key }}
ecdh_public_key: {{ .ECDHKey.public_key }}`,
				"generators": []interface{}{
					map[string]interface{}{
						"name": "HMAC",
						"type": "crypto_hmac",
						"config": map[string]interface{}{
							"algorithm": "sha256",
							"key_size":  float64(32),
						},
					},
					map[string]interface{}{
						"name": "AESKey",
						"type": "crypto_aes_key",
						"config": map[string]interface{}{
							"key_size": float64(256),
						},
					},
					map[string]interface{}{
						"name": "RSAKey",
						"type": "crypto_rsa_key",
						"config": map[string]interface{}{
							"key_size": float64(2048),
						},
					},
					map[string]interface{}{
						"name": "Ed25519Key",
						"type": "crypto_ed25519_key",
					},
					map[string]interface{}{
						"name": "ChaCha20Key",
						"type": "crypto_chacha20_key",
					},
					map[string]interface{}{
						"name": "XChaCha20Key",
						"type": "crypto_xchacha20_key",
					},
					map[string]interface{}{
						"name": "ECDSAKey",
						"type": "crypto_ecdsa_key",
						"config": map[string]interface{}{
							"curve": "P256",
						},
					},
					map[string]interface{}{
						"name": "ECDHKey",
						"type": "crypto_ecdh_key",
						"config": map[string]interface{}{
							"curve": "P256",
						},
					},
				},
				"secretType": "Opaque",
			},
		},
	}

	_, err := dynClient.Resource(secretSantaGVR).Namespace(namespace).Create(context.TODO(), secretSanta, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create SecretSanta: %v", err)
	}
	defer dynClient.Resource(secretSantaGVR).Namespace(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})

	err = wait.PollImmediate(2*time.Second, 60*time.Second, func() (bool, error) {
		_, err := client.CoreV1().Secrets(namespace).Get(context.TODO(), name, metav1.GetOptions{})
		return err == nil, nil
	})
	if err != nil {
		t.Fatalf("Secret was not created: %v", err)
	}

	secret, err := client.CoreV1().Secrets(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Failed to get secret: %v", err)
	}

	data := string(secret.Data["data"])
	expectedFields := []string{
		"hmac_key:", "hmac_algorithm:",
		"aes_key:", "aes_key_size:",
		"rsa_private_key:", "rsa_public_key:",
		"ed25519_private_key:", "ed25519_public_key:",
		"chacha20_key:", "xchacha20_key:",
		"ecdsa_private_key:", "ecdsa_public_key:",
		"ecdh_private_key:", "ecdh_public_key:",
	}
	for _, field := range expectedFields {
		if !strings.Contains(data, field) {
			t.Errorf("Field %s not found in secret data", field)
		}
	}
}
