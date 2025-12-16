package tls

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"time"
)

type SelfSignedCertGenerator struct{}

func (g *SelfSignedCertGenerator) Generate(config map[string]interface{}) (map[string]string, error) {
	// Generate private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	// Create certificate template
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: getStringConfig(config, "common_name", "localhost"),
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:    x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:    []string{getStringConfig(config, "common_name", "localhost")},
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

	return map[string]string{
		"cert_pem":            string(certPEM),
		"key_algorithm":       "RSA",
		"validity_start_time": template.NotBefore.Format(time.RFC3339),
		"validity_end_time":   template.NotAfter.Format(time.RFC3339),
		"ready_for_renewal":   "false",
	}, nil
}
