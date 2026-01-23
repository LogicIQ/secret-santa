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
	const expectedKeyBytes = 32

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

			const expectedKeySize = "256"
			if result["key_size"] != expectedKeySize {
				t.Errorf("Generate() key_size = %s, want %s", result["key_size"], expectedKeySize)
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
			if len(decoded) != expectedKeyBytes {
				t.Errorf("Generate() key length = %d, want %d", len(decoded), expectedKeyBytes)
			}

			hexDecoded, err := hex.DecodeString(result["key_hex"])
			if err != nil {
				t.Errorf("Generate() invalid hex: %v", err)
				return
			}
			if len(hexDecoded) != expectedKeyBytes {
				t.Errorf("Generate() hex key length = %d, want %d", len(hexDecoded), expectedKeyBytes)
			}
		})
	}
}
