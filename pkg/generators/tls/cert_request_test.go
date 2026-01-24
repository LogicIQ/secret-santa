package tls

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"
)

func TestCertRequestGenerator_Generate(t *testing.T) {
	gen := &CertRequestGenerator{}

	// Generate a test private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate test private key: %v", err)
	}

	pkcs8Bytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		t.Fatalf("Failed to marshal private key: %v", err)
	}

	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: pkcs8Bytes,
	})

	tests := []struct {
		name   string
		config map[string]interface{}
	}{
		{
			name: "basic CSR",
			config: map[string]interface{}{
				"private_key_pem": string(privateKeyPEM),
				"common_name":     "test.example.com",
			},
		},
		{
			name: "CSR with DNS names",
			config: map[string]interface{}{
				"private_key_pem": string(privateKeyPEM),
				"common_name":     "test.example.com",
				"dns_names":       []interface{}{"test.example.com", "www.test.example.com"},
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

			expectedKeys := []string{"cert_request_pem", "key_algorithm"}
			for _, key := range expectedKeys {
				if _, ok := result[key]; !ok {
					t.Errorf("Generate() missing key %s", key)
				}
			}

			if result["key_algorithm"] != "RSA" {
				t.Errorf("Generate() key_algorithm = %s, want RSA", result["key_algorithm"])
			}

			// Validate CSR PEM format
			certRequestPEM := result["cert_request_pem"]
			block, _ := pem.Decode([]byte(certRequestPEM))
			if block == nil {
				t.Error("Generate() invalid CSR PEM")
			}
			if block.Type != "CERTIFICATE REQUEST" {
				t.Errorf("Generate() wrong PEM type = %s, want CERTIFICATE REQUEST", block.Type)
			}

			// Parse and validate CSR
			csr, err := x509.ParseCertificateRequest(block.Bytes)
			if err != nil {
				t.Errorf("Generate() failed to parse CSR: %v", err)
			}

			commonName, ok := tt.config["common_name"].(string)
			if !ok {
				t.Error("Generate() common_name is not a string")
				return
			}
			if csr.Subject.CommonName != commonName {
				t.Errorf("Generate() CSR common name = %s, want %s", csr.Subject.CommonName, commonName)
			}
		})
	}

	// Test missing private key
	t.Run("missing private key", func(t *testing.T) {
		config := map[string]interface{}{
			"common_name": "test.example.com",
		}
		_, err := gen.Generate(config)
		if err == nil {
			t.Error("Generate() expected error for missing private key")
		}
	})

	// Test invalid private key PEM
	t.Run("invalid private key PEM", func(t *testing.T) {
		config := map[string]interface{}{
			"private_key_pem": "invalid-pem",
			"common_name":     "test.example.com",
		}
		_, err := gen.Generate(config)
		if err == nil {
			t.Error("Generate() expected error for invalid private key PEM")
		}
	})
}
