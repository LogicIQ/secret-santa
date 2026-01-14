package media

import (
	"context"
	secretsantav1alpha1 "github.com/logicIQ/secret-santa/api/v1alpha1"
)

// Media interface for different secret storage destinations
type Media interface {
	Store(ctx context.Context, secretSanta *secretsantav1alpha1.SecretSanta, data string, enableMetadata bool) error
	GetType() string
}

// MediaConfig defines configuration for media destinations
type MediaConfig struct {
	Type   string                 `json:"type"`
	Config map[string]interface{} `json:"config,omitempty"`
}
