package metrics

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

const (
	Namespace                 = "secretsanta"
	ControllerSubsystem       = "controller"
	GeneratorSubsystem        = "generator"
	KubernetesClientSubsystem = "kubernetes_client"
)

// Controller metrics
var (
	SecretsGenerated = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: Namespace,
			Subsystem: ControllerSubsystem,
			Name:      "secrets_generated_total",
			Help:      "Total number of secrets generated",
		},
		[]string{"namespace", "name", "type"},
	)

	SecretsSkipped = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: Namespace,
			Subsystem: ControllerSubsystem,
			Name:      "secrets_skipped_total",
			Help:      "Total number of secrets skipped (already exists)",
		},
		[]string{"namespace", "name"},
	)

	ReconcileErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: Namespace,
			Subsystem: ControllerSubsystem,
			Name:      "reconcile_errors_total",
			Help:      "Total number of reconcile errors",
		},
		[]string{"namespace", "name", "reason"},
	)

	ReconcileTime = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: Namespace,
			Subsystem: ControllerSubsystem,
			Name:      "reconcile_duration_seconds",
			Help:      "Time spent reconciling SecretSanta resources",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"namespace", "name"},
	)

	LastReconciliationTime = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: Namespace,
			Subsystem: ControllerSubsystem,
			Name:      "last_reconciliation_timestamp_seconds",
			Help:      "Timestamp of the last reconciliation",
		},
	)

	ManagedSecretsTotal = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: Namespace,
			Subsystem: ControllerSubsystem,
			Name:      "managed_secrets_total",
			Help:      "Total number of SecretSanta resources currently managed",
		},
	)

	ReconcileActive = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: Namespace,
			Subsystem: ControllerSubsystem,
			Name:      "reconcile_active",
			Help:      "Shows if reconcile loop is currently running",
		},
		[]string{"namespace", "name"},
	)

	LastReconcileDuration = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: Namespace,
			Subsystem: ControllerSubsystem,
			Name:      "last_reconcile_duration_seconds",
			Help:      "Duration of the last reconcile operation",
		},
		[]string{"namespace", "name"},
	)

	SecretInstances = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: Namespace,
			Subsystem: ControllerSubsystem,
			Name:      "secret_instances",
			Help:      "Number of secret instances managed per SecretSanta resource",
		},
		[]string{"namespace", "name"},
	)
)

// Generator metrics
var (
	GeneratorExecutions = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: Namespace,
			Subsystem: GeneratorSubsystem,
			Name:      "executions_total",
			Help:      "Total number of generator executions",
		},
		[]string{"generator_type", "status"},
	)

	TemplateValidationErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: Namespace,
			Subsystem: GeneratorSubsystem,
			Name:      "template_validation_errors_total",
			Help:      "Total number of template validation errors",
		},
		[]string{"namespace", "name"},
	)

	GeneratorDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: Namespace,
			Subsystem: GeneratorSubsystem,
			Name:      "duration_seconds",
			Help:      "Time spent executing generators",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"generator_type"},
	)
)

// Kubernetes client metrics
var (
	KubernetesClientRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: Namespace,
			Subsystem: KubernetesClientSubsystem,
			Name:      "requests_total",
			Help:      "Total API requests to Kubernetes",
		},
		[]string{"operation", "status"},
	)

	KubernetesClientErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: Namespace,
			Subsystem: KubernetesClientSubsystem,
			Name:      "errors_total",
			Help:      "Total API request errors to Kubernetes",
		},
		[]string{"operation"},
	)
)

// Helper functions
func RecordSecretGenerated(namespace, name, secretType string) {
	SecretsGenerated.WithLabelValues(namespace, name, secretType).Inc()
	LastReconciliationTime.SetToCurrentTime()
}

func RecordSecretSkipped(namespace, name string) {
	SecretsSkipped.WithLabelValues(namespace, name).Inc()
}

func RecordReconcileError(namespace, name, reason string) {
	ReconcileErrors.WithLabelValues(namespace, name, reason).Inc()
}

func RecordTemplateValidationError(namespace, name string) {
	TemplateValidationErrors.WithLabelValues(namespace, name).Inc()
}

func RecordGeneratorExecution(generatorType, status string) {
	GeneratorExecutions.WithLabelValues(generatorType, status).Inc()
}

func RecordKubernetesRequest(operation, status string) {
	KubernetesClientRequests.WithLabelValues(operation, status).Inc()
	if status == "error" {
		KubernetesClientErrors.WithLabelValues(operation).Inc()
	}
}

func NewReconcileTimer(namespace, name string) *prometheus.Timer {
	ReconcileActive.WithLabelValues(namespace, name).Set(1)
	return ReconcileTime.WithLabelValues(namespace, name).NewTimer()
}

func RecordReconcileComplete(namespace, name string, duration float64) {
	ReconcileActive.WithLabelValues(namespace, name).Set(0)
	LastReconcileDuration.WithLabelValues(namespace, name).Set(duration)
}

func UpdateSecretInstances(namespace, name string, count float64) {
	SecretInstances.WithLabelValues(namespace, name).Set(count)
}

func NewGeneratorTimer(generatorType string) *prometheus.Timer {
	return GeneratorDuration.WithLabelValues(generatorType).NewTimer()
}

func UpdateManagedSecretsCount(count float64) {
	ManagedSecretsTotal.Set(count)
}

var registerOnce sync.Once

func init() {
	registerOnce.Do(func() {
		metrics.Registry.MustRegister(
			// Controller metrics
			SecretsGenerated,
			SecretsSkipped,
			ReconcileErrors,
			ReconcileTime,
			LastReconciliationTime,
			ManagedSecretsTotal,
			ReconcileActive,
			LastReconcileDuration,
			SecretInstances,
			// Generator metrics
			GeneratorExecutions,
			TemplateValidationErrors,
			GeneratorDuration,
			// Client metrics
			KubernetesClientRequests,
			KubernetesClientErrors,
		)
	})
}