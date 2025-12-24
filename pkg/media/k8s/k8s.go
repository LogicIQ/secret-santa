package k8s

import (
	"context"
	"strings"

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

func (m *K8sSecretsMedia) Store(ctx context.Context, secretSanta *secretsantav1alpha1.SecretSanta, data string) error {
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

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:        secretName,
			Namespace:   secretSanta.Namespace,
			Labels:      secretSanta.Spec.Labels,
			Annotations: secretSanta.Spec.Annotations,
		},
		Type:       corev1.SecretType(secretSanta.Spec.SecretType),
		StringData: stringData,
	}

	return m.Client.Create(ctx, secret)
}

func (m *K8sSecretsMedia) GetType() string {
	return "k8s"
}