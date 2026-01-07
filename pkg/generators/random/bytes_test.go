package random

import (
	"encoding/base64"
	"testing"
)

func TestBytesGenerator_Generate(t *testing.T) {
	gen := &BytesGenerator{}

	tests := []struct {
		name   string
		config map[string]interface{}
		length int
	}{
		{
			name:   "default length",
			config: map[string]interface{}{},
			length: 16,
		},
		{
			name: "custom length",
			config: map[string]interface{}{
				"length": float64(32),
			},
			length: 32,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := gen.Generate(tt.config)
			if err != nil {
				t.Errorf("Generate() error = %v", err)
				return
			}

			expectedKeys := []string{"value", "generatedAt"}
			for _, key := range expectedKeys {
				if _, ok := result[key]; !ok {
					t.Errorf("Generate() missing key %s", key)
					return
				}
			}

			// Validate base64 value
			decoded, err := base64.StdEncoding.DecodeString(result["value"])
			if err != nil {
				t.Errorf("Generate() invalid base64 value: %v", err)
				return
			}
			if len(decoded) != tt.length {
				t.Errorf("Generate() decoded length = %d, want %d", len(decoded), tt.length)
				return
			}
		})
	}
}
