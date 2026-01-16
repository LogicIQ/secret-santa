package random

import (
	"testing"
)

func TestPasswordGenerator_Generate(t *testing.T) {
	gen := &PasswordGenerator{}

	tests := []struct {
		name   string
		config map[string]interface{}
		want   []string // expected keys
	}{
		{
			name:   "default config",
			config: map[string]interface{}{},
			want:   []string{"value", "charset", "generatedAt"},
		},
		{
			name: "custom length",
			config: map[string]interface{}{
				"length": float64(32),
			},
			want: []string{"value", "charset", "generatedAt"},
		},
		{
			name: "no special chars",
			config: map[string]interface{}{
				"length":  float64(16),
				"special": false,
			},
			want: []string{"value", "charset", "generatedAt"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := gen.Generate(tt.config)
			if err != nil {
				t.Errorf("Generate() error = %v", err)
				return
			}

			for _, key := range tt.want {
				if _, ok := result[key]; !ok {
					t.Errorf("Generate() missing key %s", key)
				}
			}

			value := result["value"]
			if len(value) == 0 {
				t.Error("Generate() value is empty")
				return
			}
			expectedLen := 16
			if l, ok := tt.config["length"].(float64); ok {
				expectedLen = int(l)
			}
			if len(value) != expectedLen {
				t.Errorf("Generate() value length = %d, want %d", len(value), expectedLen)
			}
		})
	}
}
