//go:build e2e

package k8s

import (
	"context"
	"encoding/base64"
	"encoding/pem"
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

func TestTLSCertificateGenerator(t *testing.T) {
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
	name := "test-tls-cert"

	// Create SecretSanta CR for TLS certificate
	secretSanta := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "secrets.secret-santa.io/v1alpha1",
			"kind":       "SecretSanta",
			"metadata": map[string]interface{}{
				"name":      name,
				"namespace": namespace,
			},
			"spec": map[string]interface{}{
				"template": "tls.crt: {{ .TLSCert.cert_pem | b64enc }}\ntls.key: {{ .PrivateKey.private_key_pem | b64enc }}",
				"generators": []interface{}{
					map[string]interface{}{
						"name": "PrivateKey",
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
							"common_name": "test.example.com",
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

	if secret.Type != corev1.SecretTypeTLS {
		t.Errorf("Expected secret type kubernetes.io/tls, got %s", secret.Type)
	}

	// Verify TLS certificate format
	certData, ok := secret.Data["tls.crt"]
	if !ok {
		t.Error("Missing tls.crt in secret")
	}

	keyData, ok := secret.Data["tls.key"]
	if !ok {
		t.Error("Missing tls.key in secret")
	}

	// Decode and validate PEM format
	certDecoded, err := base64.StdEncoding.DecodeString(string(certData))
	if err != nil {
		t.Fatalf("Failed to decode certificate: %v", err)
	}

	block, _ := pem.Decode(certDecoded)
	if block == nil || block.Type != "CERTIFICATE" {
		t.Error("Invalid certificate PEM format")
	}

	keyDecoded, err := base64.StdEncoding.DecodeString(string(keyData))
	if err != nil {
		t.Fatalf("Failed to decode private key: %v", err)
	}

	keyBlock, _ := pem.Decode(keyDecoded)
	if keyBlock == nil || keyBlock.Type != "RSA PRIVATE KEY" {
		t.Error("Invalid private key PEM format")
	}

	// Cleanup
	dynClient.Resource(secretSantaGVR).Namespace(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
}