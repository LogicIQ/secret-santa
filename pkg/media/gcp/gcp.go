package gcp

import (
	"context"
	"crypto/sha256"
	"fmt"
	"strings"
	"time"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"google.golang.org/api/option"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	secretsantav1alpha1 "github.com/logicIQ/secret-santa/api/v1alpha1"
)

// GCPSecretManagerMedia stores secrets in GCP Secret Manager
type GCPSecretManagerMedia struct {
	ProjectID       string
	SecretName      string
	CredentialsFile string
}

func (m *GCPSecretManagerMedia) Store(ctx context.Context, secretSanta *secretsantav1alpha1.SecretSanta, data string, enableMetadata bool) error {
	var opts []option.ClientOption

	// Use credentials file if provided, otherwise rely on workload identity/default credentials
	if m.CredentialsFile != "" {
		opts = append(opts, option.WithCredentialsFile(m.CredentialsFile))
	}

	client, err := secretmanager.NewClient(ctx, opts...)
	if err != nil {
		return fmt.Errorf("failed to create GCP Secret Manager client: %w", err)
	}
	defer client.Close()

	secretName := m.SecretName
	if secretName == "" {
		secretName = secretSanta.Spec.SecretName
		if secretName == "" {
			secretName = secretSanta.Name
		}
	}

	// Create secret name with namespace prefix
	fullSecretName := fmt.Sprintf("%s-%s", secretSanta.Namespace, secretName)
	secretPath := fmt.Sprintf("projects/%s/secrets/%s", m.ProjectID, fullSecretName)

	// Create the secret if it doesn't exist
	createReq := &secretmanagerpb.CreateSecretRequest{
		Parent:   fmt.Sprintf("projects/%s", m.ProjectID),
		SecretId: fullSecretName,
		Secret: &secretmanagerpb.Secret{
			Replication: &secretmanagerpb.Replication{
				Replication: &secretmanagerpb.Replication_Automatic_{
					Automatic: &secretmanagerpb.Replication_Automatic{},
				},
			},
		},
	}

	// Add labels from SecretSanta labels, annotations, and metadata
	labels := make(map[string]string)
	for k, v := range secretSanta.Spec.Labels {
		labels[k] = v
	}
	for k, v := range secretSanta.Spec.Annotations {
		labels[k] = v
	}

	// Add metadata labels only if enabled
	if enableMetadata {
		labels["secrets_secret-santa_io_created-at"] = time.Now().UTC().Format(time.RFC3339)
		labels["secrets_secret-santa_io_generator-types"] = m.getGeneratorTypes(secretSanta.Spec.Generators)
		labels["secrets_secret-santa_io_template-checksum"] = m.calculateTemplateChecksum(secretSanta.Spec.Template)
		labels["secrets_secret-santa_io_source-cr"] = fmt.Sprintf("%s_%s", secretSanta.Namespace, secretSanta.Name)
	}

	if len(labels) > 0 {
		createReq.Secret.Labels = labels
	}

	_, err = client.CreateSecret(ctx, createReq)
	if err != nil {
		if st, ok := status.FromError(err); ok && st.Code() == codes.AlreadyExists {
			// Secret already exists, check if it has versions (create-once policy)
			listReq := &secretmanagerpb.ListSecretVersionsRequest{
				Parent: secretPath,
			}
			versions := client.ListSecretVersions(ctx, listReq)
			if _, vErr := versions.Next(); vErr == nil {
				// Secret already has versions, skip adding new version (create-once)
				return nil
			}
		} else {
			return fmt.Errorf("failed to create secret: %w", err)
		}
	}

	// Add secret version with the data
	addVersionReq := &secretmanagerpb.AddSecretVersionRequest{
		Parent: secretPath,
		Payload: &secretmanagerpb.SecretPayload{
			Data: []byte(data),
		},
	}

	_, err = client.AddSecretVersion(ctx, addVersionReq)
	if err != nil {
		return fmt.Errorf("failed to add secret version: %w", err)
	}

	return nil
}

func (m *GCPSecretManagerMedia) GetType() string {
	return "gcp-secret-manager"
}

// getGeneratorTypes extracts generator types from the configuration
func (m *GCPSecretManagerMedia) getGeneratorTypes(generators []secretsantav1alpha1.GeneratorConfig) string {
	types := make([]string, len(generators))
	for i, gen := range generators {
		types[i] = gen.Type
	}
	return strings.Join(types, ",")
}

// calculateTemplateChecksum creates a SHA256 checksum of the template
func (m *GCPSecretManagerMedia) calculateTemplateChecksum(template string) string {
	hash := sha256.Sum256([]byte(template))
	return fmt.Sprintf("%x", hash)[:16] // Use first 16 chars for brevity
}
