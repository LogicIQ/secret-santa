//go:build e2e

package integration

import (
	"context"
	"strings"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

func TestMultipleGenerators(t *testing.T) {
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
	name := "multiple-generators-test"

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
		delErr := dynClient.Resource(secretSantaGVR).Namespace(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
		if delErr != nil {
			t.Logf("Failed to delete SecretSanta: %v", delErr)
		}
	}()

	err = wait.PollImmediate(2*time.Second, 60*time.Second, func() (bool, error) {
		_, getErr := client.CoreV1().Secrets(namespace).Get(context.TODO(), name, metav1.GetOptions{})
		if getErr != nil {
			return false, nil // Continue polling
		}
		return true, nil // Secret found
	})
	if err != nil {
		t.Fatalf("Timeout waiting for secret to be created: %v", err)
	}

	secret, err := client.CoreV1().Secrets(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Failed to get secret: %v", err)
	}

	data := string(secret.Data["data"])

	expectedFields := []string{"password:", "api_key:", "uuid:", "uuid_compact:", "port:", "port_hex:"}
	for _, field := range expectedFields {
		if !strings.Contains(data, field) {
			t.Errorf("Field %s not found in secret data", field)
		}
	}
}