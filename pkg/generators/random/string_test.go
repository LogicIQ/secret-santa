package random

import (
	"testing"
)

func TestStringGenerator_Generate(t *testing.T) {
	gen := &StringGenerator{}

	tests := []struct {
		name   string
		config map[string]interface{}
		want   []string
	}{
		{
			name:   "default config",
			config: map[string]interface{}{},
			want:   []string{"value", "charset", "generatedAt"},
		},
		{
			name: "custom length",
			config: map[string]interface{}{
				"length": float64(64),
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
		})
	}
}
