package k8s

import (
	"context"
	"crypto/sha256"
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
		// For TLS secrets, parse the template output to extract tls.crt and tls.key
		lines := strings.Split(data, "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "tls.crt:") {
				stringData["tls.crt"] = strings.TrimSpace(strings.TrimPrefix(line, "tls.crt:"))
			} else if strings.HasPrefix(line, "tls.key:") {
				stringData["tls.key"] = strings.TrimSpace(strings.TrimPrefix(line, "tls.key:"))
			}
		}
		// If we don't have the required fields, fall back to data field
		if stringData["tls.crt"] == "" || stringData["tls.key"] == "" {
			stringData = map[string]string{"data": data}
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