//go:build e2e

package integration

import (
	"context"
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

func TestTLSCertificate(t *testing.T) {
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
	name := "tls-certificate-test"

	secretSanta := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "secrets.secret-santa.io/v1alpha1",
			"kind":       "SecretSanta",
			"metadata": map[string]interface{}{
				"name":      name,
				"namespace": namespace,
			},
			"spec": map[string]interface{}{
				"template": `tls.key: {{ .TLSKey.private_key_pem }}
tls.crt: {{ .TLSCert.cert_pem }}`,
				"generators": []interface{}{
					map[string]interface{}{
						"name": "TLSKey",
						"type": "tls_private_key",
						"config": map[string]interface{}{
							"algorithm": "RSA",
							"rsa_bits":  float64(2048),
						},
					},
					map[string]interface{}{
						"name": "TLSCert",
						"type": "tls_self_signed_cert",
						"config": map[string]interface{}{
							"key_algorithm": "RSA",
							"rsa_bits":      float64(2048),
							"subject": map[string]interface{}{
								"common_name": "example.com",
							},
							"validity_period_hours": float64(8760),
						},
					},
				},
				"secretType": "kubernetes.io/tls",
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

	if secret.Type != corev1.SecretTypeTLS {
		t.Errorf("Expected secret type kubernetes.io/tls, got %s", secret.Type)
	}

	if _, ok := secret.Data["tls.crt"]; !ok {
		t.Error("Missing tls.crt in TLS secret")
	}
	if _, ok := secret.Data["tls.key"]; !ok {
		t.Error("Missing tls.key in TLS secret")
	}
}