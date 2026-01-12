package crypto

import (
	"encoding/base64"
	"encoding/hex"
	"testing"
)

func TestAESKeyGenerator_Generate(t *testing.T) {
	gen := &AESKeyGenerator{}

	tests := []struct {
		name    string
		config  map[string]interface{}
		keySize int
		bytes   int
	}{
		{
			name:    "default 256-bit",
			config:  map[string]interface{}{},
			keySize: 256,
			bytes:   32,
		},
		{
			name: "128-bit",
			config: map[string]interface{}{
				"key_size": float64(128),
			},
			keySize: 128,
			bytes:   16,
		},
		{
			name: "192-bit",
			config: map[string]interface{}{
				"key_size": float64(192),
			},
			keySize: 192,
			bytes:   24,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := gen.Generate(tt.config)
			if err != nil {
				t.Errorf("Generate() error = %v", err)
				return
			}

			expectedKeys := []string{"key_base64", "key_hex", "key_size"}
			for _, key := range expectedKeys {
				if _, ok := result[key]; !ok {
					t.Errorf("Generate() missing key %s", key)
				}
			}

			// Validate base64 key length
			decoded, err := base64.StdEncoding.DecodeString(result["key_base64"])
			if err != nil {
				t.Errorf("Generate() invalid base64: %v", err)
				return
			}
			if len(decoded) != tt.bytes {
				t.Errorf("Generate() key length = %d, want %d", len(decoded), tt.bytes)
			}

			// Validate hex key length
			hexDecoded, err := hex.DecodeString(result["key_hex"])
			if err != nil {
				t.Errorf("Generate() invalid hex: %v", err)
				return
			}
			if len(hexDecoded) != tt.bytes {
				t.Errorf("Generate() hex key length = %d, want %d", len(hexDecoded), tt.bytes)
			}
		})
	}

	// Test invalid key size
	t.Run("invalid key size", func(t *testing.T) {
		config := map[string]interface{}{
			"key_size": float64(512),
		}
		_, err := gen.Generate(config)
		if err == nil {
			t.Error("Generate() expected error for invalid key size")
		}
	})
}
