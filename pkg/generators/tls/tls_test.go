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
			certPEM := result["cert_pem"]
			if certPEM == "" {
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

			cert, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				t.Errorf("Failed to parse certificate: %v", err)
				return
			}

			// Verify subject fields if provided
			if org, ok := tt.config["organization"].([]interface{}); ok && len(org) > 0 {
				if orgStr, ok := org[0].(string); ok {
					if len(cert.Subject.Organization) == 0 || cert.Subject.Organization[0] != orgStr {
						t.Error("Certificate missing or incorrect Organization")
					}
				}
			}
			if country, ok := tt.config["country"].([]interface{}); ok && len(country) > 0 {
				if countryStr, ok := country[0].(string); ok {
					if len(cert.Subject.Country) == 0 || cert.Subject.Country[0] != countryStr {
						t.Error("Certificate missing or incorrect Country")
					}
				}
			}

			// Verify DNS names if provided
			if dnsNames, ok := tt.config["dns_names"].([]interface{}); ok {
				if len(cert.DNSNames) != len(dnsNames) {
					t.Errorf("Certificate has %d DNS names, want %d", len(cert.DNSNames), len(dnsNames))
				}
				for i, dns := range dnsNames {
					if dnsStr, ok := dns.(string); ok {
						if i >= len(cert.DNSNames) || cert.DNSNames[i] != dnsStr {
							t.Errorf("Certificate DNS name[%d] = %v, want %s", i, cert.DNSNames[i], dnsStr)
						}
					}
				}
			}
		})
	}
}
