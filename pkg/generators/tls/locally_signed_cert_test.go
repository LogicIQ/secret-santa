package tls

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"testing"
	"time"
)

func TestLocallySignedCertGenerator_Generate(t *testing.T) {
	gen := &LocallySignedCertGenerator{}

	// Generate CA key pair
	caPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate CA private key: %v", err)
	}

	// Create CA certificate
	caTemplate := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "Test CA",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	caCertDER, err := x509.CreateCertificate(rand.Reader, &caTemplate, &caTemplate, &caPrivateKey.PublicKey, caPrivateKey)
	if err != nil {
		t.Fatalf("Failed to create CA certificate: %v", err)
	}

	caCertPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caCertDER,
	})

	caPrivateKeyBytes, err := x509.MarshalPKCS8PrivateKey(caPrivateKey)
	if err != nil {
		t.Fatalf("Failed to marshal CA private key: %v", err)
	}

	caPrivateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: caPrivateKeyBytes,
	})

	// Generate CSR key pair
	csrPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate CSR private key: %v", err)
	}

	// Create CSR
	csrTemplate := x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName: "test.example.com",
		},
		DNSNames: []string{"test.example.com", "www.test.example.com"},
	}

	csrDER, err := x509.CreateCertificateRequest(rand.Reader, &csrTemplate, csrPrivateKey)
	if err != nil {
		t.Fatalf("Failed to create CSR: %v", err)
	}

	csrPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE REQUEST",
		Bytes: csrDER,
	})

	tests := []struct {
		name   string
		config map[string]interface{}
	}{
		{
			name: "basic locally signed cert",
			config: map[string]interface{}{
				"cert_request_pem":   string(csrPEM),
				"ca_private_key_pem": string(caPrivateKeyPEM),
				"ca_cert_pem":        string(caCertPEM),
			},
		},
		{
			name: "custom validity period",
			config: map[string]interface{}{
				"cert_request_pem":      string(csrPEM),
				"ca_private_key_pem":    string(caPrivateKeyPEM),
				"ca_cert_pem":           string(caCertPEM),
				"validity_period_hours": float64(720), // 30 days
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := gen.Generate(tt.config)
			if err != nil {
				t.Errorf("Generate() error = %v", err)
				return
			}

			expectedKeys := []string{"cert_pem", "ca_key_algorithm", "validity_start_time", "validity_end_time", "ready_for_renewal"}
			for _, key := range expectedKeys {
				if _, ok := result[key]; !ok {
					t.Errorf("Generate() missing key %s", key)
				}
			}

			if result["ca_key_algorithm"] != "RSA" {
				t.Errorf("Generate() ca_key_algorithm = %s, want RSA", result["ca_key_algorithm"])
			}

			if result["ready_for_renewal"] != "false" {
				t.Errorf("Generate() ready_for_renewal = %s, want false", result["ready_for_renewal"])
			}

			// Validate certificate PEM format
			block, _ := pem.Decode([]byte(result["cert_pem"]))
			if block == nil {
				t.Error("Generate() invalid certificate PEM")
			}
			if block.Type != "CERTIFICATE" {
				t.Errorf("Generate() wrong PEM type = %s, want CERTIFICATE", block.Type)
			}

			// Parse and validate certificate
			cert, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				t.Errorf("Generate() failed to parse certificate: %v", err)
			}

			if cert.Subject.CommonName != "test.example.com" {
				t.Errorf("Generate() certificate common name = %s, want test.example.com", cert.Subject.CommonName)
			}

			// Validate certificate is signed by CA
			caCert, err := x509.ParseCertificate(caCertDER)
			if err != nil {
				t.Fatalf("Failed to parse CA certificate: %v", err)
			}

			err = cert.CheckSignatureFrom(caCert)
			if err != nil {
				t.Errorf("Generate() certificate not signed by CA: %v", err)
			}
		})
	}

	// Test missing CSR
	t.Run("missing CSR", func(t *testing.T) {
		config := map[string]interface{}{
			"ca_private_key_pem": string(caPrivateKeyPEM),
			"ca_cert_pem":        string(caCertPEM),
		}
		_, err := gen.Generate(config)
		if err == nil {
			t.Error("Generate() expected error for missing CSR")
		}
	})

	// Test missing CA private key
	t.Run("missing CA private key", func(t *testing.T) {
		config := map[string]interface{}{
			"cert_request_pem": string(csrPEM),
			"ca_cert_pem":      string(caCertPEM),
		}
		_, err := gen.Generate(config)
		if err == nil {
			t.Error("Generate() expected error for missing CA private key")
		}
	})

	// Test missing CA certificate
	t.Run("missing CA certificate", func(t *testing.T) {
		config := map[string]interface{}{
			"cert_request_pem":   string(csrPEM),
			"ca_private_key_pem": string(caPrivateKeyPEM),
		}
		_, err := gen.Generate(config)
		if err == nil {
			t.Error("Generate() expected error for missing CA certificate")
		}
	})
}
