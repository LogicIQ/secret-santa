package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GeneratorConfig defines configuration for secret generators
type GeneratorConfig struct {
	Name   string                 `json:"name"`
	Type   string                 `json:"type"`
	Config map[string]interface{} `json:"config,omitempty"`
}

// SecretSantaSpec defines the desired state of SecretSanta
type SecretSantaSpec struct {
	Template    string            `json:"template"`
	Generators  []GeneratorConfig `json:"generators"`
	SecretType  string            `json:"secretType,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

// SecretSantaStatus defines the observed state of SecretSanta
type SecretSantaStatus struct {
	LastGenerated *metav1.Time       `json:"lastGenerated,omitempty"`
	Conditions    []metav1.Condition `json:"conditions,omitempty"`
}

// SecretSanta is the Schema for the secretsantas API
type SecretSanta struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              SecretSantaSpec   `json:"spec,omitempty"`
	Status            SecretSantaStatus `json:"status,omitempty"`
}

// SecretSantaList contains a list of SecretSanta
type SecretSantaList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SecretSanta `json:"items"`
}
