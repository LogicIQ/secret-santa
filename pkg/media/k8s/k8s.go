package k8s

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	secretsantav1alpha1 "github.com/logicIQ/secret-santa/api/v1alpha1"
)

// K8sSecretsMedia stores secrets as Kubernetes secrets
type K8sSecretsMedia struct {
	Client     client.Client
	SecretName string
}

func (m *K8sSecretsMedia) Store(ctx context.Context, secretSanta *secretsantav1alpha1.SecretSanta, data string, enableMetadata bool) error {
	secretName := m.SecretName
	if secretName == "" {
		secretName = secretSanta.Spec.SecretName
		if secretName == "" {
			secretName = secretSanta.Name
		}
	}

	// Handle TLS secrets specially
	stringData := map[string]string{}
	if secretSanta.Spec.SecretType == "kubernetes.io/tls" {
		// Try to parse as JSON first
		var jsonData map[string]string
		if err := json.Unmarshal([]byte(data), &jsonData); err == nil {
			if cert, ok := jsonData["tls.crt"]; ok && cert != "" {
				stringData["tls.crt"] = cert
			}
			if key, ok := jsonData["tls.key"]; ok && key != "" {
				stringData["tls.key"] = key
			}
		} else {
			// Fallback to line-by-line parsing
			lines := strings.Split(data, "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line == "" {
					continue
				}
				if strings.HasPrefix(line, "tls.crt:") {
					certValue := strings.TrimSpace(strings.TrimPrefix(line, "tls.crt:"))
					if certValue != "" {
						stringData["tls.crt"] = certValue
					}
				} else if strings.HasPrefix(line, "tls.key:") {
					keyValue := strings.TrimSpace(strings.TrimPrefix(line, "tls.key:"))
					if keyValue != "" {
						stringData["tls.key"] = keyValue
					}
				}
			}
		}
		// If we don't have the required fields, return error for TLS secrets
		if stringData["tls.crt"] == "" || stringData["tls.key"] == "" {
			return fmt.Errorf("TLS secret requires both tls.crt and tls.key fields")
		}
	} else {
		stringData["data"] = data
	}

	// Merge user annotations with metadata annotations
	annotations := make(map[string]string)
	for k, v := range secretSanta.Spec.Annotations {
		annotations[k] = v
	}

	// Add metadata annotations only if enabled
	if enableMetadata {
		annotations["secrets.secret-santa.io/created-at"] = time.Now().UTC().Format(time.RFC3339)
		annotations["secrets.secret-santa.io/generator-types"] = m.getGeneratorTypes(secretSanta.Spec.Generators)
		annotations["secrets.secret-santa.io/template-checksum"] = m.calculateTemplateChecksum(secretSanta.Spec.Template)
		annotations["secrets.secret-santa.io/source-cr"] = fmt.Sprintf("%s/%s", secretSanta.Namespace, secretSanta.Name)
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:        secretName,
			Namespace:   secretSanta.Namespace,
			Labels:      secretSanta.Spec.Labels,
			Annotations: annotations,
		},
		Type:       corev1.SecretType(secretSanta.Spec.SecretType),
		StringData: stringData,
	}

	err := m.Client.Create(ctx, secret)
	if err != nil {
		// Check if secret already exists
		if client.IgnoreAlreadyExists(err) == nil {
			return nil // Secret already exists, which is fine for create-once policy
		}
		return fmt.Errorf("failed to create secret %s/%s: %w", secretSanta.Namespace, secretName, err)
	}
	return nil
}

func (m *K8sSecretsMedia) GetType() string {
	return "k8s"
}

// getGeneratorTypes extracts generator types from the configuration
func (m *K8sSecretsMedia) getGeneratorTypes(generators []secretsantav1alpha1.GeneratorConfig) string {
	types := make([]string, len(generators))
	for i, gen := range generators {
		types[i] = gen.Type
	}
	return strings.Join(types, ",")
}

// calculateTemplateChecksum creates a SHA256 checksum of the template
func (m *K8sSecretsMedia) calculateTemplateChecksum(template string) string {
	hash := sha256.Sum256([]byte(template))
	return fmt.Sprintf("%x", hash)[:16] // Use first 16 chars for brevity
}
