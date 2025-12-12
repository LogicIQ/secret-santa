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
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

func TestE2ECryptoGenerators(t *testing.T) {
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
	name := "e2e-crypto-test"

	secretSanta := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "secrets.secret-santa.io/v1alpha1",
			"kind":       "SecretSanta",
			"metadata": map[string]interface{}{
				"name":      name,
				"namespace": namespace,
			},
			"spec": map[string]interface{}{
				"template": `aes_key: {{ .AESKey.value }}
aes_key_b64: {{ .AESKey.value | b64enc }}
hmac_key: {{ .HMACKey.value }}
hmac_hash: {{ .HMACKey.value | sha256 }}
rsa_key: {{ .RSAKey.value }}
rsa_key_length: {{ len .RSAKey.value }}`,
				"generators": []interface{}{
					map[string]interface{}{
						"name": "AESKey",
						"type": "crypto_aes_key",
						"config": map[string]interface{}{
							"key_size": float64(256),
						},
					},
					map[string]interface{}{
						"name": "HMACKey",
						"type": "crypto_hmac",
						"config": map[string]interface{}{
							"algorithm": "SHA256",
							"key_size":  float64(32),
						},
					},
					map[string]interface{}{
						"name": "RSAKey",
						"type": "crypto_rsa_key",
						"config": map[string]interface{}{
							"key_size": float64(2048),
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
	expectedFields := []string{"aes_key:", "aes_key_b64:", "hmac_key:", "hmac_hash:", "rsa_key:", "rsa_key_length:"}
	for _, field := range expectedFields {
		if !strings.Contains(data, field) {
			t.Errorf("Field %s not found in secret data", field)
		}
	}
}

func TestE2ETLSGenerators(t *testing.T) {
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
	name := "e2e-tls-test"

	secretSanta := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "secrets.secret-santa.io/v1alpha1",
			"kind":       "SecretSanta",
			"metadata": map[string]interface{}{
				"name":      name,
				"namespace": namespace,
			},
			"spec": map[string]interface{}{
				"template": `tls.key: {{ .TLSKey.value }}
tls.crt: {{ .TLSCert.value }}
key_algorithm: {{ .TLSKey.algorithm }}
cert_subject: {{ .TLSCert.subject }}`,
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

	if secret.Type != corev1.SecretTypeTLS {
		t.Errorf("Expected secret type kubernetes.io/tls, got %s", secret.Type)
	}

	data := string(secret.Data["data"])
	expectedFields := []string{"tls.key:", "tls.crt:", "key_algorithm:", "cert_subject:"}
	for _, field := range expectedFields {
		if !strings.Contains(data, field) {
			t.Errorf("Field %s not found in secret data", field)
		}
	}
}

func TestE2ETimeGenerator(t *testing.T) {
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
	name := "e2e-time-test"

	secretSanta := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "secrets.secret-santa.io/v1alpha1",
			"kind":       "SecretSanta",
			"metadata": map[string]interface{}{
				"name":      name,
				"namespace": namespace,
			},
			"spec": map[string]interface{}{
				"template": `timestamp: {{ .Timestamp.value }}
timezone: {{ .Timestamp.timezone }}
epoch: {{ .Timestamp.epoch }}
iso8601: {{ .Timestamp.iso8601 }}
formatted: {{ .Timestamp.formatted }}`,
				"generators": []interface{}{
					map[string]interface{}{
						"name": "Timestamp",
						"type": "time_static",
						"config": map[string]interface{}{
							"timezone": "UTC",
							"format":   "2006-01-02 15:04:05",
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
	expectedFields := []string{"timestamp:", "timezone:", "epoch:", "iso8601:", "formatted:"}
	for _, field := range expectedFields {
		if !strings.Contains(data, field) {
			t.Errorf("Field %s not found in secret data", field)
		}
	}
}

func TestE2ETemplateFunctions(t *testing.T) {
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
	name := "e2e-template-functions-test"

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
password_bcrypt: {{ .Password.value | bcrypt }}
password_sha256: {{ .Password.value | sha256 }}
password_length: {{ len .Password.value }}
password_entropy: {{ entropy .Password.value .Password.charset | printf "%.2f" }}
password_crc32: {{ .Password.value | crc32 }}
api_key: {{ .APIKey.value }}
api_key_b64: {{ .APIKey.value | b64enc }}
api_key_url_safe: {{ .APIKey.value | urlSafeB64 }}
api_key_upper: {{ .APIKey.value | upper }}
uuid: {{ .UUID.value }}
uuid_compact: {{ .UUID.value | compact }}
port: {{ .Port.value }}
port_hex: {{ .Port.value | toHex }}
port_binary: {{ .Port.value | toBinary }}
{{- $entropy := entropy .Password.value .Password.charset }}
{{- if gt $entropy 60.0 }}
security_level: "high"
{{- else }}
security_level: "medium"
{{- end }}`,
				"generators": []interface{}{
					map[string]interface{}{
						"name": "Password",
						"type": "random_password",
						"config": map[string]interface{}{
							"length": float64(24),
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
	templateFunctions := []string{
		"password_bcrypt:", "password_sha256:", "password_length:", "password_entropy:", "password_crc32:",
		"api_key_b64:", "api_key_url_safe:", "api_key_upper:",
		"uuid_compact:",
		"port_hex:", "port_binary:",
		"security_level:",
	}
	
	for _, field := range templateFunctions {
		if !strings.Contains(data, field) {
			t.Errorf("Template function result %s not found in secret data", field)
		}
	}
}