package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

// SetupTestRegistry configures a separate registry for tests to avoid conflicts
func SetupTestRegistry() {
	metrics.Registry = prometheus.NewRegistry()
}
