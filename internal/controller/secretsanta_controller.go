package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"text/template"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	secretsantav1alpha1 "github.com/logicIQ/secret-santa/api/v1alpha1"
	"github.com/logicIQ/secret-santa/pkg/generators"
	"github.com/logicIQ/secret-santa/pkg/media"
	"github.com/logicIQ/secret-santa/pkg/media/aws"
	"github.com/logicIQ/secret-santa/pkg/media/gcp"
	"github.com/logicIQ/secret-santa/pkg/media/k8s"
	tmplpkg "github.com/logicIQ/secret-santa/pkg/template"
	"github.com/logicIQ/secret-santa/pkg/validation"
)

const (
	SecretSantaFinalizer = "secrets.secret-santa.io/finalizer"
)

type SecretSantaReconciler struct {
	client.Client
	Scheme             *runtime.Scheme
	IncludeAnnotations []string
	ExcludeAnnotations []string
	IncludeLabels      []string
	ExcludeLabels      []string
	DryRun             bool
	EnableMetadata     bool
}

func (r *SecretSantaReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx).WithValues("secretsanta", req.Name, "namespace", req.Namespace)
	start := time.Now()
	log.V(1).Info("Starting reconcile", "dryRun", r.DryRun)

	timer := NewReconcileTimer(req.Name, req.Namespace)
	defer func() {
		duration := time.Since(start)
		timer.ObserveDuration()
		RecordReconcileComplete(req.Name, req.Namespace, duration.Seconds())
		log.V(1).Info("Completed reconcile", "duration", duration)
	}()

	var secretSanta secretsantav1alpha1.SecretSanta
	if err := r.Get(ctx, req.NamespacedName, &secretSanta); err != nil {
		if errors.IsNotFound(err) {
			log.V(1).Info("Resource not found")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	log.V(1).Info("Found SecretSanta resource", "secretType", secretSanta.Spec.SecretType, "generators", len(secretSanta.Spec.Generators))

	if secretSanta.ObjectMeta.DeletionTimestamp.IsZero() {
		if !controllerutil.ContainsFinalizer(&secretSanta, SecretSantaFinalizer) {
			controllerutil.AddFinalizer(&secretSanta, SecretSantaFinalizer)
			return ctrl.Result{}, r.Update(ctx, &secretSanta)
		}
	} else {
		return r.handleDeletion(ctx, &secretSanta)
	}

	if !r.shouldProcess(&secretSanta) {
		log.V(1).Info("Skipping resource due to filters", "includeAnnotations", r.IncludeAnnotations, "excludeAnnotations", r.ExcludeAnnotations)
		return ctrl.Result{}, nil
	}

	return r.reconcileSecret(ctx, &secretSanta)
}

func (r *SecretSantaReconciler) handleDeletion(ctx context.Context, secretSanta *secretsantav1alpha1.SecretSanta) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.Info("Handling SecretSanta deletion")

	controllerutil.RemoveFinalizer(secretSanta, SecretSantaFinalizer)
	return ctrl.Result{}, r.Update(ctx, secretSanta)
}

func (r *SecretSantaReconciler) reconcileSecret(ctx context.Context, secretSanta *secretsantav1alpha1.SecretSanta) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.V(1).Info("Reconciling secret", "templateLength", len(secretSanta.Spec.Template))

	// Handle dry-run mode (from spec or controller flag)
	if secretSanta.Spec.DryRun || r.DryRun {
		return r.handleDryRun(ctx, secretSanta)
	}

	// Check if we already processed this SecretSanta successfully
	for _, condition := range secretSanta.Status.Conditions {
		if condition.Type == "Ready" && condition.Status == metav1.ConditionTrue {
			log.V(1).Info("Secret already processed - create-once policy enforced")
			return ctrl.Result{}, nil
		}
	}

	// Determine secret name
	secretName := secretSanta.Spec.SecretName
	if secretName == "" {
		secretName = secretSanta.Name
	}

	// Check if secret already exists (created outside this controller)
	var existingSecret corev1.Secret
	err := r.Get(ctx, client.ObjectKey{Name: secretName, Namespace: secretSanta.Namespace}, &existingSecret)
	if err == nil {
		log.Info("Secret already exists - create-once policy enforced")
		RecordSecretSkipped(secretSanta.Name, secretSanta.Namespace)
		if updateErr := r.updateStatus(ctx, secretSanta, "Ready", "True", "Secret already exists"); updateErr != nil {
			log.Error(updateErr, "Failed to update status")
		}
		return ctrl.Result{}, nil
	} else if !errors.IsNotFound(err) {
		return ctrl.Result{}, err
	}

	// Validate generators first
	if err := validation.ValidateGeneratorConfigs(secretSanta.Spec.Generators); err != nil {
		log.Error(err, "Generator validation failed")
		RecordReconcileError(secretSanta.Name, secretSanta.Namespace, "generator_validation_failed")
		if updateErr := r.updateStatus(ctx, secretSanta, "DryRunFailed", "False", err.Error()); updateErr != nil {
			log.Error(updateErr, "Failed to update status")
		}
		return ctrl.Result{}, nil
	}

	templateData, err := r.generateTemplateData(secretSanta.Spec.Generators)
	if err != nil {
		log.Error(err, "Failed to generate template data")
		RecordReconcileError(secretSanta.Name, secretSanta.Namespace, "generator_failed")
		if updateErr := r.updateStatus(ctx, secretSanta, "GeneratorFailed", "False", err.Error()); updateErr != nil {
			log.Error(updateErr, "Failed to update status")
		}
		return ctrl.Result{}, nil
	}
	log.V(1).Info("Template data generated", "generators", len(secretSanta.Spec.Generators))

	if err := r.validateTemplate(secretSanta.Spec.Template); err != nil {
		log.Error(err, "Template validation failed")
		RecordTemplateValidationError(secretSanta.Name, secretSanta.Namespace)
		if updateErr := r.updateStatus(ctx, secretSanta, "TemplateFailed", "False", err.Error()); updateErr != nil {
			log.Error(updateErr, "Failed to update status")
		}
		return ctrl.Result{}, nil
	}

	secretData, err := r.executeTemplate(secretSanta.Spec.Template, templateData)
	if err != nil {
		log.Error(err, "Template execution failed")
		if updateErr := r.updateStatus(ctx, secretSanta, "TemplateExecutionFailed", "False", err.Error()); updateErr != nil {
			log.Error(updateErr, "Failed to update status")
		}
		return ctrl.Result{}, nil
	}
	log.V(1).Info("Template executed successfully", "dataSize", len(secretData))

	return r.storeSecret(ctx, secretSanta, secretData)
}

func (r *SecretSantaReconciler) storeSecret(ctx context.Context, secretSanta *secretsantav1alpha1.SecretSanta, data string) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Create media instance based on configuration
	mediaInstance, err := r.createMedia(secretSanta)
	if err != nil {
		log.Error(err, "Failed to create media instance")
		return ctrl.Result{}, err
	}

	// Store the secret using the media
	if err := mediaInstance.Store(ctx, secretSanta, data, r.EnableMetadata); err != nil {
		log.Error(err, "Failed to store secret", "mediaType", mediaInstance.GetType())
		if updateErr := r.updateStatus(ctx, secretSanta, "SecretStorageFailed", "False", err.Error()); updateErr != nil {
			log.Error(updateErr, "Failed to update status")
		}
		return ctrl.Result{}, err
	}

	RecordSecretGenerated(secretSanta.Name, secretSanta.Namespace, mediaInstance.GetType())
	UpdateSecretInstances(secretSanta.Name, secretSanta.Namespace, 1)
	if updateErr := r.updateStatus(ctx, secretSanta, "Ready", "True", "Secret stored successfully"); updateErr != nil {
		log.Error(updateErr, "Failed to update status")
	}

	log.Info("Secret stored successfully", "mediaType", mediaInstance.GetType())
	return ctrl.Result{}, nil
}

func (r *SecretSantaReconciler) createMedia(secretSanta *secretsantav1alpha1.SecretSanta) (media.Media, error) {
	// Default to K8s secrets if no media is specified
	if secretSanta.Spec.Media == nil {
		return &k8s.K8sSecretsMedia{Client: r.Client}, nil
	}

	// Parse media config
	var config map[string]interface{}
	if secretSanta.Spec.Media.Config != nil && len(secretSanta.Spec.Media.Config.Raw) > 0 {
		if err := json.Unmarshal(secretSanta.Spec.Media.Config.Raw, &config); err != nil {
			return nil, fmt.Errorf("failed to unmarshal media config: %w", err)
		}
	}
	if config == nil {
		config = make(map[string]interface{})
	}

	switch secretSanta.Spec.Media.Type {
	case "k8s", "":
		secretName := ""
		if s, ok := config["secret_name"].(string); ok {
			secretName = s
		}
		return &k8s.K8sSecretsMedia{
			Client:     r.Client,
			SecretName: secretName,
		}, nil
	case "aws-secrets-manager":
		region := ""
		if r, ok := config["region"].(string); ok {
			region = r
		}
		secretName := ""
		if s, ok := config["secret_name"].(string); ok {
			secretName = s
		}
		kmsKeyId := ""
		if k, ok := config["kms_key_id"].(string); ok {
			kmsKeyId = k
		}
		return &aws.AWSSecretsManagerMedia{
			Region:     region,
			SecretName: secretName,
			KMSKeyId:   kmsKeyId,
		}, nil
	case "aws-parameter-store":
		region := ""
		if r, ok := config["region"].(string); ok {
			region = r
		}
		parameterName := ""
		if p, ok := config["parameter_name"].(string); ok {
			parameterName = p
		}
		kmsKeyId := ""
		if k, ok := config["kms_key_id"].(string); ok {
			kmsKeyId = k
		}
		return &aws.AWSParameterStoreMedia{
			Region:        region,
			ParameterName: parameterName,
			KMSKeyId:      kmsKeyId,
		}, nil
	case "gcp-secret-manager":
		projectID := ""
		if p, ok := config["project_id"].(string); ok {
			projectID = p
		}
		secretName := ""
		if s, ok := config["secret_name"].(string); ok {
			secretName = s
		}
		credentialsFile := ""
		if c, ok := config["credentials_file"].(string); ok {
			credentialsFile = c
		}
		return &gcp.GCPSecretManagerMedia{
			ProjectID:       projectID,
			SecretName:      secretName,
			CredentialsFile: credentialsFile,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported media type: %s", secretSanta.Spec.Media.Type)
	}
}

func (r *SecretSantaReconciler) executeTemplate(tmplStr string, data map[string]interface{}) (string, error) {
	tmpl, err := template.New("secret").Funcs(tmplpkg.FuncMap()).Parse(tmplStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

func (r *SecretSantaReconciler) generateTemplateData(generatorConfigs []secretsantav1alpha1.GeneratorConfig) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	for _, config := range generatorConfigs {
		if err := r.validateGeneratorConfig(config); err != nil {
			return nil, fmt.Errorf("invalid generator config %s: %w", config.Name, err)
		}

		log := ctrl.Log.WithName("generator").WithValues("name", config.Name, "type", config.Type)
		
		// Get generator from registry
		gen, err := generators.Get(config.Type)
		if err != nil {
			return nil, fmt.Errorf("failed to get generator %s: %w", config.Name, err)
		}

		// Convert RawExtension to map[string]interface{}
		var configMap map[string]interface{}
		if config.Config != nil && len(config.Config.Raw) > 0 {
			if err := json.Unmarshal(config.Config.Raw, &configMap); err != nil {
				return nil, fmt.Errorf("failed to unmarshal config for generator %s: %w", config.Name, err)
			}
		}
		if configMap == nil {
			configMap = make(map[string]interface{})
		}

		log.V(1).Info("Executing generator", "config", configMap)
		timer := NewGeneratorTimer(config.Type)
		result, err := gen.Generate(configMap)
		timer.ObserveDuration()
		if err != nil {
			log.Error(err, "Generator failed")
			RecordGeneratorExecution(config.Type, "error")
			return nil, fmt.Errorf("generator %s failed: %w", config.Name, err)
		}
		log.V(1).Info("Generator completed", "resultKeys", getMapKeys(result))
		RecordGeneratorExecution(config.Type, "success")
		data[config.Name] = result
	}

	return data, nil
}

func (r *SecretSantaReconciler) validateTemplate(tmplStr string) error {
	if tmplStr == "" {
		return fmt.Errorf("template cannot be empty")
	}

	_, err := template.New("validation").Funcs(tmplpkg.FuncMap()).Parse(tmplStr)
	return err
}

func (r *SecretSantaReconciler) shouldProcess(secretSanta *secretsantav1alpha1.SecretSanta) bool {
	// Include annotations: ALL must be present (AND logic)
	for _, include := range r.IncludeAnnotations {
		if _, exists := secretSanta.Annotations[include]; !exists {
			return false
		}
	}

	// Exclude annotations: ANY present means skip (OR logic)
	for _, exclude := range r.ExcludeAnnotations {
		if _, exists := secretSanta.Annotations[exclude]; exists {
			return false
		}
	}

	// Include labels: ALL must be present (AND logic)
	for _, include := range r.IncludeLabels {
		if _, exists := secretSanta.Labels[include]; !exists {
			return false
		}
	}

	// Exclude labels: ANY present means skip (OR logic)
	for _, exclude := range r.ExcludeLabels {
		if _, exists := secretSanta.Labels[exclude]; exists {
			return false
		}
	}

	return true
}

func (r *SecretSantaReconciler) validateGeneratorConfig(config secretsantav1alpha1.GeneratorConfig) error {
	if config.Name == "" {
		return fmt.Errorf("generator name cannot be empty")
	}
	if config.Type == "" {
		return fmt.Errorf("generator type cannot be empty")
	}
	return nil
}

func (r *SecretSantaReconciler) updateStatus(ctx context.Context, secretSanta *secretsantav1alpha1.SecretSanta, conditionType, status, message string) error {
	now := metav1.Now()
	secretSanta.Status.LastGenerated = &now

	conditionFound := false
	for i, condition := range secretSanta.Status.Conditions {
		if condition.Type == conditionType {
			secretSanta.Status.Conditions[i].Status = metav1.ConditionStatus(status)
			secretSanta.Status.Conditions[i].Message = message
			secretSanta.Status.Conditions[i].LastTransitionTime = now
			conditionFound = true
			break
		}
	}

	if !conditionFound {
		secretSanta.Status.Conditions = append(secretSanta.Status.Conditions, metav1.Condition{
			Type:               conditionType,
			Status:             metav1.ConditionStatus(status),
			Message:            message,
			LastTransitionTime: now,
			Reason:             conditionType,
		})
	}

	err := r.Status().Update(ctx, secretSanta)
	if errors.IsNotFound(err) {
		// Resource was deleted, ignore the error
		return nil
	}
	return err
}

func (r *SecretSantaReconciler) handleDryRun(ctx context.Context, secretSanta *secretsantav1alpha1.SecretSanta) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.Info("Running dry-run with masked output")

	// Validate template first
	if err := validation.ValidateTemplate(secretSanta.Spec.Template); err != nil {
		log.Error(err, "Template validation failed")
		if updateErr := r.updateStatus(ctx, secretSanta, "DryRunFailed", "False", fmt.Sprintf("Template validation failed: %v", err)); updateErr != nil {
			log.Error(updateErr, "Failed to update dry-run status")
		}
		return ctrl.Result{}, nil
	}

	// Validate generators
	if err := validation.ValidateGeneratorConfigs(secretSanta.Spec.Generators); err != nil {
		log.Error(err, "Generator validation failed")
		if updateErr := r.updateStatus(ctx, secretSanta, "DryRunFailed", "False", fmt.Sprintf("Generator validation failed: %v", err)); updateErr != nil {
			log.Error(updateErr, "Failed to update dry-run status")
		}
		return ctrl.Result{}, nil
	}

	// Generate template data
	templateData, err := r.generateTemplateData(secretSanta.Spec.Generators)
	if err != nil {
		log.Error(err, "Failed to generate template data for dry-run")
		if updateErr := r.updateStatus(ctx, secretSanta, "DryRunFailed", "False", err.Error()); updateErr != nil {
			log.Error(updateErr, "Failed to update dry-run status")
		}
		return ctrl.Result{}, nil
	}

	// Execute template
	secretData, err := r.executeTemplate(secretSanta.Spec.Template, templateData)
	if err != nil {
		log.Error(err, "Template execution failed during dry-run")
		if updateErr := r.updateStatus(ctx, secretSanta, "DryRunFailed", "False", err.Error()); updateErr != nil {
			log.Error(updateErr, "Failed to update dry-run status")
		}
		return ctrl.Result{}, nil
	}

	// Mask sensitive data
	maskedOutput := validation.MaskSensitiveData(secretData)

	// Collect generator names used
	generatorsUsed := make([]string, len(secretSanta.Spec.Generators))
	for i, gen := range secretSanta.Spec.Generators {
		generatorsUsed[i] = fmt.Sprintf("%s (%s)", gen.Name, gen.Type)
	}

	// Create dry-run result
	now := metav1.Now()
	dryRunResult := &secretsantav1alpha1.DryRunResult{
		MaskedOutput:   maskedOutput,
		GeneratorsUsed: generatorsUsed,
		ExecutionTime:  &now,
	}

	// Update status with dry-run result
	secretSanta.Status.DryRunResult = dryRunResult

	if err := r.updateStatus(ctx, secretSanta, "DryRunComplete", "True", "Dry-run completed successfully with masked output"); err != nil {
		log.Error(err, "Failed to update dry-run status")
		return ctrl.Result{}, err
	}

	log.Info("Dry-run completed successfully", "generatorsUsed", len(generatorsUsed))
	return ctrl.Result{}, nil
}

func getMapKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func (r *SecretSantaReconciler) SetupWithManager(mgr ctrl.Manager, maxConcurrentReconciles int) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&secretsantav1alpha1.SecretSanta{}).
		Complete(r)
}
