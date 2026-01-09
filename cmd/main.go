package main

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap/zapcore"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"

	secretsantav1alpha1 "github.com/logicIQ/secret-santa/api/v1alpha1"
	"github.com/logicIQ/secret-santa/internal/config"
	"github.com/logicIQ/secret-santa/internal/controller"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
	version  = "dev"
	gitHash  = "unknown"
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(secretsantav1alpha1.AddToScheme(scheme))
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "secret-santa",
		Short: "Kubernetes operator for sensitive data generation",
		Run:   run,
	}

	rootCmd.Flags().String("metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	rootCmd.Flags().String("health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	rootCmd.Flags().Bool("leader-elect", false, "Enable leader election for controller manager.")
	rootCmd.Flags().Int("max-concurrent-reconciles", 1, "Maximum number of concurrent reconciles.")
	rootCmd.Flags().StringSlice("watch-namespaces", []string{}, "Comma-separated list of namespaces to watch (empty = all namespaces).")
	rootCmd.Flags().StringSlice("include-annotations", []string{}, "Comma-separated list of annotations to include.")
	rootCmd.Flags().StringSlice("exclude-annotations", []string{}, "Comma-separated list of annotations to exclude.")
	rootCmd.Flags().StringSlice("include-labels", []string{}, "Comma-separated list of labels to include.")
	rootCmd.Flags().StringSlice("exclude-labels", []string{}, "Comma-separated list of labels to exclude.")
	rootCmd.Flags().Bool("dry-run", false, "Enable dry-run mode (validate templates without creating secrets).")
	rootCmd.Flags().Bool("enable-metadata", true, "Enable metadata annotations/tags on generated secrets.")
	rootCmd.Flags().String("log-format", "json", "Log format: json or console")
	rootCmd.Flags().String("log-level", "info", "Log level: debug, info, warn, error")

	if err := viper.BindPFlags(rootCmd.Flags()); err != nil {
		setupLog.Error(err, "unable to bind flags")
		os.Exit(1)
	}

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) {
	logFormat := viper.GetString("log-format")
	logLevel := viper.GetString("log-level")
	dryRun := viper.GetBool("dry-run")

	// Validate log format to prevent log injection
	if logFormat != "json" && logFormat != "console" {
		logFormat = "json"
	}

	// Validate log level to prevent log injection
	validLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !validLevels[logLevel] {
		logLevel = "info"
	}

	opts := zap.Options{
		Development: logFormat == "console",
	}

	// Set log level
	switch logLevel {
	case "debug":
		opts.Level = zapcore.DebugLevel
	case "info":
		opts.Level = zapcore.InfoLevel
	case "warn":
		opts.Level = zapcore.WarnLevel
	case "error":
		opts.Level = zapcore.ErrorLevel
	}

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	setupLog.Info("Secret Santa starting", "version", version, "gitHash", gitHash)
	setupLog.Info("Logging configuration", "format", logFormat, "level", logLevel)
	if dryRun {
		setupLog.Info("Starting in DRY RUN mode - no secrets will be created")
	}

	cfg, err := config.Load()
	if err != nil {
		setupLog.Error(err, "unable to load config")
		os.Exit(1)
	}

	var cacheOpts cache.Options
	if len(cfg.WatchNamespaces) > 0 {
		cacheOpts.DefaultNamespaces = make(map[string]cache.Config)
		for _, ns := range cfg.WatchNamespaces {
			cacheOpts.DefaultNamespaces[ns] = cache.Config{}
		}
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		Metrics:                server.Options{BindAddress: cfg.MetricsBindAddress},
		HealthProbeBindAddress: cfg.HealthProbeBindAddress,
		LeaderElection:         cfg.LeaderElection,
		LeaderElectionID:       "secret-santa-leader-election",
		Cache:                  cacheOpts,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&controller.SecretSantaReconciler{
		Client:             mgr.GetClient(),
		Scheme:             mgr.GetScheme(),
		IncludeAnnotations: cfg.IncludeAnnotations,
		ExcludeAnnotations: cfg.ExcludeAnnotations,
		IncludeLabels:      cfg.IncludeLabels,
		ExcludeLabels:      cfg.ExcludeLabels,
		DryRun:             cfg.DryRun,
		EnableMetadata:     cfg.EnableMetadata,
	}).SetupWithManager(mgr, cfg.MaxConcurrentReconciles); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "SecretSanta")
		os.Exit(1)
	}

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
