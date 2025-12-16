package crypto

import (
	"encoding/base64"
	"encoding/pem"
	"testing"
)

func TestRSAKeyGenerator_Generate(t *testing.T) {
	gen := &RSAKeyGenerator{}

	tests := []struct {
		name    string
		config  map[string]interface{}
		keySize string
	}{
		{
			name:    "default 2048-bit",
			config:  map[string]interface{}{},
			keySize: "2048",
		},
		{
			name: "1024-bit",
			config: map[string]interface{}{
				"key_size": float64(1024),
			},
			keySize: "1024",
		},
		{
			name: "4096-bit",
			config: map[string]interface{}{
				"key_size": float64(4096),
			},
			keySize: "4096",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := gen.Generate(tt.config)
			if err != nil {
				t.Errorf("Generate() error = %v", err)
				return
			}

			expectedKeys := []string{"private_key_pem", "public_key_pem", "private_key_base64", "public_key_base64", "key_size", "algorithm"}
			for _, key := range expectedKeys {
				if _, ok := result[key]; !ok {
					t.Errorf("Generate() missing key %s", key)
				}
			}

			if result["key_size"] != tt.keySize {
				t.Errorf("Generate() key_size = %s, want %s", result["key_size"], tt.keySize)
			}

			if result["algorithm"] != "RSA" {
				t.Errorf("Generate() algorithm = %s, want RSA", result["algorithm"])
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

	// Test multiple generations produce different keys
	t.Run("unique keys", func(t *testing.T) {
		result1, err := gen.Generate(map[string]interface{}{})
		if err != nil {
			t.Errorf("Generate() first call error = %v", err)
			return
		}

		result2, err := gen.Generate(map[string]interface{}{})
		if err != nil {
			t.Errorf("Generate() second call error = %v", err)
			return
		}

		if result1["private_key_pem"] == result2["private_key_pem"] {
			t.Error("Generate() should produce different keys on each call")
		}
	})
}
