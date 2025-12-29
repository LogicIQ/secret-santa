package metrics

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

const (
	Namespace                 = "secretsanta"
	ControllerSubsystem       = "controller"
	GeneratorSubsystem        = "generator"
	KubernetesClientSubsystem = "kubernetes_client"
)

var (
	SuccessGenerationTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: Namespace,
			Subsystem: ControllerSubsystem,
			Name:      "success_generation_total",
			Help:      "Successful secret generations",
		},
		[]string{"secretsanta", "namespace"},
	)

	FailedGenerationTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: Namespace,
			Subsystem: ControllerSubsystem,
			Name:      "failed_generation_total",
			Help:      "Failed secret generations",
		},
		[]string{"secretsanta", "namespace", "reason"},
	)

	LoopSecondsTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: Namespace,
			Subsystem: ControllerSubsystem,
			Name:      "loop_seconds_total",
			Help:      "Total seconds spent in processing loops",
		},
	)

	SecretsSkippedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: Namespace,
			Subsystem: ControllerSubsystem,
			Name:      "secrets_skipped_total",
			Help:      "Secrets skipped (already exist)",
		},
		[]string{"secretsanta", "namespace"},
	)

	TemplateValidationFailedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: Namespace,
			Subsystem: ControllerSubsystem,
			Name:      "template_validation_failed_total",
			Help:      "Template validation failures",
		},
		[]string{"secretsanta", "namespace"},
	)

	GeneratorExecutionsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: Namespace,
			Subsystem: GeneratorSubsystem,
			Name:      "executions_total",
			Help:      "Total generator executions",
		},
		[]string{"generator_type", "status"},
	)

	GeneratorResponseTime = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: Namespace,
			Subsystem: GeneratorSubsystem,
			Name:      "response_time_seconds",
			Help:      "Generator execution times",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"generator_type"},
	)

	KubernetesClientFailTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: Namespace,
			Subsystem: KubernetesClientSubsystem,
			Name:      "fail_total",
			Help:      "Failed API requests",
		},
		[]string{"operation"},
	)

	KubernetesClientRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: Namespace,
			Subsystem: KubernetesClientSubsystem,
			Name:      "requests_total",
			Help:      "Total API requests",
		},
		[]string{"operation", "status"},
	)

	LastReconciliationTime = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: Namespace,
			Name:      "last_reconciliation_timestamp_seconds",
			Help:      "Last reconciliation timestamp",
		},
	)

	ReconciliationStatus = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: Namespace,
			Name:      "reconciliation_status",
			Help:      "Reconciliation status (1=success, 0=failure)",
		},
		[]string{"status"},
	)

	ManagedSecretsTotal = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: Namespace,
			Name:      "managed_secrets_total",
			Help:      "Total managed SecretSanta resources",
		},
	)

	SecretGenerationStatus = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: Namespace,
			Name:      "secret_generation_status",
			Help:      "Secret generation status (1=generated, 0=failed)",
		},
		[]string{"secretsanta", "namespace"},
	)
)

func RecordSuccessfulGeneration(secretSantaName, namespace string) {
	SuccessGenerationTotal.WithLabelValues(secretSantaName, namespace).Inc()
	SecretGenerationStatus.WithLabelValues(secretSantaName, namespace).Set(1)
}

func RecordFailedGeneration(secretSantaName, namespace, reason string) {
	FailedGenerationTotal.WithLabelValues(secretSantaName, namespace, reason).Inc()
	SecretGenerationStatus.WithLabelValues(secretSantaName, namespace).Set(0)
}

func RecordSecretSkipped(secretSantaName, namespace string) {
	SecretsSkippedTotal.WithLabelValues(secretSantaName, namespace).Inc()
}

func RecordTemplateValidationFailed(secretSantaName, namespace string) {
	TemplateValidationFailedTotal.WithLabelValues(secretSantaName, namespace).Inc()
}

func RecordGeneratorExecution(generatorType, status string) {
	GeneratorExecutionsTotal.WithLabelValues(generatorType, status).Inc()
}

func RecordKubernetesClientRequest(operation, status string) {
	KubernetesClientRequestsTotal.WithLabelValues(operation, status).Inc()
	if status == "failed" {
		KubernetesClientFailTotal.WithLabelValues(operation).Inc()
	}
}

func RecordLoopDuration(seconds float64) {
	LoopSecondsTotal.Add(seconds)
}

func NewGeneratorTimer(generatorType string) *prometheus.Timer {
	return prometheus.NewTimer(GeneratorResponseTime.WithLabelValues(generatorType))
}

func UpdateLastReconciliationTime() {
	LastReconciliationTime.SetToCurrentTime()
}

func UpdateReconciliationStatus(success bool) {
	if success {
		ReconciliationStatus.WithLabelValues("success").Set(1)
		ReconciliationStatus.WithLabelValues("failure").Set(0)
	} else {
		ReconciliationStatus.WithLabelValues("success").Set(0)
		ReconciliationStatus.WithLabelValues("failure").Set(1)
	}
}

func UpdateManagedSecretsCount(count float64) {
	ManagedSecretsTotal.Set(count)
}

func NewReconcileTimer(name, namespace string) *prometheus.Timer {
	ReconcileActive.WithLabelValues(name, namespace).Set(1)
	return prometheus.NewTimer(prometheus.ObserverFunc(func(v float64) {
		LastReconcileDuration.WithLabelValues(name, namespace).Set(v)
	}))
}

func RecordReconcileComplete(name, namespace string, duration float64) {
	ReconcileActive.WithLabelValues(name, namespace).Set(0)
	LastReconcileDuration.WithLabelValues(name, namespace).Set(duration)
}

func RecordReconcileError(name, namespace, reason string) {
	SyncErrorCount.WithLabelValues(name, namespace).Inc()
	FailedGenerationTotal.WithLabelValues(name, namespace, reason).Inc()
}

func RecordTemplateValidationError(name, namespace string) {
	TemplateValidationFailedTotal.WithLabelValues(name, namespace).Inc()
}

func RecordSecretGenerated(name, namespace, secretType string) {
	SuccessGenerationTotal.WithLabelValues(name, namespace).Inc()
	SyncCallCount.WithLabelValues(name, namespace).Inc()
}

func UpdateSecretInstances(name, namespace string, count float64) {
	SecretInstances.WithLabelValues(name, namespace).Set(count)
}

var (
	SyncCallCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Subsystem: "secretsanta",
		Name:      "controller_sync_call_count",
		Help:      "The number of reconciliation loops made by a controller",
	}, []string{"secretsanta_name", "secretsanta_namespace"})

	SyncErrorCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Subsystem: "secretsanta",
		Name:      "controller_sync_error_count",
		Help:      "The number of failed reconciliation loops",
	}, []string{"secretsanta_name", "secretsanta_namespace"})

	LastReconcileDuration = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Subsystem: "secretsanta",
		Name:      "controller_last_reconcile_duration_seconds",
		Help:      "Duration of the last reconcile operation",
	}, []string{"secretsanta_name", "secretsanta_namespace"})

	ReconcileActive = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Subsystem: "secretsanta",
		Name:      "controller_reconcile_active",
		Help:      "Shows if Reconcile loop is running",
	}, []string{"secretsanta_name", "secretsanta_namespace"})

	SecretInstances = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Subsystem: "secretsanta",
		Name:      "controller_secrets_instances",
		Help:      "The number of desired secret instances",
	}, []string{"secretsanta_name", "secretsanta_namespace"})
)

var registerOnce sync.Once

func init() {
	registerOnce.Do(func() {
		metrics.Registry.MustRegister(
			SuccessGenerationTotal,
			FailedGenerationTotal,
			LoopSecondsTotal,
			SecretsSkippedTotal,
			TemplateValidationFailedTotal,
			GeneratorExecutionsTotal,
			GeneratorResponseTime,
			KubernetesClientFailTotal,
			KubernetesClientRequestsTotal,
			LastReconciliationTime,
			ReconciliationStatus,
			ManagedSecretsTotal,
			SecretGenerationStatus,
			SyncCallCount,
			SyncErrorCount,
			LastReconcileDuration,
			ReconcileActive,
			SecretInstances,
		)
	})
}
