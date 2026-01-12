package time

import (
	"testing"
	"time"
)

func TestStaticGenerator_Generate(t *testing.T) {
	gen := &StaticGenerator{}

	tests := []struct {
		name   string
		config map[string]interface{}
	}{
		{
			name:   "default current time",
			config: map[string]interface{}{},
		},
		{
			name: "custom rfc3339",
			config: map[string]interface{}{
				"rfc3339": "2023-01-01T00:00:00Z",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := gen.Generate(tt.config)
			if err != nil {
				t.Errorf("Generate() error = %v", err)
				return
			}

			expectedKeys := []string{"rfc3339", "unix", "year", "month", "day", "hour", "minute", "second"}
			for _, key := range expectedKeys {
				if _, ok := result[key]; !ok {
					t.Errorf("Generate() missing key %s", key)
				}
			}

			// Validate RFC3339 format
			_, err = time.Parse(time.RFC3339, result["rfc3339"])
			if err != nil {
				t.Errorf("Generate() invalid RFC3339 format: %v", err)
				return
			}
		})
	}

	// Test invalid RFC3339
	t.Run("invalid rfc3339", func(t *testing.T) {
		config := map[string]interface{}{
			"rfc3339": "invalid-time",
		}
		_, err := gen.Generate(config)
		if err == nil {
			t.Error("Generate() expected error for invalid RFC3339")
		}
	})
}
