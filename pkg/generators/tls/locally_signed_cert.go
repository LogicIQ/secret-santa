package tls

import (
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"
)

// LocallySignedCertGenerator generates locally signed certificates
type LocallySignedCertGenerator struct{}

func (g *LocallySignedCertGenerator) Generate(config map[string]interface{}) (map[string]string, error) {
	// Parse CSR
	csrPEM := getStringConfig(config, "cert_request_pem", "")
	if csrPEM == "" {
		return nil, fmt.Errorf("cert_request_pem is required")
	}

	csrBlock, _ := pem.Decode([]byte(csrPEM))
	if csrBlock == nil {
		return nil, fmt.Errorf("failed to decode CSR PEM")
	}

	csr, err := x509.ParseCertificateRequest(csrBlock.Bytes)
	if err != nil {
		return nil, err
	}

	// Parse CA private key
	caPrivateKeyPEM := getStringConfig(config, "ca_private_key_pem", "")
	if caPrivateKeyPEM == "" {
		return nil, fmt.Errorf("ca_private_key_pem is required")
	}

	caKeyBlock, _ := pem.Decode([]byte(caPrivateKeyPEM))
	if caKeyBlock == nil {
		return nil, fmt.Errorf("failed to decode CA private key PEM")
	}

	caPrivateKey, err := x509.ParsePKCS8PrivateKey(caKeyBlock.Bytes)
	if err != nil {
		return nil, err
	}

	// Parse CA certificate
	caCertPEM := getStringConfig(config, "ca_cert_pem", "")
	if caCertPEM == "" {
		return nil, fmt.Errorf("ca_cert_pem is required")
	}

	caCertBlock, _ := pem.Decode([]byte(caCertPEM))
	if caCertBlock == nil {
		return nil, fmt.Errorf("failed to decode CA certificate PEM")
	}

	caCert, err := x509.ParseCertificate(caCertBlock.Bytes)
	if err != nil {
		return nil, err
	}

	// Create certificate template from CSR
	validityHours := getIntConfig(config, "validity_period_hours", 8760) // 1 year default
	template := x509.Certificate{
		SerialNumber:          big.NewInt(time.Now().Unix()),
		Subject:               csr.Subject,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Duration(validityHours) * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:              csr.DNSNames,
		IPAddresses:           csr.IPAddresses,
		URIs:                  csr.URIs,
		BasicConstraintsValid: true,
	}

	// Sign certificate
	certDER, err := x509.CreateCertificate(rand.Reader, &template, caCert, csr.PublicKey, caPrivateKey)
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
		"ca_key_algorithm":    getKeyAlgorithm(caPrivateKey),
		"validity_start_time": template.NotBefore.Format(time.RFC3339),
		"validity_end_time":   template.NotAfter.Format(time.RFC3339),
		"ready_for_renewal":   "false",
	}, nil
}
