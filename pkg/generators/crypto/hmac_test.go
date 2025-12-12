package crypto

import (
	"encoding/base64"
	"encoding/hex"
	"testing"
)

func TestHMACGenerator_Generate(t *testing.T) {
	gen := &HMACGenerator{}

	tests := []struct {
		name      string
		config    map[string]interface{}
		algorithm string
		keySize   int
	}{
		{
			name:      "default sha256",
			config:    map[string]interface{}{},
			algorithm: "sha256",
			keySize:   32,
		},
		{
			name: "sha512",
			config: map[string]interface{}{
				"algorithm": "sha512",
				"key_size":  float64(64),
			},
			algorithm: "sha512",
			keySize:   64,
		},
		{
			name: "with message",
			config: map[string]interface{}{
				"message": "test-payload",
			},
			algorithm: "sha256",
			keySize:   32,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := gen.Generate(tt.config)
			if err != nil {
				t.Errorf("Generate() error = %v", err)
				return
			}

			expectedKeys := []string{"key_base64", "key_hex", "signature_base64", "signature_hex", "algorithm"}
			for _, key := range expectedKeys {
				if _, ok := result[key]; !ok {
					t.Errorf("Generate() missing key %s", key)
				}
			}

			// Validate key length
			keyDecoded, err := base64.StdEncoding.DecodeString(result["key_base64"])
			if err != nil {
				t.Errorf("Generate() invalid key base64: %v", err)
			}
			if len(keyDecoded) != tt.keySize {
				t.Errorf("Generate() key length = %d, want %d", len(keyDecoded), tt.keySize)
			}

			// Validate signature exists
			sigDecoded, err := hex.DecodeString(result["signature_hex"])
			if err != nil {
				t.Errorf("Generate() invalid signature hex: %v", err)
			}
			if len(sigDecoded) == 0 {
				t.Error("Generate() empty signature")
			}

			if result["algorithm"] != tt.algorithm {
				t.Errorf("Generate() algorithm = %s, want %s", result["algorithm"], tt.algorithm)
			}
		})
	}

	// Test invalid algorithm
	t.Run("invalid algorithm", func(t *testing.T) {
		config := map[string]interface{}{
			"algorithm": "md5",
		}
		_, err := gen.Generate(config)
		if err == nil {
			t.Error("Generate() expected error for invalid algorithm")
		}
	})
}
