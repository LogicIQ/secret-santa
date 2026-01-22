//go:build test

package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

func init() {
	// Use a separate registry for tests to avoid conflicts
	metrics.Registry = prometheus.NewRegistry()
}
