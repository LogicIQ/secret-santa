package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

//+kubebuilder:object:generate=true

// DryRunResult contains masked template execution results
type DryRunResult struct {
	MaskedOutput   string       `json:"maskedOutput,omitempty"`
	GeneratorsUsed []string     `json:"generatorsUsed,omitempty"`
	ExecutionTime  *metav1.Time `json:"executionTime,omitempty"`
}

//+kubebuilder:object:generate=true

// MediaConfig defines configuration for secret storage destinations
type MediaConfig struct {
	// Type specifies the storage backend
	// Supported types: k8s, aws-secrets-manager, aws-parameter-store, azure-key-vault, gcp-secret-manager
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:Enum=k8s;aws-secrets-manager;aws-parameter-store;azure-key-vault;gcp-secret-manager
	Type string `json:"type"`
	// Config contains storage backend specific configuration parameters
	Config *runtime.RawExtension `json:"config,omitempty"`
}

//+kubebuilder:object:generate=true

// GeneratorConfig defines configuration for secret generators
type GeneratorConfig struct {
	// Name is the unique identifier for this generator within the template
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`
	// Type specifies the generator type (e.g., random_password, tls_private_key)
	// Supported types: random_password, random_string, random_uuid, random_bytes,
	// random_integer, random_id, tls_private_key, tls_self_signed_cert,
	// tls_cert_request, tls_locally_signed_cert, crypto_aes_key, crypto_rsa_key,
	// crypto_ed25519_key, crypto_hmac
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:Enum=random_password;random_string;random_uuid;random_bytes;random_integer;random_id;tls_private_key;tls_self_signed_cert;tls_cert_request;tls_locally_signed_cert;crypto_aes_key;crypto_rsa_key;crypto_ed25519_key;crypto_hmac
	Type string `json:"type"`
	// Config contains generator-specific configuration parameters
	Config *runtime.RawExtension `json:"config,omitempty"`
}

//+kubebuilder:object:generate=true

// SecretSantaSpec defines the desired state of SecretSanta
type SecretSantaSpec struct {
	// Template is the Go template string for generating secret data
	// +kubebuilder:validation:MinLength=1
	Template string `json:"template"`
	// Generators define the secret value generators used in the template
	// +kubebuilder:validation:MinItems=1
	Generators []GeneratorConfig `json:"generators"`
	// Media specifies where to store the generated secret (defaults to Kubernetes)
	Media *MediaConfig `json:"media,omitempty"`
	// SecretName overrides the default secret name (defaults to CR name)
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=253
	// +kubebuilder:validation:Pattern=`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`
	SecretName string `json:"secretName,omitempty"`
	// SecretType sets the Kubernetes secret type
	// +kubebuilder:default="Opaque"
	SecretType string `json:"secretType,omitempty"`
	// Labels to apply to the generated secret
	Labels map[string]string `json:"labels,omitempty"`
	// Annotations to apply to the generated secret
	Annotations map[string]string `json:"annotations,omitempty"`
	// DryRun enables validation mode without creating actual secrets
	// +kubebuilder:default=false
	DryRun bool `json:"dryRun,omitempty"`
}

// SecretSantaStatus defines the observed state of SecretSanta
type SecretSantaStatus struct {
	// LastGenerated timestamp of the last successful secret generation
	LastGenerated *metav1.Time `json:"lastGenerated,omitempty"`
	// Conditions represent the current state of the SecretSanta resource
	Conditions []metav1.Condition `json:"conditions,omitempty"`
	// DryRunResult contains the masked output from dry-run executions
	DryRunResult *DryRunResult `json:"dryRunResult,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:shortName=ss

// SecretSanta is the Schema for the secretsantas API
type SecretSanta struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SecretSantaSpec   `json:"spec,omitempty"`
	Status SecretSantaStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// SecretSantaList contains a list of SecretSanta
type SecretSantaList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SecretSanta `json:"items"`
}
