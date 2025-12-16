package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

//+kubebuilder:object:generate=true

// GeneratorConfig defines configuration for secret generators
type GeneratorConfig struct {
	Name   string                `json:"name"`
	Type   string                `json:"type"`
	Config *runtime.RawExtension `json:"config,omitempty"`
}

//+kubebuilder:object:generate=true

// SecretSantaSpec defines the desired state of SecretSanta
type SecretSantaSpec struct {
	Template    string            `json:"template"`
	Generators  []GeneratorConfig `json:"generators"`
	SecretName  string            `json:"secretName,omitempty"`
	SecretType  string            `json:"secretType,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

//+kubebuilder:object:generate=true

// SecretSantaStatus defines the observed state of SecretSanta
type SecretSantaStatus struct {
	LastGenerated *metav1.Time       `json:"lastGenerated,omitempty"`
	Conditions    []metav1.Condition `json:"conditions,omitempty"`
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
