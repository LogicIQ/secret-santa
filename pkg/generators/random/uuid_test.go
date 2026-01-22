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
		return
	}

	if result["value"] == "" {
		t.Error("Generate() value is empty")
		return
	}

	// Validate UUID format
	parsedUUID, err := uuid.Parse(result["value"])
	if err != nil {
		t.Errorf("Generate() invalid UUID format: %v", err)
		return
	}

	// Validate UUID version
	if parsedUUID.Version() != 4 {
		t.Errorf("Generate() expected UUID version 4, got %d", parsedUUID.Version())
		return
	}
}
