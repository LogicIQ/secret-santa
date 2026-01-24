package random

import (
	"github.com/google/uuid"
	"testing"
)

func TestUUIDGenerator_Generate(t *testing.T) {
	gen := &UUIDGenerator{}

	result, err := gen.Generate(map[string]interface{}{})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	value, ok := result["value"]
	if !ok {
		t.Fatal("Generate() missing value key")
	}

	if value == "" {
		t.Fatal("Generate() value is empty")
	}

	// Validate UUID format
	parsedUUID, err := uuid.Parse(value)
	if err != nil {
		t.Fatalf("Generate() invalid UUID format: %v", err)
	}

	// Validate UUID version
	if parsedUUID.Version() != 4 {
		t.Errorf("Generate() expected UUID version 4, got %d", parsedUUID.Version())
	}
}
