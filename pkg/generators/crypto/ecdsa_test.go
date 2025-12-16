package crypto

import (
	"encoding/base64"
	"encoding/pem"
	"testing"
)

func TestECDSAKeyGenerator_Generate(t *testing.T) {
	gen := &ECDSAKeyGenerator{}

	tests := []struct {
		name   string
		config map[string]interface{}
		curve  string
	}{
		{
			name:   "default P256",
			config: map[string]interface{}{},
			curve:  "P256",
		},
		{
			name: "P224",
			config: map[string]interface{}{
				"curve": "P224",
			},
			curve: "P224",
		},
		{
			name: "P384",
			config: map[string]interface{}{
				"curve": "P384",
			},
			curve: "P384",
		},
		{
			name: "P521",
			config: map[string]interface{}{
				"curve": "P521",
			},
			curve: "P521",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := gen.Generate(tt.config)
			if err != nil {
				t.Errorf("Generate() error = %v", err)
				return
			}

			expectedKeys := []string{"private_key_pem", "public_key_pem", "private_key_base64", "public_key_base64", "curve", "algorithm"}
			for _, key := range expectedKeys {
				if _, ok := result[key]; !ok {
					t.Errorf("Generate() missing key %s", key)
				}
			}

			if result["curve"] != tt.curve {
				t.Errorf("Generate() curve = %s, want %s", result["curve"], tt.curve)
			}

			if result["algorithm"] != "ECDSA" {
				t.Errorf("Generate() algorithm = %s, want ECDSA", result["algorithm"])
			}

			// Validate PEM format
			privateBlock, _ := pem.Decode([]byte(result["private_key_pem"]))
			if privateBlock == nil || privateBlock.Type != "PRIVATE KEY" {
				t.Error("Generate() invalid private key PEM format")
			}

			publicBlock, _ := pem.Decode([]byte(result["public_key_pem"]))
			if publicBlock == nil || publicBlock.Type != "PUBLIC KEY" {
				t.Error("Generate() invalid public key PEM format")
			}

			// Validate base64 encoding
			_, err = base64.StdEncoding.DecodeString(result["private_key_base64"])
			if err != nil {
				t.Errorf("Generate() invalid private key base64: %v", err)
			}

			_, err = base64.StdEncoding.DecodeString(result["public_key_base64"])
			if err != nil {
				t.Errorf("Generate() invalid public key base64: %v", err)
			}
		})
	}

	// Test invalid curve
	t.Run("invalid curve", func(t *testing.T) {
		config := map[string]interface{}{
			"curve": "INVALID",
		}
		_, err := gen.Generate(config)
		if err == nil {
			t.Error("Generate() expected error for invalid curve")
		}
	})
}
