package tls

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"
)

type SelfSignedCertGenerator struct{}

func (g *SelfSignedCertGenerator) Generate(config map[string]interface{}) (map[string]string, error) {
	// TODO: Add support for ECDSA and Ed25519 key algorithms through key_algorithm config parameter
	// Generate private key
	keySize := getIntConfig(config, "key_size", 2048)
	privateKey, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		return nil, err
	}

	// Create certificate template
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, err
	}

	validityDays := getIntConfig(config, "validity_days", 365)
	notBefore := time.Now()
	notAfter := notBefore.Add(time.Duration(validityDays) * 24 * time.Hour)

	dnsNames := getStringSliceConfig(config, "dns_names")
	if len(dnsNames) == 0 {
		dnsNames = []string{getStringConfig(config, "common_name", "localhost")}
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:         getStringConfig(config, "common_name", "localhost"),
			Organization:       getStringSliceConfig(config, "organization"),
			OrganizationalUnit: getStringSliceConfig(config, "organizational_unit"),
			Country:            getStringSliceConfig(config, "country"),
			Province:           getStringSliceConfig(config, "province"),
			Locality:           getStringSliceConfig(config, "locality"),
		},
		NotBefore:   notBefore,
		NotAfter:    notAfter,
		KeyUsage:    x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:    dnsNames,
	}

	// Create certificate
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, err
	}

	// Encode certificate
	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	})
	if certPEM == nil {
		return nil, fmt.Errorf("failed to encode certificate to PEM")
	}

	// Encode private key
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	})
	if privateKeyPEM == nil {
		return nil, fmt.Errorf("failed to encode private key to PEM")
	}

	return map[string]string{
		"cert_pem":            string(certPEM),
		"private_key_pem":     string(privateKeyPEM),
		"key_algorithm":       "RSA",
		"validity_start_time": template.NotBefore.Format(time.RFC3339),
		"validity_end_time":   template.NotAfter.Format(time.RFC3339),
		"ready_for_renewal":   "false",
	}, nil
}
