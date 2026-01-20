//go:build e2e

package k8s

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

func TestE2ERandomPassword(t *testing.T) {
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
	name := "e2e-password-test"

	// Create SecretSanta CR with template functions
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

	_, err = dynClient.Resource(secretSantaGVR).Namespace(namespace).Create(context.TODO(), secretSanta, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create SecretSanta: %v", err)
	}
	defer func() {
		if delErr := dynClient.Resource(secretSantaGVR).Namespace(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{}); delErr != nil {
			t.Logf("Failed to delete SecretSanta: %v", delErr)
		}
	}()

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

	data := string(secret.Data["data"])
	if len(data) == 0 {
		t.Error("Secret data is empty")
	}

	// Verify template functions worked
	if !strings.Contains(data, "password:") {
		t.Error("Password not found in secret data")
	}
	if !strings.Contains(data, "bcrypt_hash:") {
		t.Error("Bcrypt hash not found in secret data")
	}
	if !strings.Contains(data, "length: 32") {
		t.Error("Length not computed correctly")
	}
	if !strings.Contains(data, "entropy:") {
		t.Error("Entropy not computed")
	}
	if !strings.Contains(data, "sha256:") {
		t.Error("SHA256 hash not computed")
	}
}

func TestE2EMultipleGenerators(t *testing.T) {
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
	name := "e2e-multi-test"

	// Create SecretSanta CR with multiple generators
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
api_key: {{ .APIKey.value | b64enc }}
uuid: {{ .UUID.value }}
uuid_compact: {{ .UUID.value | compact }}
port: {{ .Port.value }}
port_hex: {{ .Port.value | toHex }}`,
				"generators": []interface{}{
					map[string]interface{}{
						"name": "Password",
						"type": "random_password",
						"config": map[string]interface{}{
							"length": float64(16),
						},
					},
					map[string]interface{}{
						"name": "APIKey",
						"type": "random_string",
						"config": map[string]interface{}{
							"length":  float64(32),
							"special": false,
						},
					},
					map[string]interface{}{
						"name": "UUID",
						"type": "random_uuid",
					},
					map[string]interface{}{
						"name": "Port",
						"type": "random_integer",
						"config": map[string]interface{}{
							"min": float64(8000),
							"max": float64(9000),
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
	defer func() {
		if delErr := dynClient.Resource(secretSantaGVR).Namespace(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{}); delErr != nil {
			t.Logf("Failed to delete SecretSanta: %v", delErr)
		}
	}()

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

	data := string(secret.Data["data"])

	// Verify all generators worked
	expectedFields := []string{"password:", "api_key:", "uuid:", "uuid_compact:", "port:", "port_hex:"}
	for _, field := range expectedFields {
		if !strings.Contains(data, field) {
			t.Errorf("Field %s not found in secret data", field)
		}
	}
}
