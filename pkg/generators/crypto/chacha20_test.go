package crypto

import (
	"encoding/base64"
	"encoding/hex"
	"testing"
)

type keyGenerator interface {
	Generate(config map[string]interface{}) (map[string]string, error)
}

func TestChaCha20Generators(t *testing.T) {
	tests := []struct {
		name      string
		generator keyGenerator
		algorithm string
	}{
		{
			name:      "ChaCha20",
			generator: &ChaCha20KeyGenerator{},
			algorithm: "ChaCha20",
		},
		{
			name:      "XChaCha20",
			generator: &XChaCha20KeyGenerator{},
			algorithm: "XChaCha20",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.generator.Generate(map[string]interface{}{})
			if err != nil {
				t.Errorf("Generate() error = %v", err)
				return
			}

			expectedKeys := []string{"key_base64", "key_hex", "key_size", "algorithm"}
			for _, key := range expectedKeys {
				if _, ok := result[key]; !ok {
					t.Errorf("Generate() missing key %s", key)
				}
			}

			if result["key_size"] != "256" {
				t.Errorf("Generate() key_size = %s, want 256", result["key_size"])
				return
			}

			if result["algorithm"] != tt.algorithm {
				t.Errorf("Generate() algorithm = %s, want %s", result["algorithm"], tt.algorithm)
				return
			}

			decoded, err := base64.StdEncoding.DecodeString(result["key_base64"])
			if err != nil {
				t.Errorf("Generate() invalid base64: %v", err)
				return
			}
			if len(decoded) != 32 {
				t.Errorf("Generate() key length = %d, want 32", len(decoded))
			}

			hexDecoded, err := hex.DecodeString(result["key_hex"])
			if err != nil {
				t.Errorf("Generate() invalid hex: %v", err)
				return
			}
			if len(hexDecoded) != 32 {
				t.Errorf("Generate() hex key length = %d, want 32", len(hexDecoded))
			}
		})
	}
}
