package aws

import (
	"context"
	"crypto/sha256"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager/types" // Fixed import
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	ssm_types "github.com/aws/aws-sdk-go-v2/service/ssm/types"

	secretsantav1alpha1 "github.com/logicIQ/secret-santa/api/v1alpha1"
)

// AWSSecretsManagerMedia stores secrets in AWS Secrets Manager
type AWSSecretsManagerMedia struct {
	Region    string
	SecretName string
	KMSKeyId  string
}

func (m *AWSSecretsManagerMedia) Store(ctx context.Context, secretSanta *secretsantav1alpha1.SecretSanta, data string, enableMetadata bool) error {
	cfg, err := m.loadAWSConfig()
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := secretsmanager.NewFromConfig(cfg)

	secretName := m.SecretName
	if secretName == "" {
		secretName = secretSanta.Spec.SecretName
		if secretName == "" {
			secretName = secretSanta.Name
		}
	}

	// Add namespace prefix to avoid conflicts
	secretName = fmt.Sprintf("%s/%s", secretSanta.Namespace, secretName)

	input := &secretsmanager.CreateSecretInput{
		Name:         aws.String(secretName),
		SecretString: aws.String(data),
	}

	// Add KMS key if specified
	if m.KMSKeyId != "" {
		input.KmsKeyId = aws.String(m.KMSKeyId)
	}

	// Add tags from labels, annotations, and metadata
	var tags []types.Tag
	for k, v := range secretSanta.Spec.Labels {
		tags = append(tags, types.Tag{
			Key:   aws.String(k),
			Value: aws.String(v),
		})
	}
	for k, v := range secretSanta.Spec.Annotations {
		tags = append(tags, types.Tag{
			Key:   aws.String(k),
			Value: aws.String(v),
		})
	}
	
	// Add metadata tags only if enabled
	if enableMetadata {
		tags = append(tags, 
			types.Tag{Key: aws.String("secrets.secret-santa.io/created-at"), Value: aws.String(time.Now().UTC().Format(time.RFC3339))},
			types.Tag{Key: aws.String("secrets.secret-santa.io/generator-types"), Value: aws.String(m.getGeneratorTypes(secretSanta.Spec.Generators))},
			types.Tag{Key: aws.String("secrets.secret-santa.io/template-checksum"), Value: aws.String(m.calculateTemplateChecksum(secretSanta.Spec.Template))},
			types.Tag{Key: aws.String("secrets.secret-santa.io/source-cr"), Value: aws.String(fmt.Sprintf("%s/%s", secretSanta.Namespace, secretSanta.Name))},
		)
	}
	
	if len(tags) > 0 {
		input.Tags = tags
	}

	_, err = client.CreateSecret(ctx, input)
	return err
}

func (m *AWSSecretsManagerMedia) GetType() string {
	return "aws-secrets-manager"
}

func (m *AWSSecretsManagerMedia) loadAWSConfig() (aws.Config, error) {
	var opts []func(*config.LoadOptions) error
	
	if m.Region != "" {
		opts = append(opts, config.WithRegion(m.Region))
	}

	return config.LoadDefaultConfig(context.TODO(), opts...)
}

// getGeneratorTypes extracts generator types from the configuration
func (m *AWSSecretsManagerMedia) getGeneratorTypes(generators []secretsantav1alpha1.GeneratorConfig) string {
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
func (m *AWSSecretsManagerMedia) calculateTemplateChecksum(template string) string {
	hash := sha256.Sum256([]byte(template))
	return fmt.Sprintf("%x", hash)[:16] // Use first 16 chars for brevity
}

// AWSParameterStoreMedia stores secrets in AWS Systems Manager Parameter Store
type AWSParameterStoreMedia struct {
	Region       string
	ParameterName string
	KMSKeyId     string
}

func (m *AWSParameterStoreMedia) Store(ctx context.Context, secretSanta *secretsantav1alpha1.SecretSanta, data string, enableMetadata bool) error {
	cfg, err := m.loadAWSConfig()
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := ssm.NewFromConfig(cfg)

	paramName := m.ParameterName
	if paramName == "" {
		paramName = secretSanta.Spec.SecretName
		if paramName == "" {
			paramName = secretSanta.Name
		}
	}

	// Add namespace prefix to avoid conflicts
	paramName = fmt.Sprintf("/%s/%s", secretSanta.Namespace, paramName)

	// Create tags from labels, annotations, and metadata
	var tags []ssm_types.Tag
	for k, v := range secretSanta.Spec.Labels {
		tags = append(tags, ssm_types.Tag{
			Key:   aws.String(k),
			Value: aws.String(v),
		})
	}
	for k, v := range secretSanta.Spec.Annotations {
		tags = append(tags, ssm_types.Tag{
			Key:   aws.String(k),
			Value: aws.String(v),
		})
	}
	
	// Add metadata tags only if enabled
	if enableMetadata {
		tags = append(tags,
			ssm_types.Tag{Key: aws.String("secrets.secret-santa.io/created-at"), Value: aws.String(time.Now().UTC().Format(time.RFC3339))},
			ssm_types.Tag{Key: aws.String("secrets.secret-santa.io/generator-types"), Value: aws.String(m.getGeneratorTypes(secretSanta.Spec.Generators))},
			ssm_types.Tag{Key: aws.String("secrets.secret-santa.io/template-checksum"), Value: aws.String(m.calculateTemplateChecksum(secretSanta.Spec.Template))},
			ssm_types.Tag{Key: aws.String("secrets.secret-santa.io/source-cr"), Value: aws.String(fmt.Sprintf("%s/%s", secretSanta.Namespace, secretSanta.Name))},
		)
	}

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
	return err
}

func (m *AWSParameterStoreMedia) GetType() string {
	return "aws-parameter-store"
}

func (m *AWSParameterStoreMedia) loadAWSConfig() (aws.Config, error) {
	var opts []func(*config.LoadOptions) error
	
	if m.Region != "" {
		opts = append(opts, config.WithRegion(m.Region))
	}

	return config.LoadDefaultConfig(context.TODO(), opts...)
}

// getGeneratorTypes extracts generator types from the configuration
func (m *AWSParameterStoreMedia) getGeneratorTypes(generators []secretsantav1alpha1.GeneratorConfig) string {
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
func (m *AWSParameterStoreMedia) calculateTemplateChecksum(template string) string {
	hash := sha256.Sum256([]byte(template))
	return fmt.Sprintf("%x", hash)[:16] // Use first 16 chars for brevity
}
