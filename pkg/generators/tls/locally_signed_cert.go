package tls

import (
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"
)

type LocallySignedCertGenerator struct{}

func (g *LocallySignedCertGenerator) Generate(config map[string]interface{}) (map[string]string, error) {
	// Parse CSR
	csrPEM := getStringConfig(config, "cert_request_pem", "")
	if csrPEM == "" {
		return nil, fmt.Errorf("cert_request_pem is required")
	}

	csrBlock, rest := pem.Decode([]byte(csrPEM))
	if csrBlock == nil {
		return nil, fmt.Errorf("failed to decode CSR PEM")
	}
	if len(rest) > 0 {
		return nil, fmt.Errorf("CSR PEM contains extra data")
	}

	csr, err := x509.ParseCertificateRequest(csrBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate request: %w", err)
	}

	// Validate CSR signature
	if err := csr.CheckSignature(); err != nil {
		return nil, fmt.Errorf("invalid CSR signature: %w", err)
	}

	// Parse CA private key
	caPrivateKeyPEM := getStringConfig(config, "ca_private_key_pem", "")
	if caPrivateKeyPEM == "" {
		return nil, fmt.Errorf("ca_private_key_pem is required")
	}

	caKeyBlock, rest := pem.Decode([]byte(caPrivateKeyPEM))
	if caKeyBlock == nil {
		return nil, fmt.Errorf("failed to decode CA private key PEM")
	}
	if len(rest) > 0 {
		return nil, fmt.Errorf("CA private key PEM contains extra data")
	}

	caPrivateKey, err := x509.ParsePKCS8PrivateKey(caKeyBlock.Bytes)
	if err != nil {
		// Try parsing as PKCS1 RSA key if PKCS8 fails
		rsaKey, rsaErr := x509.ParsePKCS1PrivateKey(caKeyBlock.Bytes)
		if rsaErr != nil {
			return nil, fmt.Errorf("failed to parse CA private key: %w", rsaErr)
		}
		caPrivateKey = rsaKey
	}

	// Parse CA certificate
	caCertPEM := getStringConfig(config, "ca_cert_pem", "")
	if caCertPEM == "" {
		return nil, fmt.Errorf("ca_cert_pem is required")
	}

	caCertBlock, rest := pem.Decode([]byte(caCertPEM))
	if caCertBlock == nil {
		return nil, fmt.Errorf("failed to decode CA certificate PEM")
	}
	if len(rest) > 0 {
		return nil, fmt.Errorf("CA certificate PEM contains extra data")
	}

	caCert, err := x509.ParseCertificate(caCertBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse CA certificate: %w", err)
	}

	// Validate CA private key matches CA certificate
	if !publicKeysMatch(caPrivateKey, caCert.PublicKey) {
		return nil, fmt.Errorf("CA private key does not match CA certificate public key")
	}

	// Create certificate template from CSR
	validityHours := getIntConfig(config, "validity_period_hours", 8760) // 1 year default
	if validityHours <= 0 {
		return nil, fmt.Errorf("validity_period_hours must be positive, got %d", validityHours)
	}
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, fmt.Errorf("failed to generate serial number: %w", err)
	}
	template := x509.Certificate{
		SerialNumber:          serialNumber,
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
		return nil, fmt.Errorf("failed to create certificate: %w", err)
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
