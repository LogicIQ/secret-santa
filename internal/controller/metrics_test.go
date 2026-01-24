package controller

import (
	"os"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

func TestMain(m *testing.M) {
	// Use a separate registry for tests to avoid conflicts
	metrics.Registry = prometheus.NewRegistry()
	os.Exit(m.Run())
}
