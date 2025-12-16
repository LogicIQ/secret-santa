//go:build !test

package controller

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/logicIQ/secret-santa/pkg/metrics"
)

type Timer = prometheus.Timer

// Production metrics - delegate to actual metrics package
func NewReconcileTimer(name, namespace string) *Timer {
	return metrics.NewReconcileTimer(name, namespace)
}

func RecordReconcileComplete(name, namespace string, duration float64) {
	metrics.RecordReconcileComplete(name, namespace, duration)
}

func RecordSecretSkipped(name, namespace string) {
	metrics.RecordSecretSkipped(name, namespace)
}

func RecordReconcileError(name, namespace, reason string) {
	metrics.RecordReconcileError(name, namespace, reason)
}

func RecordTemplateValidationError(name, namespace string) {
	metrics.RecordTemplateValidationError(name, namespace)
}

func RecordSecretGenerated(name, namespace, secretType string) {
	metrics.RecordSecretGenerated(name, namespace, secretType)
}

func UpdateSecretInstances(name, namespace string, count float64) {
	metrics.UpdateSecretInstances(name, namespace, count)
}

func NewGeneratorTimer(generatorType string) *Timer {
	return metrics.NewGeneratorTimer(generatorType)
}

func RecordGeneratorExecution(generatorType, status string) {
	metrics.RecordGeneratorExecution(generatorType, status)
}