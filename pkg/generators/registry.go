package generators

import (
	"fmt"
	"strings"
	"sync"
)

type Registry struct {
	mu         sync.RWMutex
	generators map[string]Generator
}

func NewRegistry() *Registry {
	return &Registry{
		generators: make(map[string]Generator),
	}
}

var globalRegistry = NewRegistry()

func Register(generatorType string, generator Generator) error {
	if strings.TrimSpace(generatorType) == "" {
		return fmt.Errorf("generator type cannot be empty")
	}
	if generator == nil {
		return fmt.Errorf("generator cannot be nil")
	}
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()
	if _, exists := globalRegistry.generators[generatorType]; exists {
		return fmt.Errorf("generator type %s is already registered", generatorType)
	}
	globalRegistry.generators[generatorType] = generator
	return nil
}

func Get(generatorType string) (Generator, error) {
	if strings.TrimSpace(generatorType) == "" {
		return nil, fmt.Errorf("generator type cannot be empty")
	}
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	generator, exists := globalRegistry.generators[generatorType]
	if !exists {
		return nil, fmt.Errorf("unsupported generator type: %s", generatorType)
	}
	return generator, nil
}

func GetSupportedTypes() []string {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	types := make([]string, 0, len(globalRegistry.generators))
	for generatorType := range globalRegistry.generators {
		types = append(types, generatorType)
	}
	return types
}

func IsSupported(generatorType string) bool {
	if strings.TrimSpace(generatorType) == "" {
		return false
	}
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	_, exists := globalRegistry.generators[generatorType]
	return exists
}

func Clear() {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()
	globalRegistry.generators = make(map[string]Generator)
}
