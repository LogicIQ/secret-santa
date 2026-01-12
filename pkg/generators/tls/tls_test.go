package tls

import (
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := gen.Generate(tt.config)
			if err != nil {
				t.Errorf("Generate() error = %v", err)
				return
			}

			expectedKeys := []string{"cert_pem", "key_algorithm", "validity_start_time", "validity_end_time"}
			for _, key := range expectedKeys {
				if _, ok := result[key]; !ok {
					t.Errorf("Generate() missing key %s", key)
				}
			}

			// Validate PEM format
			block, rest := pem.Decode([]byte(result["cert_pem"]))
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
		})
	}
}
