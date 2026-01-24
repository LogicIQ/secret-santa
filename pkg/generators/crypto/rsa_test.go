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
			name: "4096-bit",
			config: map[string]interface{}{
				"key_size": 4096,
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

			keySize, ok := result["key_size"]
			if !ok {
				t.Error("Generate() missing key_size key")
				return
			}
			if keySize != tt.keySize {
				t.Errorf("Generate() key_size = %s, want %s", keySize, tt.keySize)
			}

			algorithm, ok := result["algorithm"]
			if !ok {
				t.Error("Generate() missing algorithm key")
				return
			}
			if algorithm != "RSA" {
				t.Errorf("Generate() algorithm = %s, want RSA", algorithm)
			}

			// Validate PEM format
			privateKeyPem, ok := result["private_key_pem"]
			if !ok {
				t.Error("Generate() missing private_key_pem key")
				return
			}
			privateBlock, _ := pem.Decode([]byte(privateKeyPem))
			if privateBlock == nil || privateBlock.Type != "PRIVATE KEY" {
				t.Error("Generate() invalid private key PEM format")
			}

			publicKeyPem, ok := result["public_key_pem"]
			if !ok {
				t.Error("Generate() missing public_key_pem key")
				return
			}
			publicBlock, _ := pem.Decode([]byte(publicKeyPem))
			if publicBlock == nil || publicBlock.Type != "PUBLIC KEY" {
				t.Error("Generate() invalid public key PEM format")
			}

			// Validate base64 encoding
			privateKeyBase64, ok := result["private_key_base64"]
			if !ok {
				t.Error("Generate() missing private_key_base64 key")
				return
			}
			_, err = base64.StdEncoding.DecodeString(privateKeyBase64)
			if err != nil {
				t.Errorf("Generate() invalid private key base64: %v", err)
			}

			publicKeyBase64, ok := result["public_key_base64"]
			if !ok {
				t.Error("Generate() missing public_key_base64 key")
				return
			}
			_, err = base64.StdEncoding.DecodeString(publicKeyBase64)
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

		privateKeyPem1, ok := result1["private_key_pem"]
		if !ok {
			t.Error("Generate() first result missing private_key_pem key")
			return
		}

		privateKeyPem2, ok := result2["private_key_pem"]
		if !ok {
			t.Error("Generate() second result missing private_key_pem key")
			return
		}

		if privateKeyPem1 == privateKeyPem2 {
			t.Error("Generate() should produce different keys on each call")
		}
	})
}
