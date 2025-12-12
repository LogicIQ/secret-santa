package controller

import (
	"bytes"
	"context"
	"text/template"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	secretsantav1alpha1 "github.com/logicIQ/secret-santa/api/v1alpha1"
	"github.com/logicIQ/secret-santa/pkg/generators"
	"github.com/logicIQ/secret-santa/pkg/generators/crypto"
	"github.com/logicIQ/secret-santa/pkg/generators/random"
	timegens "github.com/logicIQ/secret-santa/pkg/generators/time"
	"github.com/logicIQ/secret-santa/pkg/generators/tls"
	"github.com/logicIQ/secret-santa/pkg/metrics"
	tmplpkg "github.com/logicIQ/secret-santa/pkg/template"
	"time"
)

// SecretSantaReconciler reconciles a SecretSanta object
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
	log.Info("Starting reconcile", "time", start.Format(time.RFC3339), "dryRun", r.DryRun)

	timer := metrics.NewReconcileTimer(req.Namespace, req.Name)
	defer func() {
		duration := time.Since(start)
		timer.ObserveDuration()
		metrics.RecordReconcileComplete(req.Namespace, req.Name, duration.Seconds())
		log.Info("Completed reconcile", "duration", duration)
	}()

	var secretSanta secretsantav1alpha1.SecretSanta
	if err := r.Get(ctx, req.NamespacedName, &secretSanta); err != nil {
		if errors.IsNotFound(err) {
			log.V(1).Info("Resource not found, skipping", "resource", req.NamespacedName)
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get SecretSanta resource")
		return ctrl.Result{}, err
	}


	log.V(1).Info("Found SecretSanta resource", "phase", "processing", "secretType", secretSanta.Spec.SecretType)

	if !r.shouldProcess(&secretSanta) {
		log.Info("Skipping resource due to annotation/label filters", "includeAnnotations", r.IncludeAnnotations, "excludeAnnotations", r.ExcludeAnnotations, "includeLabels", r.IncludeLabels, "excludeLabels", r.ExcludeLabels)
		return ctrl.Result{}, nil
	}

	log.Info("Processing SecretSanta resource", "generators", len(secretSanta.Spec.Generators), "templateLength", len(secretSanta.Spec.Template))


	log.Info("Generating template data", "generatorCount", len(secretSanta.Spec.Generators))
	templateData, err := r.generateTemplateData(secretSanta.Spec.Generators)
	if err != nil {
		log.Error(err, "Failed to generate template data")
		metrics.RecordReconcileError(secretSanta.Namespace, secretSanta.Name, "generator_failed")
		return ctrl.Result{}, err
	}
	log.V(1).Info("Template data generated successfully", "dataKeys", getMapKeys(templateData), "generatorTypes", getGeneratorTypes(secretSanta.Spec.Generators))


	if err := r.validateTemplate(secretSanta.Spec.Template); err != nil {
		log.Error(err, "Template validation failed")
		metrics.RecordTemplateValidationError(secretSanta.Namespace, secretSanta.Name)
		metrics.RecordReconcileError(secretSanta.Namespace, secretSanta.Name, "template_validation")
		if !r.DryRun {
			r.updateStatus(ctx, &secretSanta, "TemplateFailed", "False", err.Error())
		}
		return ctrl.Result{}, err
	}


	tmpl, err := template.New("secret").Funcs(tmplpkg.FuncMap()).Parse(secretSanta.Spec.Template)
	if err != nil {
		log.Error(err, "Failed to parse template")
		if !r.DryRun {
			r.updateStatus(ctx, &secretSanta, "TemplateFailed", "False", err.Error())
		}
		return ctrl.Result{}, err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, templateData); err != nil {
		log.Error(err, "Failed to execute template")
		if !r.DryRun {
			r.updateStatus(ctx, &secretSanta, "TemplateExecutionFailed", "False", err.Error())
		}
		return ctrl.Result{}, err
	}


	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:        secretSanta.Name,
			Namespace:   secretSanta.Namespace,
			Labels:      secretSanta.Spec.Labels,
			Annotations: secretSanta.Spec.Annotations,
		},
		Type: corev1.SecretType(secretSanta.Spec.SecretType),
		StringData: map[string]string{
			"data": buf.String(),
		},
	}

	if err := ctrl.SetControllerReference(&secretSanta, secret, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}

	if r.DryRun {
		log.Info("DRY RUN: Would create secret", "secretName", secret.Name, "secretType", secret.Type, "dataSize", len(buf.String()))
		log.V(1).Info("DRY RUN: Template execution successful", "template", secretSanta.Spec.Template, "generatedDataPreview", truncateString(buf.String(), 100))
		return ctrl.Result{}, nil
	}

	if err := r.Client.Create(ctx, secret); err != nil {
		if errors.IsAlreadyExists(err) {

			log.Info("Secret already exists - create-once policy enforced", "secret", secret.Name, "secretType", secret.Type)
			metrics.RecordSecretSkipped(secretSanta.Namespace, secretSanta.Name)
			if !r.DryRun {
				r.updateStatus(ctx, &secretSanta, "Ready", "True", "Secret already exists")
			}
			return ctrl.Result{}, nil
		}
		if !r.DryRun {
			r.updateStatus(ctx, &secretSanta, "SecretCreationFailed", "False", err.Error())
		}
		return ctrl.Result{}, err
	}


	if !r.DryRun {
		metrics.RecordSecretGenerated(secretSanta.Namespace, secretSanta.Name, string(secret.Type))
		metrics.UpdateSecretInstances(secretSanta.Namespace, secretSanta.Name, 1)
		r.updateStatus(ctx, &secretSanta, "Ready", "True", "Secret generated successfully")
	}
	log.Info("Secret creation completed successfully", "secretName", secret.Name, "secretType", secret.Type, "dataSize", len(buf.String()))
	return ctrl.Result{}, nil
}

func (r *SecretSantaReconciler) generateTemplateData(generatorConfigs []secretsantav1alpha1.GeneratorConfig) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	for _, config := range generatorConfigs {
		log := ctrl.Log.WithName("generator").WithValues("name", config.Name, "type", config.Type)
		log.V(1).Info("Processing generator", "config", config.Config)
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

		timer := metrics.NewGeneratorTimer(config.Type)
		result, err := gen.Generate(config.Config)
		timer.ObserveDuration()
		if err != nil {
			log.Error(err, "Generator failed")
			metrics.RecordGeneratorExecution(config.Type, "error")
			return nil, err
		}
		log.V(1).Info("Generator completed successfully")
		metrics.RecordGeneratorExecution(config.Type, "success")
		data[config.Name] = result
	}

	return data, nil
}

func (r *SecretSantaReconciler) validateTemplate(tmplStr string) error {

	tmpl, err := template.New("validation").Funcs(tmplpkg.FuncMap()).Parse(tmplStr)
	if err != nil {
		return err
	}


	var buf bytes.Buffer
	emptyData := make(map[string]interface{})
	if err := tmpl.Execute(&buf, emptyData); err != nil {

		return nil
	}
	return nil
}

func (r *SecretSantaReconciler) shouldProcess(secretSanta *secretsantav1alpha1.SecretSanta) bool {

	if len(r.IncludeAnnotations) > 0 {
		found := false
		for _, include := range r.IncludeAnnotations {
			if _, exists := secretSanta.Annotations[include]; exists {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	for _, exclude := range r.ExcludeAnnotations {
		if _, exists := secretSanta.Annotations[exclude]; exists {
			return false
		}
	}


	if len(r.IncludeLabels) > 0 {
		found := false
		for _, include := range r.IncludeLabels {
			if _, exists := secretSanta.Labels[include]; exists {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	for _, exclude := range r.ExcludeLabels {
		if _, exists := secretSanta.Labels[exclude]; exists {
			return false
		}
	}

	return true
}

func (r *SecretSantaReconciler) updateStatus(ctx context.Context, secretSanta *secretsantav1alpha1.SecretSanta, conditionType, status, message string) {
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

	r.Status().Update(ctx, secretSanta)
}

func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func getGeneratorTypes(generators []secretsantav1alpha1.GeneratorConfig) []string {
	types := make([]string, 0, len(generators))
	for _, gen := range generators {
		types = append(types, gen.Type)
	}
	return types
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func (r *SecretSantaReconciler) SetupWithManager(mgr ctrl.Manager, maxConcurrentReconciles int) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&secretsantav1alpha1.SecretSanta{}).
		Owns(&corev1.Secret{}).
		WithOptions(ctrl.Options{
			MaxConcurrentReconciles: maxConcurrentReconciles,
		}).
		Complete(r)
}
