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
	tmplpkg "github.com/logicIQ/secret-santa/pkg/template"
)

// SecretSantaReconciler reconciles a SecretSanta object
type SecretSantaReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *SecretSantaReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Fetch SecretSanta instance
	var secretSanta secretsantav1alpha1.SecretSanta
	if err := r.Get(ctx, req.NamespacedName, &secretSanta); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// Generate template context
	templateData, err := r.generateTemplateData(secretSanta.Spec.Generators)
	if err != nil {
		log.Error(err, "Failed to generate template data")
		return ctrl.Result{}, err
	}

	// Parse and execute template with custom functions
	tmpl, err := template.New("secret").Funcs(tmplpkg.FuncMap()).Parse(secretSanta.Spec.Template)
	if err != nil {
		log.Error(err, "Failed to parse template")
		return ctrl.Result{}, err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, templateData); err != nil {
		log.Error(err, "Failed to execute template")
		return ctrl.Result{}, err
	}

	// Create or update Kubernetes secret
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

	if err := r.Client.Create(ctx, secret); err != nil {
		if errors.IsAlreadyExists(err) {
			if err := r.Client.Update(ctx, secret); err != nil {
				return ctrl.Result{}, err
			}
		} else {
			return ctrl.Result{}, err
		}
	}

	log.Info("Secret created/updated", "secret", secret.Name)
	return ctrl.Result{}, nil
}

func (r *SecretSantaReconciler) generateTemplateData(generatorConfigs []secretsantav1alpha1.GeneratorConfig) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	for _, config := range generatorConfigs {
		var gen generators.Generator

		switch config.Type {
		// TLS generators
		case "tls_private_key":
			gen = &tls.PrivateKeyGenerator{}
		case "tls_self_signed_cert":
			gen = &tls.SelfSignedCertGenerator{}
		case "tls_cert_request":
			gen = &tls.CertRequestGenerator{}
		case "tls_locally_signed_cert":
			gen = &tls.LocallySignedCertGenerator{}
		// Random generators
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
		// Time generators
		case "time_static":
			gen = &timegens.StaticGenerator{}
		// Crypto generators
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

		result, err := gen.Generate(config.Config)
		if err != nil {
			return nil, err
		}

		data[config.Name] = result
	}

	return data, nil
}

func (r *SecretSantaReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&secretsantav1alpha1.SecretSanta{}).
		Owns(&corev1.Secret{}).
		Complete(r)
}
