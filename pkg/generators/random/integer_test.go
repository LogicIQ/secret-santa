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

			if _, ok := result["value"]; !ok {
				t.Error("Generate() missing value key")
			}

			val, err := strconv.Atoi(result["value"])
			if err != nil {
				t.Errorf("Generate() invalid integer: %v", err)
				return
			}

			if val < tt.min || val > tt.max {
				t.Errorf("Generate() value %d out of range [%d, %d]", val, tt.min, tt.max)
			}
		})
	}
}
