package aws

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	ssm_types "github.com/aws/aws-sdk-go-v2/service/ssm/types"

	secretsantav1alpha1 "github.com/logicIQ/secret-santa/api/v1alpha1"
)

const (
	tagKeyCreatedAt        = "secrets.secret-santa.io/created-at"
	tagKeyGeneratorTypes   = "secrets.secret-santa.io/generator-types"
	tagKeyTemplateChecksum = "secrets.secret-santa.io/template-checksum"
	tagKeySourceCR         = "secrets.secret-santa.io/source-cr"
)

// getGeneratorTypes extracts generator types from the configuration
func getGeneratorTypes(generators []secretsantav1alpha1.GeneratorConfig) string {
	if len(generators) == 0 {
		return ""
	}
	var builder strings.Builder
	for i, gen := range generators {
		if i > 0 {
			builder.WriteByte(',')
		}
		builder.WriteString(gen.Type)
	}
	return builder.String()
}

// calculateTemplateChecksum creates a SHA256 checksum of the template
func calculateTemplateChecksum(template string) string {
	hash := sha256.Sum256([]byte(template))
	return fmt.Sprintf("%x", hash)[:16]
}

// loadAWSConfig loads AWS configuration with optional region override
func loadAWSConfig(ctx context.Context, region string) (aws.Config, error) {
	var opts []func(*config.LoadOptions) error

	if region != "" {
		opts = append(opts, config.WithRegion(region))
	}

	return config.LoadDefaultConfig(ctx, opts...)
}

// resolveSecretName resolves the secret name with namespace prefix
func resolveSecretName(mediaName string, secretSanta *secretsantav1alpha1.SecretSanta, prefix string) string {
	name := mediaName
	if name == "" {
		if secretSanta.Spec.SecretName != "" {
			name = secretSanta.Spec.SecretName
		} else {
			name = secretSanta.Name
		}
	}
	return fmt.Sprintf("%s%s/%s", prefix, secretSanta.Namespace, name)
}

// createSecretsManagerTags creates tags for AWS Secrets Manager
func createSecretsManagerTags(secretSanta *secretsantav1alpha1.SecretSanta, enableMetadata bool) []types.Tag {
	var tags []types.Tag
	for k, v := range secretSanta.Spec.Labels {
		tags = append(tags, types.Tag{Key: aws.String(k), Value: aws.String(v)})
	}
	for k, v := range secretSanta.Spec.Annotations {
		tags = append(tags, types.Tag{Key: aws.String(k), Value: aws.String(v)})
	}
	if enableMetadata {
		tags = append(tags,
			types.Tag{Key: aws.String(tagKeyCreatedAt), Value: aws.String(time.Now().UTC().Format(time.RFC3339))},
			types.Tag{Key: aws.String(tagKeyGeneratorTypes), Value: aws.String(getGeneratorTypes(secretSanta.Spec.Generators))},
			types.Tag{Key: aws.String(tagKeyTemplateChecksum), Value: aws.String(calculateTemplateChecksum(secretSanta.Spec.Template))},
			types.Tag{Key: aws.String(tagKeySourceCR), Value: aws.String(fmt.Sprintf("%s/%s", secretSanta.Namespace, secretSanta.Name))},
		)
	}
	return tags
}

// createSSMTags creates tags for AWS Systems Manager Parameter Store
func createSSMTags(secretSanta *secretsantav1alpha1.SecretSanta, enableMetadata bool) []ssm_types.Tag {
	var tags []ssm_types.Tag
	for k, v := range secretSanta.Spec.Labels {
		tags = append(tags, ssm_types.Tag{Key: aws.String(k), Value: aws.String(v)})
	}
	for k, v := range secretSanta.Spec.Annotations {
		tags = append(tags, ssm_types.Tag{Key: aws.String(k), Value: aws.String(v)})
	}
	if enableMetadata {
		tags = append(tags,
			ssm_types.Tag{Key: aws.String(tagKeyCreatedAt), Value: aws.String(time.Now().UTC().Format(time.RFC3339))},
			ssm_types.Tag{Key: aws.String(tagKeyGeneratorTypes), Value: aws.String(getGeneratorTypes(secretSanta.Spec.Generators))},
			ssm_types.Tag{Key: aws.String(tagKeyTemplateChecksum), Value: aws.String(calculateTemplateChecksum(secretSanta.Spec.Template))},
			ssm_types.Tag{Key: aws.String(tagKeySourceCR), Value: aws.String(fmt.Sprintf("%s/%s", secretSanta.Namespace, secretSanta.Name))},
		)
	}
	return tags
}

// AWSSecretsManagerMedia stores secrets in AWS Secrets Manager
type AWSSecretsManagerMedia struct {
	Region     string
	SecretName string
	KMSKeyId   string
}

func (m *AWSSecretsManagerMedia) Store(ctx context.Context, secretSanta *secretsantav1alpha1.SecretSanta, data string, enableMetadata bool) error {
	cfg, err := loadAWSConfig(ctx, m.Region)
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := secretsmanager.NewFromConfig(cfg)

	secretName := resolveSecretName(m.SecretName, secretSanta, "")

	input := &secretsmanager.CreateSecretInput{
		Name:         aws.String(secretName),
		SecretString: aws.String(data),
	}

	// Add KMS key if specified
	if m.KMSKeyId != "" {
		input.KmsKeyId = aws.String(m.KMSKeyId)
	}

	tags := createSecretsManagerTags(secretSanta, enableMetadata)
	if len(tags) > 0 {
		input.Tags = tags
	}

	_, err = client.CreateSecret(ctx, input)
	if err != nil {
		var resourceExists *types.ResourceExistsException
		if errors.As(err, &resourceExists) {
			return nil
		}
		return err
	}
	return nil
}

func (m *AWSSecretsManagerMedia) GetType() string {
	return "aws-secrets-manager"
}

// AWSParameterStoreMedia stores secrets in AWS Systems Manager Parameter Store
type AWSParameterStoreMedia struct {
	Region        string
	ParameterName string
	KMSKeyId      string
}

func (m *AWSParameterStoreMedia) Store(ctx context.Context, secretSanta *secretsantav1alpha1.SecretSanta, data string, enableMetadata bool) error {
	cfg, err := loadAWSConfig(ctx, m.Region)
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := ssm.NewFromConfig(cfg)

	paramName := resolveSecretName(m.ParameterName, secretSanta, "/")

	tags := createSSMTags(secretSanta, enableMetadata)

	input := &ssm.PutParameterInput{
		Name:  aws.String(paramName),
		Value: aws.String(data),
		Type:  ssm_types.ParameterTypeSecureString,
		Tags:  tags,
	}

	// Add KMS key if specified
	if m.KMSKeyId != "" {
		input.KeyId = aws.String(m.KMSKeyId)
	}

	_, err = client.PutParameter(ctx, input)
	if err != nil {
		var paramExists *ssm_types.ParameterAlreadyExists
		if errors.As(err, &paramExists) {
			return nil
		}
		return err
	}
	return nil
}

func (m *AWSParameterStoreMedia) GetType() string {
	return "aws-parameter-store"
}
