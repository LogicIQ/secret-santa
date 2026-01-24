package tls

import (
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
)

type CertRequestGenerator struct{}

func (g *CertRequestGenerator) Generate(config map[string]interface{}) (map[string]string, error) {
	// Parse private key from config
	privateKeyPEM := getStringConfig(config, "private_key_pem", "")
	if privateKeyPEM == "" {
		return nil, fmt.Errorf("private_key_pem is required")
	}

	block, rest := pem.Decode([]byte(privateKeyPEM))
	if block == nil {
		return nil, fmt.Errorf("failed to decode private key PEM")
	}
	if len(rest) > 0 {
		return nil, fmt.Errorf("private key PEM contains extra data")
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		// Try PKCS1 for RSA keys
		rsaKey, rsaErr := x509.ParsePKCS1PrivateKey(block.Bytes)
		if rsaErr == nil {
			privateKey = rsaKey
		} else {
			// Try SEC1 for EC keys
			ecKey, ecErr := x509.ParseECPrivateKey(block.Bytes)
			if ecErr == nil {
				privateKey = ecKey
			} else {
				return nil, fmt.Errorf("failed to parse private key: PKCS8 error: %v, PKCS1 error: %v, EC error: %v", err, rsaErr, ecErr)
			}
		}
	}

	// Create CSR template
	template := x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName: getStringConfig(config, "common_name", ""),
		},
		DNSNames: getStringSliceConfig(config, "dns_names"),
	}

	// Generate CSR
	csrDER, err := x509.CreateCertificateRequest(rand.Reader, &template, privateKey)
	if err != nil {
		return nil, err
	}

	// Encode CSR
	csrPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE REQUEST",
		Bytes: csrDER,
	})

	return map[string]string{
		"cert_request_pem": string(csrPEM),
		"key_algorithm":    getKeyAlgorithm(privateKey),
	}, nil
}
