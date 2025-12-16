//go:build test

package controller

// Stub metrics functions for testing to avoid registration conflicts
type Timer struct{}

func (s *Timer) ObserveDuration() {}

func NewReconcileTimer(name, namespace string) *Timer {
	return &Timer{}
}

func RecordReconcileComplete(name, namespace string, duration float64) {}
func RecordSecretSkipped(name, namespace string)                      {}
func RecordReconcileError(name, namespace, reason string)             {}
func RecordTemplateValidationError(name, namespace string)            {}
func RecordSecretGenerated(name, namespace, secretType string)        {}
func UpdateSecretInstances(name, namespace string, count float64)     {}
func NewGeneratorTimer(generatorType string) *Timer { return &Timer{} }
func RecordGeneratorExecution(generatorType, status string)           {}