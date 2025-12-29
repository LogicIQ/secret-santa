package generators

import (
	"testing"
)

type MockGenerator struct{}

func (m *MockGenerator) Generate(config map[string]interface{}) (map[string]string, error) {
	return map[string]string{"value": "test"}, nil
}

func TestRegistry(t *testing.T) {
	Clear()

	// Test registration
	mockGen := &MockGenerator{}
	Register("test_generator", mockGen)

	// Test IsSupported
	if !IsSupported("test_generator") {
		t.Error("Expected test_generator to be supported")
	}

	if IsSupported("non_existent") {
		t.Error("Expected non_existent to not be supported")
	}

	// Test Get
	gen, err := Get("test_generator")
	if err != nil {
		t.Errorf("Expected to get test_generator, got error: %v", err)
	}
	if gen != mockGen {
		t.Error("Expected to get the same generator instance")
	}

	// Test Get with non-existent generator
	_, err = Get("non_existent")
	if err == nil {
		t.Error("Expected error for non-existent generator")
	}

	// Test GetSupportedTypes
	types := GetSupportedTypes()
	if len(types) != 1 || types[0] != "test_generator" {
		t.Errorf("Expected [test_generator], got %v", types)
	}

	// Test multiple registrations
	Register("another_generator", mockGen)
	types = GetSupportedTypes()
	if len(types) != 2 {
		t.Errorf("Expected 2 generators, got %d", len(types))
	}

	// Test Clear
	Clear()
	types = GetSupportedTypes()
	if len(types) != 0 {
		t.Errorf("Expected 0 generators after clear, got %d", len(types))
	}
}