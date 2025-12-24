package gcp

import (
	"context"
	"fmt"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"google.golang.org/api/option"

	secretsantav1alpha1 "github.com/logicIQ/secret-santa/api/v1alpha1"
)

// GCPSecretManagerMedia stores secrets in GCP Secret Manager
type GCPSecretManagerMedia struct {
	ProjectID       string
	SecretName      string
	CredentialsFile string
}

func (m *GCPSecretManagerMedia) Store(ctx context.Context, secretSanta *secretsantav1alpha1.SecretSanta, data string) error {
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

	// Add labels from SecretSanta labels and annotations
	labels := make(map[string]string)
	for k, v := range secretSanta.Spec.Labels {
		labels[k] = v
	}
	for k, v := range secretSanta.Spec.Annotations {
		labels[k] = v
	}
	if len(labels) > 0 {
		createReq.Secret.Labels = labels
	}

	secret, err := client.CreateSecret(ctx, createReq)
	if err != nil {
		return fmt.Errorf("failed to create secret: %w", err)
	}

	// Add secret version with the data
	addVersionReq := &secretmanagerpb.AddSecretVersionRequest{
		Parent: secret.Name,
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