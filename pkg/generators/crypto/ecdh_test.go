package crypto

import (
	"encoding/base64"
	"encoding/pem"
	"testing"
)

func TestECDHKeyGenerator_Generate(t *testing.T) {
	gen := &ECDHKeyGenerator{}

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
		{
			name: "X25519",
			config: map[string]interface{}{
				"curve": "X25519",
			},
			curve: "X25519",
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

			curve := result["curve"]
			if curve != tt.curve {
				t.Errorf("Generate() curve = %s, want %s", curve, tt.curve)
			}

			algorithm := result["algorithm"]
			if algorithm != "ECDH" {
				t.Errorf("Generate() algorithm = %s, want ECDH", algorithm)
			}

			// Validate PEM format
			privateKeyPem := result["private_key_pem"]
			privateBlock, _ := pem.Decode([]byte(privateKeyPem))
			if privateBlock == nil || privateBlock.Type != "PRIVATE KEY" {
				t.Error("Generate() invalid private key PEM format")
			}

			publicKeyPem := result["public_key_pem"]
			publicBlock, _ := pem.Decode([]byte(publicKeyPem))
			if publicBlock == nil || publicBlock.Type != "PUBLIC KEY" {
				t.Error("Generate() invalid public key PEM format")
			}

			// Validate base64 encoding
			privateKeyBase64 := result["private_key_base64"]
			_, err = base64.StdEncoding.DecodeString(privateKeyBase64)
			if err != nil {
				t.Errorf("Generate() invalid private key base64: %v", err)
			}

			publicKeyBase64 := result["public_key_base64"]
			_, err = base64.StdEncoding.DecodeString(publicKeyBase64)
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
