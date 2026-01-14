package random

import (
	"strconv"
	"testing"
)

func TestIntegerGenerator_Generate(t *testing.T) {
	gen := &IntegerGenerator{}

	tests := []struct {
		name   string
		config map[string]interface{}
		min    int
		max    int
	}{
		{
			name:   "default range",
			config: map[string]interface{}{},
			min:    0,
			max:    100,
		},
		{
			name: "custom range",
			config: map[string]interface{}{
				"min": float64(10),
				"max": float64(20),
			},
			min: 10,
			max: 20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := gen.Generate(tt.config)
			if err != nil {
				t.Errorf("Generate() error = %v", err)
				return
			}

			valStr, ok := result["value"]
			if !ok {
				t.Fatal("Generate() missing value key")
			}

			val, err := strconv.Atoi(valStr)
			if err != nil {
				t.Fatalf("Generate() invalid integer: %v", err)
			}

			if val < tt.min || val > tt.max {
				t.Errorf("Generate() value %d out of range [%d, %d]", val, tt.min, tt.max)
			}
		})
	}

	// Test error case
	t.Run("min greater than max", func(t *testing.T) {
		config := map[string]interface{}{
			"min": float64(20),
			"max": float64(10),
		}
		_, err := gen.Generate(config)
		if err == nil {
			t.Error("Generate() expected error when min > max")
		}
	})
}
