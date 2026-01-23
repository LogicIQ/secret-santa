package tls

import (
	"crypto/x509"
	"encoding/pem"
	"testing"
)

func TestSelfSignedCertGenerator_Generate(t *testing.T) {
	gen := &SelfSignedCertGenerator{}

	tests := []struct {
		name   string
		config map[string]interface{}
	}{
		{
			name:   "default config",
			config: map[string]interface{}{},
		},
		{
			name: "custom common name",
			config: map[string]interface{}{
				"common_name": "test.example.com",
			},
		},
		{
			name: "with subject fields",
			config: map[string]interface{}{
				"common_name":         "test.example.com",
				"organization":        []interface{}{"Test Org"},
				"organizational_unit": []interface{}{"Test Unit"},
				"country":             []interface{}{"US"},
				"province":            []interface{}{"California"},
				"locality":            []interface{}{"San Francisco"},
			},
		},
		{
			name: "with multiple dns names",
			config: map[string]interface{}{
				"common_name": "example.com",
				"dns_names":   []interface{}{"example.com", "www.example.com", "api.example.com"},
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

			expectedKeys := []string{"cert_pem", "private_key_pem", "key_algorithm", "validity_start_time", "validity_end_time"}
			for _, key := range expectedKeys {
				if _, ok := result[key]; !ok {
					t.Errorf("Generate() missing key %s", key)
				}
			}

			// Validate PEM format
			certPEM, ok := result["cert_pem"]
			if !ok || certPEM == "" {
				t.Error("Generate() missing or empty cert_pem")
				return
			}
			block, rest := pem.Decode([]byte(certPEM))
			if block == nil {
				t.Error("Generate() invalid certificate PEM")
				return
			}
			if len(rest) > 0 {
				t.Error("Generate() certificate PEM contains extra data")
			}
			if block.Type != "CERTIFICATE" {
				t.Errorf("Generate() wrong PEM type = %s, want CERTIFICATE", block.Type)
			}

			// Verify subject fields if provided
			if tt.name == "with subject fields" {
				cert, err := x509.ParseCertificate(block.Bytes)
				if err != nil {
					t.Errorf("Failed to parse certificate: %v", err)
					return
				}
				if len(cert.Subject.Organization) == 0 || cert.Subject.Organization[0] != "Test Org" {
					t.Error("Certificate missing or incorrect Organization")
				}
				if len(cert.Subject.Country) == 0 || cert.Subject.Country[0] != "US" {
					t.Error("Certificate missing or incorrect Country")
				}
			}

			// Verify multiple DNS names if provided
			if tt.name == "with multiple dns names" {
				cert, err := x509.ParseCertificate(block.Bytes)
				if err != nil {
					t.Errorf("Failed to parse certificate: %v", err)
					return
				}
				expectedDNS := []string{"example.com", "www.example.com", "api.example.com"}
				if len(cert.DNSNames) != len(expectedDNS) {
					t.Errorf("Certificate has %d DNS names, want %d", len(cert.DNSNames), len(expectedDNS))
				}
				for i, dns := range expectedDNS {
					if i >= len(cert.DNSNames) || cert.DNSNames[i] != dns {
						t.Errorf("Certificate DNS name[%d] = %v, want %s", i, cert.DNSNames[i], dns)
					}
				}
			}
		})
	}
}
