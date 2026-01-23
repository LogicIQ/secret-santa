package random

import (
	"encoding/hex"
	"testing"
)

func TestIDGenerator_Generate(t *testing.T) {
	gen := &IDGenerator{}

	tests := []struct {
		name   string
		config map[string]interface{}
		length int
	}{
		{
			name:   "default length",
			config: map[string]interface{}{},
			length: 8,
		},
		{
			name: "custom length",
			config: map[string]interface{}{
				"byte_length": 16,
			},
			length: 16,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := gen.Generate(tt.config)
			if err != nil {
				t.Errorf("Generate() error = %v", err)
				return
			}

			expectedKeys := []string{"value", "prefix", "generatedAt"}
			for _, key := range expectedKeys {
				if _, ok := result[key]; !ok {
					t.Fatalf("Generate() missing key %s", key)
				}
			}

			// Validate hex value (without prefix)
			if result["value"] == "" {
				t.Fatal("Generate() value is empty or missing")
			}
			hexValue := result["value"]
			// Strip prefix if present
			prefix := result["prefix"]
			if prefix != "" && len(hexValue) > len(prefix) {
				hexValue = hexValue[len(prefix):]
			}
			if hexValue == "" {
				t.Fatal("Generate() hex value is empty after stripping prefix")
			}
			hexDecoded, err := hex.DecodeString(hexValue)
			if err != nil {
				t.Fatalf("Generate() invalid hex value %q: %v", hexValue, err)
			}
			if len(hexDecoded) != tt.length {
				t.Errorf("Generate() hex length = %d, want %d", len(hexDecoded), tt.length)
			}
		})
	}
}
