package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
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
	"github.com/logicIQ/secret-santa/pkg/generators/crypto"
	"github.com/logicIQ/secret-santa/pkg/generators/random"
	timegens "github.com/logicIQ/secret-santa/pkg/generators/time"
	"github.com/logicIQ/secret-santa/pkg/generators/tls"
	tmplpkg "github.com/logicIQ/secret-santa/pkg/template"
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

	templateData, err := r.generateTemplateData(secretSanta.Spec.Generators)
	if err != nil {
		log.Error(err, "Failed to generate template data")
		RecordReconcileError(secretSanta.Name, secretSanta.Namespace, "generator_failed")
		return ctrl.Result{}, err
	}
	log.V(1).Info("Template data generated", "generators", len(secretSanta.Spec.Generators))

	if err := r.validateTemplate(secretSanta.Spec.Template); err != nil {
		log.Error(err, "Template validation failed")
		RecordTemplateValidationError(secretSanta.Name, secretSanta.Namespace)
		if !r.DryRun {
			if updateErr := r.updateStatus(ctx, secretSanta, "TemplateFailed", "False", err.Error()); updateErr != nil {
				log.Error(updateErr, "Failed to update status")
			}
		}
		return ctrl.Result{}, err
	}

	secretData, err := r.executeTemplate(secretSanta.Spec.Template, templateData)
	if err != nil {
		log.Error(err, "Template execution failed")
		if !r.DryRun {
			if updateErr := r.updateStatus(ctx, secretSanta, "TemplateExecutionFailed", "False", err.Error()); updateErr != nil {
				log.Error(updateErr, "Failed to update status")
			}
		}
		return ctrl.Result{}, err
	}
	log.V(1).Info("Template executed successfully", "dataSize", len(secretData))

	if r.DryRun {
		log.Info("DRY RUN: Template execution successful", "dataSize", len(secretData))
		return ctrl.Result{}, nil
	}

	return r.createOrUpdateSecret(ctx, secretSanta, secretData, secretName)
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

func (r *SecretSantaReconciler) createOrUpdateSecret(ctx context.Context, secretSanta *secretsantav1alpha1.SecretSanta, data string, secretName string) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Handle TLS secrets specially
	stringData := map[string]string{}
	if secretSanta.Spec.SecretType == "kubernetes.io/tls" {
		// For TLS secrets, parse the template output to extract tls.crt and tls.key
		lines := strings.Split(data, "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "tls.crt:") {
				stringData["tls.crt"] = strings.TrimSpace(strings.TrimPrefix(line, "tls.crt:"))
			} else if strings.HasPrefix(line, "tls.key:") {
				stringData["tls.key"] = strings.TrimSpace(strings.TrimPrefix(line, "tls.key:"))
			}
		}
		// If we don't have the required fields, fall back to data field
		if stringData["tls.crt"] == "" || stringData["tls.key"] == "" {
			stringData = map[string]string{"data": data}
		}
	} else {
		stringData["data"] = data
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:        secretName,
			Namespace:   secretSanta.Namespace,
			Labels:      secretSanta.Spec.Labels,
			Annotations: secretSanta.Spec.Annotations,
		},
		Type:       corev1.SecretType(secretSanta.Spec.SecretType),
		StringData: stringData,
	}

	if err := r.Client.Create(ctx, secret); err != nil {
		if updateErr := r.updateStatus(ctx, secretSanta, "SecretCreationFailed", "False", err.Error()); updateErr != nil {
			log.Error(updateErr, "Failed to update status")
		}
		return ctrl.Result{}, err
	}

	RecordSecretGenerated(secretSanta.Name, secretSanta.Namespace, string(secret.Type))
	UpdateSecretInstances(secretSanta.Name, secretSanta.Namespace, 1)
	if updateErr := r.updateStatus(ctx, secretSanta, "Ready", "True", "Secret generated successfully"); updateErr != nil {
		log.Error(updateErr, "Failed to update status")
	}

	log.Info("Secret created successfully", "secretName", secret.Name)
	return ctrl.Result{}, nil
}

func (r *SecretSantaReconciler) generateTemplateData(generatorConfigs []secretsantav1alpha1.GeneratorConfig) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	for _, config := range generatorConfigs {
		if err := r.validateGeneratorConfig(config); err != nil {
			return nil, fmt.Errorf("invalid generator config %s: %w", config.Name, err)
		}

		log := ctrl.Log.WithName("generator").WithValues("name", config.Name, "type", config.Type)
		var gen generators.Generator

		switch config.Type {

		case "tls_private_key":
			gen = &tls.PrivateKeyGenerator{}
		case "tls_self_signed_cert":
			gen = &tls.SelfSignedCertGenerator{}
		case "tls_cert_request":
			gen = &tls.CertRequestGenerator{}
		case "tls_locally_signed_cert":
			gen = &tls.LocallySignedCertGenerator{}

		case "random_password":
			gen = &random.PasswordGenerator{}
		case "random_string":
			gen = &random.StringGenerator{}
		case "random_uuid":
			gen = &random.UUIDGenerator{}
		case "random_integer":
			gen = &random.IntegerGenerator{}
		case "random_bytes":
			gen = &random.BytesGenerator{}
		case "random_id":
			gen = &random.IDGenerator{}

		case "time_static":
			gen = &timegens.StaticGenerator{}

		case "crypto_hmac":
			gen = &crypto.HMACGenerator{}
		case "crypto_aes_key":
			gen = &crypto.AESKeyGenerator{}
		case "crypto_rsa_key":
			gen = &crypto.RSAKeyGenerator{}
		case "crypto_ed25519_key":
			gen = &crypto.ED25519KeyGenerator{}
		case "crypto_chacha20_key":
			gen = &crypto.ChaCha20KeyGenerator{}
		case "crypto_xchacha20_key":
			gen = &crypto.XChaCha20KeyGenerator{}
		case "crypto_ecdsa_key":
			gen = &crypto.ECDSAKeyGenerator{}
		case "crypto_ecdh_key":
			gen = &crypto.ECDHKeyGenerator{}
		default:
			continue
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
