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
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:Enum=k8s;aws-secrets-manager;aws-parameter-store;azure-key-vault;gcp-secret-manager
	Type   string                `json:"type"`
	Config *runtime.RawExtension `json:"config,omitempty"`
}

//+kubebuilder:object:generate=true

// GeneratorConfig defines configuration for secret generators
type GeneratorConfig struct {
	// +kubebuilder:validation:MinLength=1
	Name   string                `json:"name"`
	// +kubebuilder:validation:MinLength=1
	Type   string                `json:"type"`
	// Configuration for the generator. Validation is performed by the controller.
	Config *runtime.RawExtension `json:"config,omitempty"`
}

//+kubebuilder:object:generate=true

// SecretSantaSpec defines the desired state of SecretSanta
type SecretSantaSpec struct {
	// +kubebuilder:validation:MinLength=1
	Template    string            `json:"template"`
	// +kubebuilder:validation:MinItems=1
	Generators  []GeneratorConfig `json:"generators"`
	Media       *MediaConfig      `json:"media,omitempty"`
	// SecretName is the name of the Kubernetes secret to create (not the secret value itself)
	SecretName  string            `json:"secretName,omitempty"`
	SecretType  string            `json:"secretType,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
	DryRun      bool              `json:"dryRun,omitempty"`
}

// SecretSantaStatus defines the observed state of SecretSanta
type SecretSantaStatus struct {
	LastGenerated *metav1.Time       `json:"lastGenerated,omitempty"`
	Conditions    []metav1.Condition `json:"conditions,omitempty"`
	DryRunResult  *DryRunResult      `json:"dryRunResult,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:shortName=ss

// SecretSanta is the Schema for the secretsantas API
type SecretSanta struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              SecretSantaSpec   `json:"spec,omitempty"`
	Status            SecretSantaStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// SecretSantaList contains a list of SecretSanta
type SecretSantaList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SecretSanta `json:"items"`
}
