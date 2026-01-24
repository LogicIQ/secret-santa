package controller

import (
	"github.com/logicIQ/secret-santa/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

// Production metrics wrapper functions
func NewReconcileTimer(name, namespace string) *prometheus.Timer {
	return metrics.NewReconcileTimer(name, namespace)
}

func RecordReconcileComplete(name, namespace string, duration float64) {
	metrics.RecordReconcileComplete(name, namespace, duration)
}

func RecordSecretSkipped(name, namespace string) {
	metrics.RecordSecretSkipped(name, namespace)
}

func RecordReconcileError(name, namespace string) {
	metrics.RecordReconcileError(name, namespace)
}

func RecordTemplateValidationFailed(name, namespace string) {
	metrics.RecordTemplateValidationFailed(name, namespace)
}

func RecordSuccessfulGeneration(name, namespace string) {
	metrics.RecordSuccessfulGeneration(name, namespace)
}

func UpdateSecretInstances(name, namespace string, count float64) {
	metrics.UpdateSecretInstances(name, namespace, count)
}

func NewGeneratorTimer(generatorType string) *prometheus.Timer {
	return metrics.NewGeneratorTimer(generatorType)
}

func RecordGeneratorExecution(generatorType, status string) {
	metrics.RecordGeneratorExecution(generatorType, status)
}
