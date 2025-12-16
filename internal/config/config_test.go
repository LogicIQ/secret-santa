package config

import (
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_Defaults(t *testing.T) {
	// Reset viper state
	viper.Reset()

	cfg, err := Load()
	require.NoError(t, err)

	assert.Equal(t, ":8080", cfg.MetricsBindAddress)
	assert.Equal(t, ":8081", cfg.HealthProbeBindAddress)
	assert.False(t, cfg.LeaderElection)
	assert.Equal(t, 1, cfg.MaxConcurrentReconciles)
	assert.Empty(t, cfg.WatchNamespaces)
	assert.Empty(t, cfg.IncludeAnnotations)
	assert.Empty(t, cfg.ExcludeAnnotations)
	assert.Empty(t, cfg.IncludeLabels)
	assert.Empty(t, cfg.ExcludeLabels)
	assert.False(t, cfg.DryRun)
	assert.Equal(t, "json", cfg.LogFormat)
	assert.Equal(t, "info", cfg.LogLevel)
}

func TestLoad_EnvironmentVariables(t *testing.T) {
	// Reset viper state
	viper.Reset()

	// Set environment variables
	envVars := map[string]string{
		"SECRET_SANTA_METRICS_BIND_ADDRESS":      ":9090",
		"SECRET_SANTA_HEALTH_PROBE_BIND_ADDRESS": ":9091",
		"SECRET_SANTA_LEADER_ELECT":              "true",
		"SECRET_SANTA_MAX_CONCURRENT_RECONCILES": "5",
		"SECRET_SANTA_WATCH_NAMESPACES":          "default,kube-system",
		"SECRET_SANTA_INCLUDE_ANNOTATIONS":       "app.kubernetes.io/name",
		"SECRET_SANTA_EXCLUDE_ANNOTATIONS":       "skip.secret-santa.io/ignore",
		"SECRET_SANTA_INCLUDE_LABELS":            "environment=prod",
		"SECRET_SANTA_EXCLUDE_LABELS":            "skip=true",
		"SECRET_SANTA_DRY_RUN":                   "true",
		"SECRET_SANTA_LOG_FORMAT":                "console",
		"SECRET_SANTA_LOG_LEVEL":                 "debug",
	}

	// Set environment variables
	for key, value := range envVars {
		os.Setenv(key, value)
	}

	// Clean up after test
	defer func() {
		for key := range envVars {
			os.Unsetenv(key)
		}
	}()

	cfg, err := Load()
	require.NoError(t, err)

	assert.Equal(t, ":9090", cfg.MetricsBindAddress)
	assert.Equal(t, ":9091", cfg.HealthProbeBindAddress)
	assert.True(t, cfg.LeaderElection)
	assert.Equal(t, 5, cfg.MaxConcurrentReconciles)
	assert.Equal(t, []string{"default", "kube-system"}, cfg.WatchNamespaces)
	assert.Equal(t, []string{"app.kubernetes.io/name"}, cfg.IncludeAnnotations)
	assert.Equal(t, []string{"skip.secret-santa.io/ignore"}, cfg.ExcludeAnnotations)
	assert.Equal(t, []string{"environment=prod"}, cfg.IncludeLabels)
	assert.Equal(t, []string{"skip=true"}, cfg.ExcludeLabels)
	assert.True(t, cfg.DryRun)
	assert.Equal(t, "console", cfg.LogFormat)
	assert.Equal(t, "debug", cfg.LogLevel)
}

func TestLoad_ViperSettings(t *testing.T) {
	// Reset viper state
	viper.Reset()

	// Set values directly in viper
	viper.Set("metrics-bind-address", ":7070")
	viper.Set("leader-elect", true)
	viper.Set("max-concurrent-reconciles", 10)
	viper.Set("watch-namespaces", []string{"test-ns"})
	viper.Set("dry-run", true)
	viper.Set("log-format", "console")
	viper.Set("log-level", "warn")

	cfg, err := Load()
	require.NoError(t, err)

	assert.Equal(t, ":7070", cfg.MetricsBindAddress)
	assert.True(t, cfg.LeaderElection)
	assert.Equal(t, 10, cfg.MaxConcurrentReconciles)
	assert.Equal(t, []string{"test-ns"}, cfg.WatchNamespaces)
	assert.True(t, cfg.DryRun)
	assert.Equal(t, "console", cfg.LogFormat)
	assert.Equal(t, "warn", cfg.LogLevel)
}

func TestLoad_EnvOverridesDefaults(t *testing.T) {
	// Reset viper state
	viper.Reset()

	// Set one environment variable
	os.Setenv("SECRET_SANTA_MAX_CONCURRENT_RECONCILES", "3")
	defer os.Unsetenv("SECRET_SANTA_MAX_CONCURRENT_RECONCILES")

	cfg, err := Load()
	require.NoError(t, err)

	// Environment variable should override default
	assert.Equal(t, 3, cfg.MaxConcurrentReconciles)
	// Other values should remain defaults
	assert.Equal(t, ":8080", cfg.MetricsBindAddress)
	assert.False(t, cfg.LeaderElection)
}

func TestLoad_CommaSeparatedEnvVars(t *testing.T) {
	// Reset viper state
	viper.Reset()

	// Test comma-separated environment variables
	os.Setenv("SECRET_SANTA_WATCH_NAMESPACES", "ns1,ns2,ns3")
	os.Setenv("SECRET_SANTA_INCLUDE_LABELS", "app=web,env=prod")
	defer func() {
		os.Unsetenv("SECRET_SANTA_WATCH_NAMESPACES")
		os.Unsetenv("SECRET_SANTA_INCLUDE_LABELS")
	}()

	cfg, err := Load()
	require.NoError(t, err)

	assert.Equal(t, []string{"ns1", "ns2", "ns3"}, cfg.WatchNamespaces)
	assert.Equal(t, []string{"app=web", "env=prod"}, cfg.IncludeLabels)
}

func TestConfig_StructFields(t *testing.T) {
	cfg := &Config{
		MetricsBindAddress:      ":8080",
		HealthProbeBindAddress:  ":8081",
		LeaderElection:          true,
		MaxConcurrentReconciles: 5,
		WatchNamespaces:         []string{"default"},
		IncludeAnnotations:      []string{"include"},
		ExcludeAnnotations:      []string{"exclude"},
		IncludeLabels:           []string{"env=prod"},
		ExcludeLabels:           []string{"skip=true"},
		DryRun:                  true,
		LogFormat:               "json",
		LogLevel:                "debug",
	}

	assert.Equal(t, ":8080", cfg.MetricsBindAddress)
	assert.Equal(t, ":8081", cfg.HealthProbeBindAddress)
	assert.True(t, cfg.LeaderElection)
	assert.Equal(t, 5, cfg.MaxConcurrentReconciles)
	assert.Equal(t, []string{"default"}, cfg.WatchNamespaces)
	assert.Equal(t, []string{"include"}, cfg.IncludeAnnotations)
	assert.Equal(t, []string{"exclude"}, cfg.ExcludeAnnotations)
	assert.Equal(t, []string{"env=prod"}, cfg.IncludeLabels)
	assert.Equal(t, []string{"skip=true"}, cfg.ExcludeLabels)
	assert.True(t, cfg.DryRun)
	assert.Equal(t, "json", cfg.LogFormat)
	assert.Equal(t, "debug", cfg.LogLevel)
}

func TestGetStringSlice(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		value    interface{}
		expected []string
	}{
		{
			name:     "comma-separated string",
			key:      "test-comma",
			value:    "a,b,c",
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "single string",
			key:      "test-single",
			value:    "single",
			expected: []string{"single"},
		},
		{
			name:     "string slice",
			key:      "test-slice",
			value:    []string{"a", "b"},
			expected: []string{"a", "b"},
		},
		{
			name:     "empty",
			key:      "test-empty",
			value:    []string{},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Reset()
			viper.Set(tt.key, tt.value)
			result := getStringSlice(tt.key)
			assert.Equal(t, tt.expected, result)
		})
	}
}
