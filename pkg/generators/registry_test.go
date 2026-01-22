package generators

import (
	"testing"
)

type MockGenerator struct{}

func (m *MockGenerator) Generate(config map[string]interface{}) (map[string]string, error) {
	return map[string]string{"value": "test"}, nil
}

func TestRegistry(t *testing.T) {
	mockGen := &MockGenerator{}

	t.Run("Register", func(t *testing.T) {
		Clear()
		err := Register("test_generator", mockGen)
		if err != nil {
			t.Errorf("Expected successful registration, got error: %v", err)
		}
		err = Register("test_generator", mockGen)
		if err == nil {
			t.Error("Expected error when registering duplicate generator")
		}
	})

	t.Run("IsSupported", func(t *testing.T) {
		Clear()
		if err := Register("test_generator", mockGen); err != nil {
			t.Fatalf("Failed to register test_generator: %v", err)
		}
		if !IsSupported("test_generator") {
			t.Error("Expected test_generator to be supported")
		}
		if IsSupported("non_existent") {
			t.Error("Expected non_existent to not be supported")
		}
	})

	t.Run("Get", func(t *testing.T) {
		Clear()
		if err := Register("test_generator", mockGen); err != nil {
			t.Fatalf("Failed to register test_generator: %v", err)
		}
		gen, err := Get("test_generator")
		if err != nil {
			t.Errorf("Expected to get test_generator, got error: %v", err)
		}
		if gen != mockGen {
			t.Error("Expected to get the same generator instance")
		}
		_, err = Get("non_existent")
		if err == nil {
			t.Error("Expected error for non-existent generator")
		}
	})

	t.Run("GetSupportedTypes", func(t *testing.T) {
		Clear()
		if err := Register("test_generator", mockGen); err != nil {
			t.Fatalf("Failed to register test_generator: %v", err)
		}
		types := GetSupportedTypes()
		if len(types) != 1 || types[0] != "test_generator" {
			t.Errorf("Expected [test_generator], got %v", types)
		}
		if err := Register("another_generator", mockGen); err != nil {
			t.Fatalf("Failed to register another_generator: %v", err)
		}
		types = GetSupportedTypes()
		if len(types) != 2 {
			t.Errorf("Expected 2 generators, got %d", len(types))
		}
	})

	t.Run("Clear", func(t *testing.T) {
		Clear()
		if err := Register("test_generator", mockGen); err != nil {
			t.Fatalf("Failed to register test_generator: %v", err)
		}
		Clear()
		types := GetSupportedTypes()
		if len(types) != 0 {
			t.Errorf("Expected 0 generators after clear, got %d", len(types))
		}
	})
}
