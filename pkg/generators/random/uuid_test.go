package random

import (
	"github.com/google/uuid"
	"testing"
)

func TestUUIDGenerator_Generate(t *testing.T) {
	gen := &UUIDGenerator{}

	result, err := gen.Generate(map[string]interface{}{})
	if err != nil {
		t.Errorf("Generate() error = %v", err)
		return
	}

	if _, ok := result["value"]; !ok {
		t.Error("Generate() missing value key")
	}

	// Validate UUID format
	_, err = uuid.Parse(result["value"])
	if err != nil {
		t.Errorf("Generate() invalid UUID format: %v", err)
	}
}
