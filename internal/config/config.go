package config

import (
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	MetricsBindAddress      string
	HealthProbeBindAddress  string
	LeaderElection          bool
	MaxConcurrentReconciles int
	WatchNamespaces         []string
	IncludeAnnotations      []string
	ExcludeAnnotations      []string
	IncludeLabels           []string
	ExcludeLabels           []string
	DryRun                  bool
	EnableMetadata          bool
	LogFormat               string
	LogLevel                string
}

func Load() *Config {
	viper.SetDefault("metrics-bind-address", ":8080")
	viper.SetDefault("health-probe-bind-address", ":8081")
	viper.SetDefault("leader-elect", false)
	viper.SetDefault("max-concurrent-reconciles", 1)
	viper.SetDefault("watch-namespaces", []string{})
	viper.SetDefault("include-annotations", []string{})
	viper.SetDefault("exclude-annotations", []string{})
	viper.SetDefault("include-labels", []string{})
	viper.SetDefault("exclude-labels", []string{})
	viper.SetDefault("dry-run", false)
	viper.SetDefault("enable-metadata", true)
	viper.SetDefault("log-format", "json")
	viper.SetDefault("log-level", "info")

	viper.SetEnvPrefix("SECRET_SANTA")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	return &Config{
		MetricsBindAddress:      viper.GetString("metrics-bind-address"),
		HealthProbeBindAddress:  viper.GetString("health-probe-bind-address"),
		LeaderElection:          viper.GetBool("leader-elect"),
		MaxConcurrentReconciles: viper.GetInt("max-concurrent-reconciles"),
		WatchNamespaces:         getStringSlice("watch-namespaces"),
		IncludeAnnotations:      getStringSlice("include-annotations"),
		ExcludeAnnotations:      getStringSlice("exclude-annotations"),
		IncludeLabels:           getStringSlice("include-labels"),
		ExcludeLabels:           getStringSlice("exclude-labels"),
		DryRun:                  viper.GetBool("dry-run"),
		EnableMetadata:          viper.GetBool("enable-metadata"),
		LogFormat:               viper.GetString("log-format"),
		LogLevel:                viper.GetString("log-level"),
	}
}

func getStringSlice(key string) []string {
	slice := viper.GetStringSlice(key)
	if len(slice) == 1 && strings.Contains(slice[0], ",") {
		parts := strings.Split(slice[0], ",")
		for i, v := range parts {
			parts[i] = strings.TrimSpace(v)
		}
		return parts
	}
	return slice
}
